package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// A user group
type UserGroup struct {
	// user group name
	// Required: true
	Name string `json:"name" validate:"required"`
	// user group description
	Desc string `json:"desc"`
	// user uids
	UserUids []string `json:"user_uids"`
}

// swagger:model
type AddUserGroupReq struct {
	UserGroup *UserGroup `json:"user_group" validate:"required"`
}

func (u *AddUserGroupReq) String() string {
	if u == nil {
		return "AddUserGroupReq{nil}"
	}
	return fmt.Sprintf("AddUserGroupReq{Name:%s}", u.UserGroup.Name)
}

// swagger:model AddUserGroupReply
type AddUserGroupReply struct {
	// Add user group reply
	Data struct {
		// user group UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *AddUserGroupReply) String() string {
	if u == nil {
		return "AddUserGroupReply{nil}"
	}
	return fmt.Sprintf("AddUserGroupReply{Uid:%s}", u.Data.Uid)
}

// swagger:parameters DelUserGroup
type DelUserGroupReq struct {
	// user group uid
	// in:path
	UserGroupUid string `param:"user_group_uid" json:"user_group_uid" validate:"required"`
}

func (u *DelUserGroupReq) String() string {
	if u == nil {
		return "DelUserGroupReq{nil}"
	}
	return fmt.Sprintf("DelUserGroupReq{Uid:%s}", u.UserGroupUid)
}

// swagger:parameters ListUserGroups
type ListUserGroupReq struct {
	// the maximum count of user to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of user groups to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy UserGroupOrderByField `query:"order_by" json:"order_by"`
	// filter the user group name
	// in:query
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
}

// swagger:enum UserGroupOrderByField
type UserGroupOrderByField string

const (
	UserGroupOrderByName UserGroupOrderByField = "name"
)

// A dms user group
type ListUserGroup struct {
	// user group uid
	UserGroupUid string `json:"uid"`
	// user group name
	Name string `json:"name"`
	// user group description
	Desc string `json:"desc"`
	// user group stat
	Stat Stat `json:"stat"`
	// users
	Users []UidWithName `json:"users"`
}

// swagger:model ListUserGroupReply
type ListUserGroupReply struct {
	// List user reply
	Data  []*ListUserGroup `json:"data"`
	Total int64            `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

type UpdateUserGroup struct {
	// Whether the user group is disabled or not
	IsDisabled *bool `json:"is_disabled" validate:"required"`
	// UserGroup description
	Desc *string `json:"desc"`
	// User uids
	UserUids *[]string `json:"user_uids" validate:"required"`
}

// swagger:model
type UpdateUserGroupReq struct {
    // swagger:ignore
	UserGroupUid string `param:"user_group_uid" json:"user_group_uid" validate:"required"`
	UserGroup *UpdateUserGroup `json:"user_group" validate:"required"`
}

func (u *UpdateUserGroupReq) String() string {
	if u == nil {
		return "UpdateUserGroupReq{nil}"
	}
	if u.UserGroup == nil {
		return "UpdateUserGroupReq{UserGroup:nil}"
	}
	return fmt.Sprintf("UpdateUserGroupReq{Uid:%s}", u.UserGroupUid)
}
