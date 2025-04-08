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

var _ biz.BusinessTagRepo = (*BusinessTagRepo)(nil)

type BusinessTagRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewBusinessTagRepo(log utilLog.Logger, s *Storage) *BusinessTagRepo {
	return &BusinessTagRepo{
		Storage: s,
		log:     utilLog.NewHelper(log, utilLog.WithMessageKey("storage.business_tag")),
	}
}

func (repo *BusinessTagRepo) toModel(businessTag *biz.BusinessTag) *model.BusinessTag {
	return &model.BusinessTag{
		Name:  businessTag.Name,
		Model: model.Model{UID: businessTag.UID},
	}
}

func (repo *BusinessTagRepo) toBiz(businessTag *model.BusinessTag) *biz.BusinessTag {
	return &biz.BusinessTag{
		Name: businessTag.Name,
		UID:  businessTag.UID,
	}
}

func (repo *BusinessTagRepo) CreateBusinessTag(ctx context.Context, businessTag *biz.BusinessTag) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(repo.toModel(businessTag)).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to create business tag: %v", err))
		}
		return nil
	})
}

func (repo *BusinessTagRepo) UpdateBusinessTag(ctx context.Context, businessTagUID, businessTagName string) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.BusinessTag{}).Where("uid = ?", businessTagUID).Update("name", businessTagName).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to update business tag: %v", err))
		}
		return nil
	})
}

func (repo *BusinessTagRepo) DeleteBusinessTag(ctx context.Context, businessTagUID string) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", businessTagUID).Delete(&model.BusinessTag{}).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to delete business tag: %v", err))
		}
		return nil
	})
}

func (repo *BusinessTagRepo) GetBusinessTagByName(ctx context.Context, name string) (*biz.BusinessTag, error) {
	var businessTag model.BusinessTag
	if err := repo.db.WithContext(ctx).Where("name = ?", name).First(&businessTag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgErr.ErrStorageNoData
		}
		return nil, fmt.Errorf("failed to get business tag by name: %w", err)
	}
	return repo.toBiz(&businessTag), nil
}

func (repo *BusinessTagRepo) GetBusinessTagByUID(ctx context.Context, uid string) (*biz.BusinessTag, error) {
	var businessTag model.BusinessTag
	if err := repo.db.WithContext(ctx).Where("uid = ?", uid).First(&businessTag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgErr.ErrStorageNoData
		}
		return nil, fmt.Errorf("failed to get business tag by uid: %w", err)
	}
	return repo.toBiz(&businessTag), nil
}
func (repo *BusinessTagRepo) ListBusinessTags(ctx context.Context, options *biz.ListBusinessTagsOption) ([]*biz.BusinessTag, int64, error) {
	var businessTags []*model.BusinessTag
	db := repo.db.WithContext(ctx)

	// 构建查询条件
	query := db.Model(&model.BusinessTag{})
	if options.Limit >= 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset >= 0 {
		query = query.Offset(options.Offset)
	}

	// 获取分页结果
	if err := query.Find(&businessTags).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list business tags: %w", err)
	}

	// 获取总数
	var count int64
	if err := repo.db.WithContext(ctx).Model(&model.BusinessTag{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count business tags: %w", err)
	}

	bizBusinessTags := make([]*biz.BusinessTag, 0, len(businessTags))
	for _, businessTag := range businessTags {
		bizBusinessTags = append(bizBusinessTags, repo.toBiz(businessTag))
	}

	return bizBusinessTags, count, nil
}
