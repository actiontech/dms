package sql_workbench

import (
	dataMaskingBiz "github.com/actiontech/dms/internal/data_masking/biz"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/sql_workbench/sqlresultmasker"
)

// DataMaskingMiddlewareConfig 配置脱敏中间件
type DataMaskingMiddlewareConfig struct {
	SqlResultMasker          sqlresultmasker.SQLResultMasker
	DBServiceUsecase         *biz.DBServiceUsecase
	SqlWorkbenchService      *SqlWorkbenchService
	UnmaskingWorkflowUsecase *dataMaskingBiz.UnmaskingWorkflowUsecase
}
