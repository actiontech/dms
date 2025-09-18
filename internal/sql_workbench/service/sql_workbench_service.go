package sql_workbench

import (
	"fmt"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	config "github.com/actiontech/dms/internal/sql_workbench/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4/middleware"
	"net/url"
)

const SQL_WORKBENCH_URL = "/odc_query"

type SqlWorkbenchService struct {
	cfg          *config.SqlWorkbenchOpts
	log          *utilLog.Helper
}

func NewAndInitSqlWorkbenchService(logger utilLog.Logger, opts *conf.DMSOptions) (*SqlWorkbenchService, error) {
	return &SqlWorkbenchService{
		cfg: opts.SqlWorkBenchOpts,
		log: utilLog.NewHelper(logger, utilLog.WithMessageKey("sql_workbench.service")),
	}, nil
}

func (sqlWorkbenchService *SqlWorkbenchService) IsConfigured() bool {
	if sqlWorkbenchService.cfg == nil {
		return false
	}
	return sqlWorkbenchService.cfg != nil && sqlWorkbenchService.cfg.Host != "" && sqlWorkbenchService.cfg.Port != ""
}

func (sqlWorkbenchService *SqlWorkbenchService) GetSqlWorkbenchConfiguration() (reply *dmsV1.GetSQLQueryConfigurationReply, err error) {
	sqlWorkbenchService.log.Infof("GetSqlWorkbenchConfiguration")
	defer func() {
		sqlWorkbenchService.log.Infof("GetSqlWorkbenchConfiguration; reply=%v, error=%v", reply, err)
	}()

	return &dmsV1.GetSQLQueryConfigurationReply{
		Data: struct {
			EnableSQLQuery  bool   `json:"enable_sql_query"`
			SQLQueryRootURI string `json:"sql_query_root_uri"`
		}{
			EnableSQLQuery:  sqlWorkbenchService.IsConfigured(),
			SQLQueryRootURI: SQL_WORKBENCH_URL+"/",
		},
	}, nil
}

func (sqlWorkbenchService *SqlWorkbenchService) GetOdcProxyTarget() ([]*middleware.ProxyTarget, error) {
	cfg := sqlWorkbenchService.cfg
	protocol := "http"
	if cfg.EnableHttps {
		protocol = "https"
	}

	rawUrl, err := url.Parse(fmt.Sprintf("%v://%v:%v", protocol, cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}
	sqlWorkbenchService.log.Infof("ODC proxy target URL: %s", rawUrl.String())
	return []*middleware.ProxyTarget{
		{
			URL: rawUrl,
		},
	}, nil
}

func (sqlWorkbenchService *SqlWorkbenchService) GetRootUri() string {
	return SQL_WORKBENCH_URL
}