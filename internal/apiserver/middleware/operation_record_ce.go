//go:build !enterprise

package middleware

import (
	"github.com/actiontech/dms/internal/dms/service"
	"github.com/labstack/echo/v4"
)

func OperationRecordMiddleware(_ *service.DMSService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}
