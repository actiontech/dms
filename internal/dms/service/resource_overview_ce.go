//go:build !enterprise

package service

import (
	"context"
	"errors"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

var ErrNotSupportResourceOverview = errors.New("resource overview related functions are enterprise version functions")

func (svc *DMSService) GetResourceOverviewStatistics(ctx context.Context, currentUserUid string) (resp *dmsV1.ResourceOverviewStatisticsRes, err error) {
	return nil, ErrNotSupportResourceOverview
}
