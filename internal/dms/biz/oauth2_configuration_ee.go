//go:build enterprise

package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	jwtpkg "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
)

func (d *Oauth2ConfigurationUsecase) UpdateOauth2Configuration(ctx context.Context, conf v1.Oauth2Configuration) error {
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
		if conf.EnableOauth2 != nil {
			oauth2C.EnableOauth2 = *conf.EnableOauth2
		}
		if conf.SkipCheckState != nil {
			oauth2C.SkipCheckState = *conf.SkipCheckState
		}
		if conf.AutoCreateUser != nil {
			oauth2C.AutoCreateUser = *conf.AutoCreateUser
		}
		if conf.AutoCreateUserPWD != nil {
			oauth2C.AutoCreateUserPWD = *conf.AutoCreateUserPWD
		}
		if conf.ClientID != nil {
			oauth2C.ClientID = *conf.ClientID
		}
		if conf.ClientKey != nil {
			oauth2C.ClientKey = *conf.ClientKey
		}
		if conf.ClientHost != nil {
			oauth2C.ClientHost = *conf.ClientHost
		}
		if conf.ServerAuthUrl != nil {
			oauth2C.ServerAuthUrl = *conf.ServerAuthUrl
		}
		if conf.ServerTokenUrl != nil {
			oauth2C.ServerTokenUrl = *conf.ServerTokenUrl
		}
		if conf.ServerUserIdUrl != nil {
			oauth2C.ServerUserIdUrl = *conf.ServerUserIdUrl
		}
		if conf.ServerLogoutUrl != nil {
			oauth2C.ServerLogoutUrl = *conf.ServerLogoutUrl
		}
		if conf.Scopes != nil {
			oauth2C.Scopes = *conf.Scopes
		}
		if conf.AccessTokenTag != nil {
			oauth2C.AccessTokenTag = *conf.AccessTokenTag
		}
		if conf.UserIdTag != nil {
			oauth2C.UserIdTag = *conf.UserIdTag
		}
		if conf.LoginTip != nil {
			oauth2C.LoginTip = *conf.LoginTip
		}
		if conf.UserEmailTag != nil {
			oauth2C.UserEmailTag = *conf.UserEmailTag
		}
		if conf.UserWeChatTag != nil {
			oauth2C.UserWeChatTag = *conf.UserWeChatTag
		}
		if conf.LoginPermExpr != nil {
			oauth2C.LoginPermExpr = *conf.LoginPermExpr
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
func (d *Oauth2ConfigurationUsecase) GenerateCallbackUri(ctx context.Context, state, code string) (data *CallbackRedirectData, claims *ClaimsInfo, err error) {
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return nil, nil, err
	}
	// TODO sqle https should also support
	data = &CallbackRedirectData{uri: oauth2C.ClientHost}
	// check callback request
	if !oauth2C.SkipCheckState && state != oauthState {
		err := fmt.Errorf("invalid state: %v", state)
		data.Error = err.Error()
		return data, nil, err
	}
	if code == "" {
		err := fmt.Errorf("code is nil")
		data.Error = err.Error()
		return data, nil, err
	}

	// get oauth2 token
	oauth2Token, err := d.generateOauth2Config(oauth2C).Exchange(ctx, code)
	if err != nil {
		data.Error = err.Error()
		return data, nil, err
	}
	data.Oauth2Token, data.RefreshToken = oauth2Token.AccessToken, oauth2Token.RefreshToken
	idToken, _ := oauth2Token.Extra("id_token").(string) // 尝试获取 id_token
	d.log.Infof("oauth2Token type: %s, got id_token len: %d", oauth2Token.Type(), len(idToken))

	defer func() {
		if oauth2C.ServerLogoutUrl != "" && err != nil {
			// 第三方平台登录成功，但后续dms流程异常，需要注销第三方平台上的会话
			logoutErr := d.backendLogout(ctx, idToken)
			if logoutErr != nil {
				d.log.Errorf("backendLogout error: %v", logoutErr)
				// err 是命名返回值才可以完成实际返回值的修改
				err = fmt.Errorf("%w; Clear OAuth2 session err: %v", err, logoutErr)
			} else {
				err = fmt.Errorf("%w; Cleared OAuth2 session", err)
			}
		}
	}()

	token, err := jwtpkg.Parse(oauth2Token.AccessToken, nil)
	if token == nil {
		return data, nil, fmt.Errorf("parse oauth2 access token failed: %v", err)
	}
	needUpdateState, canLogin, err := d.getLoginPermFromToken(ctx, token)
	if err != nil {
		return data, nil, fmt.Errorf("get login perm from token err: %v", err)
	}

	//get user is exist
	oauth2User, err := d.getOauth2User(oauth2C, oauth2Token.AccessToken)
	if err != nil {
		data.Error = err.Error()
		return data, nil, err
	}
	user, exist, err := d.userUsecase.GetUserByThirdPartyUserID(ctx, oauth2User.UID)
	if err != nil {
		data.Error = err.Error()
		return data, nil, err
	}
	data.UserExist = exist

	if !exist && needUpdateState && !canLogin {
		// 没有登录权限时且用户不存在时，返回登录错误
		err = fmt.Errorf("the user does not have login permission")
		data.Error = err.Error()
		return data, nil, err
	}

	var userID string
	if oauth2C.AutoCreateUser && !exist {
		if oauth2C.AutoCreateUserPWD == "" {
			return data, nil, fmt.Errorf("OAuth2 user default login password is empty")
		}
		args := &CreateUserArgs{
			Name:                   oauth2User.UID,
			Password:               oauth2C.AutoCreateUserPWD,
			IsDisabled:             false,
			ThirdPartyUserID:       oauth2User.UID,
			UserAuthenticationType: UserAuthenticationTypeOAUTH2,
			ThirdPartyUserInfo:     oauth2User.ThirdPartyUserInfo,
			Email:                  oauth2User.Email,
			WxID:                   oauth2User.WxID,
		}
		uid, err := d.userUsecase.CreateUser(ctx, pkgConst.UIDOfUserSys, args)
		if err != nil {
			d.log.Errorf("when generate callback uri, userUsecase.CreateUser failed,%v", err)
			return data, nil, err
		}
		userID = uid
		data.UserExist = true
	} else if exist {
		userID = user.GetUID()
		// update user whenever login via oauth2
		user.UserAuthenticationType = UserAuthenticationTypeOAUTH2
		user.WxID = oauth2User.WxID
		user.Email = oauth2User.Email
		user.ThirdPartyUserInfo = oauth2User.ThirdPartyUserInfo
		if needUpdateState && canLogin {
			user.Stat = UserStatOK
		} else if needUpdateState && !canLogin {
			user.Stat = UserStatDisable
		}
		err := d.userUsecase.SaveUser(ctx, user)
		if err != nil {
			d.log.Errorf("when generate callback uri, update user failed,%v", err)
		}
		if user.Stat == UserStatDisable {
			err = fmt.Errorf("user %s can not login", user.Name)
			data.Error = err.Error()
			return data, nil, err
		}
	}

	sub, sid, exp, iat, err := d.getClaimsInfoFromToken(ctx, token)
	if err != nil {
		return data, nil, fmt.Errorf("get claims info from oauth2 access token err: %v", err)
	}
	refExp, refIat := exp, iat
	if oauth2Token.RefreshToken != "" {
		refreshToken, err := jwtpkg.Parse(oauth2Token.RefreshToken, nil)
		if refreshToken == nil {
			return data, nil, fmt.Errorf("parse oauth2 refresh token failed: %v", err)
		}
		_, _, refExp, refIat, err = d.getClaimsInfoFromToken(ctx, refreshToken)
		if err != nil {
			return data, nil, fmt.Errorf("get refresh token claims failed: %v", err)
		}
	}
	_, err = d.oauth2SessionUsecase.CreateOrUpdateSession(ctx, userID, sub, sid, idToken, oauth2Token.RefreshToken, time.Now().Add(time.Duration(refExp-refIat)*time.Second*2))
	if err != nil {
		return data, nil, fmt.Errorf("create or update oauth2 session failed:%v", err)
	}

	claims = &ClaimsInfo{UserId: userID, Iat: iat, Exp: exp, Sub: sub, Sid: sid, RefreshIat: refIat, RefreshExp: refExp}

	return data, claims, nil
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

func (d *Oauth2ConfigurationUsecase) getLoginPermFromToken(ctx context.Context, token *jwtpkg.Token) (configured, canLogin bool, err error) {
	Oauth2Conf, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return false, false, err
	}
	if Oauth2Conf.LoginPermExpr == "" {
		return false, false, nil
	}

	claims, err := json.Marshal(token.Claims)
	if err != nil {
		return false, false, err
	}
	// claims eg:
	//
	// {
	//  "exp": 1741593656,
	//  "iat": 1741593356,
	//  "auth_time": 1741593356,
	//  "jti": "a3ce017a-f00b-42b2-aac3-b54ebfc86a86",
	//  "iss": "http://localhost:8080/realms/test",
	//  "aud": "account",
	//  "sub": "1cda12c9-ca15-40b5-af45-1f4e73df4600",
	//  "typ": "Bearer",
	//  "azp": "sqle",
	//  "sid": "036bf71b-6696-4fb4-b9fb-dadb924991ba",
	//  "acr": "1",
	//  "allowed-origins": [
	//    "*"
	//  ],
	//  "realm_access": {
	//    "roles": [
	//      "default-roles-test",
	//      "offline_access",
	//      "uma_authorization"
	//    ]
	//  },
	//  "resource_access": {
	//    "sqle": {
	//      "roles": [
	//        "login"
	//      ]
	//    },
	//    "account": {
	//      "roles": [
	//        "manage-account",
	//        "manage-account-links",
	//        "view-profile"
	//      ]
	//    }
	//  },
	//  "scope": "openid email profile",
	//  "email_verified": false,
	//  "name": "l wq",
	//  "preferred_username": "sqle",
	//  "given_name": "l",
	//  "family_name": "wq",
	//  "email": "sqle@xxx.com"
	//}

	// LoginPermExpr eg：`resource_access.sqle.roles.#(=="login")`
	// 即 判断查询的json文档的 resource_access.sqle.roles 中是否存在login元素
	canLogin = gjson.Get(string(claims), Oauth2Conf.LoginPermExpr).Exists()

	return true, canLogin, nil
}

func (d *Oauth2ConfigurationUsecase) getClaimsInfoFromToken(ctx context.Context, token *jwtpkg.Token) (sub, sid string, exp, iat float64, err error) {
	claims, ok := token.Claims.(jwtpkg.MapClaims)
	if !ok {
		return sub, sid, exp, iat, fmt.Errorf("unexpected claims type")
	}

	sub = getClaim[string](claims, "sub")
	sid = getClaim[string](claims, "sid")
	exp = getClaim[float64](claims, "exp")
	iat = getClaim[float64](claims, "iat")

	if exp == 0 || iat == 0 {
		return sub, sid, exp, iat, fmt.Errorf("invalid exp or iat")
	}

	return
}

func getClaim[T any](claims map[string]interface{}, key string) T {
	var zero T
	value, ok := claims[key]
	if !ok {
		return zero
	}

	// 检查值的类型是否与泛型类型匹配
	typedValue, ok := value.(T)
	if !ok {
		return zero
	}

	return typedValue
}

func (d *Oauth2ConfigurationUsecase) BindOauth2User(ctx context.Context, oauth2Token, refreshTokenStr, userName, password string) (claims *ClaimsInfo, err error) {

	// 获取oauth2 配置
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return nil, err
	}
	// 读取Oauth2Token中的 第三方用户ID
	oauth2User, err := d.getOauth2User(oauth2C, oauth2Token)
	if err != nil {
		return nil, fmt.Errorf("get third part user id from oauth2 token failed:%v", err)
	}

	// check third-party users have bound dms user
	_, exist, err := d.userUsecase.GetUserByThirdPartyUserID(ctx, oauth2User.UID)
	if err != nil {
		return nil, fmt.Errorf("get user by third party user id failed : %v", err)
	}
	if exist {
		return nil, pkgErr.ErrBeenBoundOrThePasswordIsWrong
	}

	user, exist, err := d.userUsecase.GetUserByName(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("get user by name failed: %v", err)
	}

	accessToken, err := jwtpkg.Parse(oauth2Token, nil)
	if accessToken == nil {
		return nil, fmt.Errorf("parse oauth2 access token failed: %v", err)
	}
	sub, sid, exp, iat, err := d.getClaimsInfoFromToken(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("getClaimsInfoFromToken failed: %v", err)
	}
	refExp, refIat := exp, iat
	if refreshTokenStr != "" {
		refreshToken, err := jwtpkg.Parse(refreshTokenStr, nil)
		if refreshToken == nil {
			return nil, fmt.Errorf("parse oauth2 refresh token failed: %v", err)
		}
		_, _, refExp, refIat, err = d.getClaimsInfoFromToken(ctx, refreshToken)
		if err != nil {
			return nil, fmt.Errorf("get refresh token claims failed: %v", err)
		}
	}
	claims = &ClaimsInfo{UserId: "", Iat: iat, Exp: exp, Sub: sub, Sid: sid, RefreshIat: refIat, RefreshExp: refExp}

	// create user if not exist
	if !exist {
		args := &CreateUserArgs{
			Name:                   userName,
			Password:               password,
			IsDisabled:             false,
			ThirdPartyUserID:       oauth2User.UID,
			UserAuthenticationType: UserAuthenticationTypeOAUTH2,
			ThirdPartyUserInfo:     oauth2User.ThirdPartyUserInfo,
			Email:                  oauth2User.Email,
			WxID:                   oauth2User.WxID,
		}
		uid, err := d.userUsecase.CreateUser(ctx, pkgConst.UIDOfUserSys, args)
		if err != nil {
			return nil, err
		}
		claims.UserId = uid
	} else {
		// check user state
		if user.Stat == UserStatDisable {
			return nil, fmt.Errorf("user %s not exist or can not login", userName)
		}
		// check password
		if user.Password != password {
			return nil, pkgErr.ErrBeenBoundOrThePasswordIsWrong
		}

		// check user login type
		if user.UserAuthenticationType != UserAuthenticationTypeOAUTH2 &&
			user.UserAuthenticationType != UserAuthenticationTypeDMS &&
			user.UserAuthenticationType != "" {
			return nil, fmt.Errorf("the user has bound other login methods")
		}

		// check user bind third party users
		if user.ThirdPartyUserID != oauth2User.UID && user.ThirdPartyUserID != "" {
			return nil, fmt.Errorf("the user has bound other third-party user")
		}

		// modify user login type
		if user.UserAuthenticationType != UserAuthenticationTypeOAUTH2 {
			user.ThirdPartyUserID = oauth2User.UID
			user.UserAuthenticationType = UserAuthenticationTypeOAUTH2
			user.WxID = oauth2User.WxID
			user.Email = oauth2User.Email
			user.ThirdPartyUserInfo = oauth2User.ThirdPartyUserInfo
			err := d.userUsecase.SaveUser(ctx, user)
			if err != nil {
				return nil, err
			}
		}
		claims.UserId = user.GetUID()
	}

	if err = d.oauth2SessionUsecase.UpdateUserIdBySub(ctx, claims.UserId, sub); err != nil {
		return nil, err
	}

	return claims, nil
}

