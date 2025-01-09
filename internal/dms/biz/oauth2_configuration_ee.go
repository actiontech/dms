//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	jwtpkg "github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
)

func (d *Oauth2ConfigurationUsecase) UpdateOauth2Configuration(ctx context.Context, enableOauth2, skipCheckState, autoCreateUser *bool, autoCreateUserPWD, clientID, clientKey, clientHost, serverAuthUrl, serverTokenUrl, serverUserIdUrl, serverLogoutUrl,
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
		if skipCheckState != nil {
			oauth2C.SkipCheckState = *skipCheckState
		}
		if autoCreateUser != nil {
			oauth2C.AutoCreateUser = *autoCreateUser
		}
		if autoCreateUserPWD != nil {
			oauth2C.AutoCreateUserPWD = *autoCreateUserPWD
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
		if serverLogoutUrl != nil {
			oauth2C.ServerLogoutUrl = *serverLogoutUrl
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

// TODO the state field is a fixed value with low security and should be changed to a random value, such as a hash value based on session ID
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
	if !oauth2C.SkipCheckState && state != oauthState {
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
	data.IdToken, _ = oauth2Token.Extra("id_token").(string) // 尝试获取 id_token
	d.log.Infof("oauth2Token type: %s, got id_token len: %d", oauth2Token.Type(), len(data.IdToken))

	defer func() {
		if oauth2C.ServerLogoutUrl != "" && err != nil {
			// 第三方平台登录成功，但后续dms流程异常，需要注销第三方平台上的会话
			logoutErr := d.BackendLogout(ctx, data.IdToken)
			if logoutErr != nil {
				// err 是命名返回值才可以完成实际返回值的修改
				d.log.Errorf("BackendLogout error: %v", logoutErr)
				err = fmt.Errorf("%w; Clear OAuth2 session err: %v", err, logoutErr)
			} else {
				err = fmt.Errorf("%w; Cleared OAuth2 session", err)
			}
		}
	}()

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
	if oauth2C.AutoCreateUser && !exist {
		if oauth2C.AutoCreateUserPWD == "" {
			return data.generateQuery(uri), "", fmt.Errorf("OAuth2 user default login password is empty")
		}
		args := &CreateUserArgs{
			Name:                   oauth2User.UID,
			Password:               oauth2C.AutoCreateUserPWD,
			IsDisabled:             false,
			ThirdPartyUserID:       oauth2User.UID,
			UserAuthenticationType: UserAuthenticationTypeOAUTH2,
			ThirdPartyUserInfo:     oauth2User.ThirdPartyUserInfo,
			ThirdPartyIdToken:      data.IdToken,
			Email:                  oauth2User.Email,
			WxID:                   oauth2User.WxID,
		}
		uid, err := d.userUsecase.CreateUser(ctx, pkgConst.UIDOfUserSys, args)
		if err != nil {
			d.log.Errorf("when generate callback uri, userUsecase.CreateUser failed,%v", err)
			return "", "", err
		}
		dmsToken, err = jwt.GenJwtToken(jwt.WithUserId(uid))
		if err != nil {
			data.Error = err.Error()
			return data.generateQuery(uri), "", err
		}
		data.DMSToken = dmsToken
		data.UserExist = true
	} else if exist {
		// the user has successfully logged in at the third party, and the token can be returned directly after checking users' state
		if user.Stat == UserStatDisable {
			err = fmt.Errorf("user %s can not login", user.Name)
			data.Error = err.Error()
			return data.generateQuery(uri), "", err
		}
		dmsToken, err = jwt.GenJwtToken(jwt.WithUserId(user.GetUID()))
		if err != nil {
			data.Error = err.Error()
			return data.generateQuery(uri), "", err
		}
		data.DMSToken = dmsToken
		// update user whenever login via oauth2
		user.UserAuthenticationType = UserAuthenticationTypeOAUTH2
		user.WxID = oauth2User.WxID
		user.Email = oauth2User.Email
		user.ThirdPartyUserInfo = oauth2User.ThirdPartyUserInfo
		user.ThirdPartyIdToken = data.IdToken
		err := d.userUsecase.SaveUser(ctx, user)
		if err != nil {
			d.log.Errorf("when generate callback uri, update user failed,%v", err)
		}
	}

	return data.generateQuery(uri), dmsToken, nil
}

type callbackRedirectData struct {
	UserExist   bool
	DMSToken    string
	Oauth2Token string
	IdToken     string
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
	if c.IdToken != "" {
		params.Set("id_token", c.IdToken)
	}
	if c.Error != "" {
		params.Set("error", c.Error)
	}
	return fmt.Sprintf("%v/user/bind?%v", uri, params.Encode())
}

func (d *Oauth2ConfigurationUsecase) getOauth2User(conf *Oauth2Configuration, token string) (user *User, err error) {
	oauth2Config := d.generateOauth2Config(conf)
	client := oauth2Config.Client(context.Background(), &oauth2.Token{AccessToken: token})

	// 兼容原有将token放到uri的情况
	uri := fmt.Sprintf("%v?%v=%v", conf.ServerUserIdUrl, conf.AccessTokenTag, token)
	resp, err := client.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo, err: %v", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get third-party user ID, unable to read response")
	}
	userId, err := ParseJsonByPath(body, conf.UserIdTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get third-party user ID, %v, resp: %s", err, body)
	}
	if userId.ToString() == "" {
		return nil, fmt.Errorf("not found third-party user ID, resp: %s", body)
	}
	user = &User{UID: userId.ToString()}
	if conf.UserWeChatTag != "" {
		userWeChat, err := ParseJsonByPath(body, conf.UserWeChatTag)
		if err != nil {
			d.log.Errorf("failed to get third-party wechat, resp: %s", body)
		} else {
			user.WxID = userWeChat.ToString()
		}
	}
	if conf.UserEmailTag != "" {
		userEmail, err := ParseJsonByPath(body, conf.UserEmailTag)
		if err != nil {
			d.log.Errorf("failed to get third-party email, resp: %s", body)
		} else {
			user.Email = userEmail.ToString()
		}
	}
	user.ThirdPartyUserInfo = string(body)
	return user, nil
}

func (d *Oauth2ConfigurationUsecase) BindOauth2User(ctx context.Context, oauth2Token, idToken, userName, password string) (token string, err error) {

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
			ThirdPartyUserInfo:     oauth2User.ThirdPartyUserInfo,
			ThirdPartyIdToken:      idToken,
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
			user.WxID = oauth2User.WxID
			user.Email = oauth2User.Email
			user.ThirdPartyUserInfo = oauth2User.ThirdPartyUserInfo
			user.ThirdPartyIdToken = idToken
			err := d.userUsecase.SaveUser(ctx, user)
			if err != nil {
				return "", err
			}
		}
		return jwt.GenJwtToken(jwt.WithUserId(user.UID))
	}
}

