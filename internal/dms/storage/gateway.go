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

var _ biz.GatewayRepo = (*GatewayRepo)(nil)

type GatewayRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewGatewayRepo(log utilLog.Logger, s *Storage) *GatewayRepo {
	return &GatewayRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.gateway"))}
}

// AddGateway 添加一个新网关
func (r *GatewayRepo) AddGateway(ctx context.Context, u *biz.Gateway) error {
	gateway := &model.Gateway{
		Model: model.Model{
			UID: u.ID,
		},
		Name:        u.Name,
		Description: u.Desc,
		Address:     u.URL,
	}

	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(gateway).Error; err != nil {
			return fmt.Errorf("failed to add gateway: %v", err)
		}
		return nil
	}); err != nil {
		return pkgErr.WrapStorageErr(r.log, err)
	}

	return nil
}

// DeleteGateway 删除网关
func (r *GatewayRepo) DeleteGateway(ctx context.Context, id string) error {
	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", id).Delete(&model.Gateway{}).Error; err != nil {
			return fmt.Errorf("failed to delete gateway: %v", err)
		}
		return nil
	}); err != nil {
		return pkgErr.WrapStorageErr(r.log, err)
	}

	return nil
}

// UpdateGateway 更新网关
func (r *GatewayRepo) UpdateGateway(ctx context.Context, u *biz.Gateway) error {
	gateway := &model.Gateway{
		Model: model.Model{
			UID: u.ID,
		},
		Name:        u.Name,
		Description: u.Desc,
		Address:     u.URL,
	}

	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Gateway{}).Where("uid = ?", u.ID).Updates(gateway).Error; err != nil {
			return fmt.Errorf("failed to update gateway: %v", err)
		}
		return nil
	}); err != nil {
		return pkgErr.WrapStorageErr(r.log, err)
	}

	return nil
}

// GetGateway 根据ID获取网关
func (r *GatewayRepo) GetGateway(ctx context.Context, id string) (*biz.Gateway, error) {
	var gateway model.Gateway

	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", id).First(&gateway).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("gateway %s not found", id)
			}
			return fmt.Errorf("failed to get gateway: %v", err)
		}
		return nil
	}); err != nil {
		return nil, pkgErr.WrapStorageErr(r.log, err)
	}

	return &biz.Gateway{
		ID:   gateway.UID,
		Name: gateway.Name,
		Desc: gateway.Description,
		URL:  gateway.Address,
	}, nil
}

// ListGateways 列出所有网关
func (r *GatewayRepo) ListGateways(ctx context.Context, opt *biz.ListGatewaysOption) ([]*biz.Gateway, int64, error) {
	var gateways []*model.Gateway
	var total int64

	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		// 查询网关列表
		{
			db := tx.WithContext(ctx).Order(string(opt.OrderBy))
			db = gormWheresWithOptions(ctx, db, opt.FilterByOptions)
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber))))).Find(&gateways)
			if err := db.Error; err != nil {
				return fmt.Errorf("failed to list gateways: %v", err)
			}
		}

		// 查询总数
		{
			db := tx.WithContext(ctx).Model(&model.Gateway{})
			db = gormWheresWithOptions(ctx, db, opt.FilterByOptions)
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count gateways: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, pkgErr.WrapStorageErr(r.log, err)
	}

	// 转换为业务模型
	result := make([]*biz.Gateway, len(gateways))
	for i, g := range gateways {
		result[i] = &biz.Gateway{
			ID:   g.UID,
			Name: g.Name,
			Desc: g.Description,
			URL:  g.Address,
		}
	}

	return result, total, nil
}

// GetGatewayTips 获取网关提示信息
func (r *GatewayRepo) GetGatewayTips(ctx context.Context) ([]*biz.Gateway, error) {
	var gateways []*model.Gateway

	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Find(&gateways).Error; err != nil {
			return fmt.Errorf("failed to get gateway tips: %v", err)
		}
		return nil
	}); err != nil {
		return nil, pkgErr.WrapStorageErr(r.log, err)
	}

	// 转换为业务模型
	result := make([]*biz.Gateway, len(gateways))
	for i, g := range gateways {
		result[i] = &biz.Gateway{
			ID:   g.UID,
			Name: g.Name,
			Desc: g.Description,
			URL:  g.Address,
		}
	}

	return result, nil
}

func (d *GatewayRepo) SyncGateways(ctx context.Context, s []*biz.Gateway) error {
	gateways := make([]*model.Gateway, len(s))
	for i, v := range s {
		gateways[i] = &model.Gateway{
			Model: model.Model{
				UID: v.ID,
			},
			Name:        v.Name,
			Description: v.Desc,
			Address:     v.URL,
		}
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// 第一步：删除所有现有的 gateways（整表清空）
		if err := tx.WithContext(ctx).Where("1 = 1").Delete(&model.Gateway{}).Error; err != nil {
			return err
		}
		// 第二步：插入新的 gateways
		return tx.WithContext(ctx).Save(gateways).Error
	})
}
