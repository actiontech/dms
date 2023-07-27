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

type ProxyTargetRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewProxyTargetRepo(log utilLog.Logger, s *Storage) *ProxyTargetRepo {
	return &ProxyTargetRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.proxytarget"))}
}

func (d *ProxyTargetRepo) SaveProxyTarget(ctx context.Context, u *biz.ProxyTarget) error {
	model, err := convertBizProxyTarget(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz proxy target: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save proxy target: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *ProxyTargetRepo) UpdateProxyTarget(ctx context.Context, u *biz.ProxyTarget) error {
	exist, err := d.CheckProxyTargetExist(ctx, []string{u.Name})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("proxy target not exist"))
	}

	target, err := convertBizProxyTarget(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz proxy target: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.ProxyTarget{}).Where("name = ?", u.Name).Omit("created_at").Save(target).Error; err != nil {
			return fmt.Errorf("failed to update proxy target: %v", err)
		}
		return nil
	})

}

func (d *ProxyTargetRepo) CheckProxyTargetExist(ctx context.Context, targetNames []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.ProxyTarget{}).Where("name in (?)", targetNames).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check proxy target exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(targetNames)) {
		return false, nil
	}
	return true, nil
}

func (d *ProxyTargetRepo) ListProxyTargets(ctx context.Context) (targets []*biz.ProxyTarget, err error) {

	var models []*model.ProxyTarget
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find targets
		if err := tx.WithContext(ctx).Find(&models).Error; err != nil {
			return fmt.Errorf("failed to list proxy targets: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// convert model to biz
	for _, model := range models {
		t, err := convertModelProxyTarget(model)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model proxy targets: %v", err))
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (d *ProxyTargetRepo) GetProxyTargetByName(ctx context.Context, name string) (*biz.ProxyTarget, error) {

	var target model.ProxyTarget
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find targets
		if err := tx.WithContext(ctx).Where(&model.ProxyTarget{Name: name}).Find(&target).Error; err != nil {
			return fmt.Errorf("failed to list proxy target: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// convert model to biz
	t, err := convertModelProxyTarget(&target)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model proxy target: %v", err))
	}

	return t, nil
}