const (
	userVariableIdToken = "${id_token}"
	userVariableSqleUrl = "${sqle_url}"
)

func (d *Oauth2ConfigurationUsecase) Logout(ctx context.Context, sub, sid string) (string, error) {
	configuration, exist, err := d.GetOauth2Configuration(ctx)
	if err != nil {
		return "", err
	}

	// 获取会话信息
	session, sessionExist, err := d.oauth2SessionUsecase.GetSessionBySubSid(ctx, sub, sid)
	if err != nil {
		return "", fmt.Errorf("failed to get session by sub:%s sid:%s err: %v", sub, sid, err)
	}
	if !exist || !sessionExist || configuration.ServerLogoutUrl == "" || session.IdToken == "" || session.LastLogoutEvent != "" {
		// 无需注销第三方平台
		return "", nil
	}

	token, _ := jwtpkg.Parse(session.IdToken, nil)
	if token == nil || token.Claims == nil {
		d.log.Warnf("failed to Parse idToken, sub:%s sid:%s", session.Sub, session.Sid)
		return "", nil
	}
	claims, ok := token.Claims.(jwtpkg.MapClaims)
	if !ok {
		d.log.Warnf("idToken has invalid Claims, sub:%s sid:%s", session.Sub, session.Sid)
		return "", nil
	}
	if err = claims.Valid(); err != nil {
		// IdToken 已过期，无需注销第三方平台
		d.log.Infof("IdToken should have expired, sub:%s sid:%s", session.Sub, session.Sid)
		return "", nil
	}

	// 配置注销地址时，可以使用这里的键作变量
	vars := map[string]string{
		userVariableIdToken: url.PathEscape(session.IdToken),
		userVariableSqleUrl: url.PathEscape(configuration.ClientHost),
	}
	logoutUrl := configuration.ServerLogoutUrl
	for k, v := range vars {
		logoutUrl = strings.ReplaceAll(logoutUrl, k, v)
	}
	return logoutUrl, nil
}

