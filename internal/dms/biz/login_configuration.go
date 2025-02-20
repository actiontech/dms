package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type LoginConfiguration struct {
	Base
	UID                 string
	LoginButtonText     string
	DisableUserPwdLogin bool
}

func defaultLoginConfiguration() (*LoginConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &LoginConfiguration{
		UID:                 uid,
		LoginButtonText:     "登录",
		DisableUserPwdLogin: false,
	}, nil
}

type LoginConfigurationRepo interface {
	UpdateLoginConfiguration(ctx context.Context, configuration *LoginConfiguration) error
	GetLastLoginConfiguration(ctx context.Context) (*LoginConfiguration, error)
}

type LoginConfigurationUsecase struct {
	tx   TransactionGenerator
	repo LoginConfigurationRepo
	log  *utilLog.Helper
}

func NewLoginConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo LoginConfigurationRepo) *LoginConfigurationUsecase {
	return &LoginConfigurationUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.Login_configuration")),
	}
}
