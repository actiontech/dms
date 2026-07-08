package sql_workbench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	dbmodel "github.com/actiontech/dms/internal/dms/storage/model"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/labstack/echo/v4"
)

type streamExecuteRequestContext struct {
	sql             string
	datasourceID    string
	schemaName      string
	isExecuteAnyway bool
}

type workflowInstanceForCreatingTask struct {
	InstanceName   string `json:"instance_name"`
	InstanceSchema string `json:"instance_schema"`
}

type autoCreateAndExecuteWorkflowReq struct {
	Instances       []*workflowInstanceForCreatingTask `json:"instances"`
	ExecMode        string                             `json:"exec_mode"`
	FileOrderMethod string                             `json:"file_order_method"`
	Sql             string                             `json:"sql"`
	Subject         string                             `json:"workflow_subject"`
	Desc            string                             `json:"desc"`
}

type autoCreateAndExecuteWorkflowRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		WorkflowID     string `json:"workflow_id"`
		WorkFlowStatus string `json:"workflow_status"`
	} `json:"data"`
}

type streamExecuteWorkflowInfo struct {
	WorkflowID     string `json:"workflowId"`
	WorkflowStatus string `json:"workflowStatus"`
	ProjectName    string `json:"projectName"`
	ExecSuccess    bool   `json:"execSuccess"`
}

// StreamExecuteWorkflowInfo is exported for response serialization in sql_workbench_service.go.
type StreamExecuteWorkflowInfo = streamExecuteWorkflowInfo

func (sqlWorkbenchService *SqlWorkbenchService) isEnableWorkflowExec(dbService *biz.DBService) bool {
	if dbService.SQLEConfig == nil || dbService.SQLEConfig.SQLQueryConfig == nil {
		return false
	}
	return dbService.SQLEConfig.AuditEnabled && dbService.SQLEConfig.SQLQueryConfig.WorkflowExecEnabled
}

func (sqlWorkbenchService *SqlWorkbenchService) shouldExecuteByWorkflow(dbService *biz.DBService, auditResults []cloudbeaver.AuditSQLResV2) bool {
	if !sqlWorkbenchService.isEnableSQLAudit(dbService) || !sqlWorkbenchService.isEnableWorkflowExec(dbService) {
		return false
	}
	for _, result := range auditResults {
		if result.SQLType != "" && result.SQLType != "dql" {
			return true
		}
	}
	return false
}

func (sqlWorkbenchService *SqlWorkbenchService) checkWorkflowPermission(ctx context.Context, userUID string, dbService *biz.DBService) (bool, error) {
	canOpGlobal, err := sqlWorkbenchService.opPermissionVerifyUsecase.CanOpGlobal(ctx, userUID, true)
	if err != nil {
		return false, fmt.Errorf("check global op permission err: %v", err)
	}
	if canOpGlobal {
		return true, nil
	}
	opPermissions, err := sqlWorkbenchService.opPermissionVerifyUsecase.GetUserOpPermissionInProject(ctx, userUID, dbService.ProjectUID)
	if err != nil {
		return false, fmt.Errorf("get user op permission in project err: %v", err)
	}

	requiredPermissions := map[string]struct{}{
		constant.UIDOfOpPermissionCreateWorkflow:  {},
		constant.UIDOfOpPermissionAuditWorkflow:   {},
		constant.UIDOfOpPermissionExecuteWorkflow: {},
	}
	dbServicePermissions := make(map[string]struct{})

	for _, opPermission := range opPermissions {
		if opPermission.OpRangeType == biz.OpRangeTypeProject && opPermission.OpPermissionUID == constant.UIDOfOpPermissionProjectAdmin {
			return true, nil
		}
		if opPermission.OpRangeType == biz.OpRangeTypeDBService {
			if _, isRequired := requiredPermissions[opPermission.OpPermissionUID]; isRequired {
				for _, rangeUID := range opPermission.RangeUIDs {
					if rangeUID == dbService.UID {
						dbServicePermissions[opPermission.OpPermissionUID] = struct{}{}
						if len(dbServicePermissions) == len(requiredPermissions) {
							return true, nil
						}
						break
					}
				}
			}
		}
	}
	return len(dbServicePermissions) == len(requiredPermissions), nil
}

