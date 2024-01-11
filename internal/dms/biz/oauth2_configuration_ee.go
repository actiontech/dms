//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	"io"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/oauth2"

	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
)

func (d *Oauth2ConfigurationUsecase) UpdateOauth2Configuration(ctx context.Context, enableOauth2 *bool, clientID, clientKey, clientHost, serverAuthUrl, serverTokenUrl, serverUserIdUrl,
	accessTokenTag, userIdTag, userWechatTag, userEmailTag, loginTip *string, scopes *[]string) error {
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到oauth2配置,默认生成一个带uid的配置
		oauth2C, err = initOauth2Configuration()
		if err != nil {
			return err
		}
	}

	{ // patch oauth2 config
		if enableOauth2 != nil {
			oauth2C.EnableOauth2 = *enableOauth2
		}
		if clientID != nil {
			oauth2C.ClientID = *clientID
		}
		if clientKey != nil {
			oauth2C.ClientKey = *clientKey
		}
		if clientHost != nil {
			oauth2C.ClientHost = *clientHost
		}
		if serverAuthUrl != nil {
			oauth2C.ServerAuthUrl = *serverAuthUrl
		}
		if serverTokenUrl != nil {
			oauth2C.ServerTokenUrl = *serverTokenUrl
		}
		if serverUserIdUrl != nil {
			oauth2C.ServerUserIdUrl = *serverUserIdUrl
		}
		if scopes != nil {
			oauth2C.Scopes = *scopes
		}
		if accessTokenTag != nil {
			oauth2C.AccessTokenTag = *accessTokenTag
		}
		if userIdTag != nil {
			oauth2C.UserIdTag = *userIdTag
		}
		if loginTip != nil {
			oauth2C.LoginTip = *loginTip
		}
		if userEmailTag != nil {
			oauth2C.UserEmailTag = *userEmailTag
		}
		if userWechatTag != nil {
			oauth2C.UserWeChatTag = *userWechatTag
		}
	}
	return d.repo.UpdateOauth2Configuration(ctx, oauth2C)
}

func (d *Oauth2ConfigurationUsecase) GetOauth2Configuration(ctx context.Context) (oauth2C *Oauth2Configuration, exist bool, err error) {
	oauth2C, err = d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return oauth2C, true, nil
}

var oauthState = "dms-action"

func (d *Oauth2ConfigurationUsecase) GenOauth2LinkURI(ctx context.Context) (uri string, err error) {
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return "", err
	}
	config := d.generateOauth2Config(oauth2C)
	uri = config.AuthCodeURL(oauthState)
	_, err = url.ParseRequestURI(uri)
	if err != nil {
		d.log.Errorf("parse oauth2 link failed: %v", err)
		return "", errors.New("parse oauth2 link failed, please check oauth2 config")
	}
	return uri, nil
}
func (d *Oauth2ConfigurationUsecase) generateOauth2Config(oauth2C *Oauth2Configuration) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     oauth2C.ClientID,
		ClientSecret: oauth2C.ClientKey,
		RedirectURL:  fmt.Sprintf("%v/v1/dms/oauth2/callback", oauth2C.ClientHost),
		Scopes:       oauth2C.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauth2C.ServerAuthUrl,
			TokenURL: oauth2C.ServerTokenUrl,
		},
	}
}

// if user is exist and valid, will return dmsToken, otherwise this parameter will be an empty string
func (d *Oauth2ConfigurationUsecase) GenerateCallbackUri(ctx context.Context, state, code string) (redirectUri string, dmsToken string, err error) {
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return "", "", err
	}
	// TODO sqle https should also support
	uri := oauth2C.ClientHost
	data := callbackRedirectData{}
	// check callback request
	if state != oauthState {
		err := fmt.Errorf("invalid state: %v", state)
		data.Error = err.Error()
		return data.generateQuery(uri), "", err
	}
	if code == "" {
		err := fmt.Errorf("code is nil")
		data.Error = err.Error()
		return data.generateQuery(uri), "", err
	}

	// get oauth2 token
	oauth2Token, err := d.generateOauth2Config(oauth2C).Exchange(ctx, code)
	if err != nil {
		data.Error = err.Error()
		return data.generateQuery(uri), "", err
	}
	data.Oauth2Token = oauth2Token.AccessToken

	//get user is exist
	oauth2User, err := d.getOauth2User(oauth2C, oauth2Token.AccessToken)
	if err != nil {
		data.Error = err.Error()
		return data.generateQuery(uri), "", err
	}
	user, exist, err := d.userUsecase.GetUserByThirdPartyUserID(ctx, oauth2User.UID)
	if err != nil {
		data.Error = err.Error()
		return data.generateQuery(uri), "", err
	}
	data.UserExist = exist

	// the user has successfully logged in at the third party, and the token can be returned directly after checking users'state
	if exist {
		if user.Stat == UserStatDisable {
			err = fmt.Errorf("user %s not exist or can not login", user.Name)
			data.Error = err.Error()
			return data.generateQuery(uri), "", err
		}
		dmsToken, err = jwt.GenJwtToken(jwt.WithUserId(user.GetUID()))
		if nil != err {
			return "", "", err
		}
		if err != nil {
			data.Error = err.Error()
			return data.generateQuery(uri), "", err
		}
		data.DMSToken = dmsToken
	}

	return data.generateQuery(uri), dmsToken, nil
}

