//go:build !release
// +build !release

package middleware

import (
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/labstack/echo/v4"
)

func licenseAdapter(l *biz.LicenseUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
