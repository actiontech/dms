package v1

import (
	"mime/multipart"

	base "github.com/actiontech/dms/api/base/v1"
)

type ComponentNameWithVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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

// swagger:response GetStaticLogoReply
type GetStaticLogoReply struct {
	// swagger:file
	// in:  body
	File []byte
}

// swagger:parameters Personalization
type PersonalizationReq struct {
	// title
	// Required: false
	// in: formData
	Title string `json:"title" form:"title"`

	// file upload
	// Required: false
	// in: formData
	// swagger:file
	File *multipart.FileHeader `json:"file" form:"file"`
}
