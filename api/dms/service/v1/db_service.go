package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

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
}

// swagger:parameters AddDBService
type AddDBServiceReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
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

// swagger:parameters CheckDBServiceIsConnectable
type CheckDBServiceIsConnectableReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// check db_service is connectable
	// in:body
	DBService dmsCommonV1.CheckDbConnectable `json:"db_service"`
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

// swagger:parameters CheckDBServiceIsConnectableById
type CheckDBServiceIsConnectableByIdReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// db service uid
	// in:path
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

// swagger:parameters UpdateDBService
type UpdateDBServiceReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
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
	// enum: create_audit_plan,create_workflow,sql_manage,create_export_task
	// in:query
	FunctionalModule string `json:"functional_module" query:"functional_module" validate:"omitempty,oneof=create_audit_plan create_workflow sql_manage create_export_task"`
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
