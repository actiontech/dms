//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotDBServiceSyncTask = errors.New("db service sync task related functions are enterprise version functions")

func (uc *DBServiceSyncTaskUsecase) AddDBServiceSyncTask(ctx context.Context, syncTask *DBServiceSyncTask, currentUserId string) (string, error) {
	return "", errNotDBServiceSyncTask
}

func (uc *DBServiceSyncTaskUsecase) checkPermission(ctx context.Context, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (uc *DBServiceSyncTaskUsecase) UpdateDBServiceSyncTask(ctx context.Context, syncTaskId string, updateValues *DBServiceSyncTask, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTasks(ctx context.Context, currentUserId string) ([]*DBServiceSyncTask, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTaskTips() ([]ListDBServiceSyncTaskTips, error) {
	return nil,errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) GetDBServiceSyncTask(ctx context.Context, syncTaskId, currentUserId string) (*DBServiceSyncTask, error) {
	return nil, errNotDBServiceSyncTask
}

func (uc *DBServiceSyncTaskUsecase) DeleteDBServiceSyncTask(ctx context.Context, syncTaskId, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (uc *DBServiceSyncTaskUsecase) SyncDBServices(ctx context.Context, syncTaskId, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) RestartSyncDBServiceSyncTask() {
}

func (d *DBServiceSyncTaskUsecase) StartSyncDBServiceSyncTask() {
}

func (d *DBServiceSyncTaskUsecase) StopSyncDBServiceSyncTask() {
}
