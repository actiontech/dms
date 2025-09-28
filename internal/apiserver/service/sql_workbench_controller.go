package service

import (
	"fmt"
	sql_workbench "github.com/actiontech/dms/internal/sql_workbench/service"

	"github.com/actiontech/dms/internal/apiserver/conf"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
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
	reply, err := cc.CloudbeaverService.GetCloudbeaverConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	getSQLQueryConfigurationReply, err := cc.SqlWorkbenchService.GetSqlWorkbenchConfiguration()
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	if getSQLQueryConfigurationReply.Data.EnableSQLQuery {
		return NewOkRespWithReply(c, getSQLQueryConfigurationReply)
	}
	return NewOkRespWithReply(c, reply)
}