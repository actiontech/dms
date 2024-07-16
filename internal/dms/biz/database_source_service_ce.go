//go:build !enterprise

package biz

import (
	"context"
	"errors"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

var errNotDBServiceSyncTask = errors.New("db service sync task related functions are enterprise version functions")

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTasks(ctx context.Context, conditions []pkgConst.FilterCondition, projectId string, currentUserId string) ([]*DBServiceSyncTaskParams, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) GetDBServiceSyncTask(ctx context.Context, dbServiceTaskId, projectId, currentUserId string) (*DBServiceSyncTaskParams, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) AddDBServiceSyncTask(ctx context.Context, params *DBServiceSyncTaskParams, currentUserId string) (string, error) {
	return "", errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) UpdateDBServiceSyncTask(ctx context.Context, dbServiceTaskId string, params *DBServiceSyncTaskParams, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) DeleteDBServiceSyncTask(ctx context.Context, dbServiceTaskId, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) ListDBServiceSyncTaskTips(ctx context.Context) ([]ListDBServiceSyncTaskTipsParams, error) {
	return nil, errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) SyncDBServices(ctx context.Context, dbServiceTaskId, currentUserId string) error {
	return errNotDBServiceSyncTask
}

func (d *DBServiceSyncTaskUsecase) StartSyncDBServices() {
}

func (d *DBServiceSyncTaskUsecase) StopSyncDBServices() {
}
