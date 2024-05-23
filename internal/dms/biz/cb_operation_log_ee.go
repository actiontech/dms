//go:build enterprise

package biz

import (
	"context"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

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

func (u *CbOperationLogUsecase) ListCbOperationLog(ctx context.Context, option *ListCbOperationLogOption, currentUid string, filterPersonID string, projectUid string) ([]*CbOperationLog, int64, error) {
	// 只有管理员可以查看所有操作日志, 其他用户只能查看自己的操作日志
	if currentUid != pkgConst.UIDOfUserSys {
		if isAdmin, err := u.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUid, projectUid); err != nil {
			return nil, 0, err
		} else if isAdmin {
			// do nothing,skip to next,because admin can view all operation logs
		} else if currentUid != filterPersonID {
			return nil, 0, nil
		}
	}

	operationLogs, count, err := u.repo.ListCbOperationLogs(ctx, option)
	if err != nil {
		return nil, 0, err
	}

	return operationLogs, count, nil
}
