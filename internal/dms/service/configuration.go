package service

import (
	"context"
	"encoding/json"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/locale"
	"golang.org/x/text/language"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) GetLoginTips(ctx context.Context) (reply *dmsV1.GetLoginTipsReply, err error) {
	loginConfiguration, err := d.LoginConfigurationUsecase.GetLoginConfiguration(ctx)
	if err != nil {
		return nil, err
	}

	return &dmsV1.GetLoginTipsReply{
		Data: dmsV1.LoginTipsResData{
			LoginButtonText:     loginConfiguration.LoginButtonText,
			DisableUserPwdLogin: loginConfiguration.DisableUserPwdLogin,
		},
	}, nil
}

func (d *DMSService) UpdateLoginConfiguration(ctx context.Context, userId string, req *dmsV1.UpdateLoginConfigurationReq) (err error) {
	d.log.Infof("UpdateLoginConfiguration")
	defer func() {
		d.log.Infof("UpdateLoginConfiguration;error=%v", err)
	}()

	// 权限校验
	if canGlobalOp, err := d.OpPermissionVerifyUsecase.CanOpGlobal(ctx, userId); err != nil {
		return fmt.Errorf("check user op permission failed: %v", err)
	} else if !canGlobalOp {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	loginConfiguration := req.LoginConfiguration
	err = d.LoginConfigurationUsecase.UpdateLoginConfiguration(ctx, loginConfiguration.LoginButtonText, loginConfiguration.DisableUserPwdLogin)
	return
}

func (d *DMSService) GetOauth2Configuration(ctx context.Context) (reply *dmsV1.GetOauth2ConfigurationReply, err error) {
	oauth2C, exist, err := d.Oauth2ConfigurationUsecase.GetOauth2Configuration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetOauth2ConfigurationReply{
			Data: dmsV1.GetOauth2ConfigurationResData{},
		}, nil
	}

	return &dmsV1.GetOauth2ConfigurationReply{
		Data: dmsV1.GetOauth2ConfigurationResData{
			EnableOauth2:         oauth2C.EnableOauth2,
			SkipCheckState:       oauth2C.SkipCheckState,
			AutoCreateUser:       oauth2C.AutoCreateUser,
			ClientID:             oauth2C.ClientID,
			ClientHost:           oauth2C.ClientHost,
			ServerAuthUrl:        oauth2C.ServerAuthUrl,
			ServerTokenUrl:       oauth2C.ServerTokenUrl,
			ServerUserIdUrl:      oauth2C.ServerUserIdUrl,
			ServerLogoutUrl:      oauth2C.ServerLogoutUrl,
			Scopes:               oauth2C.Scopes,
			AccessTokenTag:       oauth2C.AccessTokenTag,
			UserIdTag:            oauth2C.UserIdTag,
			LoginTip:             oauth2C.LoginTip,
			UserEmailTag:         oauth2C.UserEmailTag,
			UserWeChatTag:        oauth2C.UserWeChatTag,
			LoginPermExpr:        oauth2C.LoginPermExpr,
			BackChannelLogoutUri: dmsCommonV1.GroupV1 + "/dms/oauth2" + biz.BackChannelLogoutUri,
		},
	}, nil
}

func (d *DMSService) GetOauth2ConfigurationTip(ctx context.Context) (reply *dmsV1.GetOauth2TipsReply, err error) {
	d.log.Infof("GetOauth2ConfigurationTip")
	defer func() {
		d.log.Infof("GetOauth2ConfigurationTip.reply=%v;error=%v", reply, err)
	}()

	oauth2C, exist, err := d.Oauth2ConfigurationUsecase.GetOauth2Configuration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetOauth2TipsReply{
			Data: dmsV1.GetOauth2TipsResData{},
		}, nil
	}

	return &dmsV1.GetOauth2TipsReply{
		Data: dmsV1.GetOauth2TipsResData{
			EnableOauth2: oauth2C.EnableOauth2,
			LoginTip:     oauth2C.LoginTip,
		},
	}, nil
}

