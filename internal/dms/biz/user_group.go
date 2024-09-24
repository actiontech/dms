package biz

import (
	"context"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type UserGroup struct {
	Base

	UID  string
	Name string
	Desc string
	Stat UserGroupStat
}

type UserGroupStat uint

const (
	UserGroupStatOK      UserGroupStat = iota // 0
	UserGroupStatDisable                      // 1

	userGroupStatMax
)

func (u *UserGroupStat) Uint() uint {
	return uint(*u)
}

func ParseUserGroupStat(stat uint) (UserGroupStat, error) {
	if stat < uint(userGroupStatMax) {
		return UserGroupStat(stat), nil
	}
	return 0, fmt.Errorf("invalid user group stat: %d", stat)
}

func newUserGroup(name, desc string) (*UserGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &UserGroup{
		UID:  uid,
		Name: name,
		Desc: desc,
		Stat: UserGroupStatOK,
	}, nil
}

func (u *UserGroup) GetUID() string {
	return u.UID
}

type ListUserGroupsOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      UserGroupField
	FilterBy     []pkgConst.FilterCondition
}

type UserGroupRepo interface {
	SaveUserGroup(ctx context.Context, user *UserGroup) error
	UpdateUserGroup(ctx context.Context, u *UserGroup) error
	CheckUserGroupExist(ctx context.Context, userGroupUids []string) (exists bool, err error)
	ListUserGroups(ctx context.Context, opt *ListUserGroupsOption) (services []*UserGroup, total int64, err error)
	DelUserGroup(ctx context.Context, UserGroupUid string) error
	GetUserGroup(ctx context.Context, UserGroupUid string) (*UserGroup, error)
	AddUserToUserGroup(ctx context.Context, userGroupUid string, userUid string) error
	ReplaceUsersInUserGroup(ctx context.Context, userGroupUid string, userUids []string) error
	GetUsersInUserGroup(ctx context.Context, userGroupUid string) ([]*User, error)
}

type UserGroupUsecase struct {
	tx                        TransactionGenerator
	repo                      UserGroupRepo
	userRepo                  UserRepo
	pluginUsecase             *PluginUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	log                       *utilLog.Helper
}

func NewUserGroupUsecase(log utilLog.Logger, tx TransactionGenerator, repo UserGroupRepo, userRepo UserRepo, pluginUsecase *PluginUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase) *UserGroupUsecase {
	return &UserGroupUsecase{
		tx:                        tx,
		repo:                      repo,
		userRepo:                  userRepo,
		pluginUsecase:             pluginUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.usergroup")),
	}
}

type CreateUserGroupArgs struct {
	Name     string
	Desc     string
	UserUids []string
}

func (d *UserGroupUsecase) CreateUserGroup(ctx context.Context, currentUserUid string, args *CreateUserGroupArgs) (uid string, err error) {
	// check
	{
		if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
			return "", fmt.Errorf("check user is admin or global management permission : %v", err)
		} else if !canGlobalOp {
			return "", fmt.Errorf("user is not admin or global management permission")
		}
	}

	u, err := newUserGroup(args.Name, args.Desc)
	if err != nil {
		return "", fmt.Errorf("new user group failed: %v", err)
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.repo.SaveUserGroup(tx, u); err != nil {
		return "", fmt.Errorf("save user group failed: %v", err)
	}

	if err := d.InsureUsersToUserGroup(tx, args.UserUids, u.UID); err != nil {
		return "", fmt.Errorf("insure users to user group failed: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return "", fmt.Errorf("commit tx failed: %v", err)
	}

	return u.UID, nil
}

// InsureUsersToUserGroup 确保多个用户属于指定的用户组
func (d *UserGroupUsecase) InsureUsersToUserGroup(ctx context.Context, userUids []string, userGroupUid string) (err error) {
	// TODO: check user exist
	if err := d.repo.ReplaceUsersInUserGroup(ctx, userGroupUid, userUids); err != nil {
		return fmt.Errorf("insure users to user group failed: %v", err)
	}

	return nil
}

func (d *UserGroupUsecase) ListUserGroup(ctx context.Context, option *ListUserGroupsOption) (users []*UserGroup, total int64, err error) {
	users, total, err = d.repo.ListUserGroups(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list user groups failed: %v", err)
	}
	return users, total, nil
}

func (d *UserGroupUsecase) DelUserGroup(ctx context.Context, currentUserUid, UserGroupUid string) (err error) {

	// check
	{
		if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
			return fmt.Errorf("check user is admin or global management permission : %v", err)
		} else if !canGlobalOp {
			return fmt.Errorf("user is not admin or global management permission")
		}
	}

	ds, err := d.repo.GetUserGroup(ctx, UserGroupUid)
	if err != nil {
		return fmt.Errorf("get user group failed: %v", err)
	}

	// 调用其他服务对用户组进行预检查
	if err := d.pluginUsecase.DelUserGroupPreCheck(ctx, ds.GetUID()); err != nil {
		return fmt.Errorf("precheck del user group failed: %v", err)
	}

	if err := d.repo.DelUserGroup(ctx, UserGroupUid); nil != err {
		return fmt.Errorf("delete user group error: %v", err)
	}
	return nil
}

func (d *UserGroupUsecase) GetUsersInUserGroup(ctx context.Context, userGroupUid string) (users []*User, err error) {
	users, err = d.repo.GetUsersInUserGroup(ctx, userGroupUid)
	if err != nil {
		return nil, fmt.Errorf("get users in user group failed: %v", err)
	}
	return users, nil
}

func (d *UserGroupUsecase) GetUserGroup(ctx context.Context, userGroupUid string) (*UserGroup, error) {
	return d.repo.GetUserGroup(ctx, userGroupUid)
}

func (d *UserGroupUsecase) UpdateUserGroup(ctx context.Context, currentUserUid, updateUserGroupUid string, isDisabled bool, desc *string, userUids []string) error {
	// check
	{
		if canGlobalOp, err := d.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
			return fmt.Errorf("check user is admin or global management permission : %v", err)
		} else if !canGlobalOp {
			return fmt.Errorf("user is not admin or global management permission")
		}
	}

	group, err := d.GetUserGroup(ctx, updateUserGroupUid)
	if err != nil {
		return fmt.Errorf("get user group failed: %v", err)
	}

	if isDisabled {
		group.Stat = UserGroupStatDisable
	} else {
		group.Stat = UserGroupStatOK
	}

	if desc != nil {
		group.Desc = *desc
	}

	if exist, err := d.userRepo.CheckUserExist(ctx, userUids); err != nil {
		return fmt.Errorf("check user exist failed: %v", err)
	} else if !exist {
		return fmt.Errorf("user not exist")
	}

	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	if err := d.repo.ReplaceUsersInUserGroup(tx, group.UID, userUids); err != nil {
		return fmt.Errorf("replace users in user group failed: %v", err)
	}

	if err := d.repo.UpdateUserGroup(tx, group); nil != err {
		return fmt.Errorf("update user group error: %v", err)
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}
