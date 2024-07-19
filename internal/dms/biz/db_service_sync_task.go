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
	SaveDBServiceSyncTask(ctx context.Context, syncTask *DBServiceSyncTask) error
	UpdateDBServiceSyncTask(ctx context.Context, syncTask *DBServiceSyncTask) error
	GetDBServiceSyncTaskById(ctx context.Context, id string) (*DBServiceSyncTask, error)
	ListDBServiceSyncTasks(ctx context.Context) ([]*DBServiceSyncTask, error)
	ListDBServiceSyncTaskTips() ([]ListDBServiceSyncTaskTips, error)
	DeleteDBServiceSyncTask(ctx context.Context, id string) error
	UpdateDBServiceSyncTaskByFields(ctx context.Context, id string, fields map[string]interface{}) error
}

type DBServiceSyncTaskUsecase struct {
	log                       *utilLog.Helper
	repo                      DBServiceSyncTaskRepo
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	dbServiceUsecase          *DBServiceUsecase
	projectUsecase            *ProjectUsecase
	cron                      *cron.Cron
	lastSyncTime              time.Time
}

func NewDBServiceSyncTaskUsecase(log utilLog.Logger, repo DBServiceSyncTaskRepo, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, projectUsecase *ProjectUsecase, dbServiceUsecase *DBServiceUsecase) *DBServiceSyncTaskUsecase {
	return &DBServiceSyncTaskUsecase{
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.db_service_sync_task")),
		repo:                      repo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		projectUsecase:            projectUsecase,
		dbServiceUsecase:          dbServiceUsecase,
	}
}

type DBServiceSyncTask struct {
	UID                 string           `json:"uid"`
	Name                string           `json:"name"`
	Source              string           `json:"source"`
	URL                 string           `json:"url"`
	DbType              string           `json:"db_type"`
	CronExpress         string           `json:"cron_express"`
	LastSyncErr         string           `json:"last_sync_err"`
	LastSyncSuccessTime *time.Time       `json:"last_sync_success_time"`
	AdditionalParam     pkgParams.Params `json:"additional_params"`
	SQLEConfig          *SQLEConfig      `json:"sqle_config"`
}

type ListDBServiceSyncTaskTips struct {
	Type    pkgConst.DBServiceSourceName `json:"service_source_name"`
	Desc    string                       `json:"description"`
	DBTypes []pkgConst.DBType            `json:"db_type"`
	Params  pkgParams.Params             `json:"params,omitempty"`
}
