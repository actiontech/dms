// Package docs dms api.
//
// Documentation of our dms API.
//
//	Schemes: http, https
//	BasePath: /
//	Version: 0.1.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- basic
//
//	SecurityDefinitions:
//	basic:
//	 type: apiKey
//	 in: header
//	 name: Authorization
//
// swagger:meta
package service

import (
	"fmt"
	"net/http"

	"github.com/actiontech/dms/internal/apiserver/conf"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	bV1 "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	"github.com/labstack/echo/v4"

	utilConf "github.com/actiontech/dms/pkg/dms-common/pkg/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type APIServer struct {
	DMSController         *DMSController
	CloudbeaverController *CloudbeaverController
	// more controllers

	echo   *echo.Echo
	opts   *conf.DMSOptions
	logger utilLog.Logger
}

func NewAPIServer(logger utilLog.Logger, opts *conf.DMSOptions) (*APIServer, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	return &APIServer{
		logger: logger,
		opts:   opts,
		echo:   e,
	}, nil
}

func (s *APIServer) RunHttpServer(logger utilLog.Logger) error {
	if err := s.installController(); nil != err {
		return fmt.Errorf("failed to install controller: %v", err)
	}

	if s.opts.EnableClusterMode {
		if s.opts.ServerId == "" {
			return fmt.Errorf("server id is required on cluster mode")
		}

		if s.opts.ReportHost == "" {
			return fmt.Errorf("report host is required on cluster mode")
		}

		s.DMSController.DMS.ClusterUsecase.SetClusterMode(true)
		if err := s.DMSController.DMS.ClusterUsecase.Join(s.opts.ServerId); err != nil {
			return err
		}

		defer s.DMSController.DMS.ClusterUsecase.Leave()
	}

	if err := s.installMiddleware(); nil != err {
		return fmt.Errorf("failed to install middleware: %v", err)
	}
	if err := s.initRouter(); nil != err {
		return fmt.Errorf("failed to init router: %v", err)
	}
	if s.opts.APIServiceOpts.EnableHttps {
		if s.opts.APIServiceOpts.CertFilePath == "" || s.opts.APIServiceOpts.KeyFilePath == "" {
			return fmt.Errorf("cert file path and key file path are required on https mode")
		}
		if err := s.echo.StartTLS(
			s.opts.GetAPIServer().GetHTTPAddr(),
			s.opts.APIServiceOpts.CertFilePath,
			s.opts.APIServiceOpts.KeyFilePath,
		); err != nil {
			if err != http.ErrServerClosed {
				return fmt.Errorf("failed to run https server: %v", err)
			}
		}
	} else {
		if err := s.echo.Start(s.opts.GetAPIServer().GetHTTPAddr()); nil != err {
			if err != http.ErrServerClosed {
				return fmt.Errorf("failed to run http server: %v", err)
			}
		}
	}
	return nil
}

func NewErrResp(c echo.Context, err error, code apiError.ErrorCode) error {
	return c.JSON(http.StatusOK, bV1.GenericResp{Code: int(code), Message: err.Error()})
}

func NewOkRespWithReply(c echo.Context, reply bV1.GenericResper) error {
	reply.SetCode(int(apiError.StatusOK))
	return c.JSON(http.StatusOK, reply)
}

func NewOkResp(c echo.Context) error {
	return c.JSON(http.StatusOK, bV1.GenericResp{Code: int(apiError.StatusOK), Message: "OK"})
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
