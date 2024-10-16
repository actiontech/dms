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
	Debug       bool // 暂时无用
}

func NewStorage(logger pkgLog.Logger, conf *StorageConfig) (*Storage, error) {
	log := pkgLog.NewHelper(logger, pkgLog.WithMessageKey("dms.storage"))
	log.Infof("connecting to storage, host: %s, port: %s, user: %s, schema: %s",
		conf.Host, conf.Port, conf.User, conf.Schema)

	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Schema)), &gorm.Config{
		Logger:                                   pkgLog.NewGormLogWrapper(pkgLog.NewKLogWrapper(logger), gormLog.Info),
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
		if ok {
			return fmt.Sprintf("%s %s (%s)", condition.Field, condition.Operator, strings.Join(values, ",")), nil
		}
	}
	return fmt.Sprintf("%s %s ?", condition.Field, condition.Operator), condition.Value
}

func gormWheres(ctx context.Context, db *gorm.DB, conditions []pkgConst.FilterCondition) *gorm.DB {
	fuzzyWhere := db.WithContext(ctx)
	singleWhere := db.WithContext(ctx)

	for _, condition := range conditions {
		if condition.Table != "" {
			continue
		}
		if condition.KeywordSearch {
			// 模糊查询字段
			fuzzyWhere = fuzzyWhere.Or(gormWhere(singleWhere, condition))
		} else {
			db = gormWhere(db, condition)
		}
	}
	db = db.Where(fuzzyWhere)
	return db
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

func fixPageIndices(page_number uint32) int {
	if page_number <= 0 {
		page_number = 1
	}

	page_index := int(page_number - 1)
	return page_index
}
