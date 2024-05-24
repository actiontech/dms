//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportCbOperationLog = errors.New("cb operation log related functions are enterprise version functions")

func (cu *CbOperationLogUsecase) GetCbOperationLogByID(ctx context.Context, uid string) (*CbOperationLog, error) {
	return nil, errNotSupportCbOperationLog
}

func (u *CbOperationLogUsecase) SaveCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return errNotSupportCbOperationLog
}

func (u *CbOperationLogUsecase) UpdateCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return errNotSupportCbOperationLog
}

func (u *CbOperationLogUsecase) ListCbOperationLog(ctx context.Context, option *ListCbOperationLogOption, currentUid string, filterPersonID string, projectUid string) ([]*CbOperationLog, int64, error) {
	return nil, 0, errNotSupportCbOperationLog
}
