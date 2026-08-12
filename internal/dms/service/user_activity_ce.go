//go:build !enterprise

package service

import (
	"context"
	"errors"

	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
)

var errNotSupportUserActivity = errors.New("UserActivity related functions are enterprise version functions")

func (d *DMSService) GetUserActivitySummary(ctx context.Context, req *aV1.GetUserActivitySummaryReq, _ string) (*aV1.GetUserActivitySummaryReply, error) {
	reply := &aV1.GetUserActivitySummaryReply{}
	reply.GenericResp.SetCode(int(apiError.DMSServiceErr))
	reply.GenericResp.SetMsg(errNotSupportUserActivity.Error())
	return reply, nil
}

func (d *DMSService) ListUserActivityDailyTrend(ctx context.Context, req *aV1.ListUserActivityDailyTrendReq, _ string) (*aV1.ListUserActivityDailyTrendReply, error) {
	return nil, errNotSupportUserActivity
}

func (d *DMSService) ListUserActivityModuleDistribution(ctx context.Context, req *aV1.ListUserActivityModuleDistributionReq, _ string) (*aV1.ListUserActivityModuleDistributionReply, error) {
	return nil, errNotSupportUserActivity
}

func (d *DMSService) ListUserActivityHourlyDistribution(ctx context.Context, req *aV1.ListUserActivityHourlyDistributionReq, _ string) (*aV1.ListUserActivityHourlyDistributionReply, error) {
	return nil, errNotSupportUserActivity
}

func (d *DMSService) ListUserActivityUsers(ctx context.Context, req *aV1.ListUserActivityUsersReq, _ string) (*aV1.ListUserActivityUsersReply, error) {
	return nil, errNotSupportUserActivity
}
