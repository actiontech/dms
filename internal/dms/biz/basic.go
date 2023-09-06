package biz

import (
	"context"

	"github.com/actiontech/dms/internal/dms/conf"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type BasicUsecase struct {
	log         *utilLog.Helper
	proxyTarget ProxyTargetRepo
}

func NewBasicInfoUsecase(log utilLog.Logger, proxyTarget ProxyTargetRepo) *BasicUsecase {
	return &BasicUsecase{
		log:         utilLog.NewHelper(log, utilLog.WithMessageKey("biz.basic")),
		proxyTarget: proxyTarget,
	}
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
	componentDMSName = "dms"
)

func (d *BasicUsecase) GetBasicInfo(ctx context.Context) (*BasicInfo, error) {
	targets, err := d.proxyTarget.ListProxyTargets(ctx)
	if err != nil {
		return nil, err
	}

	ret := &BasicInfo{
		Components: []ComponentNameWithVersion{
			{
				Name:    componentDMSName,
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

	return ret, nil
}
