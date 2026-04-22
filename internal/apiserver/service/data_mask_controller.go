package service

import (
	"errors"

	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:route GET /v1/dms/projects/{project_uid}/masking/rules Masking ListMaskingRules
//
// 查询项目下的脱敏规则列表（内置与自定义）。
//
//	responses:
//	  200: body:ListMaskingRulesReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMaskingRules(c echo.Context) error {
	req := &aV1.ListMaskingRulesReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMaskingRules(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/templates Masking ListMaskingTemplates
//
// 查询脱敏模板列表。
//
//	responses:
//	  200: body:ListMaskingTemplatesReply
//	  default: body:GenericResp
func (ctl *DMSController) ListMaskingTemplates(c echo.Context) error {
	req := &aV1.ListMaskingTemplatesReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMaskingTemplates(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/db_services/{db_service_uid}/schemas/{schema_name}/tables/{table_name}/columns DBStructure ListTableColumns
//
// List table columns (internal API for lineage analysis).
//
//	responses:
//	  200: body:ListTableColumnsReply
//	  default: body:GenericResp
func (ctl *DMSController) ListTableColumns(c echo.Context) error {
	// 内部接口，仅允许sys/admin用户访问
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	if currentUserUid != pkgConst.UIDOfUserSys && currentUserUid != pkgConst.UIDOfUserAdmin {
		return NewErrResp(c, errors.New("insufficient permission"), apiError.UnauthorizedErr)
	}

	req := new(aV1.ListTableColumnsReq)
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListTableColumns(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/templates Masking AddMaskingTemplate
//
// 新增脱敏模板。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: masking_template
//     description: 脱敏模板信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddMaskingTemplateReq"
//
// responses:
//
//   '200':
//     description: 成功新增脱敏模板
//     schema:
//       "$ref": "#/definitions/AddMaskingTemplateReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddMaskingTemplate(c echo.Context) error {
	req := &aV1.AddMaskingTemplateReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if req.MaskingTemplate != nil {
		if err := req.MaskingTemplate.NormalizeRuleIDs(); err != nil {
			return NewErrResp(c, err, apiError.BadRequestErr)
		}
	}

	if err := ctl.DMS.AddMaskingTemplate(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.AddMaskingTemplateReply{})
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/masking/templates/{template_id} Masking UpdateMaskingTemplate
//
// 更新脱敏模板。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: template_id
//     description: 脱敏模板 ID
//     in: path
//     required: true
//     type: integer
//   - name: masking_template
//     description: 脱敏模板信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateMaskingTemplateReq"
//
// responses:
//
//   '200':
//     description: 成功更新脱敏模板
//     schema:
//       "$ref": "#/definitions/UpdateMaskingTemplateReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateMaskingTemplate(c echo.Context) error {
	req := &aV1.UpdateMaskingTemplateReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if req.MaskingTemplate != nil {
		if err := req.MaskingTemplate.NormalizeRuleIDs(); err != nil {
			return NewErrResp(c, err, apiError.BadRequestErr)
		}
	}

	if err := ctl.DMS.UpdateMaskingTemplate(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.UpdateMaskingTemplateReply{})
}

// swagger:operation DELETE /v1/dms/projects/{project_uid}/masking/templates/{template_id} Masking DeleteMaskingTemplate
//
// 删除脱敏模板。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: template_id
//     description: 脱敏模板 ID
//     in: path
//     required: true
//     type: integer
//
// responses:
//
//   '200':
//     description: 成功删除脱敏模板
//     schema:
//       "$ref": "#/definitions/DeleteMaskingTemplateReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) DeleteMaskingTemplate(c echo.Context) error {
	req := &aV1.DeleteMaskingTemplateReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if err := ctl.DMS.DeleteMaskingTemplate(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.DeleteMaskingTemplateReply{})
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/creatable-db-services Masking ListCreatableDBServicesForMaskingTask
//
// 查询可用于创建敏感数据发现任务的数据源列表。
//
//	responses:
//	  200: body:ListCreatableDBServicesForMaskingTaskReply
//	  default: body:GenericResp
func (ctl *DMSController) ListCreatableDBServicesForMaskingTask(c echo.Context) error {
	req := &aV1.ListCreatableDBServicesForMaskingTaskReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.ListCreatableDBServicesForMaskingTask(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks Masking ListSensitiveDataDiscoveryTasks
//
// 查询敏感数据发现任务列表。
//
//	responses:
//	  200: body:ListSensitiveDataDiscoveryTasksReply
//	  default: body:GenericResp
func (ctl *DMSController) ListSensitiveDataDiscoveryTasks(c echo.Context) error {
	req := &aV1.ListSensitiveDataDiscoveryTasksReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListSensitiveDataDiscoveryTasks(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks Masking AddSensitiveDataDiscoveryTask
//
// 新增敏感数据发现任务。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: task
//     description: 敏感数据发现任务信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataDiscoveryTaskReq"
//
// responses:
//
//   '200':
//     description: 成功新增敏感数据发现任务
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataDiscoveryTaskReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddSensitiveDataDiscoveryTask(c echo.Context) error {
	req := &aV1.AddSensitiveDataDiscoveryTaskReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.AddSensitiveDataDiscoveryTask(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/{task_id} Masking UpdateSensitiveDataDiscoveryTask
//
// 更新敏感数据发现任务。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: task_id
//     description: 敏感数据发现任务 ID
//     in: path
//     required: true
//     type: integer
//   - name: task
//     description: 敏感数据发现任务信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateSensitiveDataDiscoveryTaskReq"
//
// responses:
//
//   '200':
//     description: 成功更新敏感数据发现任务
//     schema:
//       "$ref": "#/definitions/UpdateSensitiveDataDiscoveryTaskReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateSensitiveDataDiscoveryTask(c echo.Context) error {
	req := &aV1.UpdateSensitiveDataDiscoveryTaskReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.UpdateSensitiveDataDiscoveryTask(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation DELETE /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/{task_id} Masking DeleteSensitiveDataDiscoveryTask
//
// 删除敏感数据发现任务。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: task_id
//     description: 敏感数据发现任务 ID
//     in: path
//     required: true
//     type: integer
//
// responses:
//
//   '200':
//     description: 成功删除敏感数据发现任务
//     schema:
//       "$ref": "#/definitions/DeleteSensitiveDataDiscoveryTaskReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) DeleteSensitiveDataDiscoveryTask(c echo.Context) error {
	req := &aV1.DeleteSensitiveDataDiscoveryTaskReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if err := ctl.DMS.DeleteSensitiveDataDiscoveryTask(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, &aV1.DeleteSensitiveDataDiscoveryTaskReply{})
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/{task_id}/histories Masking ListSensitiveDataDiscoveryTaskHistories
//
// 查询敏感数据发现任务执行历史。
//
//	responses:
//	  200: body:ListSensitiveDataDiscoveryTaskHistoriesReply
//	  default: body:GenericResp
func (ctl *DMSController) ListSensitiveDataDiscoveryTaskHistories(c echo.Context) error {
	req := &aV1.ListSensitiveDataDiscoveryTaskHistoriesReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListSensitiveDataDiscoveryTaskHistories(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/masking/rule-configs Masking ConfigureMaskingRules
//
// 批量配置脱敏规则。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: masking_rule_configs_req
//     description: 批量创建或更新脱敏规则的配置
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ConfigureMaskingRulesReq"
//
// responses:
//
//   '200':
//     description: 成功批量配置脱敏规则
//     schema:
//       "$ref": "#/definitions/ConfigureMaskingRulesReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ConfigureMaskingRules(c echo.Context) error {
	req := &aV1.ConfigureMaskingRulesReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	if err := ctl.DMS.ConfigureMaskingRules(c.Request().Context(), req, currentUserUid); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.ConfigureMaskingRulesReply{})
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/overview Masking GetMaskingOverviewTree
//
// 获取脱敏概览树。
//
//	responses:
//	  200: body:GetMaskingOverviewTreeReply
//	  default: body:GenericResp
func (ctl *DMSController) GetMaskingOverviewTree(c echo.Context) error {
	req := &aV1.GetMaskingOverviewTreeReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	reply, err := ctl.DMS.GetMaskingOverviewTree(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/tables/{table_id}/column-masking-details Masking GetTableColumnMaskingDetails
//
// 获取表字段脱敏详情。
//
//	responses:
//	  200: body:GetTableColumnMaskingDetailsReply
//	  default: body:GenericResp
func (ctl *DMSController) GetTableColumnMaskingDetails(c echo.Context) error {
	req := &aV1.GetTableColumnMaskingDetailsReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.GetTableColumnMaskingDetails(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/unmasking-workflows Masking CreateUnmaskingWorkflow
//
// Create unmasking workflow.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: unmasking_workflow
//     description: unmasking workflow info
//     in: body
//     required: true
//     schema:
//     "$ref": "#/definitions/CreateUnmaskingWorkflowReq"
//
// responses:
//
//	'200':
//	  description: Create unmasking workflow successfully
//	  schema:
//	    "$ref": "#/definitions/CreateUnmaskingWorkflowReply"
//	default:
//	  description: Generic error response
//	  schema:
//	    "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CreateUnmaskingWorkflow(c echo.Context) error {
	req := &aV1.CreateUnmaskingWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	reply, err := ctl.DMS.CreateUnmaskingWorkflow(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/unmasking-workflows Masking ListUnmaskingWorkflows
//
// List unmasking workflows.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: the maximum count of workflows to be returned
//     in: query
//     required: true
//     type: integer
//     format: uint32
//   - name: page_index
//     description: the offset of workflows to be returned, default is 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: filter_by_approval_status
//     description: filter the approval status
//     in: query
//     required: false
//     type: string
//     enum: [pending, approved, rejected, cancelled]
//   - name: filter_by_usage_status
//     description: filter the usage status
//     in: query
//     required: false
//     type: string
//     enum: [unviewed, viewed]
//   - name: filter_by_db_service_uid
//     description: filter db_service id
//     in: query
//     required: false
//     type: string
//
// responses:
//
//	'200':
//	  description: List unmasking workflows successfully
//	  schema:
//	    "$ref": "#/definitions/ListUnmaskingWorkflowsReply"
//	default:
//	  description: Generic error response
//	  schema:
//	    "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListUnmaskingWorkflows(c echo.Context) error {
	req := &aV1.ListUnmaskingWorkflowsReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	reply, err := ctl.DMS.ListUnmaskingWorkflows(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/unmasking-workflows/{workflow_id} Masking GetUnmaskingWorkflow
//
// Get unmasking workflow detail.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: workflow_id
//     description: workflow id
//     in: path
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	  description: Get unmasking workflow detail successfully
//	  schema:
//	    "$ref": "#/definitions/GetUnmaskingWorkflowReply"
//	default:
//	  description: Generic error response
//	  schema:
//	    "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GetUnmaskingWorkflow(c echo.Context) error {
	req := &aV1.GetUnmaskingWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	reply, err := ctl.DMS.GetUnmaskingWorkflow(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/unmasking-workflows/{workflow_id}/approve Masking ApproveUnmaskingWorkflow
//
// Approve unmasking workflow.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: workflow_id
//     description: workflow id
//     in: path
//     required: true
//     type: string
//   - name: approve_unmasking_workflow
//     description: approve unmasking workflow info
//     in: body
//     required: true
//     schema:
//     "$ref": "#/definitions/ApproveUnmaskingWorkflow"
//
// responses:
//
//	'200':
//	  description: Approve unmasking workflow successfully
//	  schema:
//	    "$ref": "#/definitions/ApproveUnmaskingWorkflowReply"
//	default:
//	  description: Generic error response
//	  schema:
//	    "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ApproveUnmaskingWorkflow(c echo.Context) error {
	req := &aV1.ApproveUnmaskingWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	err = ctl.DMS.ApproveUnmaskingWorkflow(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.ApproveUnmaskingWorkflowReply{})
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/unmasking-workflows/{workflow_id}/reject Masking RejectUnmaskingWorkflow
//
// Reject unmasking workflow.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: workflow_id
//     description: workflow id
//     in: path
//     required: true
//     type: string
//   - name: reject_unmasking_workflow
//     description: reject unmasking workflow info
//     in: body
//     required: true
//     schema:
//     "$ref": "#/definitions/RejectUnmaskingWorkflow"
//
// responses:
//
//	'200':
//	  description: Reject unmasking workflow successfully
//	  schema:
//	    "$ref": "#/definitions/RejectUnmaskingWorkflowReply"
//	default:
//	  description: Generic error response
//	  schema:
//	    "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) RejectUnmaskingWorkflow(c echo.Context) error {
	req := &aV1.RejectUnmaskingWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	err = ctl.DMS.RejectUnmaskingWorkflow(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.RejectUnmaskingWorkflowReply{})
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/unmasking-workflows/{workflow_id}/cancel Masking CancelUnmaskingWorkflow
//
// Cancel unmasking workflow.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: workflow_id
//     description: workflow id
//     in: path
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	  description: Cancel unmasking workflow successfully
//	  schema:
//	    "$ref": "#/definitions/CancelUnmaskingWorkflowReply"
//	default:
//	  description: Generic error response
//	  schema:
//	    "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CancelUnmaskingWorkflow(c echo.Context) error {
	req := &aV1.CancelUnmaskingWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}

	err = ctl.DMS.CancelUnmaskingWorkflow(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.CancelUnmaskingWorkflowReply{})
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/rules/{rule_id} Masking GetMaskingRuleDetail
//
// 获取脱敏规则详情。
//
//	responses:
//	  200: body:GetMaskingRuleDetailReply
//	  default: body:GenericResp
func (ctl *DMSController) GetMaskingRuleDetail(c echo.Context) error {
	req := &aV1.GetMaskingRuleDetailReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetMaskingRuleDetail(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/rules Masking AddMaskingRule
//
// 新增脱敏规则。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: rule
//     description: 脱敏规则信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddMaskingRuleReq"
//
// responses:
//
//   '200':
//     description: 成功新增脱敏规则
//     schema:
//       "$ref": "#/definitions/AddMaskingRuleReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddMaskingRule(c echo.Context) error {
	req := &aV1.AddMaskingRuleReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.AddMaskingRule(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/masking/rules/{rule_id} Masking UpdateMaskingRule
//
// 更新脱敏规则。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: rule_id
//     description: 脱敏规则 ID
//     in: path
//     required: true
//     type: integer
//   - name: rule
//     description: 脱敏规则更新信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateMaskingRuleReq"
//
// responses:
//
//   '200':
//     description: 成功更新脱敏规则
//     schema:
//       "$ref": "#/definitions/UpdateMaskingRuleReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateMaskingRule(c echo.Context) error {
	req := &aV1.UpdateMaskingRuleReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.UpdateMaskingRule(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation DELETE /v1/dms/projects/{project_uid}/masking/rules/{rule_id} Masking DeleteMaskingRule
//
// 删除脱敏规则。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: rule_id
//     description: 脱敏规则 ID
//     in: path
//     required: true
//     type: integer
//
// responses:
//
//   '200':
//     description: 成功删除脱敏规则
//     schema:
//       "$ref": "#/definitions/DeleteMaskingRuleReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) DeleteMaskingRule(c echo.Context) error {
	req := &aV1.DeleteMaskingRuleReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if err := ctl.DMS.DeleteMaskingRule(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, &aV1.DeleteMaskingRuleReply{})
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/sensitive-types Masking ListSensitiveTypes
//
// 查询敏感数据类型列表（内置与自定义）。
//
//	responses:
//	  200: body:ListSensitiveTypesReply
//	  default: body:GenericResp
func (ctl *DMSController) ListSensitiveTypes(c echo.Context) error {
	req := &aV1.ListSensitiveTypesReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListSensitiveTypes(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/sensitive-types Masking AddSensitiveDataType
//
// 创建自定义敏感数据类型（不创建脱敏规则）。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: body
//     description: 敏感数据类型定义
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataTypeReq"
//
// responses:
//
//   '200':
//     description: 成功创建敏感数据类型
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataTypeReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddSensitiveDataType(c echo.Context) error {
	req := &aV1.AddSensitiveDataTypeReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.AddSensitiveDataType(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/masking/sensitive-types/{sensitive_data_type_id} Masking UpdateSensitiveDataType
//
// 更新自定义敏感数据类型的识别条件（字段关键词、抽样正则、示例数据）。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: sensitive_data_type_id
//     description: 敏感数据类型 ID（自定义类型的主键）
//     in: path
//     required: true
//     type: integer
//   - name: body
//     description: 更新内容
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateSensitiveDataTypeReq"
//
// responses:
//
//   '200':
//     description: 成功更新敏感数据类型
//     schema:
//       "$ref": "#/definitions/UpdateSensitiveDataTypeReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateSensitiveDataType(c echo.Context) error {
	req := &aV1.UpdateSensitiveDataTypeReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.UpdateSensitiveDataType(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation DELETE /v1/dms/projects/{project_uid}/masking/sensitive-types/{sensitive_data_type_id} Masking DeleteSensitiveDataType
//
// 删除自定义敏感数据类型（未被任何脱敏规则引用时方可删除）。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: sensitive_data_type_id
//     description: 敏感数据类型 ID（自定义类型的主键）
//     in: path
//     required: true
//     type: integer
//
// responses:
//
//   '200':
//     description: 成功删除敏感数据类型
//     schema:
//       "$ref": "#/definitions/DeleteSensitiveDataTypeReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) DeleteSensitiveDataType(c echo.Context) error {
	req := &aV1.DeleteSensitiveDataTypeReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	if err := ctl.DMS.DeleteSensitiveDataType(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, &aV1.DeleteSensitiveDataTypeReply{})
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/sensitive-types/match-test Masking TestSensitiveDataTypeMatch
//
// 测试样例值是否匹配敏感数据类型规则（关键词与/或样例正则）。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: body
//     description: 匹配测试请求体
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/TestSensitiveDataTypeMatchReq"
//
// responses:
//
//   '200':
//     description: 匹配测试完成
//     schema:
//       "$ref": "#/definitions/TestSensitiveDataTypeMatchReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) TestSensitiveDataTypeMatch(c echo.Context) error {
	req := &aV1.TestSensitiveDataTypeMatchReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.TestSensitiveDataTypeMatch(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/preview Masking PreviewMaskingEffect
//
// 预览脱敏效果。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: preview_req
//     description: 脱敏效果预览请求体
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/PreviewMaskingEffectReq"
//
// responses:
//
//   '200':
//     description: 成功返回脱敏效果预览
//     schema:
//       "$ref": "#/definitions/PreviewMaskingEffectReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) PreviewMaskingEffect(c echo.Context) error {
	req := &aV1.PreviewMaskingEffectReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.PreviewMaskingEffect(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/db-service-schemas Masking ListDBServiceSchemasForMaskingTask
//
// 查询指定数据源下可用于配置敏感数据发现任务的 schema（数据库）列表。
//
//	responses:
//	  200: body:ListDBServiceSchemasForMaskingTaskReply
//	  default: body:GenericResp
func (ctl *DMSController) ListDBServiceSchemasForMaskingTask(c echo.Context) error {
	req := &aV1.ListDBServiceSchemasForMaskingTaskReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListDBServiceSchemasForMaskingTask(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/db-service-tables Masking ListDBServiceTablesForMaskingTask
//
// 查询指定数据源与 schema 下可用于配置敏感数据发现任务的表列表。
//
//	responses:
//	  200: body:ListDBServiceTablesForMaskingTaskReply
//	  default: body:GenericResp
func (ctl *DMSController) ListDBServiceTablesForMaskingTask(c echo.Context) error {
	req := &aV1.ListDBServiceTablesForMaskingTaskReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListDBServiceTablesForMaskingTask(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}
