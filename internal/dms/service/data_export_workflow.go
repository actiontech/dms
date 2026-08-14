package service

import (
	"context"
	"fmt"
	"strings"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/locale"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) AddDataExportWorkflow(ctx context.Context, req *dmsV1.AddDataExportWorkflowReq, currentUserUid string) (reply *dmsV1.AddDataExportWorkflowReply, err error) {
	// generate biz args
	tasks := make([]biz.Task, 0)
	for _, t := range req.DataExportWorkflow.Tasks {
		tasks = append(tasks, biz.Task{UID: t.Uid})
	}
	args := &biz.Workflow{
		Name:               req.DataExportWorkflow.Name,
		Desc:               req.DataExportWorkflow.Desc,
		Tasks:              tasks,
		ProjectUID:         req.ProjectUid,
		WorkflowTemplateId: req.DataExportWorkflow.WorkflowTemplateId,
		OpsTypeUID:         req.DataExportWorkflow.OpsTypeUID,
	}
	uid, err := d.DataExportWorkflowUsecase.AddDataExportWorkflow(ctx, currentUserUid, args)
	if err != nil {
		return nil, fmt.Errorf("add data export workflow failed: %w", err)
	}

	return &dmsV1.AddDataExportWorkflowReply{
		Data: struct {
			Uid string `json:"export_data_workflow_uid"`
		}{Uid: uid}}, nil
}

