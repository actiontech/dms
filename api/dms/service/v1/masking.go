package v1

import (
	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters ListMaskingRules
type ListMaskingRulesReq struct {
}

// swagger:model ListMaskingRulesData
type ListMaskingRulesData struct {
	// masking type
	// Example: "MASK_DIGIT"
	MaskingType string `json:"masking_type"`
	// description
	// Example: "mask digits"
	Description string `json:"description"`
	// effect description for users
	// Example: "保留开头2位和结尾2位，中间字符替换为*"
	Effect string `json:"effect"`
	// effect example before masking
	// Example: "13812345678"
	EffectExampleBefore string `json:"effect_example_before"`
	// effect example after masking
	// Example: "138******78"
	EffectExampleAfter string `json:"effect_example_after"`
	// masking rule id
	// Example: 1
	Id int `json:"id"`
}

// swagger:model ListMaskingRulesReply
type ListMaskingRulesReply struct {
	// list masking rule reply
	Data []ListMaskingRulesData `json:"data"`

	base.GenericResp
}

// swagger:parameters ListMaskingTemplates
type ListMaskingTemplatesReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// the maximum count of masking templates to be returned, default is 20
	// in: query
	PageSize uint32 `query:"page_size" json:"page_size"`
	// the offset of masking templates to be returned, default is 0
	// in: query
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:model ListMaskingTemplatesData
type ListMaskingTemplatesData struct {
	// masking template id
	// Example: 1
	Id int `json:"id"`
	// masking template name
	// Example: "Standard Template"
	Name string `json:"name"`
	// count of rules in the template
	// Example: 5
	RuleCount int `json:"rule_count"`
	// preview of rule name in the template, up to 3 items
	RuleNames []string `json:"rule_names"`
}

// swagger:model ListMaskingTemplatesReply
type ListMaskingTemplatesReply struct {
	// list masking templates reply
	Data []ListMaskingTemplatesData `json:"data"`
	// total count of masking templates
	// Example: 100
	Total int64 `json:"total_nums"`

	base.GenericResp
}

// swagger:model AddMaskingTemplateReq
type AddMaskingTemplateReq struct {
	// project uid
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// masking template
	// Required: true
	MaskingTemplate *AddMaskingTemplate `json:"masking_template" validate:"required"`
}

// swagger:model AddMaskingTemplate
type AddMaskingTemplate struct {
	// masking template name
	// Required: true
	// Example: "New Template"
	Name string `json:"name" validate:"required"`
	// masking rule id list
	// Required: true
	// MinLength: 1
	// Example: [1, 2, 3]
	RuleIDs []int `json:"rule_ids" validate:"required,min=1"`
}

// swagger:model AddMaskingTemplateReply
type AddMaskingTemplateReply struct {
	base.GenericResp
}

// swagger:model UpdateMaskingTemplateReq
type UpdateMaskingTemplateReq struct {
	// project uid
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// masking template id
	// in: path
	// swagger:ignore
	// Required: true
	// Example: 1
	TemplateID int `param:"template_id" json:"template_id" validate:"required"`
	// masking template
	// Required: true
	MaskingTemplate *UpdateMaskingTemplate `json:"masking_template" validate:"required"`
}

// swagger:model UpdateMaskingTemplate
type UpdateMaskingTemplate struct {
	// masking rule id list
	// Required: true
	// MinLength: 1
	// Example: [1, 2]
	RuleIDs []int `json:"rule_ids" validate:"required,min=1"`
}

// swagger:model UpdateMaskingTemplateReply
type UpdateMaskingTemplateReply struct {
	base.GenericResp
}

// swagger:parameters DeleteMaskingTemplate
type DeleteMaskingTemplateReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// masking template id
	// in: path
	// Required: true
	// Example: 1
	TemplateID int `param:"template_id" json:"template_id" validate:"required"`
}

// swagger:model DeleteMaskingTemplateReply
type DeleteMaskingTemplateReply struct {
	base.GenericResp
}

