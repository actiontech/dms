package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	pkgAes "github.com/actiontech/dms/pkg/dms-common/pkg/aes"
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
		Business:          req.DBService.Business,
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
	if err := d.DBServiceUsecase.UpdateDBService(ctx, req.DBServiceUid, args, currentUserUid); err != nil {
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
		Name:              req.DBService.Name,
		Desc:              &req.DBService.Desc,
		DBType:            req.DBService.DBType,
		Host:              req.DBService.Host,
		Port:              req.DBService.Port,
		User:              req.DBService.User,
		Password:          &req.DBService.Password,
		Business:          req.DBService.Business,
		MaintenancePeriod: d.convertMaintenanceTimeToPeriod(req.DBService.MaintenanceTimes),
		ProjectUID:        req.ProjectUid,
		Source:            string(pkgConst.DBServiceSourceNameSQLE),
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

func (d *DMSService) ListDBServices(ctx context.Context, req *dmsCommonV1.ListDBServiceReq, currentUserUid string) (reply *dmsCommonV1.ListDBServiceReply, err error) {
	d.log.Infof("ListDBServices.req=%v", req)
	defer func() {
		d.log.Infof("ListDBServices.req=%v;reply=%v;error=%v", req, reply, err)
	}()

	var orderBy biz.DBServiceField
	switch req.OrderBy {
	case dmsCommonV1.DBServiceOrderByName:
		orderBy = biz.DBServiceFieldName
	default:
		orderBy = biz.DBServiceFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByBusiness != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldBusiness),
			Operator: pkgConst.FilterOperatorEqual,
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

	ret := make([]*dmsCommonV1.ListDBService, len(service))
	for i, u := range service {
		password, err := pkgAes.AesEncrypt(u.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		ret[i] = &dmsCommonV1.ListDBService{
			DBServiceUid:     u.GetUID(),
			Name:             u.Name,
			DBType:           u.DBType,
			Host:             u.Host,
			Port:             u.Port,
			User:             u.User,
			Password:         password,
			Business:         u.Business,
			MaintenanceTimes: d.convertPeriodToMaintenanceTime(u.MaintenancePeriod),
			Desc:             u.Desc,
			Source:           u.Source,
			ProjectUID:       u.ProjectUID,
			IsEnableMasking:  u.IsMaskingSwitch,
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

	return &dmsCommonV1.ListDBServiceReply{
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
		Data: ret,
	}, nil
}
