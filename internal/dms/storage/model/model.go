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
		MemberGroup{},
		MemberGroupRoleOpRange{},
		Project{},
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
		BasicConfig{},
		CompanyNotice{},
	}
}

type Model struct {
	UID       string    `json:"uid" gorm:"primaryKey;size:32" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2018-10-21T16:40:23+08:00"`
	UpdatedAt time.Time `json:"updated_at" example:"2018-10-21T16:40:23+08:00"`
}

type DBService struct {
	Model
	Name              string          `json:"name" gorm:"size:200;not null;index:project_uid_name,unique" example:""`
	DBType            string          `json:"db_type" gorm:"size:255;column:db_type; not null" example:"mysql"`
	Host              string          `json:"host" gorm:"column:db_host;size:255; not null" example:"10.10.10.10"`
	Port              string          `json:"port" gorm:"column:db_port;size:255; not null" example:"3306"`
	User              string          `json:"user" gorm:"column:db_user;size:255; not null" example:"root"`
	Password          string          `json:"password" gorm:"column:db_password; size:255; not null"`
	Desc              string          `json:"desc" gorm:"column:desc" example:"this is a instance"`
	Business          string          `json:"business" gorm:"column:business;size:255; not null" example:"this is a business"`
	AdditionalParams  params.Params   `json:"additional_params" gorm:"type:text"`
	Source            string          `json:"source" gorm:"size:255;not null"`
	ProjectUID        string          `json:"project_uid" gorm:"size:32;column:project_uid;index:project_uid_name,unique"`
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
	ThirdPartyUserID       string         `json:"third_party_user_id" gorm:"size:255;third_party_user_id;column:third_party_user_id"`
	Email                  string         `json:"email" gorm:"size:255;column:email"`
	Phone                  string         `json:"phone" gorm:"size:255;column:phone"`
	WeChatID               string         `json:"wechat_id" gorm:"size:255;column:wechat_id"`
	Password               string         `json:"password" gorm:"size:255;column:password"`
	UserAuthenticationType string         `json:"user_authentication_type" gorm:"size:255;not null;column:user_authentication_type"`
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
	Stat uint   `json:"stat" gorm:"size:32;not null"`

	Users []*User `gorm:"many2many:user_group_users"`
}

type Role struct {
	Model
	Name string `json:"name" gorm:"size:200;uniqueIndex"`
	Desc string `json:"desc" gorm:"column:description"`
	Stat uint   `json:"stat" gorm:"size:32;not null"`

	OpPermissions []*OpPermission `gorm:"many2many:role_op_permissions"`
}

type OpPermission struct {
	Model
	Name      string `json:"name" gorm:"size:200;uniqueIndex"`
	Desc      string `json:"desc" gorm:"column:description"`
	RangeType string `json:"range_type" gorm:"size:255;column:range_type"`
}

type DMSConfig struct {
	Model
	NeedInitOpPermissions bool `json:"need_init_op_permissions" gorm:"column:need_init_op_permissions"`
	NeedInitUsers         bool `json:"need_init_users" gorm:"column:need_init_users"`
	NeedInitRoles         bool `json:"need_init_roles" gorm:"column:need_init_roles"`
	NeedInitProjects      bool `json:"need_init_projects" gorm:"column:need_init_projects"`
}

type Member struct {
	Model
	UserUID          string              `json:"user_uid" gorm:"size:32;column:user_uid;index:project_user_id,unique"`
	ProjectUID       string              `json:"project_uid" gorm:"size:32;column:project_uid;index:project_user_id,unique"`
	RoleWithOpRanges []MemberRoleOpRange `json:"role_with_op_ranges" gorm:"foreignKey:MemberUID;references:UID"`
}

type MemberRoleOpRange struct {
	MemberUID   string `json:"member_uid" gorm:"size:32;column:member_uid"`
	RoleUID     string `json:"role_uid" gorm:"size:32;column:role_uid"`
	OpRangeType string `json:"op_range_type" gorm:"size:255;column:op_range_type"`
	RangeUIDs   string `json:"range_uids" gorm:"type:text;column:range_uids"`
}

func (mg *MemberRoleOpRange) AfterSave(tx *gorm.DB) error {
	return tx.Delete(&MemberRoleOpRange{}, "member_uid IS NULL").Error
}

type MemberGroup struct {
	Model
	Name             string                   `json:"name" gorm:"size:200;index:project_uid_name,unique"`
	ProjectUID       string                   `json:"project_uid" gorm:"size:32;column:project_uid;index:project_uid_name,unique"`
	Users            []*User                  `gorm:"many2many:member_group_users"`
	RoleWithOpRanges []MemberGroupRoleOpRange `json:"role_with_op_ranges" gorm:"foreignKey:MemberGroupUID;references:UID"`
}

