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

var _ biz.OpPermissionRepo = (*OpPermissionRepo)(nil)

type OpPermissionRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOpPermissionRepo(log utilLog.Logger, s *Storage) *OpPermissionRepo {
	return &OpPermissionRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.oppermission"))}
}

func (d *OpPermissionRepo) SaveOpPermission(ctx context.Context, u *biz.OpPermission) error {
	model, err := convertBizOpPermission(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz opPermission: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save opPermission: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *OpPermissionRepo) CheckOpPermissionExist(ctx context.Context, opPermissionUids []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.OpPermission{}).Where("uid in (?)", opPermissionUids).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check opPermission exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(opPermissionUids)) {
		return false, nil
	}
	return true, nil
}

func (d *OpPermissionRepo) UpdateOpPermission(ctx context.Context, u *biz.OpPermission) error {
	exist, err := d.CheckOpPermissionExist(ctx, []string{u.UID})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("opPermission not exist"))
	}

	opPermission, err := convertBizOpPermission(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz opPermission: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.OpPermission{}).Where("uid = ?", u.UID).Omit("created_at").Save(opPermission).Error; err != nil {
			return fmt.Errorf("failed to update opPermission: %v", err)
		}
		return nil
	})

}

func (d *OpPermissionRepo) ListOpPermissions(ctx context.Context, opt *biz.ListOpPermissionsOption) (opPermissions []*biz.OpPermission, total int64, err error) {

	var models []*model.OpPermission
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list opPermissions: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.OpPermission{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count opPermissions: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelOpPermission(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model opPermissions: %v", err))
		}
		opPermissions = append(opPermissions, ds)
	}
	return opPermissions, total, nil
}

func (d *OpPermissionRepo) DelOpPermission(ctx context.Context, opPermissionUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", opPermissionUid).Delete(&model.OpPermission{}).Error; err != nil {
			return fmt.Errorf("failed to delete opPermission: %v", err)
		}
		return nil
	})
}

func (d *OpPermissionRepo) GetOpPermission(ctx context.Context, opPermissionUid string) (*biz.OpPermission, error) {
	var opPermission *model.OpPermission
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&opPermission, "uid = ?", opPermissionUid).Error; err != nil {
			return fmt.Errorf("failed to get opPermission: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelOpPermission(opPermission)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model opPermission: %v", err))
	}
	return ret, nil
}
