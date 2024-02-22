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

// swagger:response GetLicenseInfoReply
type GetLicenseInfoReply struct {
	// swagger:file
	// in:  body
	File []byte
}

type LicenseUsageItem struct {
	ResourceType     string `json:"resource_type"`
	ResourceTypeDesc string `json:"resource_type_desc"`
	Used             uint   `json:"used"`
	Limit            uint   `json:"limit"`
	IsLimited        bool   `json:"is_limited"`
}

type LicenseUsage struct {
	UsersUsage      LicenseUsageItem   `json:"users_usage"`
	DbServicesUsage []LicenseUsageItem `json:"db_services_usage"`
}

// swagger:model GetLicenseUsageReply
type GetLicenseUsageReply struct {
	// Generic reply
	base.GenericResp
	Data *LicenseUsage `json:"data"`
}