type MemberGroupRoleOpRange struct {
	MemberGroupUID string `json:"member_group_uid" gorm:"size:32;column:member_group_uid"`
	RoleUID        string `json:"role_uid" gorm:"size:32;column:role_uid"`
	OpRangeType    string `json:"op_range_type" gorm:"size:255;column:op_range_type"`
	RangeUIDs      string `json:"range_uids" gorm:"type:text;column:range_uids"`
}

func (mg *MemberGroupRoleOpRange) AfterSave(tx *gorm.DB) error {
	return tx.Delete(&MemberGroupRoleOpRange{}, "member_group_uid IS NULL").Error
}

type Project struct {
	Model
	Name          string `json:"name" gorm:"size:200;column:name;index:name,unique"`
	Desc          string `json:"desc" gorm:"column:desc"`
	CreateUserUID string `json:"create_user_uid" gorm:"size:32;column:create_user_uid"`
	Status        string `gorm:"size:64;default:'active'"`
}

type ProxyTarget struct {
	Name            string `json:"name" gorm:"primaryKey;size:200;not null;column:name"`
	Url             string `json:"url" gorm:"size:255;column:url"`
	Version         string `json:"version" gorm:"size:64;column:version"`
	ProxyUrlPrefixs string `json:"proxy_url_prefixs" gorm:"size:255;column:proxy_url_prefixs"`
}

type Plugin struct {
	Name                         string `json:"name" gorm:"primaryKey;size:200;not null;column:name"`
	AddDBServicePreCheckUrl      string `json:"add_db_service_pre_check_url" gorm:"size:255;column:add_db_service_pre_check_url"`
	DelDBServicePreCheckUrl      string `json:"del_db_service_pre_check_url" gorm:"size:255;column:del_db_service_pre_check_url"`
	DelUserPreCheckUrl           string `json:"del_user_pre_check_url" gorm:"size:255;column:del_user_pre_check_url"`
	DelUserGroupPreCheckUrl      string `json:"del_user_group_pre_check_url" gorm:"size:255;column:del_user_group_pre_check_url"`
	OperateDataResourceHandleUrl string `json:"operate_data_resource_handle_url" gorm:"size:255;column:operate_data_resource_handle_url"`
}

// Oauth2Configuration store oauth2 server configuration.
type Oauth2Configuration struct {
	Model
	EnableOauth2    bool   `json:"enable_oauth2" gorm:"column:enable_oauth2"`
	ClientID        string `json:"client_id" gorm:"size:255;column:client_id"`
	ClientKey       string `json:"-" gorm:"-"`
	ClientSecret    string `json:"client_secret" gorm:"size:255;client_secret"`
	ClientHost      string `json:"client_host" gorm:"size:255;column:client_host"`
	ServerAuthUrl   string `json:"server_auth_url" gorm:"size:255;column:server_auth_url"`
	ServerTokenUrl  string `json:"server_token_url" gorm:"size:255;column:server_token_url"`
	ServerUserIdUrl string `json:"server_user_id_url" gorm:"size:255;column:server_user_id_url"`
	Scopes          string `json:"scopes" gorm:"size:255;column:scopes"`
	AccessTokenTag  string `json:"access_token_tag" gorm:"size:255;column:access_token_tag"`
	UserIdTag       string `json:"user_id_tag" gorm:"size:255;column:user_id_tag"`
	LoginTip        string `json:"login_tip" gorm:"size:255;column:login_tip; default:'使用第三方账户登录'"`
}

// LDAPConfiguration store ldap server configuration.
type LDAPConfiguration struct {
	Model
	// whether the ldap is enabled
	Enable bool `json:"enable" gorm:"not null"`
	// whether the ssl is enabled
	EnableSSL bool `json:"enable_ssl" gorm:"not null"`
	// ldap server's ip
	Host string `json:"host" gorm:"size:255;not null"`
	// ldap server's port
	Port string `json:"port" gorm:"size:255;not null"`
	// the DN of the ldap administrative user for verification
	ConnectDn string `json:"connect_dn" gorm:"size:255;not null"`
	// the secret password of the ldap administrative user for verification
	ConnectSecretPassword string `json:"connect_secret_password" gorm:"size:255;not null"`
	// base dn used for ldap verification
	BaseDn string `json:"base_dn" gorm:"size:255;not null"`
	// the key corresponding to the user name in ldap
	UserNameRdnKey string `json:"ldap_user_name_rdn_key" gorm:"size:255;not null"`
	// the key corresponding to the user email in ldap
	UserEmailRdnKey string `json:"ldap_user_email_rdn_key" gorm:"size:255;not null"`
}

