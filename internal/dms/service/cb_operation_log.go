package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) ListCBOperationLogs(ctx context.Context, req *dmsV1.ListCBOperationLogsReq, uid string) (reply *dmsV1.ListCBOperationLogsReply, err error) {
	return d.listCBOperationLogs(ctx, req, uid)
}

func (d *DMSService) GetCBOperationLogTips(ctx context.Context, req *dmsV1.GetCBOperationLogTipsReq, uid string) (reply *dmsV1.GetCBOperationLogTipsReply, err error) {
	return d.getCBOperationLogTips(ctx, req, uid)
}

func (d *DMSService) ExportCBOperationLogs(ctx context.Context, req *dmsV1.ExportCBOperationLogsReq, uid string) ([]byte, error) {
	return d.exportCbOperationLogs(ctx, req, uid)
}
