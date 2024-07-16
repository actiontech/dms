package v1

import (
	"time"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// swagger:parameters ListDBServiceSyncTasks
type ListDBServiceSyncTasksReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model ListDBServiceSyncTasksReply
type ListDBServiceSyncTasksReply struct {
	Data []*ListDBServiceSyncTask `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters GetDBServiceSyncTask
type GetDBServiceSyncTaskReq struct {
	// Required: true
	// in:path
	DBServiceSyncTaskUid string `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetDBServiceSyncTaskReply
type GetDBServiceSyncTaskReply struct {
	Data *GetDBServiceSyncTask `json:"data"`

	// Generic reply
	base.GenericResp
}

type GetDBServiceSyncTask struct {
	DBServiceSyncTask
	UID        string `json:"uid"`
	ProjectUid string `json:"project_uid"`
}

type ListDBServiceSyncTask struct {
	DBServiceSyncTask
	UID        string `json:"uid"`
	ProjectUid string `json:"project_uid"`

	// last sync error message
	LastSyncErr         string     `json:"last_sync_err"`
	LastSyncSuccessTime *time.Time `json:"last_sync_success_time"`
}

type DBServiceSyncTask struct {
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
type AddDBServiceSyncTaskReq struct {
	// swagger:ignore
	ProjectUid        string            `param:"project_uid" json:"project_uid" validate:"required"`
	DBServiceSyncTask DBServiceSyncTask `json:"db_service_sync_task"`
}

// swagger:model AddDBServiceSyncTaskReply
type AddDBServiceSyncTaskReply struct {
	// add database source service reply
	Data struct {
		// db service UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:model
type UpdateDBServiceSyncTaskReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	DBServiceSyncTaskUid string            `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
	DBServiceSyncTask    DBServiceSyncTask `json:"db_service_sync_task" validate:"required"`
}

// swagger:parameters DeleteDBServiceSyncTask
type DeleteDBServiceSyncTaskReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DBServiceSyncTaskUid string `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
}

type DatabaseSource struct {
	// database type
	// example: MySQL
	DbTypes []string `json:"db_types"`
	// database source
	// example: actiontech-dmp
	Source string `json:"source"`
}

// swagger:model ListDBServiceSyncTaskTipsReply
type ListDBServiceSyncTaskTipsReply struct {
	Data []*DatabaseSource `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ListDBServiceSyncTaskTips
type ListDBServiceSyncTaskTipsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model
type SyncDBServicesReq struct {
	ProjectUid           string `param:"project_uid" json:"project_uid" validate:"required"`
	DBServiceSyncTaskUid string `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
}
