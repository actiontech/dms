//go:build !enterprise

package biz

import (
	"context"
	"errors"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

var errNotDBServiceSyncTask = errors.New("database source service related functions are enterprise version functions")

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTasks(ctx context.Context, conditions []pkgConst.FilterCondition, projectId string, currentUserId string) ([]*DBServiceSyncTaskParams, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) GetDBServiceSyncTask(ctx context.Context, databaseSourceServiceId, projectId, currentUserId string) (*DBServiceSyncTaskParams, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) AddDBServiceSyncTask(ctx context.Context, params *DBServiceSyncTaskParams, currentUserId string) (string, error) {
	return "", errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) UpdateDBServiceSyncTask(ctx context.Context, databaseSourceServiceId string, params *DBServiceSyncTaskParams, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) DeleteDBServiceSyncTask(ctx context.Context, databaseSourceServiceId, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTaskTips(ctx context.Context) ([]ListDBServiceSyncTaskTipsParams, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) SyncDBServiceSyncTask(ctx context.Context, databaseSourceServiceId, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) StartSyncDBServiceSyncTask() {
}

func (d *DBServiceSyncTaskUsecase) StopSyncDBServiceSyncTask() {
}
