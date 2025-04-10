package service

import (
	"context"
	"fmt"

	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	dmsV2 "github.com/actiontech/dms/api/dms/service/v2"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	dmsCommonV2 "github.com/actiontech/dms/pkg/dms-common/api/dms/v2"
	pkgAes "github.com/actiontech/dms/pkg/dms-common/pkg/aes"
	"github.com/actiontech/dms/pkg/params"
	"github.com/actiontech/dms/pkg/periods"
	"github.com/go-openapi/strfmt"
)

func (d *DMSService) DelDBService(ctx context.Context, req *dmsV1.DelDBServiceReq, currentUserUid string) (err error) {
	d.log.Infof("DelDBService.req=%v", req)
	defer func() {
		d.log.Infof("DelDBService.req=%v;error=%v", req, err)
	}()

	if err := d.DBServiceUsecase.DelDBService(ctx, req.DBServiceUid, currentUserUid); err != nil {
		return fmt.Errorf("delete db service failed: %v", err)
	}

	return nil
}

func (d *DMSService) UpdateDBService(ctx context.Context, req *dmsV2.UpdateDBServiceReq, currentUserUid string) (err error) {
	d.log.Infof("UpdateDBService.req=%v", req)
	defer func() {
		d.log.Infof("UpdateDBService.req=%v;error=%v", req, err)
	}()

	additionalParams, err := d.DBServiceUsecase.GetDriverParamsByDBType(ctx, req.DBService.DBType)
	if err != nil {
		return err
	}
	for _, additionalParam := range req.DBService.AdditionalParams {
		err = additionalParams.SetParamValue(additionalParam.Name, additionalParam.Value)
		if err != nil {
			return fmt.Errorf("set param value failed,invalid db type: %s", req.DBService.DBType)
		}
	}

	args := &biz.BizDBServiceArgs{
		DBType:            req.DBService.DBType,
		Desc:              req.DBService.Desc,
		Host:              req.DBService.Host,
		Port:              req.DBService.Port,
		User:              req.DBService.User,
		Password:          req.DBService.Password,
		EnvironmentTagUID: req.DBService.EnvironmentTagUID,
		EnableBackup:      req.DBService.EnableBackup,
		BackupMaxRows:     autoChooseBackupMaxRows(req.DBService.EnableBackup, req.DBService.BackupMaxRows),
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(req.DBService.MaintenanceTimes),
		AdditionalParams:  additionalParams,
	}

	if biz.IsDMS() {
		args.IsMaskingSwitch = req.DBService.IsEnableMasking
	}

	sqleConfig := req.DBService.SQLEConfig
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
	if err := d.DBServiceUsecase.UpdateDBServiceByArgs(ctx, req.DBServiceUid, args, currentUserUid); err != nil {
		return fmt.Errorf("update db service failed: %v", err)
	}

	return nil
}

func (d *DMSService) CheckDBServiceIsConnectable(ctx context.Context, req *dmsV1.CheckDBServiceIsConnectableReq) (reply *dmsV1.CheckDBServiceIsConnectableReply, err error) {
	results, err := d.DBServiceUsecase.IsConnectable(ctx, req.DBService)

	if err != nil {
		d.log.Errorf("IsConnectable err: %v", err)
		return nil, err
	}

	ret := &dmsV1.CheckDBServiceIsConnectableReply{}
	for _, item := range results {
		ret.Data = append(ret.Data, dmsV1.CheckDBServiceIsConnectableReplyItem{
			IsConnectable:       item.IsConnectable,
			Component:           item.Component,
			ConnectErrorMessage: item.ConnectErrorMessage,
		})
	}

	return ret, nil
}

