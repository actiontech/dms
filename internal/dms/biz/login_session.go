package biz

import (
	"context"
	"fmt"
	"net/http"

	jwtPkg "github.com/actiontech/dms/pkg/dms-common/api/jwt"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	"github.com/labstack/echo/v4"
)

type AuthLoginSessionUsecase struct {
	userUsecase               *UserUsecase
	loginConfigurationUsecase *LoginConfigurationUsecase
	log                       *utilLog.Helper
}

func NewAuthLoginSessionUsecase(log utilLog.Logger, userUsecase *UserUsecase, loginConfigurationUsecase *LoginConfigurationUsecase) *AuthLoginSessionUsecase {
	return &AuthLoginSessionUsecase{
		userUsecase:               userUsecase,
		loginConfigurationUsecase: loginConfigurationUsecase,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.login_session")),
	}
}

func (au *AuthLoginSessionUsecase) CheckSingleActiveSession(gatewayForwardedHeader string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get(gatewayForwardedHeader) != "" {
				return next(c)
			}

			tokenDetail, err := jwtPkg.GetTokenDetailFromContext(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("get token detail failed, err:%v", err))
			}
			if tokenDetail.UID == "" {
				return next(c)
			}
			if tokenDetail.LoginType == AccessTokenLogin {
				return next(c)
			}

			disableMultipleLogin, err := au.loginConfigurationUsecase.IsDisableMultipleLogin(c.Request().Context())
			if err != nil {
				return err
			}
			if !disableMultipleLogin {
				return next(c)
			}
			if tokenDetail.LoginSessionID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "login session is invalid, please login again")
			}

			latestSessionID, exists, err := au.userUsecase.GetLatestLoginSession(c.Request().Context(), tokenDetail.UID)
			if err != nil {
				return err
			}
			if !exists || latestSessionID != tokenDetail.LoginSessionID {
				return echo.NewHTTPError(http.StatusUnauthorized, "account has been logged in on another device")
			}

			return next(c)
		}
	}
}

func (d *UserUsecase) CreateLoginSession(ctx context.Context, userUID string) (sessionID string, err error) {
	sessionID, err = pkgRand.GenStrUid()
	if err != nil {
		return "", err
	}
	if err = d.RecordLoginSession(ctx, userUID, sessionID); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (d *UserUsecase) ValidateLoginSession(ctx context.Context, userUID, sessionID string) error {
	disableMultipleLogin, err := d.loginConfigurationUsecase.IsDisableMultipleLogin(ctx)
	if err != nil {
		return err
	}
	if !disableMultipleLogin {
		return nil
	}
	if sessionID == "" {
		return fmt.Errorf("login session is invalid, please login again")
	}
	latestSessionID, exists, err := d.GetLatestLoginSession(ctx, userUID)
	if err != nil {
		return err
	}
	if !exists || latestSessionID != sessionID {
		return fmt.Errorf("account has been logged in on another device")
	}
	return nil
}
