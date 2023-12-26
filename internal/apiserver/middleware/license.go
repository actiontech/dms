package middleware

import (
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/labstack/echo/v4"
)

func LicenseAdapter(l *biz.LicenseUsecase) echo.MiddlewareFunc {
	//nolint:typecheck
	return licenseAdapter(l)
}
