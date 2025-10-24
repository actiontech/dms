package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ biz.UserRepo = (*UserRepo)(nil)

type UserRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewUserRepo(log utilLog.Logger, s *Storage) *UserRepo {
	return &UserRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.user"))}
}

func (d *UserRepo) SaveUser(ctx context.Context, u *biz.User) error {
	model, err := convertBizUser(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz user: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save user: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *UserRepo) CheckUserExist(ctx context.Context, userUids []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.User{}).Where("uid in (?)", userUids).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check user exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(userUids)) {
		return false, nil
	}
	return true, nil
}

func (d *UserRepo) UpdateUser(ctx context.Context, u *biz.User) error {
	exist, err := d.CheckUserExist(ctx, []string{u.UID})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("user not exist"))
	}

	user, err := convertBizUser(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz user: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.User{}).Where("uid = ?", u.UID).Omit("created_at").Save(user).Error; err != nil {
			return fmt.Errorf("failed to update user: %v", err)
		}
		return nil
	})

}

func (d *UserRepo) AddUserToUserGroup(ctx context.Context, userGroupUid string, userUid string) error {
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

func (d *UserRepo) DelUserFromAllUserGroups(ctx context.Context, userUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.User{Model: model.Model{UID: userUid}}).Association("UserGroups").Clear()
		if err != nil {
			return fmt.Errorf("failed to del user from all user groups: %v", err)
		}
		return nil
	})
}

func (d *UserRepo) ReplaceUserGroupsInUser(ctx context.Context, userUid string, userGroupUids []string) error {
	var groups []*model.UserGroup
	for _, u := range userGroupUids {
		groups = append(groups, &model.UserGroup{
			Model: model.Model{
				UID: u,
			},
		})
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.User{Model: model.Model{
			UID: userUid,
		}}).Association("UserGroups").Replace(groups)
		if err != nil {
			return fmt.Errorf("failed to replace user groups in user: %v", err)
		}
		return nil
	})
}

func (d *UserRepo) ReplaceOpPermissionsInUser(ctx context.Context, userUid string, OpPermissionUids []string) error {
	var ops []*model.OpPermission
	for _, u := range OpPermissionUids {
		ops = append(ops, &model.OpPermission{
			Model: model.Model{
				UID: u,
			},
		})
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.User{Model: model.Model{
			UID: userUid,
		}}).Association("OpPermissions").Replace(ops)
		if err != nil {
			return fmt.Errorf("failed to replace op permissions in user: %v", err)
		}
		return nil
	})
}

// ListUsers 使用了Unscoped()查询出被软删除的记录
// 目的是在关联数据中展示出被删除用户的名字.例如：工单创建者 : test[x]
func (d *UserRepo) ListUsers(ctx context.Context, opt *biz.ListUsersOption) (users []*biz.User, total int64, err error) {

	var models []*model.User
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.Unscoped().WithContext(ctx).Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Preload("Members.Project").Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list users: %v", err)
			}
		}

		// find total
		{
			db := tx.Unscoped().WithContext(ctx).Model(&model.User{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count users: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelUser(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model users: %v", err))
		}
		users = append(users, ds)
	}
	return users, total, nil
}

func (d *UserRepo) CountUsers(ctx context.Context, opts []constant.FilterCondition) (total int64, err error) {
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find total
		db := tx.WithContext(ctx).Model(&model.User{})
		for _, f := range opts {
			db = gormWhere(db, f)
		}
		if err := db.Count(&total).Error; err != nil {
			return fmt.Errorf("failed to count users: %v", err)
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return total, nil
}

func (d *UserRepo) DelUser(ctx context.Context, userUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Select("OpPermissions").Delete(&model.User{Model: model.Model{UID: userUid}}).Error; err != nil {
			return fmt.Errorf("failed to delete user: %v", err)
		}

		// delete members, member_role_op_ranges
		if err := tx.WithContext(ctx).Exec(`delete m, mror from members m left join member_role_op_ranges mror on m.uid = mror.member_uid where m.user_uid = ?`, userUid).Error; err != nil {
			return fmt.Errorf("failed to delete user associate members: %v", err)
		}

		// delete member_group_users
		if err := tx.WithContext(ctx).Exec(`delete mgu from member_group_users mgu where mgu.user_uid = ?`, userUid).Error; err != nil {
			return fmt.Errorf("failed to delete user associate member_groups: %v", err)
		}

		return nil
	})
}

