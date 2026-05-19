package biz

import (
	"context"
	"time"

	maskingCore "github.com/actiontech/dms/internal/data_masking/core"
)

// MaskingConfigStatus 脱敏配置状态
// swagger:enum MaskingConfigStatus
type MaskingConfigStatus string

const (
	MaskingConfigStatusPendingConfirm  MaskingConfigStatus = "PENDING_CONFIRM"  // 系统发现，待人工确认
	MaskingConfigStatusConfigured      MaskingConfigStatus = "CONFIGURED"       // 用户已确认/手动配置
	MaskingConfigStatusSystemConfirmed MaskingConfigStatus = "SYSTEM_CONFIRMED" // 系统已确认
)

// ColumnMaskingConfig 列级别脱敏配置领域模型
// swagger:model ColumnMaskingConfig
type ColumnMaskingConfig struct {
	// 配置记录 ID
	ID uint `json:"id"`
	// 数据源 UID
	DBServiceUID string `json:"db_service_uid"`
	// 列 ID（db_columns.id）
	ColumnID uint `json:"column_id"`
	// Schema 名称
	SchemaName string `json:"schema_name"`
	// 表名
	TableName string `json:"table_name"`
	// 列名
	ColumnName string `json:"column_name"`
	// 是否启用脱敏
	IsMaskingEnabled bool `json:"is_masking_enabled"`
	// 脱敏规则 ID
	MaskingRuleID int `json:"masking_rule_id"`
	// 脱敏规则名称（中文）
	MaskingRuleName string `json:"masking_rule_name"`
	// 置信度
	Confidence maskingCore.Confidence `json:"confidence,omitempty"`
	// 配置状态
	Status MaskingConfigStatus `json:"status"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// swagger:model TableRef
type TableRef struct {
	// Schema 名称
	Schema string `json:"schema"`
	// 表名
	Table string `json:"table"`
	// 表别名
	Alias string `json:"alias"`
}

// swagger:model ColumnRef
type ColumnRef struct {
	// Schema 名称
	Schema string `json:"schema"`
	// 表名
	Table string `json:"table"`
	// 列名
	Column string `json:"column"`
}

// swagger:model ResultColumn
type ResultColumn struct {
	// 结果列名
	Name string `json:"name"`
	// 结果列表达式（SQL 片段）
	Expression string `json:"expression"`
	// 来源列列表
	Sources []ColumnRef `json:"sources"`
}

// swagger:model LineageNode
type LineageNode struct {
	// 节点 ID（图内唯一）
	ID string `json:"id"`
	// 节点类型
	Type NodeType `json:"type"`
	// 节点展示名
	Name string `json:"name"`
	// Schema 名称（列节点可能存在）
	Schema string `json:"schema"`
	// 表名（列节点可能存在）
	Table string `json:"table"`
	// 列名（列节点可能存在）
	Column string `json:"column"`
	// 表达式内容（表达式节点可能存在）
	Expr string `json:"expr"`
}

// swagger:model LineageEdge
type LineageEdge struct {
	// 起点节点 ID
	FromID string `json:"from_id"`
	// 终点节点 ID
	ToID string `json:"to_id"`
	// 边类型
	Type EdgeType `json:"type"`
}

// swagger:model AnalyzeResult
type AnalyzeResult struct {
	// 分析标题/摘要
	Title string `json:"title"`
	// 原始 SQL
	OriginalSQL string `json:"original_sql"`
	// 解析到的表引用列表
	Tables []TableRef `json:"tables"`
	// 解析到的源列列表
	SourceColumns []ColumnRef `json:"source_columns"`
	// 解析到的结果列列表
	ResultColumns []ResultColumn `json:"result_columns"`
	// 血缘图节点
	Nodes []LineageNode `json:"nodes"`
	// 血缘图边
	Edges []LineageEdge `json:"edges"`
	// 警告信息（解析不完整等）
	Warnings []string `json:"warnings,omitempty"`
}

// NodeType 血缘节点类型
// swagger:enum NodeType
type NodeType string

const (
	// 源列节点
	NodeTypeSource NodeType = "source_column"
	// 表达式节点
	NodeTypeExpression NodeType = "expression"
	// 结果列节点
	NodeTypeResult NodeType = "result_column"
	// 表节点
	NodeTypeTable NodeType = "table"
)

// EdgeType 血缘边类型
// swagger:enum EdgeType
type EdgeType string

const (
	// 直接依赖
	EdgeTypeDirect EdgeType = "direct"
	// 转换/计算依赖
	EdgeTypeTransform EdgeType = "transform"
	// 聚合依赖
	EdgeTypeAggregate EdgeType = "aggregate"
)

// UnmaskingWorkflowApprovalStatus 审批状态
// swagger:enum UnmaskingWorkflowApprovalStatus
type UnmaskingWorkflowApprovalStatus string

const (
	// 待审批
	UnmaskingWorkflowApprovalStatusPending UnmaskingWorkflowApprovalStatus = "pending"
	// 已批准
	UnmaskingWorkflowApprovalStatusApproved UnmaskingWorkflowApprovalStatus = "approved"
	// 已驳回
	UnmaskingWorkflowApprovalStatusRejected UnmaskingWorkflowApprovalStatus = "rejected"
	// 已取消
	UnmaskingWorkflowApprovalStatusCancelled UnmaskingWorkflowApprovalStatus = "cancelled"
)

func (s UnmaskingWorkflowApprovalStatus) String() string {
	return string(s)
}

// UnmaskingWorkflowUsageStatus 使用情况
// swagger:enum UnmaskingWorkflowUsageStatus
type UnmaskingWorkflowUsageStatus string

const (
	// 未查看
	UnmaskingWorkflowUsageStatusUnviewed UnmaskingWorkflowUsageStatus = "unviewed"
	// 已查看
	UnmaskingWorkflowUsageStatusViewed UnmaskingWorkflowUsageStatus = "viewed"
)

func (s UnmaskingWorkflowUsageStatus) String() string {
	return string(s)
}

// UnmaskingWorkflowSourceType 来源类型
// swagger:enum UnmaskingWorkflowSourceType
type UnmaskingWorkflowSourceType string

const (
	// 数据导出工单
	UnmaskingWorkflowSourceTypeDataExport UnmaskingWorkflowSourceType = "data_export"
	// SQL工作台
	UnmaskingWorkflowSourceTypeSQLWorkbench UnmaskingWorkflowSourceType = "sql_workbench"
)

func (s UnmaskingWorkflowSourceType) String() string {
	return string(s)
}

// UnmaskingOriginalDataCredentialArgs 工单凭证路径参数，供 CheckOriginalDataAccess 使用；nil 表示仅做权限直通（不入库解析工单）。enterprise 与 dms 构建共用。
type UnmaskingOriginalDataCredentialArgs struct {
	SourceType UnmaskingWorkflowSourceType
	SourceUID  string
	Credential string
}

// UnmaskingAction 操作动作类型
// swagger:enum UnmaskingAction
type UnmaskingAction string

const (
	// 提交申请
	UnmaskingActionSubmit UnmaskingAction = "submit"
	// 批准申请
	UnmaskingActionApprove UnmaskingAction = "approve"
	// 驳回申请
	UnmaskingActionReject UnmaskingAction = "reject"
	// 查看工单详情
	UnmaskingActionViewOriginalDataWorkflowDetail UnmaskingAction = "view_unmasking_workflow_detail"
	// 查看原文
	UnmaskingActionViewOriginalData UnmaskingAction = "view_full_original_data"
	// 下载原文
	UnmaskingActionDownloadOriginalData UnmaskingAction = "download_full_original_data"
	// 取消申请
	UnmaskingActionCancel UnmaskingAction = "cancel"
	// 激活查看原文（工单详情）
	UnmaskingActionActivateView UnmaskingAction = "activate_view"
)

func (a UnmaskingAction) String() string {
	return string(a)
}

// UnmaskingWorkflowViewState 申请人明文查看状态（详情页展示）
// swagger:enum UnmaskingWorkflowViewState
type UnmaskingWorkflowViewState string

const (
	UnmaskingWorkflowViewStateNotActivated      UnmaskingWorkflowViewState = "not_activated"
	UnmaskingWorkflowViewStateActive            UnmaskingWorkflowViewState = "active"
	UnmaskingWorkflowViewStateViewExpired       UnmaskingWorkflowViewState = "view_expired"
	UnmaskingWorkflowViewStateActivationExpired UnmaskingWorkflowViewState = "activation_expired"
)

func (s UnmaskingWorkflowViewState) String() string {
	return string(s)
}

// 与 DMS OpRangeType 字符串取值一致，用于 UnmaskingOpPermissionRange 判定。
const (
	UnmaskingOpRangeProject   = "project"
	UnmaskingOpRangeDBService = "db_service"
)

// UnmaskingOpPermissionRange 查看原文工单权限判定用的「操作权限 + 范围」快照。
type UnmaskingOpPermissionRange struct {
	OpPermissionUID string
	OpRangeType     string
	RangeUIDs       []string
}

// UnmaskingWorkflowOpPermissionVerifier 操作权限校验能力。
type UnmaskingWorkflowOpPermissionVerifier interface {
	IsUserDMSAdmin(ctx context.Context, userUID string) (bool, error)
	CanOpGlobal(ctx context.Context, userUID string, isBusinessWrite bool) (bool, error)
	CanViewGlobal(ctx context.Context, userUID string) (bool, error)
	IsUserProjectAdmin(ctx context.Context, userUID, projectUID string, isBusinessWrite bool) (bool, error)
	GetUserOpPermissionInProject(ctx context.Context, userUID, projectUID string) ([]UnmaskingOpPermissionRange, error)
	GetCanOpDBUsers(ctx context.Context, projectUID, dbServiceUID string, needOpPermissionTypes []string, isBusinessWrite bool) ([]string, error)
}

// UnmaskingWorkflowUserDirectory 用户领域：工单列表/详情等场景将用户 UID 解析为展示名。
// 由 DMS service 层适配 UserUsecase 实现。
type UnmaskingWorkflowUserDirectory interface {
	GetUserNamesByUIDs(ctx context.Context, uids []string) (map[string]string, error)
}

// UnmaskingWorkflowDBServiceDirectory 数据源领域：将 DBService UID 解析为实例名称。
// 由 DMS service 层适配 DBServiceUsecase 实现。
type UnmaskingWorkflowDBServiceDirectory interface {
	GetDBServiceNamesByUIDs(ctx context.Context, uids []string) (map[string]string, error)
}
