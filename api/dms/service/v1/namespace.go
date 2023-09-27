package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// A Project
type Project struct {
	// project name
	Name string `json:"name"`
	// project desc
	Desc string `json:"desc"`
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
