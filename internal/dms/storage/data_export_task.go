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

var _ biz.DataExportTaskRepo = (*DataExportTaskRepo)(nil)

type DataExportTaskRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewDataExportTaskRepo(log utilLog.Logger, s *Storage) *DataExportTaskRepo {
	return &DataExportTaskRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.dataExportTask"))}
}

func (d *DataExportTaskRepo) SaveDataExportTask(ctx context.Context, dataExportDataExportTasks []*biz.DataExportTask) error {
	models := make([]*model.DataExportTask, 0)
	for _, dataExportDataExportTask := range dataExportDataExportTasks {
		models = append(models, convertBizDataExportTask(dataExportDataExportTask))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(models).Error; err != nil {
			return fmt.Errorf("failed to save data export task: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *DataExportTaskRepo) GetDataExportTaskByIds(ctx context.Context, ids []string) (dataExportDataExportTasks []*biz.DataExportTask, err error) {
	tasks := make([]*model.DataExportTask, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Preload("DataExportTaskRecords").Find(&tasks, "uid in (?)", ids).Error; err != nil {
			return fmt.Errorf("failed to get data export tasks: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.DataExportTask, 0)
	for _, v := range tasks {
		ret = append(ret, convertModelDataExportTask(v))
	}

	return ret, nil
}

func (d *DataExportTaskRepo) ListDataExportTaskRecord(ctx context.Context, opt *biz.ListDataExportTaskRecordOption) (exportTaskRecords []*biz.DataExportTaskRecord, total int64, err error) {
	var models []*model.DataExportTaskRecord

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			db = gormWheres(ctx, db, opt.FilterBy)
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list data export task records: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.DataExportTaskRecord{})
			db = gormWheres(ctx, db, opt.FilterBy)
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count data export task records: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		exportTaskRecords = append(exportTaskRecords, convertModelDataExportTaskRecords(model))
	}
	return exportTaskRecords, total, nil
}

func (d *DataExportTaskRepo) BatchUpdateDataExportTaskStatusByIds(ctx context.Context, ids []string, status biz.DataExportTaskStatus) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DataExportTask{}).Where("uid in (?)", ids).Update("export_status", status).Error; err != nil {
			return fmt.Errorf("failed to update data export task status: %v", err)
		}

		return nil
	})
}

func (d *DataExportTaskRepo) ListDataExportTasks(ctx context.Context, opt *biz.ListDataExportTaskOption) (exportTasks []*biz.DataExportTask, total int64, err error) {
	var models []*model.DataExportTask

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			db = gormWheres(ctx, db, opt.FilterBy)
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list data export task: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.DataExportTask{})
			db = gormWheres(ctx, db, opt.FilterBy)
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count data export task: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		exportTasks = append(exportTasks, convertModelDataExportTask(model))
	}
	return exportTasks, total, nil
}

func (d *DataExportTaskRepo) DeleteUnusedDataExportTasks(ctx context.Context) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.Exec(`DELETE det,detr
		FROM data_export_tasks det
		LEFT JOIN data_export_task_records detr ON det.uid = detr.data_export_task_id
		WHERE det.uid NOT IN (
			SELECT uid FROM (
				SELECT det.uid
				FROM data_export_tasks det
				JOIN workflow_records wr ON JSON_SEARCH(wr.task_ids, "one", det.uid) IS NOT NULL
			) AS subquery
		)
		AND det.created_at < ?
			`, time.Now().Add(-time.Hour*24)).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (d *DataExportTaskRepo) BatchUpdateDataExportTaskByIds(ctx context.Context, ids []string, args map[string]interface{}) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DataExportTask{}).Where("uid in (?)", ids).Updates(args).Error; err != nil {
			return fmt.Errorf("failed to update data export task: %v", err)
		}
		return nil
	})
}

func (d *DataExportTaskRepo) SaveDataExportTaskRecords(ctx context.Context, dataExportTaskRecords []*biz.DataExportTaskRecord) error {
	models := make([]*model.DataExportTaskRecord, 0)
	for _, dataExportTaskRecord := range dataExportTaskRecords {
		models = append(models, convertBizDataExportTaskRecords(dataExportTaskRecord))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(models).Error; err != nil {
			return fmt.Errorf("failed to save data export task records: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
