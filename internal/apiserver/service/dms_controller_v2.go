package service

import (
	"fmt"
	"mime"
	"net/http"

	dmsApiV2 "github.com/actiontech/dms/api/dms/service/v2"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	commonApiV2 "github.com/actiontech/dms/pkg/dms-common/api/dms/v2"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

// swagger:operation POST /v2/dms/projects Project AddProjectV2
//
// Add project.
//
// ---
// parameters:
//   - name: project
//     description: Add new Project
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddProjectReqV2"
// responses:
//   '200':
//     description: AddProjectReplyV2
//     schema:
//       "$ref": "#/definitions/AddProjectReplyV2"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) AddProjectV2(c echo.Context) error {
	req := new(dmsApiV2.AddProjectReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.AddProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v2/dms/projects/import Project ImportProjectsV2
//
// Import projects.
//
// ---
// parameters:
//   - name: projects
//     description: import projects
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ImportProjectsReqV2"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) ImportProjectsV2(c echo.Context) error {
	req := new(dmsApiV2.ImportProjectsReq)
	err := bindAndValidateReq(c, req)
	if err != nil {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = ctl.DMS.ImportProjects(c.Request().Context(), currentUserUid, req)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}


// swagger:route POST /v2/dms/projects/preview_import Project PreviewImportProjectsV2
//
// Preview import projects.
//
//	Consumes:
//	- multipart/form-data
//
//	responses:
//	  200: PreviewImportProjectsReplyV2
//	  default: body:GenericResp
func (ctl *DMSController) PreviewImportProjectsV2(c echo.Context) error{
	file, exist, err := ReadFileContent(c, ProjectsFileParamKey)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	if !exist {
		return NewErrResp(c, fmt.Errorf("upload file is not exist"), apiError.APIServerErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.PreviewImportProjects(c.Request().Context(), currentUserUid, file)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v2/dms/projects Project ListProjectsV2
//
// List projects.
//
//	responses:
//	  200: body:ListProjectReplyV2
//	  default: body:GenericResp
func (ctl *DMSController) ListProjectsV2(c echo.Context) error {
	req := new(commonApiV2.ListProjectReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := ctl.DMS.ListProjects(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}


// swagger:operation PUT /v2/dms/projects/{project_uid} Project UpdateProjectV2
//
// update a project.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: project
//     description: Update a project
//     required: true
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateProjectReqV2"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (ctl *DMSController) UpdateProjectV2(c echo.Context) error {
	req := &dmsApiV2.UpdateProjectReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = ctl.DMS.UpdateProject(c.Request().Context(), currentUserUid, req)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}


// swagger:operation POST /v2/dms/projects/{project_uid}/db_services DBService AddDBServiceV2
//
// Add DB Service.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_service
//     description: Add new db service
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/AddDBServiceReqV2"
// responses:
//   '200':
//     description: AddDBServiceReply
//     schema:
//       "$ref": "#/definitions/AddDBServiceReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) AddDBServiceV2(c echo.Context) error{
	req := new(dmsApiV2.AddDBServiceReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, err := d.DMS.AddDBServiceV2(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:operation POST /v2/dms/projects/import_db_services Project ImportDBServicesOfProjectsV2
//
// Import DBServices.
//
// ---
// parameters:
//   - name: db_services
//     description: new db services
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ImportDBServicesOfProjectsReqV2"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) ImportDBServicesOfProjectsV2(c echo.Context) error {
	req := new(dmsApiV2.ImportDBServicesOfProjectsReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.ImportDBServicesOfProjects(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}


// swagger:operation POST /v2/dms/projects/{project_uid}/db_services/import DBService ImportDBServicesOfOneProjectV2
//
// Import DBServices.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_services
//     description: new db services
//     in: body
//     required: true
//     schema:
//       "$ref": "#/definitions/ImportDBServicesOfOneProjectReqV2"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) ImportDBServicesOfOneProjectV2(c echo.Context) error {
	req := new(dmsApiV2.ImportDBServicesOfOneProjectReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	err = d.DMS.ImportDBServicesOfOneProject(c.Request().Context(), req, currentUserUid)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	return NewOkResp(c)
}

// swagger:route POST /v2/dms/projects/{project_uid}/db_services/import_check DBService ImportDBServicesOfOneProjectCheckV2
//
// Import DBServices.
//
//	Consumes:
//	- multipart/form-data
//
//	Produces:
//	- application/json
//	- text/csv
//
//	responses:
//	  200: ImportDBServicesCheckCsvReply
//	  default: body:ImportDBServicesCheckReply
func (d *DMSController) ImportDBServicesOfOneProjectCheckV2(c echo.Context) error {
	req := new(dmsApiV2.ImportDBServicesOfOneProjectCheckReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}

	fileContent, exist, err := ReadFileContent(c, DBServicesFileParamKey)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	if !exist {
		return NewErrResp(c, fmt.Errorf("upload file is not exist"), apiError.APIServerErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, csvCheckResult, err := d.DMS.ImportDBServicesOfOneProjectCheck(c.Request().Context(), currentUserUid, req.ProjectUid, fileContent)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	if csvCheckResult != nil {
		c.Response().Header().Set(echo.HeaderContentDisposition,
			mime.FormatMediaType("attachment", map[string]string{"filename": "import_db_services_problems.csv"}))
		return c.Blob(http.StatusOK, "text/csv", csvCheckResult)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:route POST /v2/dms/projects/import_db_services_check Project ImportDBServicesOfProjectsCheckV2
//
// Import DBServices.
//
//		Consumes:
//		- multipart/form-data
//
//		Produces:
//		- application/json
//		- text/csv
//
//	responses:
//	  200: ImportDBServicesCheckCsvReply
//	  default: body:ImportDBServicesCheckReply
func (d *DMSController) ImportDBServicesOfProjectsCheckV2(c echo.Context) error {
	fileContent, exist, err := ReadFileContent(c, DBServicesFileParamKey)
	if err != nil {
		return NewErrResp(c, err, apiError.APIServerErr)
	}
	if !exist {
		return NewErrResp(c, fmt.Errorf("upload file is not exist"), apiError.APIServerErr)
	}

	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}

	reply, csvCheckResult, err := d.DMS.ImportDBServicesOfProjectsCheck(c.Request().Context(), currentUserUid, fileContent)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	if csvCheckResult != nil {
		c.Response().Header().Set(echo.HeaderContentDisposition,
			mime.FormatMediaType("attachment", map[string]string{"filename": "import_db_services_problems.csv"}))
		return c.Blob(http.StatusOK, "text/csv", csvCheckResult)
	}

	return NewOkRespWithReply(c, reply)
}

// swagger:operation PUT /v2/dms/projects/{project_uid}/db_services/{db_service_uid} DBService UpdateDBServiceV2
//
// update a DB Service.
//
// ---
// parameters:
//   - name: project_uid
//     description: project id
//     in: path
//     required: true
//     type: string
//   - name: db_service_uid
//     description: db_service_uid id
//     in: path
//     required: true
//     type: string
//   - name: db_service
//     description: Update a DB service
//     in: body
//     schema:
//       "$ref": "#/definitions/UpdateDBServiceReqV2"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) UpdateDBServiceV2(c echo.Context) error{
		req := &dmsApiV2.UpdateDBServiceReq{}
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	// get current user id
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	err = d.DMS.UpdateDBService(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkResp(c)
}

// swagger:route GET /v2/dms/db_services DBService ListGlobalDBServicesV2
//
// list global DBServices
//
//	responses:
//	  200: body:ListGlobalDBServicesReplyV2
//	  default: body:GenericResp
func (d *DMSController) ListGlobalDBServicesV2(c echo.Context) error {
	req := new(dmsApiV2.ListGlobalDBServicesReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := d.DMS.ListGlobalDBServices(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}

// swagger:route GET /v2/dms/projects/{project_uid}/db_services DBService ListDBServicesV2
//
// List db service.
//
//	responses:
//	  200: body:ListDBServiceReplyV2
//	  default: body:GenericResp
func (d *DMSController) ListDBServicesV2(c echo.Context) error {
	req := new(commonApiV2.ListDBServiceReq)
	err := bindAndValidateReq(c, req)
	if nil != err {
		return NewErrResp(c, err, apiError.BadRequestErr)
	}
	currentUserUid, err := jwt.GetUserUidStrFromContext(c)
	if err != nil {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	reply, err := d.DMS.ListDBServices(c.Request().Context(), req, currentUserUid)
	if nil != err {
		return NewErrResp(c, err, apiError.DMSServiceErr)
	}
	return NewOkRespWithReply(c, reply)
}
