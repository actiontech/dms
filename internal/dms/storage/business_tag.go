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

func (b *BusinessTagRepo) toModel(businessTag *biz.BusinessTag) *model.BusinessTag {
	return &model.BusinessTag{
		BusinessName: businessTag.BusinessTagName,
		Model:        model.Model{UID: businessTag.UID},
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
