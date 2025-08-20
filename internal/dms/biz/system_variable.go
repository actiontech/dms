package biz

import (
	"context"
	"fmt"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

const (
	SystemVariableWorkflowExpiredHours        = "system_variable_workflow_expired_hours"
	SystemVariableSqleUrl                     = "system_variable_sqle_url"
	SystemVariableOperationRecordExpiredHours = "system_variable_operation_record_expired_hours"
	SystemVariableCbOperationLogsExpiredHours = "system_variable_cb_operation_logs_expired_hours"
	SystemVariableSSHPrimaryKey               = "system_variable_ssh_primary_key"
)

const (
	DefaultOperationRecordExpiredHours        = 90 * 24
	DefaultCbOperationLogsExpiredHours        = 90 * 24
	DefaultSystemVariableWorkflowExpiredHours = 90 * 24
)

// SystemVariable 系统变量业务模型
type SystemVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SystemVariableRepo 系统变量存储接口
type SystemVariableRepo interface {
	GetSystemVariables(ctx context.Context) (map[string]SystemVariable, error)
	UpdateSystemVariables(ctx context.Context, variables []*SystemVariable) error
}

// SystemVariableUsecase 系统变量业务逻辑
type SystemVariableUsecase struct {
	repo SystemVariableRepo
	log  *utilLog.Helper
}

// NewSystemVariableUsecase 创建系统变量业务逻辑实例
func NewSystemVariableUsecase(log utilLog.Logger, repo SystemVariableRepo) *SystemVariableUsecase {
	return &SystemVariableUsecase{
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.system_variable")),
	}
}

// GetSystemVariables 获取所有系统变量
func (s *SystemVariableUsecase) GetSystemVariables(ctx context.Context) (map[string]SystemVariable, error) {
	s.log.Infof("GetSystemVariables")
	defer func() {
		s.log.Infof("GetSystemVariables completed")
	}()

	variables, err := s.repo.GetSystemVariables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system variables: %v", err)
	}

	return variables, nil
}

// UpdateSystemVariables 更新系统变量
func (s *SystemVariableUsecase) UpdateSystemVariables(ctx context.Context, variables []*SystemVariable) error {
	s.log.Infof("UpdateSystemVariables")
	defer func() {
		s.log.Infof("UpdateSystemVariables completed")
	}()

	if err := s.repo.UpdateSystemVariables(ctx, variables); err != nil {
		return fmt.Errorf("failed to update system variables: %v", err)
	}

	return nil
}
