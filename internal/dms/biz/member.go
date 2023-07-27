package biz

import (
	"context"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type Member struct {
	Base

	UID              string
	NamespaceUID     string
	UserUID          string
	RoleWithOpRanges []MemberRoleWithOpRange
}

func (u *Member) GetUID() string {
	return u.UID
}

type MemberRoleWithOpRange struct {
	RoleUID     string      // Role中包含了多个操作权限
	OpRangeType OpRangeType // OpRangeType描述操作权限的权限范围类型，目前只支持数据源
	RangeUIDs   []string    // Range描述操作权限的权限范围，如涉及哪些数据源
}

func newMember(userUid, namespaceUid string, opPermissions []MemberRoleWithOpRange) (*Member, error) {

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}

	return &Member{
		UID:              uid,
		NamespaceUID:     namespaceUid,
		UserUID:          userUid,
		RoleWithOpRanges: opPermissions,
	}, nil
}

type MemberRepo interface {
	SaveMember(ctx context.Context, u *Member) error
	ListMembers(ctx context.Context, opt *ListMembersOption) (members []*Member, total int64, err error)
	GetMember(ctx context.Context, memberUid string) (*Member, error)
	UpdateMember(ctx context.Context, m *Member) error
	CheckMemberExist(ctx context.Context, memberUids []string) (exists bool, err error)
	DelMember(ctx context.Context, memberUid string) error
	DelRoleFromAllMembers(ctx context.Context, roleUid string) error
}

type MemberUsecase struct {
	tx                        TransactionGenerator
	repo                      MemberRepo
	userUsecase               *UserUsecase
	roleUsecase               *RoleUsecase
	dbServiceUsecase          *DBServiceUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	namespaceUsecase          *NamespaceUsecase
	log                       *utilLog.Helper
}

func NewMemberUsecase(log utilLog.Logger, tx TransactionGenerator, repo MemberRepo,
	userUsecase *UserUsecase,
	roleUsecase *RoleUsecase,
	dbServiceUsecase *DBServiceUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase,
	namespaceUsecase *NamespaceUsecase) *MemberUsecase {
	return &MemberUsecase{
		tx:                        tx,
		repo:                      repo,
		userUsecase:               userUsecase,
		roleUsecase:               roleUsecase,
		dbServiceUsecase:          dbServiceUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		namespaceUsecase:          namespaceUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.member")),
	}
}

func (m *MemberUsecase) CreateMember(ctx context.Context, currentUserUid string, memberUserUid string, namespaceUid string, isNamespaceAdmin bool,
	roleAndOpRanges []MemberRoleWithOpRange) (memberUid string, err error) {
	// check
	{
		// 检查空间是否归档/删除
		if err := m.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
			return "", fmt.Errorf("create member error: %v", err)
		}
		// 检查当前用户有空间管理员权限
		if isAdmin, err := m.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, namespaceUid); err != nil {
			return "", fmt.Errorf("check user is namespace admin failed: %v", err)
		} else if !isAdmin {
			return "", fmt.Errorf("user is not namespace admin")
		}

		// 检查成员用户存在
		if exist, err := m.userUsecase.CheckUserExist(ctx, []string{memberUserUid}); err != nil {
			return "", fmt.Errorf("check user exist failed: %v", err)
		} else if !exist {
			return "", fmt.Errorf("user not exist")
		}

		if err := m.CheckRoleAndOpRanges(ctx, roleAndOpRanges); err != nil {
			return "", err
		}

		// 检查空间内成员之间的用户不同
		if _, total, err := m.ListMember(ctx, &ListMembersOption{
			PageNumber:   0,
			LimitPerPage: 10,
			FilterBy: []pkgConst.FilterCondition{
				{
					Field:    string(MemberFieldUserUID),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    memberUserUid,
				},
				{
					Field:    string(MemberFieldNamespaceUID),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    namespaceUid,
				},
			},
		}, namespaceUid); err != nil {
			return "", fmt.Errorf("check member exist failed: %v", err)
		} else if total > 0 {
			return "", fmt.Errorf("user already in namespace")
		}
	}

	member, err := newMember(memberUserUid, namespaceUid, roleAndOpRanges)
	if err != nil {
		return "", fmt.Errorf("new member failed: %v", err)
	}

	// 如果是空间管理员，则自动添加内置的空间管理员角色
	if isNamespaceAdmin {
		m.FixMemberWithNamespaceAdmin(ctx, member, namespaceUid)
	}

	tx := m.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(m.log, err)
		}
	}()

	if err := m.repo.SaveMember(tx, member); err != nil {
		return "", fmt.Errorf("save member failed: %v", err)
	}

	if err := tx.Commit(m.log); err != nil {
		return "", fmt.Errorf("commit tx failed: %v", err)
	}
	return member.GetUID(), nil

}

