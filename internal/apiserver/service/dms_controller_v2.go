package service

import "github.com/labstack/echo/v4"

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