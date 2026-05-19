package v1

import (
	"github.com/actiontech/dms/internal/data_masking/biz"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters ListUnmaskingWorkflows
type ListUnmaskingWorkflowsReq struct {
	// project id
	// Required: true
	// in: path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// the maximum count of workflows to be returned
	// in: query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of workflows to be returned, default is 0
	// in: query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// filter the approval status
	// in: query
	FilterByApprovalStatus biz.UnmaskingWorkflowApprovalStatus `query:"filter_by_approval_status" json:"filter_by_approval_status"`
	// filter the usage status
	// in: query
	FilterByUsageStatus biz.UnmaskingWorkflowUsageStatus `query:"filter_by_usage_status" json:"filter_by_usage_status"`
	// filter db_service id
	// in: query
	FilterByDBServiceUid string `query:"filter_by_db_service_uid" json:"filter_by_db_service_uid"`
}

// swagger:model ListUnmaskingWorkflowsReply
type ListUnmaskingWorkflowsReply struct {
	Data  []*UnmaskingWorkflowListItem `json:"data"`
	Total int64                        `json:"total_nums"`
	// Generic reply
	base.GenericResp
}

// swagger:model UnmaskingWorkflowListItem
type UnmaskingWorkflowListItem struct {
	// 申请编号
	WorkflowID string `json:"workflow_id"`
	// 申请人用户名
	ApplicantName string `json:"applicant_name"`
	// 申请时间 (RFC3339)
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// 数据源实例名称
	DatasourceName string `json:"datasource_name"`
	// 数据源实例ID
	DatasourceUid string `json:"datasource_uid"`
	// 来源类型
	SourceType biz.UnmaskingWorkflowSourceType `json:"source_type" validate:"oneof=data_export sql_workbench"`
	// 来源对象UID
	SourceUID string `json:"source_uid"`
	// 审批状态
	ApprovalStatus biz.UnmaskingWorkflowApprovalStatus `json:"approval_status" validate:"oneof=pending approved rejected cancelled"`
	// 使用情况
	UsageStatus biz.UnmaskingWorkflowUsageStatus `json:"usage_status" validate:"oneof=unviewed viewed"`
	// 过期时间 (RFC3339)，兼容字段：批准前为激活截止，激活后为查看截止
	ExpireTime string `json:"expire_time" example:"2024-01-16T10:30:00Z"`
	// 激活查看时刻 (RFC3339)
	ActivatedAt string `json:"activated_at"`
	// 须在此时刻前激活查看 (RFC3339)
	ActivationDeadline string `json:"activation_deadline"`
	// 明文查看截止 (RFC3339)
	ViewValidUntil string `json:"view_valid_until"`
	// 申请人明文查看状态
	ViewState biz.UnmaskingWorkflowViewState `json:"view_state"`
	// 是否可点击激活查看
	CanActivate bool `json:"can_activate"`
	// 申请理由
	ApplyReason string `json:"apply_reason"`
	// 当前待处理人
	CurrentAssignees []*UidWithName `json:"current_assignees"`
}

// swagger:model CreateUnmaskingWorkflowReq
type CreateUnmaskingWorkflowReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: body
	// Required: true
	UnmaskingWorkflow *CreateUnmaskingWorkflow `json:"unmasking_workflow" validate:"required"`
}

// swagger:model CreateUnmaskingWorkflow
type CreateUnmaskingWorkflow struct {
	// 数据源 UID
	DatasourceUID string `json:"datasource_uid" validate:"required"`
	// SQL 默认 schema
	DefaultSchema string `json:"default_schema" validate:"required"`
	// 来源类型
	SourceType biz.UnmaskingWorkflowSourceType `json:"source_type" validate:"required,oneof=data_export sql_workbench"`
	// 来源对象 UID (如数据导出任务 UID)
	SourceUID string `json:"source_uid"`
	// 申请理由
	ApplyReason string `json:"apply_reason" validate:"required"`
	// 待脱敏 SQL 列表
	UnmaskingSQLs []CreateUnmaskingSQLItem `json:"unmasking_sqls" validate:"required,gt=0"`
}

// swagger:model CreateUnmaskingSQLItem
type CreateUnmaskingSQLItem struct {
	// 来源侧 SQL 索引 id（如数据导出记录中的 SQL 序号）；与 SQL 工作台场景的索引约定可能不同，需结合 source_type、source_uid 解析
	SQLIndexID string `json:"sql_index_id" validate:"required"`
	// 原始 SQL 内容
	SQLContent string `json:"sql_content" validate:"required"`
}

// swagger:model CreateUnmaskingWorkflowReply
type CreateUnmaskingWorkflowReply struct {
	Data *CreateUnmaskingWorkflowReplyData `json:"data"`
	// Generic reply
	base.GenericResp
}

// swagger:model CreateUnmaskingWorkflowReplyData
type CreateUnmaskingWorkflowReplyData struct {
	WorkflowID string `json:"workflow_id"`
}

// swagger:parameters GetUnmaskingWorkflow
type GetUnmaskingWorkflowReq struct {
	// project id
	// Required: true
	// in: path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
}

// swagger:model GetUnmaskingWorkflowReply
type GetUnmaskingWorkflowReply struct {
	Data *UnmaskingWorkflowDetail `json:"data"`
	// Generic reply
	base.GenericResp
}

