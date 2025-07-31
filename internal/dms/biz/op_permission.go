package biz

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/actiontech/dms/api/dms/service/v1"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OpPermission struct {
	Base

	UID       string
	Name      string
	Module    Module
	RangeType OpRangeType
	Desc      string
	Service   v1.Service
}

func (o *OpPermission) GetUID() string {
	return o.UID
}

type OpRangeType string

func (o OpRangeType) String() string {
	return string(o)
}

const (
	OpRangeTypeGlobal    OpRangeType = "global"
	OpRangeTypeProject   OpRangeType = "project"
	OpRangeTypeDBService OpRangeType = "db_service"
)

type Module string

const (
	SQLWorkflow          Module = "SQL工单"
	SQLManage                   = "SQL管控"
	SQLDataSource               = "数据源管理"
	SQLWorkBench                = "SQL工作台"
	DataExport                  = "数据导出"
	QuickAudit                  = "快捷审核"
	VersionManage               = "版本管理"
	CICDIntegration             = "CI/CD集成"
	IDEAudit                    = "IDE审核"
	SQLOptimization             = "SQL优化"
	AuditRuleTemplate           = "审核规则模板"
	ApprovalFlowTemplate        = "审批流模板管理"
	MemberMange                 = "成员与权限"
	PushRule                    = "推送规则"
	AuditSQLWhiteList           = "审核SQL例外"
	SQLMangeWhiteList           = "管控SQL例外"
	RoleMange                   = "角色管理"
	DesensitizationRule         = "脱敏规则"
	AccountManagement           = "账号管理"
)

func ParseOpRangeType(t string) (OpRangeType, error) {
	switch t {
	case OpRangeTypeGlobal.String():
		return OpRangeTypeGlobal, nil
	case OpRangeTypeProject.String():
		return OpRangeTypeProject, nil
	case OpRangeTypeDBService.String():
		return OpRangeTypeDBService, nil
	default:
		return "", nil
	}
}

