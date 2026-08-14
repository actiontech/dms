package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// OpsType 项目级运维类型字典项（对标 EnvironmentTag，无色标）。
// swagger:model OpsType
type OpsType struct {
	UID string `json:"uid,omitempty"`
	// 运维类型名称最多50个字符
	Name string `json:"name" validate:"max=50"`
}

// ListOpsTypeReq 按项目分页列举运维类型（供 dmsobject / 跨服务调用）。
type ListOpsTypeReq struct {
	ProjectUID string `param:"project_uid" json:"project_uid" validate:"required"`
	PageIndex  uint32 `query:"page_index" json:"page_index"`
	PageSize   uint32 `query:"page_size" json:"page_size" validate:"required"`
}

// ListOpsTypesReply 运维类型列表响应。
type ListOpsTypesReply struct {
	Data  []*OpsType `json:"data"`
	Total int64      `json:"total_nums"`
	base.GenericResp
}
