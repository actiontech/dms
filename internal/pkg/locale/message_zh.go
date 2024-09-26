package locale

import (
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Stat
var (
	StatOK      = &i18n.Message{ID: "StatOK", Other: string(dmsCommonV1.StatOK)}
	StatDisable = &i18n.Message{ID: "StatDisable", Other: string(dmsCommonV1.StatDisable)}
	StatUnknown = &i18n.Message{ID: "StatUnknown", Other: string(dmsCommonV1.StatUnknown)}
)

// OpPermission
var (
	NameOpPermissionCreateProject          = &i18n.Message{ID: "NameOpPermissionCreateProject", Other: "创建项目"}
	NameOpPermissionProjectAdmin           = &i18n.Message{ID: "NameOpPermissionProjectAdmin", Other: "项目管理"}
	NameOpPermissionCreateWorkflow         = &i18n.Message{ID: "NameOpPermissionCreateWorkflow", Other: "创建/编辑工单"}
	NameOpPermissionAuditWorkflow          = &i18n.Message{ID: "NameOpPermissionAuditWorkflow", Other: "审核/驳回工单"}
	NameOpPermissionAuthDBServiceData      = &i18n.Message{ID: "NameOpPermissionAuthDBServiceData", Other: "授权数据源数据权限"}
	NameOpPermissionExecuteWorkflow        = &i18n.Message{ID: "NameOpPermissionExecuteWorkflow", Other: "上线工单"}
	NameOpPermissionViewOthersWorkflow     = &i18n.Message{ID: "NameOpPermissionViewOthersWorkflow", Other: "查看他人创建的工单"}
	NameOpPermissionViewOthersAuditPlan    = &i18n.Message{ID: "NameOpPermissionViewOthersAuditPlan", Other: "创建/编辑扫描任务"}
	NameOpPermissionSaveAuditPlan          = &i18n.Message{ID: "NameOpPermissionSaveAuditPlan", Other: "查看他人创建的扫描任务"}
	NameOpPermissionSQLQuery               = &i18n.Message{ID: "NameOpPermissionSQLQuery", Other: "SQL查询"}
	NameOpPermissionExportApprovalReject   = &i18n.Message{ID: "NameOpPermissionExportApprovalReject", Other: "审批/驳回数据导出工单"}
	NameOpPermissionExportCreate           = &i18n.Message{ID: "NameOpPermissionExportCreate", Other: "创建数据导出任务"}
	NameOpPermissionCreateOptimization     = &i18n.Message{ID: "NameOpPermissionCreateOptimization", Other: "创建智能调优"}
	NameOpPermissionViewOthersOptimization = &i18n.Message{ID: "NameOpPermissionViewOthersOptimization", Other: "查看他人创建的智能调优"}
	NameOpPermissionCreatePipeline         = &i18n.Message{ID: "NameOpPermissionCreatePipeline", Other: "配置流水线"}

	DescOpPermissionCreateProject          = &i18n.Message{ID: "DescOpPermissionCreateProject", Other: "创建项目；创建项目的用户自动拥有该项目管理权限"}
	DescOpPermissionProjectAdmin           = &i18n.Message{ID: "DescOpPermissionProjectAdmin", Other: "项目管理；拥有该权限的用户可以管理项目下的所有资源"}
	DescOpPermissionCreateWorkflow         = &i18n.Message{ID: "DescOpPermissionCreateWorkflow", Other: "创建/编辑工单；拥有该权限的用户可以创建/编辑工单"}
	DescOpPermissionAuditWorkflow          = &i18n.Message{ID: "DescOpPermissionAuditWorkflow", Other: "审核/驳回工单；拥有该权限的用户可以审核/驳回工单"}
	DescOpPermissionAuthDBServiceData      = &i18n.Message{ID: "DescOpPermissionAuthDBServiceData", Other: "授权数据源数据权限；拥有该权限的用户可以授权数据源数据权限"}
	DescOpPermissionExecuteWorkflow        = &i18n.Message{ID: "DescOpPermissionExecuteWorkflow", Other: "上线工单；拥有该权限的用户可以上线工单"}
	DescOpPermissionViewOthersWorkflow     = &i18n.Message{ID: "DescOpPermissionViewOthersWorkflow", Other: "查看他人创建的工单；拥有该权限的用户可以查看他人创建的工单"}
	DescOpPermissionViewOthersAuditPlan    = &i18n.Message{ID: "DescOpPermissionViewOthersAuditPlan", Other: "创建/编辑扫描任务；拥有该权限的用户可以创建/编辑扫描任务"}
	DescOpPermissionSaveAuditPlan          = &i18n.Message{ID: "DescOpPermissionSaveAuditPlan", Other: "查看他人创建的扫描任务；拥有该权限的用户可以查看他人创建的扫描任务"}
	DescOpPermissionSQLQuery               = &i18n.Message{ID: "DescOpPermissionSQLQuery", Other: "SQL查询；拥有该权限的用户可以执行SQL查询"}
	DescOpPermissionExportApprovalReject   = &i18n.Message{ID: "DescOpPermissionExportApprovalReject", Other: "审批/驳回数据导出工单；拥有该权限的用户可以执行审批导出数据工单或者驳回导出数据工单"}
	DescOpPermissionExportCreate           = &i18n.Message{ID: "DescOpPermissionExportCreate", Other: "创建数据导出任务；拥有该权限的用户可以创建数据导出任务或者工单"}
	DescOpPermissionCreateOptimization     = &i18n.Message{ID: "DescOpPermissionCreateOptimization", Other: "创建智能调优；拥有该权限的用户可以创建智能调优"}
	DescOpPermissionViewOthersOptimization = &i18n.Message{ID: "DescOpPermissionViewOthersOptimization", Other: "查看他人创建的智能调优；拥有该权限的用户可以查看他人创建的智能调优"}
	DescOpPermissionCreatePipeline         = &i18n.Message{ID: "DescOpPermissionCreatePipeline", Other: "配置流水线；拥有该权限的用户可以为数据源配置流水线"}
)

// role
var (
	NameRoleProjectAdmin   = &i18n.Message{ID: "NameRoleProjectAdmin", Other: "项目管理员"}
	NameRoleSQLEAdmin      = &i18n.Message{ID: "NameRoleSQLEAdmin", Other: "SQLE管理员"}
	NameRoleProvisionAdmin = &i18n.Message{ID: "NameRoleProvisionAdmin", Other: "provision管理员"}

	DescRoleProjectAdmin   = &i18n.Message{ID: "DescRoleProjectAdmin", Other: "project admin"}
	DescRoleSQLEAdmin      = &i18n.Message{ID: "DescRoleSQLEAdmin", Other: "拥有该权限的用户可以创建/编辑工单，审核/驳回工单，上线工单,创建/编辑扫描任务"}
	DescRoleProvisionAdmin = &i18n.Message{ID: "DescRoleProvisionAdmin", Other: "拥有该权限的用户可以授权数据源数据权限"}
)

// license
var (
	LicenseInstanceNum           = &i18n.Message{ID: "LicenseInstanceNum", Other: "实例数"}
	LicenseUserNum               = &i18n.Message{ID: "LicenseUserNum", Other: "用户数"}
	LicenseAuthorizedDurationDay = &i18n.Message{ID: "LicenseAuthorizedDurationDay", Other: "授权运行时长(天)"}
	LicenseUnlimited             = &i18n.Message{ID: "LicenseUnlimited", Other: "无限制"}
	LicenseDurationOfRunning     = &i18n.Message{ID: "LicenseDurationOfRunning", Other: "已运行时长(天)"}
	LicenseEstimatedMaturity     = &i18n.Message{ID: "LicenseEstimatedMaturity", Other: "预计到期时间"}
	LicenseResourceTypeUser      = &i18n.Message{ID: "LicenseResourceTypeUser", Other: "用户"}
	LicenseInstanceNumOfType     = &i18n.Message{ID: "LicenseInstanceNumOfType", Other: "[%v]类型实例数"}
	LicenseMachineInfo           = &i18n.Message{ID: "LicenseMachineInfo", Other: "机器信息"}
	LicenseMachineInfoOfNode     = &i18n.Message{ID: "LicenseMachineInfoOfNode", Other: "节点[%s]机器信息"}
	LicenseDmsVersion            = &i18n.Message{ID: "LicenseDmsVersion", Other: "DMS版本"}
)

// DB service
var (
	DBServiceDbName           = &i18n.Message{ID: "DBServiceDbName", Other: "数据源名称"}
	DBServiceProjName         = &i18n.Message{ID: "DBServiceProjName", Other: "所属项目(平台已有的项目名称)"}
	DBServiceBusiness         = &i18n.Message{ID: "DBServiceBusiness", Other: "所属业务(项目已有的业务名称)"}
	DBServiceDesc             = &i18n.Message{ID: "DBServiceDesc", Other: "数据源描述"}
	DBServiceDbType           = &i18n.Message{ID: "DBServiceDbType", Other: "数据源类型"}
	DBServiceHost             = &i18n.Message{ID: "DBServiceHost", Other: "数据源地址"}
	DBServicePort             = &i18n.Message{ID: "DBServicePort", Other: "数据源端口"}
	DBServiceUser             = &i18n.Message{ID: "DBServiceUser", Other: "数据源连接用户"}
	DBServicePassword         = &i18n.Message{ID: "DBServicePassword", Other: "数据源密码"}
	DBServiceOracleService    = &i18n.Message{ID: "DBServiceOracleService", Other: "服务名(Oracle需填)"}
	DBServiceDB2DbName        = &i18n.Message{ID: "DBServiceDB2DbName", Other: "数据库名(DB2需填)"}
	DBServiceOpsTime          = &i18n.Message{ID: "DBServiceOpsTime", Other: "运维时间(非必填，9:30-11:00;14:10-18:30)"}
	DBServiceRuleTemplateName = &i18n.Message{ID: "DBServiceRuleTemplateName", Other: "审核规则模板(项目已有的规则模板)"}
	DBServiceAuditLevel       = &i18n.Message{ID: "DBServiceAuditLevel", Other: "工作台查询的最高审核等级[error|warn|notice|normal]"}
	DBServiceProblem          = &i18n.Message{ID: "DBServiceProblem", Other: "问题"}

	DBServiceNoProblem               = &i18n.Message{ID: "DBServiceNoProblem", Other: "无"}
	IDBPCErrMissingOrInvalidCols     = &i18n.Message{ID: "IDBPCErrMissingOrInvalidCols", Other: "缺失或不规范的列：%s"}
	IDBPCErrInvalidInput             = &i18n.Message{ID: "IDBPCErrInvalidInput", Other: "若无特别说明每列均为必填"}
	IDBPCErrProjNonExist             = &i18n.Message{ID: "IDBPCErrProjNonExist", Other: "所属项目不存在"}
	IDBPCErrProjNotActive            = &i18n.Message{ID: "IDBPCErrProjNotActive", Other: "所属项目状态异常"}
	IDBPCErrProjNotAllowed           = &i18n.Message{ID: "IDBPCErrProjNotAllowed", Other: "所属项目不是操作中的项目"}
	IDBPCErrBusinessNonExist         = &i18n.Message{ID: "IDBPCErrBusinessNonExist", Other: "项目业务固定且所属业务不存在"}
	IDBPCErrOptTimeInvalid           = &i18n.Message{ID: "IDBPCErrOptTimeInvalid", Other: "运维时间不规范"}
	IDBPCErrDbTypeInvalid            = &i18n.Message{ID: "IDBPCErrDbTypeInvalid", Other: "数据源类型不规范或对应插件未安装"}
	IDBPCErrOracleServiceNameInvalid = &i18n.Message{ID: "IDBPCErrOracleServiceNameInvalid", Other: "Oracle服务名错误"}
	IDBPCErrDB2DbNameInvalid         = &i18n.Message{ID: "IDBPCErrDB2DbNameInvalid", Other: "DB2数据库名错误"}
	IDBPCErrRuleTemplateInvalid      = &i18n.Message{ID: "IDBPCErrRuleTemplateInvalid", Other: "审核规则模板不存在或数据源类型不匹配"}
)

// project
var (
	ProjectName         = &i18n.Message{ID: "ProjectName", Other: "项目名称"}
	ProjectDesc         = &i18n.Message{ID: "ProjectDesc", Other: "项目描述"}
	ProjectStatus       = &i18n.Message{ID: "ProjectStatus", Other: "项目状态"}
	ProjectBusiness     = &i18n.Message{ID: "ProjectBusiness", Other: "可用业务"}
	ProjectCreateTime   = &i18n.Message{ID: "ProjectCreateTime", Other: "创建时间"}
	ProjectAvailable    = &i18n.Message{ID: "ProjectAvailable", Other: "可用"}
	ProjectNotAvailable = &i18n.Message{ID: "ProjectNotAvailable", Other: "不可用"}
)

// cb
var (
	CbOpDetailDelData    = &i18n.Message{ID: "CbOpDetailDelData", Other: "在数据源:%s中删除了以下数据:%s"}
	CbOpDetailAddData    = &i18n.Message{ID: "CbOpDetailAddData", Other: "在数据源:%s中添加了以下数据:%s"}
	CbOpDetailUpdateData = &i18n.Message{ID: "CbOpDetailUpdateData", Other: "在数据源:%s中更新了以下数据:%s"}

	CbOpTotalExecutions        = &i18n.Message{ID: "CbOpTotalExecutions", Other: "执行总量:"}
	CbOpSuccessRate            = &i18n.Message{ID: "CbOpSuccessRate", Other: "执行成功率:"}
	CbOpAuditBlockedSQL        = &i18n.Message{ID: "CbOpAuditBlockedSQL", Other: "审核拦截的异常SQL:"}
	CbOpUnsuccessfulExecutions = &i18n.Message{ID: "CbOpUnsuccessfulExecutions", Other: "执行不成功的SQL:"}

	CbOpProjectName     = &i18n.Message{ID: "CbOpProjectName", Other: "项目名"}
	CbOpOperator        = &i18n.Message{ID: "CbOpOperator", Other: "操作人"}
	CbOpOperationTime   = &i18n.Message{ID: "CbOpOperationTime", Other: "操作时间"}
	CbOpDataSource      = &i18n.Message{ID: "CbOpDataSource", Other: "数据源"}
	CbOpDetails         = &i18n.Message{ID: "CbOpDetails", Other: "操作详情"}
	CbOpSessionID       = &i18n.Message{ID: "CbOpSessionID", Other: "会话ID"}
	CbOpOperationIP     = &i18n.Message{ID: "CbOpOperationIP", Other: "操作IP"}
	CbOpAuditResult     = &i18n.Message{ID: "CbOpAuditResult", Other: "审核结果"}
	CbOpExecutionResult = &i18n.Message{ID: "CbOpExecutionResult", Other: "执行结果"}
	CbOpExecutionTimeMs = &i18n.Message{ID: "CbOpExecutionTimeMs", Other: "执行时间(毫秒)"}
	CbOpResultRowCount  = &i18n.Message{ID: "CbOpResultRowCount", Other: "结果集返回行数"}
)

// DB Service Sync Task
var (
	DBServiceSyncVersion = &i18n.Message{ID: "DBServiceSyncVersion", Other: "版本(支持DMP5.23.04.0及以上版本)"}
	DBServiceSyncExpand  = &i18n.Message{ID: "DBServiceSyncExpand", Other: "数据源同步扩展服务"}
)
