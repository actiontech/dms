package v1

import (
	privilegeApplyBiz "github.com/actiontech/dms/internal/privilege_apply/biz"
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters GetPrivilegeApplyAssignees
type GetPrivilegeApplyAssigneesReq struct {
	// project id
	// Required: true
	// in: path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// datasource uid
	// Required: true
	// in: query
	DBServiceUid string `query:"db_service_uid" json:"db_service_uid" validate:"required"`
}

// swagger:model GetPrivilegeApplyAssigneesReply
type GetPrivilegeApplyAssigneesReply struct {
	Data *GetPrivilegeApplyAssigneesReplyData `json:"data"`
	base.GenericResp
}

// swagger:model GetPrivilegeApplyAssigneesReplyData
type GetPrivilegeApplyAssigneesReplyData struct {
	HasAssignee bool           `json:"has_assignee"`
	Assignees   []*UidWithName `json:"assignees"`
}

// swagger:model CreatePrivilegeApplyWorkflowReq
type CreatePrivilegeApplyWorkflowReq struct {
	// swagger:ignore
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: body
	// Required: true
	PrivilegeApplyWorkflow *CreatePrivilegeApplyWorkflow `json:"privilege_apply_workflow" validate:"required"`
}

// swagger:model CreatePrivilegeApplyWorkflow
type CreatePrivilegeApplyWorkflow struct {
	DBServiceUid         string                            `json:"db_service_uid" validate:"required"`
	SourceDBAccountUid   string                            `json:"source_db_account_uid" validate:"required"`
	RawSQL               string                            `json:"raw_sql" validate:"required"`
	ErrorMessage         string                            `json:"error_message" validate:"required"`
	RequestedObjects     []privilegeApplyBiz.PrivilegeObject `json:"requested_objects"`
	RequestedActions     []string                          `json:"requested_actions"`
	ApplyReason          string                            `json:"apply_reason" validate:"required"`
	ExpectedExpireDays   *int64                            `json:"expected_expire_days"`
}

// swagger:model CreatePrivilegeApplyWorkflowReply
type CreatePrivilegeApplyWorkflowReply struct {
	Data *CreatePrivilegeApplyWorkflowReplyData `json:"data"`
	base.GenericResp
}

// swagger:model CreatePrivilegeApplyWorkflowReplyData
type CreatePrivilegeApplyWorkflowReplyData struct {
	WorkflowID string `json:"workflow_id"`
}

// swagger:parameters GetPrivilegeApplyWorkflow
type GetPrivilegeApplyWorkflowReq struct {
	// project id
	// Required: true
	// in: path
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// in: path
	// Required: true
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
}

// swagger:model GetPrivilegeApplyWorkflowReply
type GetPrivilegeApplyWorkflowReply struct {
	Data *PrivilegeApplyWorkflowDetail `json:"data"`
	base.GenericResp
}

// swagger:model PrivilegeApplyWorkflowDetail
type PrivilegeApplyWorkflowDetail struct {
	WorkflowID           string                            `json:"workflow_id"`
	ApprovalStatus       privilegeApplyBiz.PrivilegeApplyWorkflowApprovalStatus `json:"approval_status"`
	ReissueStatus        privilegeApplyBiz.PrivilegeApplyReissueStatus `json:"reissue_status"`
	ApplyReason          string                            `json:"apply_reason"`
	DBServiceUid         string                            `json:"db_service_uid"`
	DBServiceName        string                            `json:"db_service_name"`
	SourceDBAccountUid   string                            `json:"source_db_account_uid"`
	SourceDBAccountName  string                            `json:"source_db_account_name"`
	ApplicantUid         string                            `json:"applicant_uid"`
	ApplicantName        string                            `json:"applicant_name"`
	RawSQL               string                            `json:"raw_sql"`
	ErrorMessage         string                            `json:"error_message"`
	RequestedObjects     []privilegeApplyBiz.PrivilegeObject `json:"requested_objects"`
	RequestedActions     []string                          `json:"requested_actions"`
	ApprovedObjects      []privilegeApplyBiz.PrivilegeObject `json:"approved_objects"`
	ApprovedActions      []string                          `json:"approved_actions"`
	ExpectedExpireDays   *int64                            `json:"expected_expire_days"`
	RejectReason         string                            `json:"reject_reason"`
	ReissueError         string                            `json:"reissue_error"`
	TargetDBAccountUid   string                            `json:"target_db_account_uid"`
	TargetDBAccountName  string                            `json:"target_db_account_name"`
	CreatedAt            string                            `json:"created_at" example:"2024-01-15T10:30:00Z"`
	CurrentAssignees     []*UidWithName                    `json:"current_assignees"`
	ImpactPreview        *privilegeApplyBiz.PrivilegeApplyImpactPreview `json:"impact_preview"`
}

// swagger:parameters ListPrivilegeApplyWorkflows
type ListPrivilegeApplyWorkflowsReq struct {
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	PageSize   uint32 `query:"page_size" json:"page_size" validate:"required"`
	PageIndex  uint32 `query:"page_index" json:"page_index"`
	FilterByTab string `query:"filter_by_tab" json:"filter_by_tab" validate:"required,oneof=pending handled"`
	FilterByDBServiceUid string `query:"filter_by_db_service_uid" json:"filter_by_db_service_uid"`
}

// swagger:model ListPrivilegeApplyWorkflowsReply
type ListPrivilegeApplyWorkflowsReply struct {
	Data  []*PrivilegeApplyWorkflowListItem `json:"data"`
	Total int64                             `json:"total_nums"`
	base.GenericResp
}

// swagger:model PrivilegeApplyWorkflowListItem
type PrivilegeApplyWorkflowListItem struct {
	WorkflowID          string `json:"workflow_id"`
	ApplicantUid        string `json:"applicant_uid"`
	ApplicantName       string `json:"applicant_name"`
	DBServiceUid        string `json:"db_service_uid"`
	DBServiceName       string `json:"db_service_name"`
	SourceDBAccountName string `json:"source_db_account_name"`
	ApplyReason         string `json:"apply_reason"`
	CreatedAt           string `json:"created_at"`
	ApprovalStatus      privilegeApplyBiz.PrivilegeApplyWorkflowApprovalStatus `json:"approval_status"`
	ReissueStatus       privilegeApplyBiz.PrivilegeApplyReissueStatus `json:"reissue_status"`
	CurrentAssignees    []*UidWithName `json:"current_assignees"`
}

// swagger:parameters ApprovePrivilegeApplyWorkflow
type ApprovePrivilegeApplyWorkflowReq struct {
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
	ApprovePrivilegeApplyWorkflow *ApprovePrivilegeApplyWorkflow `json:"approve_privilege_apply_workflow"`
}

type ApprovePrivilegeApplyWorkflow struct {
	ApproveReason       string `json:"approve_reason"`
	ApprovedPermissions *privilegeApplyBiz.ApprovedPermissions `json:"approved_permissions"`
	ExpectedExpireDays  *int64 `json:"expected_expire_days"`
}

type ApprovePrivilegeApplyWorkflowReply struct {
	base.GenericResp
}

// swagger:parameters RejectPrivilegeApplyWorkflow
type RejectPrivilegeApplyWorkflowReq struct {
	ProjectUid   string `param:"project_uid" json:"project_uid" validate:"required"`
	WorkflowID   string `param:"workflow_id" json:"workflow_id" validate:"required"`
	RejectReason string `json:"reject_reason" validate:"required"`
}

type RejectPrivilegeApplyWorkflowReply struct {
	base.GenericResp
}

// swagger:parameters RetryReissuePrivilegeApplyWorkflow
type RetryReissuePrivilegeApplyWorkflowReq struct {
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	WorkflowID string `param:"workflow_id" json:"workflow_id" validate:"required"`
}

type RetryReissuePrivilegeApplyWorkflowReply struct {
	base.GenericResp
}
