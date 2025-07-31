package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.MemberGroupRepo = (*MemberGroupRepo)(nil)

type MemberGroupRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewMemberGroupRepo(log utilLog.Logger, s *Storage) *MemberGroupRepo {
	return &MemberGroupRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.member_group"))}
}

func (d *MemberGroupRepo) ListMemberGroups(ctx context.Context, opt *biz.ListMemberGroupsOption) (memberGroups []*biz.MemberGroup, total int64, err error) {
	var models []*model.MemberGroup

	if err = transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Preload("RoleWithOpRanges").Preload("Users").Preload("OpPermissions").Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err = db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list member groups: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.MemberGroup{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err = db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count member groups: %v", err)
			}
		}

		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, memberGroup := range models {
		ds, err := convertModelMemberGroup(memberGroup)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user groups: %v", err))
		}
		memberGroups = append(memberGroups, ds)
	}

	return
}

func (d *MemberGroupRepo) GetMemberGroup(ctx context.Context, memberGroupId string) (*biz.MemberGroup, error) {
	var memberGroup model.MemberGroup
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Preload("RoleWithOpRanges").Preload("Users").Where("uid = ?", memberGroupId).First(&memberGroup).Error; err != nil {
			return fmt.Errorf("failed to get member group: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelMemberGroup(&memberGroup)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model member group: %v", err))
	}

	return ret, nil
}

func (d *MemberGroupRepo) CreateMemberGroup(ctx context.Context, u *biz.MemberGroup) error {
	memberGroup := convertBizMemberGroup(u)

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(memberGroup).Error; err != nil {
			return fmt.Errorf("failed to create member group err: %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *MemberGroupRepo) UpdateMemberGroup(ctx context.Context, m *biz.MemberGroup) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		memberGroup := convertBizMemberGroup(m)
		if err := tx.WithContext(ctx).Model(memberGroup).Association("Users").Clear(); err != nil {
			return fmt.Errorf("clean member group relationship with user failed: %v", err)
		}

		if err := tx.WithContext(ctx).Model(memberGroup).Association("RoleWithOpRanges").Clear(); err != nil {
			return fmt.Errorf("failed to delete member group: %v", err)
		}

		memberGroup = convertBizMemberGroup(m)
		if err := tx.WithContext(ctx).Save(memberGroup).Error; err != nil {
			return fmt.Errorf("failed to update member group err: %v", err)
		}

		return nil
	})
}

func (d *MemberGroupRepo) DeleteMemberGroup(ctx context.Context, memberGroupUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		memberGroup := model.MemberGroup{Model: model.Model{UID: memberGroupUid}}

		if err := tx.WithContext(ctx).Model(&memberGroup).Association("Users").Clear(); err != nil {
			return fmt.Errorf("clean member group relationship with user failed: %v", err)
		}

		if err := tx.WithContext(ctx).Select("RoleWithOpRanges").Delete(&memberGroup).Error; err != nil {
			return fmt.Errorf("failed to delete member group: %v", err)
		}

		return nil
	})
}

func (d *MemberGroupRepo) GetMemberGroupsByUserIDAndProjectID(ctx context.Context, userID, projectID string) ([]*biz.MemberGroup, error) {
	var memberGroups []*model.MemberGroup
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// 查询关联的用户组
		if err := tx.WithContext(ctx).Preload("RoleWithOpRanges").Preload("Users").Preload("OpPermissions").
			Joins("JOIN member_group_users ON member_groups.uid = member_group_users.member_group_uid").
			Where("member_group_users.user_uid = ? AND member_groups.project_uid = ?", userID, projectID).
			Find(&memberGroups).Error; err != nil {
			return fmt.Errorf("failed to get member groups by user id and project id: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	bizGroups := make([]*biz.MemberGroup, 0, len(memberGroups))
	for _, modelGroup := range memberGroups {
		bizGroup, err := convertModelMemberGroup(modelGroup)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model member group: %v", err))
		}
		bizGroups = append(bizGroups, bizGroup)
	}

	return bizGroups, nil
}

func (d *MemberGroupRepo) ReplaceOpPermissionsInMemberGroup(ctx context.Context, memberGroupUid string, opPermissionUids []string) error {
	if len(opPermissionUids) == 0 {
		// delete all op permissions when op permission uids is empty
		return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
			memberGroup := &model.MemberGroup{Model: model.Model{UID: memberGroupUid}}
			if err := tx.WithContext(ctx).Where("uid = ?", memberGroupUid).First(memberGroup).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("member group not found: %v", err)
				}
				return fmt.Errorf("failed to query member group existence: %v", err)
			}

			err := tx.WithContext(ctx).Model(memberGroup).Association("OpPermissions").Clear()
			if err != nil {
				return fmt.Errorf("failed to delete op permissions")
			}
			return nil
		})
	}
	var ops []*model.OpPermission
	for _, u := range opPermissionUids {
		ops = append(ops, &model.OpPermission{
			Model: model.Model{
				UID: u,
			},
		})
	}
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		memberGroup := &model.MemberGroup{Model: model.Model{UID: memberGroupUid}}
		if err := tx.WithContext(ctx).Where("uid = ?", memberGroupUid).First(memberGroup).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("member not found: %v", err)
			}
			return fmt.Errorf("failed to query member existence: %v", err)
		}

		err := tx.WithContext(ctx).Model(memberGroup).Association("OpPermissions").Replace(ops)
		if err != nil {
			return fmt.Errorf("failed to replace op permissions")
		}
		return nil
	})
}