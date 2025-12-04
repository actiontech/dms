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

	ExportStatus    DataExportTaskStatus
	ExportStartTime *time.Time
	ExportEndTime   *time.Time
	DbService       *DBService

	DataExportTaskRecords []*DataExportTaskRecord
}

func (t *DataExportTask) InstanceName() string {
	if t.DbService != nil {
		return t.DbService.Name
	}
	return ""
}

type DataExportTaskRecord struct {
	Number           uint
	DataExportTaskId string
	ExportSQL        string
	AuditLevel       string
	ExportResult     string
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