func (d *DMSService) ListDataExportWorkflow(ctx context.Context, req *dmsV1.ListDataExportWorkflowsReq, currentUserUid string) (reply *dmsV1.ListDataExportWorkflowsReply, err error) {
	// default order by
	orderBy := biz.WorkflowFieldCreateTime

	filterByOptions := pkgConst.NewFilterOptions(pkgConst.FilterLogicAnd)

	andConditions := make([]pkgConst.FilterCondition, 0)
	andConditions = append(andConditions, pkgConst.FilterCondition{
		Field:    string(biz.WorkflowFieldWorkflowType),
		Operator: pkgConst.FilterOperatorEqual,
		Value:    biz.DataExportWorkflowEventType.String(),
	})

	if req.FilterByCreateUserUid != "" {
		andConditions = append(andConditions, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldCreateUserUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByCreateUserUid,
		})
	}

	if req.FilterCreateTimeFrom != "" {
		andConditions = append(andConditions, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldCreateTime),
			Operator: pkgConst.FilterOperatorGreaterThanOrEqual,
			Value:    req.FilterCreateTimeFrom,
		})
	}

	if req.FilterCreateTimeTo != "" {
		andConditions = append(andConditions, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldCreateTime),
			Operator: pkgConst.FilterOperatorLessThanOrEqual,
			Value:    req.FilterCreateTimeTo,
		})
	}

	if req.ProjectUid != "" {
		andConditions = append(andConditions, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	if req.FilterWorkflowTemplateId != 0 {
		andConditions = append(andConditions, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldWorkflowTemplateId),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterWorkflowTemplateId,
		})
	}

	if req.FilterByOpsTypeUid != "" {
		andConditions = append(andConditions, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldOpsTypeUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByOpsTypeUid,
		})
	}

	if len(andConditions) > 0 {
		filterByOptions.Groups = append(filterByOptions.Groups, pkgConst.NewConditionGroup(pkgConst.FilterLogicAnd, andConditions...))
	}

	if req.FuzzyKeyword != "" {
		filterByOptions.Groups = append(filterByOptions.Groups, pkgConst.NewConditionGroup(
			pkgConst.FilterLogicOr,
			pkgConst.FilterCondition{
				Field:    string(biz.WorkflowFieldName),
				Operator: pkgConst.FilterOperatorContains,
				Value:    req.FuzzyKeyword,
			},
			pkgConst.FilterCondition{
				Field:    string(biz.WorkflowFieldUID),
				Operator: pkgConst.FilterOperatorContains,
				Value:    req.FuzzyKeyword,
			},
		))
	}

	listOption := &biz.ListWorkflowsOption{
		PageNumber:      req.PageIndex,
		LimitPerPage:    req.PageSize,
		OrderBy:         orderBy,
		FilterByOptions: filterByOptions,
	}

	workflows, total, err := d.DataExportWorkflowUsecase.ListDataExportWorkflows(ctx, listOption, currentUserUid, req.FilterByDBServiceUid, req.FilterCurrentStepAssigneeUserUid, string(req.FilterByStatus), req.ProjectUid)
	if nil != err {
		return nil, err
	}

	// 收集所有唯一的项目UID
	projectUIDs := make(map[string]bool)
	for _, w := range workflows {
		if w.ProjectUID != "" {
			projectUIDs[w.ProjectUID] = true
		}
	}

	// 批量获取项目信息
	projectMap := make(map[string]string)
	for projectUID := range projectUIDs {
		project, err := d.ProjectUsecase.GetProject(ctx, projectUID)
		if err != nil {
			// 如果获取项目失败，记录错误但继续处理
			projectMap[projectUID] = "Unknown Project"
		} else {
			projectMap[projectUID] = project.Name
		}
	}

	// 按项目批量解析运维类型名称（D8：禁止逐条请求字典）
	opsTypeNameByProject := d.buildOpsTypeNameMapsByProjects(ctx, projectUIDs)

	ret := make([]*dmsV1.ListDataExportWorkflow, len(workflows))
	for i, w := range workflows {
		ret[i] = &dmsV1.ListDataExportWorkflow{
			ProjectUid:           w.ProjectUID,
			ProjectName:          projectMap[w.ProjectUID],
			WorkflowID:           w.UID,
			WorkflowName:         w.Name,
			Description:          w.Desc,
			CreatedAt:            w.CreatedAt,
			Status:               dmsV1.DataExportWorkflowStatus(w.WorkflowRecord.Status),
			WorkflowTemplateId:   w.WorkflowTemplateId,
			WorkflowTemplateName: w.WorkflowTemplateName,
			OpsType:              resolveOpsTypeFromNameMap(w.OpsTypeUID, opsTypeNameByProject[w.ProjectUID]),
		}
		creater := convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, []string{w.CreateUserUID}))
		if len(creater) > 0 {
			ret[i].Creater = creater[0]
		}
		ret[i].CurrentStepAssigneeUsers = d.dataExportWorkflowCurrentStepAssigneeUsers(ctx, w)
	}

	return &dmsV1.ListDataExportWorkflowsReply{
		Data:  ret,
		Total: total,
	}, nil
}

func (d *DMSService) GetGlobalWorkflowsList(ctx context.Context, req *dmsV1.FilterGlobalDataExportWorkflowReq) (
	*dmsV1.GetGlobalDataExportWorkflowsReply, error) {

	limit, offset := d.GetLimitAndOffset(req.PageIndex, req.PageSize)
	workflows, total, err := d.DataExportWorkflowUsecase.GetGlobalWorkflowsList(ctx, req, limit, offset)
	if err != nil {
		return nil, err
	}

	// 收集所有唯一的项目UID
	projectUIDs := make(map[string]bool)
	for _, w := range workflows {
		if w.ProjectUID != "" {
			projectUIDs[w.ProjectUID] = true
		}
	}

	// 跨项目按所属项目字典批量解析名称（D8：禁止逐条请求）
	opsTypeNameByProject := d.buildOpsTypeNameMapsByProjects(ctx, projectUIDs)

	ret := make([]*dmsV1.GlobalDataExportWorkflow, len(workflows))
	for i, w := range workflows {
		ret[i] = &dmsV1.GlobalDataExportWorkflow{
			ProjectInfo:    w.ProjectInfo,
			WorkflowID:     w.UID,
			WorkflowName:   w.Name,
			Description:    w.Desc,
			CreatedAt:      w.CreatedAt,
			UpdatedAt:      w.WorkflowRecord.UpdateTime,
			Status:         dmsV1.DataExportWorkflowStatus(w.WorkflowRecord.Status),
			DBServiceInfos: w.DBServiceInfos,
			OpsType:        resolveOpsTypeFromNameMap(w.OpsTypeUID, opsTypeNameByProject[w.ProjectUID]),
		}

		creater := convertBizUidWithName(d.UserUsecase.GetBizUserIncludeDeletedWithNameByUids(ctx, []string{w.CreateUserUID}))
		if len(creater) > 0 {
			ret[i].Creater = creater[0]
		}
		ret[i].CurrentStepAssigneeUsers = d.dataExportWorkflowCurrentStepAssigneeUsers(ctx, w)
	}

	return &dmsV1.GetGlobalDataExportWorkflowsReply{
		Data:  ret,
		Total: total,
	}, nil
}

