//go:build enterprise

package service

import (
	"context"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) importDBServicesOfOneProjectCheck(ctx context.Context, userUid, projectUid, fileContent string) (*dmsV1.ImportDBServicesCheckReply, []byte, error) {
	dbs, resultContent, err := d.DBServiceUsecase.ImportDBServicesOfOneProjectCheck(ctx, userUid, projectUid, fileContent)
	if err != nil {
		return nil, nil, err
	}
	if resultContent != nil {
		return nil, resultContent, nil
	}

	ret := d.convertBizDBServiceArgs2ImportDBService(dbs)

	return &dmsV1.ImportDBServicesCheckReply{Data: ret}, nil, nil
}

func (d *DMSService) importDBServicesOfOneProject(ctx context.Context, req *dmsV1.ImportDBServicesOfOneProjectReq, uid string) error {
	ret := d.convertImportDBService2BizDBService(req.DBServices)
	return d.DBServiceUsecase.ImportDBServicesOfOneProject(ctx, ret, uid, req.ProjectUid)
}

func (d *DMSService) listGlobalDBServices(ctx context.Context, req *dmsV1.ListGlobalDBServicesReq, currentUserUid string) (reply *dmsV1.ListGlobalDBServicesReply, err error) {
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

	if req.FilterByPort != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldPort),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByPort,
		})
	}

	if req.FilterByUID != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByUID,
		})
	}

	if biz.IsDMS() && req.FilterByIsEnableMasking != nil {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldIsEnableMasking),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    *req.FilterByIsEnableMasking,
		})
	}

	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}

	if req.FilterByDBType != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldDBType),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByDBType,
		})
	}

	if req.FilterByProjectUid != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.DBServiceFieldProjectUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByProjectUid,
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
		}, pkgConst.FilterCondition{
			Field:         string(biz.DBServiceFieldName),
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

	service, total, err := d.DBServiceUsecase.ListGlobalDBServices(ctx, listOption, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsV1.ListGlobalDBService, len(service))
	for i, u := range service {
		ret[i] = &dmsV1.ListGlobalDBService{
			DBServiceUid:          u.GetUID(),
			Name:                  u.Name,
			DBType:                u.DBType,
			Host:                  u.Host,
			Port:                  u.Port,
			Business:              u.Business,
			MaintenanceTimes:      d.convertPeriodToMaintenanceTime(u.MaintenancePeriod),
			Desc:                  u.Desc,
			Source:                u.Source,
			ProjectUID:            u.ProjectUID,
			ProjectName:           u.ProjectName,
			IsEnableAudit:         false,
			IsEnableMasking:       u.IsMaskingSwitch,
			UnfinishedWorkflowNum: u.UnfinishedWorkflowNum,
		}

		if u.SQLEConfig != nil && u.SQLEConfig.SQLQueryConfig != nil {
			ret[i].IsEnableAudit = u.SQLEConfig.SQLQueryConfig.AuditEnabled
		}
	}

	return &dmsV1.ListGlobalDBServicesReply{
		Data:  ret,
		Total: total,
	}, nil
}