func (sqlWorkbenchService *SqlWorkbenchService) getSQLEURL(ctx context.Context) (string, error) {
	target, err := sqlWorkbenchService.proxyTargetRepo.GetProxyTargetByName(ctx, _const.SqleComponentName)
	if err != nil {
		return "", fmt.Errorf("get sqle proxy target failed: %v", err)
	}
	return target.URL.String(), nil
}

func (sqlWorkbenchService *SqlWorkbenchService) autoCreateAndExecuteWorkflow(ctx context.Context, projectName string, dbService *biz.DBService, sql string, instanceSchema string) (*autoCreateAndExecuteWorkflowRes, error) {
	sqleURL, err := sqlWorkbenchService.getSQLEURL(ctx)
	if err != nil {
		return nil, err
	}

	project, err := sqlWorkbenchService.projectUsecase.GetProject(ctx, dbService.ProjectUID)
	if err != nil {
		return nil, fmt.Errorf("get project failed: %v", err)
	}
	if projectName == "" {
		projectName = project.Name
	}

	instances := []*workflowInstanceForCreatingTask{{
		InstanceName:   dbService.Name,
		InstanceSchema: instanceSchema,
	}}
	req := autoCreateAndExecuteWorkflowReq{
		Instances:       instances,
		ExecMode:        "sqls",
		FileOrderMethod: "",
		Sql:             sql,
		Subject:         fmt.Sprintf("工作台工单_%s_%s", dbService.Name, time.Now().Format("20060102150405")),
		Desc:            "通过工作台执行非DQL类型的SQL时，自动创建的工单",
	}

	instancesJSON, err := json.Marshal(req.Instances)
	if err != nil {
		return nil, fmt.Errorf("marshal instances failed: %v", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	fields := map[string]string{
		"instances":         string(instancesJSON),
		"exec_mode":         req.ExecMode,
		"file_order_method": req.FileOrderMethod,
		"sql":               req.Sql,
		"workflow_subject":  req.Subject,
		"desc":              req.Desc,
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			writer.Close()
			return nil, fmt.Errorf("write field %s failed: %v", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer failed: %v", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/workflows/auto_create_and_execute", sqleURL, projectName)
	headers := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}
	var reply autoCreateAndExecuteWorkflowRes
	if err := pkgHttp.Call(ctx, http.MethodPost, url, headers, writer.FormDataContentType(), requestBody.Bytes(), &reply); err != nil {
		return nil, fmt.Errorf("request sqle failed: %v", err)
	}
	if reply.Code != 0 {
		return nil, fmt.Errorf("sqle returned error code %d: %s", reply.Code, reply.Message)
	}
	return &reply, nil
}

func (sqlWorkbenchService *SqlWorkbenchService) executeNonDQLByWorkflow(
	c echo.Context,
	userID string,
	execCtx *streamExecuteRequestContext,
	auditResult *cloudbeaver.AuditSQLReply,
	dbService *biz.DBService,
) error {
	ctx := c.Request().Context()

	hasPermission, err := sqlWorkbenchService.checkWorkflowPermission(ctx, userID, dbService)
	if err != nil {
		return sqlWorkbenchService.buildWorkflowErrorResponse(c, fmt.Sprintf("check workflow permission failed: %v", err))
	}
	if !hasPermission {
		return sqlWorkbenchService.buildWorkflowErrorResponse(c, "用户没有数据源上的创建、审批、上线工单权限")
	}

	project, err := sqlWorkbenchService.projectUsecase.GetProject(ctx, dbService.ProjectUID)
	if err != nil {
		return sqlWorkbenchService.buildWorkflowErrorResponse(c, fmt.Sprintf("get project failed: %v", err))
	}

	workflowRes, err := sqlWorkbenchService.autoCreateAndExecuteWorkflow(ctx, project.Name, dbService, execCtx.sql, execCtx.schemaName)
	if err != nil {
		sqlWorkbenchService.log.Errorf("auto create and execute workflow failed: %v", err)
		return sqlWorkbenchService.buildWorkflowErrorResponse(c, fmt.Sprintf("auto create and execute workflow failed: %v", err))
	}
	sqlWorkbenchService.log.Infof("auto create and execute workflow, workflow_id: %s, status: %s",
		workflowRes.Data.WorkflowID, workflowRes.Data.WorkFlowStatus)

	isExecFailed := !strings.Contains(workflowRes.Data.WorkFlowStatus, "finished")
	if err := sqlWorkbenchService.saveOpLogForWorkflow(c, userID, execCtx, auditResult, dbService, workflowRes.Data.WorkflowID, isExecFailed); err != nil {
		sqlWorkbenchService.log.Errorf("save operation log for workflow failed: %v", err)
	}

	return sqlWorkbenchService.buildWorkflowExecuteResponse(c, project.Name, workflowRes.Data.WorkflowID, workflowRes.Data.WorkFlowStatus, !isExecFailed)
}

func (sqlWorkbenchService *SqlWorkbenchService) buildWorkflowExecuteResponse(
	c echo.Context,
	projectName, workflowID, workflowStatus string,
	execSuccess bool,
) error {
	response := StreamExecuteResponse{
		Data: StreamExecuteData{
			ApprovalRequired: false,
			LogicalSQL:       false,
			RequestID:        nil,
			SQLs:             []StreamExecuteSQLItem{},
			WorkflowInfo: &StreamExecuteWorkflowInfo{
				WorkflowID:     workflowID,
				WorkflowStatus: workflowStatus,
				ProjectName:    projectName,
				ExecSuccess:    execSuccess,
			},
		},
		DurationMillis: 0,
		HTTPStatus:     "OK",
		RequestID:      fmt.Sprintf("dms-workflow-%d", time.Now().UnixNano()),
		Server:         "DMS",
		Successful:     true,
		Timestamp:      float64(time.Now().Unix()),
		TraceID:        c.Response().Header().Get("X-Trace-ID"),
	}
	return c.JSON(http.StatusOK, response)
}

func (sqlWorkbenchService *SqlWorkbenchService) buildWorkflowErrorResponse(c echo.Context, message string) error {
	response := StreamExecuteResponse{
		Data: StreamExecuteData{
			ApprovalRequired: false,
			LogicalSQL:       false,
			RequestID:        nil,
			SQLs:             []StreamExecuteSQLItem{},
			ErrorMessage:     message,
		},
		DurationMillis: 0,
		HTTPStatus:     "OK",
		RequestID:      fmt.Sprintf("dms-workflow-err-%d", time.Now().UnixNano()),
		Server:         "DMS",
		Successful:     false,
		Timestamp:      float64(time.Now().Unix()),
		TraceID:        c.Response().Header().Get("X-Trace-ID"),
	}
	return c.JSON(http.StatusOK, response)
}

func (sqlWorkbenchService *SqlWorkbenchService) saveOpLogForWorkflow(
	c echo.Context,
	userID string,
	execCtx *streamExecuteRequestContext,
	auditResult *cloudbeaver.AuditSQLReply,
	dbService *biz.DBService,
	workflowID string,
	isExecFailed bool,
) error {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return err
	}

	sessionID := extractSessionID(c.Request().URL.Path)
	isAuditPass := auditResult != nil && auditResult.Data != nil && auditResult.Data.PassRate == 1
	execResult := biz.CbExecOpSuccess
	if isExecFailed {
		execResult = biz.CbExecOpFailure
	}

	var auditResults dbmodel.AuditResults
	if auditResult != nil && auditResult.Data != nil {
		for _, sqlResult := range auditResult.Data.SQLResults {
			auditResults = append(auditResults, sqlWorkbenchService.convertToAuditResults(&sqlResult)...)
		}
	}

	now := time.Now()
	cbOperationLog := biz.CbOperationLog{
		UID:          uid,
		OpPersonUID:  userID,
		OpTime:       &now,
		DBServiceUID: dbService.UID,
		OpType:       biz.CbOperationLogTypeSql,
		I18nOpDetail: i18nPkg.ConvertStr2I18nAsDefaultLang(execCtx.sql),
		OpSessionID:  &sessionID,
		ProjectID:    dbService.ProjectUID,
		OpHost:       c.RealIP(),
		AuditResults: auditResults,
		IsAuditPass:  &isAuditPass,
		ExecResult:   execResult,
		WorkflowID:   &workflowID,
	}
	return sqlWorkbenchService.cbOperationLogUsecase.SaveCbOperationLog(c.Request().Context(), &cbOperationLog)
}
