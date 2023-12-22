package v1

import (
	"bytes"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:model GetLicenseReply
type GetLicenseReply struct {
	// Generic reply
	base.GenericResp
	Content string        `json:"content"`
	License []LicenseItem `json:"license"`
}

type LicenseItem struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Limit       string `json:"limit"`
}

// swagger:model CheckLicenseReply
type CheckLicenseReply struct {
	// Generic CheckLicenseReply
	base.GenericResp
	Content string        `json:"content"`
	License []LicenseItem `json:"license"`
}

// swagger:parameters SetLicense
type SetLicenseReq struct {
	// license file.
	//
	// in: formData
	//
	// swagger:file
	LicenseFile *bytes.Buffer `json:"license_file"`
}

// swagger:parameters CheckLicense
type CheckLicenseReq struct {
	// license file.
	//
	// in: formData
	//
	// swagger:file
	LicenseFile *bytes.Buffer `json:"license_file"`
}
