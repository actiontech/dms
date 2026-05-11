//go:build !dms

package service

import (
	"context"
	"errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/data_masking/biz"
	dmsBiz "github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var errNotSupportUnmaskingWorkflow = errors.New("unmasking workflow related functions are enterprise version functions")

type unmaskingWorkflowUsecase = biz.UnmaskingWorkflowUsecase

func initUnmaskingWorkflowUsecase(_ utilLog.Logger, _ *storage.Storage, _ dmsBiz.ProxyTargetRepo, _ *dmsBiz.OpPermissionVerifyUsecase, _ *dmsBiz.UserUsecase, _ *dmsBiz.DBServiceUsecase) (*unmaskingWorkflowUsecase, error) {
	return nil, nil
}

func (d *DMSService) CreateUnmaskingWorkflow(ctx context.Context, req *v1.CreateUnmaskingWorkflowReq, currentUserUid string) (*v1.CreateUnmaskingWorkflowReply, error) {
	return nil, errNotSupportUnmaskingWorkflow
}

func (d *DMSService) GetUnmaskingWorkflow(ctx context.Context, req *v1.GetUnmaskingWorkflowReq, currentUserUid string) (*v1.GetUnmaskingWorkflowReply, error) {
	return nil, errNotSupportUnmaskingWorkflow
}

func (d *DMSService) ListUnmaskingWorkflows(ctx context.Context, req *v1.ListUnmaskingWorkflowsReq, currentUserUid string) (*v1.ListUnmaskingWorkflowsReply, error) {
	return nil, errNotSupportUnmaskingWorkflow
}

func (d *DMSService) ApproveUnmaskingWorkflow(ctx context.Context, req *v1.ApproveUnmaskingWorkflowReq, currentUserUid string) error {
	return errNotSupportUnmaskingWorkflow
}

func (d *DMSService) RejectUnmaskingWorkflow(ctx context.Context, req *v1.RejectUnmaskingWorkflowReq, currentUserUid string) error {
	return errNotSupportUnmaskingWorkflow
}

func (d *DMSService) CancelUnmaskingWorkflow(ctx context.Context, req *v1.CancelUnmaskingWorkflowReq, currentUserUid string) error {
	return errNotSupportUnmaskingWorkflow
}
