//go:build !enterprise

package service

import (
	"context"
	"errors"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

var errNotSupportCBOperationLog = errors.New("cloudbeaver operation log related functions are enterprise version functions")

func (d *DMSService) listCBOperationLogs(ctx context.Context, req *dmsV1.ListCBOperationLogsReq, uid string) (reply *dmsV1.ListCBOperationLogsReply, err error) {
	return nil, errNotSupportCBOperationLog
}

func (d *DMSService) getCBOperationLogTips(ctx context.Context, req *dmsV1.GetCBOperationLogTipsReq, currentUid string) (*dmsV1.GetCBOperationLogTipsReply, error) {
	return nil, errNotSupportCBOperationLog
}
