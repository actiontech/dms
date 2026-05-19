package v1

import (
	"time"

	"github.com/actiontech/dms/internal/data_masking/biz"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

// swagger:model
type AddDataExportWorkflowReq struct {
	// swagger:ignore
	ProjectUid         string             `param:"project_uid" json:"project_uid" validate:"required"`
	DataExportWorkflow DataExportWorkflow `json:"data_export_workflow"`
}

type DataExportWorkflow struct {
	// name
	// Required: true
	// example: d1
	Name string `json:"name" validate:"required"`
	// desc
	// Required: false
	// example: transaction data export
	Desc string `json:"desc"`
	// export task info
	// Required: true
	// example: [export_task_uid1,export_task_uid2]
	Tasks []Task `json:"tasks" validate:"required"`
}

// swagger:model AddDataExportWorkflowReply
type AddDataExportWorkflowReply struct {
	// add data export workflow reply
	Data struct {
		// data export workflow UID
		Uid string `json:"export_data_workflow_uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ListDataExportWorkflows ListAllDataExportWorkflows
type ListDataExportWorkflowsReq struct {
	// project id
	// Required: false
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid"`
	// the maximum count of member to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of members to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// filter the status
	// in:query
	FilterByStatus DataExportWorkflowStatus `query:"filter_by_status" json:"filter_by_status"`
	// filter create user id
	// in:query
	FilterByCreateUserUid string `json:"filter_by_create_user_uid" query:"filter_by_create_user_uid"`
	// filter current assignee user id
	// in:query
	FilterCurrentStepAssigneeUserUid string `json:"filter_current_step_assignee_user_uid" query:"filter_current_step_assignee_user_uid"`
	// filter db_service id
	// in:query
	FilterByDBServiceUid string `json:"filter_by_db_service_uid" query:"filter_by_db_service_uid"`
	// filter create time from
	// in:query
	FilterCreateTimeFrom string `json:"filter_create_time_from" query:"filter_create_time_from"`
	// filter create time end
	// in:query
	FilterCreateTimeTo string `json:"filter_create_time_to" query:"filter_create_time_to"`
	// filter fuzzy key word for id/name
	// in:query
	FuzzyKeyword string `json:"fuzzy_keyword" query:"fuzzy_keyword"`
}

// swagger:parameters FilterGlobalDataExportWorkflowReq
type FilterGlobalDataExportWorkflowReq struct {
	// filter status list
	// in:query
	FilterStatusList []DataExportWorkflowStatus `json:"filter_status_list" query:"filter_status_list"`
	// filter db service uid
	// in:query
	FilterDBServiceUid string `json:"filter_db_service_uid" query:"filter_db_service_uid"`
	// filter project priority
	// in:query
	FilterCurrentStepAssigneeUserId string `json:"filter_current_step_assignee_user_id" query:"filter_current_step_assignee_user_id"`
	// filter project uids
	// in:query
	FilterProjectUids []string `json:"filter_project_uids" query:"filter_project_uids"`
	// filter project uid
	// in:query
	FilterProjectUid string `json:"filter_project_uid" query:"filter_project_uid"`
	// the maximum count of member to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of members to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// filter the status
	// in:query
	FilterByStatus DataExportWorkflowStatus `query:"filter_by_status" json:"filter_by_status"`
	// filter create user id
	// in:query
	FilterByCreateUserUid string `json:"filter_by_create_user_uid" query:"filter_by_create_user_uid"`
	// filter current assignee user id
	// in:query
	FilterCurrentStepAssigneeUserUid string `json:"filter_current_step_assignee_user_uid" query:"filter_current_step_assignee_user_uid"`
	// filter db_service id
	// in:query
	FilterByDBServiceUid string `json:"filter_by_db_service_uid" query:"filter_by_db_service_uid"`
	// filter fuzzy key word for id/name
	// in:query
	FuzzyKeyword string `json:"fuzzy_keyword" query:"fuzzy_keyword"`
	// enable OR-based self-relevant filtering (creator OR assignee OR viewable db_service)
	// in:query
	CheckUserCanAccess bool `json:"check_user_can_access" query:"check_user_can_access"`
	// current user id for check_user_can_access filter
	// in:query
	CurrentUserID string `json:"current_user_id" query:"current_user_id"`
	// db_service UIDs with view_others_workflow permission for check_user_can_access
	// in:query
	ViewableDBServiceUids []string `json:"viewable_db_service_uids" query:"viewable_db_service_uids"`
	// filter update time from
	// in:query
	FilterUpdateTimeFrom string `json:"filter_update_time_from" query:"filter_update_time_from"`
	// filter update time to
	// in:query
	FilterUpdateTimeTo string `json:"filter_update_time_to" query:"filter_update_time_to"`
}

// swagger:model GetGlobalDataExportWorkflowsReply
type GetGlobalDataExportWorkflowsReply struct {
	Data  []*GlobalDataExportWorkflow `json:"data"`
	Total int64                       `json:"total_nums"`
	// Generic reply
	base.GenericResp
}

type GlobalDataExportWorkflow struct {
	WorkflowID   string                   `json:"workflow_uid"`  // 数据导出工单ID
	WorkflowName string                   `json:"workflow_name"` // 数据导出工单的名称
	Description  string                   `json:"desc"`          // 数据导出工单的描述
	Creater      UidWithName              `json:"creater"`       // 数据导出工单的创建人
	CreatedAt    time.Time                `json:"created_at"`    // 数据导出工单的创建时间
	UpdatedAt    time.Time                `json:"updated_at"`    // 数据导出工单的更新时间
	Status       DataExportWorkflowStatus `json:"status"`        // 数据导出工单的状态

	CurrentStepAssigneeUsers []UidWithName                           `json:"current_step_assignee_user_list"` // 工单待操作人
	DBServiceInfos           []*dmsCommonV1.DBServiceUidWithNameInfo `json:"db_service_info,omitempty"`       // 所属数据源信息
	ProjectInfo              *dmsCommonV1.ProjectInfo                `json:"project_info,omitempty"`          // 所属项目信息
}

// swagger:model ListDataExportWorkflowsReply
type ListDataExportWorkflowsReply struct {
	Data  []*ListDataExportWorkflow `json:"data"`
	Total int64                     `json:"total_nums"`
	// Generic reply
	base.GenericResp
}

type ListDataExportWorkflow struct {
	ProjectUid   string                   `json:"project_uid"`
	ProjectName  string                   `json:"project_name"`  // 项目名称
	WorkflowID   string                   `json:"workflow_uid"`  // 数据导出工单ID
	WorkflowName string                   `json:"workflow_name"` // 数据导出工单的名称
	Description  string                   `json:"desc"`          // 数据导出工单的描述
	Creater      UidWithName              `json:"creater"`       // 数据导出工单的创建人
	CreatedAt    time.Time                `json:"created_at"`    // 数据导出工单的创建时间
	ExportedAt   time.Time                `json:"exported_at"`   // 执行数据导出工单的时间
	Status       DataExportWorkflowStatus `json:"status"`        // 数据导出工单的状态

	CurrentStepAssigneeUsers []UidWithName                           `json:"current_step_assignee_user_list"` // 工单待操作人
	DBServiceInfos           []*dmsCommonV1.DBServiceUidWithNameInfo `json:"db_service_info,omitempty"`       // 所属数据源信息
	ProjectInfo              *dmsCommonV1.ProjectInfo                `json:"project_info,omitempty"`          // 所属项目信息
}

// swagger:enum DataExportWorkflowStatus
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

// swagger:parameters GetDataExportWorkflow
type GetDataExportWorkflowReq struct {
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model GetDataExportWorkflowReply
type GetDataExportWorkflowReply struct {
	Data *GetDataExportWorkflow `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:enum WorkflowStepStatus
type WorkflowStepStatus string

const (
	WorkflowStepStatusWaitForExporting WorkflowStepStatus = "init"
	WorkflowStepStatusRejected         WorkflowStepStatus = "rejected"
	WorkflowStepStatusFinish           WorkflowStepStatus = "finish"
)

type GetDataExportWorkflow struct {
	Name                  string           `json:"workflow_name"`
	WorkflowID            string           `json:"workflow_uid"`
	Desc                  string           `json:"desc,omitempty"`
	CreateUser            UidWithName      `json:"create_user"`
	CreateTime            *time.Time       `json:"create_time"`
	WorkflowRecord        WorkflowRecord   `json:"workflow_record"`
	WorkflowRecordHistory []WorkflowRecord `json:"workflow_record_history"`
	// UnmaskingWorkflow 关联的查看原文工单摘要；无关联时为 null
	UnmaskingWorkflow *DataExportRelatedUnmaskingWorkflow `json:"unmasking_workflow"`
}

// swagger:model DataExportRelatedUnmaskingWorkflow
// DataExportRelatedUnmaskingWorkflow 数据导出工单关联的查看原文工单（仅摘要字段）
type DataExportRelatedUnmaskingWorkflow struct {
	// UnmaskingWorkflowUid 查看原文工单 UID，用于跳转详情等
	UnmaskingWorkflowUid string `json:"unmasking_workflow_uid"`
	// 创建人（uid + 展示名）
	Creator UidWithName `json:"creator"`
	// 创建时间 (RFC3339)
	CreatedAt string `json:"created_at"`
	// 过期时间 (RFC3339)，未设置时省略
	ExpireTime string `json:"expire_time,omitempty"`
	// 审批状态
	ApprovalStatus biz.UnmaskingWorkflowApprovalStatus `json:"approval_status"`
	// 使用状态
	UsageStatus biz.UnmaskingWorkflowUsageStatus `json:"usage_status"`
	// 申请理由
	ApplyReason string `json:"apply_reason"`
	// 驳回理由（整单驳回时）
	RejectReason string `json:"reject_reason,omitempty"`
	// 操作记录
	OperationLogs []*UnmaskingOperationLogItem `json:"operation_logs"`
}

type WorkflowRecord struct {
	Tasks             []*Task                  `json:"tasks"`
	CurrentStepNumber uint                     `json:"current_step_number,omitempty"`
	Status            DataExportWorkflowStatus `json:"status"`
	Steps             []*WorkflowStep          `json:"workflow_step_list,omitempty"`
}

type Task struct {
	Uid string `json:"task_uid"`
}

type WorkflowStep struct {
	Number        uint64             `json:"number"`
	Type          string             `json:"type"`
	Desc          string             `json:"desc,omitempty"`
	Users         []UidWithName      `json:"assignee_user_list,omitempty"`
	OperationUser UidWithName        `json:"operation_user,omitempty"`
	OperationTime *time.Time         `json:"operation_time,omitempty"`
	State         WorkflowStepStatus `json:"state,omitempty" `
	Reason        string             `json:"reason,omitempty"`
}

// swagger:parameters ApproveDataExportWorkflow
type ApproveDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
}

// swagger:parameters ExportDataExportWorkflow
type ExportDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
}

// swagger:parameters DownloadOriginalDataExportWorkflow
type DownloadOriginalDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
	// 已批准的查看原文工单 UID
	// Required: true
	// in:query
	UnmaskingWorkflowUid string `query:"unmasking_workflow_uid" json:"unmasking_workflow_uid" validate:"required"`
}

// swagger:response DownloadOriginalDataExportWorkflowReply
type DownloadOriginalDataExportWorkflowReply struct {
	// swagger:file
	// in: body
	File []byte
}

type RejectDataExportWorkflowPayload struct {
	// Required: true
	Reason string `json:"reason" validate:"required"`
}

// swagger:model
type RejectDataExportWorkflowReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// swagger:ignore
	DataExportWorkflowUid string                          `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
	Payload               RejectDataExportWorkflowPayload `json:"payload" validate:"required"`
}

type CancelDataExportWorkflowPayload struct {
	// Required: true
	DataExportWorkflowUids []string `json:"data_export_workflow_uids" validate:"required"`
}

// swagger:model
type CancelDataExportWorkflowReq struct {
	// swagger:ignore
	ProjectUid string                          `param:"project_uid" json:"project_uid" validate:"required"`
	Payload    CancelDataExportWorkflowPayload `json:"payload" validate:"required"`
}
