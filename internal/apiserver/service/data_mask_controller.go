package service

import (
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:operation GET /v1/dms/masking/rules Masking ListMaskingRules
//
// List masking rules.
//
// ---
// responses:
//   '200':
//     description: List masking rules successfully
//     schema:
//       "$ref": "#/definitions/ListMaskingRulesReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ListMaskingRules(c echo.Context) error {
	req := &aV1.ListMaskingRulesReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	reply, err := ctl.DMS.ListMaskingRules(c.Request().Context())
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/templates Masking ListMaskingTemplates
//
// List masking templates.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: the maximum count of masking templates to be returned, default is 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: the offset of masking templates to be returned, default is 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//   '200':
//     description: List masking templates successfully
//     schema:
//       "$ref": "#/definitions/ListMaskingTemplatesReply"
//   default:
//     description: Generic error response
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
// Add masking template.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: masking_template
//     description: masking template info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddMaskingTemplateReq"
//
// responses:
//   '200':
//     description: Add masking template successfully
//     schema:
//       "$ref": "#/definitions/AddMaskingTemplateReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddMaskingTemplate(c echo.Context) error {
	req := &aV1.AddMaskingTemplateReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if err := ctl.DMS.AddMaskingTemplate(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.AddMaskingTemplateReply{})
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/masking/templates/{template_id} Masking UpdateMaskingTemplate
//
// Update masking template.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: template_id
//     description: masking template id
//     in: path
//     required: true
//     type: integer
//   - name: masking_template
//     description: masking template info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateMaskingTemplateReq"
//
// responses:
//   '200':
//     description: Update masking template successfully
//     schema:
//       "$ref": "#/definitions/UpdateMaskingTemplateReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateMaskingTemplate(c echo.Context) error {
	req := &aV1.UpdateMaskingTemplateReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if err := ctl.DMS.UpdateMaskingTemplate(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.UpdateMaskingTemplateReply{})
}

// swagger:operation DELETE /v1/dms/projects/{project_uid}/masking/templates/{template_id} Masking DeleteMaskingTemplate
//
// Delete masking template.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: template_id
//     description: masking template id
//     in: path
//     required: true
//     type: integer
//
// responses:
//   '200':
//     description: Delete masking template successfully
//     schema:
//       "$ref": "#/definitions/DeleteMaskingTemplateReply"
//   default:
//     description: Generic error response
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
// List db services that can create sensitive data discovery task.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: the maximum count of db services to be returned, default is 100
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: the offset of db services to be returned, default is 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: keywords
//     description: fuzzy search keywords for db service name
//     in: query
//     required: false
//     type: string
//
// responses:
//   '200':
//     description: List creatable db services for masking task successfully
//     schema:
//       "$ref": "#/definitions/ListCreatableDBServicesForMaskingTaskReply"
//   default:
//     description: Generic error response
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
// List sensitive data discovery tasks.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: the maximum count of tasks to be returned, default is 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: the offset of tasks to be returned, default is 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//   '200':
//     description: List sensitive data discovery tasks successfully
//     schema:
//       "$ref": "#/definitions/ListSensitiveDataDiscoveryTasksReply"
//   default:
//     description: Generic error response
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
// Add sensitive data discovery task.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: task
//     description: sensitive data discovery task info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataDiscoveryTaskReq"
//
// responses:
//   '200':
//     description: Add sensitive data discovery task successfully
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataDiscoveryTaskReply"
//   default:
//     description: Generic error response
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
// Update sensitive data discovery task.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: task_id
//     description: sensitive data discovery task id
//     in: path
//     required: true
//     type: integer
//   - name: task
//     description: sensitive data discovery task info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/UpdateSensitiveDataDiscoveryTaskReq"
//
// responses:
//   '200':
//     description: Update sensitive data discovery task successfully
//     schema:
//       "$ref": "#/definitions/UpdateSensitiveDataDiscoveryTaskReply"
//   default:
//     description: Generic error response
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
// Delete sensitive data discovery task.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: task_id
//     description: sensitive data discovery task id
//     in: path
//     required: true
//     type: integer
//
// responses:
//   '200':
//     description: Delete sensitive data discovery task successfully
//     schema:
//       "$ref": "#/definitions/DeleteSensitiveDataDiscoveryTaskReply"
//   default:
//     description: Generic error response
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
// List sensitive data discovery task histories.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: task_id
//     description: sensitive data discovery task id
//     in: path
//     required: true
//     type: integer
//   - name: page_size
//     description: the maximum count of histories to be returned, default is 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: the offset of histories to be returned, default is 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//   '200':
//     description: List sensitive data discovery task histories successfully
//     schema:
//       "$ref": "#/definitions/ListSensitiveDataDiscoveryTaskHistoriesReply"
//   default:
//     description: Generic error response
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
// Configure masking rules in batch.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: masking_rule_configs_req
//     description: masking rule configurations for batch create or update
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ConfigureMaskingRulesReq"
//
// responses:
//   '200':
//     description: Configure masking rules successfully
//     schema:
//       "$ref": "#/definitions/ConfigureMaskingRulesReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ConfigureMaskingRules(c echo.Context) error {
	req := &aV1.ConfigureMaskingRulesReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	if err := ctl.DMS.ConfigureMaskingRules(c.Request().Context(), req); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, &aV1.ConfigureMaskingRulesReply{})
}

