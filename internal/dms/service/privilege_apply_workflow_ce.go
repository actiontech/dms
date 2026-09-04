//go:build !dms

package service

import (
	"context"
	"errors"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	privilegeApplyBiz "github.com/actiontech/dms/internal/privilege_apply/biz"
	dmsBiz "github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var errNotSupportPrivilegeApplyWorkflow = errors.New("privilege apply workflow related functions are enterprise version functions")

type privilegeApplyWorkflowUsecase = privilegeApplyBiz.PrivilegeApplyWorkflowUsecase

func initPrivilegeApplyWorkflowUsecase(_ utilLog.Logger, _ *storage.Storage, _ dmsBiz.ProxyTargetRepo, _ *dmsBiz.OpPermissionVerifyUsecase, _ *dmsBiz.UserUsecase, _ *dmsBiz.DBServiceUsecase) (*privilegeApplyWorkflowUsecase, error) {
	return nil, nil
}

func (d *DMSService) GetPrivilegeApplyAssignees(ctx context.Context, req *v1.GetPrivilegeApplyAssigneesReq) (*v1.GetPrivilegeApplyAssigneesReply, error) {
	return nil, errNotSupportPrivilegeApplyWorkflow
}

func (d *DMSService) CreatePrivilegeApplyWorkflow(ctx context.Context, req *v1.CreatePrivilegeApplyWorkflowReq, currentUserUID string) (*v1.CreatePrivilegeApplyWorkflowReply, error) {
	return nil, errNotSupportPrivilegeApplyWorkflow
}

func (d *DMSService) GetPrivilegeApplyWorkflow(ctx context.Context, req *v1.GetPrivilegeApplyWorkflowReq, currentUserUID string) (*v1.GetPrivilegeApplyWorkflowReply, error) {
	return nil, errNotSupportPrivilegeApplyWorkflow
}

func (d *DMSService) ListPrivilegeApplyWorkflows(ctx context.Context, req *v1.ListPrivilegeApplyWorkflowsReq, currentUserUID string) (*v1.ListPrivilegeApplyWorkflowsReply, error) {
	return nil, errNotSupportPrivilegeApplyWorkflow
}

func (d *DMSService) ApprovePrivilegeApplyWorkflow(ctx context.Context, req *v1.ApprovePrivilegeApplyWorkflowReq, currentUserUID string) error {
	return errNotSupportPrivilegeApplyWorkflow
}

func (d *DMSService) RejectPrivilegeApplyWorkflow(ctx context.Context, req *v1.RejectPrivilegeApplyWorkflowReq, currentUserUID string) error {
	return errNotSupportPrivilegeApplyWorkflow
}

func (d *DMSService) RetryReissuePrivilegeApplyWorkflow(ctx context.Context, req *v1.RetryReissuePrivilegeApplyWorkflowReq, currentUserUID string) error {
	return errNotSupportPrivilegeApplyWorkflow
}
