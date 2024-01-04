//go:build !enterprise

package biz

import (
	"context"

	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
)

func (cu *CloudbeaverUsecase) ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService, userId string) ([]*DBService, error) {
	return activeDBServices, nil
}

// SQLExecuteResultsDataMasking 为DMS企业版的脱敏功能，捕获cloudbeaver返回的结果集，根据配置对结果集脱敏
func (cu *CloudbeaverUsecase) SQLExecuteResultsDataMasking(ctx context.Context, result *model.SQLExecuteInfo) error {
	return nil
}
