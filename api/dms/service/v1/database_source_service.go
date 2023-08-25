package v1

import (
	"time"

	base "github.com/actiontech/dms/api/base/v1"
)

// swagger:parameters ListDatabaseSourceServices
type ListDatabaseSourceServicesReq struct {
	// filter by db service namespace uid
	// only the sys user can use an empty namespace value, which means lookup from all namespaces
	// in:query
	NamespaceId string `query:"namespace_id" json:"namespace_id"`
}

// swagger:model ListDatabaseSourceServicesReply
type ListDatabaseSourceServicesReply struct {
	Payload struct {
		DatabaseSourceServices []*ListDatabaseSourceService `json:"database_source_services"`
	} `json:"payload"`
	// Generic reply
	base.GenericResp
}

// swagger:parameters GetDatabaseSourceService
type GetDatabaseSourceServiceReq struct {
	// Required: true
	// in:path
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
	// filter by db service namespace uid
	// only the sys user can use an empty namespace value, which means lookup from all namespaces
	// in:query
	NamespaceId string `query:"namespace_id" json:"namespace_id"`
}

// swagger:model GetDatabaseSourceServiceReply
type GetDatabaseSourceServiceReply struct {
	Payload struct {
		DatabaseSourceService *GetDatabaseSourceService `json:"database_source_service"`
	} `json:"payload"`
	// Generic reply
	base.GenericResp
}

type GetDatabaseSourceService struct {
	DatabaseSourceService
	UID string `json:"uid"`
}

type ListDatabaseSourceService struct {
	DatabaseSourceService
	UID string `json:"uid"`
	// last sync error message
	LastSyncErr         string     `json:"last_sync_err"`
	LastSyncSuccessTime *time.Time `json:"last_sync_success_time"`
}

type DatabaseSourceService struct {
	// name
	// Required: true
	// example: dmp
	Name string `json:"name" validate:"required"`
	// source
	// Required: true
	// example: actiontech-dmp
	Source string `json:"source" validate:"required"`
	// version
	// Required: true
	// example: 5.23.01.0
	Version string `json:"version" validate:"required"`
	// addr
	// Required: true
	// example: http://10.186.62.56:10000
	URL string `json:"url" validate:"required"`
	// database type
	// Required: true
	// example: MySQL
	DbType string `json:"db_type" validate:"required"`
	// cron expression
	// Required: true
	// example: 0 0 * * *
	CronExpress string `json:"cron_express" validate:"required"`
	// namespace id
	// Required: true
	NamespaceUID string `json:"namespace_uid" validate:"required"`
	// SQLE config
	SQLEConfig *SQLEConfig `json:"sqle_config"`
}

// swagger:parameters AddDatabaseSourceService
type AddDatabaseSourceServiceReq struct {
	// add database source service
	// in:body
	DatabaseSourceService DatabaseSourceService `json:"database_source_service"`
}

// swagger:model AddDatabaseSourceServiceReply
type AddDatabaseSourceServiceReply struct {
	// add database source service reply
	Payload struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"payload"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters UpdateDatabaseSourceService
type UpdateDatabaseSourceServiceReq struct {
	// Required: true
	// in:path
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
	// update database source service
	// in:body
	DatabaseSourceService DatabaseSourceService `json:"database_source_service" validate:"required"`
}

// swagger:parameters DeleteDatabaseSourceService
type DeleteDatabaseSourceServiceReq struct {
	// Required: true
	// in:path
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
}

type DatabaseSource struct {
	// database type
	// example: MySQL
	DbTypes []string `json:"db_types"`
	// database source
	// example: actiontech-dmp
	Source string `json:"source"`
}

// swagger:model ListDatabaseSourceServiceTipsReply
type ListDatabaseSourceServiceTipsReply struct {
	Payload struct {
		DatabaseSources []*DatabaseSource `json:"database_sources"`
	} `json:"payload"`
	// Generic reply
	base.GenericResp
}

// swagger:parameters SyncDatabaseSourceService
type SyncDatabaseSourceServiceReq struct {
	// Required: true
	// in:path
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
}
