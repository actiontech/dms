//go:build !enterprise

package service

import (
	"context"
	"errors"

	aV1 "github.com/actiontech/dms/api/dms/service/v1"
	apiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
)

var errNotSupportOperationRecord = errors.New("OperationRecord related functions are enterprise version functions")

func (d *DMSService) AddOperationRecord(ctx context.Context, req *aV1.AddOperationRecordReq) (reply *aV1.AddOperationRecordReply, err error) {
	reply = &aV1.AddOperationRecordReply{}
	reply.GenericResp.SetCode(int(apiError.DMSServiceErr))
	reply.GenericResp.SetMsg(errNotSupportOperationRecord.Error())
	return reply, nil
}

func (d *DMSService) GetOperationRecordList(ctx context.Context, req *aV1.GetOperationRecordListReq, currentUserUid string) (reply *aV1.GetOperationRecordListReply, err error) {
	return nil, errNotSupportOperationRecord
}

func (d *DMSService) ExportOperationRecordList(ctx context.Context, req *aV1.ExportOperationRecordListReq, currentUserUid string) (reply []byte, err error) {
	return nil, errNotSupportOperationRecord
}
