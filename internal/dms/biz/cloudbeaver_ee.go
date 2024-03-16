//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
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
	// the dbaccount relation auth purpose
	AuthPurpose string `json:"auth_purpose"`
	// the dbaccount relation auth used by sql workbench
	UsedBySQLWorkbench bool `json:"used_by_workbench"`
}

type ListDBAccountReply struct {
	Data []*TempDBAccount `json:"data"`
	// Generic reply
	base.GenericResp
}

func (cu *CloudbeaverUsecase) ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService, userId string) ([]*DBService, error) {
	proxyTarget, err := cu.proxyTargetRepo.GetProxyTargetByName(ctx, "provision")
	if errors.Is(err, pkgErr.ErrStorageNoData) {
		return activeDBServices, nil
	}
	if err != nil {
		return nil, err
	}
	dbaccounts, err := cu.ListAuthDbAccount(ctx, proxyTarget.URL.String(), userId)
	if err != nil {
		return nil, err
	}

	ret := make([]*DBService, 0)
	for _, dbaccount := range dbaccounts {
		if !dbaccount.UsedBySQLWorkbench || dbaccount.AuthPurpose == "" {
			continue
		}

		for _, activeDBService := range activeDBServices {
			if activeDBService.DBType != constant.DBTypeMySQL.String() {
				ret = append(ret, activeDBService)
			} else {
				if dbaccount.DbServiceUid == activeDBService.UID {
					db := *activeDBService
					db.User = dbaccount.User
					db.Password = dbaccount.Password
					db.AccountPurpose = dbaccount.AuthPurpose
					ret = append(ret, &db)
					break
				}
			}
		}
	}

	return ret, nil
}

func (cu *CloudbeaverUsecase) ListAuthDbAccount(ctx context.Context, url, userId string) ([]*TempDBAccount, error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reply := &ListDBAccountReply{}

	if err := pkgHttp.Get(ctx, fmt.Sprintf("%v/provision/v1/auth/dbaccounts?used_by_sql_workbench=true&page_size=999&page_index=1&owner_user_id=%s", url, userId), header, nil, reply); err != nil {
		return nil, fmt.Errorf("failed to get db account from %v: %v", url, err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("http reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}
