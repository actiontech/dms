package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"gorm.io/gorm"
)

var _ biz.DMSConfigRepo = (*DMSConfigRepo)(nil)

type DMSConfigRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewDMSConfigRepo(log utilLog.Logger, s *Storage) *DMSConfigRepo {
	return &DMSConfigRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.dms_config"))}
}

func (d *DMSConfigRepo) SaveDMSConfig(ctx context.Context, u *biz.DMSConfig) error {
	model, err := convertBizDMSConfig(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz config: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save config: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *DMSConfigRepo) CheckDMSConfigExist(ctx context.Context, configUid string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DMSConfig{}).Where("uid = ?", configUid).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check dms config exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != 1 {
		return false, nil
	}
	return true, nil
}

func (d *DMSConfigRepo) UpdateDMSConfig(ctx context.Context, u *biz.DMSConfig) error {
	exist, err := d.CheckDMSConfigExist(ctx, u.UID)
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("dms config not exist"))
	}

	config, err := convertBizDMSConfig(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz dms config: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DMSConfig{}).Where("uid = ?", u.UID).Omit("created_at").Save(config).Error; err != nil {
			return fmt.Errorf("failed to update dms config: %v", err)
		}
		return nil
	})

}

func (d *DMSConfigRepo) DelDMSConfig(ctx context.Context, configUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", configUid).Delete(&model.DMSConfig{}).Error; err != nil {
			return fmt.Errorf("failed to delete dms config: %v", err)
		}
		return nil
	})
}

func (d *DMSConfigRepo) GetDMSConfig(ctx context.Context, configUid string) (*biz.DMSConfig, error) {
	var config *model.DMSConfig
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&config, "uid = ?", configUid).Error; err != nil {
			return fmt.Errorf("failed to get dms config: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelDMSConfig(config)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model dms config: %v", err))
	}
	return ret, nil
}
