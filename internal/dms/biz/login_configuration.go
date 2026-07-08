package biz

import (
	"context"
	"time"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/patrickmn/go-cache"
)

type LoginConfiguration struct {
	Base
	UID                   string
	LoginButtonText       string
	DisableUserPwdLogin   bool
	DisableMultipleLogin  bool
}

func defaultLoginConfiguration() (*LoginConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &LoginConfiguration{
		UID:                  uid,
		LoginButtonText:      "登录",
		DisableUserPwdLogin:  false,
		DisableMultipleLogin: false,
	}, nil
}

type LoginConfigurationRepo interface {
	UpdateLoginConfiguration(ctx context.Context, configuration *LoginConfiguration) error
	GetLastLoginConfiguration(ctx context.Context) (*LoginConfiguration, error)
}

type LoginConfigurationUsecase struct {
	tx                        TransactionGenerator
	repo                      LoginConfigurationRepo
	log                       *utilLog.Helper
	disableMultipleLoginCache *cache.Cache
}

func NewLoginConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo LoginConfigurationRepo) *LoginConfigurationUsecase {
	return &LoginConfigurationUsecase{
		tx:                        tx,
		repo:                      repo,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.Login_configuration")),
		disableMultipleLoginCache: cache.New(10*time.Second, 30*time.Second),
	}
}

func (d *LoginConfigurationUsecase) IsDisableMultipleLogin(ctx context.Context) (bool, error) {
	if cached, found := d.disableMultipleLoginCache.Get("disable_multiple_login"); found {
		if val, ok := cached.(bool); ok {
			return val, nil
		}
	}
	loginC, err := d.GetLoginConfiguration(ctx)
	if err != nil {
		return false, err
	}
	d.disableMultipleLoginCache.Set("disable_multiple_login", loginC.DisableMultipleLogin, cache.DefaultExpiration)
	return loginC.DisableMultipleLogin, nil
}

func (d *LoginConfigurationUsecase) invalidateDisableMultipleLoginCache() {
	d.disableMultipleLoginCache.Delete("disable_multiple_login")
}
