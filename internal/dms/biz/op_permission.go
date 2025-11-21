package biz

import (
	"context"
	"errors"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OpPermission struct {
	Base

	UID       string
	Name      string
	RangeType OpRangeType
	Desc      string
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
			UID:       pkgConst.UIDOfOpPermissionCreateProject,
			Name:      "创建项目",
			RangeType: OpRangeTypeGlobal,
			Desc:      "创建项目；创建项目的用户自动拥有该项目管理权限",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionProjectAdmin,
			Name:      "项目管理",
			RangeType: OpRangeTypeProject,
			Desc:      "项目管理；拥有该权限的用户可以管理项目下的所有资源",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionCreateWorkflow,
			Name:      "创建/编辑工单",
			RangeType: OpRangeTypeDBService,
			Desc:      "创建/编辑工单；拥有该权限的用户可以创建/编辑工单",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionAuditWorkflow,
			Name:      "审核/驳回工单",
			RangeType: OpRangeTypeDBService,
			Desc:      "审核/驳回工单；拥有该权限的用户可以审核/驳回工单",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionAuthDBServiceData,
			Name:      "授权数据源数据权限",
			RangeType: OpRangeTypeDBService,
			Desc:      "授权数据源数据权限；拥有该权限的用户可以授权数据源数据权限",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionExecuteWorkflow,
			Name:      "上线工单",
			RangeType: OpRangeTypeDBService,
			Desc:      "上线工单；拥有该权限的用户可以上线工单",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOthersWorkflow,
			Name:      "查看他人创建的工单",
			RangeType: OpRangeTypeDBService,
			Desc:      "查看他人创建的工单；拥有该权限的用户可以查看他人创建的工单",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionSaveAuditPlan,
			Name:      "创建/编辑扫描任务",
			RangeType: OpRangeTypeDBService,
			Desc:      "创建/编辑扫描任务；拥有该权限的用户可以创建/编辑扫描任务",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOthersAuditPlan,
			Name:      "查看他人创建的扫描任务",
			RangeType: OpRangeTypeDBService,
			Desc:      "查看他人创建的扫描任务；拥有该权限的用户可以查看他人创建的扫描任务",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionSQLQuery,
			Name:      "SQL查询",
			RangeType: OpRangeTypeDBService,
			Desc:      "SQL查询；拥有该权限的用户可以执行SQL查询",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionExportApprovalReject,
			Name:      "审批/驳回数据导出工单",
			RangeType: OpRangeTypeDBService,
			Desc:      "审批/驳回数据导出工单；拥有该权限的用户可以执行审批导出数据工单或者驳回导出数据工单",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionExportCreate,
			Name:      "创建数据导出任务",
			RangeType: OpRangeTypeDBService,
			Desc:      "创建数据导出任务；拥有该权限的用户可以创建数据导出任务或者工单",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionCreateOptimization,
			Name:      "创建智能调优",
			RangeType: OpRangeTypeDBService,
			Desc:      "创建智能调优；拥有该权限的用户可以创建智能调优",
		},
		{
			UID:       pkgConst.UIDOfOpPermissionViewOthersOptimization,
			Name:      "查看他人创建的智能调优",
			RangeType: OpRangeTypeDBService,
			Desc:      "查看他人创建的智能调优；拥有该权限的用户可以查看他人创建的智能调优",
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

func (d *OpPermissionUsecase) InitOpPermissions(ctx context.Context) (err error) {
	for _, op := range initOpPermission() {

		_, err := d.repo.GetOpPermission(ctx, op.GetUID())
		// already exist
		if err == nil {
			continue
		}

		// error, return directly
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return fmt.Errorf("failed to get op permission: %w", err)
		}

		// not exist, then create it
		if err := d.repo.SaveOpPermission(ctx, op); err != nil {
			return fmt.Errorf("failed to init op permission: %w", err)
		}

	}
	d.log.Debug("init op permissions success")
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
