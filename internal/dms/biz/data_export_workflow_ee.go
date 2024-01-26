//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	export "github.com/actiontech/dms/internal/dataQuery/pkg/dataExport"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	v1Base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

func (d *DataExportWorkflowUsecase) AddDataExportWorkflow(ctx context.Context, currentUserId string, params *Workflow) (string, error) {
	// geerate workflow record
	workflowRecordUid, err := pkgRand.GenStrUid()
	if err != nil {
		return "", err
	}
	workflowRecord := &WorkflowRecord{
		UID:                   workflowRecordUid,
		CurrentWorkflowStepId: 1,
		Status:                DataExportWorkflowStatusWaitForApprove,
		Tasks:                 params.Tasks,
	}

	taskIds := make([]string, 0)
	for _, t := range params.Tasks {
		taskIds = append(taskIds, t.UID)
	}
	tasks, err := d.dataExportTaskRepo.GetDataExportTaskByIds(ctx, taskIds)
	if err != nil {
		return "", err
	}
	dbServiceUids := make([]string, 0)
	for _, task := range tasks {
		dbServiceUids = append(dbServiceUids, task.DBServiceUid)
	}

	userPermissions, err := d.opPermissionVerifyUsecase.GetUserOpPermissionInProject(ctx, currentUserId, params.ProjectUID)
	if err != nil {
		return "", err
	}
	for _, dbServiceUid := range dbServiceUids {
		if !d.opPermissionVerifyUsecase.UserCanOpDB(userPermissions, []string{pkgConst.UIDOfOpPermissionExportCreate}, dbServiceUid) {
			return "", fmt.Errorf("current user has not enough permission to create export workflow: db service %v", dbServiceUid)
		}
	}

	// 获取审批人
	approveWorkflowUsers := make([]string, 0)
	approveWorkflowMapUsers := make(map[string] /*userId*/ struct{})
	for _, task := range tasks {
		opUsers, err := d.opPermissionVerifyUsecase.GetCanOpDBUsers(ctx, params.ProjectUID, task.DBServiceUid, []string{pkgConst.UIDOfOpPermissionExportApprovalReject})
		if err != nil {
			return "", fmt.Errorf("get op users fail: %v", err)
		}
		for _, opUser := range opUsers {
			if _, ok := approveWorkflowMapUsers[opUser]; !ok {
				approveWorkflowMapUsers[opUser] = struct{}{}
				approveWorkflowUsers = append(approveWorkflowUsers, opUser)
			}
		}
	}

	//  generate workflow step
	steps, err := generateWorkflowStep(workflowRecord.UID, approveWorkflowUsers)
	if err != nil {
		return "", err
	}
	workflowRecord.WorkflowSteps = steps

	// create workflow
	workflowUid, err := pkgRand.GenStrUid()
	if err != nil {
		return "", err
	}
	err = d.repo.SaveWorkflow(ctx, &Workflow{
		UID:               workflowUid,
		Name:              params.Name,
		ProjectUID:        params.ProjectUID,
		WorkflowType:      "data_export", // 枚举
		Desc:              params.Desc,
		CreateTime:        time.Now(),
		CreateUserUID:     currentUserId,
		WorkflowRecordUid: workflowRecordUid,
		WorkflowRecord:    workflowRecord,
	})

	return workflowUid, err
}

func generateWorkflowStep(workflowRecordUid string, allInspector []string) ([]*WorkflowStep, error) {
	return []*WorkflowStep{
		{
			StepId:            1,
			WorkflowRecordUid: workflowRecordUid,
			State:             "init",
			Assignees:         allInspector,
		},
	}, nil
}

func (d *DataExportWorkflowUsecase) ListDataExportWorkflows(ctx context.Context, workflowsOption *ListWorkflowsOption, currentUserId, dbServiceUID string) ([]*Workflow, int64, error) {
	// TODO 导出任务实现后对接 filter by task's db_service uid
	// filter task id,将查询的workflow_recordid 作为筛选值加入workflowOption

	services, total, err := d.repo.ListDataExportWorkflows(ctx, workflowsOption)
	if err != nil {
		return nil, 0, fmt.Errorf("list db services failed: %w", err)
	}
	return services, total, nil
}

func (d *DataExportWorkflowUsecase) GetDataExportWorkflow(ctx context.Context, workflowUid, currentUserId string) (*Workflow, error) {
	// 校验用户查看权限

	services, err := d.repo.GetDataExportWorkflow(ctx, workflowUid)
	if err != nil {
		return nil, fmt.Errorf("get data export workflow failed: %w", err)
	}
	return services, nil
}

