package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
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
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

type ListMemberGroup struct {
	Name string `json:"name"`
	// member uid
	Uid string `json:"uid"`
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member user
	Users []UidWithName `json:"users"`
	// member op permission
	RoleWithOpRanges []ListMemberRoleWithOpRange `json:"role_with_op_ranges"`
	// current project permission
	CurrentProjectOpPermissions []ProjectOpPermission `json:"current_project_op_permissions"`
	// member project manage permissions
	CurrentProjectManagePermissions []UidWithName `json:"current_project_manage_permissions"`
}

// swagger:model ListMemberGroupsReply
type ListMemberGroupsReply struct {
	// List member reply
	Data  []*ListMemberGroup `json:"data"`
	Total int64              `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

type GetMemberGroup struct {
	Name string `json:"name"`
	// member group uid
	Uid string `json:"uid"`
	// member user
	Users []UidWithName `json:"users"`
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member op permission
	RoleWithOpRanges []ListMemberRoleWithOpRange `json:"role_with_op_ranges"`
}

// swagger:parameters GetMemberGroup
type GetMemberGroupReq struct {
	// Member group id
	// Required: true
	// in:path
	MemberGroupUid string `param:"member_group_uid" json:"member_group_uid" validate:"required"`
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetMemberGroupReply
type GetMemberGroupReply struct {
	// List member reply
	Data *GetMemberGroup `json:"data"`

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
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member role with op ranges
	RoleWithOpRanges []MemberRoleWithOpRange `json:"role_with_op_ranges"`
	// member project manage permissions
	ProjectManagePermissions []string `json:"project_manage_permissions"`
}

// swagger:model
type AddMemberGroupReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	MemberGroup MemberGroup `json:"member_group" validate:"required"`
}

// swagger:model AddMemberGroupReply
type AddMemberGroupReply struct {
	// Add member group reply
	Data struct {
		// member group ID
		Id string `json:"id"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

type UpdateMemberGroup struct {
	// member user uid
	// Required: true
	UserUids []string `json:"user_uids" validate:"required"`
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member role with op ranges
	RoleWithOpRanges []MemberRoleWithOpRange `json:"role_with_op_ranges"`
	// member project manage permissions
	ProjectManagePermissions []string `json:"project_manage_permissions"`
}

// swagger:model
type UpdateMemberGroupReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	MemberGroupUid string `param:"member_group_uid" json:"member_group_uid" validate:"required"`
	MemberGroup *UpdateMemberGroup `json:"member_group" validate:"required"`
}

// swagger:parameters DeleteMemberGroup
type DeleteMemberGroupReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// member group id
	// in:path
	MemberGroupUid string `param:"member_group_uid" json:"member_group_uid" validate:"required"`
}