func initOpPermission() []*OpPermission {
	return []*OpPermission{
		{
			UID:       pkgConst.UIDOfOpPermissionGlobalView,
			Name:      "审计管理员",
			RangeType: OpRangeTypeGlobal,
			Desc:      "负责系统操作审计、数据合规检查等工作",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionGlobalManagement,
			Name:      "系统管理员",
			RangeType: OpRangeTypeGlobal,
			Desc:      "具备系统最高权限，可进行系统配置、用户管理等操作",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionCreateProject,
			Name:      "项目总监", // todo i18n 返回时会根据uid国际化，name、desc已弃用；数据库name字段是唯一键，故暂时保留
			RangeType: OpRangeTypeGlobal,
			Desc:      "创建项目、配置项目资源",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOrdinaryUser,
			Name:      "普通用户", // todo i18n 返回时会根据uid国际化，name、desc已弃用；数据库name字段是唯一键，故暂时保留
			RangeType: OpRangeTypeGlobal,
			Desc:      "基础功能操作权限，可进行日常业务操作",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionProjectAdmin,
			Name:      "项目管理",
			RangeType: OpRangeTypeProject,
			Desc:      "项目管理；拥有该权限的用户可以管理项目下的所有资源",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionCreateWorkflow,
			Name:      "创建上线工单",
			RangeType: OpRangeTypeDBService,
			Module:    SQLWorkflow,
			Desc:      "创建/编辑工单；拥有该权限的用户可以创建/编辑工单",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionAuditWorkflow,
			Name:      "审批上线工单",
			RangeType: OpRangeTypeDBService,
			Module:    SQLWorkflow,
			Desc:      "审核/驳回工单；拥有该权限的用户可以审核/驳回工单",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionExecuteWorkflow,
			Name:      "执行上线工单",
			RangeType: OpRangeTypeDBService,
			Module:    SQLWorkflow,
			Desc:      "上线工单；拥有该权限的用户可以上线工单",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOthersWorkflow,
			Name:      "查看所有工单",
			RangeType: OpRangeTypeDBService,
			Module:    SQLWorkflow,
			Desc:      "查看他人创建的工单；拥有该权限的用户可以查看他人创建的工单",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionSaveAuditPlan,
			Name:      "配置SQL管控",
			RangeType: OpRangeTypeDBService,
			Module:    SQLManage,
			Desc:      "创建/编辑扫描任务；拥有该权限的用户可以创建/编辑扫描任务",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOthersAuditPlan,
			Name:      "访问所有管控SQL",
			RangeType: OpRangeTypeDBService,
			Module:    SQLManage,
			Desc:      "查看他人创建的扫描任务；拥有该权限的用户可以查看他人创建的扫描任务",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewSQLInsight,
			Name:      "查看性能洞察",
			RangeType: OpRangeTypeDBService,
			Module:    SQLManage,
			Desc:      "查看性能洞察；拥有该权限的用户可以查看性能洞察的数据",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionSQLQuery,
			Name:      "SQL工作台操作权限",
			RangeType: OpRangeTypeDBService,
			Module:    SQLWorkBench,
			Desc:      "SQL查询；拥有该权限的用户可以执行SQL查询",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionExportApprovalReject,
			Name:      "审批导出工单",
			RangeType: OpRangeTypeDBService,
			Module:    DataExport,
			Desc:      "审批/驳回数据导出工单；拥有该权限的用户可以执行审批导出数据工单或者驳回导出数据工单",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionExportCreate,
			Name:      "创建导出工单",
			RangeType: OpRangeTypeDBService,
			Module:    DataExport,
			Desc:      "创建数据导出任务；拥有该权限的用户可以创建数据导出任务或者工单",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionCreateOptimization,
			Name:      "创建智能调优",
			RangeType: OpRangeTypeDBService,
			Module:    SQLOptimization,
			Desc:      "创建智能调优；拥有该权限的用户可以创建智能调优",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOthersOptimization,
			Name:      "查看他人创建的智能调优",
			RangeType: OpRangeTypeDBService,
			Module:    SQLOptimization,
			Desc:      "查看他人创建的智能调优；拥有该权限的用户可以查看他人创建的智能调优",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionCreatePipeline,
			Name:      "流水线增删改",
			RangeType: OpRangeTypeDBService,
			Module:    CICDIntegration,
			Desc:      "配置流水线；拥有该权限的用户可以为数据源配置流水线",
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOperationRecord,
			Name:      "查看所有操作记录",
			RangeType: OpRangeTypeDBService,
			Module:    SQLWorkBench,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewExportTask,
			Name:      "查看所有导出任务",
			RangeType: OpRangeTypeDBService,
			Module:    DataExport,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfPermissionViewQuickAuditRecord,
			Name:      "查看所有快捷审核记录",
			RangeType: OpRangeTypeDBService,
			Module:    QuickAudit,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewIDEAuditRecord,
			Name:      "查看所有IDE审核记录",
			RangeType: OpRangeTypeDBService,
			Module:    IDEAudit,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewVersionManage,
			Name:      "查看他人创建的版本记录",
			RangeType: OpRangeTypeDBService,
			Module:    VersionManage,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIDOfOpPermissionVersionManage,
			Name:      "配置版本",
			RangeType: OpRangeTypeDBService,
			Module:    VersionManage,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionViewPipeline,
			Name:      "查看所有流水线",
			RangeType: OpRangeTypeDBService,
			Module:    CICDIntegration,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionManageProjectDataSource,
			Name:      "管理项目数据源",
			RangeType: OpRangeTypeProject,
			Module:    SQLDataSource,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionManageAuditRuleTemplate,
			Name:      "管理审核规则模版",
			RangeType: OpRangeTypeProject,
			Module:    AuditRuleTemplate,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionManageApprovalTemplate,
			Name:      "管理审批流程模版",
			RangeType: OpRangeTypeProject,
			Module:    ApprovalFlowTemplate,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionManageMember,
			Name:      "管理成员与权限",
			RangeType: OpRangeTypeProject,
			Module:    MemberMange,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionPushRule,
			Name:      "管理推送规则",
			RangeType: OpRangeTypeProject,
			Module:    PushRule,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionMangeAuditSQLWhiteList,
			Name:      "审核SQL例外",
			RangeType: OpRangeTypeProject,
			Module:    AuditSQLWhiteList,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionManageSQLMangeWhiteList,
			Name:      "管控SQL例外",
			RangeType: OpRangeTypeProject,
			Module:    SQLMangeWhiteList,
			Service:   v1.ServiceSQLE,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionManageRoleMange,
			Name:      "角色管理权限",
			RangeType: OpRangeTypeProject,
			Module:    RoleMange,
			Service:   v1.ServiceDMS,
		},
		{
			UID:       pkgConst.UIdOfOpPermissionDesensitization,
			Name:      "脱敏规则配置权限",
			RangeType: OpRangeTypeProject,
			Module:    DesensitizationRule,
			Service:   v1.ServiceDMS,
		},
	}
}

func GetProxyOpPermission() map[string][]*OpPermission {
	return map[string][]*OpPermission{
		"provision": {
			{
				UID:       pkgConst.UIDOfOpPermissionAuthDBServiceData,
				Name:      "账号管理",
				RangeType: OpRangeTypeProject,
				Desc:      "账号管理；拥有该权限的用户可以授权数据源数据权限",
				Service:   v1.ServiceDMS,
				Module:    AccountManagement,
			},
		},
	}

}

type ListOpPermissionsOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      OpPermissionField
	FilterBy     []pkgConst.FilterCondition
}