// AddUserToNamespaceAdmin 将指定用户加入空间成员，并赋予空间管理员权限
func (m *MemberUsecase) AddUserToNamespaceAdminMember(ctx context.Context, userUid string, namespaceUid string) (memberUid string, err error) {
	// check
	{
		// 检查空间是否归档/删除
		if err := m.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
			return "", fmt.Errorf("add user to namespace admin member error: %v", err)
		}
		// 如果已存在则报错
		if _, total, err := m.ListMember(ctx, &ListMembersOption{
			PageNumber:   0,
			LimitPerPage: 10,
			FilterBy: []pkgConst.FilterCondition{
				{
					Field:    string(MemberFieldUserUID),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    userUid,
				},
				{
					Field:    string(MemberFieldNamespaceUID),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    namespaceUid,
				},
			},
		}, namespaceUid); err != nil {
			return "", fmt.Errorf("check member exist failed: %v", err)
		} else if total > 0 {
			return "", fmt.Errorf("user already in namespace")
		}
	}

	member, err := newMember(userUid, namespaceUid, []MemberRoleWithOpRange{m.GetNamespaceAdminRoleWithOpRange(namespaceUid)})
	if err != nil {
		return "", fmt.Errorf("new member failed: %v", err)
	}

	if err := m.repo.SaveMember(ctx, member); err != nil {
		return "", fmt.Errorf("save member failed: %v", err)
	}

	return member.GetUID(), nil

}

// CheckRoleAndOpRanges 检查角色和操作权限范围是否合法
func (m *MemberUsecase) CheckRoleAndOpRanges(ctx context.Context, roleAndOpRanges []MemberRoleWithOpRange) error {
	for _, r := range roleAndOpRanges {
		// 检查角色存在
		role, err := m.roleUsecase.GetRole(ctx, r.RoleUID)
		if err != nil {
			return fmt.Errorf("get role exist failed: %v", err)
		}

		// 获取角色的操作权限
		opPermissions, err := m.roleUsecase.GetOpPermissions(ctx, role.UID)
		if err != nil {
			return fmt.Errorf("get op permissions failed: %v", err)
		}
		for _, op := range opPermissions {
			// 检查操作权限与指定的范围类型匹配
			if op.RangeType != r.OpRangeType {
				return fmt.Errorf("range type not match, op permission range type: %v, role range type: %v", op.RangeType, r.OpRangeType)
			}
			switch op.RangeType {
			case OpRangeTypeDBService:
				// 检查数据源存在
				if exist, err := m.dbServiceUsecase.CheckDBServiceExist(ctx, r.RangeUIDs); err != nil {
					return fmt.Errorf("check db service exist failed: %v", err)
				} else if !exist {
					return fmt.Errorf("db service not exist")
				}
			// 角色目前与成员绑定，只支持配置数据源范围的权限
			case OpRangeTypeNamespace, OpRangeTypeGlobal:
				return fmt.Errorf("role currently only support the db service op range type, but got type: %v", op.RangeType)
			default:
				return fmt.Errorf("unsupported range type: %v", op.RangeType)
			}
		}
	}
	return nil
}

func (m *MemberUsecase) IsMemberNamespaceAdmin(ctx context.Context, memberUid string) (bool, error) {
	member, err := m.repo.GetMember(ctx, memberUid)
	if err != nil {
		return false, fmt.Errorf("get member failed: %v", err)
	}

	for _, r := range member.RoleWithOpRanges {
		if r.RoleUID == pkgConst.UIDOfRoleNamespaceAdmin {
			return true, nil
		}
	}
	return false, nil
}

