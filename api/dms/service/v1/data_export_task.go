package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters AddDataExportTask
type AddDataExportTaskReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// add data export workflow
	// in:body
	DataExportTask []DataExportTask `json:"data_export_workflow"`
}

type DataExportTask struct {
	// DB Service uid
	// Required: true
	DBServiceUid string `json:"db_service_uid" validate:"required"`
	// DB Service name
	// Required: false
	DatabaseName string `json:"database_name"`
	// The exported SQL statement executed. it's necessary when ExportType is SQL
	// Required: true
	// SELECT * FROM DMS_test LIMIT 20;
	ExecSQL string `json:"exec_sql" validate:"required"`
}

// swagger:model AddDataExportTaskReply
type AddDataExportTaskReply struct {
	// add data export workflow reply
	Data struct {
		// data export task UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters GetDataExportTask
type GetDataExportTaskReq struct {
	// Required: true
	// in:path
	DataExportTaskUid string `param:"data_export_task_uid" json:"data_export_task_uid" validate:"required"`
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetDataExportTaskReply
type GetDataExportTaskReply struct {
	Data *GetDataExportTask `json:"data"`

	// Generic reply
	base.GenericResp
}

type TaskDBInfo struct {
	UidWithName
	DatabaseName string `json:"database_name"`
}

// 导出任务状态常量
const (
	StatusExpired string = "已过期"
	StatusFinish  string = "已完成"
	StatusInit    string = "未导出"
	StatusSuccess string = "导出成功"
	StatusFailed  string = "导出失败"
)

type GetDataExportTask struct {
	TaskUid string     `json:"task_uid"`
	DBInfo  TaskDBInfo `json:"db_info"`
	Status  string     `json:"status"`

	InstanceName      string `json:"instance_name"`
	ExecStartTime     string `json:"exec_start_time"`
	ExecEndTime       string `json:"exec_end_time"`
	TaskPassRate      int    `json:"task_pass_rate"`
	TaskScore         int    `json:"task_score"`
	ExecutionUserName string `json:"execution_user_name"`
	// SQL审核结果
	AuditLevel string  `json:"audit_level" enums:"normal,notice,warn,error,"`
	Score      int32   `json:"score"`
	PassRate   float64 `json:"pass_rate"`
	FileName   string  `json:"file_name"` // 导出文件名
	// 下载次数，进度等
}

// swagger:parameters ListDataExportTaskSQLs
type ListDataExportTaskSQLsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// the maximum count of member to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of members to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:model ListDataExportTaskSQLsReply
type ListDataExportTaskSQLsReply struct {
	Data  []*ListDataExportTaskSQL `json:"data"`
	Total int64                    `json:"total_nums"`
	// Generic reply
	base.GenericResp
}

type ListDataExportTaskSQL struct {
	ID        string `json:"uid"`
	FileName  string `json:"file_name"` // 导出文件名
	ExportSQL string `json:"sql"`

	ExportStatus string `json:"status"`        // 导出状态
	ExportResult string `json:"export_status"` // 导出结果

	AuditLevel  string `json:"audit_level"`
	AuditStatus string `json:"audit_status"`
	AuditResult string `json:"audit_result"` // 审核结果

}
