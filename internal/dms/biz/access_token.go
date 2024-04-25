package biz

import (
	"fmt"
	"net/http"

	jwtPkg "github.com/actiontech/dms/pkg/dms-common/api/jwt"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"github.com/labstack/echo/v4"
)

const AccessTokenLogin = "access_token_login"

type AuthAccessTokenUsecase struct {
	userUsecase *UserUsecase
	log         *utilLog.Helper
}

func NewAuthAccessTokenUsecase(log utilLog.Logger, usecase *UserUsecase) *AuthAccessTokenUsecase {
	au := &AuthAccessTokenUsecase{
		userUsecase: usecase,
		log:         utilLog.NewHelper(log, utilLog.WithMessageKey("biz.accesstoken")),
	}
	return au
}

func (au *AuthAccessTokenUsecase) CheckLatestAccessToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenDetail, err := jwtPkg.GetTokenDetailFromContext(c)
			if err != nil {
				echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("get token detail failed, err:%v", err))
				return err
			}

			// LoginType为空，不需要校验access token
			if tokenDetail.LoginType == "" {
				return next(c)
			}

			if tokenDetail.LoginType != AccessTokenLogin {
				return echo.NewHTTPError(http.StatusUnauthorized, "access token login type is error")
			}

			accessTokenInfo, err := au.userUsecase.repo.GetAccessTokenByUser(c.Request().Context(), tokenDetail.UID)

			if err != nil {
				return err
			}

			if accessTokenInfo.Token != tokenDetail.TokenStr {
				return echo.NewHTTPError(http.StatusUnauthorized, "access token is not latest")
			}

			return next(c)
		}
	}
}