const (
	userVariableIdToken = "${id_token}"
	userVariableSqleUrl = "${sqle_url}"
)

func (d *Oauth2ConfigurationUsecase) Logout(ctx context.Context, uid string) (string, error) {
	user, err := d.userUsecase.GetUser(ctx, uid)
	if err != nil {
		return "", err
	}
	configuration, exist, err := d.GetOauth2Configuration(ctx)
	if err != nil {
		return "", err
	}
	if !exist || user.ThirdPartyIdToken == "" || configuration.ServerLogoutUrl == "" {
		// 无需注销第三方平台
		return "", nil
	}
	token, _ := jwtpkg.Parse(user.ThirdPartyIdToken, nil)
	if token == nil || token.Claims == nil {
		d.log.Warnf("failed to Parse ThirdPartyIdToken of user uid: %s", user.UID)
		return "", nil
	}
	claims, ok := token.Claims.(jwtpkg.MapClaims)
	if !ok {
		d.log.Warnf("ThirdPartyIdToken of user uid:%s has invalid Claims", user.UID)
		return "", nil
	}
	if err = claims.Valid(); err != nil {
		// ThirdPartyIdToken 已过期，无需注销第三方平台
		d.log.Infof("ThirdPartyIdToken of user uid:%s should have expired, %v", user.UID, err)
		return "", nil
	}

	// 配置注销地址时，可以使用这里的键作变量
	vars := map[string]string{
		userVariableIdToken: url.PathEscape(user.ThirdPartyIdToken),
		userVariableSqleUrl: url.PathEscape(configuration.ClientHost),
	}
	logoutUrl := configuration.ServerLogoutUrl
	for k, v := range vars {
		logoutUrl = strings.ReplaceAll(logoutUrl, k, v)
	}
	return logoutUrl, nil
}

func (d *Oauth2ConfigurationUsecase) BackendLogout(ctx context.Context, idToken string) error {
	configuration, exist, err := d.GetOauth2Configuration(ctx)
	if err != nil {
		return fmt.Errorf("get oauth2 configuration failed: %v", err)
	}
	if !exist {
		return fmt.Errorf("Oauth2Configuration is not exist")
	}

	logoutUrl, err := url.Parse(configuration.ServerLogoutUrl)
	if err != nil {
		return fmt.Errorf("parse logout url failed: %v", err)
	}

	query := logoutUrl.Query()
	for key := range query {
		val := query.Get(key)
		if val == userVariableIdToken {
			query.Set(key, idToken)
		} else if val == userVariableSqleUrl {
			query.Del(key)
		}
	}
	logoutUrl.RawQuery = query.Encode()
	logoutUrlStr := logoutUrl.String()
	d.log.Infof("BackendLogout url: %s", logoutUrlStr)

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Get(logoutUrlStr)
	if err != nil {
		return fmt.Errorf("request logout url failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request logout url resp Status: %v", resp.Status)
	}
	return nil
}
