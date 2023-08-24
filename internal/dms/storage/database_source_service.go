package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.DatabaseSourceServiceRepo = (*DatabaseSourceServiceRepo)(nil)

type DatabaseSourceServiceRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewDatabaseSourceServiceRepo(log utilLog.Logger, s *Storage) *DatabaseSourceServiceRepo {
	return &DatabaseSourceServiceRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.database_source_service"))}
}

func (d *DatabaseSourceServiceRepo) ListDatabaseSourceServices(ctx context.Context, conditions []pkgConst.FilterCondition) ([]*biz.DatabaseSourceServiceParams, error) {
	var items []*model.DatabaseSourceService
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		db := tx.WithContext(ctx)
		for _, f := range conditions {
			db = gormWhere(db, f)
		}

		if err := db.Find(&items).Error; err != nil {
			return fmt.Errorf("failed to list database_source_service: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.DatabaseSourceServiceParams, 0, len(items))
	// convert model to biz
	for _, item := range items {
		ds, err := convertModelDatabaseSourceService(item)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert database_source_service: %w", err))
		}
		ret = append(ret, ds)
	}

	return ret, nil
}

func (d *DatabaseSourceServiceRepo) SaveDatabaseSourceService(ctx context.Context, params *biz.DatabaseSourceServiceParams) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(convertBizDatabaseSourceService(params)).Error; err != nil {
			return fmt.Errorf("failed to save database_source_service: %v", err)
		}

		return nil
	})
}

func (d *DatabaseSourceServiceRepo) UpdateDatabaseSourceService(ctx context.Context, params *biz.DatabaseSourceServiceParams) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Omit("created_at").Save(convertBizDatabaseSourceService(params)).Error; err != nil {
			return fmt.Errorf("failed to update database_source_service: %v", err)
		}

		return nil
	})
}

func (d *DatabaseSourceServiceRepo) UpdateSyncDatabaseSourceService(ctx context.Context, databaseSourceServiceUid string, fields map[string]interface{}) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DatabaseSourceService{}).Where("uid = ?", databaseSourceServiceUid).Omit("updated_at").Updates(fields).Error; err != nil {
			return fmt.Errorf("failed to update database_source_service: %v", err)
		}

		return nil
	})
}

func (d *DatabaseSourceServiceRepo) DeleteDatabaseSourceService(ctx context.Context, databaseSourceServiceUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", databaseSourceServiceUid).Delete(&model.DatabaseSourceService{}).Error; err != nil {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to delete database_source_service: %v", err))
		}
		return nil
	})
}

func (d *DatabaseSourceServiceRepo) GetDatabaseSourceServiceById(ctx context.Context, id string) (*biz.DatabaseSourceServiceParams, error) {
	var databaseSourceService *model.DatabaseSourceService
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&databaseSourceService, "uid = ?", id).Error; err != nil {
			return fmt.Errorf("failed to get database_source_service: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelDatabaseSourceService(databaseSourceService)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model database_source_service: %w", err))
	}

	return ret, nil
}
