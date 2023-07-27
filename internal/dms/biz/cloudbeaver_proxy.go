package biz

import (
	"fmt"
	"net/url"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4/middleware"
)

type CloudbeaverProxyUsecase struct {
	log                *utilLog.Helper
	cloudbeaverUsecase *CloudbeaverUsecase
}

func NewCloudbeaverProxyUsecase(logger utilLog.Logger, cloudbeaverUsecase *CloudbeaverUsecase) *CloudbeaverProxyUsecase {
	return &CloudbeaverProxyUsecase{
		log:                utilLog.NewHelper(logger, utilLog.WithMessageKey("cloudbeaver.proxy.service")),
		cloudbeaverUsecase: cloudbeaverUsecase,
	}
}

func (c *CloudbeaverProxyUsecase) GetCloudbeaverProxyTarget() ([]*middleware.ProxyTarget, error) {
	cfg := c.cloudbeaverUsecase.cloudbeaverCfg
	protocol := "http"
	if cfg.EnableHttps {
		protocol = "https"
	}

	rawUrl, err := url.Parse(fmt.Sprintf("%v://%v:%v", protocol, cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}

	return []*middleware.ProxyTarget{
		{
			URL: rawUrl,
		},
	}, nil
}
