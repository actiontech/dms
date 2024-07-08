package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// A companynotice
type CompanyNotice struct {
	// companynotice info
	NoticeStr string `json:"notice_str"`
	// current user has been read
	ReadByCurrentUser bool `json:"read_by_current_user"`
}

// swagger:model GetCompanyNoticeReply
type GetCompanyNoticeReply struct {
	Data CompanyNotice `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:model
type UpdateCompanyNoticeReq struct {
	UpdateCompanyNotice UpdateCompanyNotice `json:"company_notice"  validate:"required"`
}

// A companynotice
type UpdateCompanyNotice struct {
	// companynotice info
	NoticeStr *string `json:"notice_str"  valid:"omitempty"`
}
