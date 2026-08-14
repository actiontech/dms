package service

import (
	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:operation POST /v1/dms/projects/{project_uid}/ops_types Project CreateOpsType
//
// Create a new ops type.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: ops_type_name
//     description: the name of ops type to be created
//     in: body
//     required: true
//     type: string
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) CreateOpsType(c echo.Context) error {
	req := new(aV1.CreateOpsTypeReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.CreateOpsType(c.Request().Context(), req.ProjectUID, currentUserUid, req.Name)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:operation PUT /v1/dms/projects/{project_uid}/ops_types/{ops_type_uid} Project UpdateOpsType
//
// Update an existing ops type.
//
// ---
// parameters:
//   - name: project_uid
//     description: project uid
//     in: path
//     required: true
//     type: string
//   - name: ops_type_uid
//     description: ops type id
//     in: path
//     required: true
//     type: string
//   - name: ops_type_name
//     description: the name of ops type to be updated
//     required: true
//     in: body
//     type: string
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateOpsType(c echo.Context) error {
	req := new(aV1.UpdateOpsTypeReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.UpdateOpsType(c.Request().Context(), req.ProjectUID, currentUserUid, req.OpsTypeUID, req.Name)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route DELETE /v1/dms/projects/{project_uid}/ops_types/{ops_type_uid} Project DeleteOpsType
//
// Delete an existing ops type.
//
//	responses:
//	  200: body:GenericResp
//	  default: body:GenericResp
func (ctl *DMSController) DeleteOpsType(c echo.Context) error {
	req := new(aV1.DeleteOpsTypeReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.DeleteOpsType(c.Request().Context(), req.ProjectUID, currentUserUid, req.OpsTypeUID)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v1/dms/projects/{project_uid}/ops_types Project ListOpsTypes
//
// List ops types.
//
//	responses:
//	  200: body:ListOpsTypesReply
//	  default: body:GenericResp
func (ctl *DMSController) ListOpsTypes(c echo.Context) error {
	req := new(aV1.ListOpsTypeReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	reply, err := ctl.DMS.ListOpsTypes(c.Request().Context(), req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}
