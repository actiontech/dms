//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/robfig/cron/v3"
)

func (uc *DBServiceSyncTaskUsecase) AddDBServiceSyncTask(ctx context.Context, syncTask *DBServiceSyncTask, currentUserId string) (string, error) {

	if err := uc.checkPermission(ctx, currentUserId); err != nil {
		return "", err
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return "", err
	}
	syncTask.UID = uid
	if err = uc.repo.SaveDBServiceSyncTask(ctx, syncTask); err != nil {
		return "", err
	}

	uc.RestartSyncDBServiceSyncTask()
	return uid, nil
}

// 权限验证：检查当前用户是否是管理员或者具有全局操作权限
func (uc *DBServiceSyncTaskUsecase) checkPermission(ctx context.Context, currentUserId string) error {
	hasGlobalOpPermission, err := uc.opPermissionVerifyUsecase.CanOpGlobal(ctx, currentUserId)
	if err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	}
	if !hasGlobalOpPermission {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}
	return nil
}

func (uc *DBServiceSyncTaskUsecase) UpdateDBServiceSyncTask(ctx context.Context, syncTaskId string, updateValues *DBServiceSyncTask, currentUserId string) error {

	if err := uc.checkPermission(ctx, currentUserId); err != nil {
		return err
	}

	syncTask, err := uc.repo.GetDBServiceSyncTaskById(ctx, syncTaskId)
	if err != nil {
		return fmt.Errorf("get db_service_sync_task failed: %v", err)
	}

	if syncTask.DbType != updateValues.DbType {
		return fmt.Errorf("update db_service_sync_task type is unsupported")
	}

	updateValues.UID = syncTaskId
	updateValues.LastSyncErr = syncTask.LastSyncErr
	updateValues.LastSyncSuccessTime = syncTask.LastSyncSuccessTime

	if err = uc.repo.UpdateDBServiceSyncTask(ctx, updateValues); err != nil {
		return err
	}

	uc.RestartSyncDBServiceSyncTask()
	return nil
}

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTasks(ctx context.Context, currentUserId string) ([]*DBServiceSyncTask, error) {
	syncTasks, err := d.repo.ListDBServiceSyncTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list db_service_sync_task failed: %w", err)
	}
	return syncTasks, nil
}

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTaskTips(ctx context.Context) ([]ListDBServiceSyncTaskTips, error) {
	return d.repo.ListDBServiceSyncTaskTips(ctx)
}

func (d *DBServiceSyncTaskUsecase) GetDBServiceSyncTask(ctx context.Context, syncTaskId, currentUserId string) (*DBServiceSyncTask, error) {

	service, err := d.repo.GetDBServiceSyncTaskById(ctx, syncTaskId)
	if err != nil {
		return nil, fmt.Errorf("get db_service_sync_task failed: %w", err)
	}

	return service, nil
}

func (uc *DBServiceSyncTaskUsecase) DeleteDBServiceSyncTask(ctx context.Context, syncTaskId, currentUserId string) error {
	_, err := uc.repo.GetDBServiceSyncTaskById(ctx, syncTaskId)
	if err != nil {
		return fmt.Errorf("get db_service_sync_task failed: %v", err)
	}
	err = uc.checkPermission(ctx, currentUserId)
	if err != nil {
		return err
	}
	//todo: currently only database_source_services data is deleted
	if err = uc.repo.DeleteDBServiceSyncTask(ctx, syncTaskId); err != nil {
		return err
	}

	uc.RestartSyncDBServiceSyncTask()
	return nil
}

func (uc *DBServiceSyncTaskUsecase) SyncDBServices(ctx context.Context, syncTaskId, currentUserId string) error {
	dbServiceSyncTask, err := uc.repo.GetDBServiceSyncTaskById(ctx, syncTaskId)
	if err != nil {
		return fmt.Errorf("get db_service_sync_task failed: %v", err)
	}
	err = uc.checkPermission(ctx, currentUserId)
	if err != nil {
		return err
	}
	sourceName, err := pkgConst.ParseDBServiceSource(dbServiceSyncTask.Source)
	if err != nil {
		return err
	}

	databaseSourceImpl, err := NewDatabaseSourceImpl(sourceName, uc)
	if err != nil {
		return err
	}

	// sync database source
	syncErr := databaseSourceImpl.SyncDatabaseSource(ctx, dbServiceSyncTask, currentUserId)
	fields := make(map[string]interface{})
	if syncErr != nil {
		fields["last_sync_err"] = syncErr.Error()
	} else {
		currentTime := time.Now()
		fields["last_sync_err"] = ""
		fields["last_sync_success_time"] = &currentTime

	}
	if err = uc.repo.UpdateDBServiceSyncTaskByFields(ctx, dbServiceSyncTask.UID, fields); err != nil {
		return fmt.Errorf("update sync database source err: %v, sync err: %v", err, syncErr)
	}

	return syncErr
}

func (d *DBServiceSyncTaskUsecase) RestartSyncDBServiceSyncTask() {
	d.StopSyncDBServiceSyncTask()
	d.StartSyncDBServiceSyncTask()
}

func (d *DBServiceSyncTaskUsecase) StartSyncDBServiceSyncTask() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	services, err := d.repo.ListDBServiceSyncTasks(ctx)
	if err != nil {
		d.log.Errorf("start timed sync err: %v", err)
		return
	}

	if d.cron == nil {
		d.cron = cron.New()
	}

	for _, service := range services {
		_, err := d.cron.AddFunc(service.CronExpress, func() {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err = d.SyncDBServices(ctx, service.UID, pkgConst.UIDOfUserAdmin); err != nil {
				d.log.Errorf("sync database_source_service err: %d", err)
			}
		})

		d.log.Infof("add database_source_service cron: name: %s, err: %v", service.Name, err)
	}

	d.cron.Start()
}

func (d *DBServiceSyncTaskUsecase) StopSyncDBServiceSyncTask() {
	if d.cron != nil {
		d.log.Infof("stop sync database source cron")
		d.cron.Stop()
		d.cron = cron.New()
	}
}
