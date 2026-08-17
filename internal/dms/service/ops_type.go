package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/pkg/locale"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// defaultOpsTypeLocaleByCanonicalName 存库中文规范名 → locale 消息（用户改名后不再映射）。
var defaultOpsTypeLocaleByCanonicalName = map[string]*i18n.Message{
	locale.DefaultOpsTypeNameDataModification.Other:   locale.DefaultOpsTypeNameDataModification,
	locale.DefaultOpsTypeNameDataExtraction.Other:     locale.DefaultOpsTypeNameDataExtraction,
	locale.DefaultOpsTypeNameServiceMaintenance.Other: locale.DefaultOpsTypeNameServiceMaintenance,
	locale.DefaultOpsTypeNameConfigModification.Other: locale.DefaultOpsTypeNameConfigModification,
}

func localizeOpsTypeDisplayName(ctx context.Context, name string) string {
	if msg, ok := defaultOpsTypeLocaleByCanonicalName[name]; ok {
		return locale.Bundle.LocalizeMsgByCtx(ctx, msg)
	}
	return name
}

func (d *DMSService) CreateOpsType(ctx context.Context, projectUid, currentUserUid, opsTypeName string) (err error) {
	d.log.Infof("CreateOpsType.req=%v", opsTypeName)
	defer func() {
		d.log.Infof("CreateOpsType.req=%v;error=%v", opsTypeName, err)
	}()

	if err := d.OpsTypeUsecase.CreateOpsType(ctx, projectUid, currentUserUid, opsTypeName); err != nil {
		return fmt.Errorf("create ops type failed: %w", err)
	}

	return nil
}

func (d *DMSService) UpdateOpsType(ctx context.Context, projectUid, currentUserUid string, opsTypeUID, opsTypeName string) (err error) {
	d.log.Infof("UpdateOpsType.req=%v", opsTypeName)
	defer func() {
		d.log.Infof("UpdateOpsType.req=%v;error=%v", opsTypeName, err)
	}()

	if err := d.OpsTypeUsecase.UpdateOpsType(ctx, projectUid, currentUserUid, opsTypeUID, opsTypeName); err != nil {
		return fmt.Errorf("update ops type failed: %w", err)
	}
	return nil
}

func (d *DMSService) DeleteOpsType(ctx context.Context, projectUid, currentUserUid string, opsTypeUID string) (err error) {
	d.log.Infof("DeleteOpsType.req=%v", opsTypeUID)
	defer func() {
		d.log.Infof("DeleteOpsType.req=%v;error=%v", opsTypeUID, err)
	}()

	// 对标 DeleteEnvironmentTag：被数据导出工单引用时不允许删除（AC-B05；本期不校验 SQL 上线侧）。
	count, err := d.DataExportWorkflowUsecase.CountDataExportWorkflowsByOpsTypeUID(ctx, projectUid, opsTypeUID)
	if err != nil {
		return fmt.Errorf("check ops type reference failed: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%s", locale.Bundle.LocalizeMsgByCtx(ctx, locale.ErrOpsTypeReferencedByDataExportWorkflow))
	}

	if err := d.OpsTypeUsecase.DeleteOpsType(ctx, projectUid, currentUserUid, opsTypeUID); err != nil {
		return fmt.Errorf("delete ops type failed: %w", err)
	}

	return nil
}

func (d *DMSService) ListOpsTypes(ctx context.Context, req *v1.ListOpsTypeReq) (reply *v1.ListOpsTypesReply, err error) {
	d.log.Infof("ListOpsTypes.req=%v", *req)
	defer func() {
		d.log.Infof("ListOpsTypes.req=%v;error=%v", *req, err)
	}()
	limit, offset := d.GetLimitAndOffset(req.PageIndex, req.PageSize)
	bizOpsTypes, count, err := d.OpsTypeUsecase.ListOpsTypes(ctx, &biz.ListOpsTypesOption{
		Limit:      limit,
		Offset:     offset,
		ProjectUID: req.ProjectUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list ops types failed: %w", err)
	}
	opsTypes := make([]*dmsCommonV1.OpsType, 0, len(bizOpsTypes))
	for _, bizOpsType := range bizOpsTypes {
		opsTypes = append(opsTypes, &dmsCommonV1.OpsType{
			UID:  bizOpsType.UID,
			Name: localizeOpsTypeDisplayName(ctx, bizOpsType.Name),
		})
	}
	return &v1.ListOpsTypesReply{
		Data:  opsTypes,
		Total: count,
	}, nil
}
