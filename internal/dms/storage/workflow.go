package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.WorkflowRepo = (*WorkflowRepo)(nil)

type WorkflowRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewWorkflowRepo(log utilLog.Logger, s *Storage) *WorkflowRepo {
	return &WorkflowRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.workflow"))}
}

func (d *WorkflowRepo) SaveWorkflow(ctx context.Context, dataExportWorkflow *biz.Workflow) error {
	model := convertBizWorkflow(dataExportWorkflow)

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to save workflow: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (w *WorkflowRepo) UpdateWorkflowRecord(ctx context.Context, dataExportWorkflowRecord *biz.WorkflowRecord) error {
	model := convertBizWorkflowRecord(dataExportWorkflowRecord)

	if err := transaction(w.log, ctx, w.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to update workflow: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *WorkflowRepo) ListDataExportWorkflows(ctx context.Context, opt *biz.ListWorkflowsOption) (Workflows []*biz.Workflow, total int64, err error) {
	var models []*model.Workflow

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find models
		{
			db := tx.WithContext(ctx).Order(fmt.Sprintf("%s DESC", opt.OrderBy))
			db = gormPreload(ctx, db, opt.FilterBy)
			db = gormWheres(ctx, db, opt.FilterBy)
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list Workflows: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.Workflow{})
			db = gormPreload(ctx, db, opt.FilterBy)
			db = gormWheres(ctx, db, opt.FilterBy)
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count Workflows: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelWorkflow(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model Workflows: %v", err))
		}
		Workflows = append(Workflows, ds)
	}
	return Workflows, total, nil
}

func (d *WorkflowRepo) GetDataExportWorkflowsForView(ctx context.Context, userUid string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT DISTINCT w.uid
	FROM  workflows w
	left join workflow_records wr on w.workflow_record_uid  = wr.uid
	LEFT JOIN workflow_steps ws on wr.uid = ws.workflow_record_uid  and wr.uid  = ws.workflow_record_uid
	left join data_export_tasks det on JSON_SEARCH(wr.task_ids ,'one',det.uid) IS NOT NULL
	WHERE  JSON_SEARCH(ws.assignees,"one",?) IS NOT NULL
	UNION
	SELECT DISTINCT w.uid
	FROM  workflows w
	WHERE w.create_user_uid = ?
	`, userUid, userUid).Find(&workflowUids).Error; err != nil {
			return fmt.Errorf("failed to find workflow for view: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return workflowUids, nil
}

func (d *WorkflowRepo) GetDataExportWorkflowsByStatus(ctx context.Context, status string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
			SELECT w.uid
		FROM  workflows w
		left join workflow_records wr on w.workflow_record_uid  = wr.uid
		WHERE wr.status  = ?
	`, status).Find(&workflowUids).Error; err != nil {
			return fmt.Errorf("failed to get worfklow by status: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return workflowUids, nil
}

func (d *WorkflowRepo) GetDataExportWorkflowsByAssignUser(ctx context.Context, userUid string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
			SELECT DISTINCT w.uid
		FROM  workflows w
		left join workflow_records wr on w.workflow_record_uid  = wr.uid
		LEFT JOIN workflow_steps ws on wr.uid = ws.workflow_record_uid  and wr.current_workflow_step_id  = ws.step_id 
		left join data_export_tasks det on JSON_SEARCH(wr.task_ids ,'one',det.uid) IS NOT NULL
		WHERE  JSON_SEARCH(ws.assignees,"one",?) IS NOT NULL AND ws.state = "init"
	`, userUid).Find(&workflowUids).Error; err != nil {
			return fmt.Errorf("failed to find workflow by assignee user: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return workflowUids, nil
}

func (d *WorkflowRepo) GetDataExportWorkflowsByDBService(ctx context.Context, dbUid string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT w.uid
		FROM  workflows w
		left join workflow_records wr on w.workflow_record_uid  = wr.uid
		left join data_export_tasks det on JSON_SEARCH(wr.task_ids ,'one',det.uid) IS NOT NULL
		WHERE det.db_service_uid = ?
	`, dbUid).Find(&workflowUids).Error; err != nil {
			return fmt.Errorf("failed to find workflow by db uid: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return workflowUids, nil
}

func (d *WorkflowRepo) GetDataExportWorkflow(ctx context.Context, workflowUid string) (*biz.Workflow, error) {
	var workflow *model.Workflow
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Preload("WorkflowRecord").Preload("WorkflowRecord.Steps").First(&workflow, "uid = ?", workflowUid).Error; err != nil {
			return fmt.Errorf("failed to get workflow: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelWorkflow(workflow)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model workflow: %v", err))
	}
	return ret, nil
}

func (d *WorkflowRepo) AuditWorkflow(ctx context.Context, dataExportWorkflowUid string, status biz.DataExportWorkflowStatus, step *biz.WorkflowStep, operateId, reason string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.WorkflowRecord{}).Where("uid = ?", dataExportWorkflowUid).Update("status", status).Error; err != nil {
			return fmt.Errorf("failed to update workflow status, err: %v", err)
		}

		operateTime := time.Now()
		fields := map[string]interface{}{"operation_user_uid": operateId, "operate_at": operateTime, "reason": reason, "state": step.State}
		if err := tx.WithContext(ctx).Model(&model.WorkflowStep{}).Where("step_id = ? and workflow_record_uid = ?", step.StepId, step.WorkflowRecordUid).Updates(fields).Error; err != nil {
			return fmt.Errorf("failed to update workflow status, err: %v", err)
		}

		return nil
	})
}

func (d *WorkflowRepo) UpdateWorkflowStatusById(ctx context.Context, dataExportWorkflowUid string, status biz.DataExportWorkflowStatus) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.WorkflowRecord{}).Where("uid = ?", dataExportWorkflowUid).Update("status", status).Error; err != nil {
			return fmt.Errorf("failed to update workflow status, err: %v", err)
		}

		return nil
	})
}

func (d *WorkflowRepo) CancelWorkflow(ctx context.Context, workflowRecordIds []string, workflowSteps []*biz.WorkflowStep, operateId string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.WorkflowRecord{}).Where("uid in (?)", workflowRecordIds).Update("status", biz.DataExportWorkflowStatusCancel).Error; err != nil {
			return fmt.Errorf("failed to update workflow status, err: %v", err)
		}

		operateTime := time.Now()
		for _, step := range workflowSteps {
			fields := map[string]interface{}{"operation_user_uid": operateId, "operate_at": operateTime}
			if err := tx.WithContext(ctx).Model(&model.WorkflowStep{}).Where("step_id = ? and workflow_record_uid = ?", step.StepId, step.WorkflowRecordUid).Updates(fields).Error; err != nil {
				return fmt.Errorf("failed to update workflow status, err: %v", err)
			}
		}

		return nil
	})
}

func (d *WorkflowRepo) GetDataExportWorkflowsByIds(ctx context.Context, ids []string) ([]*biz.Workflow, error) {
	var workflows []*model.Workflow

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Preload("WorkflowRecord").Preload("WorkflowRecord.Steps").Find(&workflows, ids).Error; err != nil {
			return fmt.Errorf("failed to get Workflows: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.Workflow, 0, len(workflows))
	for _, workflow := range workflows {
		ds, err := convertModelWorkflow(workflow)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model Workflows: %v", err))
		}

		ret = append(ret, ds)
	}

	return ret, nil
}
