package v1

import (
	"fmt"

	base "github.com/actiontech/dms/api/base/v1"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/go-openapi/strfmt"
)

// swagger:enum DBType
type DBType string

const (
	DBTypeMySQL          DBType = "MySQL"
	DBTypeOceanBaseMySQL DBType = "OceanBaseMySQL"
)

func ParseDBType(s string) (DBType, error) {
	switch s {
	case string(DBTypeMySQL):
		return DBTypeMySQL, nil
	case string(DBTypeOceanBaseMySQL):
		return DBTypeOceanBaseMySQL, nil
	default:
		return "", fmt.Errorf("invalid db type: %s", s)
	}
}

// A db service
type DBService struct {
	// Service name
	// Required: true
	Name string `json:"name" validate:"required"`
	// Service DB type
	// Required: true
	DBType DBType `json:"db_type" validate:"required"`
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
	MaintenanceTimes []*MaintenanceTime `json:"maintenance_times"`
	// DB Service namespace id
	// Required: true
	NamespaceUID string `json:"namespace_uid" validate:"required"`
	// DB Service Custom connection parameters
	// Required: false
	AdditionalParams []*dmsCommonV1.AdditionalParam `json:"additional_params"`
	// Service description
	Desc string `json:"desc"`
	// SQLE config
	SQLEConfig *SQLEConfig `json:"sqle_config"`
}

type MaintenanceTime struct {
	MaintenanceStartTime *Time `json:"maintenance_start_time"`
	MaintenanceStopTime  *Time `json:"maintenance_stop_time"`
}

type Time struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// swagger:parameters AddDBService
type AddDBServiceReq struct {
	// Add new db service
	// in:body
	DBService *DBService `json:"db_service" validate:"required"`
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
	Payload struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

func (u *AddDBServiceReply) String() string {
	if u == nil {
		return "AddDBServiceReply{nil}"
	}
	return fmt.Sprintf("AddDBServiceReply{Uid:%s}", u.Payload.Uid)
}

// swagger:parameters CheckDBServiceIsConnectable
type CheckDBServiceIsConnectableReq struct {
	// check db_service is connectable
	// in:body
	DBService dmsCommonV1.CheckDbConnectable `json:"db_service"`
}

// swagger:model CheckDBServiceIsConnectableReply
type CheckDBServiceIsConnectableReply struct {
	Payload struct {
		IsConnectable       bool   `json:"is_connectable"`
		ConnectErrorMessage string `json:"connect_error_message,omitempty"`
	} `json:"payload"`

	base.GenericResp
}

// swagger:parameters DelDBService
type DelDBServiceReq struct {
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

// swagger:parameters ListDBServices
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
	OrderBy DBServiceOrderByField `query:"order_by" json:"order_by"`
	// the db service business name
	// in:query
	FilterByBusiness string `query:"filter_by_business" json:"filter_by_business"`
	// the db service host
	// in:query
	FilterByHost string `query:"filter_by_host" json:"filter_by_host"`
	// the db service uid
	FilterByUID string `query:"filter_by_uid" json:"filter_by_uid"`
	// the db service port
	// in:query
	FilterByPort string `query:"filter_by_port" json:"filter_by_port"`
	// the db service db type
	// in:query
	// Multiple of ["MySQL","OceanBaseMySQL"], default is [""]
	FilterByDBType string `query:"filter_by_db_type" json:"filter_by_db_type"`
	// filter by db service namespace uid
	// only the sys user can use an empty namespace value, which means lookup from all namespaces
	// in:query
	FilterByNamespaceUid string `query:"filter_by_namespace_uid" json:"filter_by_namespace_uid"`
}

// swagger:enum DBServiceOrderByField
type DBServiceOrderByField string

const (
	DBServiceOrderByName DBServiceOrderByField = "name"
)

// A dms db Service
type ListDBService struct {
	// db service uid
	DBServiceUid string `json:"uid"`
	// db service name
	Name string `json:"name"`
	// Multiple of ["MySQL"], default is ["MySQL"]
	// db service DB type
	DBType DBType `json:"db_type"`
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
	MaintenanceTimes []*MaintenanceTime `json:"maintenance_times"`
	// DB desc
	Desc string `json:"desc"`
	// DB source
	Source string `json:"source"`
	// DB namespace uid
	NamespaceUID string `json:"namespace_uid"`
	// sqle config
	SQLEConfig *SQLEConfig `json:"sqle_config"`
	// auth config
	AuthConfig *AuthSyncConfig `json:"auth_config"`
}

type SQLEConfig struct {
	// DB Service rule template name
	RuleTemplateName string `json:"rule_template_name"`
	// DB Service rule template id
	RuleTemplateID string `json:"rule_template_id"`
	// DB Service SQL query config
	SQLQueryConfig *SQLQueryConfig `json:"sql_query_config"`
}

// swagger:enum SQLAllowQueryAuditLevel
type SQLAllowQueryAuditLevel string

const (
	AuditLevelNormal SQLAllowQueryAuditLevel = "normal"
	AuditLevelNotice SQLAllowQueryAuditLevel = "notice"
	AuditLevelWarn   SQLAllowQueryAuditLevel = "warn"
	AuditLevelError  SQLAllowQueryAuditLevel = "error"
)

type SQLQueryConfig struct {
	MaxPreQueryRows                  int                     `json:"max_pre_query_rows" example:"100"`
	QueryTimeoutSecond               int                     `json:"query_timeout_second" example:"10"`
	AuditEnabled                     bool                    `json:"audit_enabled" example:"false"`
	AllowQueryWhenLessThanAuditLevel SQLAllowQueryAuditLevel `json:"allow_query_when_less_than_audit_level" enums:"normal,notice,warn,error" valid:"omitempty,oneof=normal notice warn error " example:"error"`
}

type AuthSyncConfig struct {
	// last sync data result
	LastSyncDataResult string `json:"last_sync_data_result"`
	// last sync data time
	LastSyncDataTime strfmt.DateTime `json:"last_sync_data_time"`
}

// swagger:model ListDBServiceReply
type ListDBServiceReply struct {
	// List db service reply
	Payload struct {
		DBServices []*ListDBService `json:"db_services"`
		Total      int64            `json:"total"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters UpdateDBService
type UpdateDBServiceReq struct {
	// db_service_uid
	// Required: true
	// in:path
	DBServiceUid string `param:"db_service_uid" json:"db_service_uid" validate:"required"`
	// Update a DB service
	// in:body
	DBService *UpdateDBService `json:"db_service" validate:"required"`
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
	DBType DBType `json:"db_type" validate:"required"`
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
	MaintenanceTimes []*MaintenanceTime `json:"maintenance_times"`
	// DB Service Custom connection parameters
	// Required: false
	AdditionalParams []*dmsCommonV1.AdditionalParam `json:"additional_params"`
	// Service description
	Desc *string `json:"desc"`
	// SQLE config
	SQLEConfig *SQLEConfig `json:"sqle_config"`
}

// swagger:model UpdateDBServiceReply
type UpdateDBServiceReply struct {
	// update db service reply
	Payload struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

func (u *UpdateDBServiceReply) String() string {
	if u == nil {
		return "UpdateDBServiceReply{nil}"
	}
	return fmt.Sprintf("UpdateDBServiceReply{Uid:%s}", u.Payload.Uid)
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
	Payload struct {
		DatabaseDriverOptions []*DatabaseDriverOption `json:"database_driver_options"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}