func (d *Oauth2ConfigurationUsecase) BackendLogout(ctx context.Context, sub, sid string) error {
	if sub == "" && sid == "" {
		return nil
	}

	configuration, exist, err := d.GetOauth2Configuration(ctx)
	if err != nil {
		return fmt.Errorf("GetOauth2Configuration err:%v", err)
	}

	// 获取会话信息
	session, sessionExist, err := d.oauth2SessionUsecase.GetSessionBySubSid(ctx, sub, sid)
	if err != nil {
		return fmt.Errorf("failed to get session by sub:%s sid:%s err: %v", sub, sid, err)
	}
	if !exist || !sessionExist || configuration.ServerLogoutUrl == "" || session.IdToken == "" || session.LastLogoutEvent != "" {
		// 无需注销第三方平台
		return nil
	}

	return d.backendLogout(ctx, session.IdToken)
}

func (d *Oauth2ConfigurationUsecase) backendLogout(ctx context.Context, idToken string) error {
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
	d.log.Infof("backendLogout url: %s", logoutUrlStr)

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

func (d *Oauth2ConfigurationUsecase) RefreshOauth2Token(ctx context.Context, userUid, sub, sid string) (claims *ClaimsInfo, err error) {
	// 获取会话信息
	filterBy := []pkgConst.FilterCondition{
		{Field: "sub", Operator: pkgConst.FilterOperatorEqual, Value: sub},
		{Field: "sid", Operator: pkgConst.FilterOperatorEqual, Value: sid},
		{Field: "user_uid", Operator: pkgConst.FilterOperatorEqual, Value: userUid},
		{Field: "last_logout_event", Operator: pkgConst.FilterOperatorEqual, Value: ""},
	}
	sessions, err := d.oauth2SessionUsecase.GetSessions(ctx, filterBy)
	if err != nil {
		return nil, err
	}
	if len(sessions) != 1 {
		// sub(第三方用户标识)+sid(第三方会话标识)是唯一索引，至多一条记录
		// 不存在则是该会话被注销
		return nil, fmt.Errorf("invalid sessions for user:%s, sub:%s, sid:%s", userUid, sub, sid)
	}

	// 获取OAuth2配置
	oauth2C, err := d.repo.GetLastOauth2Configuration(ctx)
	if err != nil {
		return nil, fmt.Errorf("get oauth2 configuration failed: %v", err)
	}

	// 刷新token
	oauth2Config := d.generateOauth2Config(oauth2C)
	newToken, err := oauth2Config.TokenSource(ctx, &oauth2.Token{RefreshToken: sessions[0].RefreshToken}).Token()
	if err != nil {
		return nil, fmt.Errorf("refresh oauth2 token failed: %v", err)
	}
	idToken, _ := newToken.Extra("id_token").(string) // 尝试获取 id_token

	accessToken, err := jwtpkg.Parse(newToken.AccessToken, nil)
	if accessToken == nil {
		return nil, fmt.Errorf("parse oauth2 access token failed: %v", err)
	}
	// 验证登录权限
	configured, canLogin, err := d.getLoginPermFromToken(ctx, accessToken)
	if configured && !canLogin {
		return nil, fmt.Errorf("the user:%s does not have login permission", sub)
	}

	newSub, newSid, newExp, newIat, err := d.getClaimsInfoFromToken(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("get access token claims failed: %v", err)
	}

	refExp, refIat := newExp, newIat
	if newToken.RefreshToken != "" {
		refreshToken, err := jwtpkg.Parse(newToken.RefreshToken, nil)
		if refreshToken == nil {
			return nil, fmt.Errorf("parse oauth2 refresh token failed: %v", err)
		}
		_, _, refExp, refIat, err = d.getClaimsInfoFromToken(ctx, refreshToken)
		if err != nil {
			return nil, fmt.Errorf("get refresh token claims failed: %v", err)
		}
	}

	claims = &ClaimsInfo{
		UserId:     sessions[0].UserUID,
		Iat:        newIat,
		Exp:        newExp,
		Sub:        newSub,
		Sid:        newSid,
		RefreshIat: refIat,
		RefreshExp: refExp,
	}

	// 更新会话信息
	updateSession := &OAuth2Session{
		Base: Base{
			CreatedAt: sessions[0].CreatedAt,
			UpdatedAt: time.Now(),
		},
		UID:             sessions[0].UID,
		UserUID:         sessions[0].UserUID,
		Sub:             newSub,
		Sid:             newSid,
		IdToken:         idToken,
		RefreshToken:    newToken.RefreshToken,
		LastLogoutEvent: sessions[0].LastLogoutEvent,
		DeleteAfter:     time.Now().Add(time.Duration(refExp-refIat) * time.Second * 2),
	}

	return claims, d.oauth2SessionUsecase.SaveSession(ctx, updateSession)
}

func (d *Oauth2ConfigurationUsecase) BackChannelLogout(ctx context.Context, logoutToken string) (err error) {
	// parse always err: no Keyfunc was provided.
	token, err := jwtpkg.Parse(logoutToken, nil)
	if token == nil {
		return fmt.Errorf("failed to parse logout token: %v", err)
	}

	sub, sid, _, iat, err := d.getClaimsInfoFromToken(ctx, token)
	if err != nil {
		return err
	}

	return d.oauth2SessionUsecase.UpdateLogoutEvent(ctx, sub, sid, fmt.Sprint(int(iat)))
}

func (d *Oauth2ConfigurationUsecase) CheckBackChannelLogoutEvent() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Gets user token from the context.
			dmsToken, ok := c.Get("user").(*jwtpkg.Token)
			if !ok {
				// jwt skipper
				return next(c)
			}
			// 检查dms-token中是否包含OAuth2登录信息：sub(用户标识)、sid(会话标识)
			sub, sid, _, _, _ := d.getClaimsInfoFromToken(c.Request().Context(), dmsToken)
			if sub == "" && sid == "" {
				// 不包含OAuth2信息，继续处理
				return next(c)
			}

			// 包含OAuth2信息，则通过sub和sid查询OAuth2会话表，检查是否被注销
			// 获取会话信息
			filterBy := []pkgConst.FilterCondition{
				{Field: "sub", Operator: pkgConst.FilterOperatorEqual, Value: sub},
				{Field: "sid", Operator: pkgConst.FilterOperatorEqual, Value: sid},
			}
			sessions, err := d.oauth2SessionUsecase.GetSessions(c.Request().Context(), filterBy)
			if err != nil {
				return err
			}

			if len(sessions) == 0 {
				// 会话被清理
				return echo.NewHTTPError(http.StatusUnauthorized, "the session has been logged out by a third-party platform")
			}

			for _, v := range sessions {
				if v.LastLogoutEvent != "" {
					// 会话被注销
					return echo.NewHTTPError(http.StatusUnauthorized, "the session has been logged out by a third-party platform")
				}
			}

			// 会话正常
			return next(c)
		}
	}
}
