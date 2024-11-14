package v1

import (
	"bytes"
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	"github.com/go-openapi/strfmt"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// A db service
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
	// DB Service business name
	// Required: true
	Business string `json:"business" validate:"required"`
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
}

// swagger:model AddDBServiceReq
type AddDBServiceReq struct {
	// swagger:ignore
	ProjectUid string     `param:"project_uid" json:"project_uid" validate:"required"`
	DBService  *DBService `json:"db_service" validate:"required"`
}

func (u *AddDBServiceReq) String() string {
	if u == nil {
		return "AddDBServiceReq{nil}"
	}
	return fmt.Sprintf("AddDBServiceReq{Name:%s,DBType:%s Host:%s}", u.DBService.Name, u.DBService.DBType, u.DBService.Host)
}

// swagger:model AddDBServiceReply
type AddDBServiceReply struct {
	// Add db service reply
	Data struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *AddDBServiceReply) String() string {
	if u == nil {
		return "AddDBServiceReply{nil}"
	}
	return fmt.Sprintf("AddDBServiceReply{Uid:%s}", u.Data.Uid)
}

// swagger:model
type CheckDBServiceIsConnectableReq struct {
	// swagger:ignore
	ProjectUid string                         `param:"project_uid" json:"project_uid" validate:"required"`
	DBService  dmsCommonV1.CheckDbConnectable `json:"db_service"`
}

type CheckDBServiceIsConnectableReplyItem struct {
	IsConnectable       bool   `json:"is_connectable"`
	Component           string `json:"component"`
	ConnectErrorMessage string `json:"connect_error_message"`
}

// swagger:model CheckDBServiceIsConnectableReply
type CheckDBServiceIsConnectableReply struct {
	Data []CheckDBServiceIsConnectableReplyItem `json:"data"`

	base.GenericResp
}

// swagger:model
type CheckDBServicesIsConnectableReq struct {
	// swagger:ignore
	ProjectUid string                 `param:"project_uid" json:"project_uid" validate:"required"`
	DBServices []DbServiceConnections `json:"db_services"`
}

// swagger:model CheckDBServicesIsConnectableReply
type CheckDBServicesIsConnectableReply struct {
	Data []DBServiceIsConnectableReply `json:"data"`
	base.GenericResp
}

type DBServiceIsConnectableReply struct {
	DBServiceUid        string          `param:"db_service_uid" json:"db_service_uid"`
	ConnectionStatus    string          `json:"connection_status"`
	TestConnectionTime  strfmt.DateTime `json:"test_connection_time"`
	ConnectErrorMessage string          `json:"connect_error_message"`
}

// swagger:model
type CheckDBServiceIsConnectableByIdReq struct {
	ProjectUid   string `param:"project_uid" json:"project_uid" validate:"required"`
	DBServiceUid string `param:"db_service_uid" json:"db_service_uid" validate:"required"`
}

// swagger:parameters DelDBService
type DelDBServiceReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// db service uid
	// in:path
	DBServiceUid string `param:"db_service_uid" json:"db_service_uid" validate:"required"`
}

func (u *DelDBServiceReq) String() string {
	if u == nil {
		return "DelDBServiceReq{nil}"
	}
	return fmt.Sprintf("DelDBServiceReq{Uid:%s}", u.DBServiceUid)
}

// swagger:model
type UpdateDBServiceReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	DBServiceUid string           `param:"db_service_uid" json:"db_service_uid" validate:"required"`
	DBService    *UpdateDBService `json:"db_service" validate:"required"`
}

func (u *UpdateDBServiceReq) String() string {
	if u == nil {
		return "UpdateDBServiceReq{nil}"
	}
	if u.DBService == nil {
		return "UpdateDBServiceReq{DBService:nil}"
	}
	return fmt.Sprintf("UpdateDBServiceReq{Uid:%s}", u.DBServiceUid)
}

// update db service
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
	// DB Service business name
	// Required: true
	Business string `json:"business" validate:"required"`
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
}

// swagger:model UpdateDBServiceReply
type UpdateDBServiceReply struct {
	// update db service reply
	Data struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

func (u *UpdateDBServiceReply) String() string {
	if u == nil {
		return "UpdateDBServiceReply{nil}"
	}
	return fmt.Sprintf("UpdateDBServiceReply{Uid:%s}", u.Data.Uid)
}

type DatabaseDriverAdditionalParam struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description" example:"参数项中文名"`
	Type        string `json:"type" example:"int"`
}

type DatabaseDriverOption struct {
	DBType   string                           `json:"db_type"`
	LogoPath string                           `json:"logo_path"`
	Params   []*DatabaseDriverAdditionalParam `json:"params"`
}

// swagger:model ListDBServiceDriverOptionReply
type ListDBServiceDriverOptionReply struct {
	// List db service reply
	Data []*DatabaseDriverOption `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ListDBServiceTips
type ListDBServiceTipsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: false
	// in:query
	FilterDBType string `json:"filter_db_type" query:"filter_db_type"`
	// Required: false
	// enum: save_audit_plan,create_workflow,create_export_task
	// in:query
	FunctionalModule string `json:"functional_module" query:"functional_module" validate:"omitempty,oneof=save_audit_plan create_workflow create_export_task"`
}

type ListDBServiceTipItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"db_type"`
	Host string `json:"host"`
	Port string `json:"port"`
}

// swagger:model ListDBServiceTipsReply
type ListDBServiceTipsReply struct {
	// List db service reply
	Data []*ListDBServiceTipItem `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ImportDBServicesOfOneProjectCheck
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
	// the db service business name
	Business string `json:"business"`
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

// swagger:model
type ImportDBServicesOfOneProjectReq struct {
	// swagger:ignore
	ProjectUid string            `param:"project_uid" json:"project_uid" validate:"required"`
	DBServices []ImportDBService `json:"db_services" validate:"required"`
}

// swagger:model ImportDBServicesCheckReply
type ImportDBServicesCheckReply struct {
	// db services
	Data []*ImportDBService `json:"data"`
	// Generic reply
	base.GenericResp
}

// swagger:parameters ListGlobalDBServices
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
	FilterConnectionStatus *string `query:"filter_connection_status" json:"filter_connection_status" validate:"omitempty,oneof=connect_success connect_failed"`
	// the db service business name
	// in:query
	FilterByBusiness string `query:"filter_by_business" json:"filter_by_business"`
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

// swagger:model ListGlobalDBServicesReply
type ListGlobalDBServicesReply struct {
	// List global db service reply
	Data  []*ListGlobalDBService `json:"data"`
	Total int64                  `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

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
	// the db service business name
	Business string `json:"business"`
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
	// DB connection test time
	LastConnectionTestTime strfmt.DateTime `json:"last_connection_test_time"`
	// DB connect test status
	LastConnectionTestStatus string `json:"last_connection_test_status"`
	// DB connect test error message
	LastConnectionTestErrorMessage string `json:"last_connection_test_error_message,omitempty"`
}

// swagger:model ListGlobalDBServicesTipsReply
type ListGlobalDBServicesTipsReply struct {
	// List global db service tips reply
	Data *ListGlobalDBServiceTips `json:"data"`

	// Generic reply
	base.GenericResp
}

type ListGlobalDBServiceTips struct {
	DBType []string `json:"db_type"`
}
