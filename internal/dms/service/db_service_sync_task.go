package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) AddDBServiceSyncTask(ctx context.Context, req *v1.AddDBServiceSyncTaskReq, currentUserId string) (reply *v1.AddDBServiceSyncTaskReply, err error) {
	syncTask := toBizDBServiceSyncTask(&req.DBServiceSyncTask)

	uid, err := d.DBServiceSyncTaskUsecase.AddDBServiceSyncTask(ctx, syncTask, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("create db_service_sync_task failed: %w", err)
	}

	return &v1.AddDBServiceSyncTaskReply{
		Data: struct {
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func toBizDBServiceSyncTask(syncTask *v1.DBServiceSyncTask) *biz.DBServiceSyncTask {
	ret := &biz.DBServiceSyncTask{
		Name:            syncTask.Name,
		Source:          syncTask.Source,
		URL:             syncTask.URL,
		DbType:          syncTask.DbType,
		CronExpress:     syncTask.CronExpress,
		AdditionalParam: syncTask.AdditionalParam,
	}
	if syncTask.SQLEConfig != nil {
		ret.SQLEConfig = &biz.SQLEConfig{
			AuditEnabled:               syncTask.SQLEConfig.AuditEnabled,
			RuleTemplateName:           syncTask.SQLEConfig.RuleTemplateName,
			RuleTemplateID:             syncTask.SQLEConfig.RuleTemplateID,
			DataExportRuleTemplateName: syncTask.SQLEConfig.DataExportRuleTemplateName,
			DataExportRuleTemplateID:   syncTask.SQLEConfig.DataExportRuleTemplateID,
		}
		if syncTask.SQLEConfig.SQLQueryConfig != nil {
			ret.SQLEConfig.SQLQueryConfig = &biz.SQLQueryConfig{
				MaxPreQueryRows:                  syncTask.SQLEConfig.SQLQueryConfig.MaxPreQueryRows,
				QueryTimeoutSecond:               syncTask.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond,
				AuditEnabled:                     syncTask.SQLEConfig.SQLQueryConfig.AuditEnabled,
				AllowQueryWhenLessThanAuditLevel: string(syncTask.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel),
				RuleTemplateID:                   syncTask.SQLEConfig.SQLQueryConfig.RuleTemplateID,
				RuleTemplateName:                 syncTask.SQLEConfig.SQLQueryConfig.RuleTemplateName,
			}
		}
	}
	if syncTask.AdditionalParam != nil {
		ret.AdditionalParam = syncTask.AdditionalParam
	}
	return ret
}

func (svc *DMSService) UpdateDBServiceSyncTask(ctx context.Context, req *v1.UpdateDBServiceSyncTaskReq, currentUserId string) error {
	syncTask := toBizDBServiceSyncTask(&req.DBServiceSyncTask)
	err := svc.DBServiceSyncTaskUsecase.UpdateDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, syncTask, currentUserId)
	if err != nil {
		return fmt.Errorf("update db_service_sync_task failed: %w", err)
	}
	return nil
}

func (d *DMSService) ListDBServiceSyncTask(ctx context.Context, currentUserUid string) (reply *v1.ListDBServiceSyncTasksReply, err error) {
	syncTasks, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTasks(ctx, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*v1.ListDBServiceSyncTask, 0, len(syncTasks))
	for _, syncTask := range syncTasks {
		ret = append(ret, &v1.ListDBServiceSyncTask{
			UID:                 syncTask.UID,
			DBServiceSyncTask:   toApiDBServiceSyncTask(d, syncTask),
			LastSyncErr:         syncTask.LastSyncErr,
			LastSyncSuccessTime: syncTask.LastSyncSuccessTime,
		})
	}

	return &v1.ListDBServiceSyncTasksReply{
		Data: ret,
	}, nil
}

func toApiDBServiceSyncTask(d *DMSService, syncTask *biz.DBServiceSyncTask) v1.DBServiceSyncTask {
	return v1.DBServiceSyncTask{
		Name:            syncTask.Name,
		Source:          syncTask.Source,
		URL:             syncTask.URL,
		DbType:          syncTask.DbType,
		CronExpress:     syncTask.CronExpress,
		AdditionalParam: syncTask.AdditionalParam,
		SQLEConfig:      d.buildReplySqleConfig(syncTask.SQLEConfig),
	}
}

func (d *DMSService) buildReplySqleConfig(params *biz.SQLEConfig) *dmsCommonV1.SQLEConfig {
	if params == nil {
		return nil
	}

	sqlConfig := &dmsCommonV1.SQLEConfig{
		AuditEnabled:               params.AuditEnabled,
		RuleTemplateName:           params.RuleTemplateName,
		RuleTemplateID:             params.RuleTemplateID,
		DataExportRuleTemplateName: params.DataExportRuleTemplateName,
		DataExportRuleTemplateID:   params.DataExportRuleTemplateID,
		SQLQueryConfig:             &dmsCommonV1.SQLQueryConfig{},
	}
	if params.SQLQueryConfig != nil {
		sqlConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel = dmsCommonV1.SQLAllowQueryAuditLevel(params.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel)
		sqlConfig.SQLQueryConfig.AuditEnabled = params.SQLQueryConfig.AuditEnabled
		sqlConfig.SQLQueryConfig.MaxPreQueryRows = params.SQLQueryConfig.MaxPreQueryRows
		sqlConfig.SQLQueryConfig.QueryTimeoutSecond = params.SQLQueryConfig.QueryTimeoutSecond
		sqlConfig.SQLQueryConfig.RuleTemplateID = params.SQLQueryConfig.RuleTemplateID
		sqlConfig.SQLQueryConfig.RuleTemplateName = params.SQLQueryConfig.RuleTemplateName
	}

	return sqlConfig
}

func (d *DMSService) ListDBServiceSyncTaskTips(ctx context.Context) (*v1.ListDBServiceSyncTaskTipsReply, error) {
	tips, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTaskTips(ctx)
	if err != nil {
		return nil, fmt.Errorf("list db_service_sync_task tips failed: %w", err)
	}
	data := make([]v1.DBServiceSyncTaskTip, 0, len(tips))
	for _, tip := range tips {
		data = append(data, toApiDBServiceSyncTaskTips(tip))
	}
	return &v1.ListDBServiceSyncTaskTipsReply{
		Data: data,
	}, nil
}

func toApiDBServiceSyncTaskTips(meta biz.ListDBServiceSyncTaskTips) v1.DBServiceSyncTaskTip {
	return v1.DBServiceSyncTaskTip{
		Type:   meta.Type,
		Desc:   meta.Desc,
		DBType: meta.DBTypes,
		Params: meta.Params,
	}
}

func (d *DMSService) GetDBServiceSyncTask(ctx context.Context, req *v1.GetDBServiceSyncTaskReq, currentUserUid string) (reply *v1.GetDBServiceSyncTaskReply, err error) {
	syncTask, err := d.DBServiceSyncTaskUsecase.GetDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	item := &v1.GetDBServiceSyncTask{
		UID:               syncTask.UID,
		DBServiceSyncTask: toApiDBServiceSyncTask(d, syncTask),
	}

	return &v1.GetDBServiceSyncTaskReply{
		Data: item,
	}, nil
}

func (d *DMSService) DeleteDBServiceSyncTask(ctx context.Context, req *v1.DeleteDBServiceSyncTaskReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.DeleteDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {
		return fmt.Errorf("delete db_service_sync_task failed: %w", err)
	}

	return nil
}

func (d *DMSService) SyncDBServices(ctx context.Context, req *v1.SyncDBServicesReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.SyncDBServices(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {
		d.log.Errorf("sync db_service_sync_task failed: %v", err)
		return fmt.Errorf("sync db_service_sync_task failed: %v", err)
	}

	return nil
}
