package biz

import (
	"context"

	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type Oauth2Configuration struct {
	Base

	UID             string
	EnableOauth2    bool
	ClientID        string
	ClientKey       string
	ClientSecret    string
	ClientHost      string
	ServerAuthUrl   string
	ServerTokenUrl  string
	ServerUserIdUrl string
	Scopes          []string
	AccessTokenTag  string
	UserIdTag       string
	LoginTip        string
}

func initOauth2Configuration() (*Oauth2Configuration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &Oauth2Configuration{
		UID: uid,
	}, nil
}

type Oauth2ConfigurationRepo interface {
	UpdateOauth2Configuration(ctx context.Context, configuration *Oauth2Configuration) error
	GetLastOauth2Configuration(ctx context.Context) (*Oauth2Configuration, error)
}

type Oauth2ConfigurationUsecase struct {
	tx          TransactionGenerator
	repo        Oauth2ConfigurationRepo
	userUsecase *UserUsecase
	log         *utilLog.Helper
}

func NewOauth2ConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo Oauth2ConfigurationRepo, userUsecase *UserUsecase) *Oauth2ConfigurationUsecase {
	return &Oauth2ConfigurationUsecase{
		tx:          tx,
		repo:        repo,
		userUsecase: userUsecase,
		log:         utilLog.NewHelper(log, utilLog.WithMessageKey("biz.oauth2_configuration")),
	}
}
