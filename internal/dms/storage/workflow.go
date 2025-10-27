package storage

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
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
	LEFT JOIN workflow_steps ws on wr.uid = ws.workflow_record_uid
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

func (d *WorkflowRepo) GetProjectDataExportWorkflowsForView(ctx context.Context, projectUid string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
	SELECT DISTINCT w.uid
	FROM  workflows w
	WHERE w.project_uid = ?
	`, projectUid).Find(&workflowUids).Error; err != nil {
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

func (d *WorkflowRepo) GetProjectDataExportWorkflowsByDBServices(ctx context.Context, dbUid []string, projectUid string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT w.uid
		FROM  workflows w
		left join workflow_records wr on w.workflow_record_uid  = wr.uid
		left join data_export_tasks det on JSON_SEARCH(wr.task_ids ,'one',det.uid) IS NOT NULL
		WHERE det.db_service_uid IN (?) and w.project_uid = ?
	`, dbUid, projectUid).Find(&workflowUids).Error; err != nil {
			return fmt.Errorf("failed to find workflow by db uid: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return workflowUids, nil
}

func (d *WorkflowRepo) GetDataExportWorkflowsByDBServices(ctx context.Context, dbUid []string) ([]string, error) {
	workflowUids := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
		SELECT w.uid
		FROM  workflows w
		left join workflow_records wr on w.workflow_record_uid  = wr.uid
		left join data_export_tasks det on JSON_SEARCH(wr.task_ids ,'one',det.uid) IS NOT NULL
		WHERE det.db_service_uid IN (?)
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

func (d *WorkflowRepo) DeleteDataExportWorkflowsByIds(ctx context.Context, dataExportWorkflowUids []string) error {
	if len(dataExportWorkflowUids) == 0 {
		return nil
	}
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		err := tx.Exec(`DELETE 
		FROM
			workflow_steps 
		WHERE
			workflow_record_uid IN (
			SELECT
				uid
			FROM
				workflow_records 
		WHERE
			workflow_uid IN ?
			)`, dataExportWorkflowUids).Error
		if err != nil {
			return err
		}

		err = tx.Exec("DELETE FROM workflow_records WHERE workflow_uid in ?", dataExportWorkflowUids).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Exec("DELETE FROM workflows WHERE uid in ?", dataExportWorkflowUids).Error
		if err != nil {
			return err
		}
		return nil
	})
}

var workflowsQueryTpl = `
SELECT 
	   w.project_uid,
	   p.priority AS project_priority,
	   p.name AS project_name,
       w.name AS subject,
       w.uid AS workflow_uid,
       w.desc,
       w.create_user_uid,
	   CAST("" AS DATETIME)											 AS create_user_deleted_at,
       w.created_at                                                  AS create_time,
       curr_ws.assignees											 AS current_step_assignee_user_id_list,
       curr_ws.state												 AS current_step_state,
       wr.status,
       wr.current_workflow_step_id									 AS current_workflow_step_id,
       wr.uid                                                        AS workflow_record_uid,
	   t.db_service_uid 											 AS db_service_uid,
	   ds.name 											    		 AS db_service_name
{{- template "body" . -}}

ORDER BY wr.updated_at DESC
{{- if .limit }}
LIMIT :limit OFFSET :offset
{{- end -}}
`

