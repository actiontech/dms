package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	"github.com/actiontech/dms/pkg/params"
	"github.com/actiontech/dms/pkg/periods"
	"golang.org/x/text/language"

	"gorm.io/gorm"
)

var AutoMigrateList = []interface{}{
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
	Session{},
	LoginConfiguration{},
	Oauth2Configuration{},
	LDAPConfiguration{},
	SMTPConfiguration{},
	WebHookConfiguration{},
	WeChatConfiguration{},
	IMConfiguration{},
	SmsConfiguration{},
	CloudbeaverUserCache{},
	CloudbeaverConnectionCache{},
	DBServiceSyncTask{},
	BasicConfig{},
	CompanyNotice{},
	ClusterLeader{},
	ClusterNodeInfo{},
	Workflow{},
	WorkflowRecord{},
	WorkflowStep{},
	DataExportTask{},
	DataExportTaskRecord{},
	UserAccessToken{},
	CbOperationLog{},
}

type Model struct {
	UID       string    `json:"uid" gorm:"primaryKey;size:32" example:"1"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" example:"2018-10-21T16:40:23+08:00"`
	UpdatedAt time.Time `json:"updated_at" example:"2018-10-21T16:40:23+08:00"`
}

type DBService struct {
	Model
	Name                   string          `json:"name" gorm:"size:200;not null;index:project_uid_name,unique" example:""`
	DBType                 string          `json:"db_type" gorm:"size:255;column:db_type; not null" example:"mysql"`
	Host                   string          `json:"host" gorm:"column:db_host;size:255; not null" example:"10.10.10.10"`
	Port                   string          `json:"port" gorm:"column:db_port;size:255; not null" example:"3306"`
	User                   string          `json:"user" gorm:"column:db_user;size:255; not null" example:"root"`
	Password               string          `json:"password" gorm:"column:db_password; size:255; not null"`
	Desc                   string          `json:"desc" gorm:"column:desc" example:"this is a instance"`
	Business               string          `json:"business" gorm:"column:business;size:255; not null" example:"this is a business"`
	AdditionalParams       params.Params   `json:"additional_params" gorm:"type:text"`
	Source                 string          `json:"source" gorm:"size:255;not null"`
	ProjectUID             string          `json:"project_uid" gorm:"size:32;column:project_uid;index:project_uid_name,unique"`
	MaintenancePeriod      periods.Periods `json:"maintenance_period" gorm:"type:text"`
	ExtraParameters        ExtraParameters `json:"extra_parameters" gorm:"TYPE:json"`
	IsEnableMasking        bool            `json:"is_enable_masking" gorm:"column:is_enable_masking;type:bool"`
	LastConnectionStatus   *string         `json:"last_connection_status"`
	LastConnectionTime     *time.Time      `json:"last_connection_time"`
	LastConnectionErrorMsg *string         `json:"last_connection_error_msg"`
	EnableBackup           bool            `json:"enable_backup" gorm:"column:enable_backup;type:bool"`
	BackupMaxRows          uint64          `json:"backup_max_rows" gorm:"column:backup_max_rows;not null;default:1000"`
}

