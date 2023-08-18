package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/actiontech/dms/pkg/params"
	"github.com/actiontech/dms/pkg/periods"

	"gorm.io/gorm"
)

func GetAllModels() []interface{} {
	return []interface{}{
		DBService{},
		User{},
		UserGroup{},
		Role{},
		OpPermission{},
		DMSConfig{},
		Member{},
		MemberRoleOpRange{},
		Namespace{},
		ProxyTarget{},
		Plugin{},
		Oauth2Configuration{},
		LDAPConfiguration{},
		SMTPConfiguration{},
		WebHookConfiguration{},
		WeChatConfiguration{},
		IMConfiguration{},
		CloudbeaverUserCache{},
		CloudbeaverConnectionCache{},
		DatabaseSourceService{},
	}
}

type Model struct {
	UID       string    `json:"uid" gorm:"primaryKey" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2018-10-21T16:40:23+08:00"`
	UpdatedAt time.Time `json:"updated_at" example:"2018-10-21T16:40:23+08:00"`
}

type DBService struct {
	Model
	Name              string          `json:"name" gorm:"size:200;not null;uniqueIndex" example:""`
	DBType            string          `json:"db_type" gorm:"column:db_type; not null" example:"mysql"`
	Host              string          `json:"host" gorm:"column:db_host; not null" example:"10.10.10.10"`
	Port              string          `json:"port" gorm:"column:db_port; not null" example:"3306"`
	User              string          `json:"user" gorm:"column:db_user; not null" example:"root"`
	Password          string          `json:"password" gorm:"column:db_password; not null"`
	Desc              string          `json:"desc" gorm:"column:desc" example:"this is a instance"`
	Business          string          `json:"business" gorm:"column:business; not null" example:"this is a business"`
	AdditionalParams  params.Params   `json:"additional_params" gorm:"type:text"`
	Source            string          `json:"source" gorm:"not null"`
	NamespaceUID      string          `json:"namespace_uid" gorm:"column:namespace_uid"`
	MaintenancePeriod periods.Periods `json:"maintenance_period" gorm:"type:text"`
	ExtraParameters   ExtraParameters `json:"extra_parameters" gorm:"TYPE:json"`
}

type ExtraParameters struct {
	SqleConfig *SQLEConfig `json:"sqle_config"`
}

func (e ExtraParameters) Value() (driver.Value, error) {
	b, err := json.Marshal(e)
	return string(b), err
}

func (e *ExtraParameters) Scan(input interface{}) error {
	bytes, _ := input.([]byte)
	return json.Unmarshal(bytes, e)
}

type SQLEConfig struct {
	RuleTemplateID   string          `json:"rule_template_id"`
	RuleTemplateName string          `json:"rule_template_name"`
	SqlQueryConfig   *SqlQueryConfig `json:"sql_query_config"`
}
type SqlQueryConfig struct {
	MaxPreQueryRows                  int    `json:"max_pre_query_rows"`
	QueryTimeoutSecond               int    `json:"query_timeout_second"`
	AuditEnabled                     bool   `json:"audit_enabled"`
	AllowQueryWhenLessThanAuditLevel string `json:"allow_query_when_less_than_audit_level"`
}

type User struct {
	Model
	Name                   string         `json:"name" gorm:"size:200;column:name"`
	ThirdPartyUserID       string         `json:"third_party_user_id" gorm:"third_party_user_id;column:third_party_user_id"`
	Email                  string         `json:"email" gorm:"column:email"`
	Phone                  string         `json:"phone" gorm:"column:phone"`
	WeChatID               string         `json:"wechat_id" gorm:"column:wechat_id"`
	Password               string         `json:"password" gorm:"column:password"`
	UserAuthenticationType string         `json:"user_authentication_type" gorm:"not null;column:user_authentication_type"`
	Stat                   uint           `json:"stat" gorm:"not null"`
	LastLoginAt            *time.Time     `json:"last_login_at" gorm:"column:last_login_at"`
	DeletedAt              gorm.DeletedAt `json:"delete_at" gorm:"column:delete_at" sql:"index"`

	UserGroups    []*UserGroup    `gorm:"many2many:user_group_users"`
	OpPermissions []*OpPermission `gorm:"many2many:user_op_permissions"`
}

type UserGroup struct {
	Model
	Name string `json:"name" gorm:"size:200;uniqueIndex"`
	Desc string `json:"desc" gorm:"column:description"`
	Stat uint   `json:"stat" gorm:"not null"`

	Users []*User `gorm:"many2many:user_group_users"`
}

