package service

import (
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/labstack/echo/v4"
)

// swagger:operation GET /v1/dms/masking/templates Masking ListMaskingTemplates
//
// List masking templates.
//
// ---
// parameters:
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
	return NewOkRespWithReply(c, &aV1.ListMaskingTemplatesReply{})
}

// swagger:operation POST /v1/dms/masking/templates Masking AddMaskingTemplate
//
// Add masking template.
//
// ---
// parameters:
//   - name: masking_template
//     description: masking template info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddMaskingTemplate"
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
	return NewOkRespWithReply(c, &aV1.AddMaskingTemplateReply{})
}

// swagger:operation PUT /v1/dms/masking/templates/{template_id} Masking UpdateMaskingTemplate
//
// Update masking template.
//
// ---
// parameters:
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
//       "$ref": "#/definitions/UpdateMaskingTemplate"
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
	return NewOkRespWithReply(c, &aV1.UpdateMaskingTemplateReply{})
}

// swagger:operation DELETE /v1/dms/masking/templates/{template_id} Masking DeleteMaskingTemplate
//
// Delete masking template.
//
// ---
// parameters:
//   - name: template_id
//     description: masking template id
//     in: path
//     required: true
//     type: integer
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
	return NewOkRespWithReply(c, &aV1.DeleteMaskingTemplateReply{})
}

// swagger:operation GET /v1/dms/masking/sensitive-data-discovery-tasks Masking ListSensitiveDataDiscoveryTasks
//
// List sensitive data discovery tasks.
//
// ---
// parameters:
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
	return NewOkRespWithReply(c, &aV1.ListSensitiveDataDiscoveryTasksReply{})
}

// swagger:operation POST /v1/dms/masking/sensitive-data-discovery-tasks Masking AddSensitiveDataDiscoveryTask
//
// Add sensitive data discovery task.
//
// ---
// parameters:
//   - name: task
//     description: sensitive data discovery task info
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddSensitiveDataDiscoveryTask"
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
	return NewOkRespWithReply(c, &aV1.AddSensitiveDataDiscoveryTaskReply{})
}

// swagger:operation PUT /v1/dms/masking/sensitive-data-discovery-tasks/{task_id} Masking UpdateSensitiveDataDiscoveryTask
//
// Update sensitive data discovery task.
//
// ---
// parameters:
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
//       "$ref": "#/definitions/UpdateSensitiveDataDiscoveryTask"
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
	return NewOkRespWithReply(c, &aV1.UpdateSensitiveDataDiscoveryTaskReply{})
}

// swagger:operation GET /v1/dms/masking/sensitive-data-discovery-tasks/{task_id}/histories Masking ListSensitiveDataDiscoveryTaskHistories
//
// List sensitive data discovery task histories.
//
// ---
// parameters:
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
	return NewOkRespWithReply(c, &aV1.ListSensitiveDataDiscoveryTaskHistoriesReply{})
}

// swagger:operation PUT /v1/dms/masking/rule-configs Masking ConfigureMaskingRules
//
// Configure masking rules in batch.
//
// ---
// parameters:
//   - name: masking_rule_configs
//     description: masking rule configurations
//     in: body
//     required: true
//     schema:
//       type: object
//       properties:
//         masking_rule_configs:
//           type: array
//           items:
//             "$ref": "#/definitions/MaskingRuleConfig"
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
	return NewOkRespWithReply(c, &aV1.ConfigureMaskingRulesReply{})
}

// swagger:operation GET /v1/dms/masking/overview Masking GetMaskingOverviewTree
//
// Get masking overview tree.
//
// ---
// parameters:
//   - name: project_id
//     description: project id
//     in: query
//     required: true
//     type: integer
//   - name: db_service_id
//     description: data source id
//     in: query
//     required: true
//     type: integer
//   - name: search
//     description: fuzzy search keyword for database name, table name, and column name
//     in: query
//     required: false
//     type: string
//   - name: masking_config_statuses
//     description: "masking config status filters, enum: CONFIGURED/PENDING_CONFIRM"
//     in: query
//     required: false
//     type: array
//     items:
//       type: string
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
	return NewOkRespWithReply(c, &aV1.GetMaskingOverviewTreeReply{})
}

// swagger:operation GET /v1/dms/masking/tables/{table_id}/column-masking-details Masking GetTableColumnMaskingDetails
//
// Get table column masking details.
//
// ---
// parameters:
//   - name: table_id
//     description: table id from masking overview tree
//     in: path
//     required: true
//     type: integer
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
	return NewOkRespWithReply(c, &aV1.GetTableColumnMaskingDetailsReply{})
}

// swagger:operation GET /v1/dms/masking/approval-requests/pending Masking ListPendingApprovalRequests
//
// List pending approval requests.
//
// ---
// parameters:
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

// swagger:operation GET /v1/dms/masking/approval-requests/{request_id} Masking GetPlaintextAccessRequestDetail
//
// Get plaintext access request detail.
//
// ---
// parameters:
//   - name: request_id
//     description: approval request id
//     in: path
//     required: true
//     type: integer
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

// swagger:operation POST /v1/dms/masking/approval-requests/{request_id}/decisions Masking ProcessApprovalRequest
//
// Process approval request.
//
// ---
// parameters:
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
