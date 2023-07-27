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

var _ biz.MemberRepo = (*MemberRepo)(nil)

type MemberRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewMemberRepo(log utilLog.Logger, s *Storage) *MemberRepo {
	return &MemberRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.member"))}
}

func (d *MemberRepo) SaveMember(ctx context.Context, u *biz.Member) error {
	model, err := convertBizMember(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz member: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to save member: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *MemberRepo) ListMembers(ctx context.Context, opt *biz.ListMembersOption) (members []*biz.Member, total int64, err error) {

	var models []*model.Member
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Preload("RoleWithOpRanges").Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list members: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.Member{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count members: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelMember(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model members: %v", err))
		}
		members = append(members, ds)
	}
	return members, total, nil
}

func (d *MemberRepo) GetMember(ctx context.Context, memberUid string) (*biz.Member, error) {
	var member *model.Member
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Preload("RoleWithOpRanges").First(&member, "uid = ?", memberUid).Error; err != nil {
			return fmt.Errorf("failed to get member: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelMember(member)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model member: %v", err))
	}
	return ret, nil
}

func (d *MemberRepo) CheckMemberExist(ctx context.Context, memberUids []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Member{}).Where("uid in (?)", memberUids).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check member exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(memberUids)) {
		return false, nil
	}
	return true, nil
}

func (d *MemberRepo) UpdateMember(ctx context.Context, m *biz.Member) error {

	exist, err := d.CheckMemberExist(ctx, []string{m.UID})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("member not exist"))
	}

	member, err := convertBizMember(m)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz member: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Member{}).Where("uid = ?", member.UID).Omit("created_at").Save(member).Error; err != nil {
			return fmt.Errorf("failed to update member: %v", err)
		}
		if err := tx.WithContext(ctx).Model(&model.Member{Model: model.Model{UID: member.UID}}).Association("RoleWithOpRanges").Clear(); err != nil {
			return fmt.Errorf("failed to update member role with op ranges: %v", err)
		}
		if err := tx.WithContext(ctx).Model(&model.Member{Model: model.Model{UID: member.UID}}).Association("RoleWithOpRanges").Append(member.RoleWithOpRanges); err != nil {
			return fmt.Errorf("failed to update member role with op ranges: %v", err)
		}
		return nil
	})

}

func (d *MemberRepo) DelMember(ctx context.Context, memberUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Select("RoleWithOpRanges").Delete(&model.Member{Model: model.Model{UID: memberUid}}).Error; err != nil {
			return fmt.Errorf("failed to delete member: %v", err)
		}
		return nil
	})
}

func (d *MemberRepo) DelRoleFromAllMembers(ctx context.Context, roleUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("role_uid = ?", roleUid).Delete(&model.MemberRoleOpRange{}).Error; err != nil {
			return fmt.Errorf("failed to delete role from all members: %v", err)
		}
		return nil
	})
}
