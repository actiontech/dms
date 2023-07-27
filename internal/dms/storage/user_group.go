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

var _ biz.UserGroupRepo = (*UserGroupRepo)(nil)

type UserGroupRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewUserGroupRepo(log utilLog.Logger, s *Storage) *UserGroupRepo {
	return &UserGroupRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.usergroup"))}
}

func (d *UserGroupRepo) SaveUserGroup(ctx context.Context, u *biz.UserGroup) error {
	model, err := convertBizUserGroup(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz user group: %w", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Save(model).Error
		if err != nil {
			return fmt.Errorf("failed to add user to user group: %v", err)
		}
		return nil
	})
}

func (d *UserGroupRepo) CheckUserGroupExist(ctx context.Context, userGroupUids []string) (exists bool, err error) {
	var count int64

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.UserGroup{}).Where("uid in (?)", userGroupUids).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check user group exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(userGroupUids)) {
		return false, nil
	}
	return true, nil
}

func (d *UserGroupRepo) UpdateUserGroup(ctx context.Context, u *biz.UserGroup) error {
	exist, err := d.CheckUserGroupExist(ctx, []string{u.UID})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("user group not exist"))
	}
	group, err := convertBizUserGroup(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz user group: %v", err))
	}
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.UserGroup{}).Where("uid = ?", u.UID).Omit("created_at").Save(group).Error; err != nil {
			return fmt.Errorf("failed to update user group: %v", err)
		}
		return nil
	})
}

func (d *UserGroupRepo) AddUserToUserGroup(ctx context.Context, userGroupUid string, userUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.User{Model: model.Model{UID: userUid}}).Association("UserGroups").Append(&model.UserGroup{
			Model: model.Model{
				UID: userGroupUid,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to add user to user group: %v", err)
		}
		return nil
	})
}

func (d *UserGroupRepo) ReplaceUsersInUserGroup(ctx context.Context, userGroupUid string, userUids []string) error {
	var users []*model.User
	for _, u := range userUids {
		users = append(users, &model.User{
			Model: model.Model{
				UID: u,
			},
		})
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.UserGroup{Model: model.Model{
			UID: userGroupUid,
		}}).Association("Users").Replace(users)
		if err != nil {
			return fmt.Errorf("failed to replace users in user group: %v", err)
		}
		return nil
	})

}

func (d *UserGroupRepo) ListUserGroups(ctx context.Context, opt *biz.ListUserGroupsOption) (userGroups []*biz.UserGroup, total int64, err error) {

	var models []*model.UserGroup

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list user groups: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.UserGroup{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count user groups: %v", err)
			}
		}
		return nil

	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelUserGroup(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user groups: %v", err))
		}
		userGroups = append(userGroups, ds)
	}
	return userGroups, total, nil
}

func (d *UserGroupRepo) DelUserGroup(ctx context.Context, userGroupUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		userGroup := new(model.UserGroup)
		userGroup.UID = userGroupUid
		if err := tx.WithContext(ctx).Model(&userGroup).Association("Users").Clear(); err != nil {
			return fmt.Errorf("clean user group relationship with user failed: %v", err)
		}

		if err := tx.WithContext(ctx).Where("uid = ?", userGroupUid).Delete(&model.UserGroup{}).Error; err != nil {
			return fmt.Errorf("failed to delete user group: %v", err)
		}

		return nil
	})

}

func (d *UserGroupRepo) GetUserGroup(ctx context.Context, userGroupUid string) (*biz.UserGroup, error) {
	var user *model.UserGroup
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).First(&user, "uid = ?", userGroupUid).Error; err != nil {
			return fmt.Errorf("failed to get user group : %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelUserGroup(user)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user group: %v", err))
	}
	return ret, nil
}

func (d *UserGroupRepo) GetUsersInUserGroup(ctx context.Context, userGroupUid string) ([]*biz.User, error) {
	var users []*model.User

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.UserGroup{Model: model.Model{UID: userGroupUid}}).Association("Users").Find(&users); err != nil {
			return fmt.Errorf("failed to get users in user group: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ret []*biz.User
	for _, user := range users {
		r, err := convertModelUser(user)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user: %v", err))
		}
		ret = append(ret, r)
	}
	return ret, nil
}
