package v1

import (
	"fmt"

	base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
)

// swagger:parameters GetUser
type GetUserReq struct {
	// user uid
	// in:path
	UserUid string `param:"user_uid" json:"user_uid" validate:"required"`
}

// A dms user
type GetUser struct {
	// user uid
	UserUid string `json:"uid"`
	// user name
	Name string `json:"name"`
	// user email
	Email string `json:"email"`
	// user phone
	Phone string `json:"phone"`
	// user wxid
	WxID string `json:"wxid"`
	// user language
	Language string `json:"language"`
	// user stat
	Stat Stat `json:"stat"`
	// user two factor enabled
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	// user authentication type
	AuthenticationType UserAuthenticationType `json:"authentication_type"`
	// user groups
	UserGroups []UidWithName `json:"user_groups"`
	// user operation permissions
	OpPermissions []UidWithName `json:"op_permissions"`
	// is admin
	IsAdmin bool `json:"is_admin"`
	// user bind name space
	UserBindProjects   []UserBindProject `json:"user_bind_projects"`
	ThirdPartyUserInfo string            `json:"third_party_user_info"`
	// access token
	AccessTokenInfo AccessTokenInfo `json:"access_token_info"`
}

type AccessTokenInfo struct {
	AccessToken string `json:"access_token"`
	ExpiredTime string `json:"token_expired_timestamp" example:"RFC3339"`
	IsExpired   bool   `json:"is_expired"`
}

type UserBindProject struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	IsManager   bool   `json:"is_manager"`
}

// swagger:enum Stat
type Stat string

const (
	StatOK        Stat = "正常"
	StatDisable   Stat = "被禁用"
	StatUnknown   Stat = "未知"
	StatOKEn      Stat = "Normal"
	StatDisableEn Stat = "Disabled"
	StatUnknownEn Stat = "Unknown"
)

type UidWithName struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

// swagger:enum UserAuthenticationType
type UserAuthenticationType string

const (
	UserAuthenticationTypeLDAP    UserAuthenticationType = "ldap"   // user verify through ldap
	UserAuthenticationTypeDMS     UserAuthenticationType = "dms"    // user verify through dms
	UserAuthenticationTypeOAUTH2  UserAuthenticationType = "oauth2" // user verify through oauth2
	UserAuthenticationTypeUnknown UserAuthenticationType = "unknown"
)

