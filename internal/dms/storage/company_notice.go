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

var _ biz.CompanyNoticeRepo = (*CompanyNoticeRepo)(nil)

type CompanyNoticeRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewCompanyNoticeRepo(log utilLog.Logger, s *Storage) *CompanyNoticeRepo {
	return &CompanyNoticeRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.company_notice"))}
}
func (d *CompanyNoticeRepo) UpdateCompanyNotice(ctx context.Context, o2c *biz.CompanyNotice) error {
	model, err := convertBizCompanyNotice(o2c)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz company notice: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save company notice configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *CompanyNoticeRepo) GetCompanyNotice(ctx context.Context) (*biz.CompanyNotice, error) {
	var companyNotice *model.CompanyNotice
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&companyNotice).Error; err != nil {
			return fmt.Errorf("failed to get company notice: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModelCompanyNotice(companyNotice)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model company notice: %w", err))
	}
	return ret, nil
}
