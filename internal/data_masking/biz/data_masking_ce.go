//go:build !enterprise

package biz

import (
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// 数据脱敏为DMS企业版功能
type DataMaskingUseCase struct {
}

func NewDataMaskingUseCase(log utilLog.Logger) (*DataMaskingUseCase, error) {
	d := &DataMaskingUseCase{}
	return d, nil
}
