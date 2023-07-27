// Package docs dms api.
//
// Documentation of our dms API.
//
//	 Schemes: http, https
//	 BasePath: /
//	 Version: 0.1.0
//
//	 Consumes:
//	 - application/json
//
//	 Produces:
//	 - application/json
//
//	 Security:
//	 - basic
//
//	SecurityDefinitions:
//	basic:
//	  type: basic
//
// swagger:meta
package service

import (
	"fmt"
	"net/http"

	bV1 "github.com/actiontech/dms/api/base/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"

	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

type APIServer struct {
	DMSController         *DMSController
	CloudbeaverController *CloudbeaverController
	// more controllers

	echo   *echo.Echo
	opts   *conf.Options
	logger utilLog.Logger
}

func NewAPIServer(logger utilLog.Logger, opts *conf.Options) (*APIServer, error) {
	return &APIServer{
		logger: logger,
		opts:   opts,
		echo:   echo.New(),
	}, nil
}

func (s *APIServer) RunHttpServer(logger utilLog.Logger) error {
	if err := s.installController(); nil != err {
		return fmt.Errorf("failed to install controller: %v", err)
	}
	if err := s.installMiddleware(); nil != err {
		return fmt.Errorf("failed to install middleware: %v", err)
	}
	if err := s.initRouter(); nil != err {
		return fmt.Errorf("failed to init router: %v", err)
	}

	if err := s.echo.Start(s.opts.GetAPIServer().GetHTTPAddr()); nil != err {
		if err != http.ErrServerClosed {
			return fmt.Errorf("failed to run http server: %v", err)
		}
	}
	return nil
}

func NewErrResp(c echo.Context, err error, code apiError.ErrorCode) error {
	return c.JSON(http.StatusOK, bV1.GenericResp{Code: int(code), Msg: err.Error()})
}

func NewOkRespWithReply(c echo.Context, reply bV1.GenericResper) error {
	reply.SetCode(int(apiError.StatusOK))
	return c.JSON(http.StatusOK, reply)
}

func NewOkResp(c echo.Context) error {
	return c.JSON(http.StatusOK, bV1.GenericResp{Code: int(apiError.StatusOK), Msg: "OK"})
}

func bindAndValidateReq(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return fmt.Errorf("failed to bind request: %v", err)
	}

	if err := utilConf.Validate(i); err != nil {
		return fmt.Errorf("failed to validate request: %v", err)
	}
	return nil
}
