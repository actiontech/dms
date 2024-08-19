package locale

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Stat
var (
	StatOK      = &i18n.Message{ID: "StatOK", Other: "正常"}
	StatDisable = &i18n.Message{ID: "StatDisable", Other: "被禁用"}
	StatUnknown = &i18n.Message{ID: "StatUnknown", Other: "未知"}
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