func (d *DMSService) UpdateOauth2Configuration(ctx context.Context, req *dmsV1.Oauth2ConfigurationReq) (err error) {
	d.log.Infof("UpdateOauth2Configuration")
	defer func() {
		d.log.Infof("UpdateOauth2Configuration;error=%v", err)
	}()

	oauth2Configuration := req.Oauth2Configuration
	return d.Oauth2ConfigurationUsecase.UpdateOauth2Configuration(ctx, oauth2Configuration)
}

func (d *DMSService) Oauth2Link(ctx context.Context, target string) (uri string, err error) {
	d.log.Infof("Oauth2Link path: %v", target)
	defer func() {
		d.log.Infof("Oauth2Link;error=%v", err)
	}()

	uri, err = d.Oauth2ConfigurationUsecase.GenOauth2LinkURI(ctx, target)
	if err != nil {
		return "", err
	}
	return uri, nil
}

// if redirect directly to SQLE, will return token, otherwise this parameter will be an empty string
func (d *DMSService) Oauth2Callback(ctx context.Context, req *dmsV1.Oauth2CallbackReq) (data *biz.CallbackRedirectData, claims *biz.ClaimsInfo, err error) {
	d.log.Infof("Oauth2Callback")
	defer func() {
		d.log.Infof("Oauth2Callback;error=%v", err)
	}()

	return d.Oauth2ConfigurationUsecase.GenerateCallbackUri(ctx, req.State, req.Code)
}

func (d *DMSService) BindOauth2User(ctx context.Context, bindOauth2User *dmsV1.BindOauth2UserReq) (claims *biz.ClaimsInfo, err error) {
	d.log.Infof("BindOauth2User")
	defer func() {
		d.log.Infof("BindOauth2User;error=%v", err)
	}()

	return d.Oauth2ConfigurationUsecase.BindOauth2User(ctx, bindOauth2User.Oauth2Token, bindOauth2User.RefreshToken, bindOauth2User.UserName, bindOauth2User.Pwd)
}

func (d *DMSService) RefreshOauth2Token(ctx context.Context, userUid, sub, sid string) (claims *biz.ClaimsInfo, err error) {
	d.log.Infof("RefreshOauth2Token")
	defer func() {
		d.log.Infof("RefreshOauth2Token;error=%v", err)
	}()

	return d.Oauth2ConfigurationUsecase.RefreshOauth2Token(ctx, userUid, sub, sid)
}

func (d *DMSService) BackChannelLogout(ctx context.Context, logoutToken string) error {
	d.log.Infof("BackChannelLogout")

	err := d.Oauth2ConfigurationUsecase.BackChannelLogout(ctx, logoutToken)
	if err != nil {
		d.log.Errorf("BackChannelLogout logout token:%s err:%v", logoutToken, err)
		return err
	}

	return nil
}

func (d *DMSService) GetLDAPConfiguration(ctx context.Context) (reply *dmsV1.GetLDAPConfigurationReply, err error) {
	d.log.Infof("GetLDAPConfiguration")
	defer func() {
		d.log.Infof("GetLDAPConfiguration.reply=%v;error=%v", reply, err)
	}()

	ldapConfiguration, exist, err := d.LDAPConfigurationUsecase.GetLDAPConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetLDAPConfigurationReply{
			Data: dmsV1.LDAPConfigurationResData{},
		}, nil
	}

	return &dmsV1.GetLDAPConfigurationReply{
		Data: dmsV1.LDAPConfigurationResData{
			EnableLdap:          ldapConfiguration.Enable,
			EnableSSL:           ldapConfiguration.EnableSSL,
			LdapServerHost:      ldapConfiguration.Host,
			LdapServerPort:      ldapConfiguration.Port,
			LdapConnectDn:       ldapConfiguration.ConnectDn,
			LdapSearchBaseDn:    ldapConfiguration.BaseDn,
			LdapUserNameRdnKey:  ldapConfiguration.UserNameRdnKey,
			LdapUserEmailRdnKey: ldapConfiguration.UserEmailRdnKey,
		},
	}, nil
}