type Role struct {
	Model
	Name string `json:"name" gorm:"size:200;uniqueIndex"`
	Desc string `json:"desc" gorm:"column:description"`
	Stat uint   `json:"stat" gorm:"not null"`

	OpPermissions []*OpPermission `gorm:"many2many:role_op_permissions"`
}

type OpPermission struct {
	Model
	Name      string `json:"name" gorm:"size:200;uniqueIndex"`
	Desc      string `json:"desc" gorm:"column:description"`
	RangeType string `json:"range_type" gorm:"column:range_type"`
}

type DMSConfig struct {
	Model
	NeedInitOpPermissions bool `json:"need_init_op_permissions" gorm:"column:need_init_op_permissions"`
	NeedInitUsers         bool `json:"need_init_users" gorm:"column:need_init_users"`
	NeedInitRoles         bool `json:"need_init_roles" gorm:"column:need_init_roles"`
	NeedInitNamespaces    bool `json:"need_init_namespaces" gorm:"column:need_init_namespaces"`
}

type Member struct {
	Model
	UserUID          string              `json:"user_uid" gorm:"column:user_uid"`
	NamespaceUID     string              `json:"namespace_uid" gorm:"column:namespace_uid"`
	RoleWithOpRanges []MemberRoleOpRange `json:"role_with_op_ranges" gorm:"foreignKey:MemberUID;references:UID"`
}

type MemberRoleOpRange struct {
	MemberUID   string `json:"member_uid" gorm:"size:200;column:member_uid"`
	RoleUID     string `json:"role_uid" gorm:"column:role_uid"`
	OpRangeType string `json:"op_range_type" gorm:"column:op_range_type"`
	RangeUIDs   string `json:"range_uids" gorm:"type:text;column:range_uids"`
}

type Namespace struct {
	Model
	Name          string `json:"name" gorm:"column:name"`
	Desc          string `json:"desc" gorm:"column:desc"`
	CreateUserUID string `json:"create_user_uid" gorm:"column:create_user_uid"`
	Status        string `gorm:"default:'active'"`
}

type ProxyTarget struct {
	Name            string `json:"name" gorm:"primaryKey;size:200;not null;column:name"`
	Url             string `json:"url" gorm:"column:url"`
	ProxyUrlPrefixs string `json:"proxy_url_prefixs" gorm:"column:proxy_url_prefixs"`
}

type Plugin struct {
	Name                         string `json:"name" gorm:"primaryKey;size:200;not null;column:name"`
	AddDBServicePreCheckUrl      string `json:"add_db_service_pre_check_url" gorm:"column:add_db_service_pre_check_url"`
	DelDBServicePreCheckUrl      string `json:"del_db_service_pre_check_url" gorm:"column:del_db_service_pre_check_url"`
	DelUserPreCheckUrl           string `json:"del_user_pre_check_url" gorm:"column:del_user_pre_check_url"`
	DelUserGroupPreCheckUrl      string `json:"del_user_group_pre_check_url" gorm:"column:del_user_group_pre_check_url"`
	OperateDataResourceHandleUrl string `json:"operate_data_resource_handle_url" gorm:"column:operate_data_resource_handle_url"`
}

// Oauth2Configuration store oauth2 server configuration.
type Oauth2Configuration struct {
	Model
	EnableOauth2    bool   `json:"enable_oauth2" gorm:"column:enable_oauth2"`
	ClientID        string `json:"client_id" gorm:"column:client_id"`
	ClientKey       string `json:"-" gorm:"-"`
	ClientSecret    string `json:"client_secret" gorm:"client_secret"`
	ClientHost      string `json:"client_host" gorm:"column:client_host"`
	ServerAuthUrl   string `json:"server_auth_url" gorm:"column:server_auth_url"`
	ServerTokenUrl  string `json:"server_token_url" gorm:"column:server_token_url"`
	ServerUserIdUrl string `json:"server_user_id_url" gorm:"column:server_user_id_url"`
	Scopes          string `json:"scopes" gorm:"column:scopes"`
	AccessTokenTag  string `json:"access_token_tag" gorm:"column:access_token_tag"`
	UserIdTag       string `json:"user_id_tag" gorm:"column:user_id_tag"`
	LoginTip        string `json:"login_tip" gorm:"column:login_tip; default:'使用第三方账户登录'"`
}

