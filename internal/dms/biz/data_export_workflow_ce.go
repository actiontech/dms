//go:build !enterprise

package biz

import (
	"context"
	"errors"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

var errNotDataExportWorkflow = errors.New("data export workflow related functions are enterprise version functions")
var errNotDataExportTask = errors.New("data export task related functions are enterprise version functions")

func (d *DataExportWorkflowUsecase) AddDataExportWorkflow(ctx context.Context, currentUserId string, params *Workflow) (string, error) {
	return "", errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) ListDataExportWorkflows(ctx context.Context, workflowsOption *ListWorkflowsOption, currentUserId, dbServiceUID string) ([]*Workflow, int64, error) {
	return nil, 0, errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) GetDataExportWorkflow(ctx context.Context, workflowUid, currentUserId string) (*Workflow, error) {
	return nil, errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) ApproveDataExportWorkflow(ctx context.Context, projectId, workflowId, userId string) error {
	return errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) RejectDataExportWorkflow(ctx context.Context, req *dmsV1.RejectDataExportWorkflowReq, userId string) error {
	return errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) CancelDataExportWorkflow(ctx context.Context, userId string, req *dmsV1.CancelDataExportWorkflowReq) error {
	return errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) ExportDataExportWorkflow(ctx context.Context, projectUid, dataExportWorkflowUid, currentUserUid string) error {
	return errNotDataExportWorkflow
}

func (d *DataExportWorkflowUsecase) DownloadDataExportTask(ctx context.Context, userId string, req *dmsV1.DownloadDataExportTaskReq) (string, error) {
	return "", errNotDataExportTask
}

func (d *DataExportWorkflowUsecase) AddDataExportTasks(ctx context.Context, projectUid, currentUserId string, params []*DataExportTask) (taskids []string, err error) {
	return nil, errNotDataExportTask
}

func (d *DataExportWorkflowUsecase) BatchGetDataExportTask(ctx context.Context, taskUids []string, currentUserId string) ([]*DataExportTask, error) {
	return nil, errNotDataExportTask
}

func (d *DataExportWorkflowUsecase) ListDataExportTaskRecords(ctx context.Context, options *ListDataExportTaskRecordOption, currentUserId string) ([]*DataExportTaskRecord, int64, error) {
	return nil, 0, errNotDataExportTask
}
