package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	"time"
)

// swagger:parameters GetCompanyNotice
type GetCompanyNoticeReq struct {
	// When true, return the latest notice record regardless of the display time window (e.g. expired or not yet started); intended for admin edit forms.
	// in: query
	IncludeLatestOutsidePeriod bool `query:"include_latest_outside_period" json:"include_latest_outside_period"`
}

// A companynotice
type CompanyNotice struct {
	// companynotice info
	NoticeStr string `json:"notice_str"`
	// companynotice creator name
	CreateUserName string `json:"create_user_name"`
	// current user has been read
	ReadByCurrentUser bool `json:"read_by_current_user"`
	// notice show start time
	StartTime *time.Time `json:"start_time,omitempty"`
	// notice expire time
	ExpireTime *time.Time `json:"expire_time,omitempty"`
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
	NoticeStr *string `json:"notice_str" validate:"required"`
	// notice show start time
	StartTime *time.Time `json:"start_time" validate:"required"`
	// notice show end time
	EndTime *time.Time `json:"end_time" validate:"required"`
}
