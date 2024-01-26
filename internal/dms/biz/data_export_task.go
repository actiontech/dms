package biz

import (
	"context"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

type DataExportTaskStatus string

// 导出任务状态常量
const (
	DataExportTaskStatusInit      DataExportTaskStatus = "init"
	DataExportTaskStatusExporting DataExportTaskStatus = "exporting"
	DataExportTaskStatusFinish    DataExportTaskStatus = "finish"
	DataExportTaskStatusFailed    DataExportTaskStatus = "failed"
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

	ExportStatus    DataExportTaskStatus
	ExportStartTime *time.Time
	ExportEndTime   *time.Time

	DataExportTaskRecords []*DataExportTaskRecord
}

type DataExportTaskRecord struct {
	Number           uint
	DataExportTaskId string
	ExportSQL        string
	AuditLevel       string
	ExportResult     string
	AuditSQLResults  []*AuditResult
}

type AuditResult struct {
	Level    string `json:"level" example:"warn"`
	Message  string `json:"message" example:"避免使用不必要的内置函数md5()"`
	RuleName string `json:"rule_name"`
}

type ListDataExportTaskRecordOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      DataExportTaskRecordField
	FilterBy     []pkgConst.FilterCondition
}
type DataExportTaskRepo interface {
	SaveDataExportTask(ctx context.Context, dataExportDataExportTasks []*DataExportTask) error
	GetDataExportTaskByIds(ctx context.Context, ids []string) (dataExportDataExportTasks []*DataExportTask, err error)
	ListDataExportTaskRecord(ctx context.Context, opt *ListDataExportTaskRecordOption) (dataExportTaskRecords []*DataExportTaskRecord, total int64, err error)
	BatchUpdateDataExportTaskStatusByIds(ctx context.Context, ids []string, status DataExportTaskStatus) (err error)
}
