package v2

import v1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"

// swagger:parameters ListProjectsV2
type ListProjectReq struct {
	// the maximum count of Project to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of Projects to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy v1.ProjectOrderByField `query:"order_by" json:"order_by"`
	// filter the Project name
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
	// filter the Project UID
	FilterByUID string `query:"filter_by_uid" json:"filter_by_uid"`
	// filter project by project id list, using in condition
	// in:query
	FilterByProjectUids []string `query:"filter_by_project_uids" json:"filter_by_project_uids"`
	// filter project by project priority
	// in:query
	FilterByProjectPriority v1.ProjectPriority `query:"filter_by_project_priority" json:"filter_by_project_priority"`
	// filter project by business tag
	// in:query
	FilterByBusinessTag string `query:"filter_by_business_tag" json:"filter_by_business_tag"`
	// filter the Project By Project description
	FilterByDesc string `query:"filter_by_desc" json:"filter_by_desc"`
}