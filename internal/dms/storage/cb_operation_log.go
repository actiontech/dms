package storage

import (
	"context"
	"fmt"
	"time"

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

func (d *CbOperationLogRepo) GetCbOperationLogByID(ctx context.Context, uid string) (*biz.CbOperationLog, error) {
	var model model.CbOperationLog
	if err := d.db.WithContext(ctx).Preload("User").Preload("DbService").Where("uid = ?", uid).First(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to get cb operation log by uid: %v", err)
	}

	operationLog, err := convertModelCbOperationLog(&model)
	if err != nil {
		return nil, err
	}

	return operationLog, nil
}

func (d *CbOperationLogRepo) ListCbOperationLogs(ctx context.Context, opt *biz.ListCbOperationLogOption) ([]*biz.CbOperationLog, int64, error) {
	var models []*model.CbOperationLog
	var total int64

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Preload("User").Preload("DbService")
			if opt.OrderBy != "" {
				db = db.Order(fmt.Sprintf("%s DESC", opt.OrderBy))
			}
			db = gormWheres(ctx, db, opt.FilterBy)
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list cb operation logs: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.CbOperationLog{})
			db = gormWheres(ctx, db, opt.FilterBy)
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count cb operation logs: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	ret := make([]*biz.CbOperationLog, 0)
	for _, m := range models {
		operationLog, err := convertModelCbOperationLog(m)
		if err != nil {
			return nil, 0, err
		}
		ret = append(ret, operationLog)
	}

	return ret, total, nil
}

func (d *CbOperationLogRepo) CleanCbOperationLogOpTimeBefore(ctx context.Context, t time.Time) (rowsAffected int64, err error) {
	err = transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Delete(&model.CbOperationLog{}, "op_time < ?", t)
		if err := result.Error; err != nil {
			return err
		}
		rowsAffected = result.RowsAffected
		return nil
	})
	return
}
