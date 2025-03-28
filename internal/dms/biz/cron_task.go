package biz

import (
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/robfig/cron/v3"
)

type CronTaskUsecase struct {
	log                   *utilLog.Helper
	cronTask              *cronTask
	workflowUsecase       *DataExportWorkflowUsecase
	cbOperationLogUsecase *CbOperationLogUsecase
	licenseUsecase        *LicenseUsecase
	oauth2SessionUsecase  *OAuth2SessionUsecase
}
type cronTask struct {
	cron *cron.Cron
}

func NewCronTaskUsecase(log utilLog.Logger, wu *DataExportWorkflowUsecase, cu *CbOperationLogUsecase, os *OAuth2SessionUsecase) *CronTaskUsecase {
	ctu := &CronTaskUsecase{
		log: utilLog.NewHelper(log, utilLog.WithMessageKey("biz.cronTask")),
		cronTask: &cronTask{
			cron: cron.New(),
		},
		workflowUsecase:       wu,
		cbOperationLogUsecase: cu,
		oauth2SessionUsecase:  os,
	}

	return ctu
}

func (ctu *CronTaskUsecase) InitialTask() error {
	if _, err := ctu.cronTask.cron.AddFunc("@daily", ctu.workflowUsecase.RecycleWorkflow); err != nil {
		return err
	}

	if _, err := ctu.cronTask.cron.AddFunc("@hourly", ctu.workflowUsecase.RecycleDataExportTask); err != nil {
		return err
	}

	if _, err := ctu.cronTask.cron.AddFunc("@hourly", ctu.workflowUsecase.RecycleDataExportTaskFiles); err != nil {
		return err
	}

	if _, err := ctu.cronTask.cron.AddFunc("@hourly", ctu.cbOperationLogUsecase.DoClean); err != nil {
		return err
	}

	if _, err := ctu.cronTask.cron.AddFunc("@hourly", ctu.oauth2SessionUsecase.DeleteExpiredSessions); err != nil {
		return err
	}

	ctu.cronTask.cron.Start()
	return nil
}
