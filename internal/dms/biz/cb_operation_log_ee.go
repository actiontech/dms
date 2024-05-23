//go:build enterprise

package biz

import "context"

func (cu *CbOperationLogUsecase) GetCbOperationLogByID(ctx context.Context, uid string) (*CbOperationLog, error) {
	operationLog, err := cu.repo.GetCbOperationLogByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return operationLog, nil
}

func (u *CbOperationLogUsecase) SaveCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return u.repo.SaveCbOperationLog(ctx, log)
}

func (u *CbOperationLogUsecase) UpdateCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return u.repo.UpdateCbOperationLog(ctx, log)
}