func (d *DMSService) GetDataExportWorkflow(ctx context.Context, req *dmsV1.GetDataExportWorkflowReq, currentUserUid string) (reply *dmsV1.GetDataExportWorkflowReply, err error) {
	w, err := d.DataExportWorkflowUsecase.GetDataExportWorkflow(ctx, req.DataExportWorkflowUid, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("get data export workflow error: %v", err)
	}

	data := &dmsV1.GetDataExportWorkflow{
		Name:                 w.Name,
		WorkflowID:           w.UID,
		Desc:                 w.Desc,
		CreateUser:           convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, []string{w.CreateUserUID}))[0],
		CreateTime:           &w.CreateTime,
		WorkflowTemplateId:   w.WorkflowTemplateId,
		WorkflowTemplateName: w.WorkflowTemplateName,
		OpsType:              d.resolveDataExportWorkflowOpsType(ctx, w.ProjectUID, w.OpsTypeUID),
		WorkflowRecord: dmsV1.WorkflowRecord{
			CurrentStepNumber: uint(w.WorkflowRecord.CurrentWorkflowStepId),
			Status:            dmsV1.DataExportWorkflowStatus(w.WorkflowRecord.Status),
			ExportFailSummary: w.WorkflowRecord.ExportFailSummary,
		},
	}

	for _, task := range w.WorkflowRecord.Tasks {
		data.WorkflowRecord.Tasks = append(data.WorkflowRecord.Tasks, &dmsV1.Task{
			Uid: task.UID,
		})
	}
	for _, v := range w.WorkflowRecord.WorkflowSteps {
		step := &dmsV1.WorkflowStep{
			Number:        v.StepId,
			Users:         convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, v.Assignees)),
			OperationTime: v.OperateAt,
			State:         dmsV1.WorkflowStepStatus(v.State),
			Reason:        v.Reason,
		}
		if v.OperationUserUid != "" && v.State != "init" {
			step.OperationUser = convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, []string{v.OperationUserUid}))[0]
		}
		data.WorkflowRecord.Steps = append(data.WorkflowRecord.Steps, step)
	}

	d.fillGetDataExportUnmaskingWorkflowSummary(ctx, req.DataExportWorkflowUid, data)

	return &dmsV1.GetDataExportWorkflowReply{
		Data: data,
	}, nil
}

// resolveDataExportWorkflowOpsType 按项目字典解析运维类型展示名；未设置/已删/非本项目 → nil（前端「-」）。
func (d *DMSService) resolveDataExportWorkflowOpsType(ctx context.Context, projectUID, opsTypeUID string) *dmsCommonV1.OpsType {
	if opsTypeUID == "" || d.OpsTypeUsecase == nil {
		return nil
	}
	opsType, err := d.OpsTypeUsecase.GetOpsTypeByUID(ctx, opsTypeUID)
	if err != nil || opsType == nil {
		return nil
	}
	if opsType.ProjectUID != "" && projectUID != "" && opsType.ProjectUID != projectUID {
		return nil
	}
	return &dmsCommonV1.OpsType{
		UID:  opsType.UID,
		Name: localizeOpsTypeDisplayName(ctx, opsType.Name),
	}
}

