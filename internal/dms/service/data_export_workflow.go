package service

import (
	"context"
	"fmt"
	"strings"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *DMSService) AddDataExportWorkflow(ctx context.Context, req *dmsV1.AddDataExportWorkflowReq, currentUserUid string) (reply *dmsV1.AddDataExportWorkflowReply, err error) {
	// generate biz args
	tasks := make([]biz.Task, 0)
	for _, t := range req.DataExportWorkflow.Tasks {
		tasks = append(tasks, biz.Task{UID: t.Uid})
	}
	args := &biz.Workflow{
		Name:       req.DataExportWorkflow.Name,
		Desc:       req.DataExportWorkflow.Desc,
		Tasks:      tasks,
		ProjectUID: req.ProjectUid,
	}
	uid, err := d.DataExportWorkflowUsecase.AddDataExportWorkflow(ctx, currentUserUid, args)
	if err != nil {
		return nil, fmt.Errorf("add data export workflow failed: %v", err)
	}

	return &dmsV1.AddDataExportWorkflowReply{
		Data: struct {
			Uid string `json:"export_data_workflow_uid"`
		}{Uid: uid}}, nil
}

func (d *DMSService) ListDataExportWorkflow(ctx context.Context, req *dmsV1.ListDataExportWorkflowsReq, currentUserUid string) (reply *dmsV1.ListDataExportWorkflowsReply, err error) {
	// default order by
	orderBy := biz.WorkflowFieldCreateTime

	filterBy := make([]pkgConst.FilterCondition, 0)
	filterBy = append(filterBy, pkgConst.FilterCondition{
		Field:    string(biz.WorkflowFieldWorkflowType),
		Operator: pkgConst.FilterOperatorEqual,
		Value:    biz.DataExportWorkflowEventType.String(),
	})

	if req.FilterByCreateUserUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldCreateUserUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByCreateUserUid,
		})
	}

	if req.FilterCreateTimeFrom != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldCreateTime),
			Operator: pkgConst.FilterOperatorGreaterThanOrEqual,
			Value:    req.FilterCreateTimeFrom,
		})
	}

	if req.FilterCreateTimeTo != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldCreateTime),
			Operator: pkgConst.FilterOperatorLessThanOrEqual,
			Value:    req.FilterCreateTimeTo,
		})
	}

	if req.ProjectUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.WorkflowFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	if req.FuzzyKeyword != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:         string(biz.WorkflowFieldName),
			Operator:      pkgConst.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
		}, pkgConst.FilterCondition{
			Field:         string(biz.WorkflowFieldUID),
			Operator:      pkgConst.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
		},
		)
	}

	listOption := &biz.ListWorkflowsOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	workflows, total, err := d.DataExportWorkflowUsecase.ListDataExportWorkflows(ctx, listOption, currentUserUid, req.FilterByDBServiceUid, req.FilterCurrentStepAssigneeUserUid, string(req.FilterByStatus))
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListDataExportWorkflow, len(workflows))
	for i, w := range workflows {
		ret[i] = &dmsV1.ListDataExportWorkflow{
			ProjectUid:   w.ProjectUID,
			WorkflowID:   w.UID,
			WorkflowName: w.Name,
			Description:  w.Desc,
			Creater:      convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, []string{w.CreateUserUID}))[0],
			CreatedAt:    w.CreatedAt,
			Status:       dmsV1.DataExportWorkflowStatus(w.WorkflowRecord.Status),
		}
		if w.WorkflowRecord.WorkflowSteps[w.WorkflowRecord.CurrentWorkflowStepId-1].State == "init" {
			ret[i].CurrentStepAssigneeUsers = convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, w.WorkflowRecord.WorkflowSteps[w.WorkflowRecord.CurrentWorkflowStepId-1].Assignees))
		}

	}

	return &dmsV1.ListDataExportWorkflowsReply{
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
		Name:       w.Name,
		WorkflowID: w.UID,
		Desc:       w.Desc,
		CreateUser: convertBizUidWithName(d.UserUsecase.GetBizUserWithNameByUids(ctx, []string{w.CreateUserUID}))[0],
		CreateTime: &w.CreateTime,
		WorkflowRecord: dmsV1.WorkflowRecord{
			CurrentStepNumber: uint(w.WorkflowRecord.CurrentWorkflowStepId),
			Status:            dmsV1.DataExportWorkflowStatus(w.WorkflowRecord.Status),
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

	return &dmsV1.GetDataExportWorkflowReply{
		Data: data,
	}, nil
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

func (d *DMSService) BatchGetDataExportTask(ctx context.Context, req *dmsV1.BatchGetDataExportTaskReq, currentUserUid string) (reply *dmsV1.BatchGetDataExportTaskReply, err error) {
	taskUids := strings.Split(req.TaskUids, ",")
	tasks, err := d.DataExportWorkflowUsecase.BatchGetDataExportTask(ctx, taskUids, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("get data export workflow error: %v", err)
	}
	data := make([]*dmsV1.GetDataExportTask, 0)
	for _, task := range tasks {
		data = append(data, &dmsV1.GetDataExportTask{
			TaskUid:         task.UID,
			DBInfo:          dmsV1.TaskDBInfo{UidWithName: convertBizUidWithName(d.DBServiceUsecase.GetBizDBWithNameByUids(ctx, []string{task.DBServiceUid}))[0], DBType: "", DatabaseName: task.DatabaseName},
			Status:          dmsV1.DataExportTaskStatus(task.ExportStatus),
			ExportStartTime: task.ExportStartTime,
			ExportEndTime:   task.ExportEndTime,
			FileName:        task.ExportFileName,
			ExportType:      task.ExportType,
			ExportFileType:  task.ExportFileType,
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
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	taskRecords, total, err := d.DataExportWorkflowUsecase.ListDataExportTaskRecords(ctx, listOption, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListDataExportTaskSQL, len(taskRecords))
	for i, w := range taskRecords {
		ret[i] = &dmsV1.ListDataExportTaskSQL{
			ID:            w.Number,
			ExportSQL:     w.ExportSQL,
			AuditLevel:    w.AuditLevel,
			ExportResult:  w.ExportResult,
			ExportSQLType: w.ExportSQLType,
		}
		if w.AuditSQLResults != nil {
			for _, result := range w.AuditSQLResults {
				ret[i].AuditSQLResult = append(ret[i].AuditSQLResult, dmsV1.AuditSQLResult{
					Level:    result.Level,
					Message:  result.Message,
					RuleName: result.RuleName,
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
	return d.DataExportWorkflowUsecase.ApproveDataExportWorkflow(ctx, req.ProjectUid, req.DataExportWorkflowUid, userId)
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