func (d *DataExportWorkflowUsecase) ExportDataExportWorkflow(ctx context.Context, projectUid, workflowUid, currentUserId string) error {

	workflow, err := d.repo.GetDataExportWorkflow(ctx, workflowUid)
	if err != nil {
		return err
	}
	// 校验工单状态
	if workflow.WorkflowRecord.Status != DataExportWorkflowStatusWaitForExport {
		return fmt.Errorf("current workflow status is %v, not allow to export", workflow.WorkflowRecord.Status)
	}
	// 校验用户查看权限
	if currentUserId != workflow.CreateUserUID {
		return fmt.Errorf("current user is not allow to export workflow")
	}
	// 审批完成24小时后不允许导出
	for _, step := range workflow.WorkflowRecord.WorkflowSteps {
		if step.StepId == workflow.WorkflowRecord.CurrentWorkflowStepId && step.OperateAt.Before(time.Now().Add(-time.Hour*24)) {
			return fmt.Errorf("export is not allowed after more than 24 hours after approval is completed")
		}
	}

	// 更新workflow 状态
	err = d.repo.UpdateWorkflowStatusById(ctx, workflow.WorkflowRecordUid, DataExportWorkflowStatusWaitForExporting)
	if err != nil {
		return err
	}

	go d.ExportWorkflow(context.Background(), workflow)

	return nil
}

func (d *DataExportWorkflowUsecase) ExportWorkflow(ctx context.Context, workflow *Workflow) {
	var err error
	tx := d.tx.BeginTX(ctx)
	defer func() {
		// 更新工单状态
		workflow.WorkflowRecord.Status = DataExportWorkflowStatusFinish
		if err != nil {
			d.log.Warn(tx.RollbackWithError(d.log, err))
			workflow.WorkflowRecord.Status = DataExportWorkflowStatusFailed
		}
		updateErr := d.repo.UpdateWorkflowStatusById(ctx, workflow.WorkflowRecordUid, DataExportWorkflowStatus(workflow.WorkflowRecord.Status))
		if updateErr != nil {
			d.log.Error(updateErr)
		}
	}()

	// 获取导出任务
	taskIds := make([]string, 0)
	for _, t := range workflow.WorkflowRecord.Tasks {
		taskIds = append(taskIds, t.UID)
	}
	if len(taskIds) == 0 {
		d.log.Error(fmt.Errorf("workflwo has no export task"))
	}

	// 执行导出任务
	tasks, err := d.dataExportTaskRepo.GetDataExportTaskByIds(tx, taskIds)
	if err != nil {
		d.log.Error(fmt.Errorf("get data export task failed: %v", err))
	}

	for _, task := range tasks {
		task.ExportFileName = fmt.Sprintf("%s-%s.zip", workflow.Name, task.UID)
		err := d.ExecExportTask(tx, task)
		if err != nil {
			d.log.Errorf("exec export task fail: %v", err)
		}
	}

	err = d.dataExportTaskRepo.SaveDataExportTask(tx, tasks)
	if err != nil {
		d.log.Error(err)
	}

	if err := tx.Commit(d.log); err != nil {
		d.log.Error(fmt.Errorf("commit tx failed: %v", err))
	}
}

const ExportFilePath string = "./"

func (d *DataExportWorkflowUsecase) ExecExportTask(ctx context.Context, taskInfo *DataExportTask) (err error) {
	dbService, err := d.dbServiceRepo.GetDBService(ctx, taskInfo.DBServiceUid)
	if err != nil {
		return fmt.Errorf("get db service failed: %v", err)
	}
	db, err := export.NewMysqlConn(dbService.Host, dbService.Port, dbService.User, dbService.Password, taskInfo.DatabaseName)
	if err != nil {
		return err
	}
	defer db.Close()

	startTime := time.Now()
	taskInfo.ExportStartTime = &startTime

	// TODO recor状态返回
	exportTasks := make([]*export.ExportTask, 0)
	for _, record := range taskInfo.DataExportTaskRecords {
		exportTasks = append(exportTasks, export.NewExportTask().WithExtract(export.NewExtract(db, record.ExportSQL)).WithExporter(fmt.Sprintf("%s%s_%d.csv", ExportFilePath, record.DataExportTaskId, record.Number), export.NewCsvExport()))
	}

	err = export.ExportTasksToZip(taskInfo.ExportFileName, exportTasks)
	if err != nil {
		taskInfo.ExportStatus = DataExportTaskStatusFailed
	}
	taskInfo.ExportStatus = DataExportTaskStatusFinish
	endTime := time.Now()
	taskInfo.ExportEndTime = &endTime
	return
}

