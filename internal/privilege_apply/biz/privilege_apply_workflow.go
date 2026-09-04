package biz

import (
	"context"
	"errors"
	"time"
)

// PrivilegeApplyWorkflowApprovalStatus 审批状态
// swagger:enum PrivilegeApplyWorkflowApprovalStatus
type PrivilegeApplyWorkflowApprovalStatus string

const (
	PrivilegeApplyWorkflowApprovalStatusPending  PrivilegeApplyWorkflowApprovalStatus = "pending"
	PrivilegeApplyWorkflowApprovalStatusApproved PrivilegeApplyWorkflowApprovalStatus = "approved"
	PrivilegeApplyWorkflowApprovalStatusRejected PrivilegeApplyWorkflowApprovalStatus = "rejected"
	PrivilegeApplyWorkflowApprovalStatusCancelled PrivilegeApplyWorkflowApprovalStatus = "cancelled"
)

func (s PrivilegeApplyWorkflowApprovalStatus) String() string {
	return string(s)
}

// PrivilegeApplyReissueStatus 换发状态
// swagger:enum PrivilegeApplyReissueStatus
type PrivilegeApplyReissueStatus string

const (
	PrivilegeApplyReissueStatusNone      PrivilegeApplyReissueStatus = "none"
	PrivilegeApplyReissueStatusRunning   PrivilegeApplyReissueStatus = "running"
	PrivilegeApplyReissueStatusSucceeded PrivilegeApplyReissueStatus = "succeeded"
	PrivilegeApplyReissueStatusFailed    PrivilegeApplyReissueStatus = "failed"
)

func (s PrivilegeApplyReissueStatus) String() string {
	return string(s)
}

// PrivilegeObject 所需权限对象
// swagger:model PrivilegeObject
type PrivilegeObject struct {
	Schema     string `json:"schema"`
	ObjectName string `json:"object_name"`
	ObjectType string `json:"object_type"`
}

// PrivilegeApplyAction 操作动作
// swagger:enum PrivilegeApplyAction
type PrivilegeApplyAction string

const (
	PrivilegeApplyActionSubmit  PrivilegeApplyAction = "submit"
	PrivilegeApplyActionApprove PrivilegeApplyAction = "approve"
	PrivilegeApplyActionReject  PrivilegeApplyAction = "reject"
	PrivilegeApplyActionRetry   PrivilegeApplyAction = "retry_reissue"
)

func (a PrivilegeApplyAction) String() string {
	return string(a)
}

var (
	ErrNoAssignee       = errors.New("privilege_apply.no_assignee")
	ErrAccountMismatch  = errors.New("privilege_apply.account_mismatch")
	ErrInvalidArgument  = errors.New("privilege_apply.invalid_argument")
	ErrForbidden        = errors.New("privilege_apply.forbidden")
	ErrNotPending       = errors.New("privilege_apply.not_pending")
	ErrReissueFailed    = errors.New("privilege_apply.reissue_failed")
	ErrEmptyPermissions = errors.New("privilege_apply.empty_permissions")
)

// AssigneeResolveMode 审批人解析模式
type AssigneeResolveMode int

const (
	AssigneeResolveModePreview AssigneeResolveMode = iota
	AssigneeResolveModeSubmit
)

// PrivilegeApplyImpactPreview 落地影响预览
type PrivilegeApplyImpactPreview struct {
	WillCreateAccount              bool   `json:"will_create_account"`
	SourceAccountUID               string `json:"source_account_uid"`
	SourceAccountName              string `json:"source_account_name"`
	PermissionUnionSummary         string `json:"permission_union_summary"`
	UnbindSourceThenBindTarget     bool   `json:"unbind_source_then_bind_target"`
	OtherUsersOnSourceUnaffected   bool   `json:"other_users_on_source_unaffected"`
	SourceAccountGrantsUnchanged   bool   `json:"source_account_grants_unchanged"`
}

// ApprovedPermissions 审批人修正后的权限集合
type ApprovedPermissions struct {
	Objects []PrivilegeObject `json:"objects"`
	Actions []string          `json:"actions"`
}

// PrivilegeApplyWorkflow 提权申请工单
type PrivilegeApplyWorkflow struct {
	UID                   string
	ProjectUID            string
	ApplicantUID          string
	ApplicantName         string
	CreatedAt             time.Time
	DBServiceUID          string
	DBServiceName         string
	SourceDBAccountUID    string
	SourceDBAccountName   string
	RawSQL                string
	ErrorMessage          string
	RequestedObjects      []PrivilegeObject
	RequestedActions      []string
	ApprovedObjects       []PrivilegeObject
	ApprovedActions       []string
	ApplyReason           string
	ExpectedExpireDays    *int64
	ApprovalStatus        PrivilegeApplyWorkflowApprovalStatus
	ReissueStatus         PrivilegeApplyReissueStatus
	CurrentAssigneeUIDs   []string
	CurrentAssignees      []UIDWithName
	RejectReason          string
	ApproverUID           string
	ApprovedAt            *time.Time
	TargetDBAccountUID    string
	TargetDBAccountName   string
	ReissueError          string
	ImpactPreview         *PrivilegeApplyImpactPreview
	OperationLogs         []*PrivilegeApplyOperationLog
}

