package service

import (
	aV2 "github.com/actiontech/dms/api/dms/service/v2"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
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
//       "$ref": "#/definitions/AddProjectReq"
// responses:
//   '200':
//     description: AddProjectReply
//     schema:
//       "$ref": "#/definitions/AddProjectReply"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) AddProjectV2(c echo.Context) error {
	return d.AddProject(c)
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
//       "$ref": "#/definitions/ImportProjectsReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) ImportProjectsV2(c echo.Context) error {
	return d.ImportProjects(c)
}


// swagger:route POST /v2/dms/projects/preview_import Project PreviewImportProjectsV2
//
// Preview import projects.
//
//	Consumes:
//	- multipart/form-data
//
//	responses:
//	  200: PreviewImportProjectsReply
//	  default: body:GenericResp
func (d *DMSController) PreviewImportProjectsV2(c echo.Context) error{
	return d.PreviewImportProjects(c)
}

// swagger:route GET /v2/dms/projects Project ListProjectsV2
//
// List projects.
//
//	responses:
//	  200: body:ListProjectReply
//	  default: body:GenericResp
func (d *DMSController) ListProjectsV2(c echo.Context) error {
	return d.ListProjects(c)
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
	req := new(aV2.AddDBServiceReq)
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
//       "$ref": "#/definitions/ImportDBServicesOfOneProjectReq"
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
	return d.ImportDBServicesOfOneProject(c)
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
	return d.ImportDBServicesOfOneProjectCheck(c)
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
	return d.ImportDBServicesOfProjectsCheck(c)
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
//       "$ref": "#/definitions/UpdateDBServiceReq"
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
	return d.UpdateDBService(c) 
}

// swagger:route GET /v2/dms/db_services DBService ListGlobalDBServicesV2
//
// list global DBServices
//
//	responses:
//	  200: body:ListGlobalDBServicesReply
//	  default: body:GenericResp
func (d *DMSController) ListGlobalDBServicesV2(c echo.Context) error {
	return d.ListGlobalDBServices(c)
}

// swagger:route GET /v2/dms/projects/{project_uid}/db_services DBService ListDBServicesV2
//
// List db service.
//
//	responses:
//	  200: body:ListDBServiceReply
//	  default: body:GenericResp
func (d *DMSController) ListDBServicesV2(c echo.Context) error {
	return d.ListDBServices(c)
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
//       "$ref": "#/definitions/UpdateProjectReq"
// responses:
//   '200':
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
//   default:
//     description: GenericResp
//     schema:
//       "$ref": "#/definitions/GenericResp"
func (d *DMSController) UpdateProjectV2(c echo.Context) error {
	return d.UpdateProject(c)
}