// swagger:parameters ListSensitiveDataDiscoveryTasks
type ListSensitiveDataDiscoveryTasksReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// the maximum count of tasks to be returned, default is 20
	// in: query
	// Example: 20
	PageSize uint32 `query:"page_size" json:"page_size"`
	// the offset of tasks to be returned, default is 0
	// in: query
	// Example: 0
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:enum SensitiveDataDiscoveryTaskType
type SensitiveDataDiscoveryTaskType string

const (
	SensitiveDataDiscoveryTaskTypePeriodic SensitiveDataDiscoveryTaskType = "PERIODIC" // 周期性任务
	SensitiveDataDiscoveryTaskTypeOneTime  SensitiveDataDiscoveryTaskType = "ONE_TIME" // 一次性任务
)

// swagger:enum SensitiveDataDiscoveryTaskStatus
type SensitiveDataDiscoveryTaskStatus string

const (
	SensitiveDataDiscoveryTaskStatusPendingChangeConfirm SensitiveDataDiscoveryTaskStatus = "PENDING_CONFIRM"
	SensitiveDataDiscoveryTaskStatusNormal               SensitiveDataDiscoveryTaskStatus = "NORMAL"
	SensitiveDataDiscoveryTaskStatusCompleted            SensitiveDataDiscoveryTaskStatus = "COMPLETED"
	SensitiveDataDiscoveryTaskStatusRunning              SensitiveDataDiscoveryTaskStatus = "RUNNING"
	SensitiveDataDiscoveryTaskStatusFailed               SensitiveDataDiscoveryTaskStatus = "FAILED"
)

// swagger:model ListSensitiveDataDiscoveryTasksData
type ListSensitiveDataDiscoveryTasksData struct {
	// sensitive data discovery task id
	// Example: 1
	ID int `json:"id"`
	// database instance id
	// Example: "db_service_uid_1"
	DBServiceUID string `json:"db_service_uid"`
	// database instance name
	// Example: "mysql-01"
	DBServiceName string `json:"db_service_name"`
	// task type
	// Example: "PERIODIC"
	TaskType SensitiveDataDiscoveryTaskType `json:"task_type"`
	// sensitive data identification method
	// Example: "BY_FIELD_NAME"
	IdentificationMethod SensitiveDataIdentificationMethod `json:"identification_method"`
	// execution plan
	// Example: "ONE_TIME"
	ExecutionPlan SensitiveDataDiscoveryTaskType `json:"execution_plan"`
	// whether periodic scanning is enabled
	// Example: true
	IsPeriodicScanEnabled bool `json:"is_periodic_scan_enabled"`
	// cron expression of execution frequency, periodic task returns cron, one-time task returns empty
	// Example: "0 2 * * *"
	ExecutionFrequency string `json:"execution_frequency"`
	// related masking template id
	// Example: 1
	MaskingTemplateID int `json:"masking_template_id"`
	// related masking template name
	// Example: "Standard Template"
	MaskingTemplateName string `json:"masking_template_name"`
	// next run time, periodic task returns RFC3339 time, one-time task returns null
	// Format: date-time (RFC3339)
	// Example: "2024-01-15T10:30:00Z"
	NextExecutionAt *string `json:"next_execution_at"`
	// task status
	// Example: "NORMAL"
	Status SensitiveDataDiscoveryTaskStatus `json:"status"`
}

// swagger:model ListSensitiveDataDiscoveryTasksReply
type ListSensitiveDataDiscoveryTasksReply struct {
	// sensitive data discovery tasks list reply
	Data []ListSensitiveDataDiscoveryTasksData `json:"data"`
	// total count of sensitive data discovery tasks
	// Example: 100
	Total int64 `json:"total_nums"`

	base.GenericResp
}

// swagger:enum SensitiveDataIdentificationMethod
type SensitiveDataIdentificationMethod string

const (
	SensitiveDataIdentificationMethodByFieldName  SensitiveDataIdentificationMethod = "BY_FIELD_NAME"
	SensitiveDataIdentificationMethodBySampleData SensitiveDataIdentificationMethod = "BY_SAMPLE_DATA"
)

