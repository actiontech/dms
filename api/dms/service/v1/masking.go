package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters ListMaskingRules
type ListMaskingRulesReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 规则来源筛选: builtin 或 custom，为空时返回全部
	// in: query
	// Example: "custom"
	FilterRuleSource MaskingRuleSource `query:"source" json:"source"`
	// 模糊搜索关键字，匹配规则名称、描述或敏感数据类型名称
	// in: query
	// Example: "手机"
	FuzzyKeyword string `query:"keywords" json:"keywords"`
	// 分页大小，默认 20
	// in: query
	PageSize uint32 `query:"page_size" json:"page_size"`
	// 页码（从1开始），默认 1
	// in: query
	PageIndex uint32 `query:"page_index" json:"page_index"`
}

// swagger:enum MaskingRuleSource
type MaskingRuleSource string

const (
	MaskingRuleSourceBuiltin MaskingRuleSource = "builtin" // 系统内置
	MaskingRuleSourceCustom  MaskingRuleSource = "custom"  // 自定义
)

// swagger:enum SensitiveDataTypeSource
type SensitiveDataTypeSource string

const (
	SensitiveDataTypeSourceBuiltin SensitiveDataTypeSource = "builtin" // 系统内置
	SensitiveDataTypeSourceCustom  SensitiveDataTypeSource = "custom"  // 自定义
)

// swagger:enum MaskingAlgorithmDisplayName
type MaskingAlgorithmDisplayName string

const (
	MaskingAlgorithmDisplayNameCHAR    MaskingAlgorithmDisplayName = "CHAR"
	MaskingAlgorithmDisplayNameTAG     MaskingAlgorithmDisplayName = "TAG"
	MaskingAlgorithmDisplayNameREPLACE MaskingAlgorithmDisplayName = "REPLACE"
	MaskingAlgorithmDisplayNameHASH    MaskingAlgorithmDisplayName = "HASH"
)

// swagger:model ListMaskingRulesData
type ListMaskingRulesData struct {
	// 脱敏规则ID
	// Example: 1
	Id int `json:"id"`
	// 脱敏规则名称
	// Example: "手机号脱敏"
	Name string `json:"name"`
	// 规则来源: 系统内置或自定义
	// Example: "builtin"
	RuleSource MaskingRuleSource `json:"source"`
	// 敏感数据类型名称
	// Example: "电子邮箱"
	SensitiveDataTypeName string `json:"sensitive_type_name"`
	// 敏感数据类型描述信息
	// Example: "属于账户联系信息，泄露会带来账户安全风险"
	SensitiveTypeInfo string `json:"sensitive_type_info"`
	// 是否为自定义敏感类型
	// Example: false
	IsCustomType bool `json:"is_custom_type"`
	// 脱敏算法类型: CHAR, TAG, REPLACE, HASH
	// Example: "CHAR"
	AlgorithmType string `json:"algorithm_type"`
	// description
	// Example: "mask digits"
	Description string `json:"description"`
	// 脱敏效果
	// Example: "保留开头2位和结尾2位，中间字符替换为*"
	Effect string `json:"effect"`
	// 脱敏前示例
	// Example: "13812345678"
	EffectExampleBefore string `json:"effect_example_before"`
	// 脱敏后示例
	// Example: "138******78"
	EffectExampleAfter string `json:"effect_example_after"`
}

