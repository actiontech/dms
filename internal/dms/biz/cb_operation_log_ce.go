//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportCbOperationLog = errors.New("cb operation log related functions are enterprise version functions")

func (u *CbOperationLogUsecase) SaveCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return errNotSupportCbOperationLog
}

func (u *CbOperationLogUsecase) UpdateCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return errNotSupportCbOperationLog
}