func (d *DataExportWorkflowUsecase) AddDataExportTasks(ctx context.Context, projectUid, currentUserId string, params []*DataExportTask) (taskids []string, err error) {

	project, err := d.projectUsecase.GetProject(ctx, projectUid)
	if err != nil {
		return nil, err
	}
	// 校验创建导出任务权限
	userPermissions, err := d.opPermissionVerifyUsecase.GetUserOpPermissionInProject(ctx, currentUserId, projectUid)
	if err != nil {
		return nil, err
	}
	for _, task := range params {
		if !d.opPermissionVerifyUsecase.UserCanOpDB(userPermissions, []string{pkgConst.UIDOfOpPermissionExportCreate}, task.DBServiceUid) {
			return nil, fmt.Errorf("current user has not enough permission to create export workflow: db service %v", task.DBServiceUid)
		}
	}

	dataExportTasks := make([]*DataExportTask, 0)
	for _, v := range params {
		// geerate data export task record
		dataExportTaskUid, err := pkgRand.GenStrUid()
		if err != nil {
			return nil, err
		}
		taskids = append(taskids, dataExportTaskUid)

		auditTaskInfo, AuditRecord, err := d.SQLEAuditSQL(ctx, project.Name, dataExportTaskUid, v.DBServiceUid, v.ExportSQL)
		if err != nil {
			return nil, fmt.Errorf("audit export sql err: %v", err)
		}
		dataExportTasks = append(dataExportTasks, &DataExportTask{
			UID:                   dataExportTaskUid,
			DBServiceUid:          v.DBServiceUid,
			CreateUserUID:         v.CreateUserUID,
			DatabaseName:          v.DatabaseName,
			ExportType:            v.ExportType,
			ExportFileType:        v.ExportFileType,
			ExportSQL:             v.ExportSQL,
			ExportStatus:          DataExportTaskStatusInit,
			AuditPassRate:         auditTaskInfo.AuditPassRate,
			AuditScore:            auditTaskInfo.AuditScore,
			AuditLevel:            auditTaskInfo.AuditLevel,
			DataExportTaskRecords: AuditRecord,
		})
	}

	err = d.dataExportTaskRepo.SaveDataExportTask(ctx, dataExportTasks)
	if err != nil {
		return nil, fmt.Errorf("sava data export tasks err: %v", err)
	}
	return taskids, err
}

type AuditTaskInfo struct {
	AuditPassRate float64
	AuditScore    int32
	AuditLevel    string
}

/*
	接入 SQLE审核,简化版参数
	url:  /v1/projects/{project_name}/tasks/audits
*/

type CreateAuditTaskReqV1 struct {
	InstanceName   string `json:"instance_name" form:"instance_name" example:"inst_1" valid:"required"`
	InstanceSchema string `json:"instance_schema" form:"instance_schema" example:"db1"`
	Sql            string `json:"sql" form:"sql" example:"alter table tb1 drop columns c1"`
}

type GetAuditTaskResV1 struct {
	v1Base.GenericResp
	Data *AuditTaskResV1 `json:"data"`
}

type AuditTaskResV1 struct {
	Id             uint    `json:"task_id"`
	InstanceName   string  `json:"instance_name"`
	InstanceDbType string  `json:"instance_db_type"`
	InstanceSchema string  `json:"instance_schema" example:"db1"`
	AuditLevel     string  `json:"audit_level" enums:"normal,notice,warn,error,"`
	Score          int32   `json:"score"`
	PassRate       float64 `json:"pass_rate"`
	Status         string  `json:"status" enums:"initialized,audited,executing,exec_success,exec_failed,manually_executed"`
	SQLSource      string  `json:"sql_source" enums:"form_data,sql_file,mybatis_xml_file,audit_plan"`
}

/*
	审核SQL结果集
	url: /v2/tasks/audits/:task_id/sqls?page_index=1&page_size=20&no_duplicate=false
*/

type GetAuditTaskSQLsReqV2 struct {
	FilterExecStatus  string `json:"filter_exec_status" query:"filter_exec_status"`
	FilterAuditStatus string `json:"filter_audit_status" query:"filter_audit_status"`
	FilterAuditLevel  string `json:"filter_audit_level" query:"filter_audit_level"`
	NoDuplicate       bool   `json:"no_duplicate" query:"no_duplicate"`
	PageIndex         uint32 `json:"page_index" query:"page_index" valid:"required"`
	PageSize          uint32 `json:"page_size" query:"page_size" valid:"required"`
}

