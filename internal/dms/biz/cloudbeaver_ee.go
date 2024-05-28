//go:build enterprise

package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/go-openapi/strfmt"
	"github.com/labstack/echo/v4"
)

// A provision DBAccount
type TempDBAccount struct {
	DBAccountUid string         `json:"db_account_uid"`
	AccountInfo  AccountInfo    `json:"account_info"`
	Explanation  string         `json:"explanation"`
	ExpiredTime  string         `json:"expired_time"`
	DbService    v1.UidWithName `json:"db_service"`
}

type AccountInfo struct {
	User     string `json:"user"`
	Hostname string `json:"hostname"`
	Password string `json:"password"`
}

type ListDBAccountReply struct {
	Data []*TempDBAccount `json:"data"`
	// Generic reply
	base.GenericResp
}

func (cu *CloudbeaverUsecase) SupportDBType(dbType pkgConst.DBType) bool {
	return dbType == constant.DBTypeMySQL || dbType == constant.DBTypeOracle
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
		// prov不支持的数据库类型 使用管理员账号密码连接
		if !cu.SupportDBType(pkgConst.DBType(activeDBService.DBType)) {
			ret = append(ret, activeDBService)
			continue
		}

		for _, dbaccount := range dbaccounts {

			if dbaccount.ExpiredTime != "" {
				expiredTime, err := time.Parse(strfmt.RFC3339Millis, dbaccount.ExpiredTime)
				if err != nil {
					cu.log.Errorf("failed to parse expired time %v: %v", dbaccount.ExpiredTime, err)
					continue
				}
				if expiredTime.Unix() <= time.Now().Unix() {
					continue
				}
			}

			if dbaccount.DbService.Uid == activeDBService.UID {
				db := *activeDBService
				db.User = dbaccount.AccountInfo.User
				db.Password = dbaccount.AccountInfo.Password
				db.AccountPurpose = dbaccount.AccountInfo.User
				ret = append(ret, &db)
				break
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

	if err := pkgHttp.Get(ctx, fmt.Sprintf("%v/provision/v1/auth/projects//db_accounts?page_size=999&page_index=1&filter_by_password_managed=true&filter_by_status=unlock&filter_by_user=%s", url, userId), header, nil, reply); err != nil {
		return nil, fmt.Errorf("failed to get db account from %v: %v", url, err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("http reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}

func (cu *CloudbeaverUsecase) UpdateCbOp(params *graphql.RawParams, ctx context.Context, resp cloudbeaver.AuditResults) {
	value, loaded := taskIDAssocUid.LoadAndDelete(params.Variables["taskId"])
	if loaded {
		uid, ok := value.(string)
		if ok {
			operationLog, err := cu.cbOperationLogUsecase.GetCbOperationLogByID(ctx, uid)
			if err == nil {
				operationLog.AuditResults = convertToAuditResults(resp.Results)
				operationLog.IsAuditPass = &resp.IsSuccess
				err = cu.cbOperationLogUsecase.UpdateCbOperationLog(ctx, operationLog)
				if err != nil {
					cu.log.Error(err)
				}
			}
		}
	}
}

func (cu *CloudbeaverUsecase) UpdateCbOpResult(c echo.Context, cloudbeaverResBuf *bytes.Buffer, params *graphql.RawParams, ctx context.Context) error {
	resp := &struct {
		Data struct {
			Result *model.SQLExecuteInfo `json:"result"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(cloudbeaverResBuf.Bytes(), &resp); err != nil {
		cu.log.Errorf("extract task id err: %v", err)
		return fmt.Errorf("extract task id err: %v", err)
	}

	if resp.Data.Result != nil && resp.Data.Result.Results != nil {
		value, loaded := taskIDAssocUid.LoadAndDelete(params.Variables["taskId"])
		if loaded {
			uid, ok := value.(string)
			if ok {
				operationLog, err := cu.cbOperationLogUsecase.GetCbOperationLogByID(ctx, uid)
				if err == nil {
					executeInfo := resp.Data.Result
					operationLog.ExecTotalSec = int64(executeInfo.Duration)
					operationLog.ExecResult = *executeInfo.StatusMessage
					if executeInfo.Results != nil && len(executeInfo.Results) > 0 && executeInfo.Results[0].ResultSet != nil {
						operationLog.ResultSetRowCount = int64(len(executeInfo.Results[0].ResultSet.Rows))
					} else {
						operationLog.ResultSetRowCount = 0
					}
					err = cu.cbOperationLogUsecase.UpdateCbOperationLog(c.Request().Context(), operationLog)
					if err != nil {
						cu.log.Error(err)
					}
				}
			}
		}
	}

	return nil
}

func (cu *CloudbeaverUsecase) SaveCbOpLog(c echo.Context, dbService *DBService, params *graphql.RawParams, next echo.HandlerFunc) error {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		cu.log.Error(err)
		return err
	}
	cbOperationLog := newCbOperationLog(c, uid, dbService, params)

	cloudbeaverResBuf := new(bytes.Buffer)
	mw := io.MultiWriter(c.Response().Writer, cloudbeaverResBuf)
	writer := &cloudbeaverResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
	c.Response().Writer = writer

	if err = next(c); err != nil {
		return err
	}

	var taskInfo TaskInfo
	if err := json.Unmarshal(cloudbeaverResBuf.Bytes(), &taskInfo); err != nil {
		cu.log.Errorf("extract task id err: %v", err)
		return fmt.Errorf("extract task id err: %v", err)
	}

	err = cu.cbOperationLogUsecase.SaveCbOperationLog(c.Request().Context(), &cbOperationLog)
	if err != nil {
		cu.log.Error(err)
	} else if taskInfo.Data.TaskInfo != nil {
		taskID := taskInfo.Data.TaskInfo.ID
		taskIDAssocUid.Store(taskID, cbOperationLog.UID)
	}

	return nil
}

func newCbOperationLog(c echo.Context, uid string, dbService *DBService, params *graphql.RawParams) CbOperationLog {
	var cbOperationLog CbOperationLog
	cbOperationLog.UID = uid
	cbOperationLog.OpPersonUID = c.Get(dmsUserIdKey).(string)
	now := time.Now()
	cbOperationLog.OpTime = &now
	cbOperationLog.DBServiceUID = dbService.UID
	cbOperationLog.ProjectID = dbService.ProjectUID
	cbOperationLog.OpType = CbOperationLogTypeSql
	sessionID, ok := params.Variables["connectionId"]
	if ok {
		opSessionID := sessionID.(string)
		cbOperationLog.OpSessionID = &opSessionID
	}
	query, ok := params.Variables["query"]
	if ok {
		query := query.(string)
		cbOperationLog.OpDetail = query
	}
	cbOperationLog.OpHost = c.RealIP()

	return cbOperationLog
}

func convertToAuditResults(results []cloudbeaver.AuditSQLResV2) []*AuditResult {
	var auditResults []*AuditResult
	for _, res := range results {
		for _, result := range res.AuditResult {
			auditResult := &AuditResult{
				Level:    result.Level,
				Message:  result.Message,
				RuleName: result.RuleName,
			}

			auditResults = append(auditResults, auditResult)
		}
	}

	return auditResults
}
