//go:build enterprise

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

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
			Field:    string(biz.CbOperationLogFieldOpPersonUID),
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
		OrderBy:      string(biz.CbOperationLogFieldOpTime),
	}

	logs, total, err := d.CbOperationLogUsecase.ListCbOperationLog(ctx, listOption, uid, req.FilterOperationPersonUID, req.ProjectUid)
	if nil != err {
		return nil, err
	}

	listOption.PageNumber = 0
	listOption.LimitPerPage = 999999999

	execSuccessOption := *listOption
	execSuccessOption.FilterBy = append(execSuccessOption.FilterBy, constant.FilterCondition{
		Field:    string(biz.CbOperationLogFieldExecResult),
		Operator: constant.FilterOperatorEqual,
		Value:    biz.CbExecOpSuccess,
	})
	execSuccessCount, err := d.CbOperationLogUsecase.CountOperationLogs(ctx, &execSuccessOption, uid, req.FilterOperationPersonUID, req.ProjectUid)
	if err != nil {
		return nil, err
	}

	auditFailedOption := *listOption
	auditFailedOption.FilterBy = append(auditFailedOption.FilterBy, constant.FilterCondition{
		Field:    string(biz.CbOperationLogFieldIsAuditPassed),
		Operator: constant.FilterOperatorEqual,
		Value:    "0",
	})
	auditFailedSqlCount, err := d.CbOperationLogUsecase.CountOperationLogs(ctx, &auditFailedOption, uid, req.FilterOperationPersonUID, req.ProjectUid)
	if err != nil {
		return nil, err
	}

	data := make([]*dmsV1.CBOperationLog, 0, len(logs))
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
	}

	return &dmsV1.ListCBOperationLogsReply{
		ExecSQLTotal: total,
		ExecSuccessRate: func() float64 {
			if execSuccessCount == 0 || total == 0 {
				return 0
			}
			return float64(execSuccessCount) / float64(total)
		}(),
		AuditInterceptedSQLCount: auditFailedSqlCount,
		ExecFailedSQLCount:       total - execSuccessCount,
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

func (d *DMSService) exportCbOperationLogs(ctx context.Context, req *dmsV1.ExportCBOperationLogsReq, uid string) ([]byte, error) {
	filterBy := make([]constant.FilterCondition, 0)
	if req.FilterOperationPersonUID != "" {
		filterBy = append(filterBy, constant.FilterCondition{
			Field:    string(biz.CbOperationLogFieldOpPersonUID),
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
		PageNumber:   0,
		LimitPerPage: 999999999,
		FilterBy:     filterBy,
		OrderBy:      string(biz.CbOperationLogFieldOpTime),
	}

	logs, total, err := d.CbOperationLogUsecase.ListCbOperationLog(ctx, listOption, uid, req.FilterOperationPersonUID, req.ProjectUid)
	if nil != err {
		return nil, err
	}

	buff := new(bytes.Buffer)
	buff.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	csvWriter := csv.NewWriter(buff)

	var cbOpList [][]string
	cbOpList = append(cbOpList, []string{
		"项目名",
		"操作人",
		"操作时间",
		"数据源",
		"操作详情",
		"会话ID",
		"操作IP",
		"审核结果",
		"执行结果",
		"执行时间(毫秒)",
		"结果集返回行数",
	})

	var auditFailedSqlCount, execSuccessCount int
	for _, log := range logs {
		cbOpList = append(cbOpList, []string{
			log.GetProjectName(),
			log.GetUserName(),
			log.GetOpTime().String(),
			log.GetDbServiceName(),
			log.OpDetail,
			log.GetSessionID(),
			log.OpHost,
			spliceAuditResults(log.AuditResults),
			log.ExecResult,
			strconv.FormatInt(log.ExecTotalSec, 10),
			strconv.FormatInt(log.ResultSetRowCount, 10),
		})

		if log.ExecResult == biz.CbExecOpSuccess {
			execSuccessCount++
		}
		if log.IsAuditPass != nil && !*log.IsAuditPass {
			auditFailedSqlCount++
		}
	}

	err = csvWriter.WriteAll([][]string{
		{"执行总量:", strconv.FormatInt(total, 10)},
		{"执行成功率:", fmt.Sprintf("%.2f%%",
			func() float64 {
				if execSuccessCount == 0 || total == 0 {
					return 0
				}
				return float64(execSuccessCount) / float64(total) * 100
			}())},
		{"审核拦截的异常SQL:", strconv.FormatInt(int64(auditFailedSqlCount), 10)},
		{"执行不成功的SQL:", strconv.FormatInt(total-int64(execSuccessCount), 10)},
	})
	if err != nil {
		return nil, err
	}

	if err := csvWriter.WriteAll(cbOpList); err != nil {
		return nil, err
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func spliceAuditResults(auditResults []*biz.AuditResult) string {
	var results []string
	for _, auditResult := range auditResults {
		results = append(results, fmt.Sprintf("[%v]%v", auditResult.Level, auditResult.Message))
	}
	return strings.Join(results, "\n")
}
