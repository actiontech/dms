package v2

import v1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"

// swagger:parameters ListDBServicesV2
type ListDBServiceReq struct {
	// the maximum count of db service to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of users to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy v1.DBServiceOrderByField `query:"order_by" json:"order_by"`
	// the db service business name
	// in:query
	FilterByBusiness string `query:"filter_by_business" json:"filter_by_business"`
	// the db service connection
	// enum: connect_success,connect_failed
	// in:query
	FilterLastConnectionTestStatus *string `query:"filter_last_connection_test_status" json:"filter_last_connection_test_status" validate:"omitempty,oneof=connect_success connect_failed"`
	// the db service host
	// in:query
	FilterByHost string `query:"filter_by_host" json:"filter_by_host"`
	// the db service uid
	// in:query
	FilterByUID string `query:"filter_by_uid" json:"filter_by_uid"`
	// the db service name
	// in:query
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
	// the db service port
	// in:query
	FilterByPort string `query:"filter_by_port" json:"filter_by_port"`
	// the db service db type
	// in:query
	FilterByDBType string `query:"filter_by_db_type" json:"filter_by_db_type"`
	// project id
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid"`
	// filter db services by db service id list using in condition
	// in:query
	FilterByDBServiceIds []string `query:"filter_by_db_service_ids" json:"filter_by_db_service_ids"`
	// filter db services by environment tag
	// in:query
	FilterByEnvironmentTag string `query:"filter_by_environment_tag" json:"filter_by_environment_tag"`
	// the db service fuzzy keyword,include host/port
	// in:query
	FuzzyKeyword string `query:"fuzzy_keyword" json:"fuzzy_keyword"`
	// is masking
	// in:query
	IsEnableMasking *bool `query:"is_enable_masking" json:"is_enable_masking"`
}