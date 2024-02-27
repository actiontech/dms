package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) GetOauth2Configuration(ctx context.Context) (reply *dmsV1.GetOauth2ConfigurationReply, err error) {
	d.log.Infof("GetOauth2Configuration")
	defer func() {
		d.log.Infof("GetOauth2Configuration.reply=%v;error=%v", reply, err)
	}()

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
			EnableOauth2:    oauth2C.EnableOauth2,
			AutoCreateUser:  &oauth2C.AutoCreateUser,
			ClientID:        oauth2C.ClientID,
			ClientHost:      oauth2C.ClientHost,
			ServerAuthUrl:   oauth2C.ServerAuthUrl,
			ServerTokenUrl:  oauth2C.ServerTokenUrl,
			ServerUserIdUrl: oauth2C.ServerUserIdUrl,
			Scopes:          oauth2C.Scopes,
			AccessTokenTag:  oauth2C.AccessTokenTag,
			UserIdTag:       oauth2C.UserIdTag,
			LoginTip:        oauth2C.LoginTip,
			UserEmailTag:    oauth2C.UserEmailTag,
			UserWeChatTag:   oauth2C.UserWeChatTag,
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
	return d.Oauth2ConfigurationUsecase.UpdateOauth2Configuration(
		ctx,
		oauth2Configuration.EnableOauth2,
		oauth2Configuration.AutoCreateUser,
		oauth2Configuration.ClientID,
		oauth2Configuration.ClientKey,
		oauth2Configuration.ClientHost,
		oauth2Configuration.ServerAuthUrl,
		oauth2Configuration.ServerTokenUrl,
		oauth2Configuration.ServerUserIdUrl,
		oauth2Configuration.AccessTokenTag,
		oauth2Configuration.UserIdTag,
		oauth2Configuration.UserWeChatTag,
		oauth2Configuration.UserEmailTag,
		oauth2Configuration.LoginTip,
		oauth2Configuration.Scopes,
	)
}

func (d *DMSService) Oauth2Link(ctx context.Context) (uri string, err error) {
	d.log.Infof("Oauth2Link")
	defer func() {
		d.log.Infof("Oauth2Link;error=%v", err)
	}()

	uri, err = d.Oauth2ConfigurationUsecase.GenOauth2LinkURI(ctx)
	if err != nil {
		return "", err
	}
	return uri, nil
}

// if redirect directly to SQLE, will return token, otherwise this parameter will be an empty string
func (d *DMSService) Oauth2Callback(ctx context.Context, req *dmsV1.Oauth2CallbackReq) (uri string, token string, err error) {
	d.log.Infof("Oauth2Callback")
	defer func() {
		d.log.Infof("Oauth2Callback;error=%v", err)
	}()

	uri, token, err = d.Oauth2ConfigurationUsecase.GenerateCallbackUri(ctx, req.State, req.Code)
	if err != nil {
		return "", "", err
	}
	return uri, token, nil
}

func (d *DMSService) BindOauth2User(ctx context.Context, bindOauth2User *dmsV1.BindOauth2UserReq) (reply *dmsV1.BindOauth2UserReply, err error) {
	d.log.Infof("BindOauth2User")
	defer func() {
		d.log.Infof("BindOauth2User;error=%v", err)
	}()

	token, err := d.Oauth2ConfigurationUsecase.BindOauth2User(ctx, bindOauth2User.Oauth2Token, bindOauth2User.UserName, bindOauth2User.Pwd)
	if err != nil {
		return nil, err
	}
	return &dmsV1.BindOauth2UserReply{
		Data: dmsV1.BindOauth2UserResData{Token: token},
	}, nil
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
	for _, n := range biz.Notifiers {
		err = n.Notify(ctx, req.Notification.NotificationSubject, req.Notification.NotificationBody, users)
		if err != nil {
			return err
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
