package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// A role
type Role struct {
	// role name
	// Required: true
	Name string `json:"name" validate:"required"`
	// role description
	Desc string `json:"desc"`
	// op permission uid
	OpPermissionUids []string `json:"op_permission_uids"`
}

// swagger:model
type AddRoleReq struct {
	Role *Role `json:"role" validate:"required"`
}

func (u *AddRoleReq) String() string {
	if u == nil {
		return "AddRoleReq{nil}"
	}
	return fmt.Sprintf("AddRoleReq{Name:%s}", u.Role.Name)
}

// swagger:model AddRoleReply
type AddRoleReply struct {
	// Add role reply
	Data struct {
		// role UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *AddRoleReply) String() string {
	if u == nil {
		return "AddRoleReply{nil}"
	}
	return fmt.Sprintf("AddRoleReply{Uid:%s}", u.Data.Uid)
}

// swagger:parameters DelRole
type DelRoleReq struct {
	// role uid
	// in:path
	RoleUid string `param:"role_uid" json:"role_uid" validate:"required"`
}

func (u *DelRoleReq) String() string {
	if u == nil {
		return "DelRoleReq{nil}"
	}
	return fmt.Sprintf("DelRoleReq{Uid:%s}", u.RoleUid)
}

// swagger:parameters ListRoles
type ListRoleReq struct {
	// the maximum count of role to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of roles to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy RoleOrderByField `query:"order_by" json:"order_by"`
	// filter the role name
	// in:query
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
}

// swagger:enum RoleOrderByField
type RoleOrderByField string

const (
	RoleOrderByName RoleOrderByField = "name"
)

// A dms role
type ListRole struct {
	// role uid
	RoleUid string `json:"uid"`
	// role name
	Name string `json:"name"`
	// role stat
	Stat dmsCommonV1.Stat `json:"stat"`
	// role desc
	Desc string `json:"desc"`
	// op permissions
	OpPermissions []UidWithName `json:"op_permissions"`
}

// swagger:model ListRoleReply
type ListRoleReply struct {
	// List role reply
	Data  []*ListRole `json:"data"`
	Total int64       `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

type UpdateRole struct {
	// Whether the role is disabled or not
	IsDisabled *bool `json:"is_disabled" validate:"required"`
	// Role desc
	Desc *string `json:"desc"`
	// Op permission uids
	OpPermissionUids *[]string `json:"op_permission_uids" validate:"required"`
}

// swagger:model
type UpdateRoleReq struct {
	// swagger:ignore
	RoleUid string      `param:"role_uid" json:"role_uid" validate:"required"`
	Role    *UpdateRole `json:"role" validate:"required"`
}

func (u *UpdateRoleReq) String() string {
	if u == nil {
		return "UpdateRoleReq{nil}"
	}
	if u.Role == nil {
		return "UpdateRoleReq{Role:nil}"
	}
	return fmt.Sprintf("UpdateRoleReq{Uid:%s}", u.RoleUid)
}