type ExtraParameters struct {
	SqleConfig      *SQLEConfig   `json:"sqle_config"`
	AdditionalParam params.Params `json:"additional_param"`
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
	ThirdPartyUserID       string         `json:"third_party_user_id" gorm:"size:255;column:third_party_user_id"`      // used to retrieve sqle user based on third-party user ID
	ThirdPartyUserInfo     string         `json:"third_party_user_info" gorm:"type:text;column:third_party_user_info"` // used to save original third-party user information
	ThirdPartyIdToken      string         `json:"third_party_id_token" gorm:"type:text;column:third_party_id_token"`   // used to call OIDC Logout
	Email                  string         `json:"email" gorm:"size:255;column:email"`
	Phone                  string         `json:"phone" gorm:"size:255;column:phone"`
	WeChatID               string         `json:"wechat_id" gorm:"size:255;column:wechat_id"`
	Language               string         `json:"language" gorm:"size:255;column:language"`
	Password               string         `json:"password" gorm:"size:255;column:password"`
	UserAuthenticationType string         `json:"user_authentication_type" gorm:"size:255;not null;column:user_authentication_type"`
	Stat                   uint           `json:"stat" gorm:"not null"`
	TwoFactorEnabled       bool           `json:"two_factor_enabled" gorm:"default:false; not null"`
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

type UserAccessToken struct {
	Model
	Token       string    `json:"token" gorm:"size:255"`
	ExpiredTime time.Time `json:"expired_time" example:"2018-10-21T16:40:23+08:00"`
	UserID      uint      `json:"user_id" gorm:"size:32;index:user_id,unique"`

	User *User `json:"user" gorm:"foreignkey:user_id"`
}

type DMSConfig struct {
	Model
	NeedInitOpPermissions          bool `json:"need_init_op_permissions" gorm:"column:need_init_op_permissions"`
	NeedInitUsers                  bool `json:"need_init_users" gorm:"column:need_init_users"`
	NeedInitRoles                  bool `json:"need_init_roles" gorm:"column:need_init_roles"`
	NeedInitProjects               bool `json:"need_init_projects" gorm:"column:need_init_projects"`
	EnableSQLResultSetsDataMasking bool `json:"enable_sql_result_sets_data_masking" gorm:"column:enable_sql_result_sets_data_masking"`
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
	Name            string `json:"name" gorm:"size:200;column:name;index:name,unique"`
	Desc            string `json:"desc" gorm:"column:desc"`
	Business        Bus    `json:"business" gorm:"type:json"`
	IsFixedBusiness bool   `json:"is_fixed_business" gorm:"not null"`
	CreateUserUID   string `json:"create_user_uid" gorm:"size:32;column:create_user_uid"`
	Status          string `gorm:"size:64;default:'active'"`
	Priority        uint8  `json:"priority" gorm:"type:tinyint unsigned;not null;default:20;comment:'优先级：10=低, 20=中, 30=高'"`
}

const (
	ProxyScenarioInternalService     string = "internal_service"
	ProxyScenarioThirdPartyIntegrate string = "thrid_party_integrate"
)

type ProxyTarget struct {
	Name            string `json:"name" gorm:"primaryKey;size:200;not null;column:name"`
	Url             string `json:"url" gorm:"size:255;column:url"`
	Version         string `json:"version" gorm:"size:64;column:version"`
	ProxyUrlPrefixs string `json:"proxy_url_prefixs" gorm:"size:255;column:proxy_url_prefixs"`
	Scenario        string `json:"scenario" gorm:"size:64;column:scenario;default:'internal_service'"`
}

type Plugin struct {
	Name                         string `json:"name" gorm:"primaryKey;size:200;not null;column:name"`
	AddDBServicePreCheckUrl      string `json:"add_db_service_pre_check_url" gorm:"size:255;column:add_db_service_pre_check_url"`
	DelDBServicePreCheckUrl      string `json:"del_db_service_pre_check_url" gorm:"size:255;column:del_db_service_pre_check_url"`
	DelUserPreCheckUrl           string `json:"del_user_pre_check_url" gorm:"size:255;column:del_user_pre_check_url"`
	DelUserGroupPreCheckUrl      string `json:"del_user_group_pre_check_url" gorm:"size:255;column:del_user_group_pre_check_url"`
	OperateDataResourceHandleUrl string `json:"operate_data_resource_handle_url" gorm:"size:255;column:operate_data_resource_handle_url"`
	GetDatabaseDriverOptionsUrl  string `json:"get_database_driver_options_url" gorm:"size:255;column:get_database_driver_options_url"`
	GetDatabaseDriverLogosUrl    string `json:"get_database_driver_logos_url" gorm:"size:255;column:get_database_driver_logos_url"`
}

type Session struct {
	Model
	UserUID               string         `json:"user_uid" gorm:"size:32;column:user_uid;index:idx_user_uid"`
	OAuth2Sub             string         `json:"oauth2_sub" gorm:"size:255;column:oauth2_sub;index:idx_oauth2_sub_sid,unique"`
	OAuth2Sid             string         `json:"oauth2_sid" gorm:"size:255;column:oauth2_sid;index:idx_oauth2_sub_sid,unique"`
	OAuth2IdToken         string         `json:"oauth2_id_token" gorm:"type:text;column:oauth2_id_token"`
	OAuth2RefreshToken    string         `json:"oauth2_refresh_token" gorm:"type:text;column:oauth2_refresh_token"`
	OAuth2LastLogoutEvent sql.NullString `json:"oauth2_last_logout_event" gorm:"size:255;column:oauth2_last_logout_event;"`
}

// LoginConfiguration store local login configuration.
type LoginConfiguration struct {
	Model
	LoginButtonText     string `json:"login_button_text" gorm:"column:login_button_text;size:255;default:'登录';not null"`
	DisableUserPwdLogin bool   `json:"disable_user_pwd_login" gorm:"column:disable_user_pwd_login;default:false;not null"`
}

// Oauth2Configuration store oauth2 server configuration.
type Oauth2Configuration struct {
	Model
	EnableOauth2         bool   `json:"enable_oauth2" gorm:"column:enable_oauth2"`
	SkipCheckState       bool   `json:"skip_check_state" gorm:"column:skip_check_state"`
	AutoCreateUser       bool   `json:"auto_create_user" gorm:"auto_create_user"`
	AutoCreateUserPWD    string `json:"-" gorm:"-"`
	AutoCreateUserSecret string `json:"auto_create_user_pwd" gorm:"size:255;column:auto_create_user_pwd"`
	ClientID             string `json:"client_id" gorm:"size:255;column:client_id"`
	ClientKey            string `json:"-" gorm:"-"`
	ClientSecret         string `json:"client_secret" gorm:"size:255;client_secret"`
	ClientHost           string `json:"client_host" gorm:"size:255;column:client_host"`
	ServerAuthUrl        string `json:"server_auth_url" gorm:"size:255;column:server_auth_url"`
	ServerTokenUrl       string `json:"server_token_url" gorm:"size:255;column:server_token_url"`
	ServerUserIdUrl      string `json:"server_user_id_url" gorm:"size:255;column:server_user_id_url"`
	ServerLogoutUrl      string `json:"server_logout_url" gorm:"size:255;column:server_logout_url"`
	Scopes               string `json:"scopes" gorm:"size:255;column:scopes"`
	AccessTokenTag       string `json:"access_token_tag" gorm:"size:255;column:access_token_tag"`
	UserIdTag            string `json:"user_id_tag" gorm:"size:255;column:user_id_tag"`
	UserWeChatTag        string `json:"user_wechat_tag" gorm:"size:255;column:user_wechat_tag"`
	UserEmailTag         string `json:"user_email_tag" gorm:"size:255;column:user_email_tag"`
	LoginPermExpr        string `json:"login_perm_expr" gorm:"size:255;column:login_perm_expr"`
	LoginTip             string `json:"login_tip" gorm:"size:255;column:login_tip; default:'使用第三方账户登录'"`
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

type JSON json.RawMessage

type SmsConfiguration struct {
	Model
	Enable        bool   `json:"enable" gorm:"default:true;not null"`
	Type          string `json:"type" gorm:"size:255;not null"`
	Url           string `json:"url"  gorm:"size:255;not null"`
	Configuration JSON   `json:"configuration" gorm:"type:json"`
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
	Type string `json:"type" gorm:"index:unique,size:255"`
}

type CloudbeaverUserCache struct {
	DMSUserID         string `json:"dms_user_id" gorm:"column:dms_user_id;primaryKey"`
	DMSFingerprint    string `json:"dms_fingerprint" gorm:"size:255;column:dms_fingerprint"`
	CloudbeaverUserID string `json:"cloudbeaver_user_id" gorm:"size:255;column:cloudbeaver_user_id"`
}

type CloudbeaverConnectionCache struct {
	DMSDBServiceID          string `json:"dms_db_service_id" gorm:"column:dms_db_service_id;primaryKey"`
	DMSUserID               string `json:"dms_user_id" gorm:"column:dms_user_id;primaryKey"`
	DMSDBServiceFingerprint string `json:"dms_db_service_fingerprint" gorm:"size:255;column:dms_db_service_fingerprint"`
	CloudbeaverConnectionID string `json:"cloudbeaver_connection_id" gorm:"size:255;column:cloudbeaver_connection_id"`
	Purpose                 string `json:"purpose" gorm:"size:255;column:purpose;primaryKey"`
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

type ClusterLeader struct {
	Anchor       int       `gorm:"primary_key"` // 常量值，保证该表仅有一行不重复记录。无其他意义。
	ServerId     string    `gorm:"not null"`
	LastSeenTime time.Time `gorm:"not null"`
}

type ClusterNodeInfo struct {
	ServerId     string    `json:"server_id" gorm:"primary_key"`
	HardwareSign string    `json:"hardware_sign" gorm:"type:varchar(3000)"`
	CreatedAt    time.Time `json:"created_at" gorm:"<-:create" example:"2018-10-21T16:40:23+08:00"`
	UpdatedAt    time.Time `json:"updated_at" example:"2018-10-21T16:40:23+08:00"`
}
type Workflow struct {
	Model
	Name              string     `json:"name" gorm:"size:255;not null;index:project_uid_name,unique" example:""`
	ProjectUID        string     `json:"project_uid" gorm:"size:32;column:project_uid;index:project_uid_name,unique"`
	WorkflowType      string     `json:"workflow_type" gorm:"size:64;column:workflow_type; not null" example:"export"`
	Desc              string     `json:"desc" gorm:"column:desc" example:"this is a data transform export workflow"`
	CreateTime        *time.Time `json:"create_time" gorm:"column:create_time"`
	CreateUserUID     string     `json:"create_user_uid" gorm:"size:32;column:create_user_uid"`
	WorkflowRecordUid string     `json:"workflow_record_uid" gorm:"size:32;column:workflow_record_uid"`

	WorkflowRecord *WorkflowRecord `gorm:"foreignkey:WorkflowUid"`
}

type WorkflowRecord struct {
	Model
	WorkflowUid           string  `json:"workflow_uid" gorm:"size:32" `
	CurrentWorkflowStepId uint64  `json:"current_workflow_step_id"`
	Status                string  `gorm:"default:\"wait_for_export\""`
	TaskIds               Strings `json:"task_ids" gorm:"type:json"`

	Steps []*WorkflowStep `gorm:"foreignkey:WorkflowRecordUid"`
}

type Strings []string

func (t *Strings) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

func (t Strings) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type Bus []Business

type Business struct {
	Uid  string
	Name string
}

func (b *Bus) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, b)
}

func (b Bus) Value() (driver.Value, error) {
	return json.Marshal(b)
}

type WorkflowStep struct {
	StepId            uint64     `json:"step_id" gorm:"index:step_record_id,unique"`
	WorkflowRecordUid string     `gorm:"index; not null;index:step_record_id,unique"`
	OperationUserUid  string     `json:"operation_user_uid" gorm:"size:32"`
	OperateAt         *time.Time `json:"operate_at"`
	State             string     `gorm:"size:32"`
	Reason            string     `json:"reason" gorm:"size:255"`
	Assignees         Strings    `json:"assignees" gorm:"type:json"`
}

type DataExportTask struct {
	Model
	DBServiceUid      string     `json:"db_service_uid" gorm:"size:32"`
	DatabaseName      string     `json:"database_name" gorm:"size:32"`
	WorkFlowRecordUid string     `json:"workflow_record_uid" gorm:"size:255"`
	ExportType        string     `json:"export_type" gorm:"size:32"`
	ExportFileType    string     `json:"export_file_type" gorm:"size:32"`
	ExportFileName    string     `json:"export_file_name" gorm:"column:export_file_name;size:255"`
	ExportStatus      string     `json:"export_status" gorm:"column:export_status;size:32"`
	ExportStartTime   *time.Time `json:"export_start_time" gorm:"column:export_start_time"`
	ExportEndTime     *time.Time `json:"export_end_time" gorm:"column:export_end_time"`
	CreateUserUID     string     `json:"create_user_uid" gorm:"size:32;column:create_user_uid"`
	// Audit Result
	AuditPassRate float64 `json:"audit_pass_rate"`
	AuditScore    int32   `json:"audit_score"`
	AuditLevel    string  `json:"audit_level"  gorm:"size:32"`

	DataExportTaskRecords []*DataExportTaskRecord `gorm:"foreignkey:DataExportTaskId"`
}

type DataExportTaskRecord struct {
	Number           uint   `json:"number" gorm:"index:task_id_number,unique"`
	DataExportTaskId string `json:"data_export_task_id" gorm:"size:32;column:data_export_task_id;index:task_id_number,unique"`
	ExportSQL        string `json:"export_sql"`
	ExportSQLType    string `json:"export_sql_type" gorm:"column:export_sql_type;size:10"`
	ExportResult     string `json:"export_result"`
	ExportStatus     string `json:"export_status" gorm:"size:32"`

	AuditLevel   string       `json:"audit_level"`
	AuditResults AuditResults `json:"audit_results" gorm:"type:json"`
}

type AuditResult struct {
	Level               string              `json:"level"`
	RuleName            string              `json:"rule_name"`
	ExecutionFailed     bool                `json:"execution_failed"`
	I18nAuditResultInfo I18nAuditResultInfo `json:"i18n_audit_result_info"`
}

type RuleLevel string

const (
	RuleLevelNull   RuleLevel = "" // used to indicate no rank
	RuleLevelNormal RuleLevel = "normal"
	RuleLevelNotice RuleLevel = "notice"
	RuleLevelWarn   RuleLevel = "warn"
	RuleLevelError  RuleLevel = "error"
)

var ruleLevelMap = map[RuleLevel]int{
	RuleLevelNull:   -1,
	RuleLevelNormal: 0,
	RuleLevelNotice: 1,
	RuleLevelWarn:   2,
	RuleLevelError:  3,
}

func (r RuleLevel) LessOrEqual(l RuleLevel) bool {
	return ruleLevelMap[r] <= ruleLevelMap[l]
}

func (ar *AuditResult) GetAuditMsgByLangTag(lang language.Tag) string {
	return ar.I18nAuditResultInfo.GetAuditResultInfoByLangTag(lang).Message
}

func (ar *AuditResult) GetAuditErrorMsgByLangTag(lang language.Tag) string {
	return ar.I18nAuditResultInfo.GetAuditResultInfoByLangTag(lang).ErrorInfo
}

type AuditResultInfo struct {
	Message   string
	ErrorInfo string
}

type I18nAuditResultInfo map[language.Tag]AuditResultInfo

func (i *I18nAuditResultInfo) GetAuditResultInfoByLangTag(lang language.Tag) *AuditResultInfo {
	if i == nil {
		return &AuditResultInfo{}
	}

	if info, ok := (*i)[lang]; ok {
		return &info
	}

	info := (*i)[i18nPkg.DefaultLang]
	return &info
}

type AuditResults []AuditResult

func (a AuditResults) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *AuditResults) Scan(input interface{}) error {
	if input == nil {
		return nil
	}
	if data, ok := input.([]byte); !ok {
		return fmt.Errorf("AuditResults Scan input is not bytes")
	} else {
		return json.Unmarshal(data, a)
	}
}

