package storage

import (
	"context"
	"fmt"
	"strings"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	pkgLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

type Storage struct {
	db *gorm.DB
}

func (s *Storage) Close() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

type StorageConfig struct {
	User        string
	Password    string
	Host        string
	Port        string
	Schema      string
	AutoMigrate bool
	Debug       bool
}

func NewStorage(logger pkgLog.Logger, conf *StorageConfig) (*Storage, error) {
	log := pkgLog.NewHelper(logger, pkgLog.WithMessageKey("dms.storage"))
	log.Infof("connecting to storage, host: %s, port: %s, user: %s, schema: %s",
		conf.Host, conf.Port, conf.User, conf.Schema)

	// 根据 Debug 配置设置 GORM 日志级别
	// Debug = false: 使用 Warn 级别，输出警告和错误日志，但不输出 SQL 查询
	// Debug = true: 使用 Info 级别，输出详细 SQL 日志
	logLevel := gormLog.Warn
	if conf.Debug {
		logLevel = gormLog.Info
		log.Info("gorm debug mode enabled")
	}

	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Schema)), &gorm.Config{
		Logger:                                   pkgLog.NewGormLogWrapper(pkgLog.NewKLogWrapper(logger), logLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Errorf("connect to storage failed, error: %v", err)
		return nil, pkgErr.WrapStorageErr(log, err)
	}

	s := &Storage{db: db}
	if conf.AutoMigrate {
		if err := s.AutoMigrate(logger); err != nil {
			log.Errorf("auto migrate failed, error: %v", err)
			return nil, pkgErr.WrapStorageErr(log, err)
		}
		log.Info("auto migrate dms tables")
	}
	log.Info("connected to storage")
	return s, pkgErr.WrapStorageErr(log, err)
}

func (s *Storage) AutoMigrate(logger pkgLog.Logger) error {
	log := pkgLog.NewHelper(logger, pkgLog.WithMessageKey("dms.storage.AutoMigrate"))
	err := s.db.AutoMigrate(model.AutoMigrateList...)
	if err != nil {
		return pkgErr.WrapStorageErr(log, err)
	}
	return nil
}

func gormWhere(db *gorm.DB, condition pkgConst.FilterCondition) *gorm.DB {
	// TODO  临时解决ISNULL场景不需要参数问题
	query, arg := gormWhereCondition(condition)
	if arg == nil {
		return db.Where(query)
	}
	return db.Where(query, arg)
}

func gormWhereCondition(condition pkgConst.FilterCondition) (string, interface{}) {
	switch condition.Operator {
	case pkgConst.FilterOperatorIsNull:
		return fmt.Sprintf("%s IS NULL", condition.Field), nil
	case pkgConst.FilterOperatorContains, pkgConst.FilterOperatorNotContains:
		condition.Value = fmt.Sprintf("%%%s%%", condition.Value)
	case pkgConst.FilterOperatorIn:
		values, ok := condition.Value.([]string)
		if ok && len(values) > 0 {
			var itemList []string
			for _, value := range values {
				item := fmt.Sprintf(`'%v'`, value)
				itemList = append(itemList, item)
			}

			return fmt.Sprintf("%s %s (%s)", condition.Field, condition.Operator, strings.Join(itemList, ",")), nil
		}
	}
	return fmt.Sprintf("%s %s ?", condition.Field, condition.Operator), condition.Value
}

func gormWheresWithOptions(ctx context.Context, db *gorm.DB, opts pkgConst.FilterOptions) *gorm.DB {
	if len(opts.Groups) == 0 {
		return db.WithContext(ctx)
	}

	db = db.WithContext(ctx)
	groupQueries := make([]*gorm.DB, 0, len(opts.Groups))

	for _, group := range opts.Groups {
		groupQuery := buildConditionGroup(db, group)
		if groupQuery != nil {
			groupQueries = append(groupQueries, groupQuery)
		}
	}

	if len(groupQueries) == 0 {
		return db
	}

	result := groupQueries[0]
	for i := 1; i < len(groupQueries); i++ {
		if opts.Logic == pkgConst.FilterLogicOr {
			result = db.Where(result).Or(groupQueries[i])
		} else {
			result = db.Where(result).Where(groupQueries[i])
		}
	}

	return db.Where(result)
}

func buildConditionGroup(db *gorm.DB, group pkgConst.FilterConditionGroup) *gorm.DB {
	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return nil
	}

	var result *gorm.DB

	for i, condition := range group.Conditions {
		if condition.Table != "" {
			continue
		}
		condQuery := gormWhere(db, condition)
		if i == 0 || result == nil {
			result = condQuery
		} else if group.Logic == pkgConst.FilterLogicOr {
			result = db.Where(result).Or(condQuery)
		} else {
			result = db.Where(result).Where(condQuery)
		}
	}

	for _, subGroup := range group.Groups {
		subQuery := buildConditionGroup(db, subGroup)
		if subQuery == nil {
			continue
		}
		if result == nil {
			result = subQuery
		} else if group.Logic == pkgConst.FilterLogicOr {
			result = db.Where(result).Or(subQuery)
		} else {
			result = db.Where(result).Where(subQuery)
		}
	}

	return result
}

func gormPreload(ctx context.Context, db *gorm.DB, conditions []pkgConst.FilterCondition) *gorm.DB {
	for _, f := range conditions {
		// Preload 关联表
		if f.Table != "" {
			args := make([]interface{}, 0)
			// Preload 筛选参数
			if f.Field != "" {
				whereCondition, value := gormWhereCondition(f)
				args = []interface{}{whereCondition, value}
			}
			db = db.Preload(f.Table, args)
		}
	}
	return db
}

func extractPreloadConditions(opts pkgConst.FilterOptions) []pkgConst.FilterCondition {
	var conditions []pkgConst.FilterCondition
	for _, group := range opts.Groups {
		for _, condition := range group.Conditions {
			if condition.Table != "" {
				conditions = append(conditions, condition)
			}
		}
		for _, subGroup := range group.Groups {
			subConditions := extractPreloadConditions(pkgConst.FilterOptions{Groups: []pkgConst.FilterConditionGroup{subGroup}})
			conditions = append(conditions, subConditions...)
		}
	}
	return conditions
}

func gormPreloadFromOptions(ctx context.Context, db *gorm.DB, opts pkgConst.FilterOptions) *gorm.DB {
	conditions := extractPreloadConditions(opts)
	return gormPreload(ctx, db, conditions)
}

func fixPageIndices(page_number uint32) int {
	if page_number <= 0 {
		page_number = 1
	}

	page_index := int(page_number - 1)
	return page_index
}
