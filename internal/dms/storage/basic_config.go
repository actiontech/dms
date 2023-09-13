package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.BasicConfigRepo = (*BasicConfigRepo)(nil)

type BasicConfigRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewBasicConfigRepo(log utilLog.Logger, s *Storage) *BasicConfigRepo {
	return &BasicConfigRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.basic_config"))}
}

func (d *BasicConfigRepo) GetBasicConfig(ctx context.Context) (*biz.BasicConfigParams, error) {
	var basicConfig model.BasicConfig
	err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Find(&basicConfig).Error; err != nil {
			return fmt.Errorf("failed to get basic_config: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return convertModelBasicConfig(&basicConfig), nil
}

func (d *BasicConfigRepo) SaveBasicConfig(ctx context.Context, params *biz.BasicConfigParams) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(convertBizBasicConfig(params)).Error; err != nil {
			return fmt.Errorf("failed to save basic config err: %v", err)
		}

		return nil
	})
}
