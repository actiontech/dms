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

var _ biz.EnvironmentTagRepo = (*EnvironmentTagRepo)(nil)

type EnvironmentTagRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewEnvironmentTagRepo(log utilLog.Logger, s *Storage) *EnvironmentTagRepo {
	return &EnvironmentTagRepo{
		Storage: s,
		log:     utilLog.NewHelper(log, utilLog.WithMessageKey("storage.environment_tag")),
	}
}

func (repo *EnvironmentTagRepo) toModel(environmentTag *biz.EnvironmentTag) *model.EnvironmentTag {
	return &model.EnvironmentTag{
		EnvironmentName: environmentTag.Name,
		Model:           model.Model{UID: environmentTag.UID},
		ProjectUID:      environmentTag.ProjectUID,
	}
}

func (repo *EnvironmentTagRepo) toBiz(environmentTag *model.EnvironmentTag) *biz.EnvironmentTag {
	return &biz.EnvironmentTag{
		Name:       environmentTag.EnvironmentName,
		UID:        environmentTag.UID,
		ProjectUID: environmentTag.ProjectUID,
	}
}

func (repo *EnvironmentTagRepo) CreateEnvironmentTag(ctx context.Context, environmentTag *biz.EnvironmentTag) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(repo.toModel(environmentTag)).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to create environment tag: %v", err))
		}
		return nil
	})
}

func (repo *EnvironmentTagRepo) UpdateEnvironmentTag(ctx context.Context, environmentTagUID, environmentTagName string) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.EnvironmentTag{}).Where("uid = ?", environmentTagUID).Update("environment_name", environmentTagName).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to update environment tag: %v", err))
		}
		return nil
	})
}

func (repo *EnvironmentTagRepo) DeleteEnvironmentTag(ctx context.Context, environmentTagUID string) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", environmentTagUID).Delete(&model.EnvironmentTag{}).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to delete environment tag: %v", err))
		}
		return nil
	})
}

func (repo *EnvironmentTagRepo) GetEnvironmentTagByName(ctx context.Context, projetUid, name string) (bool, *biz.EnvironmentTag, error) {
	var environmentTag model.EnvironmentTag
	if err := repo.db.WithContext(ctx).Where("environment_name = ? AND project_uid = ?", name, projetUid).First(&environmentTag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to get environment tag by name: %w", err)
	}
	return true, repo.toBiz(&environmentTag), nil
}

func (repo *EnvironmentTagRepo) GetEnvironmentTagByUID(ctx context.Context, uid string) (*biz.EnvironmentTag, error) {
	var environmentTag model.EnvironmentTag
	if err := repo.db.WithContext(ctx).Where("uid = ?", uid).First(&environmentTag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgErr.ErrStorageNoData
		}
		return nil, fmt.Errorf("failed to get environment tag by uid: %w", err)
	}
	return repo.toBiz(&environmentTag), nil
}
func (repo *EnvironmentTagRepo) ListEnvironmentTags(ctx context.Context, options *biz.ListEnvironmentTagsOption) ([]*biz.EnvironmentTag, int64, error) {
	var environmentTags []*model.EnvironmentTag
	db := repo.db.WithContext(ctx)

	// 构建查询条件
	query := db.Model(&model.EnvironmentTag{})
	if options.Limit >= 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset >= 0 {
		query = query.Offset(options.Offset)
	}

	// 获取分页结果
	if err := query.Where("project_uid = ?", options.ProjectUID).Find(&environmentTags).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list environment tags: %w", err)
	}

	// 获取总数
	var count int64
	if err := repo.db.WithContext(ctx).Model(&model.EnvironmentTag{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count environment tags: %w", err)
	}

	bizEnvironmentTags := make([]*biz.EnvironmentTag, 0, len(environmentTags))
	for _, EnvironmentTag := range environmentTags {
		bizEnvironmentTags = append(bizEnvironmentTags, repo.toBiz(EnvironmentTag))
	}

	return bizEnvironmentTags, count, nil
}
