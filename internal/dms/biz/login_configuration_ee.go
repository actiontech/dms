//go:build enterprise

package biz

import (
	"context"
	"errors"

	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
)

func (d *LoginConfigurationUsecase) UpdateLoginConfiguration(ctx context.Context, LoginButtonText *string, DisableUserPwdLogin *bool) error {
	loginC, err := d.GetLoginConfiguration(ctx)
	if err != nil {
		return err
	}

	// patch login config
	{
		if LoginButtonText != nil {
			loginC.LoginButtonText = *LoginButtonText
		}

		if DisableUserPwdLogin != nil {
			loginC.DisableUserPwdLogin = *DisableUserPwdLogin
		}
	}

	return d.repo.UpdateLoginConfiguration(ctx, loginC)
}

func (d *LoginConfigurationUsecase) GetLoginConfiguration(ctx context.Context) (loginC *LoginConfiguration, err error) {
	ldapC, err := d.repo.GetLastLoginConfiguration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return defaultLoginConfiguration()
		}
		return nil, err
	}
	return ldapC, nil
}