type GetAuditTaskSQLsResV2 struct {
	v1Base.GenericResp
	Data      []*AuditTaskSQLResV2 `json:"data"`
	TotalNums uint64               `json:"total_nums"`
}

type AuditTaskSQLResV2 struct {
	Number        uint           `json:"number"`
	ExecSQL       string         `json:"exec_sql"`
	SQLSourceFile string         `json:"sql_source_file"`
	AuditResult   []*AuditResult `json:"audit_result"`
	AuditLevel    string         `json:"audit_level"`
	AuditStatus   string         `json:"audit_status"`
	ExecResult    string         `json:"exec_result"`
	ExecStatus    string         `json:"exec_status"`
	RollbackSQL   string         `json:"rollback_sql,omitempty"`
	Description   string         `json:"description"`
}

func (d *DataExportWorkflowUsecase) SQLEAuditSQL(ctx context.Context, projectName, taskId, DBServiceUid, Sqls string) (*AuditTaskInfo, []*DataExportTaskRecord, error) {
	// sqle地址、请求头
	target, err := d.dmsProxyTargetRepo.GetProxyTargetByName(ctx, cloudbeaver.SQLEProxyName)
	if err != nil {
		return nil, nil, err
	}
	sqleUrl := target.URL.String()
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	// 获取数据源名称
	dbService, err := d.dbServiceRepo.GetDBService(ctx, DBServiceUid)
	if err != nil {
		return nil, nil, err
	}

	// 执行SQL审核 /v1/projects/{project_name}/tasks/audits
	auditReply := &GetAuditTaskResV1{}
	auditUri := fmt.Sprintf("/v1/projects/%s/tasks/audits", projectName)
	err = pkgHttp.POST(ctx, fmt.Sprintf("%s%s", sqleUrl, auditUri), header, CreateAuditTaskReqV1{
		InstanceName:   dbService.Name,
		InstanceSchema: "",
		Sql:            Sqls,
	}, auditReply)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql audit, url :%v, err : %v", sqleUrl, err)
	} else if auditReply.Code != 0 {
		return nil, nil, fmt.Errorf("create sql audit code is %v", auditReply.Code)
	}
	auditTaskInfo := &AuditTaskInfo{
		AuditPassRate: auditReply.Data.PassRate,
		AuditScore:    auditReply.Data.Score,
		AuditLevel:    auditReply.Data.AuditLevel,
	}

	// 查询审核结果集 /v2/tasks/audits/:task_id/sqls?page_index=1&page_size=20&no_duplicate=false
	auditTaskSQLsReply := &GetAuditTaskSQLsResV2{}
	if err := pkgHttp.Get(ctx, fmt.Sprintf("%v/v2/tasks/audits/%d/sqls", target.URL.String(), auditReply.Data.Id), header, GetAuditTaskSQLsReqV2{
		NoDuplicate: false,
		PageIndex:   1,
		PageSize:    999,
	}, auditTaskSQLsReply); err != nil {
		return nil, nil, fmt.Errorf("failed to get task audit sql records  %v: %v", sqleUrl, err)
	}
	if auditTaskSQLsReply.Code != 0 {
		return nil, nil, fmt.Errorf("http reply code(%v) error: %v", auditTaskSQLsReply.Code, auditTaskSQLsReply.Message)
	}

	auditResultRecords := make([]*DataExportTaskRecord, 0)
	for _, record := range auditTaskSQLsReply.Data {
		auditResultRecords = append(auditResultRecords, &DataExportTaskRecord{
			DataExportTaskId: taskId,
			Number:           record.Number,
			ExportSQL:        record.ExecSQL,
			AuditLevel:       record.AuditLevel,
			AuditSQLResults:  record.AuditResult,
		})
	}

	return auditTaskInfo, auditResultRecords, nil
}

func (d *DataExportWorkflowUsecase) BatchGetDataExportTask(ctx context.Context, taskUids []string, currentUserId string) ([]*DataExportTask, error) {
	// 校验用户查看权限

	tasks, err := d.dataExportTaskRepo.GetDataExportTaskByIds(ctx, taskUids)
	if err != nil {
		return nil, fmt.Errorf("get data export workflow failed: %w", err)
	}
	return tasks, nil
}

func (d *DataExportWorkflowUsecase) ListDataExportTaskRecords(ctx context.Context, options *ListDataExportTaskRecordOption, currentUserId string) ([]*DataExportTaskRecord, int64, error) {
	// 校验用户查看权限

	tasks, total, err := d.dataExportTaskRepo.ListDataExportTaskRecord(ctx, options)
	if err != nil {
		return nil, 0, fmt.Errorf("list data export task records failed: %w", err)
	}
	return tasks, total, nil
}

