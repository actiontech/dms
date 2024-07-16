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

	services, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTasks(ctx, conditions, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*v1.ListDBServiceSyncTask, 0, len(services))
	for _, service := range services {
		item := &v1.ListDBServiceSyncTask{
			UID:        service.UID,
			ProjectUid: service.ProjectUID,
			DBServiceSyncTask: v1.DBServiceSyncTask{
				Name:        service.Name,
				Source:      service.Source,
				Version:     service.Version,
				URL:         service.URL,
				DbType:      service.DbType,
				CronExpress: service.CronExpress,
				SQLEConfig:  d.buildReplySqleConfig(service.SQLEConfig),
			},
			LastSyncErr:         service.LastSyncErr,
			LastSyncSuccessTime: service.LastSyncSuccessTime,
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
	service, err := d.DBServiceSyncTaskUsecase.GetDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	item := &v1.GetDBServiceSyncTask{
		UID:        service.UID,
		ProjectUid: service.ProjectUID,
		DBServiceSyncTask: v1.DBServiceSyncTask{
			Name:        service.Name,
			Source:      service.Source,
			Version:     service.Version,
			URL:         service.URL,
			DbType:      service.DbType,
			CronExpress: service.CronExpress,
			SQLEConfig:  d.buildReplySqleConfig(service.SQLEConfig),
		},
	}

	return &v1.GetDBServiceSyncTaskReply{
		Data: item,
	}, nil
}

func (d *DMSService) AddDBServiceSyncTask(ctx context.Context, req *v1.AddDBServiceSyncTaskReq, currentUserId string) (reply *v1.AddDBServiceSyncTaskReply, err error) {

	databaseSourceParams := &biz.DBServiceSyncTaskParams{
		Name:        req.DBServiceSyncTask.Name,
		Source:      req.DBServiceSyncTask.Source,
		Version:     req.DBServiceSyncTask.Version,
		URL:         req.DBServiceSyncTask.URL,
		DbType:      req.DBServiceSyncTask.DbType,
		CronExpress: req.DBServiceSyncTask.CronExpress,
		ProjectUID:  req.ProjectUid,
		SQLEConfig:  d.buildSQLEConfig(req.DBServiceSyncTask.SQLEConfig),
	}

	uid, err := d.DBServiceSyncTaskUsecase.AddDBServiceSyncTask(ctx, databaseSourceParams, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("create database_source_service failed: %w", err)
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

	databaseSourceParams := &biz.DBServiceSyncTaskParams{
		Name:        req.DBServiceSyncTask.Name,
		Source:      req.DBServiceSyncTask.Source,
		Version:     req.DBServiceSyncTask.Version,
		URL:         req.DBServiceSyncTask.URL,
		DbType:      req.DBServiceSyncTask.DbType,
		CronExpress: req.DBServiceSyncTask.CronExpress,
		ProjectUID:  req.ProjectUid,
		SQLEConfig:  d.buildSQLEConfig(req.DBServiceSyncTask.SQLEConfig),
	}

	err := d.DBServiceSyncTaskUsecase.UpdateDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, databaseSourceParams, currentUserId)
	if err != nil {
		return fmt.Errorf("update database_source_service failed: %w", err)
	}

	return nil
}

func (d *DMSService) DeleteDBServiceSyncTask(ctx context.Context, req *v1.DeleteDBServiceSyncTaskReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.DeleteDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {
		return fmt.Errorf("delete database_source_service failed: %w", err)
	}

	return nil
}

func (d *DMSService) ListDBServiceSyncTaskTips(ctx context.Context) (*v1.ListDBServiceSyncTaskTipsReply, error) {
	sources, err := d.DBServiceSyncTaskUsecase.ListDBServiceSyncTaskTips(ctx)
	if err != nil {
		return nil, fmt.Errorf("list database_source_service tips failed: %w", err)
	}

	ret := make([]*v1.DatabaseSource, 0, len(sources))
	for _, item := range sources {
		dbTypes := make([]string, 0, len(item.DbTypes))
		for _, val := range item.DbTypes {
			dbTypes = append(dbTypes, string(val))
		}

		ret = append(ret, &v1.DatabaseSource{
			DbTypes: dbTypes,
			Source:  string(item.Source),
		})
	}

	return &v1.ListDBServiceSyncTaskTipsReply{
		Data: ret,
	}, nil
}

func (d *DMSService) SyncDBServiceSyncTask(ctx context.Context, req *v1.SyncDBServiceSyncTaskReq, currentUserId string) (err error) {
	err = d.DBServiceSyncTaskUsecase.SyncDBServiceSyncTask(ctx, req.DBServiceSyncTaskUid, currentUserId)
	if err != nil {	
		d.log.Errorf("sync database_source_service failed: %w", err)
		return fmt.Errorf("sync database_source_service failed")
	}

	return nil
}
