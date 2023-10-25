package storage

import (
	"fmt"

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
	User     string
	Password string
	Host     string
	Port     string
	Schema   string
	Debug    bool // 暂时无用
}

func NewStorage(logger pkgLog.Logger, conf *StorageConfig) (*Storage, error) {
	log := pkgLog.NewHelper(logger, pkgLog.WithMessageKey("dms.storage"))
	log.Infof("connecting to storage, host: %s, port: %s, user: %s, schema: %s",
		conf.Host, conf.Port, conf.User, conf.Schema)

	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Schema)), &gorm.Config{
		Logger: pkgLog.NewGormLogWrapper(pkgLog.NewKLogWrapper(logger), gormLog.Info),
	})
	if err != nil {
		log.Errorf("connect to storage failed, error: %v", err)
		return nil, pkgErr.WrapStorageErr(log, err)
	}

	s := &Storage{db: db}
	if err := s.AutoMigrate(logger); err != nil {
		log.Errorf("auto migrate failed, error: %v", err)
		return nil, pkgErr.WrapStorageErr(log, err)
	}
	log.Info("connected to storage")
	return s, pkgErr.WrapStorageErr(log, err)
}

func (s *Storage) AutoMigrate(logger pkgLog.Logger) error {
	log := pkgLog.NewHelper(logger, pkgLog.WithMessageKey("dms.storage.AutoMigrate"))
	err := s.db.AutoMigrate(model.GetAllModels()...)
	if err != nil {
		return pkgErr.WrapStorageErr(log, err)
	}
	return nil
}

func gormWhere(db *gorm.DB, condition pkgConst.FilterCondition) *gorm.DB {
	if condition.Operator == pkgConst.FilterOperatorIsNull {
		return db.Where(fmt.Sprintf("%s IS NULL", condition.Field))
	}
	return db.Where(fmt.Sprintf("%s %s ?", condition.Field, condition.Operator), condition.Value)
}

func fixPageIndices(page_number uint32) int {
	if page_number <= 0 {
		page_number = 1
	}

	page_index := int(page_number - 1)
	return page_index
}
