package locale

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// 在该文件中添加 i18n.Message 后还需生成对应语言文件（active.*.toml），脚本写在Makefile中了，使用步骤如下：
// 1. 安装需要的工具，已安装则跳过：
//		make install_i18n_tool
// 2. 将新增的i18n.Message提取到语言文件(active.*.toml)中：
//		make extract_i18n
// 3. 生成待翻译的临时文件(translate.en.toml)：
//		make start_trans_i18n
// 4. 人工介入将 translate.en.toml 文件中的文本翻译替换
// 5. 根据翻译好的文本更新英文文件(active.en.toml):
//		make end_trans_i18n

// Stat
var (
	StatOK      = &i18n.Message{ID: "StatOK", Other: "正常"}
	StatDisable = &i18n.Message{ID: "StatDisable", Other: "被禁用"}
	StatUnknown = &i18n.Message{ID: "StatUnknown", Other: "未知"}
)

// OpPermission
var (
	NameOpPermissionCreateProject           = &i18n.Message{ID: "NameOpPermissionCreateProject", Other: "项目总监"}
	NameOpPermissionProjectAdmin            = &i18n.Message{ID: "NameOpPermissionProjectAdmin", Other: "项目管理"}
	NameOpPermissionCreateWorkflow          = &i18n.Message{ID: "NameOpPermissionCreateWorkflow", Other: "创建上线工单"}
	NameOpPermissionAuditWorkflow           = &i18n.Message{ID: "NameOpPermissionAuditWorkflow", Other: "审批上线工单"}
	NameOpPermissionAuthDBServiceData       = &i18n.Message{ID: "NameOpPermissionAuthDBServiceData", Other: "账号管理"}
	NameOpPermissionExecuteWorkflow         = &i18n.Message{ID: "NameOpPermissionExecuteWorkflow", Other: "执行上线工单"}
	NameOpPermissionViewOthersWorkflow      = &i18n.Message{ID: "NameOpPermissionViewOthersWorkflow", Other: "查看所有工单"}
	NameOpPermissionViewOthersAuditPlan     = &i18n.Message{ID: "NameOpPermissionViewOthersAuditPlan", Other: "访问所有管控SQL"}
	NameOpPermissionViewSQLInsight          = &i18n.Message{ID: "NameOpPermissionViewSQLInsight", Other: "查看性能洞察"}
	NameOpPermissionSaveAuditPlan           = &i18n.Message{ID: "NameOpPermissionSaveAuditPlan", Other: "配置SQL管控"}
	NameOpPermissionSQLQuery                = &i18n.Message{ID: "NameOpPermissionSQLQuery", Other: "SQL工作台操作权限"}
	NameOpPermissionExportApprovalReject    = &i18n.Message{ID: "NameOpPermissionExportApprovalReject", Other: "审批导出工单"}
	NameOpPermissionExportCreate            = &i18n.Message{ID: "NameOpPermissionExportCreate", Other: "创建导出工单"}
	NameOpPermissionCreateOptimization      = &i18n.Message{ID: "NameOpPermissionCreateOptimization", Other: "创建智能调优"}
	NameOpPermissionGlobalManagement        = &i18n.Message{ID: "NameOpPermissionGlobalManagement", Other: "系统管理员"}
	NameOpPermissionGlobalView              = &i18n.Message{ID: "NameOpPermissionGlobalView", Other: "审计管理员"}
	NameOpPermissionViewOthersOptimization  = &i18n.Message{ID: "NameOpPermissionViewOthersOptimization", Other: "查看他人创建的智能调优"}
	NameOpPermissionCreatePipeline          = &i18n.Message{ID: "NameOpPermissionCreatePipeline", Other: "流水线增删改"}
	NameOpPermissionOrdinaryUser            = &i18n.Message{ID: "NameOpPermissionOrdinaryUser", Other: "普通用户"}
	NameOpPermissionViewOperationRecord     = &i18n.Message{ID: "NameOpPermissionViewOperationRecord", Other: "查看所有操作记录"}
	NameOpPermissionViewExportTask          = &i18n.Message{ID: "NameOpPermissionViewExportTask", Other: "查看所有导出任务"}
	NamePermissionViewQuickAuditRecord      = &i18n.Message{ID: "NamePermissionViewQuickAuditRecord", Other: "查看所有快捷审核记录"}
	NameOpPermissionViewIDEAuditRecord      = &i18n.Message{ID: "NameOpPermissionViewIDEAuditRecord", Other: "查看所有IDE审核记录"}
	NameOpPermissionViewOptimizationRecord  = &i18n.Message{ID: "NameOpPermissionViewOptimizationRecord", Other: "查看所有优化记录"}
	NameOpPermissionViewVersionManage       = &i18n.Message{ID: "NameOpPermissionViewVersionManage", Other: "查看他人创建的版本记录"}
	NameOpPermissionVersionManage           = &i18n.Message{ID: "NameOpPermissionVersionManage", Other: "配置版本"}
	NameOpPermissionViewPipeline            = &i18n.Message{ID: "NameOpPermissionViewPipeline", Other: "查看所有流水线"}
	NameOpPermissionManageProjectDataSource = &i18n.Message{ID: "NameOpPermissionManageProjectDataSource", Other: "管理项目数据源"}
	NameOpPermissionManageAuditRuleTemplate = &i18n.Message{ID: "NameOpPermissionManageAuditRuleTemplate", Other: "管理审核规则模版"}
	NameOpPermissionManageApprovalTemplate  = &i18n.Message{ID: "NameOpPermissionManageApprovalTemplate", Other: "管理审批流程模版"}
	NameOpPermissionManageMember            = &i18n.Message{ID: "NameOpPermissionManageMember", Other: "管理成员与权限"}
	NameOpPermissionPushRule                = &i18n.Message{ID: "NameOpPermissionPushRule", Other: "管理推送规则"}
	NameOpPermissionMangeAuditSQLWhiteList  = &i18n.Message{ID: "NameOpPermissionMangeAuditSQLWhiteList", Other: "审核SQL例外"}
	NameOpPermissionManageSQLMangeWhiteList = &i18n.Message{ID: "NameOpPermissionManageSQLMangeWhiteList", Other: "管控SQL例外"}
	NameOpPermissionManageRoleMange         = &i18n.Message{ID: "NameOpPermissionManageRoleMange", Other: "角色管理权限"}
	NameOpPermissionDesensitization         = &i18n.Message{ID: "NameOpPermissionDesensitization", Other: "脱敏规则配置权限"}

	DescOpPermissionGlobalManagement       = &i18n.Message{ID: "DescOpPermissionGlobalManagement", Other: "具备系统最高权限，可进行系统配置、用户管理等操作"}
	DescOpPermissionGlobalView             = &i18n.Message{ID: "DescOpPermissionGlobalView", Other: "负责系统操作审计、数据合规检查等工作"}
	DescOpPermissionCreateProject          = &i18n.Message{ID: "DescOpPermissionCreateProject", Other: "创建项目、配置项目资源"}
	DescOpPermissionProjectAdmin           = &i18n.Message{ID: "DescOpPermissionProjectAdmin", Other: "项目管理；拥有该权限的用户可以管理项目下的所有资源"}
	DescOpPermissionCreateWorkflow         = &i18n.Message{ID: "DescOpPermissionCreateWorkflow", Other: "创建/编辑工单；拥有该权限的用户可以创建/编辑工单"}
	DescOpPermissionOrdinaryUser           = &i18n.Message{ID: "DescOpPermissionOrdinaryUser", Other: "基础功能操作权限，可进行日常业务操作"}
	DescOpPermissionAuditWorkflow          = &i18n.Message{ID: "DescOpPermissionAuditWorkflow", Other: "审核/驳回工单；拥有该权限的用户可以审核/驳回工单"}
	DescOpPermissionAuthDBServiceData      = &i18n.Message{ID: "DescOpPermissionAuthDBServiceData", Other: "授权数据源数据权限；拥有该权限的用户可以授权数据源数据权限"}
	DescOpPermissionExecuteWorkflow        = &i18n.Message{ID: "DescOpPermissionExecuteWorkflow", Other: "上线工单；拥有该权限的用户可以上线工单"}
	DescOpPermissionViewOthersWorkflow     = &i18n.Message{ID: "DescOpPermissionViewOthersWorkflow", Other: "查看他人创建的工单；拥有该权限的用户可以查看他人创建的工单"}
	DescOpPermissionViewOthersAuditPlan    = &i18n.Message{ID: "DescOpPermissionViewOthersAuditPlan", Other: "查看他人创建的扫描任务；拥有该权限的用户可以查看他人创建的扫描任务"}
	DescOpPermissionViewSQLInsight         = &i18n.Message{ID: "DescOpPermissionViewSQLInsight", Other: "查看性能洞察；拥有该权限的用户可以查看性能洞察的数据"}
	DescOpPermissionSaveAuditPlan          = &i18n.Message{ID: "DescOpPermissionSaveAuditPlan", Other: "创建/编辑扫描任务；拥有该权限的用户可以创建/编辑扫描任务"}
	DescOpPermissionSQLQuery               = &i18n.Message{ID: "DescOpPermissionSQLQuery", Other: "SQL工作台查询；拥有该权限的用户可以执行SQL工作台查询"}
	DescOpPermissionExportApprovalReject   = &i18n.Message{ID: "DescOpPermissionExportApprovalReject", Other: "审批/驳回数据导出工单；拥有该权限的用户可以执行审批导出数据工单或者驳回导出数据工单"}
	DescOpPermissionExportCreate           = &i18n.Message{ID: "DescOpPermissionExportCreate", Other: "创建数据导出任务；拥有该权限的用户可以创建数据导出任务或者工单"}
	DescOpPermissionCreateOptimization     = &i18n.Message{ID: "DescOpPermissionCreateOptimization", Other: "创建智能调优；拥有该权限的用户可以创建智能调优"}
	DescOpPermissionViewOthersOptimization = &i18n.Message{ID: "DescOpPermissionViewOthersOptimization", Other: "查看他人创建的智能调优；拥有该权限的用户可以查看他人创建的智能调优"}
	DescOpPermissionCreatePipeline         = &i18n.Message{ID: "DescOpPermissionCreatePipeline", Other: "配置流水线；拥有该权限的用户可以为数据源配置流水线"}
)

