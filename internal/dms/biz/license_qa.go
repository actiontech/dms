//go:build !release
// +build !release

package biz

import (
	"context"
	e "errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
)

type License struct {
}

var ErrNoLicenseRequired = e.New("dms-qa version has unlimited resources does not need to set license")

func (d *LicenseUsecase) GetLicense(c context.Context) (*v1.GetLicenseReply, error) {
	return &v1.GetLicenseReply{
		License: []v1.LicenseItem{
			{
				Description: "实例数",
				Name:        "instance_num",
				Limit:       "无限制",
			},
			{
				Description: "用户数",
				Name:        "user",
				Limit:       "无限制",
			},
			{
				Description: "授权运行时长(天)",
				Name:        "work duration day",
				Limit:       "无限制",
			},
		},
	}, nil
}

func (d *LicenseUsecase) GetLicenseInfo(ctx context.Context) ([]byte, error) {
	return []byte{}, ErrNoLicenseRequired
}

func (d *LicenseUsecase) SetLicense(ctx context.Context, data string) error {
	return ErrNoLicenseRequired
}

func (d *LicenseUsecase) CheckLicense(ctx context.Context, data string) (*v1.CheckLicenseReply, error) {
	return nil, ErrNoLicenseRequired
}

func (d *LicenseUsecase) initial() {

}
