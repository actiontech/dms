package storage

import (
	"context"
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
	return &MemberGroupRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.member"))}
}

func (d *MemberGroupRepo) ListMemberGroups(ctx context.Context, opt *biz.ListMemberGroupsOption) (memberGroups []*biz.MemberGroup, total int64, err error) {
	var models []*model.MemberGroup

	if err = transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Preload("RoleWithOpRanges").Preload("Users").Order(opt.OrderBy)
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
		if err := tx.WithContext(ctx).Debug().Create(memberGroup).Error; err != nil {
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
		if err := tx.WithContext(ctx).Debug().Model(memberGroup).Association("Users").Clear(); err != nil {
			return fmt.Errorf("clean member group relationship with user failed: %v", err)
		}

		if err := tx.WithContext(ctx).Debug().Model(memberGroup).Association("RoleWithOpRanges").Clear(); err != nil {
			return fmt.Errorf("failed to delete member group: %v", err)
		}

		memberGroup = convertBizMemberGroup(m)
		if err := tx.WithContext(ctx).Debug().Save(memberGroup).Error; err != nil {
			return fmt.Errorf("failed to update member group err: %v", err)
		}

		return nil
	})
}

func (d *MemberGroupRepo) DeleteMemberGroup(ctx context.Context, memberGroupUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		memberGroup := model.MemberGroup{Model: model.Model{UID: memberGroupUid}}

		if err := tx.WithContext(ctx).Debug().Model(&memberGroup).Association("Users").Clear(); err != nil {
			return fmt.Errorf("clean member group relationship with user failed: %v", err)
		}

		if err := tx.WithContext(ctx).Debug().Select("RoleWithOpRanges").Delete(&memberGroup).Error; err != nil {
			return fmt.Errorf("failed to delete member group: %v", err)
		}

		return nil
	})
}
