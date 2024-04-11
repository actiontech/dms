package v1

import (
	"bytes"
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// A Project
type Project struct {
	// project name
	Name string `json:"name"`
	// project desc
	Desc string `json:"desc"`
	// is fixed business
	IsFixedBusiness bool `json:"is_fixed_business"`
	// project business
	Business []string `json:"business"`
}

// swagger:parameters AddProject
type AddProjectReq struct {
	// Add new Project
	// in:body
	Project *Project `json:"project" validate:"required"`
}

func (u *AddProjectReq) String() string {
	if u == nil {
		return "AddProjectReq{nil}"
	}
	return fmt.Sprintf("AddProjectReq{ProjectName:%s}", u.Project.Name)
}

// swagger:model AddProjectReply
type AddProjectReply struct {
	// Add Project reply
	Data struct {
		// Project UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *AddProjectReply) String() string {
	if u == nil {
		return "AddProjectReply{nil}"
	}
	return fmt.Sprintf("AddProjectReply{Uid:%s}", u.Data.Uid)
}

// swagger:parameters DelProject
type DelProjectReq struct {
	// project uid
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

func (u *DelProjectReq) String() string {
	if u == nil {
		return "DelProjectReq{nil}"
	}
	return fmt.Sprintf("DelProjectReq{Uid:%s}", u.ProjectUid)
}

type UpdateProject struct {
	// Project desc
	Desc *string `json:"desc"`
	// is fixed business
	IsFixedBusiness *bool `json:"is_fixed_business"`
	// Project business
	Business []Business `json:"business"`
}

type Business struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// swagger:parameters UpdateProject
type UpdateProjectReq struct {
	// Project uid
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Update a project
	// in:body
	Project *UpdateProject `json:"project" validate:"required"`
}

func (u *UpdateProjectReq) String() string {
	if u == nil {
		return "UpdateProjectReq{nil}"
	}
	if u.Project == nil {
		return "UpdateProjectReq{Project:nil}"
	}
	return fmt.Sprintf("UpdateProjectReq{Uid:%s}", u.ProjectUid)
}

// swagger:parameters ArchiveProject
type ArchiveProjectReq struct {
	// Project uid
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:parameters UnarchiveProject
type UnarchiveProjectReq struct {
	// Project uid
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:parameters ImportProjects
type ImportProjectsReq struct {
	Projects []*ImportProjects `json:"projects" validate:"required"`
}

type ImportProjects struct {
	// Project name
	Name string `json:"name" validate:"required"`
	// Project desc
	Desc string `json:"desc"`
	// business
	Business []string `json:"business" validate:"required"`
}

// swagger:parameters PreviewImportProjects
type PreviewImportProjectsRep struct {
	// projects file.
	//
	// in: formData
	//
	// swagger:file
	ProjectsFile *bytes.Buffer `json:"projects_file"`
}

// swagger:model PreviewImportProjectsReply
type PreviewImportProjectsReply struct {
	// Generic reply
	base.GenericResp
	// list preview import projects
	Data []*PreviewImportProjects `json:"data"`
}

type PreviewImportProjects struct {
	// Project name
	Name string `json:"name"`
	// Project desc
	Desc string `json:"desc"`
	// business
	Business []string `json:"business"`
}

// swagger:parameters ExportProjects
type ExportProjectsReq struct {
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy dmsCommonV1.ProjectOrderByField `query:"order_by" json:"order_by"`
	// filter the Project name
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
	// filter the Project UID
	FilterByUID string `query:"filter_by_uid" json:"filter_by_uid"`
}

// swagger:response ExportProjectsReply
type ExportProjectsReply struct {
	// swagger:file
	// in:  body
	File []byte
}

// swagger:response GetImportProjectsTemplateReply
type GetImportProjectsTemplateReply struct {
	// swagger:file
	// in: body
	File []byte
}

// swagger:parameters GetProjectTips
type GetProjectTipsReq struct {
	// Project uid
	// in:query
	ProjectUid string `json:"project_uid"`
}

// swagger:model GetProjectTipsReply
type GetProjectTipsReply struct {
	// Generic reply
	base.GenericResp
	// project tips
	Data []*ProjectTips `json:"data"`
}

type ProjectTips struct {
	IsFixedBusiness bool     `json:"is_fixed_business"`
	Business        []string `json:"business"`
}
