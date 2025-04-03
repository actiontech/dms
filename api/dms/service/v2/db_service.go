package v2

import (
	"bytes"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// swagger:parameters ListGlobalDBServicesV2
type ListGlobalDBServicesReq struct {
	// the maximum count of db service to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of users to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy dmsCommonV1.DBServiceOrderByField `query:"order_by" json:"order_by"`
	// the db service connection
	// enum: connect_success,connect_failed
	// in:query
	FilterLastConnectionTestStatus *string `query:"filter_last_connection_test_status" json:"filter_last_connection_test_status" validate:"omitempty,oneof=connect_success connect_failed"`
	// TODO This parameter is deprecated and will be removed soon.
	// the db service business name
	// in:query
	FilterByBusiness string `query:"filter_by_business" json:"filter_by_business"`
	// filter db services by environment tag
	// in:query
	FilterByEnvironmentTag string `query:"filter_by_environment_tag" json:"filter_by_environment_tag"`
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
	// the db service project id
	// in:query
	FilterByProjectUid string `query:"filter_by_project_uid" json:"filter_by_project_uid"`
	// is masking
	// in:query
	FilterByIsEnableMasking *bool `query:"filter_by_is_enable_masking" json:"filter_by_is_enable_masking"`
	// the db service fuzzy keyword
	// in:query
	FuzzyKeyword string `query:"fuzzy_keyword" json:"fuzzy_keyword"`
}


// swagger:parameters ImportDBServicesOfOneProjectCheckV2
type ImportDBServicesOfOneProjectCheckReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// DBServices file.
	//
	// in: formData
	//
	// swagger:file
	DBServicesFile *bytes.Buffer `json:"db_services_file"`
}