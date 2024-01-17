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
	TaskID string     `json:"task_id"`
	DBInfo TaskDBInfo `json:"db_info"`
	Status string     `json:"status"`

	// SQL审核结果
	AuditLevel string  `json:"audit_level" enums:"normal,notice,warn,error,"`
	Score      int32   `json:"score"`
	PassRate   float64 `json:"pass_rate"`

	// 导出SQL的处理信息
	ExecSQLRecord []SQLRecord `json:"sql_record"`
}

type SQLRecord struct {
	ID           string        `json:"uid"`
	SQL          string        `json:"sql"`
	CheckResults []CheckResult `json:"results"`
	// 状态、进度、下载次数等
}

type CheckResult struct {
}