// FixMemberWithNamespaceAdmin 自动修改成员的角色和操作权限范围，如果是空间管理员，则自动添加内置的空间管理员角色
func (m *MemberUsecase) FixMemberWithNamespaceAdmin(ctx context.Context, member *Member, namespaceUid string) {
	member.RoleWithOpRanges = append(member.RoleWithOpRanges, m.GetNamespaceAdminRoleWithOpRange(namespaceUid))
}

func (m *MemberUsecase) GetNamespaceAdminRoleWithOpRange(namespaceUid string) MemberRoleWithOpRange {
	return MemberRoleWithOpRange{
		RoleUID:     pkgConst.UIDOfRoleNamespaceAdmin,
		OpRangeType: OpRangeTypeNamespace,
		RangeUIDs:   []string{namespaceUid},
	}
}

func (m *MemberUsecase) GetMemberRoleWithOpRange(ctx context.Context, memberUid string) (roleWithOpRange []MemberRoleWithOpRange, err error) {
	member, err := m.repo.GetMember(ctx, memberUid)
	if err != nil {
		return nil, fmt.Errorf("get member failed: %v", err)
	}
	return member.RoleWithOpRanges, nil
}

type ListMembersOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      MemberField
	FilterBy     []pkgConst.FilterCondition
}

func (m *MemberUsecase) ListMember(ctx context.Context, option *ListMembersOption, namespaceUid string) (members []*Member, total int64, err error) {
	// 检查空间是否归档/删除
	if err := m.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
		return nil, 0, fmt.Errorf("list member error: %v", err)
	}
	members, total, err = m.repo.ListMembers(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list members failed: %v", err)
	}

	return members, total, nil
}

func (m *MemberUsecase) GetMember(ctx context.Context, memberUid string) (*Member, error) {
	return m.repo.GetMember(ctx, memberUid)
}

func (m *MemberUsecase) CheckMemberExist(ctx context.Context, memberUids []string) (bool, error) {
	return m.repo.CheckMemberExist(ctx, memberUids)
}

func (m *MemberUsecase) UpdateMember(ctx context.Context, currentUserUid, updateMemberUid, namespaceUid string, isNamespaceAdmin bool,
	roleAndOpRanges []MemberRoleWithOpRange) error {
	// check
	{
		// 检查空间是否归档/删除
		if err := m.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
			return fmt.Errorf("update member error: %v", err)
		}
		// 检查当前用户有空间管理员权限
		if isAdmin, err := m.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, namespaceUid); err != nil {
			return fmt.Errorf("check user is namespace admin failed: %v", err)
		} else if !isAdmin {
			return fmt.Errorf("user is not namespace admin")
		}

		if err := m.CheckRoleAndOpRanges(ctx, roleAndOpRanges); err != nil {
			return err
		}
	}

	member, err := m.GetMember(ctx, updateMemberUid)
	if err != nil {
		return fmt.Errorf("get member failed: %v", err)
	}
	member.RoleWithOpRanges = roleAndOpRanges

	// 如果是空间管理员，则自动添加内置的空间管理员角色
	if isNamespaceAdmin {
		m.FixMemberWithNamespaceAdmin(ctx, member, namespaceUid)
	}

	tx := m.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(m.log, err)
		}
	}()

	if err := m.repo.UpdateMember(tx, member); nil != err {
		return fmt.Errorf("update member error: %v", err)
	}

	if err := tx.Commit(m.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	return nil
}

func (m *MemberUsecase) DelMember(ctx context.Context, currentUserUid, memberUid string) (err error) {
	// check
	{
		member, err := m.GetMember(ctx, memberUid)
		if err != nil {
			return fmt.Errorf("get member failed: %v", err)
		}
		// 检查空间是否归档/删除
		if err := m.namespaceUsecase.isNamespaceActive(ctx, member.NamespaceUID); err != nil {
			return fmt.Errorf("delete member error: %v", err)
		}

		// 检查当前用户有空间管理员权限
		if isAdmin, err := m.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, member.NamespaceUID); err != nil {
			return fmt.Errorf("check user is namespace admin failed: %v", err)
		} else if !isAdmin {
			return fmt.Errorf("user is not namespace admin")
		}
	}

	tx := m.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(m.log, err)
		}
	}()

	if err := m.repo.DelMember(tx, memberUid); nil != err {
		return fmt.Errorf("delete member error: %v", err)
	}

	if err := tx.Commit(m.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}
