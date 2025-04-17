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

func (svc *DMSService) GetResourceOverviewResourceTypeDistribution(ctx context.Context, currentUserUid string) (resp *dmsV1.ResourceOverviewResourceTypeDistributionRes, err error) {
	return nil, ErrNotSupportResourceOverview
}

func (svc *DMSService) GetResourceOverviewTopology(ctx context.Context, currentUserUid string, req *dmsV1.ResourceOverviewTopologyReq) (resp *dmsV1.ResourceOverviewTopologyRes, err error)  {
	return nil, ErrNotSupportResourceOverview
}

func (svc *DMSService) GetResourceOverviewList(ctx context.Context, currentUserUid string, req *dmsV1.ResourceOverviewResourceListReq) (resp *dmsV1.ResourceOverviewResourceListRes, err error) {
	return nil, ErrNotSupportResourceOverview
}