func (d *DMSService) UpdateLDAPConfiguration(ctx context.Context, req *dmsV1.UpdateLDAPConfigurationReq) (err error) {
	d.log.Infof("UpdateLDAPConfiguration")
	defer func() {
		d.log.Infof("UpdateLDAPConfiguration;error=%v", err)
	}()

	ldapConfiguration := req.LDAPConfiguration
	return d.LDAPConfigurationUsecase.UpdateLDAPConfiguration(ctx, ldapConfiguration.EnableLdap,
		ldapConfiguration.EnableSSL, ldapConfiguration.LdapServerHost, ldapConfiguration.LdapServerPort,
		ldapConfiguration.LdapConnectDn, ldapConfiguration.LdapConnectPwd, ldapConfiguration.LdapSearchBaseDn,
		ldapConfiguration.LdapUserNameRdnKey, ldapConfiguration.LdapUserEmailRdnKey)
}

func (d *DMSService) GetSMTPConfiguration(ctx context.Context) (reply *dmsV1.GetSMTPConfigurationReply, err error) {
	d.log.Infof("GetSMTPConfiguration")
	defer func() {
		d.log.Infof("GetSMTPConfiguration.reply=%v;error=%v", reply, err)
	}()

	smtpConfiguration, exist, err := d.SMTPConfigurationUsecase.GetSMTPConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetSMTPConfigurationReply{
			Data: dmsV1.SMTPConfigurationResData{},
		}, nil
	}

	return &dmsV1.GetSMTPConfigurationReply{
		Data: dmsV1.SMTPConfigurationResData{
			EnableSMTPNotify: smtpConfiguration.EnableSMTPNotify,
			Host:             smtpConfiguration.Host,
			Port:             smtpConfiguration.Port,
			Username:         smtpConfiguration.Username,
			IsSkipVerify:     smtpConfiguration.IsSkipVerify,
		},
	}, nil
}

func (d *DMSService) UpdateSMTPConfiguration(ctx context.Context, req *dmsV1.UpdateSMTPConfigurationReq) (err error) {
	d.log.Infof("UpdateSMTPConfiguration")
	defer func() {
		d.log.Infof("UpdateSMTPConfiguration;error=%v", err)
	}()

	smtpConfiguration := req.UpdateSMTPConfiguration
	return d.SMTPConfigurationUsecase.UpdateSMTPConfiguration(ctx, smtpConfiguration.Host,
		smtpConfiguration.Port, smtpConfiguration.Username, smtpConfiguration.Password,
		smtpConfiguration.EnableSMTPNotify, smtpConfiguration.IsSkipVerify)
}

func (d *DMSService) TestSMTPConfiguration(ctx context.Context, req *dmsV1.TestSMTPConfigurationReq) (reply *dmsV1.TestSMTPConfigurationReply, err error) {
	d.log.Infof("TestSMTPConfiguration,req:%v", req)
	defer func() {
		d.log.Infof("TestSMTPConfiguration;error=%v", err)
	}()

	isSMTPSendNormal, sendErrorMessage := true, "ok"
	err = d.SMTPConfigurationUsecase.TestSMTPConfiguration(ctx, req.TestSMTPConfiguration.RecipientAddr)
	if err != nil {
		isSMTPSendNormal = false
		sendErrorMessage = err.Error()
	}

	return &dmsV1.TestSMTPConfigurationReply{
		Data: dmsV1.TestSMTPConfigurationResData{
			IsSMTPSendNormal: isSMTPSendNormal,
			SendErrorMessage: sendErrorMessage,
		},
	}, nil
}

func (d *DMSService) GetWeChatConfiguration(ctx context.Context) (reply *dmsV1.GetWeChatConfigurationReply, err error) {
	d.log.Infof("GetWeChatConfiguration")
	defer func() {
		d.log.Infof("GetWeChatConfiguration.reply=%v;error=%v", reply, err)
	}()

	wechatConfiguration, exist, err := d.WeChatConfigurationUsecase.GetWeChatConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetWeChatConfigurationReply{
			Data: dmsV1.WeChatConfigurationResData{},
		}, nil
	}

	return &dmsV1.GetWeChatConfigurationReply{
		Data: dmsV1.WeChatConfigurationResData{
			EnableWeChatNotify: wechatConfiguration.EnableWeChatNotify,
			CorpID:             wechatConfiguration.CorpID,
			AgentID:            wechatConfiguration.AgentID,
			SafeEnabled:        wechatConfiguration.SafeEnabled,
			ProxyIP:            wechatConfiguration.ProxyIP,
		},
	}, nil
}

