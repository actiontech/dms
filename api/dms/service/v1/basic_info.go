package v1

import (
	base "github.com/actiontech/dms/api/base/v1"
)

type ComponentNameWithVersion struct {
	Name    string
	Version string
}
type BasicInfo struct {
	LogoUrl    string                     `json:"logo_url"`
	Title      string                     `json:"title"`
	Components []ComponentNameWithVersion `json:"components"`
}

// swagger:model GetBasicInfoReply
type GetBasicInfoReply struct {
	Payload struct {
		BasicInfo *BasicInfo `json:"basic_info"`
	} `json:"payload"`
	// Generic reply
	base.GenericResp
}
