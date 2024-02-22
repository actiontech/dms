package service

import (
	"context"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) GetLicense(ctx context.Context) (*v1.GetLicenseReply, error) {
	return d.LicenseUsecase.GetLicense(ctx)
}

func (d *DMSService) GetLicenseInfo(ctx context.Context) ([]byte, error) {
	return d.LicenseUsecase.GetLicenseInfo(ctx)
}

func (d *DMSService) SetLicense(ctx context.Context, data string) error {
	return d.LicenseUsecase.SetLicense(ctx, data)
}

func (d *DMSService) CheckLicense(ctx context.Context, data string) (*v1.CheckLicenseReply, error) {
	return d.LicenseUsecase.CheckLicense(ctx, data)
}

func (d *DMSService) GetLicenseUsage(ctx context.Context) (*v1.GetLicenseUsageReply, error) {
	return d.LicenseUsecase.GetLicenseUsage(ctx)
}