// swagger:model GetUserReply
type GetUserReply struct {
	// Get user reply
	Data *GetUser `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:parameters GetUserOpPermission
type GetUserOpPermissionReq struct {
	// user uid
	// in:path
	UserUid string `param:"user_uid" json:"user_uid" validate:"required"`

	// in:query
	// uesr project uid
	ProjectUid string `json:"project_uid" query:"project_uid"`
	// user op permission info
	// in:body
	UserOpPermission *UserOpPermission `json:"user_op_permission" `
}

type UserOpPermission struct {
	// uesr project uid
	ProjectUid string `json:"project_uid"`
}

// swagger:model GetUserOpPermissionReply
type GetUserOpPermissionReply struct {
	// user op permission reply
	// is user admin, admin has all permissions
	Data struct {
		IsAdmin bool `json:"is_admin"`
		// user op permissions
		OpPermissionList []OpPermissionItem `json:"op_permission_list"`
	} `json:"data"`
	// Generic reply
	base.GenericResp
}

type OpPermissionItem struct {
	// object uids, object type is defined by RangeType
	RangeUids []string `json:"range_uids"`
	// object type of RangeUids
	RangeType OpRangeType `json:"range_type"`
	// op permission type
	OpPermissionType OpPermissionType `json:"op_permission_type"`
}

// swagger:enum OpRangeType
type OpRangeType string

const (
	OpRangeTypeUnknown OpRangeType = "unknown"
	// 全局权限: 该权限只能被用户使用
	OpRangeTypeGlobal OpRangeType = "global"
	// 项目权限: 该权限只能被成员使用
	OpRangeTypeProject OpRangeType = "project"
	// 项目内的数据源权限: 该权限只能被成员使用
	OpRangeTypeDBService OpRangeType = "db_service"
)

func ParseOpRangeType(typ string) (OpRangeType, error) {
	switch typ {
	case string(OpRangeTypeDBService):
		return OpRangeTypeDBService, nil
	case string(OpRangeTypeProject):
		return OpRangeTypeProject, nil
	case string(OpRangeTypeGlobal):
		return OpRangeTypeGlobal, nil
	default:
		return "", fmt.Errorf("invalid op range type: %s", typ)
	}
}

// swagger:enum OpPermissionType
type OpPermissionType string

const (
	OpPermissionTypeUnknown OpPermissionType = "unknown"
	// 创建项目；创建项目的用户自动拥有该项目管理权限
	OpPermissionTypeCreateProject OpPermissionType = "create_project"
	// 项目管理；拥有该权限的用户可以管理项目下的所有资源
	OpPermissionTypeGlobalView OpPermissionType = "global_view"
	// 全局浏览；拥有该权限的用户可以浏览全局的资源
	OpPermissionTypeGlobalManagement OpPermissionType = "global_management"
	// 全局管理；拥有该权限的用户可以浏览和管理全局的资源
	OpPermissionTypeProjectAdmin OpPermissionType = "project_admin"
	// 创建/编辑工单；拥有该权限的用户可以创建/编辑工单
	OpPermissionTypeCreateWorkflow OpPermissionType = "create_workflow"
	// 审核/驳回工单；拥有该权限的用户可以审核/驳回工单
	OpPermissionTypeAuditWorkflow OpPermissionType = "audit_workflow"
	// 账号管理；拥有该权限的用户可以授权数据源数据权限
	OpPermissionTypeAuthDBServiceData OpPermissionType = "auth_db_service_data"
	// 查看其他工单权限
	OpPermissionTypeViewOthersWorkflow OpPermissionType = "view_others_workflow"
	// 上线工单；拥有该权限的用户可以上线工单
	OpPermissionTypeExecuteWorkflow OpPermissionType = "execute_workflow"
	// 查看其他扫描任务权限
	OpPermissionTypeViewOtherAuditPlan OpPermissionType = "view_other_audit_plan"
	// 创建扫描任务权限；拥有该权限的用户可以创建/更新扫描任务
	OpPermissionTypeSaveAuditPlan OpPermissionType = "save_audit_plan"
	//SQL查询；SQL查询权限
	OpPermissionTypeSQLQuery OpPermissionType = "sql_query"
	// 创建数据导出任务；拥有该权限的用户可以创建数据导出任务或者工单
	OpPermissionTypeExportCreate OpPermissionType = "create_export_task"
	// 审核/驳回数据导出工单；拥有该权限的用户可以审核/驳回数据导出工单
	OpPermissionTypeAuditExportWorkflow OpPermissionType = "audit_export_workflow"
	// 创建智能调优；拥有该权限的用户可以创建智能调优
	OpPermissionTypeCreateOptimization OpPermissionType = "create_optimization"
	// 查看他人创建的智能调优
	OpPermissionTypeViewOthersOptimization OpPermissionType = "view_others_optimization"
	// 配置流水线
	OpPermissionTypeCreatePipeline OpPermissionType = "create_pipeline"
	// SQL工作台;查看所有操作记录
	OpPermissionViewOperationRecord OpPermissionType = "view_operation_record"
	// 数据导出;查看所有导出任务
	OpPermissionViewExportTask OpPermissionType = "view_export_task"
	// 快捷审核;查看所有快捷审核记录
	OpPermissionViewQuickAuditRecord OpPermissionType = "view_quick_audit_record"
	// IDE审核;查看所有IDE审核记录
	OpPermissionViewIDEAuditRecord OpPermissionType = "view_ide_audit_record"
	// SQL优化;查看所有优化记录
	OpPermissionViewOptimizationRecord OpPermissionType = "view_optimization_record"
	// 版本管理;查看他人创建的版本记录
	OpPermissionViewVersionManage OpPermissionType = "view_version_manage"
	// 版本管理;配置版本
	OpPermissionVersionManage OpPermissionType = "version_manage"
	// CI/CD集成;查看所有流水线
	OpPermissionViewPipeline OpPermissionType = "view_pipeline"
	// 数据源管理;管理项目数据源管理
	OpPermissionManageProjectDataSource OpPermissionType = "manage_project_data_source"
	// 审核规则模版;管理审核规则模版
	OpPermissionManageAuditRuleTemplate OpPermissionType = "manage_audit_rule_template"
	// 审批流程模版;管理审批流程模版
	OpPermissionManageApprovalTemplate OpPermissionType = "manage_approval_template"
	// 成员与权限;管理成员与权限
	OpPermissionManageMember OpPermissionType = "manage_member"
	// 推送规则;管理推送规则
	OpPermissionPushRule OpPermissionType = "manage_push_rule"
	// 审核SQL例外;管理审核SQL例外
	OpPermissionMangeAuditSQLWhiteList OpPermissionType = "manage_audit_sql_white_list"
	// 管控SQL例外;管理管控SQL例外
	OpPermissionManageSQLMangeWhiteList OpPermissionType = "manage_sql_mange_white_list"
	// 角色管理;角色管理权限
	OpPermissionManageRoleMange OpPermissionType = "manage_role_mange"
	// 脱敏规则;脱敏规则配置权限
	OpPermissionDesensitization OpPermissionType = "desensitization"
	// 无任何权限
	OpPermissionTypeNone OpPermissionType = "none"
)

func ParseOpPermissionType(typ string) (OpPermissionType, error) {
	switch typ {
	case string(OpPermissionTypeGlobalView):
		return OpPermissionTypeGlobalView, nil
	case string(OpPermissionTypeGlobalManagement):
		return OpPermissionTypeGlobalManagement, nil
	case string(OpPermissionTypeCreateProject):
		return OpPermissionTypeCreateProject, nil
	case string(OpPermissionTypeProjectAdmin):
		return OpPermissionTypeProjectAdmin, nil
	case string(OpPermissionTypeCreateWorkflow):
		return OpPermissionTypeCreateWorkflow, nil
	case string(OpPermissionTypeAuditWorkflow):
		return OpPermissionTypeAuditWorkflow, nil
	case string(OpPermissionTypeAuthDBServiceData):
		return OpPermissionTypeAuthDBServiceData, nil
	case string(OpPermissionTypeViewOthersWorkflow):
		return OpPermissionTypeViewOthersWorkflow, nil
	case string(OpPermissionTypeExecuteWorkflow):
		return OpPermissionTypeExecuteWorkflow, nil
	case string(OpPermissionTypeViewOtherAuditPlan):
		return OpPermissionTypeViewOtherAuditPlan, nil
	case string(OpPermissionTypeSaveAuditPlan):
		return OpPermissionTypeSaveAuditPlan, nil
	case string(OpPermissionTypeSQLQuery):
		return OpPermissionTypeSQLQuery, nil
	case string(OpPermissionTypeExportCreate):
		return OpPermissionTypeExportCreate, nil
	case string(OpPermissionTypeAuditExportWorkflow):
		return OpPermissionTypeAuditExportWorkflow, nil
	case string(OpPermissionTypeCreateOptimization):
		return OpPermissionTypeCreateOptimization, nil
	case string(OpPermissionTypeViewOthersOptimization):
		return OpPermissionTypeViewOthersOptimization, nil
	case string(OpPermissionTypeCreatePipeline):
		return OpPermissionTypeCreatePipeline, nil
	case string(OpPermissionViewOperationRecord): return OpPermissionViewOperationRecord, nil
	case string(OpPermissionViewExportTask): return OpPermissionViewExportTask, nil
	case string(OpPermissionViewQuickAuditRecord): return OpPermissionViewQuickAuditRecord, nil
	case string(OpPermissionViewIDEAuditRecord): return OpPermissionViewIDEAuditRecord, nil
	case string(OpPermissionViewOptimizationRecord): return OpPermissionViewOptimizationRecord, nil
	case string(OpPermissionViewVersionManage): return OpPermissionViewVersionManage, nil
	case string(OpPermissionVersionManage): return OpPermissionVersionManage, nil
	case string(OpPermissionViewPipeline): return OpPermissionViewPipeline, nil
	case string(OpPermissionManageProjectDataSource): return OpPermissionManageProjectDataSource, nil
	case string(OpPermissionManageAuditRuleTemplate): return OpPermissionManageAuditRuleTemplate, nil
	case string(OpPermissionManageApprovalTemplate): return OpPermissionManageApprovalTemplate, nil
	case string(OpPermissionManageMember): return OpPermissionManageMember, nil
	case string(OpPermissionPushRule): return OpPermissionPushRule, nil
	case string(OpPermissionMangeAuditSQLWhiteList): return OpPermissionMangeAuditSQLWhiteList, nil
	case string(OpPermissionManageSQLMangeWhiteList): return OpPermissionManageSQLMangeWhiteList, nil
	case string(OpPermissionManageRoleMange): return OpPermissionManageRoleMange, nil
	case string(OpPermissionDesensitization): return OpPermissionDesensitization, nil
	case string(OpPermissionTypeNone): return OpPermissionTypeNone, nil
	default:
		return "", fmt.Errorf("invalid op permission type: %s", typ)
	}
}

func GetOperationTypeDesc(opType OpPermissionType) string {
	switch opType {
	case OpPermissionTypeUnknown:
		return "未知操作类型"
	case OpPermissionTypeCreateProject:
		return "创建项目"
	case OpPermissionTypeProjectAdmin:
		return "项目管理"
	case OpPermissionTypeCreateWorkflow:
		return "创建/编辑工单"
	case OpPermissionTypeAuditWorkflow:
		return "审核/驳回工单；拥有该权限的用户可以审核/驳回工单"
	case OpPermissionTypeAuthDBServiceData:
		return "授权数据源数据权限"
	case OpPermissionTypeViewOthersWorkflow:
		return "查看其他工单权限"
	case OpPermissionTypeExecuteWorkflow:
		return "上线工单"
	case OpPermissionTypeViewOtherAuditPlan:
		return "查看其他扫描任务权限"
	case OpPermissionTypeSaveAuditPlan:
		return "创建扫描任务权限"
	case OpPermissionTypeSQLQuery:
		return "SQL工作台查询"
	case OpPermissionTypeCreateOptimization:
		return "创建智能调优"
	case OpPermissionTypeViewOthersOptimization:
		return "查看他人创建的智能调优"
	case OpPermissionTypeCreatePipeline:
		return "配置流水线"
	default:
		return "未知操作类型"
	}
}

// swagger:parameters ListUsers
type ListUserReq struct {
	// the maximum count of user to be returned
	// in:query
	// Required: true
	PageSize uint32 `query:"page_size" json:"page_size" validate:"required"`
	// the offset of users to be returned, default is 0
	// in:query
	PageIndex uint32 `query:"page_index" json:"page_index"`
	// Multiple of ["name"], default is ["name"]
	// in:query
	OrderBy UserOrderByField `query:"order_by" json:"order_by"`
	// filter the user name
	// in:query
	FilterByName string `query:"filter_by_name" json:"filter_by_name"`
	// filter the user uids
	// in:query
	FilterByUids string `query:"filter_by_uids" json:"filter_by_uids"`
	// filter deleted user to be return ,default is false
	// in:query
	FilterDeletedUser bool `query:"filter_del_user" json:"filter_del_user"`
}

// swagger:enum UserOrderByField
type UserOrderByField string

const (
	UserOrderByName UserOrderByField = "name"
)

// A dms user
type ListUser struct {
	// user uid
	UserUid string `json:"uid"`
	// user name
	Name string `json:"name"`
	// user email
	Email string `json:"email"`
	// user phone
	Phone string `json:"phone"`
	// user wxid
	WxID string `json:"wxid"`
	// user stat
	Stat Stat `json:"stat"`
	// user authentication type
	AuthenticationType UserAuthenticationType `json:"authentication_type"`
	// user groups
	UserGroups []UidWithName `json:"user_groups"`
	// user operation permissions
	OpPermissions []UidWithName `json:"op_permissions"`
	// projects
	Projects []string `json:"projects"`
	// user is deleted
	IsDeleted bool `json:"is_deleted"`
	// third party user info
	ThirdPartyUserInfo string `json:"third_party_user_info"`
}

// swagger:model ListUserReply
type ListUserReply struct {
	// List user reply
	Data  []*ListUser `json:"data"`
	Total int64       `json:"total_nums"`

	// Generic reply
	base.GenericResp
}

// swagger:model
type GenAccessToken struct {
	ExpirationDays string `param:"expiration_days" json:"expiration_days" validate:"required"`
}

// swagger:model GenAccessTokenReply
type GenAccessTokenReply struct {
	// Get user reply
	Data *AccessTokenInfo `json:"data"`

	// Generic reply
	base.GenericResp
}
