package biz

import (
	"context"
	"strconv"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

const (
	userActivitySessionGapMinutes = 30
)

// AmbientModuleCodesForTopModule are session/bootstrap modules excluded when
// computing top_module_code for user ranking. Total request counts still include them.
var AmbientModuleCodesForTopModule = []string{"AUTH", "USER_ROLE", "SYS_CONFIG"}

// ExcludedUserActivityUIDs returns built-in/service accounts excluded from statistics.
func ExcludedUserActivityUIDs() []string {
	return []string{pkgConst.UIDOfUserSys}
}

func IsExcludedUserActivityUID(uid string) bool {
	return uid == pkgConst.UIDOfUserSys
}

type UserRequestLog struct {
	ID              uint64
	EventTime       time.Time
	UserUID         string
	HTTPMethod      string
	NormalizedRoute string
	ModuleCode      string
	StatusCode      int
	LatencyMs       int
	ClientIP        string
	UserAgent       string
	ProjectUID      string
	Node            string
}

type UserDailyActiveStat struct {
	StatDate      string
	UserUID       string
	UserName      string
	ActiveDays    int
	RequestCount  int
	ErrorCount    int
	FirstActiveAt time.Time
	LastActiveAt  time.Time
	ActiveMinutes int
	TopModuleCode string
}

type UserModuleDailyStat struct {
	StatDate     string
	ModuleCode   string
	ModuleName   string
	RequestCount int
}

type ActiveHourlyStat struct {
	StatDate     string
	StatHour     int
	RequestCount int
	ActiveUsers  int
}

type UserActivityDailyTrendItem struct {
	StatDate     string
	DAU          int
	RequestCount int
	ErrorCount   int
}

type UserActivitySummary struct {
	DAU              int
	RequestCount     int
	AvgRequestPerUser float64
	ErrorCount       int
	ErrorRate        float64
	PeakHour         int
	PeakHourRequests int
}

type ListUserActivityUsersOption struct {
	PageIndex        uint32
	PageSize         uint32
	FilterDateFrom   string
	FilterDateTo     string
	FilterFuzzyUser  string
	OrderBy          string
}

type UserActivityRepo interface {
	BatchSaveUserRequestLogs(ctx context.Context, logs []*UserRequestLog) error
	CleanUserRequestLogsBefore(ctx context.Context, t time.Time) (rowsAffected int64, err error)
	RollupDailyStats(ctx context.Context, statDate string) error
	GetUserActivitySummary(ctx context.Context, statDate string) (*UserActivitySummary, error)
	ListDailyTrend(ctx context.Context, dateFrom, dateTo string) ([]*UserActivityDailyTrendItem, error)
	ListModuleDistribution(ctx context.Context, statDate string) ([]*UserModuleDailyStat, error)
	ListHourlyDistribution(ctx context.Context, statDate string) ([]*ActiveHourlyStat, error)
	ListUserDailyStats(ctx context.Context, opt *ListUserActivityUsersOption) ([]*UserDailyActiveStat, uint64, error)
}

type UserActivityUsecase struct {
	repo                  UserActivityRepo
	systemVariableUsecase *SystemVariableUsecase
	log                   *utilLog.Helper
}

func NewUserActivityUsecase(logger utilLog.Logger, repo UserActivityRepo, svu *SystemVariableUsecase) *UserActivityUsecase {
	return &UserActivityUsecase{
		repo:                  repo,
		systemVariableUsecase: svu,
		log:                   utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.userActivity")),
	}
}

func (u *UserActivityUsecase) GetLog() *utilLog.Helper {
	return u.log
}

func (u *UserActivityUsecase) DoClean() {
	if u.systemVariableUsecase == nil {
		u.log.Errorf("failed to clean user request logs when get systemVariableUsecase")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	variables, err := u.systemVariableUsecase.GetSystemVariables(ctx)
	if err != nil {
		u.log.Errorf("failed to clean user request logs when get expired duration: %v", err)
		return
	}

	expiredHoursVar, ok := variables[SystemVariableUserRequestLogExpiredHours]
	if !ok {
		expiredHoursVar = SystemVariable{
			Key:   SystemVariableUserRequestLogExpiredHours,
			Value: strconv.Itoa(DefaultUserRequestLogExpiredHours),
		}
	}

	expiredHours, err := strconv.Atoi(expiredHoursVar.Value)
	if err != nil {
		u.log.Errorf("failed to parse user_request_log_expired_hours value: %v", err)
		return
	}
	if expiredHours <= 0 {
		u.log.Errorf("got UserRequestLogExpiredHours: %d", expiredHours)
		return
	}

	cleanTime := time.Now().Add(time.Duration(-expiredHours) * time.Hour)
	rowsAffected, err := u.repo.CleanUserRequestLogsBefore(ctx, cleanTime)
	if err != nil {
		u.log.Errorf("failed to clean user request logs: %v", err)
		return
	}
	u.log.Infof("UserRequestLog regular cleaned rows: %d event time before: %s", rowsAffected, cleanTime.Format("2006-01-02 15:04:05"))
}

func CalcActiveMinutes(eventTimes []time.Time, gapMinutes int) int {
	if len(eventTimes) == 0 {
		return 0
	}
	if gapMinutes <= 0 {
		gapMinutes = userActivitySessionGapMinutes
	}
	gap := time.Duration(gapMinutes) * time.Minute
	total := time.Duration(0)
	sessionStart := eventTimes[0]
	sessionEnd := eventTimes[0]
	for i := 1; i < len(eventTimes); i++ {
		if eventTimes[i].Sub(sessionEnd) > gap {
			total += sessionEnd.Sub(sessionStart)
			sessionStart = eventTimes[i]
		}
		sessionEnd = eventTimes[i]
	}
	total += sessionEnd.Sub(sessionStart)
	minutes := int(total.Minutes())
	if total%time.Minute != 0 {
		minutes++
	}
	return minutes
}
