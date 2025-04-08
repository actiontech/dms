package v2

import (
	"bytes"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// A Project
type Project struct {
	// project name
	Name string `json:"name"`
	// project desc
	Desc string `json:"desc"`
	// project business tag
	BusinessTag *v1.BusinessTag `json:"business_tag"`
	// project priority
	ProjectPriority dmsCommonV1.ProjectPriority `json:"project_priority"  enums:"high,medium,low"`
}

// swagger:model AddProjectReqV2
type AddProjectReq struct {
	Project *Project `json:"project" validate:"required"`
}

// swagger:model AddProjectReplyV2
type AddProjectReply struct {
	// Add Project reply
	Data struct {
		// Project UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:model UpdateProjectReqV2
type UpdateProjectReq struct {
	// swagger:ignore
	ProjectUid string         `param:"project_uid" json:"project_uid" validate:"required"`
	Project    *UpdateProject `json:"project" validate:"required"`
}

type UpdateProject struct {
	// Project desc
	Desc *string `json:"desc"`
	// project business tag
	BusinessTag *v1.BusinessTag `json:"business_tag"`
	// project priority
	ProjectPriority *dmsCommonV1.ProjectPriority `json:"project_priority"  enums:"high,medium,low"`
}

// swagger:model ImportProjectsReqV2
type ImportProjectsReq struct {
	Projects []*ImportProjects `json:"projects" validate:"required"`
}

type ImportProjects struct {
	// Project name
	Name string `json:"name" validate:"required"`
	// Project desc
	Desc string `json:"desc"`
	// project business tag
	BusinessTag *v1.BusinessTag `json:"business_tag"`
}

// swagger:parameters PreviewImportProjectsV2
type PreviewImportProjectsRep struct {
	// projects file.
	//
	// in: formData
	//
	// swagger:file
	ProjectsFile *bytes.Buffer `json:"projects_file"`
}

// swagger:model PreviewImportProjectsReplyV2
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
	// project business tag
	BusinessTag *v1.BusinessTag `json:"business_tag"`
}
