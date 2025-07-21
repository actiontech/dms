package storage

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"github.com/labstack/echo/v4/middleware"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
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
		AdditionalParams:  ds.AdditionalParams,
		Source:            ds.Source,
		MaintenancePeriod: ds.MaintenancePeriod,
		ProjectUID:        ds.ProjectUID,
		IsEnableMasking:   ds.IsMaskingSwitch,
		EnableBackup:      ds.EnableBackup,
		BackupMaxRows:     ds.BackupMaxRows,
	}
	if ds.EnvironmentTag != nil {
		dbService.EnvironmentTagUID = ds.EnvironmentTag.UID
	}
	if ds.LastConnectionStatus != nil {
		dbService.LastConnectionStatus = (*string)(ds.LastConnectionStatus)
	}
	if ds.LastConnectionTime != nil {
		dbService.LastConnectionTime = ds.LastConnectionTime
	}
	if ds.LastConnectionErrorMsg != nil {
		dbService.LastConnectionErrorMsg = ds.LastConnectionErrorMsg
	}
	{
		// add sqle config
		if ds.SQLEConfig != nil {
			dbService.ExtraParameters = model.ExtraParameters{
				SqleConfig: &model.SQLEConfig{
					AuditEnabled:               ds.SQLEConfig.AuditEnabled,
					RuleTemplateName:           ds.SQLEConfig.RuleTemplateName,
					RuleTemplateID:             ds.SQLEConfig.RuleTemplateID,
					DataExportRuleTemplateName: ds.SQLEConfig.DataExportRuleTemplateName,
					DataExportRuleTemplateID:   ds.SQLEConfig.DataExportRuleTemplateID,
				},
			}
			sqleQueryConfig := ds.SQLEConfig.SQLQueryConfig
			if sqleQueryConfig != nil {
				dbService.ExtraParameters.SqleConfig.SqlQueryConfig = &model.SqlQueryConfig{
					AllowQueryWhenLessThanAuditLevel: sqleQueryConfig.AllowQueryWhenLessThanAuditLevel,
					AuditEnabled:                     sqleQueryConfig.AuditEnabled,
					MaxPreQueryRows:                  sqleQueryConfig.MaxPreQueryRows,
					QueryTimeoutSecond:               sqleQueryConfig.QueryTimeoutSecond,
					RuleTemplateName:                 sqleQueryConfig.RuleTemplateName,
					RuleTemplateID:                   sqleQueryConfig.RuleTemplateID,
				}
			}
		}
	}
	return dbService, nil
}

