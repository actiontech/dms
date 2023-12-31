package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) ListDatabaseSourceService(ctx context.Context, req *v1.ListDatabaseSourceServicesReq, currentUserUid string) (reply *v1.ListDatabaseSourceServicesReply, err error) {
	conditions := make([]pkgConst.FilterCondition, 0)

	if req.ProjectUid != "" {
		conditions = append(conditions, pkgConst.FilterCondition{
			Field:    string(biz.DatabaseSourceServiceFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}

	services, err := d.DatabaseSourceServiceUsecase.ListDatabaseSourceServices(ctx, conditions, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*v1.ListDatabaseSourceService, 0, len(services))
	for _, service := range services {
		item := &v1.ListDatabaseSourceService{
			UID:        service.UID,
			ProjectUid: service.ProjectUID,
			DatabaseSourceService: v1.DatabaseSourceService{
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

	return &v1.ListDatabaseSourceServicesReply{
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

func (d *DMSService) GetDatabaseSourceService(ctx context.Context, req *v1.GetDatabaseSourceServiceReq, currentUserUid string) (reply *v1.GetDatabaseSourceServiceReply, err error) {
	service, err := d.DatabaseSourceServiceUsecase.GetDatabaseSourceService(ctx, req.DatabaseSourceServiceUid, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	item := &v1.GetDatabaseSourceService{
		UID:        service.UID,
		ProjectUid: service.ProjectUID,
		DatabaseSourceService: v1.DatabaseSourceService{
			Name:        service.Name,
			Source:      service.Source,
			Version:     service.Version,
			URL:         service.URL,
			DbType:      service.DbType,
			CronExpress: service.CronExpress,
			SQLEConfig:  d.buildReplySqleConfig(service.SQLEConfig),
		},
	}

	return &v1.GetDatabaseSourceServiceReply{
		Data: item,
	}, nil
}

func (d *DMSService) AddDatabaseSourceService(ctx context.Context, req *v1.AddDatabaseSourceServiceReq, currentUserId string) (reply *v1.AddDatabaseSourceServiceReply, err error) {

	databaseSourceParams := &biz.DatabaseSourceServiceParams{
		Name:        req.DatabaseSourceService.Name,
		Source:      req.DatabaseSourceService.Source,
		Version:     req.DatabaseSourceService.Version,
		URL:         req.DatabaseSourceService.URL,
		DbType:      req.DatabaseSourceService.DbType,
		CronExpress: req.DatabaseSourceService.CronExpress,
		ProjectUID:  req.ProjectUid,
		SQLEConfig:  d.buildSQLEConfig(req.DatabaseSourceService.SQLEConfig),
	}

	uid, err := d.DatabaseSourceServiceUsecase.AddDatabaseSourceService(ctx, databaseSourceParams, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("create database_source_service failed: %w", err)
	}

	return &v1.AddDatabaseSourceServiceReply{
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

func (d *DMSService) UpdateDatabaseSourceService(ctx context.Context, req *v1.UpdateDatabaseSourceServiceReq, currentUserId string) error {

	databaseSourceParams := &biz.DatabaseSourceServiceParams{
		Name:        req.DatabaseSourceService.Name,
		Source:      req.DatabaseSourceService.Source,
		Version:     req.DatabaseSourceService.Version,
		URL:         req.DatabaseSourceService.URL,
		DbType:      req.DatabaseSourceService.DbType,
		CronExpress: req.DatabaseSourceService.CronExpress,
		ProjectUID:  req.ProjectUid,
		SQLEConfig:  d.buildSQLEConfig(req.DatabaseSourceService.SQLEConfig),
	}

	err := d.DatabaseSourceServiceUsecase.UpdateDatabaseSourceService(ctx, req.DatabaseSourceServiceUid, databaseSourceParams, currentUserId)
	if err != nil {
		return fmt.Errorf("update database_source_service failed: %w", err)
	}

	return nil
}

func (d *DMSService) DeleteDatabaseSourceService(ctx context.Context, req *v1.DeleteDatabaseSourceServiceReq, currentUserId string) (err error) {
	err = d.DatabaseSourceServiceUsecase.DeleteDatabaseSourceService(ctx, req.DatabaseSourceServiceUid, currentUserId)
	if err != nil {
		return fmt.Errorf("delete database_source_service failed: %w", err)
	}

	return nil
}

func (d *DMSService) ListDatabaseSourceServiceTips(ctx context.Context) (*v1.ListDatabaseSourceServiceTipsReply, error) {
	sources, err := d.DatabaseSourceServiceUsecase.ListDatabaseSourceServiceTips(ctx)
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

	return &v1.ListDatabaseSourceServiceTipsReply{
		Data: ret,
	}, nil
}

func (d *DMSService) SyncDatabaseSourceService(ctx context.Context, req *v1.SyncDatabaseSourceServiceReq, currentUserId string) (err error) {
	err = d.DatabaseSourceServiceUsecase.SyncDatabaseSourceService(ctx, req.DatabaseSourceServiceUid, currentUserId)
	if err != nil {	
		d.log.Errorf("sync database_source_service failed: %w", err)
		return fmt.Errorf("sync database_source_service failed")
	}

	return nil
}
