package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/dms/biz"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OdcService struct {
	OdcUsecase   *biz.OdcUsecase
	ProxyUsecase *biz.OdcProxyUsecase
	log          *utilLog.Helper
}

func NewAndInitOdcService(logger utilLog.Logger, opts *conf.DMSOptions) (*OdcService, error) {
	// 简化的ODC服务初始化，避免复杂的依赖问题
	var cfg *biz.OdcCfg
	if opts.OdcOpts != nil {
		cfg = &biz.OdcCfg{
			EnableHttps:   opts.OdcOpts.EnableHttps,
			Host:          opts.OdcOpts.Host,
			Port:          opts.OdcOpts.Port,
			AdminUser:     opts.OdcOpts.AdminUser,
			AdminPassword: opts.OdcOpts.AdminPassword,
			APIKey:        opts.OdcOpts.APIKey,
			ClientID:      opts.OdcOpts.ClientID,
			ClientSecret:  opts.OdcOpts.ClientSecret,
		}
	}

	// 创建简化的ODC业务逻辑实例
	odcUsecase := biz.NewOdcUsecase(logger, cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	proxyUsecase := biz.NewOdcProxyUsecase(logger, odcUsecase)

	return &OdcService{
		OdcUsecase:   odcUsecase,
		ProxyUsecase: proxyUsecase,
		log:          utilLog.NewHelper(logger, utilLog.WithMessageKey("odc.service")),
	}, nil
}

func (os *OdcService) GetOdcConfiguration(ctx context.Context) (reply *dmsV1.GetSQLQueryConfigurationReply, err error) {
	os.log.Infof("GetOdcConfiguration")
	defer func() {
		os.log.Infof("GetOdcConfiguration; reply=%v, error=%v", reply, err)
	}()

	return &dmsV1.GetSQLQueryConfigurationReply{
		Data: struct {
			EnableSQLQuery  bool   `json:"enable_sql_query"`
			SQLQueryRootURI string `json:"sql_query_root_uri"`
		}{
			EnableSQLQuery:  os.OdcUsecase.IsOdcConfigured(),
			SQLQueryRootURI: os.OdcUsecase.GetRootUri() + "/", // 确保URL以斜杠结尾，防止DMS开启HTTPS时，Web服务器重定向到HTTP根路径导致访问错误
		},
	}, nil
}

func (os *OdcService) Logout(session string) {
	os.OdcUsecase.UnbindOdcSession(session)
}
