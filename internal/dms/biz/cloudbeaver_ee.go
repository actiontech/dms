//go:build enterprise

package biz

import (
	"context"
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
)

// A provision DBAccount
type TempDBAccount struct {
	// the dbaccount user
	User string `json:"user"`
	// the dbaccount password
	Password string `json:"password"`
	// the datasource's uid
	DbServiceUid string `json:"db_service_uid"`
}

type ListDBAccountReply struct {
	Data []*TempDBAccount `json:"data"`
	// Generic reply
	base.GenericResp
}

func (cu *CloudbeaverUsecase) ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService) ([]*DBService, error) {
	dbaccounts, err := cu.ListAuthDbAccount(ctx)
	if err != nil {
		return nil, err
	}
	ret := make([]*DBService, 0)
	for _, activeDBService := range activeDBServices {
		for _, dbaccount := range dbaccounts {
			// use db account instead of admin account
			if dbaccount.DbServiceUid == activeDBService.UID {
				activeDBService.User = dbaccount.User
				activeDBService.Password = dbaccount.Password
				ret = append(ret, activeDBService)
			}
		}
	}

	return ret, nil
}

func (cu *CloudbeaverUsecase) ListAuthDbAccount(ctx context.Context) ([]*TempDBAccount, error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	proxyTarget, err := cu.proxyTargetRepo.GetProxyTargetByName(ctx, "provision")
	if err != nil {
		return nil, err
	}

	reply := &ListDBAccountReply{}
	url := fmt.Sprintf("%v/%v", proxyTarget.URL.String(), "provision/v1/auth/dbaccounts?page_size=999&page_index=1")

	if err := pkgHttp.Get(ctx, url, header, nil, reply); err != nil {
		return nil, fmt.Errorf("failed to get db account from %v: %v", url, err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("http reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}
