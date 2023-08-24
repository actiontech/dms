package storage

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/storage/model"

	pkgAes "github.com/actiontech/dms/pkg/dms-common/pkg/aes"

	"github.com/labstack/echo/v4"
)

func convertBase(o model.Model) biz.Base {
	return biz.Base{
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func convertBizDBService(ds *biz.DBService) (*model.DBService, error) {
	encrypted, err := pkgAes.AesEncrypt(ds.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}
	dbService := &model.DBService{
		Model: model.Model{
			UID: ds.UID,
		},
		Name:              ds.Name,
		Desc:              ds.Desc,
		DBType:            ds.DBType.String(),
		Host:              ds.Host,
		Port:              ds.Port,
		User:              ds.AdminUser,
		Password:          encrypted,
		Business:          ds.Business,
		AdditionalParams:  ds.AdditionalParams,
		Source:            ds.Source,
		MaintenancePeriod: ds.MaintenancePeriod,
		NamespaceUID:      ds.NamespaceUID,
	}
	{
		// add sqle config
		if ds.SQLEConfig != nil {
			dbService.ExtraParameters = model.ExtraParameters{
				SqleConfig: &model.SQLEConfig{
					RuleTemplateName: ds.SQLEConfig.RuleTemplateName,
					RuleTemplateID:   ds.SQLEConfig.RuleTemplateID,
				},
			}
			sqleQueryConfig := ds.SQLEConfig.SQLQueryConfig
			if sqleQueryConfig != nil {
				dbService.ExtraParameters.SqleConfig.SqlQueryConfig = &model.SqlQueryConfig{
					AllowQueryWhenLessThanAuditLevel: sqleQueryConfig.AllowQueryWhenLessThanAuditLevel,
					AuditEnabled:                     sqleQueryConfig.AuditEnabled,
					MaxPreQueryRows:                  sqleQueryConfig.MaxPreQueryRows,
					QueryTimeoutSecond:               sqleQueryConfig.QueryTimeoutSecond,
				}
			}
		}
	}
	return dbService, nil
}

func convertModelDBService(ds *model.DBService) (*biz.DBService, error) {
	decrypted, err := pkgAes.AesDecrypt(ds.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %v", err)
	}
	dbType, err := pkgConst.ParseDBType(ds.DBType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db type: %v", err)
	}
	dbService := &biz.DBService{
		Base:                   convertBase(ds.Model),
		UID:                    ds.UID,
		Name:                   ds.Name,
		Desc:                   ds.Desc,
		DBType:                 dbType,
		Host:                   ds.Host,
		Port:                   ds.Port,
		AdminUser:              ds.User,
		AdminPassword:          decrypted,
		EncryptedAdminPassword: ds.Password,
		MaintenancePeriod:      ds.MaintenancePeriod,
		Business:               ds.Business,
		AdditionalParams:       ds.AdditionalParams,
		Source:                 ds.Source,
		NamespaceUID:           ds.NamespaceUID,
	}
	{
		modelSqleConfig := ds.ExtraParameters.SqleConfig
		if modelSqleConfig != nil {
			dbService.SQLEConfig = &biz.SQLEConfig{}
			dbService.SQLEConfig.RuleTemplateName = modelSqleConfig.RuleTemplateName
			dbService.SQLEConfig.RuleTemplateID = modelSqleConfig.RuleTemplateID
			sqleQueryConfig := modelSqleConfig.SqlQueryConfig
			if sqleQueryConfig != nil {
				sqc := &biz.SQLQueryConfig{
					AllowQueryWhenLessThanAuditLevel: sqleQueryConfig.AllowQueryWhenLessThanAuditLevel,
					AuditEnabled:                     sqleQueryConfig.AuditEnabled,
					MaxPreQueryRows:                  sqleQueryConfig.MaxPreQueryRows,
					QueryTimeoutSecond:               sqleQueryConfig.QueryTimeoutSecond,
				}
				dbService.SQLEConfig.SQLQueryConfig = sqc
			}
		}
	}
	return dbService, nil
}

func convertModelDatabaseSourceService(m *model.DatabaseSourceService) (*biz.DatabaseSourceServiceParams, error) {
	dbType, err := pkgConst.ParseDBType(m.DbType)
	if err != nil {
		return nil, err
	}

	ret := &biz.DatabaseSourceServiceParams{
		UID:          m.UID,
		Name:         m.Name,
		Source:       m.Source,
		Version:      m.Version,
		URL:          m.URL,
		DbType:       dbType,
		CronExpress:  m.CronExpress,
		NamespaceUID: m.NamespaceUID,
		LastSyncErr:  m.LastSyncErr,
		UpdatedAt:    m.UpdatedAt,
	}

	if m.LastSyncSuccessTime != nil {
		ret.LastSyncSuccessTime = m.LastSyncSuccessTime
	}

	modelSqleConfig := m.ExtraParameters.SqleConfig
	if modelSqleConfig != nil {
		ret.SQLEConfig = &biz.SQLEConfig{
			RuleTemplateID:   modelSqleConfig.RuleTemplateID,
			RuleTemplateName: modelSqleConfig.RuleTemplateName,
		}

		sqleQueryConfig := modelSqleConfig.SqlQueryConfig
		if sqleQueryConfig != nil {
			ret.SQLEConfig.SQLQueryConfig = &biz.SQLQueryConfig{
				AllowQueryWhenLessThanAuditLevel: sqleQueryConfig.AllowQueryWhenLessThanAuditLevel,
				AuditEnabled:                     sqleQueryConfig.AuditEnabled,
				MaxPreQueryRows:                  sqleQueryConfig.MaxPreQueryRows,
				QueryTimeoutSecond:               sqleQueryConfig.QueryTimeoutSecond,
			}
		}
	}

	return ret, nil
}

func convertBizUser(u *biz.User) (*model.User, error) {
	encrypted, err := pkgAes.AesEncrypt(u.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %v", err)
	}

	var lastLoginAt *time.Time
	if !u.LastLoginAt.IsZero() {
		lastLoginAt = &u.LastLoginAt
	}

	return &model.User{
		Model: model.Model{
			UID: u.UID,
		},
		Name:                   u.Name,
		ThirdPartyUserID:       u.ThirdPartyUserID,
		Password:               encrypted,
		Email:                  u.Email,
		Phone:                  u.Phone,
		WeChatID:               u.WxID,
		UserAuthenticationType: u.UserAuthenticationType.String(),
		Stat:                   u.Stat.Uint(),
		LastLoginAt:            lastLoginAt,
	}, nil
}

func convertBizCloudbeaverUser(u *biz.CloudbeaverUser) *model.CloudbeaverUserCache {
	return &model.CloudbeaverUserCache{
		DMSUserID:         u.DMSUserID,
		DMSFingerprint:    u.DMSFingerprint,
		CloudbeaverUserID: u.CloudbeaverUserID,
	}
}

func convertBizCloudbeaverConnection(u *biz.CloudbeaverConnection) *model.CloudbeaverConnectionCache {
	return &model.CloudbeaverConnectionCache{
		DMSDBServiceID:          u.DMSDBServiceID,
		DMSDBServiceFingerprint: u.DMSDBServiceFingerprint,
		CloudbeaverConnectionID: u.CloudbeaverConnectionID,
	}
}

func convertBizDatabaseSourceService(u *biz.DatabaseSourceServiceParams) *model.DatabaseSourceService {
	m := &model.DatabaseSourceService{
		Model:        model.Model{UID: u.UID},
		Name:         u.Name,
		Source:       u.Source,
		Version:      u.Version,
		URL:          u.URL,
		DbType:       u.DbType.String(),
		CronExpress:  u.CronExpress,
		NamespaceUID: u.NamespaceUID,
	}

	// add sqle config
	if u.SQLEConfig != nil {
		m.ExtraParameters = model.ExtraParameters{
			SqleConfig: &model.SQLEConfig{
				RuleTemplateName: u.SQLEConfig.RuleTemplateName,
				RuleTemplateID:   u.SQLEConfig.RuleTemplateID,
			},
		}
		sqleQueryConfig := u.SQLEConfig.SQLQueryConfig
		if sqleQueryConfig != nil {
			m.ExtraParameters.SqleConfig.SqlQueryConfig = &model.SqlQueryConfig{
				AllowQueryWhenLessThanAuditLevel: sqleQueryConfig.AllowQueryWhenLessThanAuditLevel,
				AuditEnabled:                     sqleQueryConfig.AuditEnabled,
				MaxPreQueryRows:                  sqleQueryConfig.MaxPreQueryRows,
				QueryTimeoutSecond:               sqleQueryConfig.QueryTimeoutSecond,
			}
		}
	}

	return m
}

func convertModelUser(u *model.User) (*biz.User, error) {
	decrypted, err := pkgAes.AesDecrypt(u.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %v", err)
	}
	stat, err := biz.ParseUserStat(u.Stat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user stat: %v", err)
	}
	typ, err := biz.ParseUserAuthenticationType(u.UserAuthenticationType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user authentication type: %v", err)
	}

	var lastLoginAt time.Time
	if u.LastLoginAt != nil {
		lastLoginAt = *u.LastLoginAt
	}

	return &biz.User{
		Base:                   convertBase(u.Model),
		UID:                    u.UID,
		ThirdPartyUserID:       u.ThirdPartyUserID,
		Name:                   u.Name,
		Email:                  u.Email,
		Phone:                  u.Phone,
		WxID:                   u.WeChatID,
		UserAuthenticationType: typ,
		Stat:                   stat,
		LastLoginAt:            lastLoginAt,
		Password:               decrypted,
		Deleted:                u.DeletedAt.Valid,
	}, nil
}

func convertModelCloudbeaverUser(u *model.CloudbeaverUserCache) *biz.CloudbeaverUser {
	return &biz.CloudbeaverUser{
		DMSUserID:         u.DMSUserID,
		DMSFingerprint:    u.DMSFingerprint,
		CloudbeaverUserID: u.CloudbeaverUserID,
	}
}

func convertModelCloudbeaverConnection(items []*model.CloudbeaverConnectionCache) []*biz.CloudbeaverConnection {
	res := make([]*biz.CloudbeaverConnection, 0, len(items))
	for _, item := range items {
		res = append(res, &biz.CloudbeaverConnection{
			DMSDBServiceID:          item.DMSDBServiceID,
			DMSDBServiceFingerprint: item.DMSDBServiceFingerprint,
			CloudbeaverConnectionID: item.CloudbeaverConnectionID,
		})
	}

	return res
}

func convertBizUserGroup(u *biz.UserGroup) (*model.UserGroup, error) {

	return &model.UserGroup{
		Model: model.Model{
			UID: u.UID,
		},
		Name: u.Name,
		Desc: u.Desc,
		Stat: u.Stat.Uint(),
	}, nil
}

func convertModelUserGroup(u *model.UserGroup) (*biz.UserGroup, error) {
	stat, err := biz.ParseUserGroupStat(u.Stat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user group stat: %v", err)
	}
	return &biz.UserGroup{
		Base: convertBase(u.Model),
		UID:  u.UID,
		Name: u.Name,
		Desc: u.Desc,
		Stat: stat,
	}, nil
}

func convertBizRole(u *biz.Role) (*model.Role, error) {

	return &model.Role{
		Model: model.Model{
			UID: u.UID,
		},
		Name: u.Name,
		Desc: u.Desc,
		Stat: u.Stat.Uint(),
	}, nil
}

func convertModelRole(u *model.Role) (*biz.Role, error) {
	stat, err := biz.ParseRoleStat(u.Stat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user stat: %v", err)
	}
	return &biz.Role{
		Base: convertBase(u.Model),
		UID:  u.UID,
		Name: u.Name,
		Desc: u.Desc,
		Stat: stat,
	}, nil
}

func convertBizOpPermission(u *biz.OpPermission) (*model.OpPermission, error) {

	return &model.OpPermission{
		Model: model.Model{
			UID: u.UID,
		},
		Name:      u.Name,
		Desc:      u.Desc,
		RangeType: u.RangeType.String(),
	}, nil
}

func convertModelOpPermission(u *model.OpPermission) (*biz.OpPermission, error) {
	return &biz.OpPermission{
		Base:      convertBase(u.Model),
		UID:       u.UID,
		Name:      u.Name,
		Desc:      u.Desc,
		RangeType: biz.OpRangeType(u.RangeType),
	}, nil
}

func convertBizDMSConfig(u *biz.DMSConfig) (*model.DMSConfig, error) {

	return &model.DMSConfig{
		Model: model.Model{
			UID: u.UID,
		},
		NeedInitOpPermissions: u.NeedInitOpPermissions,
		NeedInitUsers:         u.NeedInitUsers,
		NeedInitRoles:         u.NeedInitRoles,
		NeedInitNamespaces:    u.NeedInitNamespaces,
	}, nil
}

func convertModelDMSConfig(u *model.DMSConfig) (*biz.DMSConfig, error) {
	return &biz.DMSConfig{
		Base:                  convertBase(u.Model),
		UID:                   u.UID,
		NeedInitOpPermissions: u.NeedInitOpPermissions,
		NeedInitUsers:         u.NeedInitUsers,
		NeedInitRoles:         u.NeedInitRoles,
		NeedInitNamespaces:    u.NeedInitNamespaces,
	}, nil
}

func convertBizMember(m *biz.Member) (*model.Member, error) {
	roles := make([]model.MemberRoleOpRange, 0, len(m.RoleWithOpRanges))
	for _, p := range m.RoleWithOpRanges {

		roles = append(roles, model.MemberRoleOpRange{
			MemberUID:   m.UID,
			RoleUID:     p.RoleUID,
			OpRangeType: p.OpRangeType.String(),
			RangeUIDs:   convertBizRangeUIDs(p.RangeUIDs),
		})
	}
	return &model.Member{
		Model: model.Model{
			UID: m.UID,
		},
		UserUID:          m.UserUID,
		NamespaceUID:     m.NamespaceUID,
		RoleWithOpRanges: roles,
	}, nil
}

func convertModelMember(m *model.Member) (*biz.Member, error) {
	roles := make([]biz.MemberRoleWithOpRange, 0, len(m.RoleWithOpRanges))
	for _, p := range m.RoleWithOpRanges {
		typ, err := biz.ParseOpRangeType(p.OpRangeType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse op range type: %v", err)
		}
		roles = append(roles, biz.MemberRoleWithOpRange{
			RoleUID:     p.RoleUID,
			OpRangeType: typ,
			RangeUIDs:   convertModelRangeUIDs(p.RangeUIDs),
		})
	}
	return &biz.Member{
		Base:             convertBase(m.Model),
		UID:              m.UID,
		NamespaceUID:     m.NamespaceUID,
		UserUID:          m.UserUID,
		RoleWithOpRanges: roles,
	}, nil
}

func convertBizRangeUIDs(uids []string) string {
	return strings.Join(uids, ",")
}

func convertModelRangeUIDs(uids string) []string {
	return strings.Split(uids, ",")
}

func convertBizNamespace(m *biz.Namespace) (*model.Namespace, error) {
	return &model.Namespace{
		Model: model.Model{
			UID: m.UID,
		},
		Name:          m.Name,
		Desc:          m.Desc,
		Status:        string(m.Status),
		CreateUserUID: m.CreateUserUID,
	}, nil
}

func convertModelNamespace(m *model.Namespace) (*biz.Namespace, error) {
	return &biz.Namespace{
		Base:          convertBase(m.Model),
		UID:           m.UID,
		Name:          m.Name,
		Desc:          m.Desc,
		Status:        convertModelNamespaceStatus(m.Status),
		CreateUserUID: m.CreateUserUID,
		CreateTime:    m.CreatedAt,
	}, nil
}

func convertModelNamespaceStatus(status string) biz.NamespaceStatus {
	switch status {
	case string(biz.NamespaceStatusActive):
		return biz.NamespaceStatusActive
	case string(biz.NamespaceStatusArchived):
		return biz.NamespaceStatusArchived
	default:
		return biz.NamespaceStatusUnknown
	}
}

func convertBizProxyTarget(t *biz.ProxyTarget) (*model.ProxyTarget, error) {
	return &model.ProxyTarget{
		Name:            t.Name,
		Url:             t.URL.String(),
		ProxyUrlPrefixs: strings.Join(t.GetProxyUrlPrefixs(), ";"),
	}, nil
}

func convertModelProxyTarget(t *model.ProxyTarget) (*biz.ProxyTarget, error) {
	url, err := url.ParseRequestURI(t.Url)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %s", t.Url)
	}
	p := &biz.ProxyTarget{
		Name: t.Name,
		URL:  url,
		Meta: echo.Map{},
	}
	p.SetProxyUrlPrefix(strings.Split(t.ProxyUrlPrefixs, ";"))
	return p, nil
}

func convertBizPlugin(t *biz.Plugin) (*model.Plugin, error) {
	return &model.Plugin{
		Name:                         t.Name,
		AddDBServicePreCheckUrl:      t.AddDBServicePreCheckUrl,
		DelDBServicePreCheckUrl:      t.DelDBServicePreCheckUrl,
		DelUserPreCheckUrl:           t.DelUserPreCheckUrl,
		DelUserGroupPreCheckUrl:      t.DelUserGroupPreCheckUrl,
		OperateDataResourceHandleUrl: t.OperateDataResourceHandleUrl,
	}, nil
}

func convertModelPlugin(t *model.Plugin) (*biz.Plugin, error) {
	p := &biz.Plugin{
		Name:                         t.Name,
		AddDBServicePreCheckUrl:      t.AddDBServicePreCheckUrl,
		DelDBServicePreCheckUrl:      t.DelDBServicePreCheckUrl,
		DelUserPreCheckUrl:           t.DelUserPreCheckUrl,
		DelUserGroupPreCheckUrl:      t.DelUserGroupPreCheckUrl,
		OperateDataResourceHandleUrl: t.OperateDataResourceHandleUrl,
	}
	return p, nil
}

func convertBizOauth2Configuration(b *biz.Oauth2Configuration) (*model.Oauth2Configuration, error) {
	data, err := pkgAes.AesEncrypt(b.ClientKey)
	if err != nil {
		return nil, err
	}
	b.ClientSecret = data
	return &model.Oauth2Configuration{
		Model: model.Model{
			UID: b.UID,
		},
		EnableOauth2:    b.EnableOauth2,
		ClientID:        b.ClientID,
		ClientKey:       b.ClientKey,
		ClientSecret:    b.ClientSecret,
		ClientHost:      b.ClientHost,
		ServerAuthUrl:   b.ServerAuthUrl,
		ServerTokenUrl:  b.ServerTokenUrl,
		ServerUserIdUrl: b.ServerUserIdUrl,
		Scopes:          strings.Join(b.Scopes, ","),
		AccessTokenTag:  b.AccessTokenTag,
		UserIdTag:       b.UserIdTag,
		LoginTip:        b.LoginTip,
	}, nil
}

func convertModelOauth2Configuration(m *model.Oauth2Configuration) (*biz.Oauth2Configuration, error) {
	if m.ClientKey == "" {
		data, err := pkgAes.AesDecrypt(m.ClientSecret)
		if err != nil {
			return nil, err
		} else {
			m.ClientKey = data
		}
	}
	p := &biz.Oauth2Configuration{
		Base:            convertBase(m.Model),
		UID:             m.UID,
		EnableOauth2:    m.EnableOauth2,
		ClientID:        m.ClientID,
		ClientKey:       m.ClientKey,
		ClientSecret:    m.ClientSecret,
		ClientHost:      m.ClientHost,
		ServerAuthUrl:   m.ServerAuthUrl,
		ServerTokenUrl:  m.ServerTokenUrl,
		ServerUserIdUrl: m.ServerUserIdUrl,
		Scopes:          strings.Split(m.Scopes, ","),
		AccessTokenTag:  m.AccessTokenTag,
		UserIdTag:       m.UserIdTag,
		LoginTip:        m.LoginTip,
	}
	return p, nil
}

func convertBizLDAPConfiguration(b *biz.LDAPConfiguration) (*model.LDAPConfiguration, error) {
	connectPassword, err := pkgAes.AesEncrypt(b.ConnectPassword)
	if err != nil {
		return nil, err
	}
	return &model.LDAPConfiguration{
		Model: model.Model{
			UID: b.UID,
		},
		Enable:                b.Enable,
		EnableSSL:             b.EnableSSL,
		Host:                  b.Host,
		Port:                  b.Port,
		ConnectDn:             b.ConnectDn,
		ConnectSecretPassword: connectPassword,
		BaseDn:                b.BaseDn,
		UserNameRdnKey:        b.UserNameRdnKey,
		UserEmailRdnKey:       b.UserEmailRdnKey,
	}, nil
}

func convertModelLDAPConfiguration(m *model.LDAPConfiguration) (*biz.LDAPConfiguration, error) {
	connectSecretPassword, err := pkgAes.AesDecrypt(m.ConnectSecretPassword)
	if err != nil {
		return nil, err
	}

	p := &biz.LDAPConfiguration{
		Base:            convertBase(m.Model),
		UID:             m.UID,
		Enable:          m.Enable,
		EnableSSL:       m.EnableSSL,
		Host:            m.Host,
		Port:            m.Port,
		ConnectDn:       m.ConnectDn,
		ConnectPassword: connectSecretPassword,
		BaseDn:          m.BaseDn,
		UserNameRdnKey:  m.UserNameRdnKey,
		UserEmailRdnKey: m.UserEmailRdnKey,
	}
	return p, nil
}

func convertBizSMTPConfiguration(b *biz.SMTPConfiguration) (*model.SMTPConfiguration, error) {
	secretPassword, err := pkgAes.AesEncrypt(b.Password)
	if err != nil {
		return nil, err
	}
	return &model.SMTPConfiguration{
		Model: model.Model{
			UID: b.UID,
		},
		EnableSMTPNotify: b.EnableSMTPNotify,
		Host:             b.Host,
		Port:             b.Port,
		Username:         b.Username,
		SecretPassword:   secretPassword,
		IsSkipVerify:     b.IsSkipVerify,
	}, nil
}

func convertModeSMTPConfiguration(m *model.SMTPConfiguration) (*biz.SMTPConfiguration, error) {
	connectSecretPassword, err := pkgAes.AesDecrypt(m.SecretPassword)
	if err != nil {
		return nil, err
	}

	p := &biz.SMTPConfiguration{
		Base:             convertBase(m.Model),
		UID:              m.UID,
		EnableSMTPNotify: m.EnableSMTPNotify,
		Host:             m.Host,
		Port:             m.Port,
		Username:         m.Username,
		Password:         connectSecretPassword,
		IsSkipVerify:     m.IsSkipVerify,
	}
	return p, nil
}

func convertBizWeChatConfiguration(b *biz.WeChatConfiguration) (*model.WeChatConfiguration, error) {
	encryptedCorpSecret, err := pkgAes.AesEncrypt(b.CorpSecret)
	if err != nil {
		return nil, err
	}
	return &model.WeChatConfiguration{
		Model: model.Model{
			UID: b.UID,
		},
		EnableWeChatNotify:  b.EnableWeChatNotify,
		CorpID:              b.CorpID,
		EncryptedCorpSecret: encryptedCorpSecret,
		AgentID:             b.AgentID,
		SafeEnabled:         b.SafeEnabled,
		ProxyIP:             b.ProxyIP,
	}, nil
}

func convertModeWeChatConfiguration(m *model.WeChatConfiguration) (*biz.WeChatConfiguration, error) {
	corpSecret, err := pkgAes.AesDecrypt(m.EncryptedCorpSecret)
	if err != nil {
		return nil, err
	}

	p := &biz.WeChatConfiguration{
		Base:               convertBase(m.Model),
		UID:                m.UID,
		EnableWeChatNotify: m.EnableWeChatNotify,
		CorpID:             m.CorpID,
		CorpSecret:         corpSecret,
		AgentID:            m.AgentID,
		SafeEnabled:        m.SafeEnabled,
		ProxyIP:            m.ProxyIP,
	}
	return p, nil
}

func convertBizWebHookConfiguration(b *biz.WebHookConfiguration) (*model.WebHookConfiguration, error) {
	encryptedToken, err := pkgAes.AesEncrypt(b.Token)
	if err != nil {
		return nil, err
	}
	return &model.WebHookConfiguration{
		Model: model.Model{
			UID: b.UID,
		},
		Enable:               b.Enable,
		MaxRetryTimes:        b.MaxRetryTimes,
		RetryIntervalSeconds: b.RetryIntervalSeconds,
		EncryptedToken:       encryptedToken,
		URL:                  b.URL,
	}, nil
}

func convertModeWebHookConfiguration(m *model.WebHookConfiguration) (*biz.WebHookConfiguration, error) {
	token, err := pkgAes.AesDecrypt(m.EncryptedToken)
	if err != nil {
		return nil, err
	}

	p := &biz.WebHookConfiguration{
		Base:                 convertBase(m.Model),
		UID:                  m.UID,
		Enable:               m.Enable,
		MaxRetryTimes:        m.MaxRetryTimes,
		RetryIntervalSeconds: m.RetryIntervalSeconds,
		Token:                token,
		URL:                  m.URL,
	}
	return p, nil
}

func convertBizIMConfiguration(b *biz.IMConfiguration) (*model.IMConfiguration, error) {
	return &model.IMConfiguration{
		Model: model.Model{
			UID: b.UID,
		},
		AppKey:      b.AppKey,
		AppSecret:   b.AppSecret,
		IsEnable:    b.IsEnable,
		ProcessCode: b.ProcessCode,
		Type:        string(b.Type),
	}, nil
}

func convertModeIMConfiguration(m *model.IMConfiguration) (*biz.IMConfiguration, error) {

	p := &biz.IMConfiguration{
		Base:        convertBase(m.Model),
		UID:         m.UID,
		AppKey:      m.AppKey,
		AppSecret:   m.AppSecret,
		IsEnable:    m.IsEnable,
		ProcessCode: m.ProcessCode,
		Type:        convertModelIMConfigType(m.Type),
	}
	return p, nil
}

func convertModelIMConfigType(_type string) biz.ImType {
	switch _type {
	case string(biz.ImTypeFeishu):
		return biz.ImTypeFeishu
	default:
		return biz.ImTypeUnknow
	}
}
