package v1

import (
	"time"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// swagger:parameters ListDatabaseSourceServices
type ListDatabaseSourceServicesReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model ListDatabaseSourceServicesReply
type ListDatabaseSourceServicesReply struct {
	Data []*ListDatabaseSourceService `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters GetDatabaseSourceService
type GetDatabaseSourceServiceReq struct {
	// Required: true
	// in:path
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetDatabaseSourceServiceReply
type GetDatabaseSourceServiceReply struct {
	Data *GetDatabaseSourceService `json:"data"`

	// Generic reply
	base.GenericResp
}

type GetDatabaseSourceService struct {
	DatabaseSourceService
	UID        string `json:"uid"`
	ProjectUid string `json:"project_uid"`
}

type ListDatabaseSourceService struct {
	DatabaseSourceService
	UID        string `json:"uid"`
	ProjectUid string `json:"project_uid"`

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
	// SQLE config
	SQLEConfig *dmsCommonV1.SQLEConfig `json:"sqle_config"`
}

// swagger:model
type AddDatabaseSourceServiceReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	DatabaseSourceService DatabaseSourceService `json:"database_source_service"`
}

// swagger:model AddDatabaseSourceServiceReply
type AddDatabaseSourceServiceReply struct {
	// add database source service reply
	Data struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:model
type UpdateDatabaseSourceServiceReq struct {
    // swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
	DatabaseSourceService DatabaseSourceService `json:"database_source_service" validate:"required"`
}

// swagger:parameters DeleteDatabaseSourceService
type DeleteDatabaseSourceServiceReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
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
	Data []*DatabaseSource `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ListDatabaseSourceServiceTips
type ListDatabaseSourceServiceTipsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model
type SyncDatabaseSourceServiceReq struct {
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	DatabaseSourceServiceUid string `param:"database_source_service_uid" json:"database_source_service_uid" validate:"required"`
}