type callbackRedirectData struct {
	UserExist   bool
	DMSToken    string
	Oauth2Token string
	Error       string
}

func (c callbackRedirectData) generateQuery(uri string) string {
	params := url.Values{}
	params.Set("user_exist", strconv.FormatBool(c.UserExist))
	if c.DMSToken != "" {
		params.Set("dms_token", c.DMSToken)
	}
	if c.Oauth2Token != "" {
		params.Set("oauth2_token", c.Oauth2Token)
	}
	if c.Error != "" {
		params.Set("error", c.Error)
	}
	return fmt.Sprintf("%v/user/bind?%v", uri, params.Encode())
}

func (d *Oauth2ConfigurationUsecase) getOauth2User(conf *Oauth2Configuration, token string) (user *User, err error) {
	uri := fmt.Sprintf("%v?%v=%v", conf.ServerUserIdUrl, conf.AccessTokenTag, token)
	resp, err := (&http.Client{}).Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get third-party user ID, unable to parse response")
	}
	userId, err := ParseJsonByPath(body, conf.UserIdTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get third-party user ID, %v", err)
	}
	if userId.ToString() == "" {
		return nil, fmt.Errorf("not found third-party user ID")
	}
	user = &User{UID: userId.ToString()}
	if conf.UserWeChatTag != "" {
		userWeChat, err := ParseJsonByPath(body, conf.UserWeChatTag)
		if err != nil {
			d.log.Errorf("failed to get third-party wechat, unrecognized response format")
		} else {
			user.WxID = userWeChat.ToString()
		}
	}
	if conf.UserEmailTag != "" {
		userEmail, err := ParseJsonByPath(body, conf.UserEmailTag)
		if err != nil {
			d.log.Errorf("failed to get third-party email, unrecognized response format")
		} else {
			user.Email = userEmail.ToString()
		}
	}
	return user, nil
}

func (d *Oauth2ConfigurationUsecase) BindOauth2User(ctx context.Context, oauth2Token, userName, password string) (token string, err error) {

	// 获取oauth2 配置
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return "", err
	}
	// 读取Oauth2Token中的 第三方用户ID
	oauth2User, err := d.getOauth2User(oauth2C, oauth2Token)
	if err != nil {
		return "", fmt.Errorf("get third part user id from oauth2 token failed:%v", err)
	}

	// check third-party users have bound dms user
	_, exist, err := d.userUsecase.GetUserByThirdPartyUserID(ctx, oauth2User.UID)
	if err != nil {
		return "", fmt.Errorf("get user by third party user id failed : %v", err)
	}
	if exist {
		return "", pkgErr.ErrBeenBoundOrThePasswordIsWrong
	}

	user, exist, err := d.userUsecase.GetUserByName(ctx, userName)
	if err != nil {
		return "", fmt.Errorf("get user by name failed: %v", err)
	}

	// create user if not exist
	if !exist {
		args := &CreateUserArgs{
			Name:                   userName,
			Password:               password,
			IsDisabled:             false,
			ThirdPartyUserID:       oauth2User.UID,
			UserAuthenticationType: UserAuthenticationTypeOAUTH2,
			Email:                  oauth2User.Email,
			WxID:                   oauth2User.WxID,
		}
		uid, err := d.userUsecase.CreateUser(ctx, pkgConst.UIDOfUserSys, args)
		if err != nil {
			return "", err
		}
		return jwt.GenJwtToken(jwt.WithUserId(uid))
	} else {
		// check user state
		if user.Stat == UserStatDisable {
			return "", fmt.Errorf("user %s not exist or can not login", userName)
		}
		// check password
		if user.Password != password {
			return "", pkgErr.ErrBeenBoundOrThePasswordIsWrong
		}

		// check user login type
		if user.UserAuthenticationType != UserAuthenticationTypeOAUTH2 &&
			user.UserAuthenticationType != UserAuthenticationTypeDMS &&
			user.UserAuthenticationType != "" {
			return "", fmt.Errorf("the user has bound other login methods")
		}

		// check user bind third party users
		if user.ThirdPartyUserID != oauth2User.UID && user.ThirdPartyUserID != "" {
			return "", fmt.Errorf("the user has bound other third-party user")
		}

		// modify user login type
		if user.UserAuthenticationType != UserAuthenticationTypeOAUTH2 {
			user.ThirdPartyUserID = oauth2User.UID
			user.UserAuthenticationType = UserAuthenticationTypeOAUTH2
			if user.WxID == "" {
				user.WxID = oauth2User.WxID
			}
			if user.Email == "" {
				user.Email = oauth2User.Email
			}
			err := d.userUsecase.SaveUser(ctx, user)
			if err != nil {
				return "", err
			}
		}
		return jwt.GenJwtToken(jwt.WithUserId(user.UID))
	}
}
