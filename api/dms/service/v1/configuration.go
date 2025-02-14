package v1

import base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"

// swagger:model GetLoginTipsReply
type GetLoginTipsReply struct {
	Data LoginTipsResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type LoginTipsResData struct {
	LoginButtonText     string `json:"login_button_text"`
	DisableUserPwdLogin bool   `json:"disable_user_pwd_login"`
	Oauth2LoginTip      string `json:"oauth2_login_tip"`
	EnableOauth2        bool   `json:"enable_oauth2"`
}

// swagger:model
type UpdateLoginConfigurationReq struct {
	LoginConfiguration LoginConfiguration `json:"login" validate:"required"`
}

type LoginConfiguration struct {
	LoginButtonText     *string `json:"login_button_text"`
	DisableUserPwdLogin *bool   `json:"disable_user_pwd_login"`
}

type GetOauth2ConfigurationResData struct {
	EnableOauth2    bool     `json:"enable_oauth2"`
	SkipCheckState  bool     `json:"skip_check_state"`
	AutoCreateUser  bool     `json:"auto_create_user"`
	ClientID        string   `json:"client_id"`
	ClientHost      string   `json:"client_host"`
	ServerAuthUrl   string   `json:"server_auth_url"`
	ServerTokenUrl  string   `json:"server_token_url"`
	ServerUserIdUrl string   `json:"server_user_id_url"`
	ServerLogoutUrl string   `json:"server_logout_url"`
	Scopes          []string `json:"scopes"`
	AccessTokenTag  string   `json:"access_token_tag"`
	UserIdTag       string   `json:"user_id_tag"`
	UserEmailTag    string   `json:"user_email_tag"`
	UserWeChatTag   string   `json:"user_wechat_tag"`
	LoginTip        string   `json:"login_tip"`
}

// swagger:model GetOauth2ConfigurationResDataReply
type GetOauth2ConfigurationReply struct {
	Data GetOauth2ConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

// swagger:model
type Oauth2ConfigurationReq struct {
	Oauth2Configuration Oauth2Configuration `json:"oauth2" validate:"required"`
}
type Oauth2Configuration struct {
	EnableOauth2      *bool     `json:"enable_oauth2"`
	SkipCheckState    *bool     `json:"skip_check_state"`
	AutoCreateUser    *bool     `json:"auto_create_user"`
	AutoCreateUserPWD *string   `json:"auto_create_user_pwd"`
	ClientID          *string   `json:"client_id"`
	ClientKey         *string   `json:"client_key"`
	ClientHost        *string   `json:"client_host"`
	ServerAuthUrl     *string   `json:"server_auth_url"`
	ServerTokenUrl    *string   `json:"server_token_url"`
	ServerUserIdUrl   *string   `json:"server_user_id_url"`
	ServerLogoutUrl   *string   `json:"server_logout_url"`
	Scopes            *[]string `json:"scopes"`
	AccessTokenTag    *string   `json:"access_token_tag"`
	UserIdTag         *string   `json:"user_id_tag"`
	UserEmailTag      *string   `json:"user_email_tag"`
	UserWeChatTag     *string   `json:"user_wechat_tag"`
	// Maximum: 28
	LoginTip *string `json:"login_tip" validate:"max=28"`
}

// swagger:model GetOauth2TipsReply
type GetOauth2TipsReply struct {
	Data GetOauth2TipsResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type GetOauth2TipsResData struct {
	EnableOauth2 bool   `json:"enable_oauth2"`
	LoginTip     string `json:"login_tip"`
}

// swagger:model
type BindOauth2UserReq struct {
	UserName    string `json:"user_name" form:"user_name" validate:"required"`
	Pwd         string `json:"pwd" form:"pwd" validate:"required"`
	Oauth2Token string `json:"oauth2_token" form:"oauth2_token" validate:"required"`
	IdToken     string `json:"id_token" form:"id_token"`
}

// swagger:model BindOauth2UserReply
type BindOauth2UserReply struct {
	Data BindOauth2UserResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type BindOauth2UserResData struct {
	Token string `json:"token"`
}

// swagger:parameters Oauth2Callback
type Oauth2CallbackReq struct {
	State string `json:"state" query:"state"`
	Code  string `json:"code" query:"code" validate:"required"`
}

// swagger:model GetLDAPConfigurationResDataReply
type GetLDAPConfigurationReply struct {
	Data LDAPConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type LDAPConfigurationResData struct {
	EnableLdap          bool   `json:"enable_ldap"`
	EnableSSL           bool   `json:"enable_ssl"`
	LdapServerHost      string `json:"ldap_server_host"`
	LdapServerPort      string `json:"ldap_server_port"`
	LdapConnectDn       string `json:"ldap_connect_dn"`
	LdapSearchBaseDn    string `json:"ldap_search_base_dn"`
	LdapUserNameRdnKey  string `json:"ldap_user_name_rdn_key"`
	LdapUserEmailRdnKey string `json:"ldap_user_email_rdn_key"`
}

// swagger:model
type UpdateLDAPConfigurationReq struct {
	LDAPConfiguration LDAPConfiguration `json:"ldap" validate:"required"`
}

type LDAPConfiguration struct {
	EnableLdap          *bool   `json:"enable_ldap"`
	EnableSSL           *bool   `json:"enable_ssl"`
	LdapServerHost      *string `json:"ldap_server_host"`
	LdapServerPort      *string `json:"ldap_server_port"`
	LdapConnectDn       *string `json:"ldap_connect_dn"`
	LdapConnectPwd      *string `json:"ldap_connect_pwd"`
	LdapSearchBaseDn    *string `json:"ldap_search_base_dn"`
	LdapUserNameRdnKey  *string `json:"ldap_user_name_rdn_key"`
	LdapUserEmailRdnKey *string `json:"ldap_user_email_rdn_key"`
}

// swagger:model GetSMTPConfigurationReply
type GetSMTPConfigurationReply struct {
	Data SMTPConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type SMTPConfigurationResData struct {
	EnableSMTPNotify bool   `json:"enable_smtp_notify"`
	Host             string `json:"smtp_host"`
	Port             string `json:"smtp_port"`
	Username         string `json:"smtp_username"`
	IsSkipVerify     bool   `json:"is_skip_verify"`
}

// swagger:model
type UpdateSMTPConfigurationReq struct {
	UpdateSMTPConfiguration UpdateSMTPConfiguration `json:"smtp_configuration" validate:"required"`
}

type UpdateSMTPConfiguration struct {
	EnableSMTPNotify *bool   `json:"enable_smtp_notify" from:"enable_smtp_notify" description:"是否启用邮件通知"`
	Host             *string `json:"smtp_host" form:"smtp_host" example:"smtp.email.qq.com"`
	Port             *string `json:"smtp_port" form:"smtp_port" example:"465"`
	Username         *string `json:"smtp_username" form:"smtp_username" example:"test@qq.com"`
	Password         *string `json:"smtp_password" form:"smtp_password" example:"123"`
	IsSkipVerify     *bool   `json:"is_skip_verify" form:"is_skip_verify" description:"是否启用邮件通知"`
}

// swagger:model
type TestSMTPConfigurationReq struct {
	TestSMTPConfiguration TestSMTPConfiguration `json:"test_smtp_configuration" validate:"required,email"`
}

type TestSMTPConfiguration struct {
	RecipientAddr string `json:"recipient_addr" validate:"required,email"`
}

// swagger:model TestSMTPConfigurationReply
type TestSMTPConfigurationReply struct {
	Data TestSMTPConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type TestSMTPConfigurationResData struct {
	IsSMTPSendNormal bool   `json:"is_smtp_send_normal"`
	SendErrorMessage string `json:"send_error_message,omitempty"`
}

// swagger:model GetWeChatConfigurationReply
type GetWeChatConfigurationReply struct {
	Data WeChatConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type WeChatConfigurationResData struct {
	EnableWeChatNotify bool   `json:"enable_wechat_notify"`
	CorpID             string `json:"corp_id"`
	AgentID            int    `json:"agent_id"`
	SafeEnabled        bool   `json:"safe_enabled"`
	ProxyIP            string `json:"proxy_ip"`
}

// swagger:model
type UpdateWeChatConfigurationReq struct {
	UpdateWeChatConfiguration UpdateWeChatConfiguration `json:"update_wechat_configuration"`
}

type UpdateWeChatConfiguration struct {
	EnableWeChatNotify *bool   `json:"enable_wechat_notify" from:"enable_wechat_notify" description:"是否启用微信通知"`
	CorpID             *string `json:"corp_id" from:"corp_id" description:"企业微信ID"`
	CorpSecret         *string `json:"corp_secret" from:"corp_secret" description:"企业微信ID对应密码"`
	AgentID            *int    `json:"agent_id" from:"agent_id" description:"企业微信应用ID"`
	SafeEnabled        *bool   `json:"safe_enabled" from:"safe_enabled" description:"是否对传输信息加密"`
	ProxyIP            *string `json:"proxy_ip" from:"proxy_ip" description:"企业微信代理服务器IP"`
}

// swagger:model
type TestWeChatConfigurationReq struct {
	TestWeChatConfiguration TestWeChatConfiguration `json:"test_wechat_configuration"`
}

type TestWeChatConfiguration struct {
	RecipientID string `json:"recipient_id" from:"recipient_id" description:"消息接收者企业微信ID"`
}

// swagger:model TestWeChatConfigurationReply
type TestWeChatConfigurationReply struct {
	Data TestWeChatConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type TestWeChatConfigurationResData struct {
	IsWeChatSendNormal bool   `json:"is_wechat_send_normal"`
	SendErrorMessage   string `json:"send_error_message,omitempty"`
}

// swagger:model GetFeishuConfigurationReply
type GetFeishuConfigurationReply struct {
	Data FeishuConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type FeishuConfigurationResData struct {
	AppID                       string `json:"app_id"`
	IsFeishuNotificationEnabled bool   `json:"is_feishu_notification_enabled"`
}

// swagger:model
type UpdateFeishuConfigurationReq struct {
	UpdateFeishuConfiguration UpdateFeishuConfiguration `json:"update_feishu_configuration"`
}

type UpdateFeishuConfiguration struct {
	AppID                       *string `json:"app_id" form:"app_id"`
	AppSecret                   *string `json:"app_secret" form:"app_secret" `
	IsFeishuNotificationEnabled *bool   `json:"is_feishu_notification_enabled" from:"is_feishu_notification_enabled" description:"是否启用飞书推送"`
}

// swagger:model
type TestFeishuConfigurationReq struct {
	TestFeishuConfiguration TestFeishuConfiguration `json:"test_feishu_configuration" validate:"required"`
}

// swagger:enum FeishuAccountType
type FeishuAccountType string

const (
	FeishuAccountTypeEmail FeishuAccountType = "email"
	FeishuAccountTypePhone FeishuAccountType = "phone"
)

type TestFeishuConfiguration struct {
	AccountType FeishuAccountType `json:"account_type" form:"account_type" enums:"email,phone" validate:"required"`
	Account     string            `json:"account" form:"account" validate:"required" description:"绑定了飞书的手机号或邮箱"`
}

// swagger:model TestFeishuConfigurationReply
type TestFeishuConfigurationReply struct {
	Data TestFeishuConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type TestFeishuConfigurationResData struct {
	IsMessageSentNormally bool   `json:"is_message_sent_normally"`
	ErrorMessage          string `json:"error_message,omitempty"`
}

type GetWebHookConfigurationReplyItem struct {
	Enable bool `json:"enable" description:"是否启用"`
	// minlength(3) maxlength(100)
	MaxRetryTimes        int    `json:"max_retry_times" description:"最大重试次数"`
	RetryIntervalSeconds int    `json:"retry_interval_seconds" description:"请求重试间隔"`
	Token                string `json:"token" description:"token 令牌"`
	URL                  string `json:"url" description:"回调API URL"`
}

// swagger:model GetWebHookConfigurationReply
type GetWebHookConfigurationReply struct {
	Data GetWebHookConfigurationReplyItem `json:"data"`

	// Generic reply
	base.GenericResp
}

type WebHookConfigurationData struct {
	Enable *bool `json:"enable" description:"是否启用"`
	// minlength(3) maxlength(100)
	MaxRetryTimes        *int    `json:"max_retry_times" description:"最大重试次数"`
	RetryIntervalSeconds *int    `json:"retry_interval_seconds" description:"请求重试间隔"`
	Token                *string `json:"token" description:"token 令牌"`
	URL                  *string `json:"url" description:"回调API URL"`
}

// swagger:model
type UpdateWebHookConfigurationReq struct {
	UpdateWebHookConfiguration WebHookConfigurationData `json:"webhook_config"`
}

// swagger:model TestWebHookConfigurationReply
type TestWebHookConfigurationReply struct {
	Data TestWebHookConfigurationResData `json:"data"`

	// Generic reply
	base.GenericResp
}

type TestWebHookConfigurationResData struct {
	IsMessageSentNormally bool   `json:"is_message_sent_normally"`
	SendErrorMessage      string `json:"send_error_message,omitempty"`
}
