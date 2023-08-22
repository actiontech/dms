//go:build !enterprise

package biz

import (
	"context"
	"errors"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

var errNotDatabaseSourceService = errors.New("database source service related functions are enterprise version functions")

func (d *DatabaseSourceServiceUsecase) ListDatabaseSourceServices(ctx context.Context, conditions []pkgConst.FilterCondition, namespaceId string, currentUserId string) ([]*DatabaseSourceServiceParams, error) {
	return nil, errNotSupportOauth2
}

func (d *DatabaseSourceServiceUsecase) GetDatabaseSourceService(ctx context.Context, databaseSourceServiceId, namespaceId, currentUserId string) (*DatabaseSourceServiceParams, error) {
	return nil, errNotSupportOauth2
}

func (d *DatabaseSourceServiceUsecase) AddDatabaseSourceService(ctx context.Context, params *DatabaseSourceServiceParams, currentUserId string) (string, error) {
	return "", errNotDatabaseSourceService

}

func (d *DatabaseSourceServiceUsecase) UpdateDatabaseSourceService(ctx context.Context, databaseSourceServiceId string, params *DatabaseSourceServiceParams, currentUserId string) error {
	return errNotSupportOauth2
}

func (d *DatabaseSourceServiceUsecase) DeleteDatabaseSourceService(ctx context.Context, databaseSourceServiceId, currentUserId string) error {
	return errNotSupportOauth2
}

func (d *DatabaseSourceServiceUsecase) ListDatabaseSourceServiceTips(ctx context.Context) ([]ListDatabaseSourceServiceTipsParams, error) {
	return nil, errNotDatabaseSourceService
}

func (d *DatabaseSourceServiceUsecase) SyncDatabaseSourceService(ctx context.Context, databaseSourceServiceId, currentUserId string) error {
	return errNotDatabaseSourceService
}
