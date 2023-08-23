package service

import (
	"context"
	"time"

	"github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type cronManager struct {
	apiServer *APIServer
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewCronManager(server *APIServer) *cronManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &cronManager{
		ctx:       ctx,
		cancel:    cancel,
		apiServer: server,
	}
}

const CronManager = "dms.cronmanager"

func (c *cronManager) Start() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	_ = c.apiServer.logger.Log(log.LevelInfo, CronManager, "cron start")

	for {
		select {
		case <-ticker.C:
			_ = c.apiServer.logger.Log(log.LevelInfo, CronManager, "cron running")

			c.apiServer.DMSController.DMS.DatabaseSourceServiceUsecase.StartSyncDatabaseSourceService()
		case <-c.ctx.Done():
			_ = c.apiServer.logger.Log(log.LevelInfo, CronManager, "cron terminal")

			c.apiServer.DMSController.DMS.DatabaseSourceServiceUsecase.StopSyncDatabaseSourceService()
			return
		}
	}
}

func (c *cronManager) Stop() {
	c.cancel()
}
