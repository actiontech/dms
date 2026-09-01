//go:build !dms

package biz

import (
	"context"
)

type PrivilegeApplyWorkflowUsecase struct{}

func NewPrivilegeApplyWorkflowUsecase(ctx context.Context) *PrivilegeApplyWorkflowUsecase {
	return &PrivilegeApplyWorkflowUsecase{}
}

func (u *PrivilegeApplyWorkflowUsecase) GetPrivilegeApplyAssignees(ctx context.Context, projectUID, dbServiceUID, applicantUID string, mode AssigneeResolveMode) ([]string, error) {
	return nil, nil
}

func (u *PrivilegeApplyWorkflowUsecase) ListPrivilegeApplyAssigneesWithNames(ctx context.Context, projectUID, dbServiceUID string) ([]UIDWithName, bool, error) {
	return nil, false, nil
}

func (u *PrivilegeApplyWorkflowUsecase) CreatePrivilegeApplyWorkflow(ctx context.Context, args *CreatePrivilegeApplyWorkflowArgs) (string, error) {
	return "", nil
}

func (u *PrivilegeApplyWorkflowUsecase) GetPrivilegeApplyWorkflow(ctx context.Context, projectUID, workflowUID, userUID string) (*PrivilegeApplyWorkflow, error) {
	return nil, nil
}
