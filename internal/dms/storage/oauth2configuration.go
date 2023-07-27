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

var _ biz.Oauth2ConfigurationRepo = (*Oauth2ConfigurationRepo)(nil)

type Oauth2ConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOauth2ConfigurationRepo(log utilLog.Logger, s *Storage) *Oauth2ConfigurationRepo {
	return &Oauth2ConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.oauth2_configuration"))}
}
func (d *Oauth2ConfigurationRepo) UpdateOauth2Configuration(ctx context.Context, o2c *biz.Oauth2Configuration) error {
	model, err := convertBizOauth2Configuration(o2c)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz oauth2 configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save oauth2 configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *Oauth2ConfigurationRepo) GetLastOauth2Configuration(ctx context.Context) (*biz.Oauth2Configuration, error) {
	var oauth2Configuration *model.Oauth2Configuration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&oauth2Configuration).Error; err != nil {
			return fmt.Errorf("failed to get oauth2 configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModelOauth2Configuration(oauth2Configuration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model oauth2 configuration: %w", err))
	}
	return ret, nil
}
