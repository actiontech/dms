package v2

import (
	"bytes"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/go-openapi/strfmt"
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

// swagger:parameters ImportDBServicesOfProjectsCheckV2
type ImportDBServicesOfProjectsCheckReq struct {
	// DBServices file.
	//
	// in: formData
	//
	// swagger:file
	DBServicesFile *bytes.Buffer `json:"db_services_file"`
}

// swagger:model DBServiceV2
type DBService struct {
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
	EnvironmentTagUID string `json:"environment_tag_uid" validate:"required"`
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
	DBService  *DBService `json:"db_service" validate:"required"`
}

// swagger:model UpdateDBServiceReqV2
type UpdateDBServiceReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	DBServiceUid string           `param:"db_service_uid" json:"db_service_uid" validate:"required"`
	DBService    *UpdateDBService `json:"db_service" validate:"required"`
}

// swagger:model UpdateDBServiceV2
type UpdateDBService struct {
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
	Password *string `json:"password"`
	// DB Service environment tag
	// Required: true
	EnvironmentTag *dmsCommonV1.EnvironmentTag `json:"environment_tag"`
	// DB Service maintenance time
	// Required: true
	MaintenanceTimes []*dmsCommonV1.MaintenanceTime `json:"maintenance_times"`
	// DB Service Custom connection parameters
	// Required: false
	AdditionalParams []*dmsCommonV1.AdditionalParam `json:"additional_params"`
	// Service description
	Desc *string `json:"desc"`
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

// swagger:model ImportDBServicesOfOneProjectReqV2
type ImportDBServicesOfOneProjectReq struct {
	// swagger:ignore
	ProjectUid string            `param:"project_uid" json:"project_uid" validate:"required"`
	DBServices []ImportDBService `json:"db_services" validate:"required"`
}

// swagger:model ImportDBServiceV2
type ImportDBService struct {
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

// swagger:model ListGlobalDBServicesReplyV2
type ListGlobalDBServicesReply struct {
	// List global db service reply
	Data  []*ListGlobalDBService `json:"data"`
	Total int64                  `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

// swagger:model ListGlobalDBServiceV2
type ListGlobalDBService struct {
	// db service uid
	DBServiceUid string `json:"uid"`
	// db service name
	Name string `json:"name"`
	// db service DB type
	DBType string `json:"db_type"`
	// db service host
	Host string `json:"host"`
	// db service port
	Port string `json:"port"`
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
	// db service project_name
	ProjectName string `json:"project_name"`
	// is enable audit
	IsEnableAudit bool `json:"is_enable_audit"`
	// is enable masking
	IsEnableMasking bool `json:"is_enable_masking"`
	// db service unfinished workflow num
	UnfinishedWorkflowNum int64 `json:"unfinished_workflow_num"`
	// backup switch
	EnableBackup bool `json:"enable_backup"`
	// backup switch
	// Required: false
	BackupMaxRows uint64 `json:"backup_max_rows"`
	// DB connection test time
	LastConnectionTestTime strfmt.DateTime `json:"last_connection_test_time"`
	// DB connect test status
	LastConnectionTestStatus dmsCommonV1.LastConnectionTestStatus `json:"last_connection_test_status"`
	// DB connect test error message
	LastConnectionTestErrorMessage string `json:"last_connection_test_error_message,omitempty"`
}