func (d *DMSService) UpdateWeChatConfiguration(ctx context.Context, req *dmsV1.UpdateWeChatConfigurationReq) (err error) {
	d.log.Infof("UpdateWeChatConfiguration")
	defer func() {
		d.log.Infof("UpdateWeChatConfiguration;error=%v", err)
	}()

	wechatConfiguration := req.UpdateWeChatConfiguration
	return d.WeChatConfigurationUsecase.UpdateWeChatConfiguration(ctx, wechatConfiguration.EnableWeChatNotify,
		wechatConfiguration.SafeEnabled, wechatConfiguration.AgentID, wechatConfiguration.CorpID, wechatConfiguration.CorpSecret, wechatConfiguration.ProxyIP)
}

func (d *DMSService) TestWeChatConfiguration(ctx context.Context, req *dmsV1.TestWeChatConfigurationReq) (reply *dmsV1.TestWeChatConfigurationReply, err error) {
	d.log.Infof("TestWeChatConfiguration,req:%v", req)
	defer func() {
		d.log.Infof("TestWeChatConfiguration;error=%v", err)
	}()

	isWeChatSendNormal, sendErrorMessage := true, "ok"
	err = d.WeChatConfigurationUsecase.TestWeChatConfiguration(ctx, req.TestWeChatConfiguration.RecipientID)
	if err != nil {
		isWeChatSendNormal = false
		sendErrorMessage = err.Error()
	}

	return &dmsV1.TestWeChatConfigurationReply{
		Data: dmsV1.TestWeChatConfigurationResData{
			IsWeChatSendNormal: isWeChatSendNormal,
			SendErrorMessage:   sendErrorMessage,
		},
	}, nil
}

func (d *DMSService) GetFeishuConfiguration(ctx context.Context) (reply *dmsV1.GetFeishuConfigurationReply, err error) {
	d.log.Infof("GetFeishuConfiguration")
	defer func() {
		d.log.Infof("GetFeishuConfiguration.reply=%v;error=%v", reply, err)
	}()

	feishuConfiguration, exist, err := d.IMConfigurationUsecase.GetIMConfiguration(ctx, biz.ImTypeFeishu)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetFeishuConfigurationReply{
			Data: dmsV1.FeishuConfigurationResData{},
		}, nil
	}
	return &dmsV1.GetFeishuConfigurationReply{
		Data: dmsV1.FeishuConfigurationResData{
			AppID:                       feishuConfiguration.AppKey,
			IsFeishuNotificationEnabled: feishuConfiguration.IsEnable,
		},
	}, nil
}

func (d *DMSService) UpdateFeishuConfiguration(ctx context.Context, req *dmsV1.UpdateFeishuConfigurationReq) (err error) {
	d.log.Infof("UpdateFeishuConfiguration")
	defer func() {
		d.log.Infof("UpdateFeishuConfiguration;error=%v", err)
	}()

	feishuConfiguration := req.UpdateFeishuConfiguration
	return d.IMConfigurationUsecase.UpdateIMConfiguration(ctx, feishuConfiguration.IsFeishuNotificationEnabled,
		feishuConfiguration.AppID, feishuConfiguration.AppSecret)
}