// swagger:model AddSensitiveDataDiscoveryTaskReq
type AddSensitiveDataDiscoveryTaskReq struct {
	// project uid
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// sensitive data discovery task
	// Required: true
	Task *AddSensitiveDataDiscoveryTask `json:"task" validate:"required"`
}

// swagger:enum ConfidenceLevel
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "HIGH"
	ConfidenceMedium ConfidenceLevel = "MEDIUM"
	ConfidenceLow    ConfidenceLevel = "LOW"
)

// swagger:model AddSensitiveDataDiscoveryTask
type AddSensitiveDataDiscoveryTask struct {
	// database instance id
	// Required: true
	// Example: "1"
	DBServiceUID string `json:"db_service_uid" validate:"required"`
	// masking template id
	// Required: true
	// Example: 1
	MaskingTemplateID int `json:"masking_template_id"`
	// sensitive data identification method
	// Required: true
	// Example: "BY_FIELD_NAME"
	IdentificationMethod SensitiveDataIdentificationMethod `json:"identification_method" validate:"required,oneof=BY_FIELD_NAME BY_SAMPLE_DATA"`
	// execution plan
	// Required: true
	// Example: "ONE_TIME"
	ExecutionPlan SensitiveDataDiscoveryTaskType `json:"execution_plan" validate:"required,oneof=PERIODIC ONE_TIME"`
	// whether periodic scanning is enabled, default is true
	// Example: true
	IsPeriodicScanEnabled *bool `json:"is_periodic_scan_enabled"`
	// cron expression, required when execution_plan is PERIODIC
	// Example: "0 0 * * *"
	CronExpression string `json:"cron_expression"`
}

// swagger:model SensitiveFieldScanResult
type SensitiveFieldScanResult struct {
	// scan information for the field
	// Example: "matched by field name 'email'"
	ScanInfo string `json:"scan_info"`
	// confidence level
	// Example: "High"
	Confidence ConfidenceLevel `json:"confidence"`
	// recommended masking rule id
	// Example: 1
	RecommendedMaskingRuleID int `json:"recommended_masking_rule_id"`
	// recommended masking rule name
	// Example: "Email Masking"
	RecommendedMaskingRuleName string `json:"recommended_masking_rule_name"`
}

// swagger:model SuspectedSensitiveFieldsTree
type SuspectedSensitiveFieldsTree struct {
	// database_name -> database node
	Databases map[string]SuspectedSensitiveDatabaseNode `json:"databases"`
}

// swagger:model SuspectedSensitiveDatabaseNode
type SuspectedSensitiveDatabaseNode struct {
	// table_name -> table node
	Tables map[string]SuspectedSensitiveTableNode `json:"tables"`
}

// swagger:model SuspectedSensitiveTableNode
type SuspectedSensitiveTableNode struct {
	// field_name -> scan result
	Fields map[string]SensitiveFieldScanResult `json:"fields"`
}

// swagger:model AddSensitiveDataDiscoveryTaskData
type AddSensitiveDataDiscoveryTaskData struct {
	// suspected sensitive fields tree
	SuspectedSensitiveFieldsTree SuspectedSensitiveFieldsTree `json:"suspected_sensitive_fields_tree"`
}

// swagger:model AddSensitiveDataDiscoveryTaskReply
type AddSensitiveDataDiscoveryTaskReply struct {
	// add sensitive data discovery task reply
	Data AddSensitiveDataDiscoveryTaskData `json:"data"`

	base.GenericResp
}

// swagger:model UpdateSensitiveDataDiscoveryTaskReq
type UpdateSensitiveDataDiscoveryTaskReq struct {
	// project uid
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// sensitive data discovery task id
	// in: path
	// swagger:ignore
	// Required: true
	// Example: 1
	TaskID int `param:"task_id" json:"task_id" validate:"required"`
	// sensitive data discovery task
	// Required: true
	Task *UpdateSensitiveDataDiscoveryTask `json:"task" validate:"required"`
}

