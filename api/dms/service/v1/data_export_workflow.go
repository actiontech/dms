package v1

import (
	"time"

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
