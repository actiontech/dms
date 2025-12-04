package sql_workbench

import (
	"github.com/actiontech/dms/internal/dms/biz"
)

// OperationLogMiddlewareConfig 配置操作日志中间件
type OperationLogMiddlewareConfig struct {
	CbOperationLogUsecase *biz.CbOperationLogUsecase
	DBServiceUsecase      *biz.DBServiceUsecase
	SqlWorkbenchService   *SqlWorkbenchService
}