// swagger:model UpdateSensitiveDataDiscoveryTask
type UpdateSensitiveDataDiscoveryTask struct {
	// masking template id
	// Required: true
	// Example: 1
	MaskingTemplateID int `json:"masking_template_id"`
	// sensitive data identification method
	// Required: true
	// Example: "BY_FIELD_NAME"
	IdentificationMethod SensitiveDataIdentificationMethod `json:"identification_method" validate:"required,oneof=BY_FIELD_NAME BY_SAMPLE_DATA"`
	// execution plan
	// Required: true
	// Example: "PERIODIC"
	ExecutionPlan SensitiveDataDiscoveryTaskType `json:"execution_plan" validate:"required,oneof=PERIODIC ONE_TIME"`
	// whether periodic scanning is enabled
	// Example: true
	IsPeriodicScanEnabled *bool `json:"is_periodic_scan_enabled"`
	// cron expression, required when execution_plan is PERIODIC
	// Example: "0 0 * * *"
	CronExpression string `json:"cron_expression"`
}

// swagger:model UpdateSensitiveDataDiscoveryTaskData
type UpdateSensitiveDataDiscoveryTaskData struct {
	// suspected sensitive fields tree
	SuspectedSensitiveFieldsTree SuspectedSensitiveFieldsTree `json:"suspected_sensitive_fields_tree"`
}

// swagger:model UpdateSensitiveDataDiscoveryTaskReply
type UpdateSensitiveDataDiscoveryTaskReply struct {
	// update sensitive data discovery task reply
	Data UpdateSensitiveDataDiscoveryTaskData `json:"data"`

	base.GenericResp
}

// swagger:parameters DeleteSensitiveDataDiscoveryTask
type DeleteSensitiveDataDiscoveryTaskReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// sensitive data discovery task id
	// in: path
	// Required: true
	// Example: 1
	TaskID int `param:"task_id" json:"task_id" validate:"required"`
}

// swagger:model DeleteSensitiveDataDiscoveryTaskReply
type DeleteSensitiveDataDiscoveryTaskReply struct {
	base.GenericResp
}

// swagger:parameters ListSensitiveDataDiscoveryTaskHistories
type ListSensitiveDataDiscoveryTaskHistoriesReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// sensitive data discovery task id
	// in: path
	// Required: true
	// Example: 1
	TaskID int `param:"task_id" json:"task_id" validate:"required"`
	// the maximum count of histories to be returned, default is 20
	// in: query
	// Example: 20
	PageSize uint32 `query:"page_size" json:"page_size"`
	// the offset of histories to be returned, default is 0
	// in: query
	// Example: 0
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:model ListSensitiveDataDiscoveryTaskHistoriesData
type ListSensitiveDataDiscoveryTaskHistoriesData struct {
	// execution time in RFC3339 format
	// Format: date-time (RFC3339)
	// Example: "2024-01-15T10:30:00Z"
	ExecutedAt string `json:"executed_at"`
	// execution status
	// Example: "NORMAL"
	Status SensitiveDataDiscoveryTaskStatus `json:"status"`
	// newly discovered sensitive field count
	// Example: 10
	NewSensitiveFieldCount int `json:"new_sensitive_field_count"`
	// remark
	// Example: "scan completed successfully"
	Remark string `json:"remark"`
}

// swagger:model ListSensitiveDataDiscoveryTaskHistoriesReply
type ListSensitiveDataDiscoveryTaskHistoriesReply struct {
	// sensitive data discovery task histories reply
	Data []ListSensitiveDataDiscoveryTaskHistoriesData `json:"data"`
	// total count of sensitive data discovery task histories
	// Example: 100
	Total int64 `json:"total_nums"`

	base.GenericResp
}

// swagger:model ConfigureMaskingRulesReq
type ConfigureMaskingRulesReq struct {
	// project uid
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// masking rule configurations for batch create or update
	// Required: true
	// MinLength: 1
	MaskingRuleConfigs []MaskingRuleConfig `json:"masking_rule_configs" validate:"required,min=1"`
}

