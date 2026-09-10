//go:build !enterprise
// +build !enterprise

package sql_workbench

import "github.com/labstack/echo/v4"

// PrivilegeDeniedMiddlewareConfig 非 enterprise 构建占位。
type PrivilegeDeniedMiddlewareConfig struct {
	SqlWorkbenchService *SqlWorkbenchService
}

// GetPrivilegeDeniedMiddleware 非 enterprise 构建下为空操作。
func GetPrivilegeDeniedMiddleware(_ PrivilegeDeniedMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}
