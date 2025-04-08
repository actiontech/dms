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

// swagger:model ImportDBServicesOfProjectsReqV2
type ImportDBServicesOfProjectsReq struct {
	DBServices []ImportDBServiceV2 `json:"db_services" validate:"required"`
}

type ImportDBServiceV2 struct {
	// db service name
	Name string `json:"name"`
	// db service DB type
	DBType string `json:"db_type"`
	// db service host
	Host string `json:"host"`
	// db service port
	Port string `json:"port"`
	// db service admin user
	User string `json:"user"`
	// db service admin encrypted password
	Password string `json:"password"`
	// DB Service environment tag
	// Required: true
	EnvironmentTag *dmsCommonV1.EnvironmentTag `json:"environment_tag"`
	// DB Service maintenance time
	MaintenanceTimes []*dmsCommonV1.MaintenanceTime `json:"maintenance_times"`
	// DB desc
	Desc string `json:"desc"`
	// DB source
	Source string `json:"source"`
	// DB project uid
	ProjectUID string `json:"project_uid"`
	// sqle config
	SQLEConfig *dmsCommonV1.SQLEConfig `json:"sqle_config"`
	// DB Service Custom connection parameters
	AdditionalParams []*dmsCommonV1.AdditionalParam `json:"additional_params"`
	// is enable masking
	IsEnableMasking bool `json:"is_enable_masking"`
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

// A db service
type DBServicev2 struct {
	// Service name
	// Required: true
	Name string `json:"name" validate:"required"`
	// Service DB type
	// Required: true
	DBType string `json:"db_type" validate:"required"`
	// DB Service Host
	// Required: true
	Host string `json:"host" validate:"required,ip_addr|uri|hostname|hostname_rfc1123"`
	// DB Service port
	// Required: true
	Port string `json:"port" validate:"required"`
	// DB Service admin user
	// Required: true
	User string `json:"user" validate:"required"`
	// DB Service admin password
	// Required: true
	Password string `json:"password" validate:"required"`
	// DB Service environment tag
	// Required: true
	EnvironmentTag *dmsCommonV1.EnvironmentTag `json:"environment_tag"`
	// DB Service maintenance time
	// empty value means that maintenance time is unlimited
	// Required: true
	MaintenanceTimes []*dmsCommonV1.MaintenanceTime `json:"maintenance_times"`
	// DB Service Custom connection parameters
	// Required: false
	AdditionalParams []*dmsCommonV1.AdditionalParam `json:"additional_params"`
	// Service description
	Desc string `json:"desc"`
	// SQLE config
	SQLEConfig *dmsCommonV1.SQLEConfig `json:"sqle_config"`
	// data masking switch
	// Required: false
	IsEnableMasking bool `json:"is_enable_masking"`
	// backup switch
	// Required: false
	EnableBackup bool `json:"enable_backup"`
	// backup switch
	// Required: false
	BackupMaxRows *uint64 `json:"backup_max_rows,omitempty"`
}

// swagger:model AddDBServiceReqV2
type AddDBServiceReq struct {
	// swagger:ignore
	ProjectUid string     `param:"project_uid" json:"project_uid" validate:"required"`
	DBService  *DBServicev2 `json:"db_service" validate:"required"`
}
