package biz

import (
	"context"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgParams "github.com/actiontech/dms/pkg/params"
	"github.com/robfig/cron/v3"
)

type DBServiceSyncTaskRepo interface {
	ListDBServiceSyncTasks(ctx context.Context, conditions []pkgConst.FilterCondition) ([]*DBServiceSyncTaskParams, error)
	SaveDBServiceSyncTask(ctx context.Context, params *DBServiceSyncTaskParams) error
	UpdateDBServiceSyncTask(ctx context.Context, params *DBServiceSyncTaskParams) error
	UpdateSyncDBServices(ctx context.Context, id string, fields map[string]interface{}) error
	DeleteDBServiceSyncTask(ctx context.Context, id string) error
	GetDBServiceSyncTaskById(ctx context.Context, id string) (*DBServiceSyncTaskParams, error)
}

type DBServiceSyncTaskUsecase struct {
	log                       *utilLog.Helper
	repo                      DBServiceSyncTaskRepo
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	projectUsecase            *ProjectUsecase
	dbServiceUsecase          *DBServiceUsecase
	cron                      *cron.Cron
	lastSyncTime              time.Time
}

func NewDBServiceSyncTaskUsecase(log utilLog.Logger, repo DBServiceSyncTaskRepo, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, projectUsecase *ProjectUsecase, dbServiceUsecase *DBServiceUsecase) *DBServiceSyncTaskUsecase {
	return &DBServiceSyncTaskUsecase{
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.database_source_service")),
		repo:                      repo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		projectUsecase:            projectUsecase,
		dbServiceUsecase:          dbServiceUsecase,
	}
}

type DBServiceSyncTaskParams struct {
	UID                 string     `json:"uid"`
	Name                string     `json:"name"`
	Source              string     `json:"source"`
	Version             string     `json:"version"`
	URL                 string     `json:"url"`
	DbType              string     `json:"db_type"`
	CronExpress         string     `json:"cron_express"`
	ProjectUID          string     `json:"project_uid"`
	LastSyncErr         string     `json:"last_sync_err"`
	LastSyncSuccessTime *time.Time `json:"last_sync_success_time"`
	AdditionalParams    pkgParams.Params
	SQLEConfig          *SQLEConfig
}

type ListDBServiceSyncTaskTipsParams struct {
	Source  pkgConst.DBServiceSourceName `json:"source"`
	DbTypes []pkgConst.DBType            `json:"db_types"`
}