// role
var (
	NameRoleProjectAdmin = &i18n.Message{ID: "NameRoleProjectAdmin", Other: "项目管理员"}
	NameRoleDevEngineer  = &i18n.Message{ID: "NameRoleDevEngineer", Other: "开发工程师"}
	NameRoleDevManager   = &i18n.Message{ID: "NameRoleDevManager", Other: "开发主管"}
	NameRoleOpsEngineer  = &i18n.Message{ID: "NameRoleOpsEngineer", Other: "运维工程师"}

	DescRoleProjectAdmin = &i18n.Message{ID: "DescRoleProjectAdmin", Other: "project admin"}
	DescRoleDevEngineer  = &i18n.Message{ID: "DescRoleDevEngineer", Other: "拥有该权限的用户可以创建/编辑工单,SQL工作台查询,配置流水线,创建智能调优"}
	DescRoleDevManager   = &i18n.Message{ID: "DescRoleDevManager", Other: "拥有该权限的用户可以创建/编辑工单,审核/驳回工单,配置流水线,查看他人创建的智能调优"}
	DescRoleOpsEngineer  = &i18n.Message{ID: "DescRoleOpsEngineer", Other: "拥有该权限的用户可以上线工单,查看他人创建的工单,创建智能扫描,查看他人的扫描任务,数据导出"}
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
	DBServiceDbName                     = &i18n.Message{ID: "DBServiceDbName", Other: "数据源名称"}
	DBServiceProjName                   = &i18n.Message{ID: "DBServiceProjName", Other: "所属项目(平台已有的项目名称)"}
	DBServiceEnvironmentTag             = &i18n.Message{ID: "DBServiceEnvironmentTag", Other: "所属环境"}
	DBServiceDesc                       = &i18n.Message{ID: "DBServiceDesc", Other: "数据源描述"}
	DBServiceDbType                     = &i18n.Message{ID: "DBServiceDbType", Other: "数据源类型"}
	DBServiceHost                       = &i18n.Message{ID: "DBServiceHost", Other: "数据源地址"}
	DBServicePort                       = &i18n.Message{ID: "DBServicePort", Other: "数据源端口"}
	DBServiceUser                       = &i18n.Message{ID: "DBServiceUser", Other: "数据源连接用户"}
	DBServicePassword                   = &i18n.Message{ID: "DBServicePassword", Other: "数据源密码"}
	DBServiceOracleService              = &i18n.Message{ID: "DBServiceOracleService", Other: "服务名(Oracle需填)"}
	DBServiceDB2DbName                  = &i18n.Message{ID: "DBServiceDB2DbName", Other: "数据库名(DB2需填)"}
	DBServiceOpsTime                    = &i18n.Message{ID: "DBServiceOpsTime", Other: "运维时间(非必填，9:30-11:00;14:10-18:30)"}
	DBServiceRuleTemplateName           = &i18n.Message{ID: "DBServiceRuleTemplateName", Other: "审核规则模板(项目已有的规则模板)"}
	DBServiceSQLQueryRuleTemplateName   = &i18n.Message{ID: "DBServiceSQLQueryRuleTemplateName", Other: "工作台操作审核规则模板(需要先填写审核规则模板)"}
	DBServiceDataExportRuleTemplateName = &i18n.Message{ID: "DBServiceDataExportRuleTemplateName", Other: "数据导出审核规则模板(需要先填写审核规则模板)"}
	DBServiceAuditLevel                 = &i18n.Message{ID: "DBServiceAuditLevel", Other: "工作台查询的最高审核等级[error|warn|notice|normal]"}
	DBServiceProblem                    = &i18n.Message{ID: "DBServiceProblem", Other: "问题"}

	DBServiceNoProblem                    = &i18n.Message{ID: "DBServiceNoProblem", Other: "无"}
	IDBPCErrMissingOrInvalidCols          = &i18n.Message{ID: "IDBPCErrMissingOrInvalidCols", Other: "缺失或不规范的列：%s"}
	IDBPCErrInvalidInput                  = &i18n.Message{ID: "IDBPCErrInvalidInput", Other: "若无特别说明每列均为必填"}
	IDBPCErrProjNonExist                  = &i18n.Message{ID: "IDBPCErrProjNonExist", Other: "所属项目不存在"}
	IDBPCErrProjNotActive                 = &i18n.Message{ID: "IDBPCErrProjNotActive", Other: "所属项目状态异常"}
	IDBPCErrProjNotAllowed                = &i18n.Message{ID: "IDBPCErrProjNotAllowed", Other: "所属项目不是操作中的项目"}
	IDBPCErrOptTimeInvalid                = &i18n.Message{ID: "IDBPCErrOptTimeInvalid", Other: "运维时间不规范"}
	IDBPCErrDbTypeInvalid                 = &i18n.Message{ID: "IDBPCErrDbTypeInvalid", Other: "数据源类型不规范或对应插件未安装"}
	IDBPCErrOracleServiceNameInvalid      = &i18n.Message{ID: "IDBPCErrOracleServiceNameInvalid", Other: "Oracle服务名错误"}
	IDBPCErrDB2DbNameInvalid              = &i18n.Message{ID: "IDBPCErrDB2DbNameInvalid", Other: "DB2数据库名错误"}
	IDBPCErrRuleTemplateInvalid           = &i18n.Message{ID: "IDBPCErrRuleTemplateInvalid", Other: "审核规则模板不存在或数据源类型不匹配"}
	IDBPCErrSQLQueryRuleTemplateInvalid   = &i18n.Message{ID: "IDBPCErrSQLQueryRuleTemplateInvalid", Other: "工作台操作审核规则模板不存在或数据源类型不匹配"}
	IDBPCErrDataExportRuleTemplateInvalid = &i18n.Message{ID: "IDBPCErrDataExportRuleTemplateInvalid", Other: "数据导出审核规则模板不存在或数据源类型不匹配"}
	IDBPCErrRuleTemplateBaseCheck         = &i18n.Message{ID: "IDBPCErrRuleTemplateBaseCheck", Other: "需要先添加审核规则模板，工作台和数据导出审核模板才会生效"}
	IDBPCErrEnvironmentTagInvalid         = &i18n.Message{ID: "IDBPCErrEnvironmentTagInvalid", Other: "项目环境标签检查错误或不存在"}
)

