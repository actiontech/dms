//go:build release
// +build release

package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/license"
	"github.com/labstack/echo/v4"
)

func licenseAdapter(l *biz.LicenseUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			check := &LicenseChecker{ctx: c, LicenseUsecase: l}
			err := check.Check()
			if err != nil {
				return fmt.Errorf("the operation is outside the scope of the license, %v", err)
			}
			return next(c)
		}
	}
}

type LicenseCheckerFn func() (skip bool, err error)

type LicenseChecker struct {
	ctx            echo.Context
	license        *license.License
	LicenseUsecase *biz.LicenseUsecase
}

func (c *LicenseChecker) getLicense() (*license.License, error) {
	if c.license != nil {
		return c.license, nil
	}

	l, exist, err := c.LicenseUsecase.GetLicenseInner(c.ctx.Request().Context())
	if err != nil {
		return nil, err
	}
	if !exist || l.Content == nil {
		return nil, fmt.Errorf("license is empty")
	}
	c.license = l.Content
	return c.license, nil
}

func (c *LicenseChecker) Check() error {
	var checks = []LicenseCheckerFn{
		c.allowGetMethod,
		c.allowApi,
		c.checkHardware,
		c.checkIsExpired,
		c.checkCreateUser,
		c.checkCreateInstance,
	}
	for _, check := range checks {
		skip, err := check()
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
	}
	return nil
}

const (
	LimitTypeUser = "user"
)

func (c *LicenseChecker) allowGetMethod() (bool, error) {
	if c.ctx.Request().Method == http.MethodGet {
		return true, nil
	}
	return false, nil
}

func (c *LicenseChecker) allowApi() (bool, error) {
	path := strings.TrimSuffix(c.ctx.Path(), "/")
	switch path {
	case "/v1/dms/sessions",
		"/v1/dms/configurations/license",
		"/v1/dms/configurations/license/check",
		"/v1/dms/configurations/license/info",
		"/v1/dms/proxys",
		"/v1/dms/plugins":
		return true, nil
	default:
		return false, nil
	}
}

func (c *LicenseChecker) checkHardware() (bool, error) {
	l, err := c.getLicense()
	if err != nil {
		return true, err
	}
	s, err := license.CollectHardwareInfo()
	if err != nil {
		return false, err
	}
	err = l.CheckHardwareSignIsMatch(s)
	if err != nil {
		return false, err
	}
	return false, nil
}

func (c *LicenseChecker) checkIsExpired() (bool, error) {
	l, err := c.getLicense()
	if err != nil {
		return true, err
	}
	if err := l.CheckLicenseNotExpired(); err != nil {
		return true, err
	}
	return false, nil
}

func (c *LicenseChecker) checkCreateUser() (bool, error) {
	path := strings.TrimSuffix(c.ctx.Path(), "/")
	if c.ctx.Request().Method == http.MethodPost && path == "/v1/dms/users" {
		return c.LicenseUsecase.CheckCanCreateUser(c.ctx.Request().Context())

	}
	return false, nil
}

func (c *LicenseChecker) checkCreateInstance() (bool, error) {
	path := strings.TrimSuffix(c.ctx.Path(), "/")
	if !(c.ctx.Request().Method == http.MethodPost &&
		path == "/v1/dms/projects/:project_uid/db_services") {
		return false, nil
	}

	dbType, err := getDBTypeWithReq(c.ctx)
	if err != nil {
		return false, err
	}
	return c.LicenseUsecase.CheckCanCreateInstance(c.ctx.Request().Context(), dbType)
}

func getDBTypeWithReq(c echo.Context) (dbType string, err error) {
	bodyBytes, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	req := new(v1.AddDBServiceReq)
	if err := c.Bind(req); err != nil {
		return "", fmt.Errorf("failed to bind request: %v", err)
	}

	if err != nil {
		return "", err
	}
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if req.DBService.DBType == "" {
		return "MySQL", nil
	}
	return req.DBService.DBType, nil
}
