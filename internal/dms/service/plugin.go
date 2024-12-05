package service

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) RegisterDMSPlugin(ctx context.Context, currentUserUid string, req *dmsV1.RegisterDMSPluginReq) (err error) {
	d.log.Infof("RegisterDMSPlugin.req=%v", req)
	defer func() {
		d.log.Infof("RegisterDMSPlugin.req=%v;error=%v", req, err)
	}()

	if err := d.PluginUsecase.RegisterPlugin(ctx, &biz.Plugin{
		Name:                         req.Plugin.Name,
		OperateDataResourceHandleUrl: req.Plugin.OperateDataResourceHandleUrl,
	}, currentUserUid); err != nil {
		return fmt.Errorf("register dms plugin failed: %v", err)
	}

	return nil
}
