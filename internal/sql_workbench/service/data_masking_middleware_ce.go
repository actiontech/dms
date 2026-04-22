//go:build !enterprise
// +build !enterprise

package sql_workbench

import "github.com/labstack/echo/v4"

// GetDataMaskingMiddleware 返回用于脱敏的中间件
func GetDataMaskingMiddleware(config DataMaskingMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

// GetUnmaskingWorkflowMiddleware 处理获批后查看原文，执行 SQL 之前进行替换
func GetUnmaskingWorkflowMiddleware(config DataMaskingMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