const listOpsTypesBatchPageSize = 1000

// buildOpsTypeNameMapsByProjects 按项目一次拉取字典，构建 uid→name（列表批量回填，禁止逐条 Get）。
func (d *DMSService) buildOpsTypeNameMapsByProjects(ctx context.Context, projectUIDs map[string]bool) map[string]map[string]string {
	out := make(map[string]map[string]string, len(projectUIDs))
	if d.OpsTypeUsecase == nil {
		return out
	}
	for projectUID := range projectUIDs {
		if projectUID == "" {
			continue
		}
		opsTypes, _, err := d.OpsTypeUsecase.ListOpsTypes(ctx, &biz.ListOpsTypesOption{
			ProjectUID: projectUID,
			Limit:      listOpsTypesBatchPageSize,
			Offset:     0,
		})
		if err != nil {
			continue
		}
		nameByUID := make(map[string]string, len(opsTypes))
		for _, ot := range opsTypes {
			if ot == nil || ot.UID == "" {
				continue
			}
			nameByUID[ot.UID] = localizeOpsTypeDisplayName(ctx, ot.Name)
		}
		out[projectUID] = nameByUID
	}
	return out
}

func resolveOpsTypeFromNameMap(opsTypeUID string, nameByUID map[string]string) *dmsCommonV1.OpsType {
	if opsTypeUID == "" || nameByUID == nil {
		return nil
	}
	name, ok := nameByUID[opsTypeUID]
	if !ok {
		return nil
	}
	return &dmsCommonV1.OpsType{
		UID:  opsTypeUID,
		Name: name,
	}
}

func (d *DMSService) ExportDataExportWorkflow(ctx context.Context, req *dmsV1.ExportDataExportWorkflowReq, currentUserUid string) error {
	return d.DataExportWorkflowUsecase.ExportDataExportWorkflow(ctx, req.ProjectUid, req.DataExportWorkflowUid, currentUserUid)

}
func (d *DMSService) AddDataExportTask(ctx context.Context, req *dmsV1.AddDataExportTaskReq, currentUserUid string) (reply *dmsV1.AddDataExportTaskReply, err error) {
	// generate biz arg
	args := make([]*biz.DataExportTask, 0)
	for _, task := range req.DataExportTasks {
		args = append(args, &biz.DataExportTask{
			DBServiceUid:   task.DBServiceUid,
			CreateUserUID:  currentUserUid,
			DatabaseName:   task.DatabaseName,
			ExportType:     "SQL",
			ExportFileType: "CSV",
			ExportSQL:      task.ExportSQL,
			ExportStatus:   biz.DataExportTaskStatusInit,
		})
	}

	uids, err := d.DataExportWorkflowUsecase.AddDataExportTasks(ctx, req.ProjectUid, currentUserUid, args)
	if err != nil {
		return nil, fmt.Errorf("add data export task failed: %v", err)
	}

	return &dmsV1.AddDataExportTaskReply{
		Data: struct {
			Uids []string `json:"data_export_task_uids"`
		}{
			Uids: uids,
		},
	}, nil
}