// swagger:model UnmaskingWorkflowDetail
type UnmaskingWorkflowDetail struct {
	UnmaskingWorkflowListItem
	// 驳回理由 (整单驳回时)
	RejectReason string `json:"reject_reason"`
	// 当前待处理人
	CurrentAssignees []*UidWithName `json:"current_assignees"`
	// SQL 详情列表
	UnmaskingSQLs []*UnmaskingSQLDetail `json:"unmasking_sqls"`
	// 操作日志
	OperationLogs []*UnmaskingOperationLogItem `json:"operation_logs"`
}

// swagger:model UnmaskingSQLDetail
type UnmaskingSQLDetail struct {
	// SQL 详情 UID
	UID string `json:"uid"`
	// 来源侧 SQL 索引 id
	SQLIndexID string `json:"sql_index_id"`
	// 原始 SQL 内容
	SQLContent string `json:"sql_content"`
	// 脱敏配置快照
	MaskingConfigSnapshot []*biz.ColumnMaskingConfig `json:"masking_config_snapshot,omitempty"`
	// 血缘分析快照
	LineageAnalysisSnapshot *biz.AnalyzeResult `json:"lineage_analysis_snapshot,omitempty"`

	// 脱敏后的预览数据 (普通用户仅能看到此数据)
	MaskedData *SQLQueryResult `json:"masked_data"`
	// 原始采样数据 (仅有权限的审核人能看到)
	OriginalData *SQLQueryResult `json:"original_data"`
}

// swagger:model SQLQueryResultRow
// SQLQueryResultRow 一行数据；每个元素为单元格的字符串形式（与 columns 顺序一致）。
type SQLQueryResultRow []string

// swagger:model SQLQueryResult
type SQLQueryResult struct {
	// 列名列表
	Columns []string `json:"columns"`
	// 数据行列表 (每一行的数据顺序与 Columns 一致)
	Rows []SQLQueryResultRow `json:"rows"`
	// 总行数
	RowCount int `json:"row_count"`
}

// swagger:model UnmaskingOperationLogItem
type UnmaskingOperationLogItem struct {
	// 操作人 UID
	OperatorUID string `json:"operator_uid"`
	// 操作人姓名
	OperatorName string `json:"operator_name"`
	// 操作动作
	Action biz.UnmaskingAction `json:"action"`
	// 操作时间 (RFC3339)
	ActionTime string `json:"action_time"`
	// 额外信息 (如拦截原因)
	ExtraMessage string `json:"extra_message"`
}

type ApproveUnmaskingWorkflowReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	// swagger:ignore
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
	// in: body
	ApproveUnmaskingWorkflow *ApproveUnmaskingWorkflow `json:"approve_unmasking_workflow,omitempty"`
}

// swagger:model ApproveUnmaskingWorkflow
type ApproveUnmaskingWorkflow struct {
	// 审批理由 非必须
	ApproveReason string `json:"approve_reason"`
}

// swagger:model ApproveUnmaskingWorkflowReply
type ApproveUnmaskingWorkflowReply struct {
	// Generic reply
	base.GenericResp
}

type RejectUnmaskingWorkflowReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	// swagger:ignore
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
	// in: body
	// Required: true
	RejectUnmaskingWorkflow *RejectUnmaskingWorkflow `json:"reject_unmasking_workflow" validate:"required"`
}

// swagger:model RejectUnmaskingWorkflow
type RejectUnmaskingWorkflow struct {
	// 驳回理由
	// Required: true
	RejectReason string `json:"reject_reason" validate:"required"`
}

// swagger:model RejectUnmaskingWorkflowReply
type RejectUnmaskingWorkflowReply struct {
	// Generic reply
	base.GenericResp
}

// swagger:parameters CancelUnmaskingWorkflow
type CancelUnmaskingWorkflowReq struct {
	// project id
	// Required: true
	// in: path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
}

// swagger:model CancelUnmaskingWorkflowReply
type CancelUnmaskingWorkflowReply struct {
	// Generic reply
	base.GenericResp
}

// swagger:parameters ActivateUnmaskingWorkflowView
type ActivateUnmaskingWorkflowViewReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
}

// swagger:model ActivateUnmaskingWorkflowViewReply
type ActivateUnmaskingWorkflowViewReply struct {
	Data *ActivateUnmaskingWorkflowViewReplyData `json:"data"`
	base.GenericResp
}

// swagger:model ActivateUnmaskingWorkflowViewReplyData
type ActivateUnmaskingWorkflowViewReplyData struct {
	ViewValidUntil string                       `json:"view_valid_until"`
	ViewState      biz.UnmaskingWorkflowViewState `json:"view_state"`
}

// swagger:parameters GetUnmaskingWorkflowPlaintext
type GetUnmaskingWorkflowPlaintextReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
}

// swagger:model GetUnmaskingWorkflowPlaintextReply
type GetUnmaskingWorkflowPlaintextReply struct {
	Data *GetUnmaskingWorkflowPlaintext `json:"data"`
	base.GenericResp
}

// swagger:model GetUnmaskingWorkflowPlaintext
type GetUnmaskingWorkflowPlaintext struct {
	ViewState      biz.UnmaskingWorkflowViewState `json:"view_state"`
	ViewValidUntil string                         `json:"view_valid_until"`
	UnmaskingSQLs  []*UnmaskingPlaintextSQLItem   `json:"unmasking_sqls"`
}

// swagger:model UnmaskingPlaintextSQLItem
type UnmaskingPlaintextSQLItem struct {
	UID           string          `json:"uid"`
	SQLIndexID    string          `json:"sql_index_id"`
	OriginalData  *SQLQueryResult `json:"original_data"`
	MaskedColumns []string        `json:"masked_columns"`
	Truncated     bool            `json:"truncated"`
}
