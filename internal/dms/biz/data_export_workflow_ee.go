//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
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

func (d *DataExportWorkflowUsecase) ListDataExportWorkflows(ctx context.Context, workflowsOption *ListWorkflowsOption, currentUserId, dbServiceUID, filterCurrentStepAssigneeUserUid, filterByStatus string) ([]*Workflow, int64, error) {

	// for preload
	workflowsOption.FilterBy = append(workflowsOption.FilterBy, pkgConst.FilterCondition{
		Table: "WorkflowRecord",
	}, pkgConst.FilterCondition{
		Table: "WorkflowRecord.Steps",
	})

	// 当前用户创建 OR 当前用户能操作的工单
	filterWorkflowUids, err := d.repo.GetDataExportWorkflowsForView(ctx, currentUserId)
	if err != nil {
		return nil, 0, err
	}

	// 非workflow筛选条件
	{

		merge := func(s1, s2 []string) []string {
			visited := make(map[string]bool)
			result := []string{}

			for _, v := range s1 {
				visited[v] = true
			}

			for _, v := range s2 {
				if visited[v] {
					result = append(result, v)
				}
			}

			return result
		}

		// 跟workflowRecord相关
		if filterByStatus != "" {
			workflowUids, err := d.repo.GetDataExportWorkflowsByStatus(ctx, filterByStatus)
			if err != nil {
				return nil, 0, err
			}
			filterWorkflowUids = merge(filterWorkflowUids, workflowUids)
		}

		// 跟step 相关
		if filterCurrentStepAssigneeUserUid != "" {
			workflowUids, err := d.repo.GetDataExportWorkflowsByAssignUser(ctx, filterCurrentStepAssigneeUserUid)
			if err != nil {
				return nil, 0, err
			}
			filterWorkflowUids = merge(filterWorkflowUids, workflowUids)
		}

		// 跟task相关
		if dbServiceUID != "" {
			workflowUids, err := d.repo.GetDataExportWorkflowsByDBService(ctx, dbServiceUID)
			if err != nil {
				return nil, 0, err
			}
			filterWorkflowUids = merge(filterWorkflowUids, workflowUids)
		}
	}
	workflowsOption.FilterBy = append(workflowsOption.FilterBy, pkgConst.FilterCondition{
		Field:    string(WorkflowFieldUID),
		Operator: pkgConst.FilterOperatorIn,
		Value:    filterWorkflowUids,
	})

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

const ExportFilenameSeparate = ";"

func (d *DataExportWorkflowUsecase) ExportWorkflow(ctx context.Context, workflow *Workflow) {
	workflow.WorkflowRecord.Status = DataExportWorkflowStatusFailed

	var exportFailed bool
	defer func() {
		if err := d.repo.UpdateWorkflowStatusById(ctx, workflow.WorkflowRecordUid, workflow.WorkflowRecord.Status); err != nil {
			d.log.Error(err)
		}
	}()

	// 获取导出任务
	taskIds := make([]string, 0)
	for _, t := range workflow.WorkflowRecord.Tasks {
		taskIds = append(taskIds, t.UID)
	}
	if len(taskIds) == 0 {
		exportFailed = true
		d.log.Error(fmt.Errorf("workflwo has no export task"))
		return
	}

	if err := d.dataExportTaskRepo.BatchUpdateDataExportTaskStatusByIds(ctx, taskIds, DataExportTaskStatusExporting); err != nil {
		exportFailed = true
		d.log.Error(err)
		return
	}

	// 执行导出任务
	tasks, err := d.dataExportTaskRepo.GetDataExportTaskByIds(ctx, taskIds)
	if err != nil {
		exportFailed = true
		d.log.Error(fmt.Errorf("get data export task failed: %v", err))
		return
	}

	for _, task := range tasks {
		// export file name
		task.ExportFileName = fmt.Sprintf("%s-%s.zip", workflow.Name, task.UID)
		err := d.ExecExportTask(ctx, task)
		if err != nil {
			exportFailed = true
			d.log.Errorf("exec export task fail: %v", err)
		}

		// export_file_name filed value
		task.ExportFileName = fmt.Sprintf("%s%s%s", d.reportHost, ExportFilenameSeparate, task.ExportFileName)
	}
	if err = d.dataExportTaskRepo.SaveDataExportTask(ctx, tasks); err != nil {
		exportFailed = true
		d.log.Error(err)
		return
	}

	// 更新工单状态
	if !exportFailed {
		workflow.WorkflowRecord.Status = DataExportWorkflowStatusFinish
	}
}

const ExportFilePath string = "./export/"

func (d *DataExportWorkflowUsecase) ExecExportTask(ctx context.Context, taskInfo *DataExportTask) (err error) {
	defer func() {
		if err != nil {
			d.log.Error(err)
			taskInfo.ExportStatus = DataExportTaskStatusFailed
		}
	}()

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
		exportTasks = append(exportTasks, export.NewExportTask().WithExtract(export.NewExtract(db, record.ExportSQL)).WithExporter(fmt.Sprintf("%s_%d.csv", record.DataExportTaskId, record.Number), export.NewCsvExport()))
	}
	if _, err := os.Stat(ExportFilePath); os.IsNotExist(err) {
		// 文件夹不存在，创建它
		err = os.MkdirAll(ExportFilePath, 644)
		if err != nil {
			return err
		}
	}

	err = export.ExportTasksToZip(filepath.Join(ExportFilePath, taskInfo.ExportFileName), exportTasks)
	if err != nil {
		return err
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

		auditTaskInfo, AuditRecord, err := d.SQLEAuditSQL(ctx, project.Name, dataExportTaskUid, v.DBServiceUid, v.DatabaseName, v.ExportSQL)
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

func (d *DataExportWorkflowUsecase) SQLEAuditSQL(ctx context.Context, projectName, taskId, dBServiceUid, dbName string, Sqls string) (*AuditTaskInfo, []*DataExportTaskRecord, error) {
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
	dbService, err := d.dbServiceRepo.GetDBService(ctx, dBServiceUid)
	if err != nil {
		return nil, nil, err
	}

	// 执行SQL审核 /v1/projects/{project_name}/tasks/audits
	auditReply := &GetAuditTaskResV1{}
	auditUri := fmt.Sprintf("/v1/projects/%s/tasks/audits", projectName)
	err = pkgHttp.POST(ctx, fmt.Sprintf("%s%s", sqleUrl, auditUri), header, CreateAuditTaskReqV1{
		InstanceName:   dbService.Name,
		InstanceSchema: dbName,
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

func (d *DataExportWorkflowUsecase) DownloadDataExportTask(ctx context.Context, userId string, req *dmsV1.DownloadDataExportTaskReq) (bool, string, error) {
	tasks, err := d.dataExportTaskRepo.GetDataExportTaskByIds(ctx, []string{req.DataExportTaskUid})
	if err != nil {
		return false, "", err
	}

	if len(tasks) == 0 {
		return false, "", fmt.Errorf("task does not exist")
	}

	if userId != tasks[0].CreateUserUID {
		return false, "", fmt.Errorf("current user not executable download")
	}

	if len(tasks[0].ExportFileName) == 0 {
		return false, "", fmt.Errorf("download file does not exist")
	}

	//exportFilename = "ip:port;filename.zip"
	splitResult := strings.Split(tasks[0].ExportFileName, ExportFilenameSeparate)
	reportHost, exportFilename := splitResult[0], splitResult[1]

	//proxy request other node to download files
	if d.reportHost != reportHost {
		return true, reportHost, nil
	}

	filename := filepath.Join(ExportFilePath, exportFilename)

	if _, err = os.Stat(filename); err != nil {
		return false, "", err
	} else if os.IsNotExist(err) {
		return false, "", fmt.Errorf("file %s does not exist", filename)
	}

	if tasks[0].ExportEndTime.Before(time.Now().Add(-24 * time.Hour)) {
		return false, "", fmt.Errorf("the file %s download has exceeded 24 hours", tasks[0].ExportFileName)
	}

	return false, filename, nil
}

func (d *DataExportWorkflowUsecase) RecycleWorkflow() {
	if d.clusterUsecase.IsClusterMode() && !d.clusterUsecase.IsLeader() {
		d.log.Info("current node is not leader,don't to recycle workflow")
		return
	}
	// 回收工单
	expiredWorkflowUids := make([]string, 0)
	for pageIndex, pageSize := 1, 100; ; pageIndex++ {
		workflows, _, err := d.repo.ListDataExportWorkflows(context.Background(), &ListWorkflowsOption{
			PageNumber:   uint32(pageIndex),
			LimitPerPage: uint32(pageSize),
			OrderBy:      WorkflowFieldCreateTime,
			FilterBy: []pkgConst.FilterCondition{
				{
					Field:    string(WorkflowFieldCreateTime),
					Operator: pkgConst.FilterOperatorLessThanOrEqual,
					Value:    time.Now().Add(-time.Hour * 365 * 24),
				},
			},
		})
		if err != nil {
			d.log.Error(err)
		}
		for _, workflow := range workflows {
			expiredWorkflowUids = append(expiredWorkflowUids, workflow.UID)
		}
		if len(workflows) < pageSize {
			break
		}
	}
	if len(expiredWorkflowUids) == 0 {
		return
	}

	err := d.repo.DeleteDataExportWorkflowsByIds(context.Background(), expiredWorkflowUids)
	if err != nil {
		d.log.Error(err)
	}
	d.log.Infof("delete expired workflow %v success ", expiredWorkflowUids)
}

func (d *DataExportWorkflowUsecase) RecycleDataExportTask() {
	if d.clusterUsecase.IsClusterMode() && !d.clusterUsecase.IsLeader() {
		d.log.Info("current node is not leader,don't to recycle data export task")
		return
	}
	// 回收无关联工单的数据导出任务
	err := d.dataExportTaskRepo.DeleteUnusedDataExportTasks(context.Background())
	if err != nil {
		d.log.Error(err)
		return
	}
	d.log.Infof("recycle data export task  success")
}

func (d *DataExportWorkflowUsecase) RecycleDataExportTaskFiles() {
	// 回收超时导出文件
	recycleDataExportTasks := make([]*DataExportTask, 0)
	for pageIndex, pageSize := 1, 100; ; pageIndex++ {
		dataExportTasks, _, err := d.dataExportTaskRepo.ListDataExportTasks(context.Background(), &ListDataExportTaskOption{
			PageNumber:   uint32(pageIndex),
			LimitPerPage: uint32(pageSize),
			OrderBy:      DataExportTaskFieldExportEndTime,
			FilterBy: []pkgConst.FilterCondition{
				{
					Field:    string(DataExportTaskFieldExportEndTime),
					Operator: pkgConst.FilterOperatorLessThanOrEqual,
					Value:    time.Now().Add(-24 * time.Hour).String(),
				},
				{
					Field:    string(DataExportTaskFieldExportStatus),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    DataExportTaskStatusFinish.String(),
				},
			},
		})
		if err != nil {
			d.log.Error(err)
		}

		recycleDataExportTasks = append(recycleDataExportTasks, dataExportTasks...)
		if len(dataExportTasks) < pageSize {
			break
		}
	}

	for _, t := range recycleDataExportTasks {
		// 文件已被删除
		hostFileName := strings.Split(t.ExportFileName, ExportFilenameSeparate)
		if hostFileName[0] != d.reportHost {
			continue
		}
		realFileName := hostFileName[1]
		_, err := os.Stat(path.Join(ExportFilePath, realFileName))
		if err != nil {
			if !os.IsNotExist(err) {
				d.log.Error(err)
				continue
			}
		} else {
			err = os.Remove(filepath.Join(ExportFilePath, realFileName))
			if err != nil {
				d.log.Errorf("remove expired file %v , err: %v", realFileName, err)
				continue
			}
		}
		// 文件移除后更新字段
		err = d.dataExportTaskRepo.BatchUpdateDataExportTaskByIds(context.Background(), []string{t.UID}, map[string]interface{}{
			"export_status": DataExportTaskStatusFileDelted.String(),
		})
		if err != nil {
			d.log.Errorf("update data export task failed: %v", err)
		}
		d.log.Debugf("remove expired file %v success", t.ExportFileName)
	}

	d.log.Infof("recycle data export task file success")
}
