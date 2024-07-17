package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgAes "github.com/actiontech/dms/pkg/dms-common/pkg/aes"
)

func (d *DMSService) ListDBServiceSyncTask(ctx context.Context, currentUserUid string) (reply *v1.ListDBServiceSyncTasksReply, err error) {
	syncTasksParams, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTasks(ctx, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*v1.ListDBServiceSyncTask, 0, len(syncTasksParams))
	for _, params := range syncTasksParams {
		dbServiceSyncTask, err := toApiDBServiceSyncTask(d, params)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &v1.ListDBServiceSyncTask{
			UID:                 params.UID,
			DBServiceSyncTask:   *dbServiceSyncTask,
			LastSyncErr:         params.LastSyncErr,
			LastSyncSuccessTime: params.LastSyncSuccessTime,
		})
	}

	return &v1.ListDBServiceSyncTasksReply{
		Data: ret,
	}, nil
}

func (d *DMSService) GetDBServiceSyncTask(ctx context.Context, req *v1.GetDBServiceSyncTaskReq, currentUserUid string) (reply *v1.GetDBServiceSyncTaskReply, err error) {
	syncTaskParams, err := d.DBServiceSyncTaskUsecase.GetDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, currentUserUid)
	if nil != err {
		return nil, err
	}
	dbServiceSyncTask, err := toApiDBServiceSyncTask(d, syncTaskParams)
	if err != nil {
		return nil, err
	}
	item := &v1.GetDBServiceSyncTask{
		UID:               syncTaskParams.UID,
		DBServiceSyncTask: *dbServiceSyncTask,
	}

	return &v1.GetDBServiceSyncTaskReply{
		Data: item,
	}, nil
}

