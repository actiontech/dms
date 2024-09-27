package biz

import (
	"context"
	"errors"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type RoleStat uint

const (
	RoleStatOK      RoleStat = iota // 0
	RoleStatDisable                 // 1

	roleStatMax
)

func (u *RoleStat) Uint() uint {
	return uint(*u)
}

func ParseRoleStat(stat uint) (RoleStat, error) {
	if stat < uint(roleStatMax) {
		return RoleStat(stat), nil
	}
	return 0, fmt.Errorf("invalid role stat: %d", stat)
}

type Role struct {
	Base

	UID  string
	Name string
	Desc string
	Stat RoleStat
}

func newRole(name, desc string) (*Role, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &Role{
		UID:  uid,
		Name: name,
		Desc: desc,
	}, nil
}

func (u *Role) GetUID() string {
	return u.UID
}

func initRole() []*Role {
	return []*Role{
		{
			UID:  pkgConst.UIDOfRoleProjectAdmin,
			Name: "项目管理员", // todo i18n 返回时会根据uid国际化，name、desc已弃用；数据库name字段是唯一键，故暂时保留
			Desc: "project admin",
		},
		{
			UID:  pkgConst.UIDOfRoleSQLEAdmin,
			Name: "SQLE管理员",
			Desc: "拥有该权限的用户可以创建/编辑工单，审核/驳回工单，上线工单,创建/编辑扫描任务",
		},
		{
			UID:  pkgConst.UIDOfRoleProvisionAdmin,
			Name: "provision管理员",
			Desc: "拥有该权限的用户可以授权数据源数据权限",
		},
	}
}

type ListRolesOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      RoleField
	FilterBy     []pkgConst.FilterCondition
}

type RoleRepo interface {
	SaveRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, u *Role) error
	CheckRoleExist(ctx context.Context, roleUids []string) (exists bool, err error)
	ListRoles(ctx context.Context, opt *ListRolesOption) (roles []*Role, total int64, err error)
	DelRole(ctx context.Context, roleUid string) error
	GetRole(ctx context.Context, roleUid string) (*Role, error)
	AddOpPermissionToRole(ctx context.Context, OpPermissionUid string, roleUid string) error
	ReplaceOpPermissionsInRole(ctx context.Context, roleUid string, OpPermissionUids []string) error
	GetOpPermissionsByRole(ctx context.Context, roleUid string) ([]*OpPermission, error)
	DelAllOpPermissionsFromRole(ctx context.Context, roleUid string) error
}

type RoleUsecase struct {
	tx                        TransactionGenerator
	repo                      RoleRepo
	opPermissionRepo          OpPermissionRepo
	memberRepo                MemberRepo
	pluginUsecase             *PluginUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	log                       *utilLog.Helper
}

func NewRoleUsecase(log utilLog.Logger, tx TransactionGenerator, repo RoleRepo, opPermissionRepo OpPermissionRepo, memberRepo MemberRepo, pluginUsecase *PluginUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase) *RoleUsecase {
	return &RoleUsecase{
		tx:                        tx,
		repo:                      repo,
		opPermissionRepo:          opPermissionRepo,
		memberRepo:                memberRepo,
		pluginUsecase:             pluginUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.role")),
	}
}

func (d *RoleUsecase) InitRoles(ctx context.Context) (err error) {
	for _, r := range initRole() {

		_, err = d.repo.GetRole(ctx, r.GetUID())
		// already exist
		if err == nil {
			continue
		}

		// error, return directly
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return fmt.Errorf("failed to get role: %v", err)
		}

		// not exist, then create it
		if err := d.repo.SaveRole(ctx, r); err != nil {
			return fmt.Errorf("failed to init role: %v", err)
		}

		roleId := r.GetUID()
		switch roleId {
		case pkgConst.UIDOfRoleProjectAdmin:
			if err = d.InsureOpPermissionsToRole(ctx, []string{pkgConst.UIDOfOpPermissionProjectAdmin}, roleId); err != nil {
				return fmt.Errorf("insure op permissions in role failed: %v", err)
			}
		case pkgConst.UIDOfRoleSQLEAdmin:
			if err = d.InsureOpPermissionsToRole(ctx, []string{pkgConst.UIDOfOpPermissionCreateWorkflow,
				pkgConst.UIDOfOpPermissionAuditWorkflow, pkgConst.UIDOfOpPermissionExecuteWorkflow,
				pkgConst.UIDOfOpPermissionViewOthersWorkflow, pkgConst.UIDOfOpPermissionSaveAuditPlan,
				pkgConst.UIDOfOpPermissionViewOthersAuditPlan, pkgConst.UIDOfOpPermissionSQLQuery,
				pkgConst.UIDOfOpPermissionExportApprovalReject, pkgConst.UIDOfOpPermissionExportCreate,
				pkgConst.UIDOfOpPermissionCreateOptimization, pkgConst.UIDOfOpPermissionViewOthersOptimization,
				pkgConst.UIDOfOpPermissionCreatePipeline}, roleId); err != nil {
				return fmt.Errorf("insure op permissions in role failed: %v", err)
			}
		case pkgConst.UIDOfRoleProvisionAdmin:
			if err = d.InsureOpPermissionsToRole(ctx, []string{pkgConst.UIDOfOpPermissionAuthDBServiceData}, roleId); err != nil {
				return fmt.Errorf("insure op permissions in role failed: %v", err)
			}
		default:
			return fmt.Errorf("invalid role uid: %s", r.GetUID())
		}

	}
	d.log.Debug("init roles success")
	return nil
}