func (d *DMSService) CheckDBServiceIsConnectableById(ctx context.Context, req *dmsV1.CheckDBServiceIsConnectableByIdReq) (reply *dmsV1.CheckDBServiceIsConnectableReply, err error) {
	dbService, err := d.DBServiceUsecase.GetDBService(ctx, req.DBServiceUid)
	if err != nil {
		return nil, err
	}

	var additionParams []*dmsCommonV1.AdditionalParam
	for _, item := range dbService.AdditionalParams {
		additionParams = append(additionParams, &dmsCommonV1.AdditionalParam{
			Name:  item.Key,
			Value: item.Value,
		})
	}

	checkDbConnectableParams := dmsCommonV1.CheckDbConnectable{
		DBType:           dbService.DBType,
		User:             dbService.User,
		Host:             dbService.Host,
		Port:             dbService.Port,
		Password:         dbService.Password,
		AdditionalParams: additionParams,
	}
	results, err := d.DBServiceUsecase.IsConnectable(ctx, checkDbConnectableParams)
	if err != nil {
		d.log.Errorf("IsConnectable err: %v", err)
		d.updateConnectionStatus(ctx, false, err.Error(), dbService)
		return nil, err
	}
	isSuccess, connectMsg := isConnectedSuccess(results)
	d.updateConnectionStatus(ctx, isSuccess, connectMsg, dbService)
	ret := &dmsV1.CheckDBServiceIsConnectableReply{}
	for _, item := range results {
		ret.Data = append(ret.Data, dmsV1.CheckDBServiceIsConnectableReplyItem{
			IsConnectable:       item.IsConnectable,
			Component:           item.Component,
			ConnectErrorMessage: item.ConnectErrorMessage,
		})
	}

	return ret, nil
}

func (d *DMSService) updateConnectionStatus(ctx context.Context, isSuccess bool, errorMsg string, dbService *biz.DBService) {
	lastConnectionStatus := biz.LastConnectionStatusSuccess
	lastConnectionTime := time.Now()
	dbService.LastConnectionTime = &lastConnectionTime
	if !isSuccess {
		lastConnectionStatus = biz.LastConnectionStatusFailed
		dbService.LastConnectionStatus = &lastConnectionStatus
		dbService.LastConnectionErrorMsg = &errorMsg
	} else {
		lastConnectionStatus = biz.LastConnectionStatusSuccess
		dbService.LastConnectionStatus = &lastConnectionStatus
		dbService.LastConnectionErrorMsg = nil
	}
	err := d.DBServiceUsecase.UpdateDBService(ctx, dbService, pkgConst.UIDOfUserAdmin)
	if err != nil {
		d.log.Errorf("dbService name: %v,UpdateDBServiceByBiz err: %v", dbService.Name, err)
	}
}

func isConnectedSuccess(results []*biz.IsConnectableReply) (bool, string) {
	if len(results) == 0 {
		return false, "check db connectable failed"
	}
	for _, connectableResult := range results {
		if !connectableResult.IsConnectable {
			return false, connectableResult.ConnectErrorMessage
		}
	}
	return true, ""
}

