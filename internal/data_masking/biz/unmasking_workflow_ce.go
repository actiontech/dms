//go:build !dms

package biz

import (
	"context"
)

type UnmaskingWorkflowUsecase struct {
}

// UnmaskingSQL 非 dms 构建占位；与 enterprise 中间件读取的字段对齐。
type UnmaskingSQL struct {
	UID        string
	SQLContent string
}

// UnmaskingWorkflow 非 dms 构建下的占位类型；enterprise 会读 UID / ProjectUID 等（如 data_export_workflow_ee）。
type UnmaskingWorkflow struct {
	UID           string
	ProjectUID    string
	ApplicantUID  string
	UnmaskingSQLs []*UnmaskingSQL
}

type CreateUnmaskingWorkflowArgs struct{}

type UnmaskingDBConfig struct{}

type ListUnmaskingWorkflowsOption struct{}

func NewUnmaskingWorkflowUsecase(ctx context.Context) *UnmaskingWorkflowUsecase {
	return &UnmaskingWorkflowUsecase{}
}

func (u *UnmaskingWorkflowUsecase) CreateUnmaskingWorkflow(ctx context.Context, args *CreateUnmaskingWorkflowArgs) (string, error) {
	return "", nil
}

func (u *UnmaskingWorkflowUsecase) GetUnmaskingWorkflow(ctx context.Context, workflowUID string) (*UnmaskingWorkflow, error) {
	return nil, nil
}

func (u *UnmaskingWorkflowUsecase) ListUnmaskingWorkflows(ctx context.Context, opt *ListUnmaskingWorkflowsOption) ([]*UnmaskingWorkflow, error) {
	return nil, nil
}

func (u *UnmaskingWorkflowUsecase) ApproveUnmaskingWorkflow(ctx context.Context, workflowUID string) error {
	return nil
}

func (u *UnmaskingWorkflowUsecase) RejectUnmaskingWorkflow(ctx context.Context, workflowUID string) error {
	return nil
}

func (u *UnmaskingWorkflowUsecase) CancelUnmaskingWorkflow(ctx context.Context, workflowUID string) error {
	return nil
}

func (u *UnmaskingWorkflowUsecase) GetUnmaskingWorkflowAssignees(ctx context.Context, projectUID, datasourceUID, applicantUID string) ([]string, error) {
	return nil, nil
}

func (u *UnmaskingWorkflowUsecase) CheckOriginalDataAccess(ctx context.Context, projectUID, datasourceUID, userUID string, credential *UnmaskingOriginalDataCredentialArgs) (bool, *UnmaskingWorkflow, error) {
	return false, nil, nil
}

// MarkWorkflowUsage 完整实现仅在 dms 构建中提供；非 dms 占位。
func (u *UnmaskingWorkflowUsecase) MarkWorkflowUsage(_ context.Context, _, _, _ string, _ UnmaskingAction) error {
	return nil
}

// GetUnmaskingWorkflowDetail 完整实现仅在 dms 构建中提供；非 dms 占位（运行时 enterprise 且未启用 dms 时 usecase 一般为 nil）。
func (u *UnmaskingWorkflowUsecase) GetUnmaskingWorkflowDetail(_ context.Context, _, _, _ string) (*UnmaskingWorkflow, error) {
	return &UnmaskingWorkflow{}, nil
}

// AnalyzeLineageAndBuildMaskingSnapshot 完整实现仅在 dms 构建中提供；非 dms 构建占位以满足跨标签编译。
func (u *UnmaskingWorkflowUsecase) AnalyzeLineageAndBuildMaskingSnapshot(
	_ context.Context,
	_, _, _, _ string,
) (*AnalyzeResult, []*ColumnMaskingConfig) {
	return nil, nil
}
