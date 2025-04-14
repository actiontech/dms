//go:build !enterprise

package service

import (
	"context"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (svc *DMSService) GetResourceOverviewStatistics(ctx context.Context, currentUserUid string) (resp *dmsV1.ResourceOverviewStatisticsRes, err error) {
	return nil, biz.ErrNotSupportResourceOverview
}
