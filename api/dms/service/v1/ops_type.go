package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// swagger:model
type CreateOpsTypeReq struct {
	// swagger:ignore
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`
	Name       string `json:"ops_type_name" validate:"required,min=1,max=50"`
}

// swagger:model
type UpdateOpsTypeReq struct {
	// swagger:ignore
	OpsTypeUID string `param:"ops_type_uid" json:"ops_type_uid" validate:"required"`
	// swagger:ignore
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`

	Name string `json:"ops_type_name" validate:"required,min=1,max=50"`
}

// swagger:parameters ListOpsTypes
type ListOpsTypeReq struct {
	// in:path
	// Required: true
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
}

// swagger:model ListOpsTypesReply
type ListOpsTypesReply struct {
	Data  []*dmsCommonV1.OpsType `json:"data"`
	Total int64                  `json:"total_nums"`
	base.GenericResp
}

// swagger:parameters DeleteOpsType
type DeleteOpsTypeReq struct {
	// in:path
	// Required: true
	OpsTypeUID string `param:"ops_type_uid" json:"ops_type_uid" validate:"required"`
	// in:path
	// Required: true
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`
}
