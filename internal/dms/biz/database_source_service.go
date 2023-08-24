package biz

import (
	"context"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgParams "github.com/actiontech/dms/pkg/params"
	"github.com/robfig/cron/v3"
)

type DatabaseSourceServiceRepo interface {
	ListDatabaseSourceServices(ctx context.Context, conditions []pkgConst.FilterCondition) ([]*DatabaseSourceServiceParams, error)
	SaveDatabaseSourceService(ctx context.Context, params *DatabaseSourceServiceParams) error
	UpdateDatabaseSourceService(ctx context.Context, params *DatabaseSourceServiceParams) error
	UpdateSyncDatabaseSourceService(ctx context.Context, id string, fields map[string]interface{}) error
	DeleteDatabaseSourceService(ctx context.Context, id string) error
	GetDatabaseSourceServiceById(ctx context.Context, id string) (*DatabaseSourceServiceParams, error)
}

type DatabaseSourceServiceUsecase struct {
	log                       *utilLog.Helper
	repo                      DatabaseSourceServiceRepo
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	namespaceUsecase          *NamespaceUsecase
	dbServiceUsecase          *DBServiceUsecase
	cron                      *cron.Cron
	lastSyncTime              time.Time
}

func NewDatabaseSourceServiceUsecase(log utilLog.Logger, repo DatabaseSourceServiceRepo, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, namespaceUsecase *NamespaceUsecase, dbServiceUsecase *DBServiceUsecase) *DatabaseSourceServiceUsecase {
	return &DatabaseSourceServiceUsecase{
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.database_source_service")),
		repo:                      repo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		namespaceUsecase:          namespaceUsecase,
		dbServiceUsecase:          dbServiceUsecase,
	}
}

type DatabaseSourceServiceParams struct {
	UID                 string          `json:"uid"`
	Name                string          `json:"name"`
	Source              string          `json:"source"`
	Version             string          `json:"version"`
	URL                 string          `json:"url"`
	DbType              pkgConst.DBType `json:"db_type"`
	CronExpress         string          `json:"cron_express"`
	NamespaceUID        string          `json:"namespace_uid"`
	LastSyncErr         string          `json:"last_sync_err"`
	LastSyncSuccessTime *time.Time      `json:"last_sync_success_time"`
	AdditionalParams    pkgParams.Params
	SQLEConfig          *SQLEConfig
	UpdatedAt           time.Time
}

type ListDatabaseSourceServiceTipsParams struct {
	Source  pkgConst.DBServiceSourceName `json:"source"`
	DbTypes []pkgConst.DBType            `json:"db_types"`
}