// SMTPConfiguration store SMTP server configuration.
type SMTPConfiguration struct {
	Model
	EnableSMTPNotify bool   `json:"enable_smtp_notify" gorm:"default:true"`
	Host             string `json:"smtp_host" gorm:"size:255;column:smtp_host; not null"`
	Port             string `json:"smtp_port" gorm:"size:255;column:smtp_port; not null"`
	Username         string `json:"smtp_username" gorm:"size:255;column:smtp_username; not null"`
	SecretPassword   string `json:"secret_smtp_password" gorm:"size:255;column:secret_smtp_password; not null"`
	IsSkipVerify     bool   `json:"is_skip_verify" gorm:"default:false; not null"`
}

// WeChatConfiguration store WeChat configuration.
type WeChatConfiguration struct {
	Model
	EnableWeChatNotify  bool   `json:"enable_wechat_notify" gorm:"not null"`
	CorpID              string `json:"corp_id" gorm:"size:255;not null"`
	EncryptedCorpSecret string `json:"encrypted_corp_secret" gorm:"size:255;not null"`
	AgentID             int    `json:"agent_id" gorm:"not null"`
	SafeEnabled         bool   `json:"safe_enabled" gorm:"not null"`
	ProxyIP             string `json:"proxy_ip" gorm:"size:255"`
}

type WebHookConfiguration struct {
	Model
	Enable               bool   `json:"enable" gorm:"default:true;not null"`
	MaxRetryTimes        int    `json:"max_retry_times" gorm:"not null"`
	RetryIntervalSeconds int    `json:"retry_interval_seconds" gorm:"not null"`
	EncryptedToken       string `json:"encrypted_token" gorm:"size:255;not null"`
	URL                  string `json:"url" gorm:"size:255;not null"`
}

const (
	ImTypeDingTalk = "dingTalk"
	ImTypeFeishu   = "feishu"
)

// Instant messaging config
type IMConfiguration struct {
	Model
	AppKey      string `json:"app_key" gorm:"size:255;column:app_key"`
	AppSecret   string `json:"app_secret" gorm:"size:255;column:app_secret"`
	IsEnable    bool   `json:"is_enable" gorm:"column:is_enable"`
	ProcessCode string `json:"process_code" gorm:"size:255;column:process_code"`
	// 类型唯一
	Type string `json:"type" gorm:"unique"`
}

type CloudbeaverUserCache struct {
	DMSUserID         string `json:"dms_user_id" gorm:"column:dms_user_id;primaryKey"`
	DMSFingerprint    string `json:"dms_fingerprint" gorm:"size:255;column:dms_fingerprint"`
	CloudbeaverUserID string `json:"cloudbeaver_user_id" gorm:"size:255;column:cloudbeaver_user_id"`
}

type CloudbeaverConnectionCache struct {
	DMSDBServiceID          string `json:"dms_db_service_id" gorm:"column:dms_db_service_id;primaryKey"`
	DMSDBServiceFingerprint string `json:"dms_db_service_fingerprint" gorm:"size:255;column:dms_db_service_fingerprint"`
	CloudbeaverConnectionID string `json:"cloudbeaver_connection_id" gorm:"size:255;column:cloudbeaver_connection_id"`
}

type DatabaseSourceService struct {
	Model
	Name       string `json:"name" gorm:"size:200;not null;index:project_uid_name,unique" example:""`
	Source     string `json:"source" gorm:"size:255;not null"`
	Version    string `json:"version" gorm:"size:255;not null"`
	URL        string `json:"url" gorm:"size:255;not null"`
	DbType     string `json:"db_type" gorm:"size:255;not null"`
	ProjectUID string `json:"project_uid" gorm:"size:32;column:project_uid;index:project_uid_name,unique"`
	// Cron表达式
	CronExpress         string          `json:"cron_express" gorm:"size:255;column:cron_express; not null"`
	LastSyncErr         string          `json:"last_sync_err" gorm:"column:last_sync_err"`
	LastSyncSuccessTime *time.Time      `json:"last_sync_success_time" gorm:"column:last_sync_success_time"`
	ExtraParameters     ExtraParameters `json:"extra_parameters" gorm:"TYPE:json"`
}

type BasicConfig struct {
	Model
	Logo  []byte `json:"logo" gorm:"type:mediumblob"`
	Title string `json:"title" gorm:"size:100;not null;uniqueIndex" example:""`
}

type CompanyNotice struct {
	Model
	NoticeStr   string    `gorm:"type:mediumtext;comment:'企业公告'" json:"notice_str"`
	ReadUserIds ReadUsers `gorm:"type:longtext" json:"read_user_ids"`
}

type ReadUsers []string

func (t *ReadUsers) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

func (t ReadUsers) Value() (driver.Value, error) {
	return json.Marshal(t)
}
