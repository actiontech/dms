package v1

import (
	"time"

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
	DataExportTasks []DataExportTask `json:"data_export_tasks"`
}

type DataExportTask struct {
	// DB Service uid
	// Required: true
	DBServiceUid string `json:"db_service_uid" validate:"required"`
	// DB Service name
	// Required: false
	DatabaseName string `json:"database_name"`
	// The exported SQL statement executed. it's necessary when ExportType is SQL
	// SELECT * FROM DMS_test LIMIT 20;
	ExportSQL string `json:"export_sql"`
}

// swagger:model AddDataExportTaskReply
type AddDataExportTaskReply struct {
	// add data export workflow reply
	Data struct {
		// data export task UIDs
		Uids []string `json:"data_export_task_uids"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters BatchGetDataExportTask
type BatchGetDataExportTaskReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in: query
	TaskUids string `query:"data_export_task_uids" json:"data_export_task_uids" validate:"required"`
}

// swagger:model BatchGetDataExportTaskReply
type BatchGetDataExportTaskReply struct {
	Data []*GetDataExportTask `json:"data"`

	// Generic reply
	base.GenericResp
}

type TaskDBInfo struct {
	UidWithName
	DBType       string `json:"db_type"`
	DatabaseName string `json:"database_name"`
}

// swagger:enum DataExportTaskStatus
type DataExportTaskStatus string

// 导出任务状态常量
const (
	StatusInit        DataExportTaskStatus = "init"
	StatusExporting   DataExportTaskStatus = "exporting"
	StatusFinish      DataExportTaskStatus = "finish"
	StatusFailed      DataExportTaskStatus = "failed"
	StatusFileDeleted DataExportTaskStatus = "file_deleted"
)

type GetDataExportTask struct {
	TaskUid         string               `json:"task_uid"`
	DBInfo          TaskDBInfo           `json:"db_info"`
	Status          DataExportTaskStatus `json:"status"`
	ExportStartTime *time.Time           `json:"export_start_time,omitempty"`
	ExportEndTime   *time.Time           `json:"export_end_time,omitempty"`
	FileName        string               `json:"file_name"` // 导出文件名
	AuditResult     AuditTaskResult      `json:"audit_result"`
	ExportType      string               `json:"export_type"`      // Export Type example: SQL Meta
	ExportFileType  string               `json:"export_file_type"` // Export Content example: CSV SQL EXCEL

}

// SQL审核结果
type AuditTaskResult struct {
	AuditLevel string  `json:"audit_level" enums:"normal,notice,warn,error,"`
	Score      int32   `json:"score"`
	PassRate   float64 `json:"pass_rate"`
}

// swagger:parameters ListDataExportTaskSQLs
type ListDataExportTaskSQLsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportTaskUid string `param:"data_export_task_uid" json:"data_export_task_uid" validate:"required"`
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
	ID             uint             `json:"uid"`
	ExportSQL      string           `json:"sql"`
	ExportResult   string           `json:"export_result"` // 导出结果
	ExportSQLType  string           `json:"export_sql_type"`
	AuditLevel     string           `json:"audit_level"`
	AuditSQLResult []AuditSQLResult `json:"audit_sql_result"`
}
type AuditSQLResult struct {
	Level    string `json:"level" example:"warn"`
	Message  string `json:"message" example:"避免使用不必要的内置函数md5()"`
	RuleName string `json:"rule_name"`
	DBType   string `json:"db_type"`
}

// swagger:parameters DownloadDataExportTask
type DownloadDataExportTaskReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportTaskUid string `param:"data_export_task_uid" json:"data_export_task_uid" validate:"required"`
}

// swagger:response DownloadDataExportTaskReply
type DownloadDataExportTaskReply struct {
	// swagger:file
	// in:  body
	File []byte
}

// swagger:parameters DownloadDataExportTaskSQLs
type DownloadDataExportTaskSQLsReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportTaskUid string `param:"data_export_task_uid" json:"data_export_task_uid" validate:"required"`
}
