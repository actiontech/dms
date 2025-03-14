//go:build !enterprise

package biz

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
)

var errNotSupportOauth2 = errors.New("oauth2 related functions are enterprise version functions")

func (d *Oauth2ConfigurationUsecase) UpdateOauth2Configuration(ctx context.Context, conf v1.Oauth2Configuration) error {

	return errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) GetOauth2Configuration(ctx context.Context) (oauth2C *Oauth2Configuration, exist bool, err error) {
	return nil, false, errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) GenOauth2LinkURI(ctx context.Context) (uri string, err error) {
	return "", errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) GenerateCallbackUri(ctx context.Context, state, code string) (data *CallbackRedirectData, claims *ClaimsInfo, err error) {
	return nil, nil, errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) BindOauth2User(ctx context.Context, oauth2Token, idToken, userName, password string) (claims *ClaimsInfo, err error) {
	return nil, errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) Logout(ctx context.Context, uid string) (string, error) {
	return "", nil
}

func (d *Oauth2ConfigurationUsecase) RefreshOauth2Token(ctx context.Context, userUid, sub, sid string) (claims *ClaimsInfo, err error) {
	return nil, errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) BackChannelLogout(ctx context.Context, logoutToken string) (err error) {
	return errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) CheckBackChannelLogoutEvent() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// do nothing
			return next(c)
		}
	}
}
