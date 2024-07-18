package v1

import (
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	pkgParams "github.com/actiontech/dms/pkg/params"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

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
}

// swagger:model GetDBServiceSyncTaskReply
type GetDBServiceSyncTaskReply struct {
	Data *GetDBServiceSyncTask `json:"data"`

	// Generic reply
	base.GenericResp
}

type GetDBServiceSyncTask struct {
	DBServiceSyncTask
	UID string `json:"uid"`
}

type ListDBServiceSyncTask struct {
	DBServiceSyncTask
	UID string `json:"uid"`

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
	// additional params
	// Required: false
	AdditionalParam pkgParams.Params `json:"additional_params"`
	// SQLE config
	SQLEConfig *dmsCommonV1.SQLEConfig `json:"sqle_config"`
}


// swagger:parameters AddDBServiceSyncTask
type AddDBServiceSyncTaskReq struct {
	// in:body
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

// swagger:parameters UpdateDBServiceSyncTask
type UpdateDBServiceSyncTaskReq struct {
	// Required: true
	// in:path
	DBServiceSyncTaskUid string `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
	// Required: true
	// in:body
	DBServiceSyncTask DBServiceSyncTask `json:"db_service_sync_task" validate:"required"`
}

// swagger:parameters DeleteDBServiceSyncTask
type DeleteDBServiceSyncTaskReq struct {
	// Required: true
	// in:path
	DBServiceSyncTaskUid string `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
}

// swagger:model ListDBServiceSyncTaskTipsReply
type ListDBServiceSyncTaskTipsReply struct {
	Data []DBServiceSyncTaskTip `json:"data"`

	// Generic reply
	base.GenericResp
}

type DBServiceSyncTaskTip struct {
	Type   pkgConst.DBServiceSourceName `json:"service_source_name"`
	Desc   string                       `json:"description"`
	DBType []pkgConst.DBType            `json:"db_type"` // 使用constant.DBType
	Params pkgParams.Params             `json:"params,omitempty"`
}

// swagger:parameters SyncDBServices
type SyncDBServicesReq struct {
	// Required: true
	// in:path
	DBServiceSyncTaskUid string `param:"db_service_sync_task_uid" json:"db_service_sync_task_uid" validate:"required"`
}
