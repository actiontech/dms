//go:build release
// +build release

package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

func init() {
	model.AutoMigrateList = append(model.AutoMigrateList, biz.License{})
}

var _ biz.LicenseRepo = (*LicenseRepo)(nil)

type LicenseRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewLicenseRepo(log utilLog.Logger, s *Storage) *LicenseRepo {
	return &LicenseRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.license"))}
}
func (d *LicenseRepo) SaveLicense(ctx context.Context, license interface{}) error {
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(license).Error; err != nil {
			return fmt.Errorf("failed to save license: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *LicenseRepo) GetLastLicense(ctx context.Context) (interface{}, bool, error) {
	var license *biz.License
	var err error

	if err = d.db.Order("created_at DESC").First(&license).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		} else {
			return nil, false, fmt.Errorf("failed to get license: %v", err)
		}
	}

	return license, true, nil
}

func (d *LicenseRepo) GetLicenseById(ctx context.Context, id string) (interface{}, bool, error) {
	var license *biz.License
	var err error

	if err = d.db.Where("uid = ?", id).First(&license).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		} else {
			return nil, false, fmt.Errorf("failed to get license: %v", err)
		}
	}

	return license, true, nil
}

func (d *LicenseRepo) UpdateLicense(ctx context.Context, license interface{}) error {
	l, ok := license.(*biz.License)
	if !ok {
		return fmt.Errorf("license is invalid")
	}
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&biz.License{}).Where("uid = ?", l.UID).Omit("created_at").Save(l).Error; err != nil {
			return fmt.Errorf("failed to update license: %v", err)
		}
		return nil
	})

}

func (d *LicenseRepo) DelLicense(ctx context.Context) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("1 = 1").Delete(&biz.License{}).Error; err != nil {
			return fmt.Errorf("failed to delete license: %v", err)
		}
		return nil
	})
}
