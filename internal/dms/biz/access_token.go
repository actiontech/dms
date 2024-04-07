package biz

import (
	"net/http"

	"github.com/actiontech/dms/pkg/dms-common/api/accesstoken"
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
			token, exist, err := accesstoken.GetTokenFromContext(c)
			if err != nil {
				return err
			}
			if !exist {
				return next(c)
			}
			uid, exist, err := accesstoken.GetUidFromAccessToken(token)
			if err != nil {
				return err
			}
			if !exist {
				return next(c)
			}

			accessTokenInfo, err := au.userUsecase.repo.GetAccessTokenByUser(c.Request().Context(), uid)

			if err != nil {
				return err
			}

			if accessTokenInfo.Token != token.Raw {
				return echo.NewHTTPError(http.StatusUnauthorized, "access token is not latest")
			}

			return next(c)
		}
	}
}
