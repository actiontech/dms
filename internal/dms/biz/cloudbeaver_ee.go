//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"

	maskBiz "github.com/actiontech/dms/internal/data_masking/biz"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
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
	for _, activeDBService := range activeDBServices {
		if activeDBService.DBType == constant.DBTypeMySQL.String() {
			for _, dbaccount := range dbaccounts {
				// use db account instead of admin account
				if dbaccount.DbServiceUid == activeDBService.UID {
					activeDBService.User = dbaccount.User
					activeDBService.Password = dbaccount.Password
					ret = append(ret, activeDBService)
				}
			}
		} else {
			ret = append(ret, activeDBService)
		}
	}

	return ret, nil
}

func (cu *CloudbeaverUsecase) ListAuthDbAccount(ctx context.Context, url, userId string) ([]*TempDBAccount, error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reply := &ListDBAccountReply{}

	if err := pkgHttp.Get(ctx, fmt.Sprintf("%v/provision/v1/auth/dbaccounts?page_size=999&page_index=1&owner_user_id=%s", url, userId), header, nil, reply); err != nil {
		return nil, fmt.Errorf("failed to get db account from %v: %v", url, err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("http reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}

// SQLExecuteResultsDataMasking 为DMS企业版的脱敏功能，捕获cloudbeaver返回的结果集，根据配置对结果集脱敏
func (cu *CloudbeaverUsecase) SQLExecuteResultsDataMasking(ctx context.Context, result *model.SQLExecuteInfo) error {
	for _, r := range result.Results {
		if r.ResultSet == nil {
			continue
		}
		c := make([]*maskBiz.SqlResultColumn, len(r.ResultSet.Columns))
		for i := range r.ResultSet.Columns {
			c[i] = &maskBiz.SqlResultColumn{
				Name: *r.ResultSet.Columns[i].Name,
			}
		}

		params := maskBiz.NewMaskSqlExecuteResultParams(c)
		for i := range r.ResultSet.Rows {
			params.AddRows(r.ResultSet.Rows[i])
		}

		if err := cu.dataMaskingUseCase.MaskSqlExecuteResultByAutoDetection(params); nil != err {
			return err
		}
	}
	return nil
}
