//go:build enterprise

package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *DMSService) listCBOperationLogs(ctx context.Context, req *dmsV1.ListCBOperationLogsReq, uid string) (reply *dmsV1.ListCBOperationLogsReply, err error) {
	d.log.Infof("ListCbOperationLogs.req=%v", req)
	defer func() {
		d.log.Infof("ListCbOperationLogs.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	filterBy := make([]constant.FilterCondition, 0)
	if req.FilterOperationPersonUID != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldUID),
			Operator: constant.FilterOperatorEqual,
			Value:    req.FilterOperationPersonUID,
		})
	}

	if req.FilterOperationTimeFrom != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldOpTime),
			Operator: constant.FilterOperatorGreaterThanOrEqual,
			Value:    req.FilterOperationTimeFrom,
		})
	}

	if req.FilterOperationTimeTo != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldOpTime),
			Operator: constant.FilterOperatorLessThanOrEqual,
			Value:    req.FilterOperationTimeTo,
		})
	}

	if req.FilterDBServiceUID != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldDBServiceUID),
			Operator: constant.FilterOperatorEqual,
			Value:    req.FilterDBServiceUID,
		})
	}

	if req.FilterExecResult != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldExecResult),
			Operator: constant.FilterOperatorEqual,
			Value:    req.FilterExecResult,
		})
	}

	if req.ProjectUid != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldProjectID),
			Operator: constant.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	if req.FuzzyKeyword != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:         string(biz.CbOperationLogFieldExecResult),
			Operator:      constant.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
		})
		filterBy = append(filterBy, constant.FilterCondition{
			Field:         string(biz.CbOperationLogFieldOpHost),
			Operator:      constant.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
		})
	}

	listOption := &biz.ListCbOperationLogOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		FilterBy:     filterBy,
	}

	logs, total, err := d.CbOperationLogUsecase.ListCbOperationLog(ctx, listOption, uid, req.FilterOperationPersonUID, req.ProjectUid)
	if nil != err {
		return nil, err
	}

	data := make([]*dmsV1.CBOperationLog, 0, len(logs))
	var auditFailedSqlCount, execSuccessCount int
	for _, log := range logs {
		dmsLog := &dmsV1.CBOperationLog{
			UID:           log.UID,
			OperationTime: log.GetOpTime(),
			Operation: dmsV1.Operation{
				OperationType:   dmsV1.CbOperationType(log.OpType),
				OperationDetail: log.OpDetail,
			},
			SessionID:         log.GetSessionID(),
			OperationIp:       log.OpHost,
			ExecResult:        log.ExecResult,
			ExecTimeSecond:    int(log.ExecTotalSec),
			ResultSetRowCount: log.ResultSetRowCount,
		}
		if log.User != nil {
			dmsLog.OperationPerson = dmsV1.UidWithName{Uid: log.User.UID, Name: log.User.Name}
		}
		if log.DbService != nil {
			dmsLog.DBService = dmsV1.UidWithDBServiceName{Uid: log.DbService.UID, Name: log.DbService.Name}
		}
		if log.AuditResults != nil {
			dmsLog.AuditResult = make([]*dmsV1.AuditSQLResult, 0, len(log.AuditResults))
			for _, auditResult := range log.AuditResults {
				dmsLog.AuditResult = append(dmsLog.AuditResult, &dmsV1.AuditSQLResult{
					Level:    auditResult.Level,
					Message:  auditResult.Message,
					RuleName: auditResult.RuleName,
				})
			}
		}

		data = append(data, dmsLog)

		if log.ExecResult == "Success" {
			execSuccessCount++
		}
		if log.IsAuditPass != nil && !*log.IsAuditPass {
			auditFailedSqlCount++
		}
	}

	return &dmsV1.ListCBOperationLogsReply{
		ExecSQLTotal: total,
		ExecSuccessRate: func() float64 {
			if execSuccessCount == 0 || total == 0 {
				return 0
			}
			return float64(execSuccessCount) / float64(total)
		}(),
		AuditInterceptedSQLCount: int64(auditFailedSqlCount),
		ExecFailedSQLCount:       total - int64(execSuccessCount),
		Data:                     data,
		Total:                    total,
	}, nil
}

func (d *DMSService) getCBOperationLogTips(ctx context.Context, req *dmsV1.GetCBOperationLogTipsReq, currentUid string) (*dmsV1.GetCBOperationLogTipsReply, error) {
	filterBy := make([]constant.FilterCondition, 0)
	if req.ProjectUid != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldProjectID),
			Operator: constant.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	listOption := &biz.ListCbOperationLogOption{
		PageNumber:   0,
		LimitPerPage: 999999,
		FilterBy:     filterBy,
	}

	logs, _, err := d.CbOperationLogUsecase.ListCbOperationLog(ctx, listOption, currentUid, currentUid, req.ProjectUid)
	if nil != err {
		return nil, err
	}

	data := &dmsV1.GetCBOperationLogTipsReply{
		Data: &dmsV1.CBOperationLogTips{
			ExecResult: make([]string, 0, len(logs)),
		},
	}

	execResultMap := make(map[string]struct{})
	for _, log := range logs {
		_, ok := execResultMap[log.ExecResult]
		if log.ExecResult != "" && !ok {
			data.Data.ExecResult = append(data.Data.ExecResult, log.ExecResult)
			execResultMap[log.ExecResult] = struct{}{}
		}
	}

	return data, nil
}