var workflowsQueryBodyTpl = `
{{ define "body" }}
FROM workflows w
INNER JOIN projects p ON w.project_uid = p.uid
INNER JOIN workflow_records AS wr ON w.workflow_record_uid = wr.uid
INNER JOIN data_export_tasks t ON JSON_CONTAINS(wr.task_ids, JSON_QUOTE(t.uid), '$')
LEFT JOIN workflow_steps AS curr_ws ON wr.uid = curr_ws.workflow_record_uid AND wr.current_workflow_step_id = curr_ws.step_id
LEFT JOIN db_services ds ON t.db_service_uid = ds.uid

WHERE
w.workflow_type='data_export' 
{{- if .check_user_can_access }}
AND (
w.create_user_uid = :current_user_id 
OR curr_ws.assignees REGEXP :current_user_id


{{- if .viewable_db_service_uids }} 
OR t.db_service_uid IN (:viewable_db_service_uids)
{{- end }}

)
{{- end }}

{{- if .filter_subject }}
AND w.name = :filter_subject
{{- end }}

{{- if .filter_by_create_user_uid }}
AND w.create_user_uid = :filter_by_create_user_uid
{{- end }}

{{- if .filter_create_user_id }}
AND w.create_user_id = :filter_create_user_id
{{- end }}

{{- if .filter_status }}
AND wr.status IN (:filter_status)
{{- end }}

{{- if .filter_current_step_assignee_user_id }}
AND curr_ws.assignees REGEXP :filter_current_step_assignee_user_id
{{- end }}

{{- if .filter_db_service_uid }}
AND t.db_service_uid = :filter_db_service_uid
{{- end }}

{{- if .filter_workflow_id }}
AND w.uid = :filter_workflow_id
{{- end }}

{{- if .filter_project_uid }}
AND w.project_uid = :filter_project_uid
{{- end }}

{{- if .filter_status_list }}
AND wr.status IN (:filter_status_list)
{{- end }}

{{- if .filter_project_uids }}
AND w.project_uid IN (:filter_project_uids)
{{- end }}

{{- if .fuzzy_keyword }}
AND (w.name like :fuzzy_keyword or w.uid like :fuzzy_keyword or w.desc like :fuzzy_keyword)
{{- end }}



{{ end }}

`

func (d *WorkflowRepo) GetGlobalWorkflowsByParameterMap(ctx context.Context, data map[string]interface{}) (workflows []*biz.Workflow, total int64, err error) {
	type workflowQueryResult struct {
		ProjectUID                    string        `gorm:"column:project_uid"`
		ProjectName                   string        `gorm:"column:project_name"`
		ProjectPriority               uint8         `gorm:"column:project_priority"`
		Subject                       string        `gorm:"column:subject"`
		WorkflowUID                   string        `gorm:"column:workflow_uid"`
		Desc                          string        `gorm:"column:desc"`
		CreateUserUID                 string        `gorm:"column:create_user_uid"`
		CreateUserDeletedAt           *time.Time    `gorm:"column:create_user_deleted_at"`
		CreateTime                    time.Time     `gorm:"column:create_time"`
		CurrentStepAssigneeUserIDList model.Strings `gorm:"column:current_step_assignee_user_id_list"`
		CurrentStepState              string        `gorm:"column:current_step_state"`
		Status                        string        `gorm:"column:status"`
		CurrentWorkflowStepId         uint64        `gorm:"column:current_workflow_step_id"`
		WorkflowRecordUID             string        `gorm:"column:workflow_record_uid"`
		DBServiceUID                  string        `gorm:"column:db_service_uid"`
		DBServiceName                 string        `gorm:"column:db_service_name"`
	}

	var results []workflowQueryResult
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find results
		{
			sqlQuery, params, err := d.buildWorkflowQuerySQL(data)
			if err != nil {
				return fmt.Errorf("failed to build workflow query SQL: %v", err)
			}

			if err := tx.WithContext(ctx).Raw(sqlQuery, params...).Scan(&results).Error; err != nil {
				return fmt.Errorf("failed to query workflows: %v", err)
			}
		}

		// find total
		{
			sqlQuery, params, err := d.buildWorkflowCountSQL(data)
			if err != nil {
				return fmt.Errorf("failed to build workflow count SQL: %v", err)
			}

			if err := tx.WithContext(ctx).Raw(sqlQuery, params...).Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count workflows: %v", err)
			}
		}

		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert results to biz.Workflow
	workflows = make([]*biz.Workflow, 0, len(results))
	for _, result := range results {
		workflow := &biz.Workflow{
			Base: biz.Base{
				CreatedAt: result.CreateTime,
			},
			UID:           result.WorkflowUID,
			Name:          result.Subject,
			ProjectUID:    result.ProjectUID,
			Desc:          result.Desc,
			CreateTime:    result.CreateTime,
			CreateUserUID: result.CreateUserUID,
			Status:        result.Status,
			ProjectInfo: &dmsCommonV1.ProjectInfo{
				ProjectUid:      result.ProjectUID,
				ProjectName:     result.ProjectName,
				ProjectPriority: dmsCommonV1.ToPriority(result.ProjectPriority),
			},
		}

		if result.DBServiceUID != "" {
			workflow.DBServiceInfos = []*dmsCommonV1.DBServiceUidWithNameInfo{
				{
					DBServiceUid:  result.DBServiceUID,
					DBServiceName: result.DBServiceName,
				},
			}
		}

		// Build WorkflowRecord with WorkflowSteps array
		workflow.WorkflowRecord = &biz.WorkflowRecord{
			UID:                   result.WorkflowRecordUID,
			Status:                biz.DataExportWorkflowStatus(result.Status),
			CurrentWorkflowStepId: result.CurrentWorkflowStepId,
			WorkflowSteps:         make([]*biz.WorkflowStep, result.CurrentWorkflowStepId),
		}

		// Fill the WorkflowSteps array with placeholder steps
		for i := uint64(0); i < result.CurrentWorkflowStepId; i++ {
			workflow.WorkflowRecord.WorkflowSteps[i] = &biz.WorkflowStep{
				StepId:            i + 1,
				WorkflowRecordUid: result.WorkflowRecordUID,
				State:             "",
				Assignees:         []string{},
			}
		}

		// Set the current step data
		if result.CurrentWorkflowStepId > 0 {
			currentStepIndex := result.CurrentWorkflowStepId - 1
			workflow.WorkflowRecord.WorkflowSteps[currentStepIndex] = &biz.WorkflowStep{
				StepId:            result.CurrentWorkflowStepId,
				WorkflowRecordUid: result.WorkflowRecordUID,
				State:             result.CurrentStepState,
				Assignees:         result.CurrentStepAssigneeUserIDList,
			}
			workflow.WorkflowRecord.CurrentStep = workflow.WorkflowRecord.WorkflowSteps[currentStepIndex]
		}

		workflows = append(workflows, workflow)
	}

	return workflows, total, nil
}

