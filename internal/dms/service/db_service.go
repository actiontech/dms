package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgAes "github.com/actiontech/dms/pkg/dms-common/pkg/aes"
	"github.com/actiontech/dms/pkg/params"
	"github.com/actiontech/dms/pkg/periods"
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

func (d *DMSService) UpdateDBService(ctx context.Context, req *dmsV1.UpdateDBServiceReq, currentUserUid string) (err error) {
	d.log.Infof("UpdateDBService.req=%v", req)
	defer func() {
		d.log.Infof("UpdateDBService.req=%v;error=%v", req, err)
	}()

	additionalParams := params.AdditionalParameter[string(req.DBService.DBType)]
	for _, additionalParam := range req.DBService.AdditionalParams {
		err = additionalParams.SetParamValue(additionalParam.Name, additionalParam.Value)
		if err != nil {
			return fmt.Errorf("set param value failed,invalid db type: %s", req.DBService.DBType)
		}
	}

	var dbType pkgConst.DBType
	switch req.DBService.DBType {
	case dmsV1.DBTypeMySQL:
		dbType = pkgConst.DBTypeMySQL
	case dmsV1.DBTypeOceanBaseMySQL:
		dbType = pkgConst.DBTypeOceanBaseMySQL
	default:
		return fmt.Errorf("invalid db type: %s", req.DBService.DBType)
	}
	args := &biz.BizDBServiceArgs{
		DBType:            dbType,
		Desc:              req.DBService.Desc,
		Host:              req.DBService.Host,
		Port:              req.DBService.Port,
		AdminUser:         req.DBService.User,
		AdminPassword:     req.DBService.Password,
		Business:          req.DBService.Business,
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(req.DBService.MaintenanceTimes),
		AdditionalParams:  additionalParams,
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
	if err := d.DBServiceUsecase.UpdateDBService(ctx, req.DBServiceUid, args, currentUserUid); err != nil {
		return fmt.Errorf("update db service failed: %v", err)
	}

	return nil
}

func (d *DMSService) CheckDBServiceIsConnectable(ctx context.Context, req *dmsV1.CheckDBServiceIsConnectableReq) (reply *dmsV1.CheckDBServiceIsConnectableReply, err error) {
	d.log.Infof("CheckDBServiceIsConnectable.req=%v", req)
	defer func() {
		d.log.Infof("CheckDBServiceIsConnectable.req=%v; error=%v", req, err)
	}()
	
	results, err := d.DBServiceUsecase.IsConnectable(ctx, req.DBService)

	if err != nil {
		d.log.Errorf("IsConnectable err: %v", err)
		return nil, err
	}

	ret := &dmsV1.CheckDBServiceIsConnectableReply{}
	for _, item := range results {
		ret.Payload.Connections = append(ret.Payload.Connections, dmsV1.CheckDBServiceIsConnectableReplyItem{
			IsConnectable:       item.IsConnectable,
			Component:           item.Component,
			ConnectErrorMessage: item.ConnectErrorMessage,
		})
	}

	return ret, nil
}

func (d *DMSService) AddDBService(ctx context.Context, req *dmsV1.AddDBServiceReq, currentUserUid string) (reply *dmsV1.AddDBServiceReply, err error) {
	d.log.Infof("AddDBServices.req=%v", req)
	defer func() {
		d.log.Infof("AddDBServices.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	// TODO 预期这里不做校验，从dms同步出去的数据自行判断数据源类型是否支持
	var dbType pkgConst.DBType
	switch req.DBService.DBType {
	case dmsV1.DBTypeMySQL:
		dbType = pkgConst.DBTypeMySQL
	case dmsV1.DBTypeOceanBaseMySQL:
		dbType = pkgConst.DBTypeOceanBaseMySQL
	default:
		return nil, fmt.Errorf("invalid db type: %s", req.DBService.DBType)
	}

	additionalParams := params.AdditionalParameter[string(req.DBService.DBType)]
	for _, additionalParam := range req.DBService.AdditionalParams {
		err = additionalParams.SetParamValue(additionalParam.Name, additionalParam.Value)
		if err != nil {
			return nil, fmt.Errorf("set param value failed,invalid db type: %s", req.DBService.DBType)
		}
	}

	args := &biz.BizDBServiceArgs{
		Name:              req.DBService.Name,
		Desc:              &req.DBService.Desc,
		DBType:            dbType,
		Host:              req.DBService.Host,
		Port:              req.DBService.Port,
		AdminUser:         req.DBService.User,
		AdminPassword:     &req.DBService.Password,
		Business:          req.DBService.Business,
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(req.DBService.MaintenanceTimes),
		NamespaceUID:      req.DBService.NamespaceUID,
		AdditionalParams:  additionalParams,
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
		Payload: struct {
			// db service UID
			Uid string `json:"uid"`
		}{Uid: uid},
	}, nil
}

func (d *DMSService) convertMaintenanceTimeToPeriod(mt []*dmsV1.MaintenanceTime) periods.Periods {
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

func (d *DMSService) convertPeriodToMaintenanceTime(p periods.Periods) []*dmsV1.MaintenanceTime {
	periods := make([]*dmsV1.MaintenanceTime, len(p))
	for i, time := range p {
		periods[i] = &dmsV1.MaintenanceTime{
			MaintenanceStartTime: &dmsV1.Time{
				Hour:   time.StartHour,
				Minute: time.StartMinute,
			},
			MaintenanceStopTime: &dmsV1.Time{
				Hour:   time.EndHour,
				Minute: time.EndMinute,
			},
		}
	}
	return periods
}

func (d *DMSService) ListDBServices(ctx context.Context, req *dmsV1.ListDBServiceReq, currentUserUid string) (reply *dmsV1.ListDBServiceReply, err error) {
	d.log.Infof("ListDBServices.req=%v", req)
	defer func() {
		d.log.Infof("ListDBServices.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.DBServiceField
	switch req.OrderBy {
	case dmsV1.DBServiceOrderByName:
		orderBy = biz.DBServiceFieldName
	default:
		orderBy = biz.DBServiceFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByBusiness != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldBusiness),
			Operator: pkgConst.FilterOperatorContains,
			Value:    req.FilterByBusiness,
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

	if req.FilterByNamespaceUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldNamespaceUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByNamespaceUid,
		})
	}

	listOption := &biz.ListDBServicesOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	service, total, err := d.DBServiceUsecase.ListDBService(ctx, listOption, req.FilterByNamespaceUid, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListDBService, len(service))
	for i, u := range service {
		password, err := pkgAes.AesEncrypt(u.AdminPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		ret[i] = &dmsV1.ListDBService{
			DBServiceUid:     u.GetUID(),
			Name:             u.Name,
			DBType:           dmsV1.DBType(u.DBType),
			Host:             u.Host,
			Port:             u.Port,
			User:             u.AdminUser,
			Password:         password,
			Business:         u.Business,
			MaintenanceTimes: d.convertPeriodToMaintenanceTime(u.MaintenancePeriod),
			Desc:             u.Desc,
			Source:           u.Source,
			NamespaceUID:     u.NamespaceUID,
			// TODO 从provision获取
			// LastSyncDataResult: "TODO",
			// LastSyncDataTime:"".
		}
		if u.SQLEConfig != nil {
			sqlConfig := &dmsV1.SQLEConfig{
				RuleTemplateName: u.SQLEConfig.RuleTemplateName,
				RuleTemplateID:   u.SQLEConfig.RuleTemplateID,
				SQLQueryConfig:   &dmsV1.SQLQueryConfig{},
			}
			if u.SQLEConfig.SQLQueryConfig != nil {
				sqlConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel = dmsV1.SQLAllowQueryAuditLevel(u.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel)
				sqlConfig.SQLQueryConfig.AuditEnabled = u.SQLEConfig.SQLQueryConfig.AuditEnabled
				sqlConfig.SQLQueryConfig.MaxPreQueryRows = u.SQLEConfig.SQLQueryConfig.MaxPreQueryRows
				sqlConfig.SQLQueryConfig.QueryTimeoutSecond = u.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond
			}
			ret[i].SQLEConfig = sqlConfig
		}
	}

	return &dmsV1.ListDBServiceReply{
		Payload: struct {
			DBServices []*dmsV1.ListDBService `json:"db_services"`
			Total      int64                  `json:"total"`
		}{DBServices: ret, Total: total},
	}, nil
}

func (d *DMSService) ListDBServiceDriverOption(ctx context.Context) (reply *dmsV1.ListDBServiceDriverOptionReply, err error) {
	options, err := d.DBServiceUsecase.ListDBServiceDriverOption(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]*dmsV1.DatabaseDriverOption, 0, len(options))
	for _, item := range options {
		additionalParams := make([]*dmsV1.DatabaseDriverAdditionalParam, 0, len(item.Params))
		for _, param := range item.Params {
			additionalParams = append(additionalParams, &dmsV1.DatabaseDriverAdditionalParam{
				Name:        param.Key,
				Value:       param.Value,
				Type:        string(param.Type),
				Description: param.Desc,
			})
		}

		ret = append(ret, &dmsV1.DatabaseDriverOption{
			DBType:   item.DbType,
			LogoPath: item.LogoPath,
			Params:   additionalParams,
		})
	}

	return &dmsV1.ListDBServiceDriverOptionReply{
		Payload: struct {
			DatabaseDriverOptions []*dmsV1.DatabaseDriverOption `json:"database_driver_options"`
		}{
			ret,
		},
	}, nil
}