func (d *DMSService) AddDBServiceSyncTask(ctx context.Context, req *v1.AddDBServiceSyncTaskReq, currentUserId string) (reply *v1.AddDBServiceSyncTaskReply, err error) {
	syncTaskParams, err := toBizDBServiceSyncTaskParams(ctx, d, &req.DBServiceSyncTask)
	if err != nil {
		return nil, err
	}

	uid, err := d.DBServiceSyncTaskUsecase.AddDBServiceSyncTask(ctx, syncTaskParams, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("create db_service_sync_task failed: %w", err)
	}

	return &v1.AddDBServiceSyncTaskReply{
		Data: struct {
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) UpdateDBServiceSyncTask(ctx context.Context, req *v1.UpdateDBServiceSyncTaskReq, currentUserId string) error {
	syncTaskParams, err := toBizDBServiceSyncTaskParams(ctx, d, &req.DBServiceSyncTask)
	if err != nil {
		return err
	}

	err = d.DBServiceSyncTaskUsecase.UpdateDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, syncTaskParams, currentUserId)
	if err != nil {
		return fmt.Errorf("update db_service_sync_task failed: %w", err)
	}

	return nil
}

func toApiDBServiceSyncTask(d *DMSService, task *biz.DBServiceSyncTaskParams) (*v1.DBServiceSyncTask, error) {
	password, err := pkgAes.AesEncrypt(*task.DBServiceDefaultConfig.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}
	sqleConfig := &dmsCommonV1.SQLEConfig{
		RuleTemplateName: task.DBServiceDefaultConfig.RuleTemplateName,
		RuleTemplateID:   task.DBServiceDefaultConfig.RuleTemplateID,
		SQLQueryConfig:   &dmsCommonV1.SQLQueryConfig{},
	}
	if task.DBServiceDefaultConfig.SQLQueryConfig != nil {
		sqleConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel = dmsCommonV1.SQLAllowQueryAuditLevel(task.DBServiceDefaultConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel)
		sqleConfig.SQLQueryConfig.AuditEnabled = task.DBServiceDefaultConfig.SQLQueryConfig.AuditEnabled
		sqleConfig.SQLQueryConfig.MaxPreQueryRows = task.DBServiceDefaultConfig.SQLQueryConfig.MaxPreQueryRows
		sqleConfig.SQLQueryConfig.QueryTimeoutSecond = task.DBServiceDefaultConfig.SQLQueryConfig.QueryTimeoutSecond
	}
	dbServiceSyncTask := &v1.DBServiceSyncTask{
		Name:            task.Name,
		Source:          task.Source,
		URL:             task.URL,
		DbType:          task.DbType,
		CronExpress:     task.CronExpress,
		AdditionalParam: task.AdditionalParams,
		DBServiceDefaultConfig: v1.DBServiceDefaultConfig{
			Name:             task.DBServiceDefaultConfig.Name,
			Port:             task.DBServiceDefaultConfig.Port,
			Password:         password,
			Business:         task.DBServiceDefaultConfig.Business,
			MaintenanceTimes: d.convertPeriodToMaintenanceTime(task.DBServiceDefaultConfig.MaintenancePeriod),
			Desc:             *task.DBServiceDefaultConfig.Desc,
			SQLEConfig:       sqleConfig,
			IsEnableMasking:  task.DBServiceDefaultConfig.IsMaskingSwitch,
		},
	}

	if task.DBServiceDefaultConfig.AdditionalParams != nil {
		additionalParams := make([]*dmsCommonV1.AdditionalParam, 0, len(task.DBServiceDefaultConfig.AdditionalParams))
		for _, param := range task.DBServiceDefaultConfig.AdditionalParams {
			additionalParams = append(additionalParams, &dmsCommonV1.AdditionalParam{
				Name:        param.Key,
				Value:       param.Value,
				Description: param.Desc,
				Type:        string(param.Type),
			})
		}
		dbServiceSyncTask.DBServiceDefaultConfig.AdditionalParams = additionalParams
	}
	return dbServiceSyncTask, nil
}

func toBizDBServiceSyncTaskParams(ctx context.Context, d *DMSService, syncTask *v1.DBServiceSyncTask) (*biz.DBServiceSyncTaskParams, error) {
	additionalParams, err := d.DBServiceUsecase.GetDriverParamsByDBType(ctx, syncTask.DbType)
	if err != nil {
		return nil, err
	}
	for _, additionalParam := range syncTask.DBServiceDefaultConfig.AdditionalParams {
		err = additionalParams.SetParamValue(additionalParam.Name, additionalParam.Value)
		if err != nil {
			return nil, fmt.Errorf("set param value failed,invalid db type: %s", syncTask.DbType)
		}
	}
	args := &biz.BizDBServiceArgs{
		DBType:            syncTask.DbType,
		Name:              syncTask.DBServiceDefaultConfig.Name,
		Desc:              &syncTask.DBServiceDefaultConfig.Desc,
		Port:              syncTask.DBServiceDefaultConfig.Port,
		User:              syncTask.DBServiceDefaultConfig.User,
		Business:          syncTask.DBServiceDefaultConfig.Business,
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(syncTask.DBServiceDefaultConfig.MaintenanceTimes),
		Source:            syncTask.Source,
		AdditionalParams:  additionalParams,
	}

	if syncTask.DBServiceDefaultConfig.Password != "" {
		encryptedPassworld, err := pkgAes.AesEncrypt(syncTask.DBServiceDefaultConfig.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		args.Password = &encryptedPassworld
	}

	if biz.IsDMS() {
		args.IsMaskingSwitch = syncTask.DBServiceDefaultConfig.IsEnableMasking
	}

	sqleConfig := syncTask.DBServiceDefaultConfig.SQLEConfig
	if sqleConfig != nil {
		args.RuleTemplateName = sqleConfig.RuleTemplateName
		args.RuleTemplateID = sqleConfig.RuleTemplateID
		if sqleConfig.SQLQueryConfig != nil {
			args.SQLQueryConfig = &biz.SQLQueryConfig{
				MaxPreQueryRows:                  sqleConfig.SQLQueryConfig.MaxPreQueryRows,
				QueryTimeoutSecond:               sqleConfig.SQLQueryConfig.QueryTimeoutSecond,
				AuditEnabled:                     sqleConfig.SQLQueryConfig.AuditEnabled,
				AllowQueryWhenLessThanAuditLevel: string(sqleConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel),
			}
		}
	}
	return &biz.DBServiceSyncTaskParams{
		Name:                   syncTask.Name,
		Source:                 syncTask.Source,
		URL:                    syncTask.URL,
		DbType:                 syncTask.DbType,
		CronExpress:            syncTask.CronExpress,
		DBServiceDefaultConfig: args,
	}, nil
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
	v1Tips := make([]v1.DBServiceSyncTaskTip, 0, len(tips))
	for _, tip := range tips {
		v1Tips = append(v1Tips, convertDBServiceSyncTaskTips(tip))
	}
	return &v1.ListDBServiceSyncTaskTipsReply{
		Tips: v1Tips,
	}, nil
}

func convertDBServiceSyncTaskTips(meta biz.ListDBServiceSyncTaskTips) v1.DBServiceSyncTaskTip {
	return v1.DBServiceSyncTaskTip{
		Type:   meta.Type,
		Desc:   meta.Desc,
		DBType: meta.DBTypes,
		Params: meta.Params,
	}
}

func (d *DMSService) SyncDBServices(ctx context.Context, req *v1.SyncDBServicesReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.SyncDBServices(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {
		d.log.Errorf("sync db_service_sync_task failed: %v", err)
		return fmt.Errorf("sync db_service_sync_task failed")
	}

	return nil
}
