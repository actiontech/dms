package v1

import (
	"fmt"

	base "github.com/actiontech/dms/api/base/v1"
)

// A user
type User struct {
	// user name
	// Required: true
	Name string `json:"name" validate:"required"`
	// user description
	Desc string `json:"desc"`
	// user email
	Email string `json:"email"`
	// user phone
	Phone string `json:"phone"`
	// user wxid
	WxID string `json:"wxid"`
	// user password
	Password string `json:"password"`
	// user group uid
	UserGroupUids []string `json:"user_group_uids"`
	// user op permission uid
	OpPermissionUids []string `json:"op_permission_uids"`
}

// swagger:parameters AddUser
type AddUserReq struct {
	// Add new user
	// in:body
	User *User `json:"user" validate:"required"`
}

func (u *AddUserReq) String() string {
	if u == nil {
		return "AddUserReq{nil}"
	}
	return fmt.Sprintf("AddUserReq{Name:%s}", u.User.Name)
}

// swagger:model AddUserReply
type AddUserReply struct {
	// Add user reply
	Payload struct {
		// user UID
		Uid string `json:"uid"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

func (u *AddUserReply) String() string {
	if u == nil {
		return "AddUserReply{nil}"
	}
	return fmt.Sprintf("AddUserReply{Uid:%s}", u.Payload.Uid)
}

// swagger:parameters DelUser
type DelUserReq struct {
	// user uid
	// in:path
	UserUid string `param:"user_uid" json:"user_uid" validate:"required"`
}

func (u *DelUserReq) String() string {
	if u == nil {
		return "DelUserReq{nil}"
	}
	return fmt.Sprintf("DelUserReq{Uid:%s}", u.UserUid)
}

type UpdateUser struct {
	// Whether the user is disabled or not
	IsDisabled *bool `json:"is_disabled" validate:"required"`
	// User password, User can login with this password
	Password *string `json:"password"`
	// User email
	Email *string `json:"email"`
	// User phone
	Phone *string `json:"phone"`
	// User wxid
	WxID *string `json:"wxid"`
	// User group uids
	UserGroupUids *[]string `json:"user_group_uids" validate:"required"`
	// User operation permission uids
	OpPermissionUids *[]string `json:"op_permission_uids" validate:"required"`
}

// swagger:parameters UpdateUser
type UpdateUserReq struct {
	// User uid
	// Required: true
	// in:path
	UserUid string `param:"user_uid" json:"user_uid" validate:"required"`
	// Update a user
	// in:body
	User *UpdateUser `json:"user" validate:"required"`
}

func (u *UpdateUserReq) String() string {
	if u == nil {
		return "UpdateUserReq{nil}"
	}
	if u.User == nil {
		return "UpdateUserReq{User:nil}"
	}
	return fmt.Sprintf("UpdateUserReq{Uid:%s}", u.UserUid)
}

type VerifyUserLoginReq struct {
	// user name
	UserName string `json:"user_name" validate:"required"`
	// user password
	Password string `json:"password" validate:"required"`
}

func (u *VerifyUserLoginReq) String() string {
	if u == nil {
		return "VerifyUserLoginReq{nil}"
	}
	return fmt.Sprintf("VerifyUserLoginReq{UserName:%s}", u.UserName)
}

type VerifyUserLoginReply struct {
	Payload struct {
		// If verify Successful, return empty string, otherwise return error message
		VerifyFailedMsg string `json:"verify_failed_msg"`
		// If verify Successful, return user uid
		UserUid string `json:"user_uid"`
	} `json:"payload"`
}

func (u *VerifyUserLoginReply) String() string {
	if u == nil {
		return "VerifyUserLoginReply{nil}"
	}
	return fmt.Sprintf("VerifyUserLoginReply{UserUid:%s}", u.Payload.UserUid)
}

type AfterUserLoginReq struct {
	// user uid
	UserUid string `json:"user_uid" validate:"required"`
}

func (u *AfterUserLoginReq) String() string {
	if u == nil {
		return "AfterUserLoginReq{nil}"
	}
	return fmt.Sprintf("AfterUserLoginReq{UserUid:%s}", u.UserUid)
}
