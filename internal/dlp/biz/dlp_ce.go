//go:build !enterprise

package biz

import (
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// 数据脱敏为DMS企业版功能
type DataLossProtectionUseCase struct {
}

func NewDataLossProtectionUseCase(log utilLog.Logger) (*DataLossProtectionUseCase, error) {
	d := &DataLossProtectionUseCase{}
	return d, nil
}