// swagger:model ListMaskingRulesReply
type ListMaskingRulesReply struct {
	// list masking rule reply
	Data []ListMaskingRulesData `json:"data"`
	// total count of masking rules
	Total int64 `json:"total_nums"`

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

// swagger:model MaskingTemplateRuleRef
type MaskingTemplateRuleRef struct {
	// rule id
	// Required: true
	RuleID int `json:"rule_id"`
	// rule source: builtin or custom
	RuleSource string `json:"rule_source"`
}

// swagger:model AddMaskingTemplate
type AddMaskingTemplate struct {
	// masking template name
	// Required: true
	// Example: "New Template"
	Name string `json:"name" validate:"required"`
	// masking rule id list (deprecated, use rule_refs instead)
	// Example: [1, 2, 3]
	RuleIDs []int `json:"rule_ids"`
	// masking rule references with source info
	// Example: [{"rule_id": 1, "rule_source": "builtin"}]
	RuleRefs []MaskingTemplateRuleRef `json:"rule_refs"`
}

// NormalizeRuleIDs populates RuleIDs from RuleRefs if RuleIDs is empty.
// Returns an error if neither RuleIDs nor RuleRefs is provided.
func (t *AddMaskingTemplate) NormalizeRuleIDs() error {
	if len(t.RuleIDs) == 0 && len(t.RuleRefs) > 0 {
		t.RuleIDs = make([]int, 0, len(t.RuleRefs))
		for _, ref := range t.RuleRefs {
			t.RuleIDs = append(t.RuleIDs, ref.RuleID)
		}
	}
	if len(t.RuleIDs) == 0 {
		return fmt.Errorf("rule_ids or rule_refs is required and must not be empty")
	}
	return nil
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
	// masking rule id list (deprecated, use rule_refs instead)
	// Example: [1, 2]
	RuleIDs []int `json:"rule_ids"`
	// masking rule references with source info
	// Example: [{"rule_id": 1, "rule_source": "builtin"}]
	RuleRefs []MaskingTemplateRuleRef `json:"rule_refs"`
}

// NormalizeRuleIDs populates RuleIDs from RuleRefs if RuleIDs is empty.
// Returns an error if neither RuleIDs nor RuleRefs is provided.
func (t *UpdateMaskingTemplate) NormalizeRuleIDs() error {
	if len(t.RuleIDs) == 0 && len(t.RuleRefs) > 0 {
		t.RuleIDs = make([]int, 0, len(t.RuleRefs))
		for _, ref := range t.RuleRefs {
			t.RuleIDs = append(t.RuleIDs, ref.RuleID)
		}
	}
	if len(t.RuleIDs) == 0 {
		return fmt.Errorf("rule_ids or rule_refs is required and must not be empty")
	}
	return nil
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
	SensitiveDataDiscoveryTaskStatusStopped              SensitiveDataDiscoveryTaskStatus = "STOPPED"
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
	// database instance host
	// Example: "10.10.10.10"
	DBServiceHost string `json:"db_service_host"`
	// database instance port
	// Example: "3306"
	DBServicePort string `json:"db_service_port"`
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
	// scan scope - database (schema) names, empty means all databases
	// Example: ["db1", "db2"]
	SchemaNames []string `json:"schema_names"`
	// scan scope - table names, empty means all tables
	// Example: ["users", "orders"]
	TableNames []string `json:"table_names"`
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
	// scan scope - database (schema) names, empty means all databases
	// Example: ["db1", "db2"]
	SchemaNames []string `json:"schema_names"`
	// scan scope - table names, empty means all tables
	// Example: ["users", "orders"]
	TableNames []string `json:"table_names"`
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
	// 规则来源: 系统内置或自定义
	// Example: "builtin"
	RuleSource MaskingRuleSource `json:"rule_source" validate:"required,oneof=builtin custom"`
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

// swagger:enum SensitiveDataDiscoveryTaskAction
type SensitiveDataDiscoveryTaskAction string

const (
	SensitiveDataDiscoveryTaskActionEnable    SensitiveDataDiscoveryTaskAction = "ENABLE"
	SensitiveDataDiscoveryTaskActionTerminate SensitiveDataDiscoveryTaskAction = "TERMINATE"
	SensitiveDataDiscoveryTaskActionUpdate    SensitiveDataDiscoveryTaskAction = "UPDATE"
)

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
	// action type: ENABLE(启用周期扫描), TERMINATE(终止周期扫描), UPDATE(更新配置)
	// Required: true
	// Example: "ENABLE"
	Action SensitiveDataDiscoveryTaskAction `json:"action" validate:"required,oneof=ENABLE TERMINATE UPDATE"`
	// task update data, required when action is UPDATE
	Task *UpdateSensitiveDataDiscoveryTask `json:"task"`
}

// swagger:model UpdateSensitiveDataDiscoveryTask
type UpdateSensitiveDataDiscoveryTask struct {
	// masking template id
	// Example: 1
	MaskingTemplateID int `json:"masking_template_id"`
	// sensitive data identification method
	// Example: "BY_FIELD_NAME"
	IdentificationMethod SensitiveDataIdentificationMethod `json:"identification_method" validate:"oneof=BY_FIELD_NAME BY_SAMPLE_DATA"`
	// execution plan
	// Example: "PERIODIC"
	ExecutionPlan SensitiveDataDiscoveryTaskType `json:"execution_plan" validate:"oneof=PERIODIC ONE_TIME"`
	// cron expression, only used when execution_plan is PERIODIC
	// Example: "0 0 * * *"
	CronExpression string `json:"cron_expression"`
	// scan scope - database (schema) names to add (edit mode: can only add, not remove)
	// Example: ["db1", "db2"]
	SchemaNames []string `json:"schema_names"`
	// scan scope - table names to add (edit mode: can only add, not remove)
	// Example: ["users", "orders"]
	TableNames []string `json:"table_names"`
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
	// Example: true
	IsMaskingEnabled bool `json:"is_masking_enabled"`
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
	// current masking rule source: "builtin" or "custom"
	//
	// Example: "builtin"
	MaskingRuleSource MaskingRuleSource `json:"masking_rule_source,omitempty"`
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
	// whether masking is enabled for this column
	IsMaskingEnabled bool `json:"is_masking_enabled"`
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
	Action ApprovalAction `json:"action" validate:"required,oneof=APPROVE REJECT"`
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

// swagger:parameters ListCreatableDBServicesForMaskingTask
// 用于获取可以创建敏感数据扫描任务的数据源列表
type ListCreatableDBServicesForMaskingTaskReq struct {
	// project uid
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// the maximum count of db services to be returned, default is 100
	// in: query
	// Example: 100
	PageSize uint32 `query:"page_size" json:"page_size"`
	// the offset of db services to be returned, default is 0
	// in: query
	// Example: 0
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// fuzzy search keywords for db service name
	// in: query
	// Example: "mysql"
	Keywords string `query:"keywords" json:"keywords"`
}

// swagger:model ListCreatableDBServicesForMaskingTaskData
// 可创建扫描任务的数据源数据
type ListCreatableDBServicesForMaskingTaskData struct {
	// database instance uid
	// Example: "db_service_uid_1"
	DBServiceUID string `json:"db_service_uid"`
	// database instance name
	// Example: "mysql-01"
	DBServiceName string `json:"db_service_name"`
	// database type
	// Example: "MySQL"
	DBType string `json:"db_type"`
	// database instance host
	// Example: "10.10.10.10"
	DBServiceHost string `json:"db_service_host"`
	// database instance port
	// Example: "3306"
	DBServicePort string `json:"db_service_port"`
	// whether this db service already has a masking task created
	HasTask bool `json:"has_task"`
}

// swagger:model ListCreatableDBServicesForMaskingTaskReply
type ListCreatableDBServicesForMaskingTaskReply struct {
	// list of db services that can create masking discovery task
	Data []ListCreatableDBServicesForMaskingTaskData `json:"data"`
	// total count of db services
	// Example: 10
	Total int64 `json:"total_nums"`

	base.GenericResp
}

// ===== 脱敏规则 API 结构体 =====

// swagger:model AddMaskingRuleReq
type AddMaskingRuleReq struct {
	// 项目 UID
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 脱敏规则
	// Required: true
	Rule AddMaskingRule `json:"rule" validate:"required"`
}

// swagger:model AddMaskingRule
type AddMaskingRule struct {
	// 规则名称
	// Required: true
	// Example: "自定义手机号脱敏"
	Name string `json:"name" validate:"required"`
	// 规则说明
	// Example: "对手机号进行脱敏处理"
	Description string `json:"description"`
	// 敏感数据类型记录uid
	// Required: true
	// Example: "20260409170256090"
	SensitiveDataTypeUID string `json:"sensitive_data_type_uid" validate:"required"`
	// 脱敏算法配置
	// Required: true
	MaskingAlgorithmConfig MaskingAlgorithmConfig `json:"masking_algorithm_config" validate:"required"`
}

// swagger:enum MaskingAlgorithmMaskType
type MaskingAlgorithmMaskType string

const (
	MaskingAlgorithmMaskTypeCHAR    MaskingAlgorithmMaskType = "CHAR"
	MaskingAlgorithmMaskTypeTAG     MaskingAlgorithmMaskType = "TAG"
	MaskingAlgorithmMaskTypeREPLACE MaskingAlgorithmMaskType = "REPLACE"
	MaskingAlgorithmMaskTypeALGO    MaskingAlgorithmMaskType = "ALGO"
)

// swagger:model MaskingAlgorithmConfig
type MaskingAlgorithmConfig struct {
	// 脱敏算法名称，对应已注册的算法标识（如 "FULL_MASK"、"GODLP" 等），由前端从算法列表中选择后传入
	// Required: true
	// Example: "FULL_MASK"
	MaskingAlgorithmName string `json:"masking_algorithm_name" validate:"required"`
	// 脱敏方式：CHAR（按字符掩码）、TAG（标签）、REPLACE（整体替换）、ALGO（算法）
	// Required: true
	// Example: "CHAR"
	MaskType MaskingAlgorithmMaskType `json:"mask_type" validate:"required,oneof=CHAR TAG REPLACE ALGO"`
	// 脱敏取值：CHAR 时为替换字符，TAG 时为标签文案，REPLACE 时为替换全文，ALGO 时为算法名称
	// Example: "*"
	Value string `json:"value"`
	// CHAR 方式时从开头起保留的字符数（偏移量）
	// Example: 3
	Offset int32 `json:"offset"`
	// CHAR 方式时从末尾起保留的字符数
	// Example: 4
	Padding int32 `json:"padding"`
	// CHAR 方式时中间需脱敏的字符长度，0 表示中间全部脱敏
	// Example: 0
	Length int32 `json:"length"`
	// CHAR 方式时是否从字符串末尾起计算脱敏区域
	// Example: false
	Reverse bool `json:"reverse"`
	// CHAR 方式脱敏时需忽略的字符集合
	// Example: "-"
	IgnoreCharSet string `json:"ignore_char_set"`
	// 示例数据
	// Example: ["13812345678", "13812345679"]
	SampleDataList []string `json:"sample_data_list"`
}

// swagger:model NewCustomSensitiveType
type NewCustomSensitiveType struct {
	// 中文名称
	// Required: true
	// Example: "自定义手机号"
	CnName string `json:"cn_name" validate:"required"`
	// 英文标识, 仅允许 [a-z0-9_]
	// Required: true
	// Example: "custom_phone"
	EnIdentifier string `json:"en_identifier" validate:"required"`
	// 字段名关键词
	// Example: ["phone", "mobile"]
	FieldKeywords []string `json:"field_keywords"`
	// 抽样数据正则表达式
	// Example: "^1[3-9]\\d{9}$"
	SampleDataRegexList []string `json:"sample_data_regex_list"`
	// 示例数据，当填写了抽样数据正则表达式时，必填
	// Example: ["John", "j***@example.com"]
	SampleDataList []string `json:"sample_data_list"`
}

// swagger:model AddSensitiveDataTypeReq
type AddSensitiveDataTypeReq struct {
	// 项目 UID
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 自定义敏感数据类型
	// Required: true
	Type *NewCustomSensitiveType `json:"type" validate:"required"`
}

// swagger:model AddSensitiveDataTypeReply
type AddSensitiveDataTypeReply struct {
	Data struct {
		// 新建敏感数据类型 id
		// Example: 1
		SensitiveDataTypeID uint `json:"sensitive_data_type_id"`
	} `json:"data"`

	base.GenericResp
}

// swagger:model AddMaskingRuleReply
type AddMaskingRuleReply struct {
	// 添加脱敏规则响应
	Data struct {
		// 新建规则 id
		// Example: 10001
		RuleID uint `json:"rule_id"`
	} `json:"data"`

	base.GenericResp
}

// swagger:model UpdateMaskingRuleReq
type UpdateMaskingRuleReq struct {
	// 项目 UID
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 规则 id
	// in: path
	// swagger:ignore
	// Required: true
	// Example: 10001
	RuleID uint `param:"rule_id" json:"rule_id" validate:"required"`
	// 脱敏规则更新内容
	// Required: true
	Rule *UpdateMaskingRule `json:"rule" validate:"required"`
}

// swagger:model UpdateMaskingRule
type UpdateMaskingRule struct {
	// 规则名称
	// Required: true
	// Example: "自定义手机号脱敏"
	Name string `json:"name" validate:"required"`
	// 规则说明
	// Example: "对手机号进行脱敏处理"
	Description string `json:"description"`
	// 脱敏算法配置，不修改可以不传递
	MaskingAlgorithmConfig *MaskingAlgorithmConfig `json:"masking_algorithm_config"`
}

// swagger:model UpdateSensitiveDataType
type UpdateSensitiveDataType struct {
	// 字段名关键词
	// Example: ["phone", "mobile"]
	FieldKeywords []string `json:"field_keywords"`
	// 抽样数据正则表达式
	// Example: "^1[3-9]\\d{9}$"
	SampleDataRegexList []string `json:"sample_data_regex_list"`
	// 示例数据，当填写了抽样数据正则表达式列表或字段名关键词时，必填
	// Example: ["John", "j***@example.com"]
	SampleDataList []string `json:"sample_data_list"`
}

// swagger:parameters UpdateSensitiveDataType
type UpdateSensitiveDataTypeReq struct {
	// 项目 UID
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 敏感数据类型 id（自定义类型的主键）
	// in: path
	// swagger:ignore
	// Required: true
	// Example: 1
	SensitiveDataTypeID uint `param:"sensitive_data_type_id" json:"sensitive_data_type_id" validate:"required"`
	// 更新内容
	// Required: true
	Type *UpdateSensitiveDataType `json:"type" validate:"required"`
}

// swagger:model UpdateSensitiveDataTypeReply
type UpdateSensitiveDataTypeReply struct {
	base.GenericResp
}

// swagger:model UpdateMaskingRuleReply
type UpdateMaskingRuleReply struct {
	base.GenericResp
}

// swagger:parameters DeleteMaskingRule
type DeleteMaskingRuleReq struct {
	// 项目 UID
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 规则 id
	// in: path
	// Required: true
	// Example: 10001
	RuleID uint `param:"rule_id" json:"rule_id" validate:"required"`
}

// swagger:model DeleteMaskingRuleReply
type DeleteMaskingRuleReply struct {
	base.GenericResp
}

// swagger:parameters GetMaskingRuleDetail
type GetMaskingRuleDetailReq struct {
	// 项目 UID
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 脱敏规则 id
	// in: path
	// Required: true
	// Example: 10001
	RuleID uint `param:"rule_id" json:"rule_id" validate:"required"`
}

// swagger:model GetMaskingRuleDetailReply
type GetMaskingRuleDetailReply struct {
	Data GetMaskingRuleDetailData `json:"data"`

	base.GenericResp
}

// swagger:model GetMaskingRuleDetailData
type GetMaskingRuleDetailData struct {
	// 规则 id
	// Example: 10001
	RuleID uint `json:"rule_id"`
	// 规则名称
	// Example: "自定义手机号脱敏"
	Name string `json:"name"`
	// 规则说明
	// Example: "对手机号进行脱敏处理"
	Description string `json:"description"`
	// 关联的敏感数据类型（内置或自定义）
	SensitiveDataType SensitiveTypeData `json:"sensitive_data_type"`
	// 脱敏算法配置
	MaskingAlgorithmConfig MaskingAlgorithmConfig `json:"masking_algorithm_config"`
}

// swagger:parameters DeleteSensitiveDataType
type DeleteSensitiveDataTypeReq struct {
	// 项目 UID
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 敏感数据类型 id（自定义类型为主键 id）
	// in: path
	// Required: true
	// Example: 1
	SensitiveDataTypeID uint `param:"sensitive_data_type_id" json:"sensitive_data_type_id" validate:"required"`
}

// swagger:model DeleteSensitiveDataTypeReply
type DeleteSensitiveDataTypeReply struct {
	base.GenericResp
}

// swagger:model TestSensitiveDataTypeMatchReq
type TestSensitiveDataTypeMatchReq struct {
	// 项目 UID
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 已存在的自定义敏感数据类型 id；与下方内联识别条件二选一或组合（由实现约定）
	// Example: 1
	SensitiveDataTypeID *uint `json:"sensitive_data_type_id"`
	// 字段名关键词（内联测试用）
	// Example: ["phone", "mobile"]
	FieldKeywords []string `json:"field_keywords"`
	// 抽样数据正则表达式列表（内联测试用）
	// Example: ["^1[3-9]\\d{9}$"]
	SampleDataRegexList []string `json:"sample_data_regex_list"`
	// 待匹配的样例值列表
	// Required: true
	SampleValues []string `json:"sample_values" validate:"required,min=1"`
}

// swagger:model TestSensitiveDataTypeMatchReply
type TestSensitiveDataTypeMatchReply struct {
	Data TestSensitiveDataTypeMatchData `json:"data"`

	base.GenericResp
}

// swagger:model TestSensitiveDataTypeMatchData
type TestSensitiveDataTypeMatchData struct {
	// 与请求顺序一一对应的匹配结果
	Results []SensitiveDataTypeMatchResult `json:"results"`
}

// swagger:model SensitiveDataTypeMatchResult
type SensitiveDataTypeMatchResult struct {
	// 样例值（与请求项对应）
	// Example: "13812345678"
	Value string `json:"value"`
	// 是否判定为匹配该敏感数据类型
	Matched bool `json:"matched"`
}

// swagger:parameters ListSensitiveTypes
type ListSensitiveTypesReq struct {
	// 项目 UID
	// in: path
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
}

// swagger:model SensitiveTypeData
type SensitiveTypeData struct {
	// ID
	// Example: 1
	ID uint `json:"id"`
	// 敏感数据类型来源: 系统内置或自定义
	// Example: "builtin"
	SensitiveDataTypeSource SensitiveDataTypeSource `json:"sensitive_data_type_source" validate:"required,oneof=builtin custom"`
	// 英文标识, 仅允许 [a-z0-9_]
	// Required: true
	// Example: "custom_phone"
	EnIdentifier string `json:"en_identifier" validate:"required"`
	// 中文展示名称
	// Example: "手机号"
	CnName string `json:"cn_name"`
	// 字段名关键词
	// Example: ["phone", "mobile"]
	FieldKeywords []string `json:"field_keywords"`
	// 抽样数据正则表达式
	// Example: "^1[3-9]\\d{9}$"
	SampleDataRegexList []string `json:"sample_data_regex_list"`
	// 示例数据，当填写了抽样数据正则表达式列表或字段名关键词时，必填
	// Example: ["John", "j***@example.com"]
	SampleDataList []string `json:"sample_data_list"`
	// 引用此敏感类型的脱敏规则数量（用于前端判断是否可编辑/删除）
	// Example: 2
	RuleCount int64 `json:"rule_count"`
}

// swagger:model ListSensitiveTypesReply
type ListSensitiveTypesReply struct {
	// 敏感数据类型列表
	Data []SensitiveTypeData `json:"data"`

	base.GenericResp
}

// swagger:model PreviewMaskingEffectReq
type PreviewMaskingEffectReq struct {
	// 项目 UID
	// in: path
	// swagger:ignore
	// Required: true
	// Example: "project_uid"
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// 示例数据列表
	// Required: true
	// Example: "13812345678"
	SampleInputList []string `json:"sample_input_list" validate:"required"`
	// 脱敏算法配置
	// Required: true
	MaskingAlgorithmConfig *MaskingAlgorithmConfig `json:"masking_algorithm_config" validate:"required"`
}

// swagger:model PreviewMaskingEffectReply
type PreviewMaskingEffectReply struct {
	// 脱敏效果预览响应
	Data struct {
		// 脱敏结果列表
		// Example: "138******78"
		MaskedOutputList []string `json:"masked_output_list"`
	} `json:"data"`

	base.GenericResp
}

// swagger:parameters ListDBServiceSchemasForMaskingTask
type ListDBServiceSchemasForMaskingTaskReq struct {
	// project uid
	// in: path
	// Required: true
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// database service uid
	// in: query
	// Required: true
	DBServiceUID string `query:"db_service_uid" json:"db_service_uid" validate:"required"`
}

// swagger:model DBServiceSchemaData
type DBServiceSchemaData struct {
	// schema (database) name
	// Example: "mydb"
	Name string `json:"name"`
}

// swagger:model ListDBServiceSchemasForMaskingTaskReply
type ListDBServiceSchemasForMaskingTaskReply struct {
	Data []DBServiceSchemaData `json:"data"`
	base.GenericResp
}

// swagger:parameters ListDBServiceTablesForMaskingTask
type ListDBServiceTablesForMaskingTaskReq struct {
	// project uid
	// in: path
	// Required: true
	ProjectUid string `param:"project_uid" json:"project_uid" validate:"required"`
	// database service uid
	// in: query
	// Required: true
	DBServiceUID string `query:"db_service_uid" json:"db_service_uid" validate:"required"`
	// schema (database) name
	// in: query
	// Required: true
	SchemaName string `query:"schema_name" json:"schema_name" validate:"required"`
}

// swagger:model DBServiceTableData
type DBServiceTableData struct {
	// table name
	// Example: "users"
	Name string `json:"name"`
}

// swagger:model ListDBServiceTablesForMaskingTaskReply
type ListDBServiceTablesForMaskingTaskReply struct {
	Data []DBServiceTableData `json:"data"`
	base.GenericResp
}
