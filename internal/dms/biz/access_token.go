package biz

import (
	"fmt"
	"net/http"

	jwtPkg "github.com/actiontech/dms/pkg/dms-common/api/jwt"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/golang-jwt/jwt/v4"
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
			user := c.Get("user")
			// 获取token为空，代表该请求不需要校验token，例如：/v1/dms/oauth2
			if user == nil {
				return next(c)
			}
			token, ok := user.(*jwt.Token)
			if !ok {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to convert user from jwt token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to convert token claims to jwt")
			}

			// 如果不存在JWTLoginType字段，代表是账号密码登录获取的token或者是扫描任务的凭证，不进行校验
			loginType, ok := claims[jwtPkg.JWTLoginType]
			if !ok {
				return next(c)
			}
			if loginType != AccessTokenLogin {
				return echo.NewHTTPError(http.StatusUnauthorized, "access token login type is error")
			}
			uidStr := fmt.Sprintf("%v", claims[jwtPkg.JWTUserId])
			accessTokenInfo, err := au.userUsecase.repo.GetAccessTokenByUser(c.Request().Context(), uidStr)
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