// swagger:operation GET /v1/dms/projects/{project_uid}/masking/overview Masking GetMaskingOverviewTree
//
// Get masking overview tree.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: db_service_uid
//     description: data source id
//     in: query
//     required: true
//     type: string
//   - name: keywords
//     description: fuzzy search keyword for database name, table name, and column name
//     in: query
//     required: false
//     type: string
//   - name: masking_config_statuses
//     description: "masking config status filters, enum: CONFIGURED/PENDING_CONFIRM"
//     in: query
//     required: false
//     type: string
//
// responses:
//   '200':
//     description: Get masking overview tree successfully
//     schema:
//       "$ref": "#/definitions/GetMaskingOverviewTreeReply"
//   default:
//     description: Generic error response
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
// Get table column masking details.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: table_id
//     description: table id from masking overview tree
//     in: path
//     required: true
//     type: integer
//   - name: keywords
//     description: fuzzy search keyword for column name
//     in: query
//     required: false
//     type: string
//
// responses:
//   '200':
//     description: Get table column masking details successfully
//     schema:
//       "$ref": "#/definitions/GetTableColumnMaskingDetailsReply"
//   default:
//     description: Generic error response
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
// List pending approval requests.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: page_size
//     description: the maximum count of requests to be returned, default is 20
//     in: query
//     required: false
//     type: integer
//     format: uint32
//   - name: page_index
//     description: the offset of requests to be returned, default is 0
//     in: query
//     required: false
//     type: integer
//     format: uint32
//
// responses:
//   '200':
//     description: List pending approval requests successfully
//     schema:
//       "$ref": "#/definitions/ListPendingApprovalRequestsReply"
//   default:
//     description: Generic error response
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
// Get plaintext access request detail.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: request_id
//     description: approval request id
//     in: path
//     required: true
//     type: integer
//
// responses:
//   '200':
//     description: Get plaintext access request detail successfully
//     schema:
//       "$ref": "#/definitions/GetPlaintextAccessRequestDetailReply"
//   default:
//     description: Generic error response
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
// Process approval request.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: request_id
//     description: approval request id
//     in: path
//     required: true
//     type: integer
//   - name: action
//     description: process action info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ProcessApprovalRequestReq"
//
// responses:
//   '200':
//     description: Process approval request successfully
//     schema:
//       "$ref": "#/definitions/ProcessApprovalRequestReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ProcessApprovalRequest(c echo.Context) error {
	req := &aV1.ProcessApprovalRequestReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	return NewOkRespWithReply(c, &aV1.ProcessApprovalRequestReply{})
}
