package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// A member
type Member struct {
	// member user uid
	// Required: true
	UserUid string `json:"user_uid" validate:"required"`
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member role with op ranges
	RoleWithOpRanges []MemberRoleWithOpRange `json:"role_with_op_ranges"`
	// member project manage permissions
	ProjectManagePermissions []string `json:"project_manage_permissions"`
}

type MemberRoleWithOpRange struct {
	// role uid
	RoleUID string `json:"role_uid" validate:"required"`
	// op permission range type, only support db service now
	OpRangeType OpRangeType `json:"op_range_type" validate:"required"`
	// op range uids
	RangeUIDs []string `json:"range_uids" validate:"required"`
}

// swagger:model
type AddMemberReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	Member *Member `json:"member" validate:"required"`
}

func (u *AddMemberReq) String() string {
	if u == nil {
		return "AddMemberReq{nil}"
	}
	return fmt.Sprintf("AddMemberReq{UserUid:%s}", u.Member.UserUid)
}

// swagger:model AddMemberReply
type AddMemberReply struct {
	// Add member reply
	Data struct {
		// member UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *AddMemberReply) String() string {
	if u == nil {
		return "AddMemberReply{nil}"
	}
	return fmt.Sprintf("AddMemberReply{Uid:%s}", u.Data.Uid)
}

// swagger:parameters ListMembers
type ListMemberReq struct {
	// the maximum count of member to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of members to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy MemberOrderByField `query:"order_by" json:"order_by"`
	// filter the member user uid
	// in:query
	FilterByUserUid string `query:"filter_by_user_uid" json:"filter_by_user_uid"`
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:enum MemberOrderByField
type MemberOrderByField string

const (
	MemberOrderByUserUid MemberOrderByField = "user_uid"
)

type ListMemberRoleWithOpRange struct {
	// role uid
	RoleUID UidWithName `json:"role_uid" validate:"required"`
	// op permission range type, only support db service now
	OpRangeType OpRangeType `json:"op_range_type" validate:"required"`
	// op range uids
	RangeUIDs []UidWithName `json:"range_uids" validate:"required"`
	// member op permissions
	OpPermissions []UidWithName  `json:"op_permissions"`
	// member group
	MemberGroup *ProjectMemberGroup `json:"member_group"`
}

// A dms member
type ListMember struct {
	// member uid
	MemberUid string `json:"uid"`
	// member user
	User UidWithName `json:"user"`
	// Whether the member is a group member
	IsGroupMember  bool `json:"is_group_member"`
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// current project admin info
	CurrentProjectAdmin CurrentProjectAdmin `json:"current_project_admin"`
	// member op permission
	RoleWithOpRanges []ListMemberRoleWithOpRange `json:"role_with_op_ranges"`
	// current project permission
	CurrentProjectOpPermissions []ProjectOpPermission `json:"current_project_op_permissions"`
	// current project manage permissions
	CurrentProjectManagePermissions []ProjectManagePermission `json:"current_project_manage_permissions"`
	// member platform roles
	PlatformRoles []UidWithName `json:"platform_roles"`
	// member projects
	Projects []string `json:"projects"`
}

type CurrentProjectAdmin struct {
	IsAdmin bool `json:"is_admin"`
	MemberGroups []string `json:"member_groups"`
}

type ProjectManagePermission struct {
	Uid string `json:"uid"`
	Name string `json:"name"`
	MemberGroup string `json:"member_group"`
}

type ProjectOpPermission struct {
	DataSource string `json:"data_source"`
	Roles []ProjectRole `json:"roles"`
}

type ProjectRole struct {
	Uid string `json:"uid"`
	Name string `json:"name"`
	OpPermissions []UidWithName `json:"op_permissions"`
	MemberGroup *ProjectMemberGroup `json:"member_group"`
}

type ProjectMemberGroup  struct {
	Uid string `json:"uid"`
	Name string `json:"name"`
	Users []UidWithName `json:"users"`
	OpPermissions []UidWithName `json:"op_permissions"`
}

// swagger:model ListMemberReply
type ListMemberReply struct {
	// List member reply
	Data  []*ListMember `json:"data"`
	Total int64         `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ListMemberTips
type ListMemberTipsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:parameters ListMemberGroupTips
type ListMemberGroupTipsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model ListMemberGroupTipsReply
type ListMemberGroupTipsReply struct {
	// List member tip reply
	Data []UidWithName `json:"data"`

	// Generic reply
	base.GenericResp
}


type ListMemberTipsItem struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// swagger:model ListMemberTipsReply
type ListMemberTipsReply struct {
	// List member tip reply
	Data []ListMemberTipsItem `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters DelMember
type DelMemberReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// member uid
	// in:path
	MemberUid string `param:"member_uid" json:"member_uid" validate:"required"`
}

func (u *DelMemberReq) String() string {
	if u == nil {
		return "DelMemberReq{nil}"
	}
	return fmt.Sprintf("DelMemberReq{Uid:%s}", u.MemberUid)
}

type UpdateMember struct {
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member role with op ranges
	RoleWithOpRanges []MemberRoleWithOpRange `json:"role_with_op_ranges"`
	// member project manage permissions
	ProjectManagePermissions []string `json:"project_manage_permissions"`
}

// swagger:model
type UpdateMemberReq struct {
    // swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	MemberUid string `param:"member_uid" json:"member_uid" validate:"required"`
	Member *UpdateMember `json:"member" validate:"required"`
}

func (u *UpdateMemberReq) String() string {
	if u == nil {
		return "UpdateMemberReq{nil}"
	}
	if u.Member == nil {
		return "UpdateMemberReq{Member:nil}"
	}
	return fmt.Sprintf("UpdateMemberReq{Uid:%s}", u.MemberUid)
}
