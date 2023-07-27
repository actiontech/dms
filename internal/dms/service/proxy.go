package service

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) RegisterDMSProxyTarget(ctx context.Context, currentUserUid string, req *dmsV1.RegisterDMSProxyTargetReq) (err error) {
	d.log.Infof("RegisterDMSProxyTarget.req=%v", req)
	defer func() {
		d.log.Infof("RegisterDMSProxyTarget.req=%v;error=%v", req, err)
	}()

	if err := d.DmsProxyUsecase.RegisterDMSProxyTarget(ctx, currentUserUid, biz.RegisterDMSProxyTargetArgs{
		Name:            req.DMSProxyTarget.Name,
		Addr:            req.DMSProxyTarget.Addr,
		ProxyUrlPrefixs: req.DMSProxyTarget.ProxyUrlPrefixs,
	}); err != nil {
		return fmt.Errorf("register dms proxy target failed: %v", err)
	}

	return nil
}