func (d *DMSService) CheckDBServiceIsConnectableByIds(ctx context.Context, projectUID, currentUserUid string, dbServiceList []dmsV1.DbServiceConnections) (*dmsV1.DBServicesConnectionReqReply, error) {
	if len(dbServiceList) == 0 {
		return &dmsV1.DBServicesConnectionReqReply{
			Data: []dmsV1.DBServiceIsConnectableReply{},
		}, nil
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	var dbServiceUidList []string
	for _, dbService := range dbServiceList {
		dbServiceUidList = append(dbServiceUidList, dbService.DBServiceUid)
	}

	filterBy = append(filterBy, pkgConst.FilterCondition{
		Field:    string(biz.DBServiceFieldUID),
		Operator: pkgConst.FilterOperatorIn,
		Value:    dbServiceUidList,
	})

	listOption := &biz.ListDBServicesOption{
		PageNumber:   1,
		LimitPerPage: uint32(len(dbServiceList)),
		OrderBy:      biz.DBServiceFieldName,
		FilterBy:     filterBy,
	}

	DBServiceList, _, err := d.DBServiceUsecase.ListDBService(ctx, listOption, projectUID, currentUserUid)
	if err != nil {
		return nil, err
	}

	resp := d.DBServiceUsecase.TestDbServiceConnections(ctx, DBServiceList, currentUserUid)

	return &dmsV1.DBServicesConnectionReqReply{
		Data: resp,
	}, nil
}

const DefaultBackupMaxRows uint64 = 1000

// autoChooseBackupMaxRows 函数根据是否启用备份以及备份最大行数的设置来确定最终的备份最大行数。
// 参数:
//   - enableBackup: 是否启用备份。
//   - backupMaxRows: 备份最大行数的设置，如果未设置，则为 nil。
//
// 返回值:
//   - int64: 最终确定的备份最大行数。
func autoChooseBackupMaxRows(enableBackup bool, backupMaxRows *uint64) uint64 {
	// 如果启用了备份并且备份最大行数的设置不为 nil，则返回设置的备份最大行数。
	if enableBackup && backupMaxRows != nil {
		return *backupMaxRows
	}
	// 如果未启用备份或者备份最大行数的设置为 nil，则返回默认的备份最大行数 DefaultBackupMaxRows。
	return DefaultBackupMaxRows
}

// Deprecated
func (d *DMSService) AddDBService(ctx context.Context, req *dmsV1.AddDBServiceReq, currentUserUid string) (reply *dmsV1.AddDBServiceReply, err error) {
	d.log.Infof("AddDBServices.req=%v", req)
	defer func() {
		d.log.Infof("AddDBServices.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	additionalParams, err := d.DBServiceUsecase.GetDriverParamsByDBType(ctx, req.DBService.DBType)
	for _, additionalParam := range req.DBService.AdditionalParams {
		err = additionalParams.SetParamValue(additionalParam.Name, additionalParam.Value)
		if err != nil {
			return nil, fmt.Errorf("set param value failed,invalid db type: %s", req.DBService.DBType)
		}
	}

	args := &biz.BizDBServiceArgs{
		Name:     req.DBService.Name,
		Desc:     &req.DBService.Desc,
		DBType:   req.DBService.DBType,
		Host:     req.DBService.Host,
		Port:     req.DBService.Port,
		User:     req.DBService.User,
		Password: &req.DBService.Password,
		// Business:          req.DBService.Business,
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(req.DBService.MaintenanceTimes),
		ProjectUID:        req.ProjectUid,
		Source:            string(pkgConst.DBServiceSourceNameSQLE),
		AdditionalParams:  additionalParams,
		EnableBackup:      req.DBService.EnableBackup,
		BackupMaxRows:     autoChooseBackupMaxRows(req.DBService.EnableBackup, req.DBService.BackupMaxRows),
	}

	if biz.IsDMS() {
		args.IsMaskingSwitch = req.DBService.IsEnableMasking
	}

	sqleConfig := req.DBService.SQLEConfig
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
	uid, err := d.DBServiceUsecase.CreateDBService(ctx, args, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("create db service failed: %w", err)
	}

	return &dmsV1.AddDBServiceReply{
		Data: struct {
			// db service UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) AddDBServiceV2(ctx context.Context, req *dmsV2.AddDBServiceReq, currentUserUid string) (reply *dmsV1.AddDBServiceReply, err error) {
	d.log.Infof("AddDBServices.req=%v", req)
	defer func() {
		d.log.Infof("AddDBServices.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	additionalParams, err := d.DBServiceUsecase.GetDriverParamsByDBType(ctx, req.DBService.DBType)
	for _, additionalParam := range req.DBService.AdditionalParams {
		err = additionalParams.SetParamValue(additionalParam.Name, additionalParam.Value)
		if err != nil {
			return nil, fmt.Errorf("set param value failed,invalid db type: %s", req.DBService.DBType)
		}
	}

	args := &biz.BizDBServiceArgs{
		Name:              req.DBService.Name,
		Desc:              &req.DBService.Desc,
		DBType:            req.DBService.DBType,
		Host:              req.DBService.Host,
		Port:              req.DBService.Port,
		User:              req.DBService.User,
		Password:          &req.DBService.Password,
		EnvironmentTagUID: req.DBService.EnvironmentTagUID,
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(req.DBService.MaintenanceTimes),
		ProjectUID:        req.ProjectUid,
		Source:            string(pkgConst.DBServiceSourceNameSQLE),
		AdditionalParams:  additionalParams,
		EnableBackup:      req.DBService.EnableBackup,
		BackupMaxRows:     autoChooseBackupMaxRows(req.DBService.EnableBackup, req.DBService.BackupMaxRows),
	}

	if biz.IsDMS() {
		args.IsMaskingSwitch = req.DBService.IsEnableMasking
	}

	sqleConfig := req.DBService.SQLEConfig
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
	uid, err := d.DBServiceUsecase.CreateDBService(ctx, args, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("create db service failed: %w", err)
	}

	return &dmsV1.AddDBServiceReply{
		Data: struct {
			// db service UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) convertMaintenanceTimeToPeriod(mt []*dmsCommonV1.MaintenanceTime) periods.Periods {
	ps := make(periods.Periods, len(mt))
	for i, time := range mt {
		ps[i] = &periods.Period{
			StartHour:   time.MaintenanceStartTime.Hour,
			StartMinute: time.MaintenanceStartTime.Minute,
			EndHour:     time.MaintenanceStopTime.Hour,
			EndMinute:   time.MaintenanceStopTime.Minute,
		}
	}
	return ps
}

func (d *DMSService) convertPeriodToMaintenanceTime(p periods.Periods) []*dmsCommonV1.MaintenanceTime {
	periods := make([]*dmsCommonV1.MaintenanceTime, len(p))
	for i, time := range p {
		periods[i] = &dmsCommonV1.MaintenanceTime{
			MaintenanceStartTime: &dmsCommonV1.Time{
				Hour:   time.StartHour,
				Minute: time.StartMinute,
			},
			MaintenanceStopTime: &dmsCommonV1.Time{
				Hour:   time.EndHour,
				Minute: time.EndMinute,
			},
		}
	}
	return periods
}

func (d *DMSService) convertBizDBServiceArgs2ImportDBService(dbs []*biz.BizDBServiceArgs) []*dmsV2.ImportDBService {
	ret := make([]*dmsV2.ImportDBService, len(dbs))
	for i, u := range dbs {
		ret[i] = &dmsV2.ImportDBService{
			ImportDBServiceCommon: dmsV2.ImportDBServiceCommon{
				Name:             u.Name,
				DBType:           u.DBType,
				Host:             u.Host,
				Port:             u.Port,
				User:             u.User,
				Password:         *u.Password,
				MaintenanceTimes: d.convertPeriodToMaintenanceTime(u.MaintenancePeriod),
				Desc:             *u.Desc,
				Source:           u.Source,
				ProjectUID:       u.ProjectUID,
				SQLEConfig: &dmsCommonV1.SQLEConfig{
					RuleTemplateName: u.RuleTemplateName,
					RuleTemplateID:   u.RuleTemplateID,
					SQLQueryConfig:   nil,
				},
				AdditionalParams: nil,
				IsEnableMasking:  false,
			},
			EnvironmentTagUID: u.EnvironmentTagUID,
		}

		if u.AdditionalParams != nil {
			additionalParams := make([]*dmsCommonV1.AdditionalParam, 0, len(u.AdditionalParams))
			for _, item := range u.AdditionalParams {
				additionalParams = append(additionalParams, &dmsCommonV1.AdditionalParam{
					Name:        item.Key,
					Value:       item.Value,
					Description: item.Desc,
					Type:        string(item.Type),
				})
			}
			ret[i].AdditionalParams = additionalParams
		}

		if u.SQLQueryConfig != nil {
			ret[i].SQLEConfig.SQLQueryConfig = &dmsCommonV1.SQLQueryConfig{
				MaxPreQueryRows:                  u.SQLQueryConfig.MaxPreQueryRows,
				QueryTimeoutSecond:               u.SQLQueryConfig.QueryTimeoutSecond,
				AuditEnabled:                     u.SQLQueryConfig.AuditEnabled,
				AllowQueryWhenLessThanAuditLevel: dmsCommonV1.SQLAllowQueryAuditLevel(u.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel),
			}
		}
	}

	return ret
}

func (d *DMSService) convertImportDBService2BizDBService(importDbs []dmsV2.ImportDBService) []*biz.DBService {
	ret := make([]*biz.DBService, len(importDbs))
	for i, u := range importDbs {
		ret[i] = &biz.DBService{
			UID:               "",
			Name:              u.ImportDBServiceCommon.Name,
			Desc:              u.ImportDBServiceCommon.Desc,
			DBType:            u.ImportDBServiceCommon.DBType,
			Host:              u.ImportDBServiceCommon.Host,
			Port:              u.ImportDBServiceCommon.Port,
			User:              u.ImportDBServiceCommon.User,
			Password:          u.ImportDBServiceCommon.Password,
			AdditionalParams:  nil,
			ProjectUID:        u.ImportDBServiceCommon.ProjectUID,
			MaintenancePeriod: d.convertMaintenanceTimeToPeriod(u.ImportDBServiceCommon.MaintenanceTimes),
			Source:            u.ImportDBServiceCommon.Source,
			SQLEConfig:        nil,
			IsMaskingSwitch:   u.ImportDBServiceCommon.IsEnableMasking,
			AccountPurpose:    "",
		}
		ret[i].EnvironmentTag = &dmsCommonV1.EnvironmentTag{
			UID: u.EnvironmentTagUID,
		}
		if u.AdditionalParams != nil {
			additionalParams := make([]*params.Param, 0, len(u.AdditionalParams))
			for _, item := range u.AdditionalParams {
				additionalParams = append(additionalParams, &params.Param{
					Key:   item.Name,
					Value: item.Value,
					Desc:  item.Description,
					Type:  params.ParamType(item.Type),
				})
			}
			ret[i].AdditionalParams = additionalParams
		}

		if u.SQLEConfig != nil {
			sqlConfig := &biz.SQLEConfig{
				RuleTemplateName: u.SQLEConfig.RuleTemplateName,
				RuleTemplateID:   u.SQLEConfig.RuleTemplateID,
				SQLQueryConfig:   &biz.SQLQueryConfig{},
			}
			if u.SQLEConfig.SQLQueryConfig != nil {
				sqlConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel = string(u.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel)
				sqlConfig.SQLQueryConfig.AuditEnabled = u.SQLEConfig.SQLQueryConfig.AuditEnabled
				sqlConfig.SQLQueryConfig.MaxPreQueryRows = u.SQLEConfig.SQLQueryConfig.MaxPreQueryRows
				sqlConfig.SQLQueryConfig.QueryTimeoutSecond = u.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond
			}
			ret[i].SQLEConfig = sqlConfig
		}
	}
	return ret
}

func (d *DMSService) ListDBServices(ctx context.Context, req *dmsCommonV2.ListDBServiceReq, currentUserUid string) (reply *dmsCommonV2.ListDBServiceReply, err error) {
	var orderBy biz.DBServiceField
	switch req.OrderBy {
	case dmsCommonV1.DBServiceOrderByName:
		orderBy = biz.DBServiceFieldName
	default:
		orderBy = biz.DBServiceFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)

	if req.FilterByEnvironmentTag != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldEnvironmentTagUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByEnvironmentTag,
		})
	}

	if req.FilterByHost != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldHost),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByHost,
		})
	}

	if req.FilterByUID != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByUID,
		})
	}

	if biz.IsDMS() && req.IsEnableMasking != nil {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldIsEnableMasking),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    *req.IsEnableMasking,
		})
	}

	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}

	if req.FilterByPort != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldPort),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByPort,
		})
	}

	if req.FilterByDBType != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldDBType),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByDBType,
		})
	}

	if req.ProjectUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.ProjectUid,
		})
	}
	if req.FilterLastConnectionTestStatus != nil {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldLastConnectionStatus),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    *req.FilterLastConnectionTestStatus,
		})
	}
	if len(req.FilterByDBServiceIds) > 0 {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldUID),
			Operator: pkgConst.FilterOperatorIn,
			Value:    req.FilterByDBServiceIds,
		})
	}
	if req.FuzzyKeyword != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:         string(biz.DBServiceFieldPort),
			Operator:      pkgConst.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
		}, pkgConst.FilterCondition{
			Field:         string(biz.DBServiceFieldHost),
			Operator:      pkgConst.FilterOperatorContains,
			Value:         req.FuzzyKeyword,
			KeywordSearch: true,
		},
		)
	}

	listOption := &biz.ListDBServicesOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	service, total, err := d.DBServiceUsecase.ListDBService(ctx, listOption, req.ProjectUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsCommonV2.ListDBService, len(service))
	for i, u := range service {
		password, err := pkgAes.AesEncrypt(u.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		ret[i] = &dmsCommonV2.ListDBService{
			DBServiceUid:        u.GetUID(),
			Name:                u.Name,
			DBType:              u.DBType,
			Host:                u.Host,
			Port:                u.Port,
			User:                u.User,
			Password:            password,
			EnvironmentTag:      u.EnvironmentTag,
			MaintenanceTimes:    d.convertPeriodToMaintenanceTime(u.MaintenancePeriod),
			Desc:                u.Desc,
			Source:              u.Source,
			ProjectUID:          u.ProjectUID,
			IsEnableMasking:     u.IsMaskingSwitch,
			InstanceAuditPlanID: u.InstanceAuditPlanID,
			AuditPlanTypes:      u.AuditPlanTypes,
			EnableBackup:        u.EnableBackup,
			BackupMaxRows:       u.BackupMaxRows,
		}

		if u.LastConnectionTime != nil {
			ret[i].LastConnectionTestTime = strfmt.DateTime(*u.LastConnectionTime)
		}
		if u.LastConnectionStatus != nil {
			ret[i].LastConnectionTestStatus = dmsCommonV1.LastConnectionTestStatus(*u.LastConnectionStatus)
		}
		if u.LastConnectionErrorMsg != nil {
			ret[i].LastConnectionTestErrorMessage = *u.LastConnectionErrorMsg
		}

		if u.AdditionalParams != nil {
			additionalParams := make([]*dmsCommonV1.AdditionalParam, 0, len(u.AdditionalParams))
			for _, item := range u.AdditionalParams {
				additionalParams = append(additionalParams, &dmsCommonV1.AdditionalParam{
					Name:        item.Key,
					Value:       item.Value,
					Description: item.Desc,
					Type:        string(item.Type),
				})
			}
			ret[i].AdditionalParams = additionalParams
		}

		if u.SQLEConfig != nil {
			sqlConfig := &dmsCommonV1.SQLEConfig{
				RuleTemplateName: u.SQLEConfig.RuleTemplateName,
				RuleTemplateID:   u.SQLEConfig.RuleTemplateID,
				SQLQueryConfig:   &dmsCommonV1.SQLQueryConfig{},
			}
			if u.SQLEConfig.SQLQueryConfig != nil {
				sqlConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel = dmsCommonV1.SQLAllowQueryAuditLevel(u.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel)
				sqlConfig.SQLQueryConfig.AuditEnabled = u.SQLEConfig.SQLQueryConfig.AuditEnabled
				sqlConfig.SQLQueryConfig.MaxPreQueryRows = u.SQLEConfig.SQLQueryConfig.MaxPreQueryRows
				sqlConfig.SQLQueryConfig.QueryTimeoutSecond = u.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond
			}
			ret[i].SQLEConfig = sqlConfig
		}
	}

	return &dmsCommonV2.ListDBServiceReply{
		Data:  ret,
		Total: total,
	}, nil
}

