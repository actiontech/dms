package service

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/apiserver/conf"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
)

var OpPermissionNameByUID = map[string]*i18n.Message{
	pkgConst.UIDOfOpPermissionCreateProject:          locale.NameOpPermissionCreateProject,
	pkgConst.UIDOfOpPermissionProjectAdmin:           locale.NameOpPermissionProjectAdmin,
	pkgConst.UIDOfOpPermissionCreateWorkflow:         locale.NameOpPermissionCreateWorkflow,
	pkgConst.UIDOfOpPermissionAuditWorkflow:          locale.NameOpPermissionAuditWorkflow,
	pkgConst.UIDOfOpPermissionAuthDBServiceData:      locale.NameOpPermissionAuthDBServiceData,
	pkgConst.UIDOfOpPermissionExecuteWorkflow:        locale.NameOpPermissionExecuteWorkflow,
	pkgConst.UIDOfOpPermissionViewOthersWorkflow:     locale.NameOpPermissionViewOthersWorkflow,
	pkgConst.UIDOfOpPermissionSaveAuditPlan:          locale.NameOpPermissionSaveAuditPlan,
	pkgConst.UIDOfOpPermissionViewOthersAuditPlan:    locale.NameOpPermissionViewOthersAuditPlan,
	pkgConst.UIDOfOpPermissionSQLQuery:               locale.NameOpPermissionSQLQuery,
	pkgConst.UIDOfOpPermissionExportApprovalReject:   locale.NameOpPermissionExportApprovalReject,
	pkgConst.UIDOfOpPermissionExportCreate:           locale.NameOpPermissionExportCreate,
	pkgConst.UIDOfOpPermissionCreateOptimization:     locale.NameOpPermissionCreateOptimization,
	pkgConst.UIDOfOpPermissionViewOthersOptimization: locale.NameOpPermissionViewOthersOptimization,
	pkgConst.UIDOfOpPermissionGlobalManagement:       locale.NameOpPermissionGlobalManagement,
	pkgConst.UIDOfOpPermissionGlobalView:             locale.NameOpPermissionGlobalView,
	pkgConst.UIDOfOpPermissionCreatePipeline:         locale.NameOpPermissionCreatePipeline,
	pkgConst.UIDOfOrdinaryUser:                       locale.NameOpPermissionOrdinaryUser,
	pkgConst.UIDOfOpPermissionViewOperationRecord:    locale.NameOpPermissionViewOperationRecord,
	pkgConst.UIDOfOpPermissionViewExportTask:		  locale.NameOpPermissionViewExportTask,
	pkgConst.UIDOfPermissionViewQuickAuditRecord:     locale.NamePermissionViewQuickAuditRecord,
	pkgConst.UIDOfOpPermissionViewIDEAuditRecord:     locale.NameOpPermissionViewIDEAuditRecord,
	pkgConst.UIDOfOpPermissionViewOptimizationRecord: locale.NameOpPermissionViewOptimizationRecord,
	pkgConst.UIDOfOpPermissionVersionManage:          locale.NameOpPermissionVersionManage,
	pkgConst.UIdOfOpPermissionViewPipeline:			  locale.NameOpPermissionViewPipeline,
	pkgConst.UIdOfOpPermissionManageProjectDataSource:locale.NameOpPermissionManageProjectDataSource,
	pkgConst.UIdOfOpPermissionManageAuditRuleTemplate:locale.NameOpPermissionManageAuditRuleTemplate,
	pkgConst.UIdOfOpPermissionManageApprovalTemplate: locale.NameOpPermissionManageApprovalTemplate,
	pkgConst.UIdOfOpPermissionManageMember:           locale.NameOpPermissionManageMember,
	pkgConst.UIdOfOpPermissionPushRule:				  locale.NameOpPermissionPushRule,
	pkgConst.UIdOfOpPermissionMangeAuditSQLWhiteList: locale.NameOpPermissionMangeAuditSQLWhiteList,
	pkgConst.UIdOfOpPermissionManageSQLMangeWhiteList:locale.NameOpPermissionManageSQLMangeWhiteList,
	pkgConst.UIdOfOpPermissionManageRoleMange:		  locale.NameOpPermissionManageRoleMange,
	pkgConst.UIdOfOpPermissionDesensitization:		  locale.NameOpPermissionDesensitization,
}

