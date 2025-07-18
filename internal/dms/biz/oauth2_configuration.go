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
	EnableManuallyBind   bool
	AutoBindSameNameUser bool
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
	tx                   TransactionGenerator
	repo                 Oauth2ConfigurationRepo
	userUsecase          *UserUsecase
	oauth2SessionUsecase *OAuth2SessionUsecase
	log                  *utilLog.Helper
}

func NewOauth2ConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo Oauth2ConfigurationRepo, userUsecase *UserUsecase, oauth2SessionUsecase *OAuth2SessionUsecase) *Oauth2ConfigurationUsecase {
	return &Oauth2ConfigurationUsecase{
		tx:                   tx,
		repo:                 repo,
		userUsecase:          userUsecase,
		oauth2SessionUsecase: oauth2SessionUsecase,
		log:                  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.oauth2_configuration")),
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
	UserExist    bool
	DMSToken     string
	Oauth2Token  string
	RefreshToken string
	Error        string
	State        string
	uri          string
}

var oauthRedirectPrefixState = "target="

func (c CallbackRedirectData) Generate() string {
	redirectUrl := fmt.Sprintf("%v/user/bind", c.uri)
	params := url.Values{}
	params.Set("user_exist", strconv.FormatBool(c.UserExist))
	if c.DMSToken != "" {
		params.Set("dms_token", c.DMSToken)
	}
	if c.Oauth2Token != "" {
		params.Set("oauth2_token", c.Oauth2Token)
	}
	if c.RefreshToken != "" {
		params.Set("refresh_token", c.RefreshToken)
	}
	if c.Error != "" {
		params.Set("error", c.Error)
	}
	if val := strings.Split(c.State, oauthRedirectPrefixState); len(val) > 1 {
		params.Set("target", val[1])
	}
	return fmt.Sprintf("%v?%v", redirectUrl, params.Encode())
}

type ClaimsInfo struct {
	UserId     string  `json:"user_id"`     // dms用户ID
	Iat        float64 `json:"iat"`         // 第三方AccessToken 签发时间 (Issued At)，Unix 时间戳
	Exp        float64 `json:"exp"`         // 第三方AccessToken 过期时间 (Expiration Time)，Unix 时间戳
	Sub        string  `json:"sub"`         // 第三方AccessToken 主题 (Subject)，通常是用户ID或唯一标识符
	Sid        string  `json:"sid"`         // 第三方AccessToken 会话ID (Session ID)，用于跟踪用户会话
	RefreshIat float64 `json:"refresh_iat"` // 第三方RefreshToken 签发时间 (Issued At)，Unix 时间戳
	RefreshExp float64 `json:"refresh_exp"` // 第三方RefreshToken 过期时间 (Expiration Time)，Unix 时间戳
}

func (c *ClaimsInfo) DmsToken() (token string, cookieExp time.Duration, err error) {
	c.setDefaults()
	// 为了在第三方会话“快过期”时去刷新第三方token，故此时（通过OAuth2登录）签发的DmsToken有效期为第三方平台的0.9
	cookieExp = time.Duration((c.Exp-c.Iat)*0.9) * time.Second
	token, err = jwt.GenJwtToken(jwt.WithUserId(c.UserId), jwt.WithExpiredTime(cookieExp), jwt.WithSub(c.Sub), jwt.WithSid(c.Sid))
	return
}

func (c *ClaimsInfo) DmsRefreshToken() (token string, cookieExp time.Duration, err error) {
	c.setDefaults()
	// cookie有效期更久，和第三方refresh token有效期保持一致
	// 这样在DmsRefreshToken过期时，cookie仍可获取，用于注销第三方会话
	cookieExp = time.Duration(c.RefreshExp-c.RefreshIat) * time.Second
	token, err = jwt.GenRefreshToken(jwt.WithUserId(c.UserId), jwt.WithExpiredTime(time.Duration(c.Exp-c.Iat)*time.Second), jwt.WithSub(c.Sub), jwt.WithSid(c.Sid))
	return
}

func (c *ClaimsInfo) setDefaults() {
	now := time.Now()

	if c.Iat == 0 {
		c.Iat = float64(now.Unix())
	}
	if c.Exp == 0 {
		c.Exp = float64(now.Add(jwt.DefaultDmsTokenExpHours * time.Hour).Unix())
	}
	if c.RefreshIat == 0 {
		c.RefreshIat = c.Iat
	}
	if c.RefreshExp == 0 {
		c.RefreshExp = c.Exp
	}

	return
}
