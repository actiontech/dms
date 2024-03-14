//go:build !dms

package biz

import (
	"context"
	"errors"

	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
)

var errNotDataMasking = errors.New("data masking unimplemented")

func (d *DataMaskingUsecase) ListMaskingRules(ctx context.Context) ([]ListMaskingRule, error) {
	return nil, errNotDataMasking
}

// SQLExecuteResultsDataMasking 为DMS企业版的脱敏功能，捕获cloudbeaver返回的结果集，根据配置对结果集脱敏
func (d *DataMaskingUsecase) SQLExecuteResultsDataMasking(ctx context.Context, result *model.SQLExecuteInfo) error {
	return nil
}

func IsDMS() bool {
	return false
}