func (d *DataExportWorkflowUsecase) ApproveDataExportWorkflow(ctx context.Context, projectId, workflowId, userId string) error {
	return d.auditDataExportWorkflow(ctx, DataExportWorkflowStatusWaitForExport, DataExportWorkflowStatusFinish.String(), workflowId, userId, "")
}

func (d *DataExportWorkflowUsecase) auditDataExportWorkflow(ctx context.Context, workflowStatus DataExportWorkflowStatus, stepStatus, workflowId, userId, reason string) error {
	workflow, err := d.repo.GetDataExportWorkflow(ctx, workflowId)
	if err != nil {
		return fmt.Errorf("get data export workflow failed: %w", err)
	}

	if workflow.WorkflowRecord.Status != DataExportWorkflowStatusWaitForApprove {
		return fmt.Errorf("workflow %s status %s cannot be performed", workflow.Name, workflow.WorkflowRecord.Status)
	}

	var workflowStep *WorkflowStep
	isApproveUser := false
	for _, item := range workflow.WorkflowRecord.WorkflowSteps {
		if item.StepId != workflow.WorkflowRecord.CurrentWorkflowStepId {
			continue
		}
		item.State = stepStatus
		for _, assignee := range item.Assignees {
			if assignee == userId {
				isApproveUser = true
				workflowStep = item
			}
		}
	}

	if !isApproveUser {
		return fmt.Errorf("current user not executable approve workflow")
	}

	return d.repo.AuditWorkflow(ctx, workflow.WorkflowRecord.UID, workflowStatus, workflowStep, userId, reason)
}

func (d *DataExportWorkflowUsecase) RejectDataExportWorkflow(ctx context.Context, req *dmsV1.RejectDataExportWorkflowReq, userId string) error {
	return d.auditDataExportWorkflow(ctx, DataExportWorkflowStatusRejected, DataExportWorkflowStatusRejected.String(), req.DataExportWorkflowUid, userId, req.Payload.Reason)
}

func (d *DataExportWorkflowUsecase) CancelDataExportWorkflow(ctx context.Context, userId string, req *dmsV1.CancelDataExportWorkflowReq) error {
	isAdmin, err := d.opPermissionVerifyUsecase.IsUserDMSAdmin(ctx, userId)
	if err != nil {
		return err
	}

	workflows, err := d.repo.GetDataExportWorkflowsByIds(ctx, req.Payload.DataExportWorkflowUids)
	if err != nil {
		return err
	}

	cancelWorkflowRecordIds := make([]string, 0, len(workflows))
	cancelWorkflowSteps := make([]*WorkflowStep, 0, len(workflows))
	for _, workflow := range workflows {
		if workflow.WorkflowRecord.Status != DataExportWorkflowStatusWaitForApprove && workflow.WorkflowRecord.Status != DataExportWorkflowStatusRejected {
			return fmt.Errorf("workflow %s status %s cannot be performed", workflow.Name, workflow.WorkflowRecord.Status)
		}

		if !isAdmin && workflow.CreateUserUID != userId {
			return fmt.Errorf("current user not executable cancel workflow")
		}

		for _, step := range workflow.WorkflowRecord.WorkflowSteps {
			if workflow.WorkflowRecord.CurrentWorkflowStepId == step.StepId {
				cancelWorkflowSteps = append(cancelWorkflowSteps, step)
			}
		}

		cancelWorkflowRecordIds = append(cancelWorkflowRecordIds, workflow.WorkflowRecord.UID)
	}

	return d.repo.CancelWorkflow(ctx, cancelWorkflowRecordIds, cancelWorkflowSteps, userId)
}

func (d *DataExportWorkflowUsecase) DownloadDataExportTask(ctx context.Context, userId string, req *dmsV1.DownloadDataExportTaskReq) (string, error) {
	tasks, err := d.dataExportTaskRepo.GetDataExportTaskByIds(ctx, []string{req.DataExportTaskUid})
	if err != nil {
		return "", err
	}

	if len(tasks) == 0 {
		return "", fmt.Errorf("task does not exist")
	}

	if userId != tasks[0].CreateUserUID {
		return "", fmt.Errorf("current user not executable download")
	}

	filename := filepath.Join(ExportFilePath, tasks[0].ExportFileName)

	if _, err = os.Stat(filename); err != nil {
		return "", err
	} else if os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", filename)
	}

	if tasks[0].ExportEndTime.Before(time.Now().Add(-24 * time.Hour)) {
		return "", fmt.Errorf("the file %s download has exceeded 24 hours", tasks[0].ExportFileName)
	}

	return filename, nil
}
