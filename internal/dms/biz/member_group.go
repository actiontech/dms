package biz

import (
	"context"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type MemberGroup struct {
	Base

	IsNamespaceAdmin bool
	UID              string
	Name             string
	NamespaceUID     string
	UserUids         []string
	Users            []UserIdWithName
	RoleWithOpRanges []MemberRoleWithOpRange
}

type UserIdWithName struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

type MemberGroupRepo interface {
	ListMemberGroups(ctx context.Context, opt *ListMemberGroupsOption) (memberGroups []*MemberGroup, total int64, err error)
	GetMemberGroup(ctx context.Context, memberGroupId string) (*MemberGroup, error)
	CreateMemberGroup(ctx context.Context, mg *MemberGroup) error
	UpdateMemberGroup(ctx context.Context, mg *MemberGroup) error
	DeleteMemberGroup(ctx context.Context, memberGroupId string) error
}

type MemberGroupUsecase struct {
	tx                        TransactionGenerator
	repo                      MemberGroupRepo
	userUsecase               *UserUsecase
	roleUsecase               *RoleUsecase
	dbServiceUsecase          *DBServiceUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	namespaceUsecase          *NamespaceUsecase
	memberUsecase             *MemberUsecase
	log                       *utilLog.Helper
}

func NewMemberGroupUsecase(log utilLog.Logger, tx TransactionGenerator, repo MemberGroupRepo,
	userUsecase *UserUsecase,
	roleUsecase *RoleUsecase,
	dbServiceUsecase *DBServiceUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase,
	namespaceUsecase *NamespaceUsecase,
	memberUsecase *MemberUsecase) *MemberGroupUsecase {
	return &MemberGroupUsecase{
		tx:                        tx,
		repo:                      repo,
		userUsecase:               userUsecase,
		roleUsecase:               roleUsecase,
		dbServiceUsecase:          dbServiceUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		namespaceUsecase:          namespaceUsecase,
		memberUsecase:             memberUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.member_group")),
	}
}

type ListMemberGroupsOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      MemberGroupField
	FilterBy     []pkgConst.FilterCondition
}

func (m *MemberGroupUsecase) ListMemberGroups(ctx context.Context, option *ListMemberGroupsOption, namespaceUid string) ([]*MemberGroup, int64, error) {
	// 检查空间是否归档/删除
	if err := m.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
		return nil, 0, fmt.Errorf("list member groups error: %v", err)
	}
	members, total, err := m.repo.ListMemberGroups(ctx, option)
	if err != nil {
		return nil, 0, fmt.Errorf("list member groups failed: %v", err)
	}

	return members, total, nil
}

func (m *MemberGroupUsecase) IsMemberGroupNamespaceAdmin(ctx context.Context, memberGroupUid string) (bool, error) {
	member, err := m.repo.GetMemberGroup(ctx, memberGroupUid)
	if err != nil {
		return false, fmt.Errorf("get member group failed: %v", err)
	}

	for _, r := range member.RoleWithOpRanges {
		if r.RoleUID == pkgConst.UIDOfRoleNamespaceAdmin {
			return true, nil
		}
	}

	return false, nil
}

func (m *MemberGroupUsecase) GetMemberGroup(ctx context.Context, memberGroupUid, namespaceUid string) (*MemberGroup, error) {
	// 检查空间是否归档/删除
	if err := m.namespaceUsecase.isNamespaceActive(ctx, namespaceUid); err != nil {
		return nil, fmt.Errorf("get member groups error: %v", err)
	}

	return m.repo.GetMemberGroup(ctx, memberGroupUid)
}

func (m *MemberGroupUsecase) CreateMemberGroup(ctx context.Context, currentUserUid string, mg *MemberGroup) (string, error) {
	// check
	if err := m.checkMemberGroupBeforeUpsert(ctx, currentUserUid, mg); err != nil {
		return "", fmt.Errorf("create member group error: %v", err)
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return "", err
	}
	mg.UID = uid

	if mg.IsNamespaceAdmin {
		mg.RoleWithOpRanges = append(mg.RoleWithOpRanges, MemberRoleWithOpRange{
			RoleUID:     pkgConst.UIDOfRoleNamespaceAdmin,
			OpRangeType: OpRangeTypeNamespace,
			RangeUIDs:   []string{mg.NamespaceUID},
		})
	}

	if err = m.repo.CreateMemberGroup(ctx, mg); err != nil {
		return "", fmt.Errorf("save member group failed: %v", err)
	}

	return uid, nil

}

func (m *MemberGroupUsecase) checkMemberGroupBeforeUpsert(ctx context.Context, currentUserUid string, mg *MemberGroup) error {
	// 检查空间是否归档/删除
	if err := m.namespaceUsecase.isNamespaceActive(ctx, mg.NamespaceUID); err != nil {
		return fmt.Errorf("create member error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := m.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, mg.NamespaceUID); err != nil {
		return fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not namespace admin")
	}

	// 检查成员组成员用户存在
	if exist, err := m.userUsecase.CheckUserExist(ctx, mg.UserUids); err != nil {
		return fmt.Errorf("check user exist failed: %v", err)
	} else if !exist {
		return fmt.Errorf("user not exist")
	}

	return m.memberUsecase.CheckRoleAndOpRanges(ctx, mg.RoleWithOpRanges)
}

func (m *MemberGroupUsecase) UpdateMemberGroup(ctx context.Context, currentUserUid string, mg *MemberGroup) error {
	// check
	if err := m.checkMemberGroupBeforeUpsert(ctx, currentUserUid, mg); err != nil {
		return fmt.Errorf("update member group error: %v", err)
	}

	memberGroup, err := m.GetMemberGroup(ctx, mg.UID, mg.NamespaceUID)
	if err != nil {
		return err
	}

	if mg.IsNamespaceAdmin {
		mg.RoleWithOpRanges = append(mg.RoleWithOpRanges, MemberRoleWithOpRange{
			RoleUID:     pkgConst.UIDOfRoleNamespaceAdmin,
			OpRangeType: OpRangeTypeNamespace,
			RangeUIDs:   []string{mg.NamespaceUID},
		})
	}

	mg.UID = memberGroup.UID
	mg.Name = memberGroup.Name
	mg.CreatedAt = memberGroup.CreatedAt

	if err = m.repo.UpdateMemberGroup(ctx, mg); err != nil {
		return fmt.Errorf("update member group failed: %v", err)
	}

	return nil
}

func (m *MemberGroupUsecase) DeleteMemberGroup(ctx context.Context, currentUserUid, memberGroupUid, namespaceUid string) (err error) {
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
	}

	return m.repo.DeleteMemberGroup(ctx, memberGroupUid)
}
