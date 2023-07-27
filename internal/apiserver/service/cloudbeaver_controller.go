package service

import (
	"fmt"

	"github.com/actiontech/dms/internal/apiserver/conf"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/internal/dms/service"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

type CloudbeaverController struct {
	CloudbeaverService *service.CloudbeaverService

	shutdownCallback func() error
}

func NewCloudbeaverController(logger utilLog.Logger, opts *conf.Options) (*CloudbeaverController, error) {
	cloudbeaverService, err := service.NewAndInitCloudbeaverService(logger, opts)
	if nil != err {
		return nil, fmt.Errorf("failed to init cloudbeaver service: %v", err)
	}

	return &CloudbeaverController{
		CloudbeaverService: cloudbeaverService,
		shutdownCallback: func() error {
			return nil
		},
	}, nil
}

func (cc *CloudbeaverController) Shutdown() error {
	if nil != cc.shutdownCallback {
		return cc.shutdownCallback()
	}
	return nil
}

// swagger:route GET /v1/dms/configurations/sql_query cloudbeaver GetSQLQueryConfiguration
//
// get sql_query configuration.
//
//	responses:
//	  200: body:GetSQLQueryConfigurationReply
//	  default: body:GenericResp
func (cc *CloudbeaverController) GetSQLQueryConfiguration(c echo.Context) error {
	reply, err := cc.CloudbeaverService.GetCloudbeaverConfiguration(c.Request().Context())
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	return NewOkRespWithReply(c, reply)
}
