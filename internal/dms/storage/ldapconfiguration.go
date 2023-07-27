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

var _ biz.LDAPConfigurationRepo = (*LDAPConfigurationRepo)(nil)

type LDAPConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewLDAPConfigurationRepo(log utilLog.Logger, s *Storage) *LDAPConfigurationRepo {
	return &LDAPConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.ldap_configuration"))}
}
func (d *LDAPConfigurationRepo) UpdateLDAPConfiguration(ctx context.Context, ldapC *biz.LDAPConfiguration) error {
	model, err := convertBizLDAPConfiguration(ldapC)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz ldap configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save ldap configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *LDAPConfigurationRepo) GetLastLDAPConfiguration(ctx context.Context) (*biz.LDAPConfiguration, error) {
	var ldapConfiguration *model.LDAPConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&ldapConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get ldap configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModelLDAPConfiguration(ldapConfiguration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model ldap configuration: %w", err))
	}
	return ret, nil
}
