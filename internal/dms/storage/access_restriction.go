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

var _ biz.AccessRestrictionRepo = (*AccessRestrictionRepo)(nil)

type AccessRestrictionRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewAccessRestrictionRepo(log utilLog.Logger, s *Storage) *AccessRestrictionRepo {
	return &AccessRestrictionRepo{
		Storage: s,
		log:     utilLog.NewHelper(log, utilLog.WithMessageKey("storage.access_restriction")),
	}
}

func (r *AccessRestrictionRepo) ListRules(ctx context.Context) ([]*biz.AccessWhitelistRule, error) {
	var rows []*model.AccessWhitelistRule
	if err := transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Order("updated_at DESC").Find(&rows).Error; err != nil {
			return fmt.Errorf("failed to list access whitelist rules: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	out := make([]*biz.AccessWhitelistRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, convertModelAccessWhitelistRule(row))
	}
	return out, nil
}

func (r *AccessRestrictionRepo) GetRuleByUID(ctx context.Context, uid string) (*biz.AccessWhitelistRule, error) {
	var row model.AccessWhitelistRule
	if err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to get access whitelist rule: %v", err))
	}
	return convertModelAccessWhitelistRule(&row), nil
}

func (r *AccessRestrictionRepo) GetRuleBySource(ctx context.Context, source string) (*biz.AccessWhitelistRule, error) {
	var row model.AccessWhitelistRule
	if err := r.db.WithContext(ctx).Where("source = ?", source).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to get access whitelist rule by source: %v", err))
	}
	return convertModelAccessWhitelistRule(&row), nil
}

func (r *AccessRestrictionRepo) CreateRule(ctx context.Context, rule *biz.AccessWhitelistRule) error {
	return transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(convertBizAccessWhitelistRule(rule)).Error; err != nil {
			return pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to create access whitelist rule: %v", err))
		}
		return nil
	})
}

func (r *AccessRestrictionRepo) UpdateRule(ctx context.Context, rule *biz.AccessWhitelistRule) error {
	return transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Model(&model.AccessWhitelistRule{}).Where("uid = ?", rule.UID).Updates(map[string]interface{}{
			"source":      rule.Source,
			"policy_type": rule.PolicyType,
			"remark":      rule.Remark,
		})
		if result.Error != nil {
			return pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to update access whitelist rule: %v", result.Error))
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("规则不存在")
		}
		return nil
	})
}

func (r *AccessRestrictionRepo) DeleteRule(ctx context.Context, uid string) error {
	return transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Where("uid = ?", uid).Delete(&model.AccessWhitelistRule{})
		if result.Error != nil {
			return pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to delete access whitelist rule: %v", result.Error))
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("规则不存在")
		}
		return nil
	})
}

func (r *AccessRestrictionRepo) GetEnabled(ctx context.Context) (bool, error) {
	var row model.SystemVariable
	if err := r.db.WithContext(ctx).Where("`key` = ?", biz.AccessRestrictionEnabledKey).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to get access restriction switch: %v", err))
	}
	return row.Value == "true", nil
}

func (r *AccessRestrictionRepo) SetEnabled(ctx context.Context, enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	return transaction(r.log, ctx, r.db, func(tx *gorm.DB) error {
		row := &model.SystemVariable{
			Key:   biz.AccessRestrictionEnabledKey,
			Value: value,
		}
		if err := tx.WithContext(ctx).Save(row).Error; err != nil {
			return pkgErr.WrapStorageErr(r.log, fmt.Errorf("failed to set access restriction switch: %v", err))
		}
		return nil
	})
}

func convertBizAccessWhitelistRule(b *biz.AccessWhitelistRule) *model.AccessWhitelistRule {
	return &model.AccessWhitelistRule{
		Model: model.Model{
			UID:       b.UID,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		},
		Source:     b.Source,
		PolicyType: b.PolicyType,
		Remark:     b.Remark,
	}
}

func convertModelAccessWhitelistRule(m *model.AccessWhitelistRule) *biz.AccessWhitelistRule {
	return &biz.AccessWhitelistRule{
		Base:       convertBase(m.Model),
		UID:        m.UID,
		Source:     m.Source,
		PolicyType: m.PolicyType,
		Remark:     m.Remark,
	}
}
