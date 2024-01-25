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
}

type MemberRoleWithOpRange struct {
	// role uid
	RoleUID string `json:"role_uid" validate:"required"`
	// op permission range type, only support db service now
	OpRangeType OpRangeType `json:"op_range_type" validate:"required"`
	// op range uids
	RangeUIDs []string `json:"range_uids" validate:"required"`
}

// swagger:parameters AddMember
type AddMemberReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Add new member
	// in:body
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
}

// A dms member
type ListMember struct {
	// member uid
	MemberUid string `json:"uid"`
	// member user
	User UidWithName `json:"user"`
	// Whether the member has project admin permission
	IsProjectAdmin bool `json:"is_project_admin"`
	// member op permission
	RoleWithOpRanges []ListMemberRoleWithOpRange `json:"role_with_op_ranges"`
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
}

// swagger:parameters UpdateMember
type UpdateMemberReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Member uid
	// Required: true
	// in:path
	MemberUid string `param:"member_uid" json:"member_uid" validate:"required"`
	// Update a member
	// in:body
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