func (d *DMSService) BatchGetDataExportTask(ctx context.Context, req *dmsV1.BatchGetDataExportTaskReq) (reply *dmsV1.BatchGetDataExportTaskReply, err error) {
	taskUids := strings.Split(req.TaskUids, ",")
	tasks, err := d.DataExportWorkflowUsecase.BatchGetDataExportTask(ctx, taskUids)
	if err != nil {
		return nil, fmt.Errorf("get data export workflow error: %v", err)
	}
	data := make([]*dmsV1.GetDataExportTask, 0)
	for _, task := range tasks {
		data = append(data, &dmsV1.GetDataExportTask{
			TaskUid:          task.UID,
			DBInfo:           dmsV1.TaskDBInfo{UidWithName: convertBizUidWithName(d.DBServiceUsecase.GetBizDBWithNameByUids(ctx, []string{task.DBServiceUid}))[0], DBType: "", DatabaseName: task.DatabaseName},
			Status:           dmsV1.DataExportTaskStatus(task.ExportStatus),
			ExportStartTime:  task.ExportStartTime,
			ExportEndTime:    task.ExportEndTime,
			FileName:         task.ExportFileName,
			ExportType:       task.ExportType,
			ExportFileType:   task.ExportFileType,
			ExportFailStage:  task.ExportFailStage,
			ExportFailReason: task.ExportFailReason,
			AuditResult: dmsV1.AuditTaskResult{
				AuditLevel: task.AuditLevel,
				Score:      task.AuditScore,
				PassRate:   task.AuditPassRate,
			},
		})
	}
	return &dmsV1.BatchGetDataExportTaskReply{
		Data: data,
	}, nil
}

func (d *DMSService) ListDataExportTaskSQLs(ctx context.Context, req *dmsV1.ListDataExportTaskSQLsReq, currentUserUid string) (reply *dmsV1.ListDataExportTaskSQLsReply, err error) {

	orderBy := biz.DataExportTaskRecordFieldNumber

	filterBy := []pkgConst.FilterCondition{{
		Field:    string(biz.DataExportTaskRecordFieldDataExportTaskId),
		Operator: pkgConst.FilterOperatorEqual,
		Value:    req.DataExportTaskUid,
	}}

	listOption := &biz.ListDataExportTaskRecordOption{
		PageNumber:      req.PageIndex,
		LimitPerPage:    req.PageSize,
		OrderBy:         orderBy,
		FilterByOptions: pkgConst.ConditionsToFilterOptions(filterBy),
	}

	taskRecords, total, err := d.DataExportWorkflowUsecase.ListDataExportTaskRecords(ctx, listOption, currentUserUid)
	if nil != err {
		return nil, err
	}

	tasks, err := d.DataExportWorkflowUsecase.BatchGetDataExportTask(ctx, []string{req.DataExportTaskUid})
	if err != nil || len(tasks) == 0 {
		return nil, fmt.Errorf("failed to get data export task: %w", err)
	}
	task := tasks[0]

	ret := make([]*dmsV1.ListDataExportTaskSQL, len(taskRecords))
	for i, w := range taskRecords {
		ret[i] = &dmsV1.ListDataExportTaskSQL{
			ID:            w.Number,
			ExportSQL:     w.ExportSQL,
			AuditLevel:    w.AuditLevel,
			ExportResult:  w.ExportResult,
			ExportStatus:  w.ExportStatus.String(),
			ExportSQLType: w.ExportSQLType,
		}
		if d.UnmaskingWorkflowUsecase != nil {
			lineage, snapshot := d.UnmaskingWorkflowUsecase.AnalyzeLineageAndBuildMaskingSnapshot(ctx, req.ProjectUid, task.DBServiceUid, task.DatabaseName, w.ExportSQL)
			ret[i].LineageAnalysisSnapshot = lineage
			ret[i].MaskingConfigSnapshot = snapshot
		}
		if w.AuditSQLResults != nil {
			for _, result := range w.AuditSQLResults {
				ret[i].AuditSQLResult = append(ret[i].AuditSQLResult, dmsV1.AuditSQLResult{
					Level:           result.Level,
					Message:         result.GetAuditMsgByLangTag(locale.Bundle.GetLangTagFromCtx(ctx)),
					ErrorInfo:       result.GetAuditErrorMsgByLangTag(locale.Bundle.GetLangTagFromCtx(ctx)),
					ExecutionFailed: result.ExecutionFailed,
					RuleName:        result.RuleName,
				})
			}
		}
	}

	return &dmsV1.ListDataExportTaskSQLsReply{
		Data:  ret,
		Total: total,
	}, nil
}

