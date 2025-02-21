//go:build !enterprise

package biz

import (
	"context"
	"errors"
)

var errNotSupportLoginConfiguration = errors.New("login configuration related functions are enterprise version functions")

func (d *LoginConfigurationUsecase) UpdateLoginConfiguration(ctx context.Context, LoginButtonText *string, DisableUserPwdLogin *bool) error {
	return errNotSupportLoginConfiguration
}

func (d *LoginConfigurationUsecase) GetLoginConfiguration(ctx context.Context) (loginC *LoginConfiguration, err error) {
	return &LoginConfiguration{
		LoginButtonText:     "登录",
		DisableUserPwdLogin: false,
	}, nil
}