// swagger:model MaskingRuleConfig
type MaskingRuleConfig struct {
	// data source id
	// Required: true
	// Example: "1"
	DBServiceUID string `json:"db_service_uid" validate:"required"`
	// schema name
	// Required: true
	// Example: "db1"
	SchemaName string `json:"schema_name" validate:"required"`
	// table name
	// Required: true
	// Example: "users"
	TableName string `json:"table_name" validate:"required"`
	// column name
	// Required: true
	// Example: "email"
	ColumnName string `json:"column_name" validate:"required"`
	// masking rule id
	// Required: true
	// Example: 1
	MaskingRuleID int `json:"masking_rule_id" validate:"required"`
	// whether to enable masking for this column
	// Required: true
	// Example: true
	IsMaskingEnabled bool `json:"is_masking_enabled" validate:"required"`
}

// swagger:model ConfigureMaskingRulesReply
type ConfigureMaskingRulesReply struct {
	base.GenericResp
}

// swagger:enum MaskingConfigStatus
type MaskingConfigStatus string

const (
	MaskingConfigStatusConfigured      MaskingConfigStatus = "CONFIGURED"
	MaskingConfigStatusPendingConfirm  MaskingConfigStatus = "PENDING_CONFIRM"
	MaskingConfigStatusSystemConfirmed MaskingConfigStatus = "SYSTEM_CONFIRMED"
)

// swagger:parameters GetMaskingOverviewTree
type GetMaskingOverviewTreeReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// data source id
	// in: query
	// Required: true
	// Example: "1"
	DBServiceUID string `query:"db_service_uid" json:"db_service_uid" validate:"required"`
	// fuzzy search keywords for column name
	// in: query
	// Example: "user"
	Keywords string `query:"keywords" json:"keywords"`
	// masking config status filters
	// in: query
	MaskingConfigStatus MaskingConfigStatus `query:"masking_config_statuses" json:"masking_config_statuses"`
}

// swagger:model MaskingOverviewDashboard
type MaskingOverviewDashboard struct {
	// total count of tables that contain sensitive data
	// Example: 50
	TotalSensitiveTables int `json:"total_sensitive_tables"`
	// total count of columns with configured masking
	// Example: 120
	ConfiguredMaskingColumns int `json:"configured_masking_columns"`
	// total count of columns pending masking confirmation
	// Example: 5
	PendingConfirmMaskingColumns int `json:"pending_confirm_masking_columns"`
}

// swagger:model MaskingOverviewTableData
type MaskingOverviewTableData struct {
	// table id
	// Example: 1
	TableID int `json:"table_id"`
	// configured masking column count for this table
	// Example: 3
	ConfiguredMaskingColumns int `json:"configured_masking_columns"`
	// pending masking confirmation column count for this table
	// Example: 1
	PendingConfirmMaskingColumns int `json:"pending_confirm_masking_columns"`
}

// swagger:model MaskingOverviewDatabaseNode
type MaskingOverviewDatabaseNode struct {
	// table_name -> table overview data
	Tables map[string]MaskingOverviewTableData `json:"tables"`
}

// swagger:model GetMaskingOverviewTreeData
type GetMaskingOverviewTreeData struct {
	// dashboard summary for the selected data source
	Dashboard MaskingOverviewDashboard `json:"dashboard"`
	// database_name -> database node
	Databases map[string]MaskingOverviewDatabaseNode `json:"databases"`
}

// swagger:model GetMaskingOverviewTreeReply
type GetMaskingOverviewTreeReply struct {
	// masking overview tree reply
	Data GetMaskingOverviewTreeData `json:"data"`

	base.GenericResp
}

// swagger:parameters GetTableColumnMaskingDetails
type GetTableColumnMaskingDetailsReq struct {
	// project uid
	//
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// table id from masking overview tree
	// in: path
	// Required: true
	// Example: 1
	TableID int `param:"table_id" json:"table_id" validate:"required"`
	// fuzzy search keywords for column name
	// in: query
	// Example: "phone"
	Keywords string `query:"keywords" json:"keywords"`
}