func (d *UserRepo) GetUser(ctx context.Context, userUid string) (*biz.User, error) {
	var user *model.User
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&user, "uid = ?", userUid).Error; err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelUser(user)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user: %v", err))
	}
	return ret, nil
}

// GetUserIncludeDeleted 查找用户，包括已删除用户
func (d *UserRepo) GetUserIncludeDeleted(ctx context.Context, userUid string) (*biz.User, error) {
	var user *model.User
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// 不加全局软删除约束
		if err := tx.Unscoped().First(&user, "uid = ?", userUid).Error; err != nil {
			return fmt.Errorf("failed to get user(include deleted): %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if user.DeletedAt.Valid {
		user.Name = user.Name + "[x]"
	}
	ret, err := convertModelUser(user)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user(include deleted): %v", err))
	}
	return ret, nil
}

func (d *UserRepo) GetUserByName(ctx context.Context, userName string) (*biz.User, error) {
	var user *model.User
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&user, "name = ?", userName).Error; err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelUser(user)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user: %v", err))
	}
	return ret, nil
}

func (d *UserRepo) GetUserGroupsByUser(ctx context.Context, userUid string) ([]*biz.UserGroup, error) {
	var userGroups []*model.UserGroup

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.User{Model: model.Model{UID: userUid}}).Association("UserGroups").Find(&userGroups); err != nil {
			return fmt.Errorf("failed to get user groups by user: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ret []*biz.UserGroup
	for _, userGroup := range userGroups {
		r, err := convertModelUserGroup(userGroup)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user group: %v", err))
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func (d *UserRepo) GetOpPermissionsByUser(ctx context.Context, userUid string) ([]*biz.OpPermission, error) {
	var ops []*model.OpPermission

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.User{Model: model.Model{UID: userUid}}).Association("OpPermissions").Find(&ops); err != nil {
			return fmt.Errorf("failed to get op permissions by user: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var ret []*biz.OpPermission
	for _, op := range ops {
		r, err := convertModelOpPermission(op)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model op permission: %v", err))
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func (d *UserRepo) GetUserByThirdPartyUserID(ctx context.Context, thirdPartyUserUID string) (*biz.User, error) {
	var user *model.User
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&user, "third_party_user_id = ?", thirdPartyUserUID).Error; err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelUser(user)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model user: %v", err))
	}
	return ret, nil
}

func (d *UserRepo) SaveAccessToken(ctx context.Context, tokenInfo *biz.AccessTokenInfo) error {
	userAccessToekn := &model.UserAccessToken{
		Model: model.Model{
			UID: tokenInfo.UID,
		},
		UserID:      tokenInfo.UserID,
		Token:       tokenInfo.Token,
		ExpiredTime: tokenInfo.ExpiredTime,
	}

	tx := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"token": tokenInfo.Token, "expired_time": tokenInfo.ExpiredTime}),
	}).Create(userAccessToekn)

	if tx.Error != nil {
		return fmt.Errorf("failed to save access token: %v", tx.Error)
	}

	return nil
}

func (d *UserRepo) GetAccessTokenByUser(ctx context.Context, userUid string) (*biz.AccessTokenInfo, error) {
	var userToken *model.UserAccessToken
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&userToken, "user_id = ?", userUid).Error; err != nil {
			// 未找到记录返回空，不影响获取用户信息的功能
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("failed to get user access token: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &biz.AccessTokenInfo{Token: userToken.Token, ExpiredTime: userToken.ExpiredTime}, nil
}
