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

var _ biz.RoleRepo = (*RoleRepo)(nil)

type RoleRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewRoleRepo(log utilLog.Logger, s *Storage) *RoleRepo {
	return &RoleRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.role"))}
}

func (d *RoleRepo) SaveRole(ctx context.Context, u *biz.Role) error {
	model, err := convertBizRole(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz role: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save role: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *RoleRepo) CheckRoleExist(ctx context.Context, roleUids []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Role{}).Where("uid in (?)", roleUids).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check role exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(roleUids)) {
		return false, nil
	}
	return true, nil
}

func (d *RoleRepo) UpdateRole(ctx context.Context, u *biz.Role) error {
	exist, err := d.CheckRoleExist(ctx, []string{u.UID})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("role not exist"))
	}

	role, err := convertBizRole(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz role: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Role{}).Where("uid = ?", u.UID).Omit("created_at").Save(role).Error; err != nil {
			return fmt.Errorf("failed to update role: %v", err)
		}
		return nil
	})

}

func (d *RoleRepo) AddOpPermissionToRole(ctx context.Context, OpPermissionUid string, roleUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.Role{Model: model.Model{UID: roleUid}}).Association("OpPermissions").Append(&model.OpPermission{
			Model: model.Model{
				UID: OpPermissionUid,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to add op permission to role: %v", err)
		}
		return nil
	})
}

func (d *RoleRepo) ReplaceOpPermissionsInRole(ctx context.Context, roleUid string, OpPermissionUids []string) error {
	var ops []*model.OpPermission
	for _, u := range OpPermissionUids {
		ops = append(ops, &model.OpPermission{
			Model: model.Model{
				UID: u,
			},
		})
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.Role{Model: model.Model{
			UID: roleUid,
		}}).Association("OpPermissions").Replace(ops)
		if err != nil {
			return fmt.Errorf("failed to replace op permissions in role: %v", err)
		}
		return nil
	})
}

func (d *RoleRepo) ListRoles(ctx context.Context, opt *biz.ListRolesOption) (roles []*biz.Role, total int64, err error) {
	// 取出权限条件的值
	opPermissionValue := ""
	for i := 0; i < len(opt.FilterBy); {
		if opt.FilterBy[i].Field == string(biz.RoleFieldOpPermission) {
			opPermissionValue = opt.FilterBy[i].Value.(string)
			opt.FilterBy = append(opt.FilterBy[:i], opt.FilterBy[i+1:]...)
		} else {
			i++
		}
	}

	var models []*model.Role
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				if f.Field != string(biz.RoleFieldOpPermission) {
					db = gormWhere(db, f)
				}
			}
			if opPermissionValue != "" {
				db = db.Joins("JOIN role_op_permissions on roles.uid = role_op_permissions.role_uid").
					Joins("JOIN op_permissions ON op_permissions.uid = role_op_permissions.op_permission_uid").
					Where("op_permissions.name like ?", "%"+opPermissionValue+"%").
					Distinct()
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list roles: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.Role{})
			for _, f := range opt.FilterBy {
				if f.Field != string(biz.RoleFieldOpPermission) {
					db = gormWhere(db, f)
				}
			}
			if opPermissionValue != "" {
				db = db.Joins("JOIN role_op_permissions on roles.uid = role_op_permissions.role_uid").
					Joins("JOIN op_permissions ON op_permissions.uid = role_op_permissions.op_permission_uid").
					Where("op_permissions.name like ?", "%"+opPermissionValue+"%").
					Group("roles.uid")
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count roles: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelRole(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model roles: %v", err))
		}
		roles = append(roles, ds)
	}
	return roles, total, nil
}

func (d *RoleRepo) DelRole(ctx context.Context, roleUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", roleUid).Delete(&model.Role{}).Error; err != nil {
			return fmt.Errorf("failed to delete role: %v", err)
		}
		return nil
	})
}

func (d *RoleRepo) GetRole(ctx context.Context, roleUid string) (*biz.Role, error) {
	var role *model.Role
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&role, "uid = ?", roleUid).Error; err != nil {
			return fmt.Errorf("failed to get role: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelRole(role)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model role: %v", err))
	}
	return ret, nil
}

func (d *RoleRepo) GetOpPermissionsByRole(ctx context.Context, roleUid string) ([]*biz.OpPermission, error) {
	var ops []*model.OpPermission

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Role{Model: model.Model{UID: roleUid}}).Association("OpPermissions").Find(&ops); err != nil {
			return fmt.Errorf("failed to get op permissions by role: %v", err)
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

func (d *RoleRepo) DelAllOpPermissionsFromRole(ctx context.Context, roleUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&model.Role{Model: model.Model{UID: roleUid}}).Association("OpPermissions").Clear()
		if err != nil {
			return fmt.Errorf("failed to del all op permissions from role: %v", err)
		}
		return nil
	})
}
