package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// Use this struct to add a new session
type AddSession struct {
	// User name
	// Required: true
	UserName string `json:"username" example:"admin" description:"username" validate:"required"`
	// User password
	// Required: true
	Password string `json:"password" example:"admin" description:"password" validate:"required"`
	// VerifyCode
	VerifyCode *string `json:"verify_code" example:"1111" description:"verify_code"`
}

// swagger:model
type AddSessionReq struct {
	Session *AddSession `json:"session" validate:"required"`
}

// swagger:model AddSessionReply
type AddSessionReply struct {
	// Add user reply
	Data struct {
		// Session token
		Token string `json:"token"`
		// Message
		Message string `json:"message"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:model DelSessionReply
type DelSessionReply struct {
	// Del session reply
	Data struct {
		// Session token
		Location string `json:"location"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters GetUserBySession
type GetUserBySessionReq struct {
	UserUid string `json:"user_uid" validate:"required"`
}

// swagger:model GetUserBySessionReply
type GetUserBySessionReply struct {
	// Get user reply
	Data struct {
		// User UID
		UserUid string `json:"user_uid"`
		// User name
		Name string `json:"name"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}
