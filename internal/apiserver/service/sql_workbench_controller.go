package service

import (
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	sql_workbench "github.com/actiontech/dms/internal/sql_workbench/service"

	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/dms/service"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

type SqlWorkbenchController struct {
	CloudbeaverService  *service.CloudbeaverService
	SqlWorkbenchService *sql_workbench.SqlWorkbenchService
	shutdownCallback    func() error
}

func NewSqlWorkbenchController(logger utilLog.Logger, opts *conf.DMSOptions) (*SqlWorkbenchController, error) {
	cloudbeaverService, err := service.NewAndInitCloudbeaverService(logger, opts)
	if nil != err {
		return nil, fmt.Errorf("failed to init cloudbeaver service: %v", err)
	}
	sqlWorkbenchService, err := sql_workbench.NewAndInitSqlWorkbenchService(logger, opts)
	if nil != err {
		return nil, fmt.Errorf("failed to init sql workbench service: %v", err)
	}
	return &SqlWorkbenchController{
		CloudbeaverService:  cloudbeaverService,
		SqlWorkbenchService: sqlWorkbenchService,
		shutdownCallback: func() error {
			return nil
		},
	}, nil
}

func (cc *SqlWorkbenchController) Shutdown() error {
	if nil != cc.shutdownCallback {
		return cc.shutdownCallback()
	}
	return nil
}

// swagger:route GET /v1/dms/configurations/sql_query CloudBeaver GetSQLQueryConfiguration
//
// get sql_query configuration.
//
//	responses:
//	  200: body:GetSQLQueryConfigurationReply
//	  default: body:GenericResp
func (cc *SqlWorkbenchController) GetSQLQueryConfiguration(c echo.Context) error {
	reply := &dmsV1.GetSQLQueryConfigurationReply{
		Data: struct {
			EnableSQLQuery  bool   `json:"enable_sql_query"`
			SQLQueryRootURI string `json:"sql_query_root_uri"`
			EnableOdcQuery  bool   `json:"enable_odc_query"`
			OdcQueryRootURI string `json:"odc_query_root_uri"`
		}{
			EnableSQLQuery:  cc.CloudbeaverService.CloudbeaverUsecase.IsCloudbeaverConfigured(),
			SQLQueryRootURI: cc.CloudbeaverService.CloudbeaverUsecase.GetRootUri() + "/", // 确保URL以斜杠结尾，防止DMS开启HTTPS时，Web服务器重定向到HTTP根路径导致访问错误
			EnableOdcQuery:  cc.SqlWorkbenchService.IsConfigured(),
			OdcQueryRootURI: cc.SqlWorkbenchService.GetRootUri(),
		},
	}
	return NewOkRespWithReply(c, reply)
}
