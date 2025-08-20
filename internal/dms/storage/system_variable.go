package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.SystemVariableRepo = (*SystemVariableRepo)(nil)

// SystemVariableRepo 系统变量存储实现
type SystemVariableRepo struct {
	*Storage
	log *utilLog.Helper
}

// NewSystemVariableRepo 创建系统变量存储实例
func NewSystemVariableRepo(log utilLog.Logger, s *Storage) *SystemVariableRepo {
	return &SystemVariableRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.system_variable"))}
}

// GetSystemVariables 获取所有系统变量
func (s *SystemVariableRepo) GetSystemVariables(ctx context.Context) (map[string]biz.SystemVariable, error) {
	var variables []*model.SystemVariable

	if err := transaction(s.log, ctx, s.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Find(&variables).Error; err != nil {
			return fmt.Errorf("failed to get system variables: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return convertModelSystemVariables(variables), nil
}

// UpdateSystemVariables 更新系统变量
func (s *SystemVariableRepo) UpdateSystemVariables(ctx context.Context, variables []*biz.SystemVariable) error {
	if err := transaction(s.log, ctx, s.db, func(tx *gorm.DB) error {
		for _, variable := range variables {
			modelVar := &model.SystemVariable{
				Key:   variable.Key,
				Value: variable.Value,
			}

			// 使用 Upsert 操作，如果存在则更新，不存在则插入
			if err := tx.WithContext(ctx).Save(modelVar).Error; err != nil {
				return fmt.Errorf("failed to update system variable %s: %v", variable.Key, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// convertModelSystemVariables 转换存储模型到业务模型
func convertModelSystemVariables(variables []*model.SystemVariable) map[string]biz.SystemVariable {
	sysVariables := make(map[string]biz.SystemVariable, len(variables))
	for _, sv := range variables {
		sysVariables[sv.Key] = biz.SystemVariable{
			Key:   sv.Key,
			Value: sv.Value,
		}
	}

	if _, ok := sysVariables[biz.SystemVariableOperationRecordExpiredHours]; !ok {
		sysVariables[biz.SystemVariableOperationRecordExpiredHours] = biz.SystemVariable{
			Key:   biz.SystemVariableOperationRecordExpiredHours,
			Value: strconv.Itoa(biz.DefaultOperationRecordExpiredHours),
		}
	}

	if _, ok := sysVariables[biz.SystemVariableCbOperationLogsExpiredHours]; !ok {
		sysVariables[biz.SystemVariableCbOperationLogsExpiredHours] = biz.SystemVariable{
			Key:   biz.SystemVariableCbOperationLogsExpiredHours,
			Value: strconv.Itoa(biz.DefaultCbOperationLogsExpiredHours),
		}
	}

	if _, ok := sysVariables[biz.SystemVariableWorkflowExpiredHours]; !ok {
		sysVariables[biz.SystemVariableWorkflowExpiredHours] = biz.SystemVariable{
			Key:   biz.SystemVariableWorkflowExpiredHours,
			Value: strconv.Itoa(biz.DefaultSystemVariableWorkflowExpiredHours),
		}
	}

	if _, ok := sysVariables[biz.SystemVariableSqlManageRawExpiredHours]; !ok {
		sysVariables[biz.SystemVariableSqlManageRawExpiredHours] = biz.SystemVariable{
			Key:   biz.SystemVariableSqlManageRawExpiredHours,
			Value: strconv.Itoa(biz.DefaultSystemVariableSqlManageRawExpiredHours),
		}
	}

	return sysVariables
}
