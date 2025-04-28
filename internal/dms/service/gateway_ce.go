//go:build !enterprise

package service

import (
	"context"
	"errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
)

var errNotSupportGateways = errors.New("gateways related functions are enterprise version functions")

func (d *DMSService) AddGateway(ctx context.Context, req *v1.Gateway) (err error) {
	return errNotSupportImportDBServices
}

func (d *DMSService) DeleteGateway(ctx context.Context, req *v1.DeleteGatewayReq) (err error) {
	return errNotSupportImportDBServices
}

func (d *DMSService) UpdateGateway(ctx context.Context, req *v1.UpdateGatewayReq) (err error) {
	return errNotSupportImportDBServices
}

func (d *DMSService) SyncGateways(ctx context.Context, req *v1.SyncGatewayReq) (err error) {
	return errNotSupportImportDBServices
}

func (d *DMSService) GetGateway(ctx context.Context, req *v1.GetGatewayReq) (reply *v1.GetGatewayReply, err error) {
	return nil, errNotSupportImportDBServices
}

func (d *DMSService) GetGatewayTips(ctx context.Context) (reply *v1.GetGatewayTipsReply, err error) {
	return nil, errNotSupportImportDBServices
}

func (d *DMSService) ListGateways(ctx context.Context, req *v1.ListGatewaysReq) (reply *v1.ListGatewaysReply, err error) {
	return nil, errNotSupportImportDBServices
}
