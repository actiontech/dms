//go:build !enterprise
// +build !enterprise

package sql_workbench

import "github.com/labstack/echo/v4"

// GetOperationLogBodyDumpMiddleware 返回用于记录操作日志的 BodyDump 中间件
func GetOperationLogBodyDumpMiddleware(config OperationLogMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
