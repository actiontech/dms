package biz

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/actiontech/dms/internal/dms/conf"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type BasicConfigRepo interface {
	GetBasicConfig(ctx context.Context) (*BasicConfigParams, error)
	SaveBasicConfig(ctx context.Context, params *BasicConfigParams) error
}

type BasicUsecase struct {
	basicConfigRepo BasicConfigRepo
	log             *utilLog.Helper
	dmsProxyUsecase *DmsProxyUsecase
}

func NewBasicInfoUsecase(log utilLog.Logger, dmsProxyUsecase *DmsProxyUsecase, repo BasicConfigRepo) *BasicUsecase {
	return &BasicUsecase{
		basicConfigRepo: repo,
		log:             utilLog.NewHelper(log, utilLog.WithMessageKey("biz.basic")),
		dmsProxyUsecase: dmsProxyUsecase,
	}
}

type BasicConfigParams struct {
	Base
	UID   string                `json:"uid"`
	Title string                `json:"title"`
	File  *multipart.FileHeader `json:"file"`
	Logo  []byte                `json:"logo"`
}

type ComponentNameWithVersion struct {
	Name    string
	Version string
}
type BasicInfo struct {
	LogoUrl    string                     `json:"logo_url"`
	Title      string                     `json:"title"`
	Components []ComponentNameWithVersion `json:"components"`
}

const (
	PersonalizationUrl = "/dms/personalization/logo"
)

func (d *BasicUsecase) GetBasicInfo(ctx context.Context) (*BasicInfo, error) {
	targets, err := d.dmsProxyUsecase.ListProxyTargets(ctx)
	if err != nil {
		return nil, err
	}

	basicConfig, err := d.basicConfigRepo.GetBasicConfig(ctx)
	if err != nil {
		return nil, err
	}

	ret := &BasicInfo{
		Components: []ComponentNameWithVersion{
			{
				Name:    _const.DmsComponentName,
				Version: conf.Version,
			},
		},
	}
	for _, target := range targets {
		ret.Components = append(ret.Components, ComponentNameWithVersion{
			Name:    target.Name,
			Version: target.Version,
		})
	}

	if basicConfig.Title != "" {
		ret.Title = basicConfig.Title
	}

	if len(basicConfig.Logo) > 0 {
		ret.LogoUrl = fmt.Sprintf("%s%s", dmsCommonV1.CurrentGroupVersion, PersonalizationUrl)
	}

	return ret, nil
}
