package biz

import (
	"context"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type DataExportWorkflowStatus string

const (
	DataExportWorkflowStatusWaitForApprove   DataExportWorkflowStatus = "wait_for_approve"
	DataExportWorkflowStatusWaitForExport    DataExportWorkflowStatus = "wait_for_export"
	DataExportWorkflowStatusWaitForExporting DataExportWorkflowStatus = "exporting"
	DataExportWorkflowStatusRejected         DataExportWorkflowStatus = "rejected"
	DataExportWorkflowStatusCancel           DataExportWorkflowStatus = "cancel"
	DataExportWorkflowStatusFailed           DataExportWorkflowStatus = "failed"
	DataExportWorkflowStatusFinish           DataExportWorkflowStatus = "finish"
)

func (dews DataExportWorkflowStatus) String() string {
	return string(dews)
}

type EventType string

const (
	DataExportWorkflowEventType EventType = "data_export"
)

func (et EventType) String() string {
	return string(et)
}

type DataExportWorkflowEventAction string

const (
	DataExportWorkflowEventActionCreate  DataExportWorkflowEventAction = "create"
	DataExportWorkflowEventActionApprove DataExportWorkflowEventAction = "approve"
	DataExportWorkflowEventActionReject  DataExportWorkflowEventAction = "reject"
	DataExportWorkflowEventActionCancel  DataExportWorkflowEventAction = "cancel"
	DataExportWorkflowEventActionExport  DataExportWorkflowEventAction = "export"
)

func (et DataExportWorkflowEventAction) String() string {
	return string(et)
}

type Workflow struct {
	Base

	UID               string
	Name              string
	ProjectUID        string
	WorkflowType      string
	Desc              string
	CreateTime        time.Time
	CreateUserUID     string
	Status            string
	WorkflowRecordUid string
	Tasks             []Task

	WorkflowRecord *WorkflowRecord
}

type Task struct {
	UID string
}

type WorkflowRecord struct {
	UID                   string
	Status                DataExportWorkflowStatus
	Tasks                 []Task
	CurrentWorkflowStepId uint64
	WorkflowSteps         []*WorkflowStep
}

type WorkflowStep struct {
	StepId            uint64
	WorkflowRecordUid string
	OperationUserUid  string
	OperateAt         *time.Time
	State             string
	Reason            string
	Assignees         []string
}

type WorkflowRepo interface {
	SaveWorkflow(ctx context.Context, dataExportWorkflow *Workflow) error
	ListDataExportWorkflows(ctx context.Context, opt *ListWorkflowsOption) ([]*Workflow, int64, error)
	GetDataExportWorkflow(ctx context.Context, dataExportWorkflowUid string) (*Workflow, error)
	UpdateWorkflowStatusById(ctx context.Context, dataExportWorkflowUid string, status DataExportWorkflowStatus) error
	GetDataExportWorkflowsByIds(ctx context.Context, dataExportWorkflowUid []string) ([]*Workflow, error)
	CancelWorkflow(ctx context.Context, workflowRecordIds []string, workflowSteps []*WorkflowStep, operateId string) error
	AuditWorkflow(ctx context.Context, dataExportWorkflowUid string, status DataExportWorkflowStatus, step *WorkflowStep, operateId, reason string) error
	GetDataExportWorkflowsForView(ctx context.Context, userUid string) ([]string, error)
	GetAllDataExportWorkflowsForView(ctx context.Context) ([]string, error)
	GetDataExportWorkflowsByStatus(ctx context.Context, status string) ([]string, error)
	GetDataExportWorkflowsByAssignUser(ctx context.Context, userUid string) ([]string, error)
	GetDataExportWorkflowsByDBService(ctx context.Context, dbUid string) ([]string, error)
	GetDataExportWorkflowsByDBServices(ctx context.Context, dbUid []string) ([]string, error)
	DeleteDataExportWorkflowsByIds(ctx context.Context, dataExportWorkflowUid []string) error
}

type DataExportWorkflowUsecase struct {
	tx                        TransactionGenerator
	repo                      WorkflowRepo
	dbServiceRepo             DBServiceRepo
	dataExportTaskRepo        DataExportTaskRepo
	dmsProxyTargetRepo        ProxyTargetRepo
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	projectUsecase            *ProjectUsecase
	clusterUsecase            *ClusterUsecase
	webhookUsecase            *WebHookConfigurationUsecase
	userUsecase               *UserUsecase
	log                       *utilLog.Helper
	reportHost                string
}

func NewDataExportWorkflowUsecase(logger utilLog.Logger, tx TransactionGenerator, repo WorkflowRepo, dataExportTaskRepo DataExportTaskRepo, dbServiceRepo DBServiceRepo, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, projectUsecase *ProjectUsecase, proxyTargetRepo ProxyTargetRepo, clusterUseCase *ClusterUsecase, webhookUsecase *WebHookConfigurationUsecase, userUsecase *UserUsecase, reportHost string) *DataExportWorkflowUsecase {
	return &DataExportWorkflowUsecase{
		tx:                        tx,
		repo:                      repo,
		dbServiceRepo:             dbServiceRepo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		projectUsecase:            projectUsecase,
		dmsProxyTargetRepo:        proxyTargetRepo,
		dataExportTaskRepo:        dataExportTaskRepo,
		clusterUsecase:            clusterUseCase,
		webhookUsecase:            webhookUsecase,
		userUsecase:               userUsecase,
		log:                       utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.dtaExportWorkflow")),
		reportHost:                reportHost,
	}
}

type ListWorkflowsOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      WorkflowField
	FilterBy     []pkgConst.FilterCondition
}
