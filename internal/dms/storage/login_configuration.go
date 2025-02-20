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

var _ biz.LoginConfigurationRepo = (*LoginConfigurationRepo)(nil)

type LoginConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewLoginConfigurationRepo(log utilLog.Logger, s *Storage) *LoginConfigurationRepo {
	return &LoginConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.Login_configuration"))}
}

func (d *LoginConfigurationRepo) UpdateLoginConfiguration(ctx context.Context, loginConfiguration *biz.LoginConfiguration) error {
	modelLoginConfiguration, err := convertBizLoginConfiguration(loginConfiguration)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz Login configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(modelLoginConfiguration).Where("uid = ?", modelLoginConfiguration.UID).Omit("created_at").Save(modelLoginConfiguration).Error; err != nil {
			return fmt.Errorf("failed to save Login configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *LoginConfigurationRepo) GetLastLoginConfiguration(ctx context.Context) (*biz.LoginConfiguration, error) {
	var LoginConfiguration *model.LoginConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&LoginConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get Login configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModelLoginConfiguration(LoginConfiguration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model Login configuration: %w", err))
	}
	return ret, nil
}