func convertModelDBService(ds *model.DBService) (*biz.DBService, error) {
	if ds == nil {
		return nil, nil
	}
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
		AdditionalParams:  ds.AdditionalParams,
		Source:            ds.Source,
		ProjectUID:        ds.ProjectUID,
		IsMaskingSwitch:   ds.IsEnableMasking,
		EnableBackup:      ds.EnableBackup,
		BackupMaxRows:     ds.BackupMaxRows,
	}

	if ds.LastConnectionStatus != nil {
		dbService.LastConnectionStatus = (*biz.LastConnectionStatus)(ds.LastConnectionStatus)
	}
	if ds.LastConnectionTime != nil {
		dbService.LastConnectionTime = ds.LastConnectionTime
	}
	if ds.LastConnectionErrorMsg != nil {
		dbService.LastConnectionErrorMsg = ds.LastConnectionErrorMsg
	}

	if ds.EnvironmentTag != nil {
		dbService.EnvironmentTag = &dmsCommonV1.EnvironmentTag{
			UID:  ds.EnvironmentTagUID,
			Name: ds.EnvironmentTag.EnvironmentName,
		}
	}

	{
		modelSqleConfig := ds.ExtraParameters.SqleConfig
		if modelSqleConfig != nil {
			dbService.SQLEConfig = &biz.SQLEConfig{}
			dbService.SQLEConfig.AuditEnabled = modelSqleConfig.AuditEnabled
			dbService.SQLEConfig.RuleTemplateName = modelSqleConfig.RuleTemplateName
			dbService.SQLEConfig.RuleTemplateID = modelSqleConfig.RuleTemplateID
			dbService.SQLEConfig.DataExportRuleTemplateName = modelSqleConfig.DataExportRuleTemplateName
			dbService.SQLEConfig.DataExportRuleTemplateID = modelSqleConfig.DataExportRuleTemplateID
			sqleQueryConfig := modelSqleConfig.SqlQueryConfig
			if sqleQueryConfig != nil {
				sqc := &biz.SQLQueryConfig{
					AllowQueryWhenLessThanAuditLevel: sqleQueryConfig.AllowQueryWhenLessThanAuditLevel,
					AuditEnabled:                     sqleQueryConfig.AuditEnabled,
					MaxPreQueryRows:                  sqleQueryConfig.MaxPreQueryRows,
					QueryTimeoutSecond:               sqleQueryConfig.QueryTimeoutSecond,
					RuleTemplateName:                 sqleQueryConfig.RuleTemplateName,
					RuleTemplateID:                   sqleQueryConfig.RuleTemplateID,
				}
				dbService.SQLEConfig.SQLQueryConfig = sqc
			}
		}
	}
	return dbService, nil
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
		TwoFactorEnabled:       u.TwoFactorEnabled,
		Name:                   u.Name,
		ThirdPartyUserID:       u.ThirdPartyUserID,
		ThirdPartyUserInfo:     u.ThirdPartyUserInfo,
		Password:               encrypted,
		Email:                  u.Email,
		Phone:                  u.Phone,
		WeChatID:               u.WxID,
		Language:               u.Language,
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
	if u == nil {
		return nil, nil
	}
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

	projects := make([]string, 0)
	for _, member := range u.Members {
		if member != nil && member.Project != nil {
			projects = append(projects, member.Project.Name)
		}
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
		Language:               u.Language,
		Projects:               projects,
		UserAuthenticationType: typ,
		Stat:                   stat,
		TwoFactorEnabled:       u.TwoFactorEnabled,
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
	projects := make([]string, 0)
	if m.User != nil && len(m.User.Members) > 0 {
		projects = append(projects, m.User.Members[0].Project.Name)
	}
	return &biz.Member{
		Base:             convertBase(m.Model),
		UID:              m.UID,
		ProjectUID:       m.ProjectUID,
		Projects:         projects,
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

func convertBizProject(m *biz.Project) *model.Project {
	return &model.Project{
		Model: model.Model{
			UID: m.UID,
		},
		Name:           m.Name,
		Desc:           m.Desc,
		BusinessTagUID: m.BusinessTag.UID,
		Status:         string(m.Status),
		CreateUserUID:  m.CreateUserUID,
		Priority:       dmsCommonV1.ToPriorityNum(m.Priority),
	}
}

func convertModelProject(m *model.Project) (*biz.Project, error) {
	return &biz.Project{
		Base:          convertBase(m.Model),
		UID:           m.UID,
		Name:          m.Name,
		Desc:          m.Desc,
		BusinessTag:   biz.BusinessTag{UID: m.BusinessTagUID},
		Status:        convertModelProjectStatus(m.Status),
		CreateUserUID: m.CreateUserUID,
		CreateTime:    m.CreatedAt,
		Priority:      dmsCommonV1.ToPriority(m.Priority),
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
		OperateDataResourceHandleUrl: t.OperateDataResourceHandleUrl,
		GetDatabaseDriverOptionsUrl:  t.GetDatabaseDriverOptionsUrl,
		GetDatabaseDriverLogosUrl:    t.GetDatabaseDriverLogosUrl,
	}, nil
}

func convertModelPlugin(t *model.Plugin) (*biz.Plugin, error) {
	p := &biz.Plugin{
		Name:                         t.Name,
		OperateDataResourceHandleUrl: t.OperateDataResourceHandleUrl,
		GetDatabaseDriverOptionsUrl:  t.GetDatabaseDriverOptionsUrl,
		GetDatabaseDriverLogosUrl:    t.GetDatabaseDriverLogosUrl,
	}
	return p, nil
}

func convertBizOAuth2Session(s *biz.OAuth2Session) (*model.OAuth2Session, error) {
	return &model.OAuth2Session{
		Model: model.Model{
			UID:       s.UID,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		},
		UserUID:         s.UserUID,
		Sub:             s.Sub,
		Sid:             s.Sid,
		IdToken:         s.IdToken,
		RefreshToken:    s.RefreshToken,
		LastLogoutEvent: sql.NullString{String: s.LastLogoutEvent, Valid: true},
		DeleteAfter:     s.DeleteAfter,
	}, nil
}

func convertModelOAuth2Session(m *model.OAuth2Session) (*biz.OAuth2Session, error) {
	return &biz.OAuth2Session{
		Base:            convertBase(m.Model),
		UID:             m.UID,
		UserUID:         m.UserUID,
		Sub:             m.Sub,
		Sid:             m.Sid,
		IdToken:         m.IdToken,
		RefreshToken:    m.RefreshToken,
		LastLogoutEvent: m.LastLogoutEvent.String,
		DeleteAfter:     m.DeleteAfter,
	}, nil
}

func convertBizLoginConfiguration(b *biz.LoginConfiguration) (*model.LoginConfiguration, error) {
	return &model.LoginConfiguration{
		Model: model.Model{
			UID: b.UID,
		},
		LoginButtonText:     b.LoginButtonText,
		DisableUserPwdLogin: b.DisableUserPwdLogin,
	}, nil
}

func convertModelLoginConfiguration(m *model.LoginConfiguration) (*biz.LoginConfiguration, error) {
	return &biz.LoginConfiguration{
		Base:                convertBase(m.Model),
		UID:                 m.UID,
		LoginButtonText:     m.LoginButtonText,
		DisableUserPwdLogin: m.DisableUserPwdLogin,
	}, nil
}

func convertBizOauth2Configuration(b *biz.Oauth2Configuration) (*model.Oauth2Configuration, error) {
	data, err := pkgAes.AesEncrypt(b.ClientKey)
	if err != nil {
		return nil, err
	}
	pwd, err := pkgAes.AesEncrypt(b.AutoCreateUserPWD)
	if err != nil {
		return nil, err
	}
	b.ClientSecret = data
	b.AutoCreateUserSecret = pwd
	return &model.Oauth2Configuration{
		Model: model.Model{
			UID: b.UID,
		},
		EnableOauth2:         b.EnableOauth2,
		SkipCheckState:       b.SkipCheckState,
		AutoCreateUser:       b.AutoCreateUser,
		AutoCreateUserPWD:    b.AutoCreateUserPWD,
		AutoCreateUserSecret: b.AutoCreateUserSecret,
		ClientID:             b.ClientID,
		ClientKey:            b.ClientKey,
		ClientSecret:         b.ClientSecret,
		ClientHost:           b.ClientHost,
		ServerAuthUrl:        b.ServerAuthUrl,
		ServerTokenUrl:       b.ServerTokenUrl,
		ServerUserIdUrl:      b.ServerUserIdUrl,
		ServerLogoutUrl:      b.ServerLogoutUrl,
		Scopes:               strings.Join(b.Scopes, ","),
		AccessTokenTag:       b.AccessTokenTag,
		UserIdTag:            b.UserIdTag,
		LoginTip:             b.LoginTip,
		UserWeChatTag:        b.UserWeChatTag,
		UserEmailTag:         b.UserEmailTag,
		LoginPermExpr:        b.LoginPermExpr,
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
	if m.AutoCreateUserPWD == "" {
		data, err := pkgAes.AesDecrypt(m.AutoCreateUserSecret)
		if err != nil {
			return nil, err
		} else {
			m.AutoCreateUserPWD = data
		}
	}
	p := &biz.Oauth2Configuration{
		Base:                 convertBase(m.Model),
		UID:                  m.UID,
		EnableOauth2:         m.EnableOauth2,
		SkipCheckState:       m.SkipCheckState,
		AutoCreateUser:       m.AutoCreateUser,
		AutoCreateUserPWD:    m.AutoCreateUserPWD,
		AutoCreateUserSecret: m.AutoCreateUserSecret,
		ClientID:             m.ClientID,
		ClientKey:            m.ClientKey,
		ClientSecret:         m.ClientSecret,
		ClientHost:           m.ClientHost,
		ServerAuthUrl:        m.ServerAuthUrl,
		ServerTokenUrl:       m.ServerTokenUrl,
		ServerUserIdUrl:      m.ServerUserIdUrl,
		ServerLogoutUrl:      m.ServerLogoutUrl,
		Scopes:               strings.Split(m.Scopes, ","),
		AccessTokenTag:       m.AccessTokenTag,
		UserIdTag:            m.UserIdTag,
		LoginTip:             m.LoginTip,
		UserWeChatTag:        m.UserWeChatTag,
		UserEmailTag:         m.UserEmailTag,
		LoginPermExpr:        m.LoginPermExpr,
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
		AuditResults:     b.AuditSQLResults,
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
		AuditSQLResults:  m.AuditResults,
	}
	return b
}

func convertBizCbOperationLog(src *biz.CbOperationLog) *model.CbOperationLog {
	return &model.CbOperationLog{
		Model: model.Model{
			UID: src.UID,
		},
		ProjectID:         src.ProjectID,
		OpPersonUID:       src.OpPersonUID,
		OpTime:            src.OpTime,
		DBServiceUID:      src.DBServiceUID,
		OpType:            string(src.OpType),
		I18nOpDetail:      src.I18nOpDetail,
		OpSessionID:       src.OpSessionID,
		OpHost:            src.OpHost,
		AuditResult:       src.AuditResults,
		IsAuditPassed:     src.IsAuditPass,
		ExecResult:        src.ExecResult,
		ExecTotalSec:      src.ExecTotalSec,
		ResultSetRowCount: src.ResultSetRowCount,
	}
}

func convertModelCbOperationLog(model *model.CbOperationLog) (*biz.CbOperationLog, error) {
	user, err := convertModelUser(model.User)
	if err != nil {
		return nil, err
	}

	dbService, err := convertModelDBService(model.DbService)
	if err != nil {
		return nil, err
	}

	project, err := convertModelProject(model.Project)
	if err != nil {
		return nil, err
	}

	return &biz.CbOperationLog{
		UID:               model.UID,
		ProjectID:         model.ProjectID,
		OpPersonUID:       model.OpPersonUID,
		OpTime:            model.OpTime,
		DBServiceUID:      model.DBServiceUID,
		OpType:            biz.CbOperationLogType(model.OpType),
		I18nOpDetail:      model.I18nOpDetail,
		OpSessionID:       model.OpSessionID,
		OpHost:            model.OpHost,
		AuditResults:      model.AuditResult,
		IsAuditPass:       model.IsAuditPassed,
		ExecResult:        model.ExecResult,
		ExecTotalSec:      model.ExecTotalSec,
		ResultSetRowCount: model.ResultSetRowCount,
		User:              user,
		DbService:         dbService,
		Project:           project,
	}, nil
}

func toModelDBServiceSyncTask(u *biz.DBServiceSyncTask) *model.DBServiceSyncTask {
	ret := &model.DBServiceSyncTask{
		Model:               model.Model{UID: u.UID},
		Name:                u.Name,
		Source:              u.Source,
		URL:                 u.URL,
		DbType:              u.DbType,
		CronExpress:         u.CronExpress,
		LastSyncErr:         u.LastSyncErr,
		LastSyncSuccessTime: u.LastSyncSuccessTime,
	}
	if u.LastSyncSuccessTime != nil {
		ret.LastSyncSuccessTime = u.LastSyncSuccessTime
	}
	if u.SQLEConfig != nil {
		ret.ExtraParameters.SqleConfig = &model.SQLEConfig{
			AuditEnabled:               u.SQLEConfig.AuditEnabled,
			RuleTemplateName:           u.SQLEConfig.RuleTemplateName,
			RuleTemplateID:             u.SQLEConfig.RuleTemplateID,
			DataExportRuleTemplateName: u.SQLEConfig.DataExportRuleTemplateName,
			DataExportRuleTemplateID:   u.SQLEConfig.DataExportRuleTemplateID,
		}
		if u.SQLEConfig.SQLQueryConfig != nil {
			ret.ExtraParameters.SqleConfig.SqlQueryConfig = &model.SqlQueryConfig{
				MaxPreQueryRows:                  u.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond,
				QueryTimeoutSecond:               u.SQLEConfig.SQLQueryConfig.QueryTimeoutSecond,
				AuditEnabled:                     u.SQLEConfig.SQLQueryConfig.AuditEnabled,
				AllowQueryWhenLessThanAuditLevel: u.SQLEConfig.SQLQueryConfig.AllowQueryWhenLessThanAuditLevel,
				RuleTemplateName:                 u.SQLEConfig.SQLQueryConfig.RuleTemplateName,
				RuleTemplateID:                   u.SQLEConfig.SQLQueryConfig.RuleTemplateID,
			}
		}
	}
	if u.AdditionalParam != nil {
		ret.ExtraParameters.AdditionalParam = u.AdditionalParam
	}
	return ret
}

func toBizDBServiceSyncTask(m *model.DBServiceSyncTask) *biz.DBServiceSyncTask {
	ret := &biz.DBServiceSyncTask{
		UID:         m.UID,
		Name:        m.Name,
		Source:      m.Source,
		URL:         m.URL,
		DbType:      m.DbType,
		CronExpress: m.CronExpress,
		LastSyncErr: m.LastSyncErr,
	}
	if m.LastSyncSuccessTime != nil {
		ret.LastSyncSuccessTime = m.LastSyncSuccessTime
	}
	if m.ExtraParameters.SqleConfig != nil {
		ret.SQLEConfig = &biz.SQLEConfig{
			AuditEnabled:               m.ExtraParameters.SqleConfig.AuditEnabled,
			RuleTemplateName:           m.ExtraParameters.SqleConfig.RuleTemplateName,
			RuleTemplateID:             m.ExtraParameters.SqleConfig.RuleTemplateID,
			DataExportRuleTemplateName: m.ExtraParameters.SqleConfig.DataExportRuleTemplateName,
			DataExportRuleTemplateID:   m.ExtraParameters.SqleConfig.DataExportRuleTemplateID,
		}
		if m.ExtraParameters.SqleConfig.SqlQueryConfig != nil {
			ret.SQLEConfig.SQLQueryConfig = &biz.SQLQueryConfig{
				MaxPreQueryRows:                  m.ExtraParameters.SqleConfig.SqlQueryConfig.QueryTimeoutSecond,
				QueryTimeoutSecond:               m.ExtraParameters.SqleConfig.SqlQueryConfig.QueryTimeoutSecond,
				AuditEnabled:                     m.ExtraParameters.SqleConfig.SqlQueryConfig.AuditEnabled,
				AllowQueryWhenLessThanAuditLevel: m.ExtraParameters.SqleConfig.SqlQueryConfig.AllowQueryWhenLessThanAuditLevel,
				RuleTemplateName:                 m.ExtraParameters.SqleConfig.SqlQueryConfig.RuleTemplateName,
				RuleTemplateID:                   m.ExtraParameters.SqleConfig.SqlQueryConfig.RuleTemplateID,
			}
		}
	}
	if m.ExtraParameters.AdditionalParam != nil {
		ret.AdditionalParam = m.ExtraParameters.AdditionalParam
	}
	return ret
}
