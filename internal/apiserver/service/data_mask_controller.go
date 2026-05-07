package service

import (
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/rules Masking ListMaskingRules
//
// 查询项目下的脱敏规则列表（内置与自定义）。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//
// responses:
//
//   '200':
//     description: 成功返回脱敏规则列表
//     schema:
//       "$ref": "#/definitions/ListMaskingRulesReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/templates Masking ListMaskingTemplates
//
// 查询脱敏模板列表。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: 单次返回的脱敏模板数量上限，默认 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: 返回结果的偏移量，默认 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//
//   '200':
//     description: 成功返回脱敏模板列表
//     schema:
//       "$ref": "#/definitions/ListMaskingTemplatesReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/creatable-db-services Masking ListCreatableDBServicesForMaskingTask
//
// 查询可用于创建敏感数据发现任务的数据源列表。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: 单次返回的数据源数量上限，默认 100
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: 数据源列表的偏移量，默认 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: keywords
//     description: 按数据源名称模糊搜索的关键词
//     in: query
//     required: false
//     type: string
//
// responses:
//
//   '200':
//     description: 成功返回可创建敏感数据发现任务的数据源列表
//     schema:
//       "$ref": "#/definitions/ListCreatableDBServicesForMaskingTaskReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks Masking ListSensitiveDataDiscoveryTasks
//
// 查询敏感数据发现任务列表。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: 单次返回的任务数量上限，默认 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: 任务列表的偏移量，默认 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//
//   '200':
//     description: 成功返回敏感数据发现任务列表
//     schema:
//       "$ref": "#/definitions/ListSensitiveDataDiscoveryTasksReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/sensitive-data-discovery-tasks/{task_id}/histories Masking ListSensitiveDataDiscoveryTaskHistories
//
// 查询敏感数据发现任务执行历史。
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
//   - name: page_size
//     description: 单次返回的历史记录数量上限，默认 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: 历史记录的偏移量，默认 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//
//   '200':
//     description: 成功返回敏感数据发现任务执行历史
//     schema:
//       "$ref": "#/definitions/ListSensitiveDataDiscoveryTaskHistoriesReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/overview Masking GetMaskingOverviewTree
//
// 获取脱敏概览树。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: db_service_uid
//     description: 数据源 UID
//     in: query
//     required: true
//     type: string
//   - name: keywords
//     description: 按库名、表名、列名模糊搜索的关键词
//     in: query
//     required: false
//     type: string
//   - name: masking_config_statuses
//     description: "脱敏配置状态过滤，枚举：CONFIGURED/PENDING_CONFIRM"
//     in: query
//     required: false
//     type: string
//
// responses:
//
//   '200':
//     description: 成功返回脱敏概览树
//     schema:
//       "$ref": "#/definitions/GetMaskingOverviewTreeReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/tables/{table_id}/column-masking-details Masking GetTableColumnMaskingDetails
//
// 获取表字段脱敏详情。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: table_id
//     description: 脱敏概览树中的表 ID
//     in: path
//     required: true
//     type: integer
//   - name: keywords
//     description: 按列名模糊搜索的关键词
//     in: query
//     required: false
//     type: string
//
// responses:
//
//   '200':
//     description: 成功返回表字段脱敏详情
//     schema:
//       "$ref": "#/definitions/GetTableColumnMaskingDetailsReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/approval-requests/pending Masking ListPendingApprovalRequests
//
// 查询待审批申请列表。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: 单次返回的申请数量上限，默认 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: 申请列表的偏移量，默认 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//
//   '200':
//     description: 成功返回待审批申请列表
//     schema:
//       "$ref": "#/definitions/ListPendingApprovalRequestsReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListPendingApprovalRequests(c echo.Context) error {
	req := &aV1.ListPendingApprovalRequestsReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	return NewOkRespWithReply(c, &aV1.ListPendingApprovalRequestsReply{})
}

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/approval-requests/{request_id} Masking GetPlaintextAccessRequestDetail
//
// 获取明文访问申请详情。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: request_id
//     description: 审批申请 ID
//     in: path
//     required: true
//     type: integer
//
// responses:
//
//   '200':
//     description: 成功返回明文访问申请详情
//     schema:
//       "$ref": "#/definitions/GetPlaintextAccessRequestDetailReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GetPlaintextAccessRequestDetail(c echo.Context) error {
	req := &aV1.GetPlaintextAccessRequestDetailReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	return NewOkRespWithReply(c, &aV1.GetPlaintextAccessRequestDetailReply{})
}

// swagger:operation POST /v1/dms/projects/{project_uid}/masking/approval-requests/{request_id}/decisions Masking ProcessApprovalRequest
//
// 处理审批申请。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//   - name: request_id
//     description: 审批申请 ID
//     in: path
//     required: true
//     type: integer
//   - name: action
//     description: 处理动作信息
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ProcessApprovalRequestReq"
//
// responses:
//
//   '200':
//     description: 成功处理审批申请
//     schema:
//       "$ref": "#/definitions/ProcessApprovalRequestReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ProcessApprovalRequest(c echo.Context) error {
	req := &aV1.ProcessApprovalRequestReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	return NewOkRespWithReply(c, &aV1.ProcessApprovalRequestReply{})
}

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/rules/{rule_id} Masking GetMaskingRuleDetail
//
// 获取脱敏规则详情。
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
//     description: 成功返回脱敏规则详情
//     schema:
//       "$ref": "#/definitions/GetMaskingRuleDetailReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/sensitive-types Masking ListSensitiveTypes
//
// 查询敏感数据类型列表（内置与自定义）。
//
// ---
// parameters:
//   - name: project_uid
//     description: 项目 UID
//     in: path
//     required: true
//     type: string
//
// responses:
//
//   '200':
//     description: 成功返回敏感数据类型列表
//     schema:
//       "$ref": "#/definitions/ListSensitiveTypesReply"
//   default:
//     description: 通用错误响应
//     schema:
//       "$ref": "#/definitions/GenericResp"
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