func (d *DMSService) ApproveDataExportWorkflow(ctx context.Context, req *dmsV1.ApproveDataExportWorkflowReq, userId string) (err error) {
	return d.DataExportWorkflowUsecase.ApproveDataExportWorkflow(ctx, req.ProjectUid, req.DataExportWorkflowUid, userId, req.Payload.Reason)
}

func (d *DMSService) RejectDataExportWorkflow(ctx context.Context, req *dmsV1.RejectDataExportWorkflowReq, userId string) (err error) {
	return d.DataExportWorkflowUsecase.RejectDataExportWorkflow(ctx, req, userId)
}

func (d *DMSService) CancelDataExportWorkflow(ctx context.Context, req *dmsV1.CancelDataExportWorkflowReq, userId string) (err error) {
	return d.DataExportWorkflowUsecase.CancelDataExportWorkflow(ctx, userId, req)
}

func (d *DMSService) DownloadDataExportTask(ctx context.Context, req *dmsV1.DownloadDataExportTaskReq, userId string) (bool, string, error) {
	return d.DataExportWorkflowUsecase.DownloadDataExportTask(ctx, userId, req)
}

func (d *DMSService) DownloadDataExportTaskSQLs(ctx context.Context, req *dmsV1.DownloadDataExportTaskSQLsReq, userId string) (string, []byte, error) {
	return d.DataExportWorkflowUsecase.DownloadDataExportTaskSQLs(ctx, req, userId)
}

// dataExportWorkflowCurrentStepAssigneeUsers 返回工单待操作人；终态（完成/关闭/失败）不返回待操作人。
func (d *DMSService) dataExportWorkflowCurrentStepAssigneeUsers(ctx context.Context, w *biz.Workflow) []dmsV1.UidWithName {
	if isDataExportWorkflowTerminalStatus(w.Status) {
		return nil
	}
	if w.Status == string(dmsV1.DataExportWorkflowStatusWaitForExport) ||
		w.Status == string(dmsV1.DataExportWorkflowStatusWaitForExporting) ||
		w.Status == string(dmsV1.DataExportWorkflowStatusRejected) {
		// wait_for_export/wait_for_exporting/rejected 状态下，待操作人为工单创建者
		return convertBizUidWithName(d.UserUsecase.GetBizUserIncludeDeletedWithNameByUids(ctx, []string{w.CreateUserUID}))
	}
	if w.WorkflowRecord != nil && w.WorkflowRecord.CurrentWorkflowStepId > 0 &&
		int(w.WorkflowRecord.CurrentWorkflowStepId-1) < len(w.WorkflowRecord.WorkflowSteps) {
		return convertBizUidWithName(d.UserUsecase.GetBizUserIncludeDeletedWithNameByUids(ctx, w.WorkflowRecord.WorkflowSteps[w.WorkflowRecord.CurrentWorkflowStepId-1].Assignees))
	}
	return nil
}

func isDataExportWorkflowTerminalStatus(status string) bool {
	switch status {
	case string(dmsV1.DataExportWorkflowStatusFinish),
		string(dmsV1.DataExportWorkflowStatusCancel),
		string(dmsV1.DataExportWorkflowStatusFailed):
		return true
	default:
		return false
	}
}

func (d *DMSService) CheckDataExportWorkflowTemplateUsed(ctx context.Context, req *dmsV1.CheckDataExportWorkflowTemplateUsedReq) (*dmsV1.CheckDataExportWorkflowTemplateUsedReply, error) {
	isUsed, count, err := d.DataExportWorkflowUsecase.CheckDataExportWorkflowTemplateUsed(ctx, req.ProjectUid, req.WorkflowTemplateId)
	if err != nil {
		return nil, err
	}
	reply := &dmsV1.CheckDataExportWorkflowTemplateUsedReply{}
	reply.Data.IsUsed = isUsed
	reply.Data.Count = count
	return reply, nil
}
