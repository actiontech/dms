package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.CbOperationLogRepo = (*CbOperationLogRepo)(nil)

type CbOperationLogRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewCbOperationLogRepo(log utilLog.Logger, s *Storage) *CbOperationLogRepo {
	return &CbOperationLogRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.cbOperationLog"))}
}

func (d *CbOperationLogRepo) SaveCbOperationLog(ctx context.Context, log *biz.CbOperationLog) error {
	model := convertBizCbOperationLog(log)

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save cb operation log: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *CbOperationLogRepo) UpdateCbOperationLog(ctx context.Context, operationLog *biz.CbOperationLog) error {
	opLog := convertBizCbOperationLog(operationLog)

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.CbOperationLog{}).Where("uid = ?", opLog.UID).Omit("created_at").Save(opLog).Error; err != nil {
			return fmt.Errorf("failed to update cb operation log: %v", err)
		}
		return nil
	})
}
