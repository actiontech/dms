//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportOauth2 = errors.New("oauth2 related functions are enterprise version functions")

func (d *Oauth2ConfigurationUsecase) UpdateOauth2Configuration(ctx context.Context, enableOauth2, skipCheckState, autoCreateUser *bool, autoCreateUserPWD, clientID, clientKey, clientHost, serverAuthUrl, serverTokenUrl, serverUserIdUrl, serverLogoutUrl,
	accessTokenTag, userIdTag, userWechatTag, userEmailTag, loginTip *string, scopes *[]string) error {

	return errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) GetOauth2Configuration(ctx context.Context) (oauth2C *Oauth2Configuration, exist bool, err error) {
	return nil, false, errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) GenOauth2LinkURI(ctx context.Context) (uri string, err error) {
	return "", errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) GenerateCallbackUri(ctx context.Context, state, code string) (string, string, error) {
	return "", "", errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) BindOauth2User(ctx context.Context, oauth2Token, idToken, userName, password string) (token string, err error) {
	return "", errNotSupportOauth2
}

func (d *Oauth2ConfigurationUsecase) Logout(ctx context.Context, uid string) (string, error) {
	return "", nil
}
