package biz

import (
	"fmt"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4/middleware"
	"net/url"
)

// OdcProxyUsecase ODC代理业务逻辑
type OdcProxyUsecase struct {
	odcUsecase *OdcUsecase
	log        *utilLog.Helper
}

// NewOdcProxyUsecase 创建ODC代理业务逻辑实例
func NewOdcProxyUsecase(log utilLog.Logger, odcUsecase *OdcUsecase) *OdcProxyUsecase {
	return &OdcProxyUsecase{
		odcUsecase: odcUsecase,
		log:        utilLog.NewHelper(log, utilLog.WithMessageKey("biz.odc_proxy")),
	}
}

func (c *OdcProxyUsecase) GetOdcProxyTarget() ([]*middleware.ProxyTarget, error) {
	cfg := c.odcUsecase.odcCfg
	protocol := "http"
	if cfg.EnableHttps {
		protocol = "https"
	}

	rawUrl, err := url.Parse(fmt.Sprintf("%v://%v:%v", protocol, cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}
	c.log.Infof("ODC proxy target URL: %s", rawUrl.String())
	return []*middleware.ProxyTarget{
		{
			URL: rawUrl,
		},
	}, nil
}