// UIDWithName 用户 UID 与展示名
type UIDWithName struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

// PrivilegeApplyOperationLog 操作日志
type PrivilegeApplyOperationLog struct {
	UID          string
	WorkflowUID  string
	OperatorUID  string
	OperatorName string
	ActionType   PrivilegeApplyAction
	ActionTime   time.Time
	ExtraMessage string
}

// ListPrivilegeApplyWorkflowsOption 列表查询选项
type ListPrivilegeApplyWorkflowsOption struct {
	PageSize              uint32
	PageIndex             uint32
	ProjectUID            string
	FilterByTab           string // pending | handled
	FilterByDBServiceUID  string
	ViewerUID             string
	// FilterByApplicantUID 由 usecase 按身份写入；普通成员仅见本人申请（AC-016）。禁止信任前端可见范围参。
	FilterByApplicantUID  string
	CanAuditDBServiceUIDs map[string]struct{}
}

// PrivilegeApplyWorkflowRepository 提权申请持久化
type PrivilegeApplyWorkflowRepository interface {
	CreateWorkflow(ctx context.Context, workflow *PrivilegeApplyWorkflow) error
	GetWorkflow(ctx context.Context, workflowUID string) (*PrivilegeApplyWorkflow, error)
	UpdateWorkflow(ctx context.Context, workflow *PrivilegeApplyWorkflow) error
	ListWorkflows(ctx context.Context, opt *ListPrivilegeApplyWorkflowsOption) ([]*PrivilegeApplyWorkflow, int64, error)
}

// PrivilegeApplyOperationLogRepository 操作日志持久化
type PrivilegeApplyOperationLogRepository interface {
	CreateOperationLog(ctx context.Context, log *PrivilegeApplyOperationLog) error
	ListOperationLogs(ctx context.Context, workflowUID string) ([]*PrivilegeApplyOperationLog, error)
}

// PrivilegeApplyOpPermissionVerifier 操作权限校验
type PrivilegeApplyOpPermissionVerifier interface {
	GetCanOpDBUsers(ctx context.Context, projectUID, dbServiceUID string, needOpPermissionTypes []string, isBusinessWrite bool) ([]string, error)
	UserCanAuditPrivilegeApply(ctx context.Context, projectUID, dbServiceUID, userUID string) (bool, error)
}

// PrivilegeApplyProvisionGateway provision 编排网关
type PrivilegeApplyProvisionGateway interface {
	RunReissueOrchestration(ctx context.Context, operatorUID string, workflow *PrivilegeApplyWorkflow, approved *ApprovedPermissions, effectiveDays int64) (targetUID, targetName string, err error)
	BuildImpactPreview(ctx context.Context, operatorUID string, workflow *PrivilegeApplyWorkflow, approved *ApprovedPermissions) (*PrivilegeApplyImpactPreview, error)
}

// PrivilegeApplyUserDirectory 用户展示名解析
type PrivilegeApplyUserDirectory interface {
	GetUserNamesByUIDs(ctx context.Context, uids []string) (map[string]string, error)
}

// PrivilegeApplyDBServiceDirectory 数据源展示名解析
type PrivilegeApplyDBServiceDirectory interface {
	GetDBServiceNamesByUIDs(ctx context.Context, uids []string) (map[string]string, error)
}

// PrivilegeApplyAccountVerifier 托管账号绑定校验
type PrivilegeApplyAccountVerifier interface {
	VerifyUserBoundAccount(ctx context.Context, projectUID, userUID, dbServiceUID, accountUID string) error
}

// ApprovePrivilegeApplyWorkflowArgs 批准参数
type ApprovePrivilegeApplyWorkflowArgs struct {
	ProjectUID          string
	WorkflowUID         string
	OperatorUID         string
	ApproveReason       string
	ApprovedPermissions *ApprovedPermissions
	ExpectedExpireDays  *int64
}

// RejectPrivilegeApplyWorkflowArgs 驳回参数
type RejectPrivilegeApplyWorkflowArgs struct {
	ProjectUID   string
	WorkflowUID  string
	OperatorUID  string
	RejectReason string
}

// CreatePrivilegeApplyWorkflowArgs 创建提权申请参数
type CreatePrivilegeApplyWorkflowArgs struct {
	ProjectUID           string
	ApplicantUID         string
	DBServiceUID         string
	SourceDBAccountUID   string
	SourceDBAccountName  string
	RawSQL               string
	ErrorMessage         string
	RequestedObjects     []PrivilegeObject
	RequestedActions     []string
	ApplyReason          string
	ExpectedExpireDays   *int64
}
