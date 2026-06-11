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

var _ biz.OperationRecordRepo = (*operationRecordRepo)(nil)

// 内置系统账号产生的操作记录在列表/导出中不展示
const operationRecordUserNameSys = "sys"

type operationRecordRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOperationRecordRepo(log utilLog.Logger, s *Storage) biz.OperationRecordRepo {
	return &operationRecordRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.operationRecord"))}
}

func (d *operationRecordRepo) SaveOperationRecord(ctx context.Context, record *biz.OperationRecord) error {
	model := convertBizOperationRecord(record)

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to save operation record: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func applyOperationRecordBaseFilters(db *gorm.DB, opt *biz.ListOperationRecordOption) *gorm.DB {
	db = db.Where("operation_user_name <> ?", operationRecordUserNameSys)
	if opt.FilterOperateTimeFrom != "" {
		db = db.Where("operation_time > ?", opt.FilterOperateTimeFrom)
	}
	if opt.FilterOperateTimeTo != "" {
		db = db.Where("operation_time < ?", opt.FilterOperateTimeTo)
	}
	if opt.FilterOperateProjectName != nil {
		db = db.Where("operation_project_name = ?", *opt.FilterOperateProjectName)
	} else {
		if !opt.CanViewGlobal && len(opt.AccessibleProjectNames) > 0 {
			db = db.Where("operation_project_name IN ?", opt.AccessibleProjectNames)
		}
	}
	return db
}

func applyOperationRecordFilters(db *gorm.DB, opt *biz.ListOperationRecordOption) *gorm.DB {
	db = applyOperationRecordBaseFilters(db, opt)

	if opt.FuzzySearchOperateUserName != "" &&
		opt.FuzzySearchOperateUserName == opt.FuzzySearchOperateContent {
		kw := "%" + opt.FuzzySearchOperateUserName + "%"
		db = db.Where(
			"(operation_user_name LIKE ? OR operation_req_ip LIKE ? OR operation_i18n_content LIKE ?)",
			kw, kw, kw,
		)
	} else {
		if opt.FuzzySearchOperateUserName != "" {
			kw := "%" + opt.FuzzySearchOperateUserName + "%"
			db = db.Where("(operation_user_name LIKE ? OR operation_req_ip LIKE ?)", kw, kw)
		}
		if opt.FuzzySearchOperateContent != "" {
			db = db.Where("operation_i18n_content LIKE ?", "%"+opt.FuzzySearchOperateContent+"%")
		}
	}

	if opt.FilterFuzzyOperateUserName != "" {
		kw := "%" + opt.FilterFuzzyOperateUserName + "%"
		db = db.Where("(operation_user_name LIKE ? OR operation_req_ip LIKE ?)", kw, kw)
	}
	if len(opt.FilterOperateTypeNames) > 0 {
		db = db.Where("operation_type_name IN ?", opt.FilterOperateTypeNames)
	}
	if len(opt.FilterOperateActions) > 0 {
		db = db.Where("operation_action IN ?", opt.FilterOperateActions)
	}
	if opt.FilterOperateStatus != "" {
		db = db.Where("operation_status = ?", opt.FilterOperateStatus)
	}
	return db
}

func (d *operationRecordRepo) ListOperationRecords(ctx context.Context, opt *biz.ListOperationRecordOption) ([]*biz.OperationRecord, uint64, error) {
	var models []*model.OperationRecord
	var total int64

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Model(&model.OperationRecord{})
			db = applyOperationRecordFilters(db, opt)

			// Order and pagination
			db = db.Order("operation_time DESC")
			offset := (opt.PageIndex - 1) * opt.PageSize
			db = db.Limit(int(opt.PageSize)).Offset(int(offset))

			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list operation records: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.OperationRecord{})
			db = applyOperationRecordFilters(db, opt)

			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count operation records: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	ret := make([]*biz.OperationRecord, 0)
	for _, m := range models {
		record := convertModelOperationRecord(m)
		ret = append(ret, record)
	}

	return ret, uint64(total), nil
}

func (d *operationRecordRepo) ExportOperationRecords(ctx context.Context, opt *biz.ListOperationRecordOption) ([]*biz.OperationRecord, error) {
	var models []*model.OperationRecord

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		db := tx.WithContext(ctx).Model(&model.OperationRecord{})
		db = applyOperationRecordFilters(db, opt)

		// Order by time DESC, no pagination for export
		db = db.Order("operation_time DESC")

		if err := db.Find(&models).Error; err != nil {
			return fmt.Errorf("failed to export operation records: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.OperationRecord, 0)
	for _, m := range models {
		record := convertModelOperationRecord(m)
		ret = append(ret, record)
	}

	return ret, nil
}

func (d *operationRecordRepo) ListDistinctOperationUserNames(ctx context.Context, opt *biz.ListOperationRecordOption) ([]*biz.OperationRecordUserNameItem, error) {
	var userNames []string

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		db := tx.WithContext(ctx).Model(&model.OperationRecord{})
		db = applyOperationRecordBaseFilters(db, opt)
		if err := db.Select("DISTINCT operation_user_name").
			Order("operation_user_name ASC").
			Find(&userNames).Error; err != nil {
			return fmt.Errorf("failed to list distinct operation user names: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.OperationRecordUserNameItem, 0, len(userNames))
	for _, userName := range userNames {
		ret = append(ret, &biz.OperationRecordUserNameItem{
			OperationUserName: userName,
		})
	}
	return ret, nil
}

func (d *operationRecordRepo) CleanOperationRecordOpTimeBefore(ctx context.Context, t time.Time) (rowsAffected int64, err error) {
	err = transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Unscoped().Delete(&model.OperationRecord{}, "operation_time < ?", t)
		if err := result.Error; err != nil {
			return err
		}
		rowsAffected = result.RowsAffected
		return nil
	})
	return
}

func convertBizOperationRecord(src *biz.OperationRecord) *model.OperationRecord {
	return &model.OperationRecord{
		ID:                   src.ID,
		OperationTime:        src.OperationTime,
		OperationUserName:    src.OperationUserName,
		OperationReqIP:       src.OperationReqIP,
		OperationUserAgent:   src.OperationUserAgent,
		OperationTypeName:    src.OperationTypeName,
		OperationAction:      src.OperationAction,
		OperationProjectName: src.OperationProjectName,
		OperationStatus:      src.OperationStatus,
		OperationI18nContent: src.OperationI18nContent,
	}
}

func convertModelOperationRecord(model *model.OperationRecord) *biz.OperationRecord {
	return &biz.OperationRecord{
		ID:                   model.ID,
		OperationTime:        model.OperationTime,
		OperationUserName:    model.OperationUserName,
		OperationReqIP:       model.OperationReqIP,
		OperationUserAgent:   model.OperationUserAgent,
		OperationTypeName:    model.OperationTypeName,
		OperationAction:      model.OperationAction,
		OperationProjectName: model.OperationProjectName,
		OperationStatus:      model.OperationStatus,
		OperationI18nContent: model.OperationI18nContent,
	}
}