func (a *AuditResults) String() string {
	msgs := make([]string, len(*a))
	for i := range *a {
		res := (*a)[i]
		msg := fmt.Sprintf("[%s]%s", res.Level, res.GetAuditMsgByLangTag(i18nPkg.DefaultLang)) // todo i18n other lang?
		msgs[i] = msg
	}
	return strings.Join(msgs, "\n")
}

func (a *AuditResults) Append(nar *AuditResult) {
	for i := range *a {
		ar := (*a)[i]
		if ar.Level == nar.Level && ar.RuleName == nar.RuleName && ar.GetAuditMsgByLangTag(i18nPkg.DefaultLang) == nar.GetAuditMsgByLangTag(i18nPkg.DefaultLang) {
			return
		}
	}
	*a = append(*a, AuditResult{Level: nar.Level, RuleName: nar.RuleName, I18nAuditResultInfo: nar.I18nAuditResultInfo})
}

type CbOperationLog struct {
	Model
	OpPersonUID       string          `json:"op_person_uid" gorm:"size:32"`
	OpTime            *time.Time      `json:"op_time"`
	DBServiceUID      string          `json:"db_service_uid" gorm:"size:32"`
	OpType            string          `json:"op_type" gorm:"size:255"`
	I18nOpDetail      i18nPkg.I18nStr `json:"i18n_op_detail" gorm:"column:i18n_op_detail; type:json"`
	OpSessionID       *string         `json:"op_session_id" gorm:"size:255"`
	ProjectID         string          `json:"project_id" gorm:"size:32"`
	OpHost            string          `json:"op_host" gorm:"size:255"`
	AuditResult       AuditResults    `json:"audit_result" gorm:"type:json"`
	IsAuditPassed     *bool           `json:"is_audit_passed"`
	ExecResult        string          `json:"exec_result" gorm:"type:text"`
	ExecTotalSec      int64           `json:"exec_total_sec"`
	ResultSetRowCount int64           `json:"result_set_row_count"`

	User      *User      `json:"user" gorm:"foreignKey:OpPersonUID"`
	DbService *DBService `json:"db_service" gorm:"foreignKey:DBServiceUID"`
	Project   *Project   `json:"project" gorm:"foreignKey:ProjectID"`
}

type DBServiceSyncTask struct {
	Model
	Name                string          `json:"name" gorm:"size:200;not null;unique" example:""`
	Source              string          `json:"source" gorm:"size:255;not null"`
	URL                 string          `json:"url" gorm:"size:255;not null"`
	DbType              string          `json:"db_type" gorm:"size:255;not null"`
	CronExpress         string          `json:"cron_express" gorm:"size:255;column:cron_express; not null"`
	LastSyncErr         string          `json:"last_sync_err" gorm:"column:last_sync_err"`
	LastSyncSuccessTime *time.Time      `json:"last_sync_success_time" gorm:"column:last_sync_success_time"`
	ExtraParameters     ExtraParameters `json:"extra_parameters" gorm:"TYPE:json"`
}