func (d *WorkflowRepo) buildWorkflowQuerySQL(data map[string]interface{}) (string, []interface{}, error) {
	return renderSQL(workflowsQueryTpl, workflowsQueryBodyTpl, data)
}

func (d *WorkflowRepo) buildWorkflowCountSQL(data map[string]interface{}) (string, []interface{}, error) {
	const countQueryTpl = `SELECT COUNT(*) as total {{- template "body" . -}}`

	templateData := make(map[string]interface{})
	for k, v := range data {
		templateData[k] = v
	}

	delete(templateData, "limit")
	delete(templateData, "offset")

	return renderSQL(countQueryTpl, workflowsQueryBodyTpl, templateData)
}

func renderSQL(mainTpl, bodyTpl string, params map[string]interface{}) (string, []interface{}, error) {
	// Parse templates together so that the main template can reference the body definition
	tmpl, err := template.New("workflows").Parse(bodyTpl)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse body template: %v", err)
	}

	if _, err = tmpl.Parse(mainTpl); err != nil {
		return "", nil, fmt.Errorf("failed to parse main template: %v", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, params); err != nil {
		return "", nil, fmt.Errorf("failed to execute template: %v", err)
	}

	query := buf.String()

	// Replace named params like :param with positional placeholders and collect args
	var args []interface{}
	re := regexp.MustCompile(`:([a-zA-Z_]+)`)
	query = re.ReplaceAllStringFunc(query, func(m string) string {
		key := m[1:] // Remove the leading ':'
		if v, ok := params[key]; ok {
			// Handle IN clause for slices
			if v != nil && reflect.TypeOf(v).Kind() == reflect.Slice {
				s := reflect.ValueOf(v)
				if s.Len() == 0 {
					// Empty slice, return NULL to avoid SQL error
					return "NULL"
				}
				placeholders := make([]string, s.Len())
				for i := 0; i < s.Len(); i++ {
					args = append(args, s.Index(i).Interface())
					placeholders[i] = "?"
				}
				// Return comma-separated placeholders
				// SQL template has: IN (:param), this becomes: IN (?, ?, ?)
				return strings.Join(placeholders, ", ")
			}
			args = append(args, v)
			return "?"
		}
		// If key not present, keep original (should not happen due to template conditionals)
		return m
	})

	return query, args, nil
}