type OpPermissionRepo interface {
	SaveOpPermission(ctx context.Context, op *OpPermission) error
	UpdateOpPermission(ctx context.Context, op *OpPermission) error
	CheckOpPermissionExist(ctx context.Context, opUids []string) (exists bool, err error)
	ListOpPermissions(ctx context.Context, opt *ListOpPermissionsOption) (ops []*OpPermission, total int64, err error)
	DelOpPermission(ctx context.Context, OpPermissionUid string) error
	GetOpPermission(ctx context.Context, OpPermissionUid string) (*OpPermission, error)
}

type OpPermissionUsecase struct {
	tx            TransactionGenerator
	repo          OpPermissionRepo
	pluginUsecase *PluginUsecase
	log           *utilLog.Helper
}

func NewOpPermissionUsecase(log utilLog.Logger, tx TransactionGenerator, repo OpPermissionRepo, pluginUsecase *PluginUsecase) *OpPermissionUsecase {
	return &OpPermissionUsecase{
		tx:            tx,
		repo:          repo,
		pluginUsecase: pluginUsecase,
		log:           utilLog.NewHelper(log, utilLog.WithMessageKey("biz.op_permission")),
	}
}

func (d *OpPermissionUsecase) InitOpPermissions(ctx context.Context, opPermissions []*OpPermission) (err error) {
	for _, opPermission := range opPermissions {

		_, err := d.repo.GetOpPermission(ctx, opPermission.GetUID())
		// already exist
		if err == nil {
			continue
		}

		// error, return directly
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return fmt.Errorf("failed to get op permission: %w", err)
		}

		// not exist, then create it
		if err := d.repo.SaveOpPermission(ctx, opPermission); err != nil {
			return fmt.Errorf("failed to init op permission: %w", err)
		}

	}
	d.log.Debug("update op permissions success")
	return nil
}

func (d *OpPermissionUsecase) IsGlobalOpPermissions(ctx context.Context, opUids []string) (bool, error) {
	for _, opUid := range opUids {
		op, err := d.repo.GetOpPermission(ctx, opUid)
		if err != nil {
			return false, err
		}
		if op.RangeType != OpRangeTypeGlobal {
			return false, nil
		}
	}
	return true, nil
}

func (d *OpPermissionUsecase) CheckOpPermissionExist(ctx context.Context, opUids []string) (exists bool, err error) {
	return d.repo.CheckOpPermissionExist(ctx, opUids)
}

func (d *OpPermissionUsecase) ListOpPermissions(ctx context.Context, opt *ListOpPermissionsOption) (ops []*OpPermission, total int64, err error) {

	ops, total, err = d.repo.ListOpPermissions(ctx, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list op permissions: %v", err)
	}
	return ops, total, nil
}

func (d *OpPermissionUsecase) ListUserOpPermissions(ctx context.Context, opt *ListOpPermissionsOption) (ops []*OpPermission, total int64, err error) {
	// 用户只能被赋予全局权限
	opt.FilterBy = append(opt.FilterBy, pkgConst.FilterCondition{
		Field:    string(OpPermissionFieldRangeType),
		Operator: pkgConst.FilterOperatorEqual,
		Value:    OpRangeTypeGlobal,
	})

	ops, total, err = d.repo.ListOpPermissions(ctx, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user op permissions: %v", err)
	}
	return ops, total, nil
}

func (d *OpPermissionUsecase) ListMemberOpPermissions(ctx context.Context, opt *ListOpPermissionsOption) (ops []*OpPermission, total int64, err error) {
	// 成员属于项目，只能被赋予非全局权限
	opt.FilterBy = append(opt.FilterBy, pkgConst.FilterCondition{
		Field:    string(OpPermissionFieldRangeType),
		Operator: pkgConst.FilterOperatorNotEqual,
		Value:    OpRangeTypeGlobal,
	})

	// 设置成员权限时，有单独的“项目管理权限”选项代表项目权限，所以这里不返回项目权限
	opt.FilterBy = append(opt.FilterBy, pkgConst.FilterCondition{
		Field:    string(OpPermissionFieldRangeType),
		Operator: pkgConst.FilterOperatorNotEqual,
		Value:    OpRangeTypeProject,
	})

	ops, total, err = d.repo.ListOpPermissions(ctx, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list member op permissions: %v", err)
	}

	return ops, total, nil
}

func (d *OpPermissionUsecase) ListProjectOpPermissions(ctx context.Context, opt *ListOpPermissionsOption) (ops []*OpPermission, total int64, err error) {
	opt.FilterBy = append(opt.FilterBy, pkgConst.FilterCondition{
		Field:    string(OpPermissionFieldRangeType),
		Operator: pkgConst.FilterOperatorEqual,
		Value:    OpRangeTypeProject,
	})

	opt.FilterBy = append(opt.FilterBy, pkgConst.FilterCondition{
		Field:    string(OpPermissionFieldUID),
		Operator: pkgConst.FilterOperatorNotEqual,
		Value:    pkgConst.UIDOfOpPermissionProjectAdmin,
	})

	ops, total, err = d.repo.ListOpPermissions(ctx, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user op permissions: %v", err)
	}
	return ops, total, nil
}
