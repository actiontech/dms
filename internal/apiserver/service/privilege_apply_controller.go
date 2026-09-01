package service

import (
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:operation GET /v1/dms/projects/{project_uid}/privilege-apply-workflows/assignees PrivilegeApply GetPrivilegeApplyAssignees
//
// 预检提权申请审批人。
//
// ---
// parameters:
//   - name: project_uid
//     in: path
//     required: true
//     type: string
//   - name: db_service_uid
//     in: query
//     required: true
//     type: string
//
// responses:
//   '200':
//     description: Get privilege apply assignees successfully
//     schema:
//       "$ref": "#/definitions/GetPrivilegeApplyAssigneesReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GetPrivilegeApplyAssignees(c echo.Context) error {
	req := &aV1.GetPrivilegeApplyAssigneesReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.GetPrivilegeApplyAssignees(c.Request().Context(), req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v1/dms/projects/{project_uid}/privilege-apply-workflows PrivilegeApply CreatePrivilegeApplyWorkflow
//
// 创建提权申请。
//
// ---
// parameters:
//   - name: project_uid
//     in: path
//     required: true
//     type: string
//   - name: privilege_apply_workflow
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/CreatePrivilegeApplyWorkflowReq"
//
// responses:
//   '200':
//     description: Create privilege apply workflow successfully
//     schema:
//       "$ref": "#/definitions/CreatePrivilegeApplyWorkflowReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CreatePrivilegeApplyWorkflow(c echo.Context) error {
	req := &aV1.CreatePrivilegeApplyWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUID, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}
	reply, err := ctl.DMS.CreatePrivilegeApplyWorkflow(c.Request().Context(), req, currentUserUID)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation GET /v1/dms/projects/{project_uid}/privilege-apply-workflows/{workflow_id} PrivilegeApply GetPrivilegeApplyWorkflow
//
// 查询提权申请详情（申请人只读）。
//
// ---
// parameters:
//   - name: project_uid
//     in: path
//     required: true
//     type: string
//   - name: workflow_id
//     in: path
//     required: true
//     type: string
//
// responses:
//   '200':
//     description: Get privilege apply workflow successfully
//     schema:
//       "$ref": "#/definitions/GetPrivilegeApplyWorkflowReply"
//   default:
//     description: Generic error response
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) GetPrivilegeApplyWorkflow(c echo.Context) error {
	req := &aV1.GetPrivilegeApplyWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUID, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}
	reply, err := ctl.DMS.GetPrivilegeApplyWorkflow(c.Request().Context(), req, currentUserUID)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

func (ctl *DMSController) ListPrivilegeApplyWorkflows(c echo.Context) error {
	req := &aV1.ListPrivilegeApplyWorkflowsReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUID, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}
	reply, err := ctl.DMS.ListPrivilegeApplyWorkflows(c.Request().Context(), req, currentUserUID)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

func (ctl *DMSController) ApprovePrivilegeApplyWorkflow(c echo.Context) error {
	req := &aV1.ApprovePrivilegeApplyWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUID, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}
	if err := ctl.DMS.ApprovePrivilegeApplyWorkflow(c.Request().Context(), req, currentUserUID); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, &aV1.ApprovePrivilegeApplyWorkflowReply{})
}

func (ctl *DMSController) RejectPrivilegeApplyWorkflow(c echo.Context) error {
	req := &aV1.RejectPrivilegeApplyWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUID, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}
	if err := ctl.DMS.RejectPrivilegeApplyWorkflow(c.Request().Context(), req, currentUserUID); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, &aV1.RejectPrivilegeApplyWorkflowReply{})
}

func (ctl *DMSController) RetryReissuePrivilegeApplyWorkflow(c echo.Context) error {
	req := &aV1.RetryReissuePrivilegeApplyWorkflowReq{}
	if err := bindAndValidateReq(c, req); err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUID, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.UnauthorizedErr)
	}
	if err := ctl.DMS.RetryReissuePrivilegeApplyWorkflow(c.Request().Context(), req, currentUserUID); err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, &aV1.RetryReissuePrivilegeApplyWorkflowReply{})
}
