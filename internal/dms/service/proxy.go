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
	scenairo, err := convertProxyScenario(req.DMSProxyTarget.Scenario)
	if err != nil {
		return err
	}
	if err := d.DmsProxyUsecase.RegisterDMSProxyTarget(ctx, currentUserUid, biz.RegisterDMSProxyTargetArgs{
		Name:            req.DMSProxyTarget.Name,
		Addr:            req.DMSProxyTarget.Addr,
		Version:         req.DMSProxyTarget.Version,
		ProxyUrlPrefixs: req.DMSProxyTarget.ProxyUrlPrefixs,
		Scenario:        scenairo,
	}); err != nil {
		return fmt.Errorf("register dms proxy target failed: %v", err)
	}
	return nil
}

func convertProxyScenario(scenario dmsV1.ProxyScenario) (biz.ProxyScenario, error) {
	switch scenario {
	case dmsV1.ProxyScenarioInternalService:
		return biz.ProxyScenarioInternalService, nil
	case dmsV1.ProxyScenarioThirdPartyIntegrate:
		return biz.ProxyScenarioThirdPartyIntegrate, nil
	case "":
		// 兼容旧版本SQLE
		return biz.ProxyScenarioInternalService, nil
	default:
		return "", biz.ErrUnknownProxyScenario
	}
}