func (d *DMSService) TestFeishuConfiguration(ctx context.Context, req *dmsV1.TestFeishuConfigurationReq) (reply *dmsV1.TestFeishuConfigurationReply, err error) {
	d.log.Infof("TestFeishuConfiguration,req:%v", req)
	defer func() {
		d.log.Infof("TestFeishuConfiguration;error=%v", err)
	}()

	var users []*biz.User
	switch req.TestFeishuConfiguration.AccountType {
	case dmsV1.FeishuAccountTypeEmail:
		// TODO 校验email格式
		// err := controller.Validate(struct {
		// 	Email string `valid:"email"`
		// }{req.Account})
		// if err != nil {
		// 	return controller.JSONBaseErrorReq(c, errors.New(errors.DataInvalid, err))
		// }
		users = append(users, &biz.User{Email: req.TestFeishuConfiguration.Account})
	case dmsV1.FeishuAccountTypePhone:
		users = append(users, &biz.User{Phone: req.TestFeishuConfiguration.Account})
	default:
		return nil, fmt.Errorf("unknown account type: %v", req.TestFeishuConfiguration.AccountType)
	}
	isFeishuSendNormal, sendErrorMessage := true, "ok"
	err = d.IMConfigurationUsecase.TestFeishuConfiguration(ctx, users)
	if err != nil {
		isFeishuSendNormal = false
		sendErrorMessage = err.Error()
	}

	return &dmsV1.TestFeishuConfigurationReply{
		Data: dmsV1.TestFeishuConfigurationResData{
			IsMessageSentNormally: isFeishuSendNormal,
			ErrorMessage:          sendErrorMessage,
		},
	}, nil
}

func (d *DMSService) GetWebHookConfiguration(ctx context.Context) (reply *dmsV1.GetWebHookConfigurationReply, err error) {
	webhookConfiguration, exist, err := d.WebHookConfigurationUsecase.GetWebHookConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetWebHookConfigurationReply{
			Data: dmsV1.GetWebHookConfigurationReplyItem{},
		}, nil
	}

	return &dmsV1.GetWebHookConfigurationReply{
		Data: dmsV1.GetWebHookConfigurationReplyItem{
			Enable:               webhookConfiguration.Enable,
			MaxRetryTimes:        webhookConfiguration.MaxRetryTimes,
			RetryIntervalSeconds: webhookConfiguration.RetryIntervalSeconds,
			Token:                webhookConfiguration.Token,
			URL:                  webhookConfiguration.URL,
		},
	}, nil
}

func (d *DMSService) UpdateWebHookConfiguration(ctx context.Context, req *dmsV1.UpdateWebHookConfigurationReq) (err error) {
	d.log.Infof("UpdateWebHookConfiguration")
	defer func() {
		d.log.Infof("UpdateWebHookConfiguration;error=%v", err)
	}()

	webhookConfiguration := req.UpdateWebHookConfiguration
	return d.WebHookConfigurationUsecase.UpdateWebHookConfiguration(ctx, webhookConfiguration.Enable, webhookConfiguration.MaxRetryTimes, webhookConfiguration.RetryIntervalSeconds,
		webhookConfiguration.Token, webhookConfiguration.URL)
}

func (d *DMSService) TestWebHookConfiguration(ctx context.Context) (reply *dmsV1.TestWebHookConfigurationReply, err error) {
	d.log.Infof("TestWebHookConfiguration")
	defer func() {
		d.log.Infof("TestWebHookConfiguration;error=%v", err)
	}()

	isWebHookSendNormal, sendErrorMessage := true, "ok"
	err = d.WebHookConfigurationUsecase.TestWebHookConfiguration(ctx)
	if err != nil {
		isWebHookSendNormal = false
		sendErrorMessage = err.Error()
	}

	return &dmsV1.TestWebHookConfigurationReply{
		Data: dmsV1.TestWebHookConfigurationResData{
			IsMessageSentNormally: isWebHookSendNormal,
			SendErrorMessage:      sendErrorMessage,
		},
	}, nil
}

func (d *DMSService) UpdateSmsConfiguration(ctx context.Context, req *dmsV1.UpdateSmsConfigurationReq) (err error) {
	d.log.Infof("UpdateSmsConfiguration")
	defer func() {
		d.log.Infof("UpdateSmsConfiguration;error=%v", err)
	}()
	return d.SmsConfigurationUseCase.UpdateSmsConfiguration(ctx, req.UpdateSmsConfiguration.EnableSms, req.UpdateSmsConfiguration.Url, req.UpdateSmsConfiguration.SmsType, req.UpdateSmsConfiguration.Configuration)
}

