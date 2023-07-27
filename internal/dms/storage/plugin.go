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

type PluginRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewPluginRepo(log utilLog.Logger, s *Storage) *PluginRepo {
	return &PluginRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.plugin"))}
}

func (d *PluginRepo) SavePlugin(ctx context.Context, u *biz.Plugin) error {
	model, err := convertBizPlugin(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz plugin: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save plugin: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *PluginRepo) UpdatePlugin(ctx context.Context, u *biz.Plugin) error {
	exist, err := d.CheckPluginExist(ctx, []string{u.Name})
	if err != nil {
		return err
	}
	if !exist {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("plugin not exist"))
	}

	plugin, err := convertBizPlugin(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz plugin: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Plugin{}).Where("name = ?", u.Name).Omit("created_at").Save(plugin).Error; err != nil {
			return fmt.Errorf("failed to update plugin: %v", err)
		}
		return nil
	})

}

func (d *PluginRepo) CheckPluginExist(ctx context.Context, pluginNames []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Plugin{}).Where("name in (?)", pluginNames).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check plugin exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(pluginNames)) {
		return false, nil
	}
	return true, nil
}

func (d *PluginRepo) ListPlugins(ctx context.Context) (plugins []*biz.Plugin, err error) {

	var models []*model.Plugin
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find plugins
		if err := tx.WithContext(ctx).Find(&models).Error; err != nil {
			return fmt.Errorf("failed to list plugins: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// convert model to biz
	for _, model := range models {
		t, err := convertModelPlugin(model)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model plugins: %v", err))
		}
		plugins = append(plugins, t)
	}
	return plugins, nil
}
