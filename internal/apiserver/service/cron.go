package service

import (
	"context"
	"time"

	commonLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var cronManagerMap = map[string]cronImpl{}

const (
	DatabaseSourceService = "database_source_service"
)

type cronImpl interface {
	start(server *APIServer, groupCtx context.Context)
	stop()
}

func init() {
	ctx, cancel := context.WithCancel(context.Background())

	cronManagerMap[DatabaseSourceService] = &databaseSourceServiceCronManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func StartAllCronJob(server *APIServer, groupCtx context.Context) {
	for _, manager := range cronManagerMap {
		manager.start(server, groupCtx)
	}
}

func StopAllCronJob() {
	for _, manager := range cronManagerMap {
		manager.stop()
	}
}

const CronManager = "dms.cronmanager"

type databaseSourceServiceCronManager struct {
	apiServer *APIServer
	ctx       context.Context
	cancel    context.CancelFunc
}

func (c *databaseSourceServiceCronManager) start(server *APIServer, groupCtx context.Context) {
	c.apiServer = server

	go func() {
		logger := commonLog.NewHelper(c.apiServer.logger, commonLog.WithMessageKey(CronManager))
		logger.Info("cron start")

		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			// loop, wait apiServer initial
			if c.apiServer.DMSController != nil {
				break
			}
		}

		ticker.Stop()

		c.apiServer.DMSController.DMS.DBServiceSyncTaskUsecase.StartSyncDBServiceSyncTask()

		select {
		case <-groupCtx.Done():
			logger.Infof("cron terminal, err: %s", groupCtx.Err())

			c.apiServer.DMSController.DMS.DBServiceSyncTaskUsecase.StopSyncDBServiceSyncTask()
			return
		case <-c.ctx.Done():
			logger.Infof("cron terminal, err: %s", groupCtx.Err())

			c.apiServer.DMSController.DMS.DBServiceSyncTaskUsecase.StopSyncDBServiceSyncTask()
			return
		}
	}()
}

func (c *databaseSourceServiceCronManager) stop() {
	c.cancel()
}