// swagger:model TableColumnMaskingDetail
type TableColumnMaskingDetail struct {
	// column name
	// Example: "email"
	ColumnName string `json:"column_name"`
	// current masking rule id, null if no masking rule is applied
	//
	// Example: 1
	MaskingRuleID *int `json:"masking_rule_id"`
	// current masking rule name, null if no masking rule is applied
	//
	// Example: "Email Masking"
	MaskingRuleName *string `json:"masking_rule_name"`
	// confidence level of masking recommendation，null if no masking rule is applied
	//
	// Example: 2
	Confidence *ConfidenceLevel `json:"confidence"`
	// current masking config status
	Status MaskingConfigStatus `json:"status"`
}

// swagger:model GetTableColumnMaskingDetailsReply
type GetTableColumnMaskingDetailsReply struct {
	// table column masking details reply
	Data []TableColumnMaskingDetail `json:"data"`

	base.GenericResp
}

// swagger:parameters ListPendingApprovalRequests
type ListPendingApprovalRequestsReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// the maximum count of requests to be returned, default is 20
	// in: query
	// Example: 20
	PageSize uint32 `query:"page_size" json:"page_size"`
	// the offset of requests to be returned, default is 0
	// in: query
	// Example: 0
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:model PendingApprovalRequestData
type PendingApprovalRequestData struct {
	// approval request id
	// Example: 1
	ID int `json:"id"`
	// applicant name
	// Example: "admin"
	ApplicantName string `json:"applicant_name"`
	// application time in RFC3339 format
	// Format: date-time (RFC3339)
	// Example: "2024-01-15T10:30:00Z"
	AppliedAt string `json:"applied_at"`
	// application reason
	// Example: "data analysis"
	Reason string `json:"reason"`
	// data scope
	// Example: "database 'db1', table 'users'"
	DataScope string `json:"data_scope"`
}

// swagger:model ListPendingApprovalRequestsReply
type ListPendingApprovalRequestsReply struct {
	// pending approval requests reply
	Data []PendingApprovalRequestData `json:"data"`
	// total count of pending approval requests
	// Example: 100
	Total int64 `json:"total_nums"`

	base.GenericResp
}

// swagger:enum ApprovalAction
type ApprovalAction string

const (
	ApprovalActionApprove ApprovalAction = "APPROVE"
	ApprovalActionReject  ApprovalAction = "REJECT"
)

// swagger:model ProcessApprovalRequestReq
type ProcessApprovalRequestReq struct {
	// project uid
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// approval request id
	// in: path
	// swagger:ignore
	// Required: true
	// Example: 1
	RequestID int `param:"request_id" json:"request_id" validate:"required"`
	// process action
	// Required: true
	// Example: "APPROVE"
	Action ApprovalAction `json:"action" validate:"required"`
	// reject reason, required when action is REJECT
	// Example: "insufficient reason"
	RejectReason string `json:"reject_reason"`
	// approval remark, optional when action is APPROVE
	// Example: "approved for one-time access"
	ApproveRemark string `json:"approve_remark"`
}

// swagger:model ProcessApprovalRequestReply
type ProcessApprovalRequestReply struct {
	base.GenericResp
}

// swagger:parameters GetPlaintextAccessRequestDetail
type GetPlaintextAccessRequestDetailReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// approval request id
	// in: path
	// Required: true
	// Example: 1
	RequestID int `param:"request_id" json:"request_id" validate:"required"`
}

// swagger:model MaskingPreviewData
type MaskingPreviewData struct {
	// preview columns
	// Example: ["id", "name", "email"]
	Columns []string `json:"columns"`
	// preview rows
	// Example: [["1", "John", "j***@example.com"], ["2", "Alice", "a***@example.com"]]
	Rows [][]string `json:"rows"`
}

// swagger:model GetPlaintextAccessRequestDetailReply
type GetPlaintextAccessRequestDetailReply struct {
	// plaintext access request detail reply
	Data struct {
		// query sql statement
		// Example: "SELECT * FROM users"
		QuerySQL string `json:"query_sql"`
		// masking result preview
		MaskingPreview MaskingPreviewData `json:"masking_preview"`
		// application reason
		// Example: "troubleshooting"
		Reason string `json:"reason"`
	} `json:"data"`

	base.GenericResp
}
