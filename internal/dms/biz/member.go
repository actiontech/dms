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
	ProjectUID       string
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

func newMember(userUid, projectUid string, opPermissions []MemberRoleWithOpRange) (*Member, error) {

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}

	return &Member{
		UID:              uid,
		ProjectUID:       projectUid,
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
	projectUsecase            *ProjectUsecase
	log                       *utilLog.Helper
}

func NewMemberUsecase(log utilLog.Logger, tx TransactionGenerator, repo MemberRepo,
	userUsecase *UserUsecase,
	roleUsecase *RoleUsecase,
	dbServiceUsecase *DBServiceUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase,
	projectUsecase *ProjectUsecase) *MemberUsecase {
	return &MemberUsecase{
		tx:                        tx,
		repo:                      repo,
		userUsecase:               userUsecase,
		roleUsecase:               roleUsecase,
		dbServiceUsecase:          dbServiceUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		projectUsecase:            projectUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.member")),
	}
}

func (m *MemberUsecase) CreateMember(ctx context.Context, currentUserUid string, memberUserUid string, projectUid string, isProjectAdmin bool,
	roleAndOpRanges []MemberRoleWithOpRange) (memberUid string, err error) {
	// check
	{
		// 检查项目是否归档/删除
		if err := m.projectUsecase.isProjectActive(ctx, projectUid); err != nil {
			return "", fmt.Errorf("create member error: %v", err)
		}
		// 检查当前用户有项目管理员权限
		if isAdmin, err := m.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, projectUid); err != nil {
			return "", fmt.Errorf("check user is project admin failed: %v", err)
		} else if !isAdmin {
			return "", fmt.Errorf("user is not project admin")
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

		// 检查项目内成员之间的用户不同
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
					Field:    string(MemberFieldProjectUID),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    projectUid,
				},
			},
		}, projectUid); err != nil {
			return "", fmt.Errorf("check member exist failed: %v", err)
		} else if total > 0 {
			return "", fmt.Errorf("user already in project")
		}
	}

	member, err := newMember(memberUserUid, projectUid, roleAndOpRanges)
	if err != nil {
		return "", fmt.Errorf("new member failed: %v", err)
	}

	// 如果是项目管理员，则自动添加内置的项目管理员角色
	if isProjectAdmin {
		m.FixMemberWithProjectAdmin(ctx, member, projectUid)
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

// AddUserToProjectAdmin 将指定用户加入项目成员，并赋予项目管理员权限
func (m *MemberUsecase) AddUserToProjectAdminMember(ctx context.Context, userUid string, projectUid string) (memberUid string, err error) {
	// check
	{
		// 检查项目是否归档/删除
		if err := m.projectUsecase.isProjectActive(ctx, projectUid); err != nil {
			return "", fmt.Errorf("add user to project admin member error: %v", err)
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
					Field:    string(MemberFieldProjectUID),
					Operator: pkgConst.FilterOperatorEqual,
					Value:    projectUid,
				},
			},
		}, projectUid); err != nil {
			return "", fmt.Errorf("check member exist failed: %v", err)
		} else if total > 0 {
			return "", fmt.Errorf("user already in project")
		}
	}

	member, err := newMember(userUid, projectUid, []MemberRoleWithOpRange{m.GetProjectAdminRoleWithOpRange(projectUid)})
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
			case OpRangeTypeProject, OpRangeTypeGlobal:
				return fmt.Errorf("role currently only support the db service op range type, but got type: %v", op.RangeType)
			default:
				return fmt.Errorf("unsupported range type: %v", op.RangeType)
			}
		}
	}
	return nil
}

func (m *MemberUsecase) IsMemberProjectAdmin(ctx context.Context, memberUid string) (bool, error) {
	member, err := m.repo.GetMember(ctx, memberUid)
	if err != nil {
		return false, fmt.Errorf("get member failed: %v", err)
	}

	for _, r := range member.RoleWithOpRanges {
		if r.RoleUID == pkgConst.UIDOfRoleProjectAdmin {
			return true, nil
		}
	}
	return false, nil
}

// FixMemberWithProjectAdmin 自动修改成员的角色和操作权限范围，如果是项目管理员，则自动添加内置的项目管理员角色
func (m *MemberUsecase) FixMemberWithProjectAdmin(ctx context.Context, member *Member, projectUid string) {
	member.RoleWithOpRanges = append(member.RoleWithOpRanges, m.GetProjectAdminRoleWithOpRange(projectUid))
}

func (m *MemberUsecase) GetProjectAdminRoleWithOpRange(projectUid string) MemberRoleWithOpRange {
	return MemberRoleWithOpRange{
		RoleUID:     pkgConst.UIDOfRoleProjectAdmin,
		OpRangeType: OpRangeTypeProject,
		RangeUIDs:   []string{projectUid},
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

func (m *MemberUsecase) ListMember(ctx context.Context, option *ListMembersOption, projectUid string) (members []*Member, total int64, err error) {
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

func (m *MemberUsecase) UpdateMember(ctx context.Context, currentUserUid, updateMemberUid, projectUid string, isProjectAdmin bool,
	roleAndOpRanges []MemberRoleWithOpRange) error {
	// check
	{
		// 检查项目是否归档/删除
		if err := m.projectUsecase.isProjectActive(ctx, projectUid); err != nil {
			return fmt.Errorf("update member error: %v", err)
		}
		// 检查当前用户有项目管理员权限
		if isAdmin, err := m.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, projectUid); err != nil {
			return fmt.Errorf("check user is project admin failed: %v", err)
		} else if !isAdmin {
			return fmt.Errorf("user is not project admin")
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

	// 如果是项目管理员，则自动添加内置的项目管理员角色
	if isProjectAdmin {
		m.FixMemberWithProjectAdmin(ctx, member, projectUid)
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
		// 检查项目是否归档/删除
		if err := m.projectUsecase.isProjectActive(ctx, member.ProjectUID); err != nil {
			return fmt.Errorf("delete member error: %v", err)
		}

		// 检查当前用户有项目管理员权限
		if isAdmin, err := m.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, member.ProjectUID); err != nil {
			return fmt.Errorf("check user is project admin failed: %v", err)
		} else if !isAdmin {
			return fmt.Errorf("user is not project admin")
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