var OpPermissionDescByUID = map[string]*i18n.Message{
	pkgConst.UIDOfOpPermissionCreateProject:          locale.DescOpPermissionCreateProject,
	pkgConst.UIDOfOpPermissionProjectAdmin:           locale.DescOpPermissionProjectAdmin,
	pkgConst.UIDOfOpPermissionCreateWorkflow:         locale.DescOpPermissionCreateWorkflow,
	pkgConst.UIDOfOpPermissionAuditWorkflow:          locale.DescOpPermissionAuditWorkflow,
	pkgConst.UIDOfOpPermissionAuthDBServiceData:      locale.DescOpPermissionAuthDBServiceData,
	pkgConst.UIDOfOpPermissionExecuteWorkflow:        locale.DescOpPermissionExecuteWorkflow,
	pkgConst.UIDOfOpPermissionViewOthersWorkflow:     locale.DescOpPermissionViewOthersWorkflow,
	pkgConst.UIDOfOpPermissionSaveAuditPlan:          locale.DescOpPermissionSaveAuditPlan,
	pkgConst.UIDOfOpPermissionViewOthersAuditPlan:    locale.DescOpPermissionViewOthersAuditPlan,
	pkgConst.UIDOfOpPermissionSQLQuery:               locale.DescOpPermissionSQLQuery,
	pkgConst.UIDOfOpPermissionExportApprovalReject:   locale.DescOpPermissionExportApprovalReject,
	pkgConst.UIDOfOpPermissionExportCreate:           locale.DescOpPermissionExportCreate,
	pkgConst.UIDOfOpPermissionCreateOptimization:     locale.DescOpPermissionCreateOptimization,
	pkgConst.UIDOfOpPermissionViewOthersOptimization: locale.DescOpPermissionViewOthersOptimization,
	pkgConst.UIDOfOpPermissionGlobalManagement:       locale.DescOpPermissionGlobalManagement,
	pkgConst.UIDOfOpPermissionGlobalView:             locale.DescOpPermissionGlobalView,
	pkgConst.UIDOfOpPermissionCreatePipeline:         locale.DescOpPermissionCreatePipeline,
	pkgConst.UIDOfOrdinaryUser:						  locale.DescOpPermissionOrdinaryUser,
}

func (d *DMSService) ListOpPermissions(ctx context.Context, req *dmsV1.ListOpPermissionReq) (reply *dmsV1.ListOpPermissionReply, err error) {

	var orderBy biz.OpPermissionField
	switch req.OrderBy {
	case dmsV1.OpPermissionOrderByName:
		orderBy = biz.OpPermissionFieldName
	default:
		orderBy = biz.OpPermissionFieldName
	}

	listOption := &biz.ListOpPermissionsOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
	}

	// 不支持智能调优时，隐藏相关权限
	if !conf.IsOptimizationEnabled() {
		listOption.FilterBy = append(listOption.FilterBy,
			pkgConst.FilterCondition{
				Field:    string(biz.OpPermissionFieldUID),
				Operator: pkgConst.FilterOperatorNotEqual,
				Value:    pkgConst.UIDOfOpPermissionCreateOptimization,
			},
			pkgConst.FilterCondition{
				Field:    string(biz.OpPermissionFieldUID),
				Operator: pkgConst.FilterOperatorNotEqual,
				Value:    pkgConst.UIDOfOpPermissionViewOthersOptimization,
			})
	}

	if req.Service != nil {
		listOption.FilterBy = append(listOption.FilterBy,
			pkgConst.FilterCondition{
				Field:    string(biz.OpPermissionFieldService),
				Operator: pkgConst.FilterOperatorEqual,
				Value:    *req.Service,
			})
	}

	var ops []*biz.OpPermission
	var total int64
	switch req.FilterByTarget {
	case dmsV1.OpPermissionTargetAll:
		listOption.FilterBy = append(listOption.FilterBy,
			pkgConst.FilterCondition{
				Field:    string(biz.OpPermissionFieldRangeType),
				Operator: pkgConst.FilterOperatorNotEqual,
				Value:    biz.OpRangeTypeGlobal,
			})
		ops, total, err = d.OpPermissionUsecase.ListOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	case dmsV1.OpPermissionTargetUser:
		ops, total, err = d.OpPermissionUsecase.ListUserOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	case dmsV1.OpPermissionTargetMember:
		ops, total, err = d.OpPermissionUsecase.ListMemberOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	case dmsV1.OpPermissionTargetProject:
		ops, total, err = d.OpPermissionUsecase.ListProjectOpPermissions(ctx, listOption)
		if nil != err {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid filter by target: %v", req.FilterByTarget)
	}

	ret := make([]*dmsV1.ListOpPermission, len(ops))
	for i, o := range ops {
		opRangeTyp, err := dmsV1.ParseOpRangeType(o.RangeType.String())
		if err != nil {
			return nil, fmt.Errorf("parse op range type failed: %v", err)
		}
		ret[i] = &dmsV1.ListOpPermission{
			OpPermission: dmsV1.UidWithName{
				Uid:  o.GetUID(),
				Name: locale.Bundle.LocalizeMsgByCtx(ctx, OpPermissionNameByUID[o.GetUID()]),
			},
			Description: locale.Bundle.LocalizeMsgByCtx(ctx, OpPermissionDescByUID[o.GetUID()]),
			RangeType:   opRangeTyp,
			Module: string(o.Module),
			Service: o.Service,
		}
	}

	return &dmsV1.ListOpPermissionReply{
		Data: ret, Total: total,
	}, nil
}
