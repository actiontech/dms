//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportUserActivity = errors.New("UserActivity related functions are enterprise version functions")

func (u *UserActivityUsecase) BatchSaveUserRequestLogs(_ context.Context, _ []*UserRequestLog) error {
	return errNotSupportUserActivity
}

func (u *UserActivityUsecase) DoDailyRollup(_ context.Context) error {
	return errNotSupportUserActivity
}

func (u *UserActivityUsecase) GetUserActivitySummary(_ context.Context, _ string) (*UserActivitySummary, error) {
	return nil, errNotSupportUserActivity
}

func (u *UserActivityUsecase) ListDailyTrend(_ context.Context, _, _ string) ([]*UserActivityDailyTrendItem, error) {
	return nil, errNotSupportUserActivity
}

func (u *UserActivityUsecase) ListModuleDistribution(_ context.Context, _ string) ([]*UserModuleDailyStat, error) {
	return nil, errNotSupportUserActivity
}

func (u *UserActivityUsecase) ListHourlyDistribution(_ context.Context, _ string) ([]*ActiveHourlyStat, error) {
	return nil, errNotSupportUserActivity
}

func (u *UserActivityUsecase) ListUserDailyStats(_ context.Context, _ *ListUserActivityUsersOption) ([]*UserDailyActiveStat, uint64, error) {
	return nil, 0, errNotSupportUserActivity
}
