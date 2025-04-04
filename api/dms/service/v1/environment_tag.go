package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:model
type EnvironmentTag struct {
	UID string `json:"uid,omitempty"`
	// 环境属性标签至少1个字符，最多50个字符
	Name string `json:"name" validate:"min=1,max=50"`
}

// swagger:model
type CreateEnvironmentTagReq struct {
	// swagger:ignore
	ProjectUID     string          `param:"project_uid" json:"project_uid" validate:"required"`
	EnvironmentTag *EnvironmentTag `json:"environment_tag" validate:"required"`
}

// swagger:model
type UpdateEnvironmentTagReq struct {
	// swagger:ignore
	EnvironmentTagUID string `json:"environment_tag_uid" validate:"required"`
	// swagger:ignore
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`

	EnvironmentTag *EnvironmentTag `json:"environment_tag" validate:"required"`
}

// swagger:parameters ListEnvironmentTags
type ListEnvironmentTagReq struct {
	// in:path
	// Required: true
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
}

// swagger:model ListEnvironmentTagsReply
type ListEnvironmentTagsReply struct {
	Data  []*EnvironmentTag `json:"data"`
	Total int64             `json:"total_nums"`
	base.GenericResp
}

// swagger:parameters DeleteEnvironmentTag
type DeleteEnvironmentTagReq struct {
	// in:path
	// Required: true
	EnvironmentTagUID string `json:"environment_tag_uid" validate:"required"`
	// in:path
	// Required: true
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`
}
