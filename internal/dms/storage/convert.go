package storage

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"github.com/labstack/echo/v4/middleware"

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
	encrypted, err := pkgAes.AesEncrypt(ds.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}
	dbService := &model.DBService{
		Model: model.Model{
			UID: ds.UID,
		},
		Name:              ds.Name,
		Desc:              ds.Desc,
		DBType:            ds.DBType,
		Host:              ds.Host,
		Port:              ds.Port,
		User:              ds.User,
		Password:          encrypted,
		Business:          ds.Business,
		AdditionalParams:  ds.AdditionalParams,
		Source:            ds.Source,
		MaintenancePeriod: ds.MaintenancePeriod,
		ProjectUID:        ds.ProjectUID,
		IsEnableMasking:   ds.IsMaskingSwitch,
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

	dbService := &biz.DBService{
		Base:              convertBase(ds.Model),
		UID:               ds.UID,
		Name:              ds.Name,
		Desc:              ds.Desc,
		DBType:            ds.DBType,
		Host:              ds.Host,
		Port:              ds.Port,
		User:              ds.User,
		Password:          decrypted,
		MaintenancePeriod: ds.MaintenancePeriod,
		Business:          ds.Business,
		AdditionalParams:  ds.AdditionalParams,
		Source:            ds.Source,
		ProjectUID:        ds.ProjectUID,
		IsMaskingSwitch:   ds.IsEnableMasking,
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

	ret := &biz.DatabaseSourceServiceParams{
		UID:         m.UID,
		Name:        m.Name,
		Source:      m.Source,
		Version:     m.Version,
		URL:         m.URL,
		DbType:      m.DbType,
		CronExpress: m.CronExpress,
		ProjectUID:  m.ProjectUID,
		LastSyncErr: m.LastSyncErr,
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
		ThirdPartyUserInfo:     u.ThirdPartyUserInfo,
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
		DMSUserID:               u.DMSUserId,
		DMSDBServiceFingerprint: u.DMSDBServiceFingerprint,
		Purpose:                 u.Purpose,
		CloudbeaverConnectionID: u.CloudbeaverConnectionID,
	}
}

func convertBizDatabaseSourceService(u *biz.DatabaseSourceServiceParams) *model.DatabaseSourceService {
	m := &model.DatabaseSourceService{
		Model:               model.Model{UID: u.UID},
		Name:                u.Name,
		Source:              u.Source,
		Version:             u.Version,
		URL:                 u.URL,
		DbType:              u.DbType,
		CronExpress:         u.CronExpress,
		ProjectUID:          u.ProjectUID,
		LastSyncErr:         u.LastSyncErr,
		LastSyncSuccessTime: u.LastSyncSuccessTime,
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

func convertBizBasicConfig(u *biz.BasicConfigParams) *model.BasicConfig {
	m := &model.BasicConfig{
		Model: model.Model{UID: u.UID, CreatedAt: u.CreatedAt},
		Title: u.Title,
		Logo:  u.Logo,
	}

	return m
}

func convertModelBasicConfig(m *model.BasicConfig) *biz.BasicConfigParams {
	return &biz.BasicConfigParams{
		Base:  convertBase(m.Model),
		UID:   m.UID,
		Title: m.Title,
		Logo:  m.Logo,
	}
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
		ThirdPartyUserInfo:     u.ThirdPartyUserInfo,
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
			DMSUserId:               item.DMSUserID,
			DMSDBServiceFingerprint: item.DMSDBServiceFingerprint,
			Purpose:                 item.Purpose,
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

func convertModelMemberGroup(mg *model.MemberGroup) (*biz.MemberGroup, error) {
	roles := make([]biz.MemberRoleWithOpRange, 0, len(mg.RoleWithOpRanges))
	for _, p := range mg.RoleWithOpRanges {
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

	users := make([]biz.UIdWithName, 0, len(mg.Users))
	for _, user := range mg.Users {
		users = append(users, biz.UIdWithName{
			Uid:  user.UID,
			Name: user.Name,
		})
	}

	return &biz.MemberGroup{
		Base:             convertBase(mg.Model),
		UID:              mg.UID,
		ProjectUID:       mg.ProjectUID,
		Name:             mg.Name,
		Users:            users,
		RoleWithOpRanges: roles,
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
		NeedInitOpPermissions:          u.NeedInitOpPermissions,
		NeedInitUsers:                  u.NeedInitUsers,
		NeedInitRoles:                  u.NeedInitRoles,
		NeedInitProjects:               u.NeedInitProjects,
		EnableSQLResultSetsDataMasking: u.EnableSQLResultSetsDataMasking,
	}, nil
}

func convertModelDMSConfig(u *model.DMSConfig) (*biz.DMSConfig, error) {
	return &biz.DMSConfig{
		Base:                           convertBase(u.Model),
		UID:                            u.UID,
		NeedInitOpPermissions:          u.NeedInitOpPermissions,
		NeedInitUsers:                  u.NeedInitUsers,
		NeedInitRoles:                  u.NeedInitRoles,
		NeedInitProjects:               u.NeedInitProjects,
		EnableSQLResultSetsDataMasking: u.EnableSQLResultSetsDataMasking,
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
		ProjectUID:       m.ProjectUID,
		RoleWithOpRanges: roles,
	}, nil
}

func convertBizMemberGroup(m *biz.MemberGroup) *model.MemberGroup {
	roles := make([]model.MemberGroupRoleOpRange, 0, len(m.RoleWithOpRanges))
	for _, p := range m.RoleWithOpRanges {
		roles = append(roles, model.MemberGroupRoleOpRange{
			MemberGroupUID: m.UID,
			RoleUID:        p.RoleUID,
			OpRangeType:    p.OpRangeType.String(),
			RangeUIDs:      convertBizRangeUIDs(p.RangeUIDs),
		})
	}

	var users []*model.User
	for _, uid := range m.UserUids {
		users = append(users, &model.User{Model: model.Model{UID: uid}})
	}

	return &model.MemberGroup{
		Model: model.Model{
			UID:       m.UID,
			CreatedAt: m.CreatedAt,
		},
		Name:             m.Name,
		ProjectUID:       m.ProjectUID,
		RoleWithOpRanges: roles,
		Users:            users,
	}
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
		ProjectUID:       m.ProjectUID,
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

func convertBizProject(m *biz.Project) (*model.Project, error) {
	busList := make([]model.Business, 0)
	for _, business := range m.Business {
		busList = append(busList, model.Business{
			Uid:  business.Uid,
			Name: business.Name,
		})
	}

	return &model.Project{
		Model: model.Model{
			UID: m.UID,
		},
		Name:            m.Name,
		Desc:            m.Desc,
		Business:        busList,
		Status:          string(m.Status),
		IsFixedBusiness: m.IsFixedBusiness,
		CreateUserUID:   m.CreateUserUID,
	}, nil
}

func convertModelProject(m *model.Project) (*biz.Project, error) {
	businessList := make([]biz.Business, 0)
	for _, business := range m.Business {
		businessList = append(businessList, biz.Business{
			Uid:  business.Uid,
			Name: business.Name,
		})
	}

	return &biz.Project{
		Base:            convertBase(m.Model),
		UID:             m.UID,
		Name:            m.Name,
		Desc:            m.Desc,
		IsFixedBusiness: m.IsFixedBusiness,
		Business:        businessList,
		Status:          convertModelProjectStatus(m.Status),
		CreateUserUID:   m.CreateUserUID,
		CreateTime:      m.CreatedAt,
	}, nil
}

func convertModelProjectStatus(status string) biz.ProjectStatus {
	switch status {
	case string(biz.ProjectStatusActive):
		return biz.ProjectStatusActive
	case string(biz.ProjectStatusArchived):
		return biz.ProjectStatusArchived
	default:
		return biz.ProjectStatusUnknown
	}
}

func convertModelProxyScenario(scenario string) biz.ProxyScenario {
	switch scenario {
	case "internal_service":
		return biz.ProxyScenarioInternalService
	case "thrid_party_integrate":
		return biz.ProxyScenarioThirdPartyIntegrate
	default:
		return biz.ProxyScenarioUnknown
	}
}

func convertBizProxyTarget(t *biz.ProxyTarget) (*model.ProxyTarget, error) {
	return &model.ProxyTarget{
		Name:            t.Name,
		Url:             t.URL.String(),
		Version:         t.Version,
		ProxyUrlPrefixs: strings.Join(t.GetProxyUrlPrefixs(), ";"),
		Scenario:        string(t.Scenario),
	}, nil
}

func convertModelProxyTarget(t *model.ProxyTarget) (*biz.ProxyTarget, error) {
	url, err := url.ParseRequestURI(t.Url)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %s", t.Url)
	}
	p := &biz.ProxyTarget{
		ProxyTarget: middleware.ProxyTarget{
			Name: t.Name,
			URL:  url,
			Meta: echo.Map{},
		},
		Version:  t.Version,
		Scenario: convertModelProxyScenario(t.Scenario),
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
		SkipCheckState:  b.SkipCheckState,
		AutoCreateUser:  b.AutoCreateUser,
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
		UserWeChatTag:   b.UserWeChatTag,
		UserEmailTag:    b.UserEmailTag,
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
		SkipCheckState:  m.SkipCheckState,
		AutoCreateUser:  m.AutoCreateUser,
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
		UserWeChatTag:   m.UserWeChatTag,
		UserEmailTag:    m.UserEmailTag,
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

func convertBizCompanyNotice(b *biz.CompanyNotice) (*model.CompanyNotice, error) {
	return &model.CompanyNotice{
		Model: model.Model{
			UID: b.UID,
		},
		NoticeStr:   b.NoticeStr,
		ReadUserIds: b.ReadUserIds,
	}, nil
}

func convertModelCompanyNotice(m *model.CompanyNotice) (*biz.CompanyNotice, error) {
	p := &biz.CompanyNotice{
		Base:        convertBase(m.Model),
		UID:         m.UID,
		NoticeStr:   m.NoticeStr,
		ReadUserIds: m.ReadUserIds,
	}

	return p, nil
}

func convertModelClusterLeader(m *model.ClusterLeader) (*biz.ClusterLeader, error) {
	return &biz.ClusterLeader{
		Anchor:       m.Anchor,
		ServerId:     m.ServerId,
		LastSeenTime: m.LastSeenTime,
	}, nil
}

func convertBizCLusterNodeInfo(b *biz.ClusterNodeInfo) *model.ClusterNodeInfo {
	return &model.ClusterNodeInfo{
		ServerId:     b.ServerId,
		HardwareSign: b.HardwareSign,
	}
}

func convertModelClusterNodeInfo(m *model.ClusterNodeInfo) (*biz.ClusterNodeInfo, error) {
	return &biz.ClusterNodeInfo{
		ServerId:     m.ServerId,
		HardwareSign: m.HardwareSign,
	}, nil
}

func convertBizWorkflow(b *biz.Workflow) *model.Workflow {
	workflow := &model.Workflow{
		Model: model.Model{
			UID: b.UID,
		},
		Name:              b.Name,
		ProjectUID:        b.ProjectUID,
		WorkflowType:      b.WorkflowType,
		Desc:              b.Desc,
		CreateTime:        &b.CreateTime,
		CreateUserUID:     b.CreateUserUID,
		WorkflowRecordUid: b.WorkflowRecordUid,
	}
	if b.WorkflowRecord != nil {
		workflow.WorkflowRecord = convertBizWorkflowRecord(b.WorkflowRecord)
	}
	return workflow
}

func convertModelWorkflow(m *model.Workflow) (w *biz.Workflow, err error) {
	w = &biz.Workflow{
		Base: convertBase(m.Model),

		UID:               m.UID,
		Name:              m.Name,
		ProjectUID:        m.ProjectUID,
		WorkflowType:      m.WorkflowType,
		Desc:              m.Desc,
		CreateTime:        *m.CreateTime,
		CreateUserUID:     m.CreateUserUID,
		WorkflowRecordUid: m.WorkflowRecordUid,
	}
	if m.WorkflowRecord != nil {
		w.WorkflowRecord, err = convertModelWorkflowRecord(m.WorkflowRecord)
		if err != nil {
			return w, err
		}
		w.Status = m.WorkflowRecord.Status
		for _, taskId := range m.WorkflowRecord.TaskIds {
			w.WorkflowRecord.Tasks = append(w.WorkflowRecord.Tasks, biz.Task{UID: taskId})
		}
	}

	return w, nil
}

func convertBizWorkflowRecord(b *biz.WorkflowRecord) *model.WorkflowRecord {
	var taskIds model.Strings
	for _, task := range b.Tasks {
		taskIds = append(taskIds, task.UID)
	}
	workflowRecord := &model.WorkflowRecord{
		Model: model.Model{
			UID: b.UID,
		},
		CurrentWorkflowStepId: b.CurrentWorkflowStepId,
		Status:                b.Status.String(),
		TaskIds:               taskIds,
	}
	if b.WorkflowSteps != nil {
		for _, step := range b.WorkflowSteps {
			workflowRecord.Steps = append(workflowRecord.Steps, convertBizWorkflowStep(step))
		}
	}
	return workflowRecord
}

func convertModelWorkflowRecord(m *model.WorkflowRecord) (wr *biz.WorkflowRecord, err error) {
	wr = &biz.WorkflowRecord{
		UID:                   m.UID,
		Status:                biz.DataExportWorkflowStatus(m.Status),
		Tasks:                 make([]biz.Task, 0),
		CurrentWorkflowStepId: m.CurrentWorkflowStepId,
	}
	if m.Steps != nil {
		wr.WorkflowSteps = make([]*biz.WorkflowStep, 0)
		for _, step := range m.Steps {
			wr.WorkflowSteps = append(wr.WorkflowSteps, &biz.WorkflowStep{
				StepId:            step.StepId,
				WorkflowRecordUid: step.WorkflowRecordUid,
				OperationUserUid:  step.OperationUserUid,
				OperateAt:         step.OperateAt,
				State:             step.State,
				Reason:            step.Reason,
				Assignees:         step.Assignees,
			})
		}
	}
	return wr, nil
}

func convertBizWorkflowStep(b *biz.WorkflowStep) *model.WorkflowStep {
	return &model.WorkflowStep{
		StepId:            b.StepId,
		WorkflowRecordUid: b.WorkflowRecordUid,
		OperationUserUid:  b.OperationUserUid,
		OperateAt:         b.OperateAt,
		State:             b.State,
		Reason:            b.Reason,
		Assignees:         b.Assignees,
	}
}

func convertModelWorkflowStep(m *model.WorkflowStep) (*biz.WorkflowStep, error) {
	return &biz.WorkflowStep{}, nil
}

func convertBizDataExportTask(b *biz.DataExportTask) *model.DataExportTask {
	dataExportTask := &model.DataExportTask{
		Model:             model.Model{UID: b.UID},
		CreateUserUID:     b.CreateUserUID,
		DBServiceUid:      b.DBServiceUid,
		DatabaseName:      b.DatabaseName,
		WorkFlowRecordUid: b.WorkFlowRecordUid,
		ExportType:        b.ExportType,
		ExportFileType:    b.ExportFileType,
		ExportFileName:    b.ExportFileName,
		ExportStatus:      b.ExportStatus.String(),
		ExportStartTime:   b.ExportStartTime,
		ExportEndTime:     b.ExportEndTime,
		AuditPassRate:     b.AuditPassRate,
		AuditScore:        b.AuditScore,
		AuditLevel:        b.AuditLevel,
	}
	if b.DataExportTaskRecords != nil {
		for _, record := range b.DataExportTaskRecords {
			dataExportTask.DataExportTaskRecords = append(dataExportTask.DataExportTaskRecords, convertBizDataExportTaskRecords(record))
		}

	}
	return dataExportTask
}

func convertModelDataExportTask(m *model.DataExportTask) *biz.DataExportTask {
	w := &biz.DataExportTask{
		Base:              convertBase(m.Model),
		UID:               m.UID,
		DBServiceUid:      m.DBServiceUid,
		CreateUserUID:     m.CreateUserUID,
		DatabaseName:      m.DatabaseName,
		WorkFlowRecordUid: m.WorkFlowRecordUid,
		ExportType:        m.ExportType,
		ExportFileType:    m.ExportFileType,
		ExportFileName:    m.ExportFileName,
		AuditPassRate:     m.AuditPassRate,
		AuditScore:        m.AuditScore,
		AuditLevel:        m.AuditLevel,
		ExportStatus:      biz.DataExportTaskStatus(m.ExportStatus),
		ExportStartTime:   m.ExportStartTime,
		ExportEndTime:     m.ExportEndTime,
	}
	if m.DataExportTaskRecords != nil {
		for _, r := range m.DataExportTaskRecords {
			w.DataExportTaskRecords = append(w.DataExportTaskRecords, convertModelDataExportTaskRecords(r))
		}

	}
	return w
}

func convertBizDataExportTaskRecords(b *biz.DataExportTaskRecord) *model.DataExportTaskRecord {
	m := &model.DataExportTaskRecord{
		Number:           b.Number,
		DataExportTaskId: b.DataExportTaskId,
		ExportSQL:        b.ExportSQL,
		ExportSQLType:    b.ExportSQLType,
		ExportResult:     b.ExportResult,
		AuditLevel:       b.AuditLevel,
	}
	if len(b.AuditSQLResults) != 0 {
		for _, v := range b.AuditSQLResults {
			m.AuditResults = append(m.AuditResults, model.AuditResult{
				Level:    v.Level,
				Message:  v.Message,
				RuleName: v.RuleName,
			})
		}
	}
	return m
}

func convertModelDataExportTaskRecords(m *model.DataExportTaskRecord) *biz.DataExportTaskRecord {
	b := &biz.DataExportTaskRecord{
		Number:           m.Number,
		DataExportTaskId: m.DataExportTaskId,
		ExportSQL:        m.ExportSQL,
		ExportSQLType:    m.ExportSQLType,
		AuditLevel:       m.AuditLevel,
		ExportResult:     m.ExportResult,
	}
	if len(m.AuditResults) != 0 {
		for _, v := range m.AuditResults {
			b.AuditSQLResults = append(b.AuditSQLResults, &biz.AuditResult{
				Level:    v.Level,
				Message:  v.Message,
				RuleName: v.RuleName,
			})
		}
	}
	return b
}
