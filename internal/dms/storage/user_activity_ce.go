//go:build !enterprise

package storage

import (
	"context"
	"errors"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var errNotSupportUserActivity = errors.New("UserActivity related functions are enterprise version functions")

type userActivityRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewUserActivityRepo(log utilLog.Logger, s *Storage) biz.UserActivityRepo {
	return &userActivityRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.userActivity"))}
}

func (r *userActivityRepo) BatchSaveUserRequestLogs(_ context.Context, _ []*biz.UserRequestLog) error {
	return errNotSupportUserActivity
}

func (r *userActivityRepo) CleanUserRequestLogsBefore(_ context.Context, _ time.Time) (int64, error) {
	return 0, errNotSupportUserActivity
}

func (r *userActivityRepo) RollupDailyStats(_ context.Context, _ string) error {
	return errNotSupportUserActivity
}

func (r *userActivityRepo) GetUserActivitySummary(_ context.Context, _ string) (*biz.UserActivitySummary, error) {
	return nil, errNotSupportUserActivity
}

func (r *userActivityRepo) ListDailyTrend(_ context.Context, _, _ string) ([]*biz.UserActivityDailyTrendItem, error) {
	return nil, errNotSupportUserActivity
}

func (r *userActivityRepo) ListModuleDistribution(_ context.Context, _ string) ([]*biz.UserModuleDailyStat, error) {
	return nil, errNotSupportUserActivity
}

func (r *userActivityRepo) ListHourlyDistribution(_ context.Context, _ string) ([]*biz.ActiveHourlyStat, error) {
	return nil, errNotSupportUserActivity
}

func (r *userActivityRepo) ListUserDailyStats(_ context.Context, _ *biz.ListUserActivityUsersOption) ([]*biz.UserDailyActiveStat, uint64, error) {
	return nil, 0, errNotSupportUserActivity
}