// project
var (
	ProjectName         = &i18n.Message{ID: "ProjectName", Other: "项目名称"}
	ProjectDesc         = &i18n.Message{ID: "ProjectDesc", Other: "项目描述"}
	ProjectStatus       = &i18n.Message{ID: "ProjectStatus", Other: "项目状态"}
	ProjectBusiness     = &i18n.Message{ID: "ProjectBusiness", Other: "所属业务"}
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

// OAuth2
var (
	OAuth2GetConfigErr                          = &i18n.Message{ID: "OAuth2GetConfigErr", Other: "获取OAuth2配置失败: %v"}
	OAuth2ProcessErr                            = &i18n.Message{ID: "OAuth2ProcessErr", Other: "OAuth2流程错误: %v"}
	OAuth2GetTokenErr                           = &i18n.Message{ID: "OAuth2GetTokenErr", Other: "OAuth2流程获取Token错误: %v"}
	OAuth2BackendLogoutFailed                   = &i18n.Message{ID: "OAuth2BackendLogoutFailed", Other: "；注销第三方平台会话失败: %v"}
	OAuth2BackendLogoutSuccess                  = &i18n.Message{ID: "OAuth2BackendLogoutSuccess", Other: "；已注销第三方平台会话"}
	OAuth2HandleTokenErr                        = &i18n.Message{ID: "OAuth2HandleTokenErr", Other: "处理 OAuth2 Token 错误: %v"}
	OAuth2GetUserInfoErr                        = &i18n.Message{ID: "OAuth2GetUserInfoErr", Other: "获取 OAuth2 用户信息错误: %v"}
	OAuth2QueryBindUserByOAuthIDErr             = &i18n.Message{ID: "OAuth2QueryBindUserByOAuthIDErr", Other: "通过 OAuth2 用户ID查询绑定用户错误: %v"}
	OAuth2QueryBindUserBySameNameErr            = &i18n.Message{ID: "OAuth2QueryBindUserBySameNameErr", Other: "通过 OAuth2 用户ID查询同名用户错误: %v"}
	OAuth2SameNameUserIsBoundErr                = &i18n.Message{ID: "OAuth2SameNameUserIsBoundErr", Other: "通过 OAuth2 用户ID %q 查询到的同名用户已经被绑定"}
	OAuth2UserNotBoundAndNoPermErr              = &i18n.Message{ID: "OAuth2UserNotBoundAndNoPermErr", Other: "该OAuth2用户未绑定且没有登陆权限"}
	OAuth2AutoCreateUserWithoutDefaultPwdErr    = &i18n.Message{ID: "OAuth2AutoCreateUserWithoutDefaultPwdErr", Other: "自动创建用户失败，默认密码未配置"}
	OAuth2AutoCreateUserErr                     = &i18n.Message{ID: "OAuth2AutoCreateUserErr", Other: "自动创建用户失败: %v"}
	OAuth2UserNotBoundAndDisableManuallyBindErr = &i18n.Message{ID: "OAuth2UserNotBoundAndDisableManuallyBindErr", Other: "未查询到 %q 关联的用户且关闭了手动绑定功能，请联系系统管理员"}
	OAuth2UserStatIsDisableErr                  = &i18n.Message{ID: "OAuth2UserStatIsDisableErr", Other: "用户 %q 被禁用"}
	OAuth2SyncSessionErr                        = &i18n.Message{ID: "OAuth2SyncSessionErr", Other: "同步OAuth2会话失败: %v"}
)

// Data Export Workflow
var (
	DataWorkflowDefault                     = &i18n.Message{ID: "DataWorkflowDefault", Other: "❓数据导出工单未知请求"}
	DataWorkflowExportFailed                = &i18n.Message{ID: "DataWorkflowExportFailed", Other: "⚠️ 数据导出失败"}
	DataWorkflowExportSuccess               = &i18n.Message{ID: "DataWorkflowExportSuccess", Other: "✅ 数据导出成功"}
	DataWorkflowReject                      = &i18n.Message{ID: "DataWorkflowReject", Other: "❌ 数据导出工单被驳回"}
	DataWorkflowWaitExporting               = &i18n.Message{ID: "DataWorkflowWaitExporting", Other: "⏳ 数据导出工单待导出"}
	DataWorkflowWaiting                     = &i18n.Message{ID: "DataWorkflowWaiting", Other: "🔍 数据导出工单待审批"}
	NotifyDataWorkflowBodyConfigUrl         = &i18n.Message{ID: "NotifyDataWorkflowBodyConfigUrl", Other: "请在系统设置-全局配置中补充全局url"}
	NotifyDataWorkflowBodyHead              = &i18n.Message{ID: "NotifyDataWorkflowBodyHead", Other: "\n📋 数据导出工单主题: %v\n📍 所属项目： %v\n🆔 数据导出工单ID: %v\n📝 数据导出工单描述: %v\n👤 申请人: %v\n⏰ 创建时间: %v\n"}
	NotifyDataWorkflowBodyInstanceAndSchema = &i18n.Message{ID: "NotifyDataWorkflowBodyInstanceAndSchema", Other: "🗄️ 数据源: %v\n📊 schema: %v\n"}
	NotifyDataWorkflowBodyLink              = &i18n.Message{ID: "NotifyDataWorkflowBodyLink", Other: "🔗 数据导出工单链接: %v"}
	NotifyDataWorkflowBodyReason            = &i18n.Message{ID: "NotifyDataWorkflowBodyReason", Other: "❌ 驳回原因: %v"}
	NotifyDataWorkflowBodyReport            = &i18n.Message{ID: "NotifyDataWorkflowBodyReport", Other: "⭐ 数据导出工单审核得分: %v"}
	NotifyDataWorkflowBodyStartEnd          = &i18n.Message{ID: "NotifyDataWorkflowBodyStartEnd", Other: "▶️ 数据导出开始时间: %v\n◀️ 数据导出结束时间: %v"}
	NotifyDataWorkflowBodyWorkFlowErr       = &i18n.Message{ID: "NotifyDataWorkflowBodyWorkFlowErr", Other: "❌ 读取工单任务内容失败，请通过SQLE界面确认工单状态"}
	NotifyDataWorkflowBodyApprovalReminder  = &i18n.Message{ID: "NotifyDataWorkflowBodyApprovalReminder", Other: "⏰ 导出工单已审批通过，请在1天内完成导出，过期后将无法执行"}
)
