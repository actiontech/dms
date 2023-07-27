package biz

import (
	"context"
	"errors"

	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type LDAPConfiguration struct {
	Base

	UID             string
	Enable          bool
	EnableSSL       bool
	Host            string
	Port            string
	ConnectDn       string
	ConnectPassword string
	BaseDn          string
	UserNameRdnKey  string
	UserEmailRdnKey string
}

func initLDAPConfiguration() (*LDAPConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &LDAPConfiguration{
		UID: uid,
	}, nil
}

type LDAPConfigurationRepo interface {
	UpdateLDAPConfiguration(ctx context.Context, configuration *LDAPConfiguration) error
	GetLastLDAPConfiguration(ctx context.Context) (*LDAPConfiguration, error)
}

type LDAPConfigurationUsecase struct {
	tx   TransactionGenerator
	repo LDAPConfigurationRepo
	log  *utilLog.Helper
}

func NewLDAPConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo LDAPConfigurationRepo) *LDAPConfigurationUsecase {
	return &LDAPConfigurationUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.ldap_configuration")),
	}
}

func (d *LDAPConfigurationUsecase) UpdateLDAPConfiguration(ctx context.Context, enableLdap, enableSSL *bool, ldapServerHost, ldapServerPort, ldapConnectDn, ldapConnectPwd, ldapSearchBaseDn,
	ldapUserNameRdnKey, ldapUserEmailRdnKey *string) error {
	ldapC, err := d.repo.GetLastLDAPConfiguration(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到ldap配置,默认生成一个带uid的配置
		ldapC, err = initLDAPConfiguration()
		if err != nil {
			return err
		}
	}

	// patch ldap config
	{

		if enableLdap != nil {
			ldapC.Enable = *enableLdap
		}

		if enableSSL != nil {
			ldapC.EnableSSL = *enableSSL
		}

		if ldapServerHost != nil {
			ldapC.Host = *ldapServerHost
		}

		if ldapServerPort != nil {
			ldapC.Port = *ldapServerPort
		}

		if ldapConnectDn != nil {
			ldapC.ConnectDn = *ldapConnectDn
		}

		if ldapConnectPwd != nil {
			ldapC.ConnectPassword = *ldapConnectPwd
		}

		if ldapSearchBaseDn != nil {
			ldapC.BaseDn = *ldapSearchBaseDn
		}

		if ldapUserNameRdnKey != nil {
			ldapC.UserNameRdnKey = *ldapUserNameRdnKey
		}

		if ldapUserEmailRdnKey != nil {
			ldapC.UserEmailRdnKey = *ldapUserEmailRdnKey
		}

	}
	return d.repo.UpdateLDAPConfiguration(ctx, ldapC)
}

func (d *LDAPConfigurationUsecase) GetLDAPConfiguration(ctx context.Context) (ldapc *LDAPConfiguration, exist bool, err error) {
	ldapC, err := d.repo.GetLastLDAPConfiguration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return ldapC, true, nil
}
