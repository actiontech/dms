package v1

import (
	"time"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters AddDataExportWorkflow
type AddDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// add data export workflow
	// in:body
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
	// example: export tasks
	TaskIds []string `json:"task_ids" validate:"required"`
}

// swagger:model AddDataExportWorkflowReply
type AddDataExportWorkflowReply struct {
	// add data export workflow reply
	Data struct {
		// data export workflow UID
		Uid string `json:"uid"`
	} `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters ListDataExportWorkflows
type ListDataExportWorkflowsReq struct {
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
	// filter the status
	// in:query
	FilterByStatus string `query:"filter_by_status" json:"filter_by_status"`
	// filter fuzzy key word
	// in:query
	FuzzyKeyword string `json:"fuzzy_keyword" query:"fuzzy_keyword"`
	// filter workflow subject
	// in:query
	FilterBySubject string `json:"filter_subject" query:"filter_subject"`
	// filter workflow id
	// in:query
	FilterByWorkflowID string `json:"filter_workflow_id" query:"filter_workflow_id"`
	// filter create user id
	// in:query
	FilterByCreateUserId string `json:"filter_create_user_id" query:"filter_create_user_id"`
}

// swagger:model ListDataExportWorkflowsReply
type ListDataExportWorkflowsReply struct {
	Data  []*ListDataExportWorkflow `json:"data"`
	Total int64                     `json:"total_nums"`
	// Generic reply
	base.GenericResp
}

type ListDataExportWorkflow struct {
	UID        string `json:"uid"`
	ProjectUid string `json:"project_uid"`

	WorkflowID               string        `json:"workflow_id"`   // 数据导出工单ID
	WorkflowName             string        `json:"workflow_name"` // 数据导出工单的名称
	Description              string        `json:"desc"`          // 数据导出工单的描述
	Creater                  UidWithName   `json:"creater"`       // 数据导出工单的创建人
	CreatedAt                time.Time     `json:"createdAt"`     // 数据导出工单的创建时间
	ExportedAt               time.Time     `json:"exportedAt"`    // 执行数据导出工单的时间
	CurrentStepType          string        `json:"current_step_type"`
	CurrentStepAssigneeUsers []UidWithName `json:"current_step_assignee_user_name_list"`
	Status                   string        `json:"status"` // 数据导出工单的状态
}

// 工单的状态常量
const (
	StatusPending      string = "待审核"
	StatusApproved     string = "审核通过"
	StatusExporting    string = "正在执行"
	StatusExportFailed string = "执行失败"
	StatusExported     string = "执行成功"
	StatusRejected     string = "已驳回"
	StatusClosed       string = "已关闭"
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

type GetDataExportWorkflow struct {
	UID                   string           `json:"uid"`
	Name                  string           `json:"workflow_name"`
	WorkflowID            string           `json:"workflow_id"`
	Desc                  string           `json:"desc,omitempty"`
	CreateUser            string           `json:"create_user_name"`
	CreateTime            *time.Time       `json:"create_time"`
	ProjectUid            string           `json:"project_uid"`
	WorkflowRecord        WorkflowRecord   `json:"workflow_record"`
	WorkflowRecordHistory []WorkflowRecord `json:"workflow_record_history"`
}

type WorkflowRecord struct {
	Tasks             []*Task         `json:"tasks"`
	CurrentStepNumber uint            `json:"current_step_number,omitempty"`
	Status            string          `json:"status" enums:"wait_for_audit,wait_for_execution,rejected,canceled,exec_failed,executing,finished"`
	Steps             []*WorkflowStep `json:"workflow_step_list,omitempty"`
}
type Task struct {
	Id string `json:"task_id"`
}

type WorkflowStep struct {
	Id            uint       `json:"workflow_step_id,omitempty"`
	Number        uint       `json:"number"`
	Type          string     `json:"type"`
	Desc          string     `json:"desc,omitempty"`
	Users         []string   `json:"assignee_user_name_list,omitempty"`
	OperationUser string     `json:"operation_user_name,omitempty"`
	OperationTime *time.Time `json:"operation_time,omitempty"`
	State         string     `json:"state,omitempty" enums:"initialized,approved,rejected"`
	Reason        string     `json:"reason,omitempty"`
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

// swagger:parameters ExecDataExportWorkflow
type ExecDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
}

// swagger:parameters RejectDataExportWorkflow
type RejectDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid" json:"data_export_workflow_uid" validate:"required"`
}

// swagger:parameters UpdateDataExportWorkflow
type UpdateDataExportWorkflowReq struct {
	// project id
	// Required: true
	// in:path
	ProjectUid string `param:"project_uid" validate:"required"`
	// Required: true
	// in:path
	DataExportWorkflowUid string `param:"data_export_workflow_uid"  validate:"required"`
	// desc
	// Required: false
	// example: transaction data export
	Desc string `json:"desc"`
}