func (d *DMSService) TestSmsConfiguration(ctx context.Context, req *dmsV1.TestSmsConfigurationReq) (reply *dmsV1.TestSmsConfigurationReply, err error) {
	d.log.Infof("TestSmsConfiguration")
	defer func() {
		d.log.Infof("TestSmsConfiguration;error=%v", err)
	}()
	isSmsSendNormal, sendErrorMessage := true, "ok"
	err = d.SmsConfigurationUseCase.TestSmsConfiguration(ctx, req.TestSmsConfiguration.RecipientPhone)
	if err != nil {
		isSmsSendNormal = false
		sendErrorMessage = err.Error()
	}
	return &dmsV1.TestSmsConfigurationReply{
		Data: dmsV1.TestSmsConfigurationResData{
			IsSmsSendNormal:  isSmsSendNormal,
			SendErrorMessage: sendErrorMessage,
		},
	}, nil
}

func (d *DMSService) GetSmsConfiguration(ctx context.Context) (reply *dmsV1.GetSmsConfigurationReply, err error) {
	smsConfiguration, exist, err := d.SmsConfigurationUseCase.GetSmsConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return &dmsV1.GetSmsConfigurationReply{
			Data: dmsV1.GetSmsConfigurationReplyItem{},
		}, nil
	}
	configuration := map[string]string{}
	err = json.Unmarshal([]byte(smsConfiguration.Configuration), &configuration)
	return &dmsV1.GetSmsConfigurationReply{
		Data: dmsV1.GetSmsConfigurationReplyItem{
			Enable:        smsConfiguration.Enable,
			Url:           smsConfiguration.Url,
			Configuration: configuration,
			SmsType:       smsConfiguration.Type,
		},
	}, nil
}

func (d *DMSService) SendSmsCode(ctx context.Context, username string) (reply *dmsV1.SendSmsCodeReply, err error) {
	d.log.Infof("send sms code")
	defer func() {
		d.log.Infof("send sms code;error=%v", err)
	}()
	return d.SmsConfigurationUseCase.SendSmsCode(ctx, username)
}

func (d *DMSService) VerifySmsCode(request *dmsV1.VerifySmsCodeReq) (reply *dmsV1.VerifySmsCodeReply) {
	d.log.Infof("verify sms code")
	defer func() {
		d.log.Infof("verify sms code %v", reply)
	}()
	return d.SmsConfigurationUseCase.VerifySmsCode(request.Code, request.Username)
}

func (d *DMSService) NotifyMessage(ctx context.Context, req *dmsCommonV1.NotificationReq) (err error) {
	d.log.Infof("notifyMessage")
	defer func() {
		d.log.Infof("notifyMessage;error=%v", err)
	}()

	users, _, err := d.UserUsecase.ListUser(ctx, &biz.ListUsersOption{
		FilterBy: []pkgConst.FilterCondition{
			{
				Field:    string(biz.UserFieldUID),
				Operator: pkgConst.FilterOperatorIn,
				Value:    req.Notification.UserUids,
			},
		},
		OrderBy:      biz.UserFieldName,
		PageNumber:   1,
		LimitPerPage: uint32(len(req.Notification.UserUids)),
	})

	lang2Users := make(map[language.Tag][]*biz.User, len(locale.Bundle.LanguageTags()))
	for _, user := range users {
		langTag := locale.Bundle.MatchLangTag(user.Language)
		lang2Users[langTag] = append(lang2Users[langTag], user)
	}

	for _, n := range biz.Notifiers {
		for langTag, u := range lang2Users {
			err = n.Notify(ctx, req.Notification.NotificationSubject.GetStrInLang(langTag), req.Notification.NotificationBody.GetStrInLang(langTag), u)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DMSService) WebHookSendMessage(ctx context.Context, req *dmsCommonV1.WebHookSendMessageReq) (err error) {
	d.log.Infof("WebHookSendMessage")
	defer func() {
		d.log.Infof("WebHookSendMessage;error=%v", err)
	}()

	return d.WebHookConfigurationUsecase.SendWebHookMessage(ctx, string(req.WebHookMessage.TriggerEventType), req.WebHookMessage.Message)
}
