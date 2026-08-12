package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/actiontech/dms/internal/dms/biz"
	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/labstack/echo/v4"
)

const accessRestrictionDenyMsg = "禁止访问：来源 IP 不在访问白名单"

// AccessRestriction enforces IP/CIDR access restriction for all DMS entry points.
// Order (must not reorder short-circuits):
//  1. never-block register channel POST /v1/dms/proxys
//  2. switch off → allow
//  3. switch on → registered ProxyTarget host IP → allow
//  4. whitelist hit → allow
//  5. else HTTP 403 (not 401)
//
// Loopback is NOT auto-allowed. CloudBeaver paths are not exempted.
func AccessRestriction(u *biz.AccessRestrictionUsecase, proxy *biz.DmsProxyUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if isNeverBlockRegisterPath(c) {
				return next(c)
			}
			if u == nil {
				return next(c)
			}
			enabled, err := u.IsEnabled(c.Request().Context())
			if err != nil || !enabled {
				// Fail-open on read error; AC-003: disabled → allow.
				return next(c)
			}

			clientIP := biz.ExtractClientIP(c.Request())
			if proxy != nil && proxy.IsRegisteredServiceIP(clientIP) {
				return next(c)
			}
			matched, err := u.MatchClientIP(c.Request().Context(), clientIP)
			if err != nil {
				// Fail-open on match errors to avoid locking out operators on transient DB faults.
				return next(c)
			}
			if matched {
				return next(c)
			}
			msg := accessRestrictionDenyMsg
			if clientIP != "" {
				msg = fmt.Sprintf("%s（识别 IP：%s）", accessRestrictionDenyMsg, clientIP)
			}
			return echo.NewHTTPError(http.StatusForbidden, msg)
		}
	}
}

func isNeverBlockRegisterPath(c echo.Context) bool {
	if c.Request().Method != http.MethodPost {
		return false
	}
	path := strings.TrimSuffix(c.Request().URL.Path, "/")
	// Exact register channel: /v1/dms/proxys (ProxyRouterGroup under /v1).
	return path == "/v1"+dmsV1.ProxyRouterGroup
}
