package v1

import base "github.com/actiontech/dms/api/base/v1"

// Use this struct to add a new session
type AddSession struct {
	// User name
	// Required: true
	UserName string `json:"username" example:"admin" description:"username" validate:"required"`
	// User password
	// Required: true
	Password string `json:"password" example:"admin" description:"password" validate:"required"`
}

// swagger:parameters AddSession
type AddSessionReq struct {
	// Add a new session
	// in:body
	Session *AddSession `json:"session" validate:"required"`
}

// swagger:model AddSessionReply
type AddSessionReply struct {
	// Add user reply
	Payload struct {
		// User UID
		UserUid string `json:"user_uid"`
		// Session token
		Token string `json:"token"`
	} `json:"payload"`

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
	Payload struct {
		// User UID
		UserUid string `json:"user_uid"`
		// User name
		Name string `json:"name"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}
