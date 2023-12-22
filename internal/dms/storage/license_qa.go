//go:build !release
// +build !release

package storage

import (
	"context"

	"github.com/actiontech/dms/internal/dms/biz"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var _ biz.LicenseRepo = (*LicenseRepo)(nil)

type LicenseRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewLicenseRepo(log utilLog.Logger, s *Storage) *LicenseRepo {
	return &LicenseRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.license"))}
}

func (d *LicenseRepo) SaveLicense(ctx context.Context, license interface{}) error {
	return nil
}

func (d *LicenseRepo) GetLastLicense(ctx context.Context) (interface{}, bool, error) {
	return nil, false, nil
}

func (d *LicenseRepo) GetLicenseById(ctx context.Context, id string) (interface{}, bool, error) {
	return nil, true, nil
}

func (d *LicenseRepo) UpdateLicense(ctx context.Context, l interface{}) error {
	return nil
}

func (d *LicenseRepo) DelLicense(ctx context.Context) error {
	return nil
}
