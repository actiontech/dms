package v1

import (
	"mime/multipart"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
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
	Data *BasicInfo `json:"data"`
	// Generic reply
	base.GenericResp
}

// swagger:response GetStaticLogoReply
type GetStaticLogoReply struct {
	// swagger:file
	// in:  body
	File []byte
}

// swagger:model
type PersonalizationReq struct {
	Title string `json:"title" form:"title"`
	File *multipart.FileHeader `json:"file" form:"file"`
}
