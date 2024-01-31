//go:build trial
// +build trial

package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/labstack/echo/v4"
)

var dbServiceLimit int64 = 20

func licenseAdapter(l *biz.LicenseUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := strings.TrimSuffix(c.Path(), "/")
			if !(c.Request().Method == http.MethodPost &&
				path == "/v1/dms/projects/:project_uid/db_services") {
				return next(c)
			}

			dbTypeCounts, err := l.DBService.CountDBService(c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}

			var total int64
			for _, count := range dbTypeCounts {
				total += count.Count
			}

			if total >= dbServiceLimit {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("the number of db services exceeds the limit of %v", dbServiceLimit))
			}

			return next(c)
		}
	}
}