func (d *DMSService) ListDBServiceTips(ctx context.Context, req *dmsV1.ListDBServiceTipsReq, userId string) (reply *dmsV1.ListDBServiceTipsReply, err error) {
	dbServices, err := d.DBServiceUsecase.ListDBServiceTips(ctx, req, userId)
	if err != nil {
		return nil, err
	}

	ret := make([]*dmsV1.ListDBServiceTipItem, 0, len(dbServices))
	for _, item := range dbServices {
		ret = append(ret, &dmsV1.ListDBServiceTipItem{
			Id:   item.UID,
			Name: item.Name,
			Host: item.Host,
			Port: item.Port,
			Type: item.DBType,
		})
	}

	return &dmsV1.ListDBServiceTipsReply{
		Data: ret,
	}, nil
}

func (d *DMSService) ListDBServiceDriverOption(ctx context.Context) (reply *dmsV1.ListDBServiceDriverOptionReply, err error) {
	options, err := d.DBServiceUsecase.ListDBServiceDriverOption(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]*dmsV1.DatabaseDriverOption, 0, len(options))
	for _, item := range options {
		ret = append(ret, &dmsV1.DatabaseDriverOption{
			DBType:   item.DBType,
			LogoPath: item.LogoPath,
			Params:   item.Params,
		})
	}

	return &dmsV1.ListDBServiceDriverOptionReply{
		Data: ret,
	}, nil
}

func (d *DMSService) ListGlobalDBServices(ctx context.Context, req *dmsV2.ListGlobalDBServicesReq, currentUserUid string) (reply *dmsV2.ListGlobalDBServicesReply, err error) {
	return d.listGlobalDBServices(ctx, req, currentUserUid)
}

func (d *DMSService) ListGlobalDBServicesTips(ctx context.Context, currentUserUid string) (reply *dmsV1.ListGlobalDBServicesTipsReply, err error) {
	return d.listGlobalDBServicesTips(ctx, currentUserUid)
}

func (d *DMSService) ImportDBServicesOfOneProjectCheck(ctx context.Context, userUid, projectUid, fileContent string) (*dmsV2.ImportDBServicesCheckReply, []byte, error) {
	return d.importDBServicesOfOneProjectCheck(ctx, userUid, projectUid, fileContent)
}

func (d *DMSService) ImportDBServicesOfOneProject(ctx context.Context, req *dmsV2.ImportDBServicesOfOneProjectReq, uid string) error {
	return d.importDBServicesOfOneProject(ctx, req, uid)
}
