package biz

import (
	"context"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/storage/model"
)

type DataExportTaskStatus string

// 导出任务状态常量
const (
	DataExportTaskStatusInit       DataExportTaskStatus = "init"
	DataExportTaskStatusExporting  DataExportTaskStatus = "exporting"
	DataExportTaskStatusFinish     DataExportTaskStatus = "finish"
	DataExportTaskStatusFailed     DataExportTaskStatus = "failed"
	DataExportTaskStatusFileDelted DataExportTaskStatus = "file_deleted"
)

func (dets DataExportTaskStatus) String() string {
	return string(dets)
}

type DataExportTask struct {
	Base

	UID               string
	CreateUserUID     string
	DBServiceUid      string
	DatabaseName      string
	WorkFlowRecordUid string
	ExportType        string
	ExportFileType    string
	ExportFileName    string
	ExportSQL         string
	AuditPassRate     float64
	AuditScore        int32
	AuditLevel        string

	ExportStatus     DataExportTaskStatus
	ExportStartTime  *time.Time
	ExportEndTime    *time.Time
	ExportFailStage  string
	ExportFailReason string
	DbService        *DBService

	DataExportTaskRecords []*DataExportTaskRecord
}

// 任务级导出失败阶段（与 S2 §8.2 / swagger / UI 文案表一致）
const (
	DataExportFailStageTaskSchedule = "task_schedule"
	DataExportFailStageConnect      = "connect"
	DataExportFailStagePrepare      = "prepare"
	DataExportFailStageSQLExecute   = "sql_execute"
	DataExportFailStageFileGenerate = "file_generate"
)

func (t *DataExportTask) InstanceName() string {
	if t.DbService != nil {
		return t.DbService.Name
	}
	return ""
}

// SQL 级导出执行状态（与 S3 §8.2 wire 一致；成败以本字段为准，不以 export_result 判）
type DataExportSQLExportStatus string

const (
	DataExportSQLStatusSuccess     DataExportSQLExportStatus = "success"
	DataExportSQLStatusFailed      DataExportSQLExportStatus = "failed"
	DataExportSQLStatusNotExecuted DataExportSQLExportStatus = "not_executed"
)

func (s DataExportSQLExportStatus) String() string {
	return string(s)
}

// 未执行说明文案（产品锁定）
const DataExportSQLNotExecutedResult = "导出任务已失败，本条 SQL 未执行"

// 确无原因时任务/工单统一兜底（产品锁定）；不得覆盖已有业务错误
const DataExportFailReasonFallback = "导出失败，暂未获取到具体原因，请联系管理员查看服务日志"

type DataExportTaskRecord struct {
	Number           uint
	DataExportTaskId string
	ExportSQL        string
	AuditLevel       string
	ExportResult     string
	ExportStatus     DataExportSQLExportStatus
	ExportSQLType    string
	AuditSQLResults  model.AuditResults
}

type ListDataExportTaskRecordOption struct {
	PageNumber      uint32
	LimitPerPage    uint32
	OrderBy         DataExportTaskRecordField
	FilterByOptions pkgConst.FilterOptions
}
type ListDataExportTaskOption struct {
	PageNumber      uint32
	LimitPerPage    uint32
	OrderBy         DataExportTaskField
	FilterByOptions pkgConst.FilterOptions
}

type DataExportTaskRepo interface {
	SaveDataExportTask(ctx context.Context, dataExportDataExportTasks []*DataExportTask) error
	GetDataExportTaskByIds(ctx context.Context, ids []string) (dataExportDataExportTasks []*DataExportTask, err error)
	ListDataExportTaskRecord(ctx context.Context, opt *ListDataExportTaskRecordOption) (dataExportTaskRecords []*DataExportTaskRecord, total int64, err error)
	BatchUpdateDataExportTaskStatusByIds(ctx context.Context, ids []string, status DataExportTaskStatus) (err error)
	ListDataExportTasks(ctx context.Context, opt *ListDataExportTaskOption) (exportTasks []*DataExportTask, total int64, err error)
	DeleteUnusedDataExportTasks(ctx context.Context) error
	BatchUpdateDataExportTaskByIds(ctx context.Context, ids []string, args map[string]interface{}) error
	SaveDataExportTaskRecords(ctx context.Context, dataExportTaskRecords []*DataExportTaskRecord) error
}
