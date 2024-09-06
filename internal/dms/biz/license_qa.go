//go:build !release
// +build !release

package biz

import (
	"context"
	e "errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/pkg/locale"
)

type License struct {
}

var ErrNoLicenseRequired = e.New("dms-qa version has unlimited resources does not need to set license")

func (d *LicenseUsecase) GetLicense(ctx context.Context) (*v1.GetLicenseReply, error) {
	return &v1.GetLicenseReply{
		License: []v1.LicenseItem{
			{
				Description: locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseInstanceNum),
				Name:        "instance_num",
				Limit:       locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseUnlimited),
			},
			{
				Description: locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseUserNum),
				Name:        "user",
				Limit:       locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseUnlimited),
			},
			{
				Description: locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseAuthorizedDurationDay),
				Name:        "work duration day",
				Limit:       locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseUnlimited),
			},
		},
	}, nil
}

func (d *LicenseUsecase) GetLicenseInfo(ctx context.Context) ([]byte, error) {
	return []byte{}, ErrNoLicenseRequired
}

func (d *LicenseUsecase) GetLicenseUsage(ctx context.Context) (*v1.GetLicenseUsageReply, error) {
	usersTotal, err := d.userUsecase.repo.CountUsers(ctx, nil)
	if err != nil {
		return nil, err
	}

	instanceStatistics, err := d.DBService.CountDBService(ctx)
	if err != nil {
		return nil, err
	}

	dbServicesUsage := make([]v1.LicenseUsageItem, 0, len(instanceStatistics))
	for _, item := range instanceStatistics {
		dbServicesUsage = append(dbServicesUsage, v1.LicenseUsageItem{
			ResourceType:     item.DBType,
			ResourceTypeDesc: item.DBType,
			Used:             uint(item.Count),
			Limit:            0,
			IsLimited:        false,
		})
	}

	return &v1.GetLicenseUsageReply{
		Data: &v1.LicenseUsage{
			UsersUsage: v1.LicenseUsageItem{
				ResourceType:     "user",
				ResourceTypeDesc: locale.Bundle.ShouldLocalizeMsg(ctx, locale.LicenseResourceTypeUser),
				Used:             uint(usersTotal),
				Limit:            0,
				IsLimited:        false,
			},
			DbServicesUsage: dbServicesUsage,
		},
	}, nil
}

func (d *LicenseUsecase) SetLicense(ctx context.Context, data string) error {
	return ErrNoLicenseRequired
}

func (d *LicenseUsecase) CheckLicense(ctx context.Context, data string) (*v1.CheckLicenseReply, error) {
	return nil, ErrNoLicenseRequired
}

func (d *LicenseUsecase) initial() {

}
