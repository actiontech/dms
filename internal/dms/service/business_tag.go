package service

import (
	"context"
	"fmt"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
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
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	if err := d.BusinessTagUsecase.CreateBusinessTag(ctx, businessTag.Name); err != nil {
		return fmt.Errorf("create business tag failed: %w", err)
	}

	return nil
}
