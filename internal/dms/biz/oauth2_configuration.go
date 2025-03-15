package biz

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	jsoniter "github.com/json-iterator/go"
)

const BackChannelLogoutUri = "/backchannel_logout"

type Oauth2Configuration struct {
	Base

	UID                  string
	EnableOauth2         bool
	SkipCheckState       bool
	AutoCreateUser       bool
	AutoCreateUserPWD    string
	AutoCreateUserSecret string
	ClientID             string
	ClientKey            string
	ClientSecret         string
	ClientHost           string
	ServerAuthUrl        string
	ServerTokenUrl       string
	ServerUserIdUrl      string
	ServerLogoutUrl      string
	Scopes               []string
	AccessTokenTag       string
	UserIdTag            string
	UserEmailTag         string
	UserWeChatTag        string
	LoginPermExpr        string
	LoginTip             string
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
	tx             TransactionGenerator
	repo           Oauth2ConfigurationRepo
	userUsecase    *UserUsecase
	sessionUsecase *SessionUsecase
	log            *utilLog.Helper
}

func NewOauth2ConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo Oauth2ConfigurationRepo, userUsecase *UserUsecase, sessionUsecase *SessionUsecase) *Oauth2ConfigurationUsecase {
	return &Oauth2ConfigurationUsecase{
		tx:             tx,
		repo:           repo,
		userUsecase:    userUsecase,
		sessionUsecase: sessionUsecase,
		log:            utilLog.NewHelper(log, utilLog.WithMessageKey("biz.oauth2_configuration")),
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

type CallbackRedirectData struct {
	UserExist   bool
	DMSToken    string
	Oauth2Token string
	IdToken     string
	Error       string
	uri         string
}

func (c CallbackRedirectData) Generate() string {
	params := url.Values{}
	params.Set("user_exist", strconv.FormatBool(c.UserExist))
	if c.DMSToken != "" {
		params.Set("dms_token", c.DMSToken)
	}
	if c.Oauth2Token != "" {
		params.Set("oauth2_token", c.Oauth2Token)
	}
	if c.IdToken != "" {
		params.Set("id_token", c.IdToken)
	}
	if c.Error != "" {
		params.Set("error", c.Error)
	}
	return fmt.Sprintf("%v/user/bind?%v", c.uri, params.Encode())
}

type ClaimsInfo struct {
	UserId string
	Iat    float64
	Exp    float64
	Sub    string
	Sid    string
}

func (c ClaimsInfo) DmsToken() (token string, expDura time.Duration, err error) {
	expDura = time.Duration((c.Exp-c.Iat)*0.9) * time.Second
	token, err = jwt.GenJwtToken(jwt.WithUserId(c.UserId), jwt.WithExpiredTime(expDura), jwt.WithSub(c.Sub), jwt.WithSid(c.Sid))
	return
}

func (c ClaimsInfo) DmsRefreshToken() (token string, expDura time.Duration, err error) {
	expDura = time.Duration(c.Exp-c.Iat) * time.Second
	token, err = jwt.GenRefreshToken(jwt.WithUserId(c.UserId), jwt.WithExpiredTime(expDura), jwt.WithSub(c.Sub), jwt.WithSid(c.Sid))
	return
}
