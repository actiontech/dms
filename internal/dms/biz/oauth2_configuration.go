package biz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	jsoniter "github.com/json-iterator/go"
)

type Oauth2Configuration struct {
	Base

	UID             string
	EnableOauth2    bool
	SkipCheckState  bool
	AutoCreateUser  bool
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
	UserEmailTag    string
	UserWeChatTag   string
	LoginTip        string
}

func initOauth2Configuration() (*Oauth2Configuration, error) { //nolint
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

// the path should formate like path.to.parameter or path.to.slice.0.parameter
func ParseJsonByPath(jsonBytes []byte, jsonPath string) (jsoniter.Any, error) {
	pathSlice := strings.Split(jsonPath, ".")
	if len(pathSlice) == 0 {
		return nil, fmt.Errorf("empty json path")
	}
	var jsonObject jsoniter.Any = jsoniter.Get(jsonBytes)
	for _, path := range pathSlice {
		if index, err := strconv.Atoi(path); err == nil {
			jsonObject = jsonObject.Get(index)
		} else {
			jsonObject = jsonObject.Get(path)
		}
		if jsonObject.LastError() != nil {
			return nil, jsonObject.LastError()
		}
	}
	return jsonObject, nil
}