func (d *RoleUsecase) CreateRole(ctx context.Context, currentUserUid, name, desc string, opPermissionUids []string) (uid string, err error) {
	// check
	{
		if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
			return "", fmt.Errorf("check user is admin or global management permission : %v", err)
		} else if !canGlobalOp {
			return "", fmt.Errorf("user is not admin or global management permission")
		}
	}

	u, err := newRole(name, desc)
	if err != nil {
		return "", fmt.Errorf("new role failed: %v", err)
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.repo.SaveRole(tx, u); err != nil {
		return "", fmt.Errorf("save role failed: %v", err)
	}

	if err := d.InsureOpPermissionsToRole(tx, opPermissionUids, u.UID); err != nil {
		return "", fmt.Errorf("insure op permissions in role failed: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return "", fmt.Errorf("commit tx failed: %v", err)
	}

	return u.UID, nil
}

// InsureOpPermissionsToRole 确保操作权限属于指定的角色
func (d *RoleUsecase) InsureOpPermissionsToRole(ctx context.Context, opPermissionUids []string, roleUid string) (err error) {

	if exist, err := d.opPermissionRepo.CheckOpPermissionExist(ctx, opPermissionUids); err != nil {
		return fmt.Errorf("check op permission exist failed: %v", err)
	} else if !exist {
		return fmt.Errorf("op permission not exist")
	}

	if err := d.repo.ReplaceOpPermissionsInRole(ctx, roleUid, opPermissionUids); err != nil {
		return fmt.Errorf("replace op permissions in role failed: %v", err)
	}

	return nil
}

func (d *RoleUsecase) ListRole(ctx context.Context, option *ListRolesOption) (roles []*Role, total int64, err error) {
	// DMS-125：项目管理员为内置角色，不对外展示
	option.FilterBy = append(option.FilterBy, pkgConst.FilterCondition{
		Field:    string(RoleFieldUID),
		Operator: pkgConst.FilterOperatorNotEqual,
		Value:    pkgConst.UIDOfRoleProjectAdmin,
	})
	roles, total, err = d.repo.ListRoles(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list roles failed: %v", err)
	}
	return roles, total, nil
}

func (d *RoleUsecase) DelRole(ctx context.Context, currentUserUid, roleUid string) (err error) {
	// check
	{
		if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
			return fmt.Errorf("check user is admin or global management permission : %v", err)
		} else if !canGlobalOp {
			return fmt.Errorf("user is not admin or global management permission")
		}
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.memberRepo.DelRoleFromAllMembers(tx, roleUid); nil != err {
		return fmt.Errorf("delete role from all members error: %v", err)
	}

	if err := d.repo.DelAllOpPermissionsFromRole(tx, roleUid); nil != err {
		return fmt.Errorf("delete all op permissions from role error: %v", err)
	}

	if err := d.repo.DelRole(tx, roleUid); nil != err {
		return fmt.Errorf("delete role error: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}

func (d *RoleUsecase) GetOpPermissions(ctx context.Context, roleUid string) (opPermissions []*OpPermission, err error) {
	ops, err := d.repo.GetOpPermissionsByRole(ctx, roleUid)
	if err != nil {
		return nil, fmt.Errorf("get op permissions by role failed: %v", err)
	}
	return ops, nil
}

func (d *RoleUsecase) GetRole(ctx context.Context, roleUid string) (*Role, error) {
	return d.repo.GetRole(ctx, roleUid)
}

func (d *RoleUsecase) CheckRoleExist(ctx context.Context, roleUids []string) (bool, error) {
	return d.repo.CheckRoleExist(ctx, roleUids)
}

func (d *RoleUsecase) UpdateRole(ctx context.Context, currentUserUid, updateRoleUid string, isDisabled bool, desc *string, opPermissionUids []string) error {
	// check
	{
		if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
			return fmt.Errorf("check user is admin or global management permission : %v", err)
		} else if !canGlobalOp {
			return fmt.Errorf("user is not admin or global management permission")
		}
	}

	role, err := d.GetRole(ctx, updateRoleUid)
	if err != nil {
		return fmt.Errorf("get role failed: %v", err)
	}

	if isDisabled {
		role.Stat = RoleStatDisable
	} else {
		role.Stat = RoleStatOK
	}

	if desc != nil {
		role.Desc = *desc
	}

	if exist, err := d.opPermissionRepo.CheckOpPermissionExist(ctx, opPermissionUids); err != nil {
		return fmt.Errorf("check op permissions exist failed: %v", err)
	} else if !exist {
		return fmt.Errorf("op permissions not exist")
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.repo.ReplaceOpPermissionsInRole(tx, role.UID, opPermissionUids); err != nil {
		return fmt.Errorf("replace op permissionss in role failed: %v", err)
	}

	if err := d.repo.UpdateRole(tx, role); nil != err {
		return fmt.Errorf("update role error: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	return nil
}
