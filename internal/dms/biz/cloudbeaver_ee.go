//go:build enterprise

package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
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

	dbaccounts := make(map[string]*TempDBAccount, 0)
	projectTips := make(map[string]struct{}, 0)
	for _, db := range activeDBServices {
		if _, ok := projectTips[db.ProjectUID]; ok {
			continue
		}
		projectTips[db.ProjectUID] = struct{}{}
		accounts, err := cu.ListAuthDbAccount(ctx, proxyTarget.URL.String(), db.ProjectUID, userId)
		if err != nil {
			return nil, err
		}
		for _, account := range accounts {
			dbaccounts[account.DBAccountUid] = account
		}

	}

	ret := make([]*DBService, 0)
	for _, activeDBService := range activeDBServices {
		// prov不支持的数据库类型 使用数据源账号密码连接
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

func (cu *CloudbeaverUsecase) ListAuthDbAccount(ctx context.Context, url, projectUid, userId string) ([]*TempDBAccount, error) {
	// gen token with claims
	token, err := jwt.GenJwtToken(jwt.WithUserId(userId))
	if nil != err {
		return nil, err
	}
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	reply := &ListDBAccountReply{}

	if err := pkgHttp.Get(ctx, fmt.Sprintf("%v/provision/v1/auth/projects/%s/db_accounts?page_size=999&page_index=1&filter_by_password_managed=true&filter_by_status=unlock", url, projectUid), header, nil, reply); err != nil {
		return nil, fmt.Errorf("failed to get db account from %v: %v", url, err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("http reply code(%v) error: %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}

func (cu *CloudbeaverUsecase) SaveUiOp(c echo.Context, buf *bytes.Buffer, params *graphql.RawParams) error {
	resp := &struct {
		Data struct {
			Result *model.SQLExecuteInfo `json:"result"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		return err
	}

	dbService, err := cu.getDbService(c.Request().Context(), params)
	if err != nil {
		return err
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return err
	}

	operationLog, err := newCbOperationLog(c, uid, dbService, params, "")
	if err != nil {
		return err
	}

	result := resp.Data.Result
	if result != nil {
		operationLog.ExecResult = CbExecOpSuccess
		operationLog.ExecTotalSec = int64(result.Duration)
		results := result.Results
		if results != nil && len(results) > 0 {
			operationLog.ResultSetRowCount = int64(*results[0].UpdateRowCount)
		}
	} else {
		marshal, err := json.Marshal(buf.String())
		if err != nil {
			return err
		}
		operationLog.ExecResult = fmt.Sprintf("执行失败: %s", string(marshal))
	}

	err = cu.cbOperationLogUsecase.SaveCbOperationLog(c.Request().Context(), &operationLog)
	if err != nil {
		return err
	}

	return nil
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

func (cu *CloudbeaverUsecase) SaveCbOpLog(c echo.Context, dbService *DBService, params *graphql.RawParams, resp cloudbeaver.AuditResults, cloudbeaverResBuf *bytes.Buffer) error {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		cu.log.Error(err)
		return nil
	}
	cbOperationLog, err := newCbOperationLog(c, uid, dbService, params, CbOperationLogTypeSql)
	if err != nil {
		cu.log.Error(err)
		return nil
	}

	cbOperationLog.AuditResults = convertToAuditResults(resp.Results)
	cbOperationLog.IsAuditPass = &resp.IsSuccess

	var taskInfo TaskInfo
	if err := json.Unmarshal(cloudbeaverResBuf.Bytes(), &taskInfo); err != nil {
		cu.log.Errorf("extract task id err: %v", err)
		return nil
	}

	err = cu.cbOperationLogUsecase.SaveCbOperationLog(c.Request().Context(), &cbOperationLog)
	if err != nil {
		cu.log.Error(err)
		return nil
	} else if taskInfo.Data.TaskInfo != nil {
		taskID := &taskInfo.Data.TaskInfo.ID
		taskIDAssocUid.Store(*taskID, cbOperationLog.UID)
		return nil
	}

	return nil
}

func (cu *CloudbeaverUsecase) SaveCbOperationLog(c echo.Context, dbService *DBService, params *graphql.RawParams, resp cloudbeaver.AuditResults) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		cu.log.Error(err)
		return
	}
	cbOperationLog, err := newCbOperationLog(c, uid, dbService, params, CbOperationLogTypeSql)
	if err != nil {
		cu.log.Error(err)
		return
	}

	cbOperationLog.AuditResults = convertToAuditResults(resp.Results)
	cbOperationLog.IsAuditPass = &resp.IsSuccess

	err = cu.cbOperationLogUsecase.SaveCbOperationLog(c.Request().Context(), &cbOperationLog)
	if err != nil {
		cu.log.Error(err)
		return
	}

	return
}

func newCbOperationLog(c echo.Context, uid string, dbService *DBService, params *graphql.RawParams, opType CbOperationLogType) (CbOperationLog, error) {
	var cbOperationLog CbOperationLog
	cbOperationLog.UID = uid
	cbOperationLog.OpPersonUID = c.Get(dmsUserIdKey).(string)
	now := time.Now()
	cbOperationLog.OpTime = &now
	cbOperationLog.DBServiceUID = dbService.UID
	cbOperationLog.ProjectID = dbService.ProjectUID
	cbOperationLog.OpType = opType
	sessionID, ok := params.Variables["connectionId"]
	if ok {
		opSessionID := sessionID.(string)
		cbOperationLog.OpSessionID = &opSessionID
	}
	var opDetailReq string
	query, ok := params.Variables["query"]
	if ok {
		query := query.(string)
		opDetailReq = query
	}
	cbOperationLog.OpHost = c.RealIP()

	opDetail, ok := params.Variables["deletedRows"]
	if ok {
		marshal, err := json.Marshal(opDetail)
		if err != nil {
			return CbOperationLog{}, err
		}
		opDetailReq = fmt.Sprintf("在数据源:%s中删除了以下数据:%s", dbService.Name, string(marshal))
	}
	opDetail, ok = params.Variables["addedRows"]
	if ok {
		marshal, err := json.Marshal(opDetail)
		if err != nil {
			return CbOperationLog{}, err
		}
		opDetailReq = fmt.Sprintf("在数据源:%s中添加了以下数据:%s", dbService.Name, string(marshal))
	}
	opDetail, ok = params.Variables["updatedRows"]
	if ok {
		marshal, err := json.Marshal(opDetail)
		if err != nil {
			return CbOperationLog{}, err
		}
		opDetailReq = fmt.Sprintf("在数据源:%s中更新了以下数据:%s", dbService.Name, string(marshal))
	}
	cbOperationLog.OpDetail = opDetailReq

	return cbOperationLog, nil
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

func (cu *CloudbeaverUsecase) SaveCbLogSqlAuditNotEnable(c echo.Context, dbService *DBService, params *graphql.RawParams, cloudbeaverResBuf *bytes.Buffer) error {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return err
	}

	cbOperationLog, err := newCbOperationLog(c, uid, dbService, params, CbOperationLogTypeSql)
	if err != nil {
		return err
	}

	var taskInfo TaskInfo
	if err := json.Unmarshal(cloudbeaverResBuf.Bytes(), &taskInfo); err != nil {
		return err
	}

	err = cu.cbOperationLogUsecase.SaveCbOperationLog(c.Request().Context(), &cbOperationLog)
	if err != nil {
		return err
	} else if taskInfo.Data.TaskInfo != nil {
		taskID := &taskInfo.Data.TaskInfo.ID
		taskIDAssocUid.Store(*taskID, cbOperationLog.UID)
	}

	return nil
}