// LDAPConfiguration store ldap server configuration.
type LDAPConfiguration struct {
	Model
	// whether the ldap is enabled
	Enable bool `json:"enable" gorm:"not null"`
	// whether the ssl is enabled
	EnableSSL bool `json:"enable_ssl" gorm:"not null"`
	// ldap server's ip
	Host string `json:"host" gorm:"not null"`
	// ldap server's port
	Port string `json:"port" gorm:"not null"`
	// the DN of the ldap administrative user for verification
	ConnectDn string `json:"connect_dn" gorm:"not null"`
	// the secret password of the ldap administrative user for verification
	ConnectSecretPassword string `json:"connect_secret_password" gorm:"not null"`
	// base dn used for ldap verification
	BaseDn string `json:"base_dn" gorm:"not null"`
	// the key corresponding to the user name in ldap
	UserNameRdnKey string `json:"ldap_user_name_rdn_key" gorm:"not null"`
	// the key corresponding to the user email in ldap
	UserEmailRdnKey string `json:"ldap_user_email_rdn_key" gorm:"not null"`
}

// SMTPConfiguration store SMTP server configuration.
type SMTPConfiguration struct {
	Model
	EnableSMTPNotify bool   `json:"enable_smtp_notify" gorm:"default:true"`
	Host             string `json:"smtp_host" gorm:"column:smtp_host; not null"`
	Port             string `json:"smtp_port" gorm:"column:smtp_port; not null"`
	Username         string `json:"smtp_username" gorm:"column:smtp_username; not null"`
	SecretPassword   string `json:"secret_smtp_password" gorm:"column:secret_smtp_password; not null"`
	IsSkipVerify     bool   `json:"is_skip_verify" gorm:"default:false; not null"`
}

// WeChatConfiguration store WeChat configuration.
type WeChatConfiguration struct {
	Model
	EnableWeChatNotify  bool   `json:"enable_wechat_notify" gorm:"not null"`
	CorpID              string `json:"corp_id" gorm:"not null"`
	EncryptedCorpSecret string `json:"encrypted_corp_secret" gorm:"not null"`
	AgentID             int    `json:"agent_id" gorm:"not null"`
	SafeEnabled         bool   `json:"safe_enabled" gorm:"not null"`
	ProxyIP             string `json:"proxy_ip"`
}

type WebHookConfiguration struct {
	Model
	Enable               bool   `json:"enable" gorm:"default:true;not null"`
	MaxRetryTimes        int    `json:"max_retry_times" gorm:"not null"`
	RetryIntervalSeconds int    `json:"retry_interval_seconds" gorm:"not null"`
	EncryptedToken       string `json:"encrypted_token" gorm:"not null"`
	URL                  string `json:"url" gorm:"not null"`
}

const (
	ImTypeDingTalk = "dingTalk"
	ImTypeFeishu   = "feishu"
)

// Instant messaging config
type IMConfiguration struct {
	Model
	AppKey      string `json:"app_key" gorm:"column:app_key"`
	AppSecret   string `json:"app_secret" gorm:"column:app_secret"`
	IsEnable    bool   `json:"is_enable" gorm:"column:is_enable"`
	ProcessCode string `json:"process_code" gorm:"column:process_code"`
	// 类型唯一
	Type string `json:"type" gorm:"unique"`
}

type CloudbeaverUserCache struct {
	DMSUserID         string `json:"dms_user_id" gorm:"column:dms_user_id;primaryKey"`
	DMSFingerprint    string `json:"dms_fingerprint" gorm:"column:dms_fingerprint"`
	CloudbeaverUserID string `json:"cloudbeaver_user_id" gorm:"column:cloudbeaver_user_id"`
}

type CloudbeaverConnectionCache struct {
	DMSDBServiceID          string `json:"dms_db_service_id" gorm:"column:dms_db_service_id;primaryKey"`
	DMSDBServiceFingerprint string `json:"dms_db_service_fingerprint" gorm:"column:dms_db_service_fingerprint"`
	CloudbeaverConnectionID string `json:"cloudbeaver_connection_id" gorm:"column:cloudbeaver_connection_id"`
}

type DatabaseSourceService struct {
	Model
	Name         string `json:"name" gorm:"size:200;not null;uniqueIndex" example:""`
	Source       string `json:"source" gorm:"not null"`
	Version      string `json:"version" gorm:"not null"`
	URL          string `json:"url" gorm:"not null"`
	DbType       string `json:"db_type" gorm:"not null"`
	NamespaceUID string `json:"namespace_uid" gorm:"column:namespace_uid"`
	// Cron表达式
	CronExpress         string          `json:"cron_express" gorm:"column:cron_express; not null"`
	LastSyncErr         string          `json:"last_sync_err" gorm:"column:last_sync_err"`
	LastSyncSuccessTime *time.Time      `json:"last_sync_success_time" gorm:"column:last_sync_success_time"`
	ExtraParameters     ExtraParameters `json:"extra_parameters" gorm:"TYPE:json"`
}
