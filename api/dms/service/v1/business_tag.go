package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:model
type BusinessTag struct {
	UID string `json:"uid,omitempty"`
	// 业务标签至少1个字符，最多50个字符
	Name string `json:"name" validate:"min=1,max=50"`
}

// swagger:model
type CreateBusinessTagReq struct {
	BusinessTag *BusinessTag `json:"business_tag" validate:"required"`
}

// swagger:model
type UpdateBusinessTagReq struct {
	// swagger:ignore
	BusinessTagUID string       `json:"business_tag_uid" validate:"required"`
	BusinessTag    *BusinessTag `json:"business_tag" validate:"required"`
}

// swagger:parameters ListBusinessTags
type ListBusinessTagReq struct {
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
}

// swagger:model ListBusinessTagsReply
type ListBusinessTagsReply struct {
	Data  []*BusinessTag `json:"data"`
	Total int64          `json:"total_nums"`
	base.GenericResp
}

// swagger:parameters DeleteBusinessTag
type DeleteBusinessTagReq struct {
	// in:path
	// Required: true
	BusinessTagUID string `param:"business_tag_uid" json:"business_tag_uid" validate:"required"`
}