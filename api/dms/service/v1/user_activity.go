package v1

import (
	"time"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters GetUserActivitySummary
type GetUserActivitySummaryReq struct {
	// in:query
	// Required: true
	StatDate string `json:"stat_date" query:"stat_date" validate:"required"`
}

// swagger:model GetUserActivitySummaryReply
type GetUserActivitySummaryReply struct {
	Data *UserActivitySummary `json:"data"`
	base.GenericResp
}

type UserActivitySummary struct {
	DAU               int     `json:"dau"`
	RequestCount      int     `json:"request_count"`
	AvgRequestPerUser float64 `json:"avg_request_per_user"`
	ErrorCount        int     `json:"error_count"`
	ErrorRate         float64 `json:"error_rate"`
	PeakHour          int     `json:"peak_hour"`
	PeakHourRequests  int     `json:"peak_hour_requests"`
}

// swagger:parameters ListUserActivityDailyTrend
type ListUserActivityDailyTrendReq struct {
	// in:query
	// Required: true
	FilterDateFrom string `json:"filter_date_from" query:"filter_date_from" validate:"required"`
	// in:query
	// Required: true
	FilterDateTo string `json:"filter_date_to" query:"filter_date_to" validate:"required"`
}

// swagger:model ListUserActivityDailyTrendReply
type ListUserActivityDailyTrendReply struct {
	Data []UserActivityDailyTrendItem `json:"data"`
	base.GenericResp
}

type UserActivityDailyTrendItem struct {
	StatDate     string `json:"stat_date"`
	DAU          int    `json:"dau"`
	RequestCount int    `json:"request_count"`
	ErrorCount   int    `json:"error_count"`
}

// swagger:parameters ListUserActivityModuleDistribution
type ListUserActivityModuleDistributionReq struct {
	// in:query
	// Required: true
	StatDate string `json:"stat_date" query:"stat_date" validate:"required"`
}

// swagger:model ListUserActivityModuleDistributionReply
type ListUserActivityModuleDistributionReply struct {
	Data []UserActivityModuleDistributionItem `json:"data"`
	base.GenericResp
}

type UserActivityModuleDistributionItem struct {
	ModuleCode   string  `json:"module_code"`
	ModuleName   string  `json:"module_name"`
	RequestCount int     `json:"request_count"`
	Percent      float64 `json:"percent"`
}

// swagger:parameters ListUserActivityHourlyDistribution
type ListUserActivityHourlyDistributionReq struct {
	// in:query
	// Required: true
	StatDate string `json:"stat_date" query:"stat_date" validate:"required"`
}

// swagger:model ListUserActivityHourlyDistributionReply
type ListUserActivityHourlyDistributionReply struct {
	Data []UserActivityHourlyDistributionItem `json:"data"`
	base.GenericResp
}

type UserActivityHourlyDistributionItem struct {
	StatHour     int `json:"stat_hour"`
	RequestCount int `json:"request_count"`
	ActiveUsers  int `json:"active_users"`
}

// swagger:parameters ListUserActivityUsers
type ListUserActivityUsersReq struct {
	// in:query
	// Required: true
	FilterDateFrom string `json:"filter_date_from" query:"filter_date_from" validate:"required"`
	// in:query
	// Required: true
	FilterDateTo string `json:"filter_date_to" query:"filter_date_to" validate:"required"`
	// in:query
	// Required: true
	PageIndex uint32 `json:"page_index" query:"page_index" validate:"required"`
	// in:query
	// Required: true
	PageSize uint32 `json:"page_size" query:"page_size" validate:"required"`
}

// swagger:model ListUserActivityUsersReply
type ListUserActivityUsersReply struct {
	Data      []UserActivityUserItem `json:"data"`
	TotalNums uint64                 `json:"total_nums"`
	base.GenericResp
}

type UserActivityUserItem struct {
	UserUID       string     `json:"user_uid"`
	UserName      string     `json:"user_name"`
	ActiveDays    int        `json:"active_days"`
	RequestCount  int        `json:"request_count"`
	TopModuleCode string     `json:"top_module_code"`
	TopModuleName string     `json:"top_module_name"`
	LastActiveAt  *time.Time `json:"last_active_at"`
}
