package service

import (
	"fmt"

	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/dms/service"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OdcController struct {
	OdcService *service.OdcService

	shutdownCallback func() error
}

func NewOdcController(logger utilLog.Logger, opts *conf.DMSOptions) (*OdcController, error) {
	odcService, err := service.NewAndInitOdcService(logger, opts)
	if nil != err {
		return nil, fmt.Errorf("failed to init odc service: %v", err)
	}

	return &OdcController{
		OdcService: odcService,
		shutdownCallback: func() error {
			return nil
		},
	}, nil
}

func (oc *OdcController) Shutdown() error {
	if nil != oc.shutdownCallback {
		return oc.shutdownCallback()
	}
	return nil
}
