package biz

import utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

type ResourceOverviewUsecase struct {
	log                       *utilLog.Helper
	projectRepo               ProjectRepo
	dbServiceRepo             DBServiceRepo
	opPermissionVerifyUsecase OpPermissionVerifyUsecase
}

func NewResourceOverviewUsecase(
	log utilLog.Logger,
	projectRepo ProjectRepo,
	dbServiceRepo DBServiceRepo,
	opPermissionVerifyUsecase OpPermissionVerifyUsecase,
) *ResourceOverviewUsecase {
	return &ResourceOverviewUsecase{
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.ResourceOverview")),
		projectRepo:               projectRepo,
		dbServiceRepo:             dbServiceRepo,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
	}
}

type ResourceOverviewVisibility string
