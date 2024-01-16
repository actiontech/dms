package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/robfig/cron/v3"
)

type LicenseRepo interface {
	SaveLicense(ctx context.Context, license interface{}) error
	GetLastLicense(ctx context.Context) (interface{}, bool, error)
	GetLicenseById(ctx context.Context, id string) (interface{}, bool, error)
	UpdateLicense(ctx context.Context, license interface{}) error
	DelLicense(ctx context.Context) error
}

type LicenseUsecase struct {
	tx             TransactionGenerator
	repo           LicenseRepo
	userUsecase    *UserUsecase
	DBService      *DBServiceUsecase
	log            *utilLog.Helper
	cron           *cron.Cron
	clusterUsecase *ClusterUsecase
}

func NewLicenseUsecase(log utilLog.Logger, tx TransactionGenerator, repo LicenseRepo, usecase *UserUsecase, serviceUsecase *DBServiceUsecase, clusterUsecase *ClusterUsecase) *LicenseUsecase {
	lu := &LicenseUsecase{
		tx:             tx,
		repo:           repo,
		log:            utilLog.NewHelper(log, utilLog.WithMessageKey("biz.license")),
		userUsecase:    usecase,
		DBService:      serviceUsecase,
		clusterUsecase: clusterUsecase,
	}

	lu.initial()

	return lu
}
