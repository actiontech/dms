package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) CreateEnvironmentTag(ctx context.Context, projectUid, currentUserUid, environmentTagName string) (err error) {
	d.log.Infof("CreateEnvironmentTag.req=%v", environmentTagName)
	defer func() {
		d.log.Infof("CreateEnvironmentTag.req=%v;error=%v", environmentTagName, err)
	}()

	if err := d.EnvironmentTagUsecase.CreateEnvironmentTag(ctx, projectUid, currentUserUid, environmentTagName); err != nil {
		return fmt.Errorf("create environment tag failed: %w", err)
	}

	return nil
}

func (d *DMSService) UpdateEnvironmentTag(ctx context.Context, projectUid, currentUserUid string, environmentTagUID, environmentTagName string) (err error) {
	d.log.Infof("UpdateEnvironmentTag.req=%v", environmentTagName)
	defer func() {
		d.log.Infof("UpdateEnvironmentTag.req=%v;error=%v", environmentTagName, err)
	}()

	if err := d.EnvironmentTagUsecase.UpdateEnvironmentTag(ctx, projectUid, currentUserUid, environmentTagUID, environmentTagName); err != nil {
		return fmt.Errorf("update environment tag failed: %w", err)
	}
	return nil
}

func (d *DMSService) DeleteEnvironmentTag(ctx context.Context, projectUid, currentUserUid string, environmentTagUID string) (err error) {
	d.log.Infof("DeleteEnvironmentTag.req=%v", environmentTagUID)
	defer func() {
		d.log.Infof("DeleteEnvironmentTag.req=%v;error=%v", environmentTagUID, err)
	}()

	// 环境标签被数据源关联时，不允许删除
	filterBy := []pkgConst.FilterCondition{
		{
			Field:    string(biz.DBServiceFieldEnvironmentTagUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    environmentTagUID,
		},
	}

	_, total, err := d.DBServiceUsecase.ListDBService(ctx, &biz.ListDBServicesOption{FilterByOptions: pkgConst.ConditionsToFilterOptions(filterBy)}, projectUid, currentUserUid)
	if err != nil {
		return fmt.Errorf("list project failed: %w", err)
	}
	if total > 0 {
		return fmt.Errorf("environment tag is used by db service, please detach environment tag with db service first")
	}

	if err := d.EnvironmentTagUsecase.DeleteEnvironmentTag(ctx, projectUid, currentUserUid, environmentTagUID); err != nil {
		return fmt.Errorf("delete environment tag failed: %w", err)
	}

	return nil
}

func (d *DMSService) ListEnvironmentTags(ctx context.Context, req *v1.ListEnvironmentTagReq) (reply *v1.ListEnvironmentTagsReply, err error) {
	d.log.Infof("ListEnvironmentTags.req=%v", *req)
	defer func() {
		d.log.Infof("ListEnvironmentTags.req=%v;error=%v", *req, err)
	}()
	limit, offset := d.GetLimitAndOffset(req.PageIndex, req.PageSize)
	bizEnvironmentTags, count, err := d.EnvironmentTagUsecase.ListEnvironmentTags(ctx, &biz.ListEnvironmentTagsOption{
		Limit:      limit,
		Offset:     offset,
		ProjectUID: req.ProjectUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list environment tags failed: %w", err)
	}
	environmentTags := make([]*dmsCommonV1.EnvironmentTag, 0, len(bizEnvironmentTags))
	for _, bizEnvironmentTag := range bizEnvironmentTags {
		environmentTags = append(environmentTags, &dmsCommonV1.EnvironmentTag{
			UID:  bizEnvironmentTag.UID,
			Name: bizEnvironmentTag.Name,
		})
	}
	return &v1.ListEnvironmentTagsReply{
		Data:  environmentTags,
		Total: count,
	}, nil
}
