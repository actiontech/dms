package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// A user
type User struct {
	UID string `json:"uid"`
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
	// 对接登录的参数
	ThirdPartyUserID       string `json:"third_party_user_id"`
	ThirdPartyUserInfo     string `json:"third_party_user_info"`
	UserAuthenticationType string `json:"user_authentication_type"`
}

// swagger:model
type AddUserReq struct {
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
	Data struct {
		// user UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *AddUserReply) String() string {
	if u == nil {
		return "AddUserReply{nil}"
	}
	return fmt.Sprintf("AddUserReply{Uid:%s}", u.Data.Uid)
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
	// User language
	Language *string `json:"language"`
	// User group uids
	UserGroupUids *[]string `json:"user_group_uids"`
	// User operation permission uids
	OpPermissionUids *[]string `json:"op_permission_uids" validate:"required"`
	// 对接登录的参数
	ThirdPartyUserID       *string `json:"third_party_user_id"`
	ThirdPartyUserInfo     *string `json:"third_party_user_info"`
	UserAuthenticationType *string `json:"user_authentication_type"`
	// User system
	System *UserSystem `json:"system"`
}

// swagger:model
type UpdateUserReq struct {
	// swagger:ignore
	UserUid string      `param:"user_uid" json:"user_uid" validate:"required"`
	User    *UpdateUser `json:"user" validate:"required"`
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

// swagger:model
type UpdateCurrentUserReq struct {
	User *UpdateCurrentUser `json:"current_user" validate:"required"`
}

func (u *UpdateCurrentUserReq) String() string {
	if u == nil {
		return "UpdateCurrentUserReq{nil}"
	}
	if u.User == nil {
		return "UpdateCurrentUserReq{User:nil}"
	}
	return fmt.Sprintf("UpdateCurrentUserReq{Uid:%v}", u.User)
}

type UpdateCurrentUser struct {
	// User old password
	OldPassword *string `json:"old_password"`
	// User new password
	Password *string `json:"password"`
	// User email
	Email *string `json:"email"`
	// User phone
	Phone *string `json:"phone"`
	// User wxid
	WxID *string `json:"wxid"`
	// User language
	Language *string `json:"language"`
	// User two factor enabled
	TwoFactorEnabled *bool `json:"two_factor_enabled"`
	// User system
	System *UserSystem `json:"system"`
}

type VerifyUserLoginReq struct {
	// user name
	UserName string `json:"user_name" validate:"required"`
	// user password
	Password string `json:"password" validate:"required"`
	// VerifyCode
	VerifyCode *string `json:"verify_code" description:"verify_code"`
}

func (u *VerifyUserLoginReq) String() string {
	if u == nil {
		return "VerifyUserLoginReq{nil}"
	}
	return fmt.Sprintf("VerifyUserLoginReq{UserName:%s}", u.UserName)
}

// swagger:model
type VerifyUserLoginReply struct {
	base.GenericResp
	Data struct {
		// If verify Successful, return empty string, otherwise return error message
		VerifyFailedMsg string `json:"verify_failed_msg"`
		// If verify Successful, return user uid
		UserUid          string `json:"user_uid"`
		Phone            string `json:"phone"`
		TwoFactorEnabled bool   `json:"two_factor_enabled"`
	} `json:"data"`
}

func (u *VerifyUserLoginReply) String() string {
	if u == nil {
		return "VerifyUserLoginReply{nil}"
	}
	return fmt.Sprintf("VerifyUserLoginReply{UserUid:%s}", u.Data.UserUid)
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

// swagger:enum UserSystem
type UserSystem string

const (
	UserSystemWorkbench  UserSystem = "WORKBENCH"
	UserSystemManagement UserSystem = "MANAGEMENT"
)
