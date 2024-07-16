package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) ListDBServiceSyncTask(ctx context.Context, req *v1.ListDBServiceSyncTasksReq, currentUserUid string) (reply *v1.ListDBServiceSyncTasksReply, err error) {
	conditions := make([]pkgConst.FilterCondition, 0)

	if req.ProjectUid != "" {
		conditions = append(conditions, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceSyncTaskFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	syncTasks, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTasks(ctx, conditions, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*v1.ListDBServiceSyncTask, 0, len(syncTasks))
	for _, task := range syncTasks {
		item := &v1.ListDBServiceSyncTask{
			UID:        task.UID,
			ProjectUid: task.ProjectUID,
			DBServiceSyncTask: v1.DBServiceSyncTask{
				Name:        task.Name,
				Source:      task.Source,
				Version:     task.Version,
				URL:         task.URL,
				DbType:      task.DbType,
				CronExpress: task.CronExpress,
				SQLEConfig:  d.buildReplySqleConfig(task.SQLEConfig),
			},
			LastSyncErr:         task.LastSyncErr,
			LastSyncSuccessTime: task.LastSyncSuccessTime,
		}

		ret = append(ret, item)
	}

	return &v1.ListDBServiceSyncTasksReply{
		Data: ret,
	}, nil
}

func (d *DMSService) buildReplySqleConfig(params *biz.SQLEConfig) *dmsCommonV1.SQLEConfig {
	if params == nil {
		return nil
	}

	sqlConfig := &dmsCommonV1.SQLEConfig{
		RuleTemplateName: params.RuleTemplateName,
		RuleTemplateID:   params.RuleTemplateID,
		SQLQueryConfig:   &dmsCommonV1.SQLQueryConfig{},
	}
	if params.SQLQueryConfig != nil {
		sqlConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel = dmsCommonV1.SQLAllowQueryAuditLevel(params.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel)
		sqlConfig.SQLQueryConfig.AuditEnabled = params.SQLQueryConfig.AuditEnabled
		sqlConfig.SQLQueryConfig.MaxPreQueryRows = params.SQLQueryConfig.MaxPreQueryRows
		sqlConfig.SQLQueryConfig.QueryTimeoutSecond = params.SQLQueryConfig.QueryTimeoutSecond
	}

	return sqlConfig
}

func (d *DMSService) GetDBServiceSyncTask(ctx context.Context, req *v1.GetDBServiceSyncTaskReq, currentUserUid string) (reply *v1.GetDBServiceSyncTaskReply, err error) {
	syncTask, err := d.DBServiceSyncTaskUsecase.GetDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	item := &v1.GetDBServiceSyncTask{
		UID:        syncTask.UID,
		ProjectUid: syncTask.ProjectUID,
		DBServiceSyncTask: v1.DBServiceSyncTask{
			Name:        syncTask.Name,
			Source:      syncTask.Source,
			Version:     syncTask.Version,
			URL:         syncTask.URL,
			DbType:      syncTask.DbType,
			CronExpress: syncTask.CronExpress,
			SQLEConfig:  d.buildReplySqleConfig(syncTask.SQLEConfig),
		},
	}

	return &v1.GetDBServiceSyncTaskReply{
		Data: item,
	}, nil
}

func (d *DMSService) AddDBServiceSyncTask(ctx context.Context, req *v1.AddDBServiceSyncTaskReq, currentUserId string) (reply *v1.AddDBServiceSyncTaskReply, err error) {

	dbServiceTaskParams := &biz.DBServiceSyncTaskParams{
		Name:        req.DBServiceSyncTask.Name,
		Source:      req.DBServiceSyncTask.Source,
		Version:     req.DBServiceSyncTask.Version,
		URL:         req.DBServiceSyncTask.URL,
		DbType:      req.DBServiceSyncTask.DbType,
		CronExpress: req.DBServiceSyncTask.CronExpress,
		ProjectUID:  req.ProjectUid,
		SQLEConfig:  d.buildSQLEConfig(req.DBServiceSyncTask.SQLEConfig),
	}

	uid, err := d.DBServiceSyncTaskUsecase.AddDBServiceSyncTask(ctx, dbServiceTaskParams, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("create db_service_sync_task failed: %w", err)
	}

	return &v1.AddDBServiceSyncTaskReply{
		Data: struct {
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) buildSQLEConfig(params *dmsCommonV1.SQLEConfig) *biz.SQLEConfig {
	if params == nil {
		return nil
	}

	sqleConf := &biz.SQLEConfig{
		RuleTemplateName: params.RuleTemplateName,
		RuleTemplateID:   params.RuleTemplateID,
	}

	if params.SQLQueryConfig != nil {
		sqleConf.SQLQueryConfig = &biz.SQLQueryConfig{
			MaxPreQueryRows:                  params.SQLQueryConfig.MaxPreQueryRows,
			QueryTimeoutSecond:               params.SQLQueryConfig.QueryTimeoutSecond,
			AuditEnabled:                     params.SQLQueryConfig.AuditEnabled,
			AllowQueryWhenLessThanAuditLevel: string(params.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel),
		}
	}

	return sqleConf
}

func (d *DMSService) UpdateDBServiceSyncTask(ctx context.Context, req *v1.UpdateDBServiceSyncTaskReq, currentUserId string) error {

	dbServiceTaskParams := &biz.DBServiceSyncTaskParams{
		Name:        req.DBServiceSyncTask.Name,
		Source:      req.DBServiceSyncTask.Source,
		Version:     req.DBServiceSyncTask.Version,
		URL:         req.DBServiceSyncTask.URL,
		DbType:      req.DBServiceSyncTask.DbType,
		CronExpress: req.DBServiceSyncTask.CronExpress,
		ProjectUID:  req.ProjectUid,
		SQLEConfig:  d.buildSQLEConfig(req.DBServiceSyncTask.SQLEConfig),
	}

	err := d.DBServiceSyncTaskUsecase.UpdateDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, dbServiceTaskParams, currentUserId)
	if err != nil {
		return fmt.Errorf("update db_service_sync_task failed: %w", err)
	}

	return nil
}

func (d *DMSService) DeleteDBServiceSyncTask(ctx context.Context, req *v1.DeleteDBServiceSyncTaskReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.DeleteDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {
		return fmt.Errorf("delete db_service_sync_task failed: %w", err)
	}

	return nil
}

func (d *DMSService) ListDBServiceSyncTaskTips(ctx context.Context) (*v1.ListDBServiceSyncTaskTipsReply, error) {
	tips, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTaskTips(ctx)
	if err != nil {
		return nil, fmt.Errorf("list db_service_sync_task tips failed: %w", err)
	}
	v1Tips := make([]v1.DBServiceSyncTaskTip , 0, len(tips))
	for _, tip := range tips {
		v1Tips = append(v1Tips, convertDBServiceSyncTaskTips(tip))
	}
	return &v1.ListDBServiceSyncTaskTipsReply{
		Tips: v1Tips,
	}, nil
}

func convertDBServiceSyncTaskTips(meta biz.ListDBServiceSyncTaskTips) v1.DBServiceSyncTaskTip  {
	return v1.DBServiceSyncTaskTip {
		Type:   meta.Type,
		Desc:   meta.Desc,
		DBType: meta.DBTypes,
		Params: meta.Params,
	}
}

func (d *DMSService) SyncDBServices(ctx context.Context, req *v1.SyncDBServicesReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.SyncDBServices(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {
		d.log.Errorf("sync db_service_sync_task failed: %w", err)
		return fmt.Errorf("sync db_service_sync_task failed")
	}

	return nil
}
