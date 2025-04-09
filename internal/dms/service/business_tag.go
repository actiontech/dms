package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func (d *DMSService) CreateBusinessTag(ctx context.Context, currentUserUid string, businessTag *v1.BusinessTag) (err error) {
	d.log.Infof("CreateBusinessTag.req=%v", businessTag)
	defer func() {
		d.log.Infof("CreateBusinessTag.req=%v;error=%v", businessTag, err)
	}()

	// 权限校验
	if canGlobalOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
		return fmt.Errorf("check user op permission failed: %v", err)
	} else if !canGlobalOp {
		return fmt.Errorf("user is not project admin or global op permission user")
	}

	if err := d.BusinessTagUsecase.CreateBusinessTag(ctx, businessTag.Name); err != nil {
		return fmt.Errorf("create business tag failed: %w", err)
	}

	return nil
}

func (d *DMSService) UpdateBusinessTag(ctx context.Context, currentUserUid string, businessTagUID string, businessTagForUpdate *v1.BusinessTag) (err error) {
	d.log.Infof("UpdateBusinessTag.req=%v", businessTagForUpdate)
	defer func() {
		d.log.Infof("UpdateBusinessTag.req=%v;error=%v", businessTagForUpdate, err)
	}()

	// 权限校验
	if canGlobalOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
		return fmt.Errorf("check user op permission failed: %v", err)
	} else if !canGlobalOp {
		return fmt.Errorf("user is not project admin or global op permission user")
	}

	if err := d.BusinessTagUsecase.UpdateBusinessTag(ctx, businessTagUID, businessTagForUpdate.Name); err != nil {
		return fmt.Errorf("update business tag failed: %w", err)
	}
	return nil
}

func (d *DMSService) DeleteBusinessTag(ctx context.Context, currentUserUid string, businessTagUID string) (err error) {
	d.log.Infof("DeleteBusinessTag.req=%v", businessTagUID)
	defer func() {
		d.log.Infof("DeleteBusinessTag.req=%v;error=%v", businessTagUID, err)
	}()

	// 权限校验
	if canGlobalOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserUid); err != nil {
		return fmt.Errorf("check user op permission failed: %v", err)
	} else if !canGlobalOp {
		return fmt.Errorf("user is not project admin or global op permission user")
	}

	// 业务标签被项目关联时，不允许删除
	filterBy := []pkgConst.FilterCondition{
		{
			Field:    string(biz.ProjectFieldBusinessTagUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    businessTagUID,
		},
	}

	_, total, err := d.ProjectUsecase.ListProject(ctx, &biz.ListProjectsOption{FilterBy: filterBy}, currentUserUid)
	if err != nil {
		return fmt.Errorf("list project failed: %w", err)
	}
	if total > 0 {
		return fmt.Errorf("business tag is used by project, please detach business tag with project first")
	}

	if err := d.BusinessTagUsecase.DeleteBusinessTag(ctx, businessTagUID); err != nil {
		return fmt.Errorf("delete business tag failed: %w", err)
	}

	return nil
}

func (d *DMSService) ListBusinessTags(ctx context.Context, req *v1.ListBusinessTagReq) (reply *v1.ListBusinessTagsReply, err error) {
	d.log.Infof("ListBusinessTags.req=%v", *req)
	defer func() {
		d.log.Infof("ListBusinessTags.req=%v;error=%v", *req, err)
	}()
	limit, offset := d.GetLimitAndOffset(req.PageIndex, req.PageSize)
	bizBusinessTags, count, err := d.BusinessTagUsecase.ListBusinessTags(ctx, &biz.ListBusinessTagsOption{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list business tags failed: %w", err)
	}
	businessTags := make([]*v1.BusinessTag, 0, len(bizBusinessTags))
	for _, bizBusinessTag := range bizBusinessTags {
		businessTags = append(businessTags, &v1.BusinessTag{
			UID:  bizBusinessTag.UID,
			Name: bizBusinessTag.Name,
		})
	}
	return &v1.ListBusinessTagsReply{
		Data:  businessTags,
		Total: count,
	}, nil
}
