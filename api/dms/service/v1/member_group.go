package v1

import (
	base "github.com/actiontech/dms/api/base/v1"
)

// swagger:enum MemberGroupOrderByField
type MemberGroupOrderByField string

const (
	MemberGroupOrderByName MemberGroupOrderByField = "name"
)

// swagger:parameters ListMemberGroups
type ListMemberGroupsReq struct {
	// the maximum count of member to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of members to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy MemberGroupOrderByField `query:"order_by" json:"order_by"`
	// filter the user group name
	// in:query
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
	// the member namespace uid
	// in:query
	// Required: true
	NamespaceUid string `query:"namespace_uid" json:"namespace_uid" validate:"required"`
}

type ListMemberGroup struct {
	Name string `json:"name"`
	// member uid
	Uid string `json:"uid"`
	// Whether the member has namespace admin permission
	IsNamespaceAdmin bool `json:"is_namespace_admin"`
	// member user
	Users []UidWithName `json:"users"`
	// member op permission
	RoleWithOpRanges []ListMemberRoleWithOpRange `json:"role_with_op_ranges"`
}

// swagger:model ListMemberGroupsReply
type ListMemberGroupsReply struct {
	// List member reply
	Payload struct {
		MemberGroups []*ListMemberGroup `json:"member_groups"`
		Total        int64              `json:"total"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

type GetMemberGroup struct {
	Name string `json:"name"`
	// member group uid
	Uid string `json:"uid"`
	// member user
	Users []UidWithName `json:"users"`
	// Whether the member has namespace admin permission
	IsNamespaceAdmin bool `json:"is_namespace_admin"`
	// member op permission
	RoleWithOpRanges []ListMemberRoleWithOpRange `json:"role_with_op_ranges"`
}

// swagger:parameters GetMemberGroup
type GetMemberGroupReq struct {
	// Member group id
	// Required: true
	// in:path
	MemberGroupUid string `param:"member_group_uid" json:"member_group_uid" validate:"required"`
	// filter by db service namespace uid
	// only the sys user can use an empty namespace value, which means lookup from all namespaces
	// in:query
	NamespaceId string `query:"namespace_id" json:"namespace_id"`
}

// swagger:model GetMemberGroupReply
type GetMemberGroupReply struct {
	// List member reply
	Payload struct {
		MemberGroup *GetMemberGroup `json:"member_group"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

type MemberGroup struct {
	// member group name
	// Required: true
	Name string `json:"name" validate:"required"`
	// member user uid
	// Required: true
	UserUids []string `json:"user_uids" validate:"required"`
	// Whether the member has namespace admin permission
	IsNamespaceAdmin bool `json:"is_namespace_admin"`
	// namespace uid
	// Required: true
	NamespaceUid string `json:"namespace_uid" validate:"required"`
	// member role with op ranges
	RoleWithOpRanges []MemberRoleWithOpRange `json:"role_with_op_ranges"`
}

// swagger:parameters AddMemberGroup
type AddMemberGroupReq struct {
	// Add new member group
	// in:body
	MemberGroup MemberGroup `json:"member_group" validate:"required"`
}

// swagger:model AddMemberGroupReply
type AddMemberGroupReply struct {
	// Add member group reply
	Payload struct {
		// member group ID
		Id string `json:"id"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

type UpdateMemberGroup struct {
	// member user uid
	// Required: true
	UserUids []string `json:"user_uids" validate:"required"`
	// Whether the member has namespace admin permission
	IsNamespaceAdmin bool `json:"is_namespace_admin"`
	// namespace uid
	// Required: true
	NamespaceUid string `json:"namespace_uid" validate:"required"`
	// member role with op ranges
	RoleWithOpRanges []MemberRoleWithOpRange `json:"role_with_op_ranges"`
}

// swagger:parameters UpdateMemberGroup
type UpdateMemberGroupReq struct {
	// Member group id
	// Required: true
	// in:path
	MemberGroupUid string `param:"member_group_uid" json:"member_group_uid" validate:"required"`
	// Update a member group
	// in:body
	MemberGroup *UpdateMemberGroup `json:"member_group" validate:"required"`
}

// swagger:parameters DeleteMemberGroup
type DeleteMemberGroupReq struct {
	// filter by db service namespace uid
	// in:query
	NamespaceId string `query:"namespace_id" json:"namespace_id" validate:"required"`
	// member group id
	// in:path
	MemberGroupUid string `param:"member_group_uid" json:"member_group_uid" validate:"required"`
}
