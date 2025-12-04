package sql_workbench

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/storage"
	dbmodel "github.com/actiontech/dms/internal/dms/storage/model"
	"github.com/actiontech/dms/internal/sql_workbench/client"
	config "github.com/actiontech/dms/internal/sql_workbench/config"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	_const "github.com/actiontech/dms/pkg/dms-common/pkg/const"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const SQL_WORKBENCH_URL = "/odc_query"
const SQL_WORKBENCH_PREFIX = "DMS-"
const SQL_WORKBENCH_DEFAULT_PASSWORD = "DMS123__"
const SQL_WORKBENCH_REAL_PASSWORD = "DMS123__@"
const INDIVIDUAL_SPACE = 4

// ODC 会话相关的 cookie 名称
const (
	ODCSessionCookieName   = "JSESSIONID"
	ODCXsrfTokenCookieName = "XSRF-TOKEN"
)

// odcSession ODC 会话缓存结构
type odcSession struct {
	dmsToken   string
	jsessionID string
	xsrfToken  string
}

var (
	dmsUserIdODCSessionMap = make(map[string]odcSession)
	odcSessionMutex        = &sync.Mutex{}
)

// generateSqlWorkbenchUsername 生成 SQL Workbench 用户名
func (s *SqlWorkbenchService) generateSqlWorkbenchUsername(dmsUserName string) string {
	return SQL_WORKBENCH_PREFIX + dmsUserName
}

// TempDBAccount 临时数据库账号
type TempDBAccount struct {
	DBAccountUid string            `json:"db_account_uid"`
	AccountInfo  AccountInfo       `json:"account_info"`
	Explanation  string            `json:"explanation"`
	ExpiredTime  string            `json:"expired_time"`
	DbService    dmsV1.UidWithName `json:"db_service"`
}

// AccountInfo 账号信息
type AccountInfo struct {
	User     string `json:"user"`
	Hostname string `json:"hostname"`
	Password string `json:"password"`
}

// ListDBAccountReply 数据库账号列表响应
type ListDBAccountReply struct {
	Data    []*TempDBAccount `json:"data"`
	Code    int              `json:"code"`
	Message string           `json:"message"`
}

type SqlWorkbenchService struct {
	cfg                        *config.SqlWorkbenchOpts
	log                        *utilLog.Helper
	client                     *client.SqlWorkbenchClient
	userUsecase                *biz.UserUsecase
	dbServiceUsecase           *biz.DBServiceUsecase
	projectUsecase             *biz.ProjectUsecase
	opPermissionVerifyUsecase  *biz.OpPermissionVerifyUsecase
	sqlWorkbenchUserRepo       biz.SqlWorkbenchUserRepo
	sqlWorkbenchDatasourceRepo biz.SqlWorkbenchDatasourceRepo
	proxyTargetRepo            biz.ProxyTargetRepo
}

func NewAndInitSqlWorkbenchService(logger utilLog.Logger, opts *conf.DMSOptions) (*SqlWorkbenchService, error) {
	var sqlWorkbenchClient *client.SqlWorkbenchClient
	if opts.SqlWorkBenchOpts != nil {
		sqlWorkbenchClient = client.NewSqlWorkbenchClient(opts.SqlWorkBenchOpts, logger)
	}

	// 初始化存储层
	st, err := storage.NewStorage(logger, &storage.StorageConfig{
		User:        opts.ServiceOpts.Database.UserName,
		Password:    opts.ServiceOpts.Database.Password,
		Host:        opts.ServiceOpts.Database.Host,
		Port:        opts.ServiceOpts.Database.Port,
		Schema:      opts.ServiceOpts.Database.Database,
		Debug:       opts.ServiceOpts.Database.Debug,
		AutoMigrate: opts.ServiceOpts.Database.AutoMigrate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	// 初始化事务生成器
	tx := storage.NewTXGenerator()

	// 初始化基础存储层
	opPermissionVerifyRepo := storage.NewOpPermissionVerifyRepo(logger, st)
	opPermissionVerifyUsecase := biz.NewOpPermissionVerifyUsecase(logger, tx, opPermissionVerifyRepo)

	// 初始化用户相关
	userRepo := storage.NewUserRepo(logger, st)
	userGroupRepo := storage.NewUserGroupRepo(logger, st)
	pluginRepo := storage.NewPluginRepo(logger, st)
	pluginUsecase, err := biz.NewDMSPluginUsecase(logger, pluginRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to new dms plugin usecase: %v", err)
	}

	opPermissionRepo := storage.NewOpPermissionRepo(logger, st)
	opPermissionUsecase := biz.NewOpPermissionUsecase(logger, tx, opPermissionRepo, pluginUsecase)

	cloudbeaverRepo := storage.NewCloudbeaverRepo(logger, st)
	loginConfigurationRepo := storage.NewLoginConfigurationRepo(logger, st)
	loginConfigurationUsecase := biz.NewLoginConfigurationUsecase(logger, tx, loginConfigurationRepo)

	ldapConfigurationRepo := storage.NewLDAPConfigurationRepo(logger, st)
	ldapConfigurationUsecase := biz.NewLDAPConfigurationUsecase(logger, tx, ldapConfigurationRepo)

	userUsecase := biz.NewUserUsecase(logger, tx, userRepo, userGroupRepo, pluginUsecase, opPermissionUsecase, opPermissionVerifyUsecase, loginConfigurationUsecase, ldapConfigurationUsecase, cloudbeaverRepo, nil)

	// 初始化项目相关
	memberUsecase := &biz.MemberUsecase{}
	environmentTagUsecase := biz.EnvironmentTagUsecase{}
	businessTagUsecase := biz.NewBusinessTagUsecase(storage.NewBusinessTagRepo(logger, st), logger)
	projectRepo := storage.NewProjectRepo(logger, st)
	projectUsecase := biz.NewProjectUsecase(logger, tx, projectRepo, memberUsecase, opPermissionVerifyUsecase, pluginUsecase, businessTagUsecase, &environmentTagUsecase)

	// 初始化数据源相关
	dbServiceRepo := storage.NewDBServiceRepo(logger, st)
	environmentTagUsecase = *biz.NewEnvironmentTagUsecase(storage.NewEnvironmentTagRepo(logger, st), logger, projectUsecase, opPermissionVerifyUsecase)
	proxyTargetRepo := storage.NewProxyTargetRepo(logger, st)
	dbServiceUsecase := biz.NewDBServiceUsecase(logger, dbServiceRepo, pluginUsecase, opPermissionVerifyUsecase, projectUsecase, proxyTargetRepo, &environmentTagUsecase)

	// 初始化SqlWorkbench相关的存储层
	sqlWorkbenchUserRepo := storage.NewSqlWorkbenchRepo(logger, st)
	sqlWorkbenchDatasourceRepo := storage.NewSqlWorkbenchDatasourceRepo(logger, st)

	return &SqlWorkbenchService{
		cfg:                        opts.SqlWorkBenchOpts,
		log:                        utilLog.NewHelper(logger, utilLog.WithMessageKey("sql_workbench.service")),
		client:                     sqlWorkbenchClient,
		userUsecase:                userUsecase,
		dbServiceUsecase:           dbServiceUsecase,
		projectUsecase:             projectUsecase,
		opPermissionVerifyUsecase:  opPermissionVerifyUsecase,
		sqlWorkbenchUserRepo:       sqlWorkbenchUserRepo,
		sqlWorkbenchDatasourceRepo: sqlWorkbenchDatasourceRepo,
		proxyTargetRepo:            proxyTargetRepo,
	}, nil
}

func (sqlWorkbenchService *SqlWorkbenchService) IsConfigured() bool {
	if sqlWorkbenchService.cfg == nil {
		return false
	}
	return sqlWorkbenchService.cfg != nil && sqlWorkbenchService.cfg.Host != "" && sqlWorkbenchService.cfg.Port != ""
}

func (sqlWorkbenchService *SqlWorkbenchService) GetOdcProxyTarget() ([]*middleware.ProxyTarget, error) {
	cfg := sqlWorkbenchService.cfg
	rawUrl, err := url.Parse(fmt.Sprintf("http://%v:%v", cfg.Host, cfg.Port))
	if err != nil {
		return nil, err
	}
	sqlWorkbenchService.log.Infof("ODC proxy target URL: %s", rawUrl.String())
	return []*middleware.ProxyTarget{
		{
			URL: rawUrl,
		},
	}, nil
}

func (sqlWorkbenchService *SqlWorkbenchService) GetRootUri() string {
	return SQL_WORKBENCH_URL
}

func (sqlWorkbenchService *SqlWorkbenchService) Login() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 从Cookie中获取DMS token
			var dmsToken string
			for _, cookie := range c.Cookies() {
				if cookie.Name == pkgConst.DMSToken {
					dmsToken = cookie.Value
					break
				}
			}

			if dmsToken == "" {
				sqlWorkbenchService.log.Errorf("dmsToken is empty")
				return fmt.Errorf("dms user token is empty")
			}

			dmsUserId, err := jwt.ParseUidFromJwtTokenStr(dmsToken)
			if err != nil {
				sqlWorkbenchService.log.Errorf("ParseUidFromJwtTokenStr err: %v", err)
				return fmt.Errorf("parse dms user uid from token err: %v", err)
			}

			// 检查缓存中是否有有效的 ODC 会话
			odcSession := sqlWorkbenchService.getODCSession(dmsUserId, dmsToken)
			if odcSession != nil {
				// 验证会话是否有效
				if sqlWorkbenchService.validateODCSession(odcSession.jsessionID, odcSession.xsrfToken) {
					// 会话有效，设置 cookie 到请求中
					sqlWorkbenchService.setODCCookiesToRequest(c, odcSession.jsessionID, odcSession.xsrfToken)
					sqlWorkbenchService.log.Debugf("Using cached ODC session for user: %s", dmsUserId)
					return next(c)
				}
				// 会话无效，清除缓存
				sqlWorkbenchService.log.Debugf("Cached ODC session invalid, clearing cache for user: %s", dmsUserId)
				sqlWorkbenchService.clearODCSession(dmsUserId)
			}

			// 缓存不存在或会话无效，需要重新登录
			sqlWorkbenchService.log.Debugf("No valid cached ODC session, performing login for user: %s", dmsUserId)

			// 1. 根据dmsUserId从数据库获取用户信息
			user, err := sqlWorkbenchService.userUsecase.GetUser(c.Request().Context(), dmsUserId)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to get user by dmsUserId %s: %v", dmsUserId, err)
				return err
			}

			// 2. 从数据库表SqlWorkbenchUserCache中判断sqlworkbench中是否存在该用户
			sqlWorkbenchUser, exists, err := sqlWorkbenchService.sqlWorkbenchUserRepo.GetSqlWorkbenchUserByDMSUserID(c.Request().Context(), dmsUserId)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to get sql workbench user cache: %v", err)
				return err
			}

			// 3. 如果用户不存在，调用sqlworkbench创建用户接口进行创建
			if !exists {
				err = sqlWorkbenchService.createSqlWorkbenchUser(c.Request().Context(), user)
				if err != nil {
					sqlWorkbenchService.log.Errorf("Failed to create sql workbench user: %v", err)
					return err
				}
				// 重新获取创建后的用户信息
				sqlWorkbenchUser, _, err = sqlWorkbenchService.sqlWorkbenchUserRepo.GetSqlWorkbenchUserByDMSUserID(c.Request().Context(), dmsUserId)
				if err != nil {
					sqlWorkbenchService.log.Errorf("Failed to get created sql workbench user: %v", err)
					return err
				}
			}

			// 4. 将DMS中的数据源同步给SqlWorkbench
			err = sqlWorkbenchService.syncDatasources(c.Request().Context(), user, sqlWorkbenchUser)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to sync datasources: %v", err)
				return err
			}

			// 5. 调用登录接口进行登录，并且从登录接口的返回值中获取Cookie设置到c echo.Context的上下文中
			jsessionID, xsrfToken, err := sqlWorkbenchService.loginSqlWorkbenchUser(c, user, dmsUserId, dmsToken)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to login sql workbench user: %v", err)
				return err
			}

			// 将新会话设置到请求中
			sqlWorkbenchService.setODCCookiesToRequest(c, jsessionID, xsrfToken)

			return next(c)
		}
	}
}

// createSqlWorkbenchUser 创建SqlWorkbench用户
func (sqlWorkbenchService *SqlWorkbenchService) createSqlWorkbenchUser(ctx context.Context, dmsUser *biz.User) error {
	cookie, _, publicKey, err := sqlWorkbenchService.getUserCookie(sqlWorkbenchService.cfg.AdminUser, sqlWorkbenchService.cfg.AdminPassword)
	if err != nil {
		return err
	}

	// 创建用户请求
	sqlWorkbenchUsername := sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name)
	createUserReq := []client.CreateUserRequest{
		{
			AccountName: sqlWorkbenchUsername,
			Name:        dmsUser.Name,
			Password:    SQL_WORKBENCH_DEFAULT_PASSWORD,
			Enabled:     true,
			RoleIDs:     []int64{INDIVIDUAL_SPACE},
		},
	}

	// 调用创建用户接口
	createUserResp, err := sqlWorkbenchService.client.CreateUsers(createUserReq, publicKey, cookie)
	if err != nil {
		return fmt.Errorf("failed to create user in sql workbench: %v", err)
	}

	if len(createUserResp.Data.Contents) == 0 {
		return fmt.Errorf("no user created in sql workbench")
	}

	// 激活用户
	activateUserResp, err := sqlWorkbenchService.client.ActivateUser(
		sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name),
		SQL_WORKBENCH_DEFAULT_PASSWORD,
		SQL_WORKBENCH_REAL_PASSWORD,
		publicKey,
		cookie,
	)
	if err != nil {
		return fmt.Errorf("failed to activate user in sql workbench: %v", err)
	}

	// 保存用户缓存
	sqlWorkbenchUser := &biz.SqlWorkbenchUser{
		SqlWorkbenchUsername: sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name),
		DMSUserID:            dmsUser.UID,
		SqlWorkbenchUserId:   activateUserResp.Data.ID,
	}

	err = sqlWorkbenchService.sqlWorkbenchUserRepo.SaveSqlWorkbenchUserCache(ctx, sqlWorkbenchUser)
	if err != nil {
		return fmt.Errorf("failed to save sql workbench user cache: %v", err)
	}

	sqlWorkbenchService.log.Infof("Successfully created and activated sql workbench user for DMS user %s (ID: %d)", dmsUser.Name, activateUserResp.Data.ID)
	return nil
}

// loginSqlWorkbenchUser 使用SqlWorkbench用户登录并设置Cookie
// 返回 jsessionID 和 xsrfToken，并缓存会话
func (sqlWorkbenchService *SqlWorkbenchService) loginSqlWorkbenchUser(c echo.Context, dmsUser *biz.User, dmsUserId, dmsToken string) (string, string, error) {
	// 获取公钥
	publicKey, err := sqlWorkbenchService.client.GetPublicKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to get public key: %v", err)
	}

	// 使用SqlWorkbench用户登录
	loginResp, err := sqlWorkbenchService.client.Login(sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name), SQL_WORKBENCH_REAL_PASSWORD, publicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to login sql workbench user: %v", err)
	}

	// 获取组织信息
	orgResp, err := sqlWorkbenchService.client.GetOrganizations(loginResp.Cookie)
	if err != nil {
		return "", "", fmt.Errorf("failed to get organizations: %v", err)
	}

	// 提取cookie值
	jsessionID := sqlWorkbenchService.client.ExtractCookieValue(loginResp.Cookie, ODCSessionCookieName)
	xsrfToken := sqlWorkbenchService.client.ExtractCookieValue(orgResp.XsrfToken, ODCXsrfTokenCookieName)

	// 设置Cookie到echo.Context中
	c.SetCookie(&http.Cookie{
		Name:     ODCSessionCookieName,
		Value:    jsessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	c.SetCookie(&http.Cookie{
		Name:     ODCXsrfTokenCookieName,
		Value:    xsrfToken,
		Path:     "/",
		HttpOnly: false, // XSRF-TOKEN通常需要JavaScript访问
		Secure:   false,
	})

	// 缓存会话
	sqlWorkbenchService.setODCSession(dmsUserId, dmsToken, jsessionID, xsrfToken)

	sqlWorkbenchService.log.Infof("Successfully logged in sql workbench user %s", dmsUser.Name)
	return jsessionID, xsrfToken, nil
}

// getODCSession 获取缓存的 ODC 会话
func (sqlWorkbenchService *SqlWorkbenchService) getODCSession(dmsUserId, dmsToken string) *odcSession {
	odcSessionMutex.Lock()
	defer odcSessionMutex.Unlock()

	if item, ok := dmsUserIdODCSessionMap[dmsUserId]; ok {
		if dmsToken == item.dmsToken {
			// 返回副本以避免并发安全问题
			session := item
			return &session
		}
	}

	return nil
}

// setODCSession 设置 ODC 会话缓存
func (sqlWorkbenchService *SqlWorkbenchService) setODCSession(dmsUserId, dmsToken, jsessionID, xsrfToken string) {
	odcSessionMutex.Lock()
	defer odcSessionMutex.Unlock()

	dmsUserIdODCSessionMap[dmsUserId] = odcSession{
		dmsToken:   dmsToken,
		jsessionID: jsessionID,
		xsrfToken:  xsrfToken,
	}
}

// clearODCSession 清除 ODC 会话缓存
func (sqlWorkbenchService *SqlWorkbenchService) clearODCSession(dmsUserId string) {
	odcSessionMutex.Lock()
	defer odcSessionMutex.Unlock()

	delete(dmsUserIdODCSessionMap, dmsUserId)
}

// validateODCSession 验证 ODC 会话是否有效
// 通过调用 GetOrganizations API 来验证会话
func (sqlWorkbenchService *SqlWorkbenchService) validateODCSession(jsessionID, xsrfToken string) bool {
	cookie := fmt.Sprintf("%s=%s; %s=%s", ODCSessionCookieName, jsessionID, ODCXsrfTokenCookieName, xsrfToken)
	_, err := sqlWorkbenchService.client.GetOrganizations(cookie)
	if err != nil {
		sqlWorkbenchService.log.Debugf("ODC session validation failed: %v", err)
		return false
	}
	return true
}

// setODCCookiesToRequest 将 ODC cookies 设置到请求中
func (sqlWorkbenchService *SqlWorkbenchService) setODCCookiesToRequest(c echo.Context, jsessionID, xsrfToken string) {
	// 更新请求的 Cookie header
	// 获取请求中现有的 Cookie header，用于保留客户端已有的 cookies
	currentCookies := c.Request().Header.Get("Cookie")
	cookieMap := make(map[string]string)
	// 解析现有 Cookie 字符串为 map，避免覆盖客户端已有的 cookies
	// Cookie 格式为 "key1=value1; key2=value2"，使用分号分隔
	if currentCookies != "" {
		existingCookies := strings.Split(currentCookies, ";")
		for _, cookie := range existingCookies {
			cookie = strings.TrimSpace(cookie)
			if cookie != "" {
				// 使用 SplitN 限制分割次数为 2，防止 cookie 值中包含 "=" 时被错误分割
				parts := strings.SplitN(cookie, "=", 2)
				if len(parts) == 2 {
					cookieMap[parts[0]] = parts[1]
				}
			}
		}
	}

	// 设置 ODC cookies
	cookieMap[ODCSessionCookieName] = jsessionID
	cookieMap[ODCXsrfTokenCookieName] = xsrfToken

	// 构建新的 Cookie header
	var cookieStrings []string
	for name, value := range cookieMap {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", name, value))
	}

	if len(cookieStrings) > 0 {
		c.Request().Header.Set("Cookie", strings.Join(cookieStrings, "; "))
	}

	// 更新请求 header 中的 X-XSRF-TOKEN
	c.Request().Header.Set("X-XSRF-TOKEN", xsrfToken)
}

// syncDatasources 同步DMS数据源到SqlWorkbench
func (sqlWorkbenchService *SqlWorkbenchService) syncDatasources(ctx context.Context, dmsUser *biz.User, sqlWorkbenchUser *biz.SqlWorkbenchUser) error {
	// 获取用户有权限访问的数据源
	activeDBServices, err := sqlWorkbenchService.getUserAccessibleDBServices(ctx, dmsUser)
	if err != nil {
		return fmt.Errorf("failed to get user accessible db services: %v", err)
	}

	// 获取当前用户Cookie
	userCookie, organizationId, _, err := sqlWorkbenchService.getUserCookie(sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name), SQL_WORKBENCH_REAL_PASSWORD)
	if err != nil {
		return fmt.Errorf("failed to get user cookie: %v", err)
	}

	// 同步数据源
	return sqlWorkbenchService.syncDBServicesToSqlWorkbench(ctx, activeDBServices, sqlWorkbenchUser, userCookie, organizationId)
}

// getUserAccessibleDBServices 获取用户有权限访问的数据源
func (sqlWorkbenchService *SqlWorkbenchService) getUserAccessibleDBServices(ctx context.Context, dmsUser *biz.User) ([]*biz.DBService, error) {
	// 获取所有活跃的数据源
	activeDBServices, err := sqlWorkbenchService.dbServiceUsecase.GetActiveDBServices(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get active db services: %v", err)
	}

	// 检查用户是否有全局权限
	hasGlobalOpPermission, err := sqlWorkbenchService.opPermissionVerifyUsecase.CanOpGlobal(ctx, dmsUser.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to check global op permission: %v", err)
	}

	if hasGlobalOpPermission {
		return activeDBServices, nil
	}

	// 获取用户的项目和数据源权限
	opPermissions, err := sqlWorkbenchService.opPermissionVerifyUsecase.GetUserOpPermission(ctx, dmsUser.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user op permissions: %v", err)
	}

	// 过滤有权限的数据源
	activeDBServices, err = sqlWorkbenchService.filterDBServicesByPermissions(ctx, activeDBServices, opPermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to filter db services by permissions: %v", err)
	}

	// 根据 provision “账号管理” 功能配置的权限进行过滤、修改连接用户
	activeDBServices, err = sqlWorkbenchService.ResetDbServiceByAuth(ctx, activeDBServices, dmsUser.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to reset db service by auth: %v", err)
	}

	return activeDBServices, nil
}

// filterDBServicesByPermissions 根据权限过滤数据源
func (sqlWorkbenchService *SqlWorkbenchService) filterDBServicesByPermissions(ctx context.Context, dbServices []*biz.DBService, opPermissions []biz.OpPermissionWithOpRange) ([]*biz.DBService, error) {
	projectIdMap := make(map[string]struct{})
	dbServiceIdMap := make(map[string]struct{})

	for _, opPermission := range opPermissions {
		// 项目权限
		if opPermission.OpRangeType == biz.OpRangeTypeProject && opPermission.OpPermissionUID == pkgConst.UIDOfOpPermissionProjectAdmin {
			for _, rangeUid := range opPermission.RangeUIDs {
				projectIdMap[rangeUid] = struct{}{}
			}
		}

		// 数据源权限
		if opPermission.OpRangeType == biz.OpRangeTypeDBService && opPermission.OpPermissionUID == pkgConst.UIDOfOpPermissionSQLQuery {
			for _, rangeUid := range opPermission.RangeUIDs {
				dbServiceIdMap[rangeUid] = struct{}{}
			}
		}
	}

	var filteredDBServices []*biz.DBService
	for _, dbService := range dbServices {
		// 检查项目权限
		if _, hasProjectPermission := projectIdMap[dbService.ProjectUID]; hasProjectPermission {
			filteredDBServices = append(filteredDBServices, dbService)
			continue
		}

		// 检查数据源权限
		if _, hasDBServicePermission := dbServiceIdMap[dbService.UID]; hasDBServicePermission {
			filteredDBServices = append(filteredDBServices, dbService)
		}
	}

	return filteredDBServices, nil
}

// getUserCookie 获取当前用户Cookie
func (sqlWorkbenchService *SqlWorkbenchService) getUserCookie(dmsUsername string, dmsUserPassword string) (string, int64, string, error) {
	// 获取公钥
	publicKey, err := sqlWorkbenchService.client.GetPublicKey()
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to get public key: %v", err)
	}

	// 使用当前用户账号登录
	loginResp, err := sqlWorkbenchService.client.Login(dmsUsername, dmsUserPassword, publicKey)
	if err != nil {
		return "", 0, publicKey, fmt.Errorf("failed to login as user: %v", err)
	}

	// 获取组织信息
	orgResp, err := sqlWorkbenchService.client.GetOrganizations(loginResp.Cookie)
	if err != nil {
		return "", 0, publicKey, fmt.Errorf("failed to get organizations: %v", err)
	}

	// 检查是否有足够的组织
	if len(orgResp.Data.Contents) < 2 {
		return "", 0, publicKey, fmt.Errorf("insufficient organizations, expected at least 2, got %d", len(orgResp.Data.Contents))
	}

	// 合并Cookie
	return sqlWorkbenchService.client.MergeCookies(orgResp.XsrfToken, loginResp.Cookie), orgResp.Data.Contents[1].ID, publicKey, nil
}

// getEnvironmentID 获取环境ID
func (sqlWorkbenchService *SqlWorkbenchService) getEnvironmentID(organizationID int64, cookie string) (int64, error) {
	// 调用GetEnvironments接口获取环境信息
	envResp, err := sqlWorkbenchService.client.GetEnvironments(organizationID, cookie)
	if err != nil {
		return 0, fmt.Errorf("failed to get environments: %v", err)
	}

	// 检查是否有环境
	if len(envResp.Data.Contents) == 0 {
		return 0, fmt.Errorf("no environments found")
	}

	// 优先选择"默认"环境，如果没有则选择第一个
	for _, env := range envResp.Data.Contents {
		if env.Name == "默认" {
			return env.ID, nil
		}
	}

	// 如果没有找到"默认"环境，返回第一个环境的ID
	return envResp.Data.Contents[0].ID, nil
}

// syncDBServicesToSqlWorkbench 同步数据源到SqlWorkbench
func (sqlWorkbenchService *SqlWorkbenchService) syncDBServicesToSqlWorkbench(ctx context.Context, dbServices []*biz.DBService, sqlWorkbenchUser *biz.SqlWorkbenchUser, userCookie string, organizationID int64) error {
	// 获取公钥
	publicKey, err := sqlWorkbenchService.client.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %v", err)
	}

	// 获取环境ID
	environmentID, err := sqlWorkbenchService.getEnvironmentID(organizationID, userCookie)
	if err != nil {
		return fmt.Errorf("failed to get environment id: %v", err)
	}

	// 获取用户现有的数据源缓存
	existingDatasources, err := sqlWorkbenchService.sqlWorkbenchDatasourceRepo.GetSqlWorkbenchDatasourcesByUserID(ctx, sqlWorkbenchUser.DMSUserID)
	if err != nil {
		return fmt.Errorf("failed to get existing datasources: %v", err)
	}

	// 创建数据源映射
	existingDatasourceMap := make(map[string]*biz.SqlWorkbenchDatasource)
	for _, ds := range existingDatasources {
		key := sqlWorkbenchService.getDatasourceKey(ds.DMSDBServiceID, ds.Purpose)
		existingDatasourceMap[key] = ds
	}

	if len(dbServices) == 0 {
		sqlWorkbenchService.log.Infof("No accessible db services for user, cleaning up all existing datasources")
		return sqlWorkbenchService.cleanupObsoleteDatasources(ctx, dbServices, existingDatasourceMap, userCookie, organizationID)
	}

	// 处理每个数据源
	for _, dbService := range dbServices {
		key := sqlWorkbenchService.getDatasourceKey(dbService.UID, dbService.AccountPurpose)

		if existingDatasource, exists := existingDatasourceMap[key]; exists {
			// 检查是否需要更新
			if sqlWorkbenchService.shouldUpdateDatasource(dbService, existingDatasource) {
				err = sqlWorkbenchService.updateDatasourceInSqlWorkbench(ctx, dbService, existingDatasource, publicKey, userCookie, organizationID, environmentID)
				if err != nil {
					sqlWorkbenchService.log.Errorf("Failed to update datasource %s: %v", dbService.Name, err)
				}
			}
		} else {
			// 创建新数据源
			err = sqlWorkbenchService.createDatasourceInSqlWorkbench(ctx, dbService, sqlWorkbenchUser, publicKey, userCookie, organizationID, environmentID)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to create datasource %s: %v", dbService.Name, err)
			}
		}
	}

	// 删除不再需要的数据源
	return sqlWorkbenchService.cleanupObsoleteDatasources(ctx, dbServices, existingDatasourceMap, userCookie, organizationID)
}

// getDatasourceKey 获取数据源唯一键
func (sqlWorkbenchService *SqlWorkbenchService) getDatasourceKey(dmsDBServiceID, purpose string) string {
	return fmt.Sprintf("%s:%s", dmsDBServiceID, purpose)
}

// shouldUpdateDatasource 判断是否需要更新数据源
func (sqlWorkbenchService *SqlWorkbenchService) shouldUpdateDatasource(dbService *biz.DBService, existingDatasource *biz.SqlWorkbenchDatasource) bool {
	// 比较数据源指纹，如果指纹不同则需要更新
	currentFingerprint := sqlWorkbenchService.dbServiceUsecase.GetDBServiceFingerprint(dbService)
	return existingDatasource.DMSDBServiceFingerprint != currentFingerprint
}

// createDatasourceInSqlWorkbench 在SqlWorkbench中创建数据源
func (sqlWorkbenchService *SqlWorkbenchService) createDatasourceInSqlWorkbench(ctx context.Context, dbService *biz.DBService, sqlWorkbenchUser *biz.SqlWorkbenchUser, publicKey, userCookie string, organizationID, environmentID int64) error {
	// 构建创建数据源请求
	createReq, err := sqlWorkbenchService.buildCreateDatasourceRequest(ctx, dbService, sqlWorkbenchUser, environmentID)
	if err != nil {
		return fmt.Errorf("failed to build create datasource request: %v", err)
	}

	// 调用创建接口
	createResp, err := sqlWorkbenchService.client.CreateDatasources(createReq, publicKey, userCookie, organizationID)
	if err != nil {
		return fmt.Errorf("failed to create datasource: %v", err)
	}

	// 保存缓存
	datasourceCache := &biz.SqlWorkbenchDatasource{
		DMSDBServiceID:           dbService.UID,
		DMSUserID:                sqlWorkbenchUser.DMSUserID,
		DMSDBServiceFingerprint:  sqlWorkbenchService.dbServiceUsecase.GetDBServiceFingerprint(dbService),
		SqlWorkbenchDatasourceID: createResp.Data.ID,
		Purpose:                  dbService.AccountPurpose,
	}

	err = sqlWorkbenchService.sqlWorkbenchDatasourceRepo.SaveSqlWorkbenchDatasourceCache(ctx, datasourceCache)
	if err != nil {
		return fmt.Errorf("failed to save datasource cache: %v", err)
	}

	sqlWorkbenchService.log.Infof("Successfully created datasource %s (ID: %d)", dbService.Name, createResp.Data.ID)
	return nil
}

// updateDatasourceInSqlWorkbench 在SqlWorkbench中更新数据源
func (sqlWorkbenchService *SqlWorkbenchService) updateDatasourceInSqlWorkbench(ctx context.Context, dbService *biz.DBService, existingDatasource *biz.SqlWorkbenchDatasource, publicKey, userCookie string, organizationID, environmentID int64) error {
	// 构建更新数据源请求
	updateReq, err := sqlWorkbenchService.buildUpdateDatasourceRequest(ctx, dbService, environmentID)
	if err != nil {
		return fmt.Errorf("failed to build update datasource request: %v", err)
	}

	// 调用更新接口
	_, err = sqlWorkbenchService.client.UpdateDatasource(existingDatasource.SqlWorkbenchDatasourceID, updateReq, publicKey, userCookie, organizationID)
	if err != nil {
		return fmt.Errorf("failed to update datasource: %v", err)
	}

	// 更新缓存中的指纹
	updatedDatasourceCache := &biz.SqlWorkbenchDatasource{
		DMSDBServiceID:           dbService.UID,
		DMSUserID:                existingDatasource.DMSUserID,
		DMSDBServiceFingerprint:  sqlWorkbenchService.dbServiceUsecase.GetDBServiceFingerprint(dbService),
		SqlWorkbenchDatasourceID: existingDatasource.SqlWorkbenchDatasourceID,
		Purpose:                  dbService.AccountPurpose,
	}

	err = sqlWorkbenchService.sqlWorkbenchDatasourceRepo.SaveSqlWorkbenchDatasourceCache(ctx, updatedDatasourceCache)
	if err != nil {
		return fmt.Errorf("failed to update datasource cache: %v", err)
	}

	sqlWorkbenchService.log.Infof("Successfully updated datasource %s (ID: %d)", dbService.Name, existingDatasource.SqlWorkbenchDatasourceID)
	return nil
}

// cleanupObsoleteDatasources 清理过时的数据源
func (sqlWorkbenchService *SqlWorkbenchService) cleanupObsoleteDatasources(ctx context.Context, currentDBServices []*biz.DBService, existingDatasourceMap map[string]*biz.SqlWorkbenchDatasource, userCookie string, organizationID int64) error {
	currentKeys := make(map[string]bool)
	for _, dbService := range currentDBServices {
		key := sqlWorkbenchService.getDatasourceKey(dbService.UID, dbService.AccountPurpose)
		currentKeys[key] = true
	}

	for key, existingDatasource := range existingDatasourceMap {
		if !currentKeys[key] {
			// 删除数据源
			_, err := sqlWorkbenchService.client.DeleteDatasource(existingDatasource.SqlWorkbenchDatasourceID, userCookie, organizationID)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to delete datasource %d: %v", existingDatasource.SqlWorkbenchDatasourceID, err)
				continue
			}

			// 删除缓存
			err = sqlWorkbenchService.sqlWorkbenchDatasourceRepo.DeleteSqlWorkbenchDatasourceCache(ctx, existingDatasource.DMSDBServiceID, existingDatasource.DMSUserID, existingDatasource.Purpose)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to delete datasource cache: %v", err)
			}

			sqlWorkbenchService.log.Infof("Successfully deleted obsolete datasource %d", existingDatasource.SqlWorkbenchDatasourceID)
		}
	}

	return nil
}

// getProjectName 获取项目名称
func (sqlWorkbenchService *SqlWorkbenchService) getProjectName(ctx context.Context, projectUID string) (string, error) {
	project, err := sqlWorkbenchService.projectUsecase.GetProject(ctx, projectUID)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %v", err)
	}
	return project.Name, nil
}

// buildDatasourceName 构建数据源名称，格式为项目名:数据源名
func (sqlWorkbenchService *SqlWorkbenchService) buildDatasourceName(ctx context.Context, dbService *biz.DBService) (string, error) {
	projectName, err := sqlWorkbenchService.getProjectName(ctx, dbService.ProjectUID)
	if err != nil {
		return "", fmt.Errorf("failed to get project name: %v", err)
	}
	return fmt.Sprintf("%s:%s", projectName, dbService.Name), nil
}

// datasourceBaseInfo 数据源基础信息
type datasourceBaseInfo struct {
	Name          string
	Type          string
	Username      string
	Password      string
	Host          string
	Port          string
	ServiceName   *string
	EnvironmentID int64
}

// buildDatasourceBaseInfo 构建数据源基础信息
func (sqlWorkbenchService *SqlWorkbenchService) buildDatasourceBaseInfo(ctx context.Context, dbService *biz.DBService, environmentID int64) (*datasourceBaseInfo, error) {
	datasourceName, err := sqlWorkbenchService.buildDatasourceName(ctx, dbService)
	if err != nil {
		return nil, err
	}

	baseInfo := &datasourceBaseInfo{
		Name:          datasourceName,
		Type:          sqlWorkbenchService.convertDBType(dbService.DBType),
		Username:      dbService.User,
		Password:      dbService.Password,
		Host:          dbService.Host,
		Port:          dbService.Port,
		EnvironmentID: environmentID,
	}

	// Oracle 特殊处理
	if dbService.DBType == "Oracle" {
		serviceName := dbService.AdditionalParams.GetParam("service_name").Value
		baseInfo.ServiceName = &serviceName
	}

	return baseInfo, nil
}

// buildCreateDatasourceRequest 构建创建数据源请求
func (sqlWorkbenchService *SqlWorkbenchService) buildCreateDatasourceRequest(ctx context.Context, dbService *biz.DBService, sqlWorkbenchUser *biz.SqlWorkbenchUser, environmentID int64) (client.CreateDatasourceRequest, error) {
	baseInfo, err := sqlWorkbenchService.buildDatasourceBaseInfo(ctx, dbService, environmentID)
	if err != nil {
		return client.CreateDatasourceRequest{}, err
	}

	return client.CreateDatasourceRequest{
		CreatorID:     sqlWorkbenchUser.SqlWorkbenchUserId,
		Type:          baseInfo.Type,
		Name:          baseInfo.Name,
		Username:      baseInfo.Username,
		Password:      baseInfo.Password,
		Host:          baseInfo.Host,
		Port:          baseInfo.Port,
		ServiceName:   baseInfo.ServiceName,
		SSLConfig:     client.SSLConfig{Enabled: false},
		EnvironmentID: baseInfo.EnvironmentID,
	}, nil
}

// buildUpdateDatasourceRequest 构建更新数据源请求
func (sqlWorkbenchService *SqlWorkbenchService) buildUpdateDatasourceRequest(ctx context.Context, dbService *biz.DBService, environmentID int64) (client.UpdateDatasourceRequest, error) {
	baseInfo, err := sqlWorkbenchService.buildDatasourceBaseInfo(ctx, dbService, environmentID)
	if err != nil {
		return client.UpdateDatasourceRequest{}, err
	}

	return client.UpdateDatasourceRequest{
		Type:          baseInfo.Type,
		Name:          &baseInfo.Name,
		Username:      baseInfo.Username,
		Password:      &baseInfo.Password,
		Host:          baseInfo.Host,
		Port:          baseInfo.Port,
		ServiceName:   baseInfo.ServiceName,
		SSLConfig:     client.SSLConfig{Enabled: false},
		EnvironmentID: baseInfo.EnvironmentID,
	}, nil
}

// convertDBType 转换数据库类型
func (sqlWorkbenchService *SqlWorkbenchService) convertDBType(dmsDBType string) string {
	// 这里需要根据实际的数据库类型映射关系进行转换
	// ODC目前支持的数据源有: OB_MYSQL, OB_ORACLE, ORACLE, MYSQL, ODP_SHARDING_OB_MYSQL, DORIS, POSTGRESQL
	// 其余调用创建数据源接口会直接失败
	switch dmsDBType {
	case "MySQL":
		return "MYSQL"
	case "PostgreSQL":
		return "POSTGRESQL"
	case "Oracle":
		return "ORACLE"
	case "SQLServer":
		return "SQLSERVER"
	case "OceanBase For Oracle":
		return "OB_ORACLE"
	case "OceanBase For MySQL":
		return "OB_MYSQL"
	default:
		return dmsDBType
	}
}

func (sqlWorkbenchService *SqlWorkbenchService) SupportDBType(dbType pkgConst.DBType) bool {
	return dbType == pkgConst.DBTypeMySQL || dbType == pkgConst.DBTypeOracle || dbType == pkgConst.DBTypeOceanBaseMySQL
}

// buildDatabaseUser 当是ob-mysql时需要给账号管理的账号附加租户名集群名等字符: root@oms_mysql#oms_resource_4250
func buildDatabaseUser(account string, dbServiceUser string, dbType string) string {
	if dbType == string(pkgConst.DBTypeOceanBaseMySQL) {
		index := strings.Index(dbServiceUser, "@")
		if index == -1 {
			return account
		}
		return account + dbServiceUser[index:]
	}
	return account
}

func ListAuthDbAccount(ctx context.Context, baseURL, projectUid, userId string) ([]*TempDBAccount, error) {
	// Generate token
	token, err := generateAuthToken(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token for user %s: %w", userId, err)
	}

	// Prepare request headers
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	// Build request URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Set the path for the endpoint
	u.Path = fmt.Sprintf("/provision/v1/auth/projects/%s/db_accounts", projectUid)

	// Add query parameters
	query := u.Query()
	query.Set("page_size", "999")
	query.Set("page_index", "1")
	query.Set("filter_by_password_managed", "true")
	query.Set("filter_by_status", "unlock")
	query.Set("filter_by_users", userId)
	u.RawQuery = query.Encode()
	requestURL := u.String()

	// Execute request
	reply := &ListDBAccountReply{}
	if err := makeHttpRequest(ctx, requestURL, header, reply); err != nil {
		return nil, err
	}

	// Validate response
	if reply.Code != 0 {
		return nil, fmt.Errorf("unexpected HTTP reply code (%v): %v", reply.Code, reply.Message)
	}

	return reply.Data, nil
}

// Helper function: Generate JWT token
func generateAuthToken(userId string) (string, error) {
	token, err := jwt.GenJwtToken(jwt.WithUserId(userId))
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}
	return token, nil
}

// Helper function: Make HTTP GET request
func makeHttpRequest(ctx context.Context, url string, headers map[string]string, reply interface{}) error {
	err := pkgHttp.Get(ctx, url, headers, nil, reply)
	if err != nil {
		return fmt.Errorf("HTTP request to %s failed: %w", url, err)
	}
	return nil
}

// AuditMiddleware 拦截工作台odc请求进行加工
func (sqlWorkbenchService *SqlWorkbenchService) AuditMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 只拦截包含 /streamExecute 的请求
			if !strings.Contains(c.Request().URL.Path, "/streamExecute") {
				return next(c)
			}

			// 读取请求体
			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return fmt.Errorf("failed to read request body: %w", err)
			}
			// 恢复请求体，供后续处理使用
			c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 解析请求体获取 SQL 和 datasource ID
			sql, datasourceID, err := sqlWorkbenchService.parseStreamExecuteRequest(bodyBytes)
			if err != nil {
				return fmt.Errorf("failed to parse streamExecute request, skipping audit: %v", err)
			}

			if sql == "" || datasourceID == "" {
				return fmt.Errorf("SQL or datasource ID is empty, skipping audit")
			}

			// 获取当前用户 ID
			dmsUserId, err := sqlWorkbenchService.getDMSUserIdFromRequest(c)
			if err != nil {
				return fmt.Errorf("failed to get DMS user ID: %v", err)
			}

			// 从缓存表获取 dms_db_service_id
			dmsDBServiceID, err := sqlWorkbenchService.getDMSDBServiceIDFromCache(c.Request().Context(), datasourceID, dmsUserId)
			if err != nil {
				return fmt.Errorf("failed to get dms_db_service_id from cache: %v", err)
			}

			if dmsDBServiceID == "" {
				return fmt.Errorf("dms_db_service_id not found in cache for datasource: %s", datasourceID)
			}

			// 获取 DBService 信息
			dbService, err := sqlWorkbenchService.dbServiceUsecase.GetDBService(c.Request().Context(), dmsDBServiceID)
			if err != nil {
				return fmt.Errorf("failed to get DBService: %v", err)
			}

			// 检查是否启用 SQL 审核
			if !sqlWorkbenchService.isEnableSQLAudit(dbService) {
				return fmt.Errorf("SQL audit is not enabled for DBService: %s", dmsDBServiceID)
			}

			// 调用 SQLE 审核接口
			auditResult, err := sqlWorkbenchService.callSQLEAudit(c.Request().Context(), sql, dbService)
			if err != nil {
				return fmt.Errorf("call SQLE audit failed: %v", err)
			}

			// 拦截响应并添加审核结果
			return sqlWorkbenchService.interceptAndAddAuditResult(c, next, auditResult, dbService)
		}
	}
}

// parseStreamExecuteRequest 解析 streamExecute 请求体，提取 SQL 和 datasource ID
func (sqlWorkbenchService *SqlWorkbenchService) parseStreamExecuteRequest(bodyBytes []byte) (sql string, datasourceID string, err error) {
	var requestBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal request body: %v", err)
	}

	// 从 sql 字段获取 SQL
	if sqlVal, ok := requestBody["sql"]; ok {
		if sqlStr, ok := sqlVal.(string); ok {
			sql = sqlStr
		}
	}

	// 从 sid 字段解析 datasource ID
	// sid 格式: sid:{base64编码的JSON}:d:dms
	// base64 JSON 包含: {"dbId":623,"dsId":28,"from":"192.168.21.47","logicalSession":false,"realId":"ee9b8ab276"}
	if sidVal, ok := requestBody["sid"]; ok {
		if sidStr, ok := sidVal.(string); ok {
			dsId, parseErr := sqlWorkbenchService.parseSidToDatasourceID(sidStr)
			if parseErr != nil {
				sqlWorkbenchService.log.Debugf("Failed to parse sid to datasource ID: %v", parseErr)
			} else {
				datasourceID = dsId
			}
		}
	}

	return sql, datasourceID, nil
}

// parseSidToDatasourceID 从 sid 字符串中解析出 datasource ID
// sid 格式: sid:{base64编码的JSON}:d:dms
func (sqlWorkbenchService *SqlWorkbenchService) parseSidToDatasourceID(sid string) (string, error) {
	// 检查 sid 格式: sid:...:d:dms
	if !strings.HasPrefix(sid, "sid:") {
		return "", fmt.Errorf("invalid sid format, missing 'sid:' prefix")
	}

	// 移除 "sid:" 前缀
	sid = strings.TrimPrefix(sid, "sid:")

	// 查找最后一个 ":d" 后缀并移除从 ":d" 开始的所有字符
	if idx := strings.LastIndex(sid, ":d"); idx != -1 {
		sid = sid[:idx]
	}

	// 解码 base64
	decodedBytes, err := base64.StdEncoding.DecodeString(sid)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 sid: %v", err)
	}

	// 解析 JSON
	var sidData struct {
		DbId           int    `json:"dbId"`
		DsId           int    `json:"dsId"`
		From           string `json:"from"`
		LogicalSession bool   `json:"logicalSession"`
		RealId         string `json:"realId"`
	}

	if err := json.Unmarshal(decodedBytes, &sidData); err != nil {
		return "", fmt.Errorf("failed to unmarshal sid JSON: %v", err)
	}

	// 返回 dsId 作为字符串
	return fmt.Sprintf("%d", sidData.DsId), nil
}

// getDMSUserIdFromRequest 从请求中获取 DMS 用户 ID
func (sqlWorkbenchService *SqlWorkbenchService) getDMSUserIdFromRequest(c echo.Context) (string, error) {
	var dmsToken string
	for _, cookie := range c.Cookies() {
		if cookie.Name == pkgConst.DMSToken {
			dmsToken = cookie.Value
			break
		}
	}

	if dmsToken == "" {
		return "", fmt.Errorf("dms token is empty")
	}

	dmsUserId, err := jwt.ParseUidFromJwtTokenStr(dmsToken)
	if err != nil {
		return "", fmt.Errorf("failed to parse dms user id from token: %v", err)
	}

	return dmsUserId, nil
}

// getDMSDBServiceIDFromCache 从 sql_workbench_datasource_caches 表获取 dms_db_service_id
func (sqlWorkbenchService *SqlWorkbenchService) getDMSDBServiceIDFromCache(ctx context.Context, datasourceID, dmsUserID string) (string, error) {
	// 尝试将 datasourceID 转换为 int64（ODC 的 datasource ID 通常是数字）
	var sqlWorkbenchDatasourceID int64
	if _, err := fmt.Sscanf(datasourceID, "%d", &sqlWorkbenchDatasourceID); err != nil {
		// 如果转换失败，尝试直接使用字符串作为 datasource ID
		sqlWorkbenchService.log.Debugf("Failed to convert datasourceID to int64, trying to find by string: %s", datasourceID)
	}

	// 从缓存表中查找，需要根据 sql_workbench_datasource_id 和 dms_user_id 查找
	// 由于缓存表可能没有直接存储 sql_workbench_datasource_id，我们需要通过其他方式查找
	// 这里先尝试通过用户 ID 获取所有数据源，然后匹配
	datasources, err := sqlWorkbenchService.sqlWorkbenchDatasourceRepo.GetSqlWorkbenchDatasourcesByUserID(ctx, dmsUserID)
	if err != nil {
		return "", fmt.Errorf("failed to get datasources by user id: %v", err)
	}

	// 如果 datasourceID 是数字，尝试匹配 SqlWorkbenchDatasourceID
	if sqlWorkbenchDatasourceID > 0 {
		for _, ds := range datasources {
			if ds.SqlWorkbenchDatasourceID == sqlWorkbenchDatasourceID {
				return ds.DMSDBServiceID, nil
			}
		}
	}

	// 如果找不到，返回第一个匹配的数据源（临时方案，后续可能需要更精确的匹配逻辑）
	if len(datasources) > 0 {
		// 这里可以根据实际业务逻辑选择合适的数据源
		// 暂时返回第一个数据源的 dms_db_service_id
		return datasources[0].DMSDBServiceID, nil
	}

	return "", fmt.Errorf("no datasource found for datasourceID: %s, userID: %s", datasourceID, dmsUserID)
}

// isEnableSQLAudit 检查是否启用 SQL 审核
func (sqlWorkbenchService *SqlWorkbenchService) isEnableSQLAudit(dbService *biz.DBService) bool {
	if dbService.SQLEConfig == nil || dbService.SQLEConfig.SQLQueryConfig == nil {
		return false
	}
	return dbService.SQLEConfig.AuditEnabled && dbService.SQLEConfig.SQLQueryConfig.AuditEnabled
}

// callSQLEAudit 调用 SQLE 直接审核接口
func (sqlWorkbenchService *SqlWorkbenchService) callSQLEAudit(ctx context.Context, sql string, dbService *biz.DBService) (*auditSQLReply, error) {
	// 获取 SQLE 服务地址
	target, err := sqlWorkbenchService.proxyTargetRepo.GetProxyTargetByName(ctx, _const.SqleComponentName)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQLE proxy target: %v", err)
	}

	sqleAddr := fmt.Sprintf("%s/v2/sql_audit", target.URL.String())

	// 构建审核请求
	auditReq := cloudbeaver.AuditSQLReq{
		InstanceType:     dbService.DBType,
		SQLContent:       sql,
		SQLType:          "sql",
		ProjectId:        dbService.ProjectUID,
		RuleTemplateName: dbService.SQLEConfig.SQLQueryConfig.RuleTemplateName,
	}

	// 设置请求头
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	// 调用 SQLE 审核接口
	reply := &auditSQLReply{}
	if err := pkgHttp.POST(ctx, sqleAddr, header, auditReq, reply); err != nil {
		return nil, fmt.Errorf("failed to call SQLE audit API: %v", err)
	}

	if reply.Code != 0 {
		return nil, fmt.Errorf("SQLE audit API returned error code %v: %v", reply.Code, reply.Message)
	}

	return reply, nil
}

// interceptAndAddAuditResult 拦截响应并添加审核结果
func (sqlWorkbenchService *SqlWorkbenchService) interceptAndAddAuditResult(c echo.Context, next echo.HandlerFunc, auditResult *auditSQLReply, dbService *biz.DBService) error {
	// 判断是否需要审批
	allowQueryWhenLessThanAuditLevel := dbService.GetAllowQueryWhenLessThanAuditLevel()
	needApproval := sqlWorkbenchService.shouldRequireApproval(auditResult.Data.SQLResults, allowQueryWhenLessThanAuditLevel)

	// 如果需要审批，直接返回审核结果，不请求真实的 streamExecute 接口
	if needApproval {
		return sqlWorkbenchService.buildAuditResponseWithoutExecution(c, auditResult, dbService)
	}

	// 不需要审批，执行真实请求并添加审核结果
	return sqlWorkbenchService.executeAndAddAuditResult(c, next, auditResult, dbService)
}

// buildAuditResponseWithoutExecution 构造审核响应，不执行真实的 SQL
func (sqlWorkbenchService *SqlWorkbenchService) buildAuditResponseWithoutExecution(c echo.Context, auditResult *auditSQLReply, dbService *biz.DBService) error {
	// 构造 SQL 条目列表
	sqlItems := make([]StreamExecuteSQLItem, 0, len(auditResult.Data.SQLResults))
	for _, sqlResult := range auditResult.Data.SQLResults {
		// 转换审核结果为 violatedRules 格式
		violatedRules := sqlWorkbenchService.convertSQLEAuditToViolatedRules(&sqlResult)

		sqlItem := StreamExecuteSQLItem{
			SQLTuple: StreamExecuteSQLTuple{
				ExecutedSQL: sqlResult.ExecSQL,
				Offset:      int(sqlResult.Number),
				OriginalSQL: sqlResult.ExecSQL,
				SQLID:       fmt.Sprintf("sqle-audit-%d", sqlResult.Number),
			},
			ViolatedRules: violatedRules,
		}
		sqlItems = append(sqlItems, sqlItem)
	}

	// 构造响应数据
	responseData := StreamExecuteData{
		ApprovalRequired:        true, // 需要审批
		LogicalSQL:              false,
		RequestID:               nil,
		SQLs:                    sqlItems,
		UnauthorizedDBResources: nil,
		ViolatedRules:           []interface{}{},
	}

	// 构造完整响应
	response := StreamExecuteResponse{
		Data:           responseData,
		DurationMillis: 0,
		HTTPStatus:     "OK",
		RequestID:      fmt.Sprintf("dms-audit-%d", time.Now().UnixNano()),
		Server:         "DMS",
		Successful:     true,
		Timestamp:      float64(time.Now().Unix()),
		TraceID:        c.Response().Header().Get("X-Trace-ID"),
	}

	// 返回 JSON 响应
	return c.JSON(http.StatusOK, response)
}

// executeAndAddAuditResult 执行真实请求并添加审核结果
func (sqlWorkbenchService *SqlWorkbenchService) executeAndAddAuditResult(c echo.Context, next echo.HandlerFunc, auditResult *auditSQLReply, dbService *biz.DBService) error {
	// 创建响应拦截器
	srw := newStreamExecuteResponseWriter(c)
	cloudbeaverResBuf := srw.Buffer
	c.Response().Writer = srw

	defer func() {
		// 在 defer 中处理响应
		if srw.status != 0 {
			srw.original.WriteHeader(srw.status)
		}

		// 读取响应内容
		responseBytes := cloudbeaverResBuf.Bytes()
		if len(responseBytes) == 0 {
			return
		}

		// 如果是 gzip 压缩响应，先解压
		responseBytes, wasGzip, err := sqlWorkbenchService.decodeResponseBody(cloudbeaverResBuf.Bytes(), srw.Header().Get("Content-Encoding"))
		if err != nil {
			sqlWorkbenchService.log.Debugf("Failed to decode response body, returning original response: %v", err)
			srw.original.Write(cloudbeaverResBuf.Bytes())
			return
		}

		// 如果解压过，先移除 Content-Encoding，后续根据需要重新设置
		if wasGzip {
			srw.original.Header().Del("Content-Encoding")
		}

		// 解析响应 JSON
		var responseBody StreamExecuteResponse
		if err := json.Unmarshal(responseBytes, &responseBody); err != nil {
			sqlWorkbenchService.log.Debugf("Failed to unmarshal response, returning original response: %v", err)
			// 如果解析失败，直接返回原始响应
			srw.original.Write(cloudbeaverResBuf.Bytes())
			return
		}

		// 添加审核结果到响应的 data 字段中
		if auditResult != nil && auditResult.Data != nil && auditResult.Data.PassRate != 1 {
			// 将 SQLE 审核结果整合到每个 SQL 条目中
			sqlWorkbenchService.mergeSQLEAuditResults(&responseBody.Data, auditResult.Data, dbService)
		}

		// 重新序列化响应
		modifiedResponse, err := json.Marshal(responseBody)
		if err != nil {
			sqlWorkbenchService.log.Errorf("Failed to marshal modified response: %v", err)
			srw.original.Write(cloudbeaverResBuf.Bytes())
			return
		}

		finalResponse := modifiedResponse
		if wasGzip {
			encoded, err := sqlWorkbenchService.encodeResponseBody(modifiedResponse)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to re-encode gzip response: %v", err)
				srw.original.Write(cloudbeaverResBuf.Bytes())
				return
			}
			finalResponse = encoded
			srw.original.Header().Set("Content-Encoding", "gzip")
		}

		// 更新 Content-Length
		header := srw.original.Header()
		header.Set("Content-Length", fmt.Sprintf("%d", len(finalResponse)))

		// 如果拦截过程中未显式写入状态码，默认使用 200
		statusCode := srw.status
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		srw.original.WriteHeader(statusCode)

		// 写入修改后的响应
		if _, err := srw.original.Write(finalResponse); err != nil {
			sqlWorkbenchService.log.Errorf("Failed to write modified response: %v", err)
		}
	}()

	// 执行下一个处理器
	return next(c)
}

// decodeResponseBody 根据 Content-Encoding 判断是否需要解压
func (sqlWorkbenchService *SqlWorkbenchService) decodeResponseBody(body []byte, contentEncoding string) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	isGzip := utils.IsGzip(body)
	if !isGzip {
		return body, false, nil
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, true, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, true, fmt.Errorf("failed to decompress gzip body: %w", err)
	}

	return decompressed, true, nil
}

// encodeResponseBody 将响应体按照 gzip 编码
func (sqlWorkbenchService *SqlWorkbenchService) encodeResponseBody(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(body); err != nil {
		gzipWriter.Close()
		return nil, fmt.Errorf("failed to gzip response body: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize gzip response body: %w", err)
	}
	return buf.Bytes(), nil
}

// mergeSQLEAuditResults 将 SQLE 审核结果整合到 sqls 数组中
func (sqlWorkbenchService *SqlWorkbenchService) mergeSQLEAuditResults(data *StreamExecuteData, auditData *auditResDataV2, dbService *biz.DBService) {
	// 创建 SQL 到审核结果的映射
	sqlAuditMap := make(map[string]*auditSQLResV2)
	for i := range auditData.SQLResults {
		sqlResult := &auditData.SQLResults[i]
		// 使用 exec_sql 作为 key，去除首尾空格和分号
		normalizedSQL := strings.TrimSpace(strings.TrimSuffix(sqlResult.ExecSQL, ";"))
		sqlAuditMap[normalizedSQL] = sqlResult
	}

	// 设置 ApprovalRequired 为 false，表示已通过审核可以执行
	// 注意：如果需要审批，在 interceptAndAddAuditResult 中已经拦截，不会执行到这里
	data.ApprovalRequired = false

	// 遍历 sqls 数组，为每个 SQL 条目添加 SQLE 审核结果
	for i := range data.SQLs {
		sqlItem := &data.SQLs[i]

		// 尝试从 originalSql 或 executedSql 匹配审核结果
		var matchedAuditResult *auditSQLResV2
		normalizedSQL := strings.TrimSpace(strings.TrimSuffix(sqlItem.SQLTuple.OriginalSQL, ";"))
		if auditResult, found := sqlAuditMap[normalizedSQL]; found {
			matchedAuditResult = auditResult
		}

		if matchedAuditResult == nil {
			normalizedSQL = strings.TrimSpace(strings.TrimSuffix(sqlItem.SQLTuple.ExecutedSQL, ";"))
			if auditResult, found := sqlAuditMap[normalizedSQL]; found {
				matchedAuditResult = auditResult
			}
		}

		// 如果找到匹配的审核结果，将其转换为 violatedRules 格式并添加
		if matchedAuditResult != nil {
			sqleViolatedRules := sqlWorkbenchService.convertSQLEAuditToViolatedRules(matchedAuditResult)
			if len(sqleViolatedRules) > 0 {
				sqlItem.ViolatedRules = sqleViolatedRules
			}
		}
	}
}

// shouldRequireApproval 根据审核放行等级判断是否需要审批
func (sqlWorkbenchService *SqlWorkbenchService) shouldRequireApproval(sqlResults []auditSQLResV2, allowQueryWhenLessThanAuditLevel string) bool {
	// 如果没有设置审核放行等级，那么直接放行
	if allowQueryWhenLessThanAuditLevel == "" {
		return false
	}

	// 遍历所有 SQL 审核结果
	for _, sqlResult := range sqlResults {
		// 检查是否有执行失败的审核项
		for _, auditItem := range sqlResult.AuditResult {
			if auditItem.ExecutionFailed {
				return true
			}
		}

		// 比较审核等级：如果 SQL 的审核等级大于允许的等级，则需要审批
		// 使用 RuleLevel 的 LessOrEqual 方法进行比较
		sqlAuditLevel := dbmodel.RuleLevel(sqlResult.AuditLevel)
		allowedLevel := dbmodel.RuleLevel(allowQueryWhenLessThanAuditLevel)

		// 如果 SQL 的审核等级大于允许的等级（即 !LessOrEqual），则需要审批
		if !sqlAuditLevel.LessOrEqual(allowedLevel) {
			return true
		}
	}

	// 所有 SQL 的审核等级都小于等于允许的等级，不需要审批
	return false
}

// convertSQLEAuditToViolatedRules 将 SQLE 审核结果转换为 violatedRules 格式
func (sqlWorkbenchService *SqlWorkbenchService) convertSQLEAuditToViolatedRules(auditResult *auditSQLResV2) []StreamExecuteRule {
	var violatedRules []StreamExecuteRule

	// 将 SQLE 的 audit_result 转换为 violatedRules 格式
	for _, auditItem := range auditResult.AuditResult {
		// 映射 level 字符串到数字
		levelNum := sqlWorkbenchService.mapAuditLevelToNumber(auditItem.Level)

		violatedRule := StreamExecuteRule{
			AppliedDialectTypes: nil,
			CreateTime:          nil,
			Enabled:             nil,
			ID:                  nil,
			Level:               levelNum,
			Metadata:            nil,
			OrganizationID:      nil,
			Properties:          nil,
			RulesetID:           nil,
			UpdateTime:          nil,
			Violation: StreamExecuteViolation{
				Level:            levelNum,
				LocalizedMessage: auditItem.Message,
				Offset:           0, // SQLE 审核结果可能没有 offset 信息
				Start:            0,
				Stop:             0,
				Text:             auditResult.ExecSQL,
			},
		}
		violatedRules = append(violatedRules, violatedRule)
	}

	return violatedRules
}

// mapAuditLevelToNumber 将审核级别字符串映射到数字
// normal=0, notice=3, warn=1, error=2
func (sqlWorkbenchService *SqlWorkbenchService) mapAuditLevelToNumber(level string) int {
	switch strings.ToLower(level) {
	case "normal":
		return 0
	case "notice":
		return 3
	case "warn":
		return 1
	case "error":
		return 2
	default:
		return 0 // 默认为 normal
	}
}

// StreamExecuteResponse streamExecute 接口响应结构
type StreamExecuteResponse struct {
	Data           StreamExecuteData `json:"data"`
	DurationMillis int64             `json:"durationMillis"`
	HTTPStatus     string            `json:"httpStatus"`
	RequestID      string            `json:"requestId"`
	Server         string            `json:"server"`
	Successful     bool              `json:"successful"`
	Timestamp      float64           `json:"timestamp"`
	TraceID        string            `json:"traceId"`
}

// StreamExecuteData streamExecute 响应中的 data 字段
type StreamExecuteData struct {
	ApprovalRequired        bool                   `json:"approvalRequired"`
	LogicalSQL              bool                   `json:"logicalSql"`
	RequestID               *string                `json:"requestId"`
	SQLs                    []StreamExecuteSQLItem `json:"sqls"`
	UnauthorizedDBResources interface{}            `json:"unauthorizedDBResources"`
	ViolatedRules           []interface{}          `json:"violatedRules"`
}

// StreamExecuteSQLItem SQL 条目
type StreamExecuteSQLItem struct {
	SQLTuple      StreamExecuteSQLTuple `json:"sqlTuple"`
	ViolatedRules []StreamExecuteRule   `json:"violatedRules"`
}

// StreamExecuteSQLTuple SQL 元组
type StreamExecuteSQLTuple struct {
	ExecutedSQL string `json:"executedSql"`
	Offset      int    `json:"offset"`
	OriginalSQL string `json:"originalSql"`
	SQLID       string `json:"sqlId"`
}

// StreamExecuteRule 违反的规则
type StreamExecuteRule struct {
	AppliedDialectTypes interface{}            `json:"appliedDialectTypes"`
	CreateTime          interface{}            `json:"createTime"`
	Enabled             interface{}            `json:"enabled"`
	ID                  interface{}            `json:"id"`
	Level               int                    `json:"level"`
	Metadata            interface{}            `json:"metadata"`
	OrganizationID      interface{}            `json:"organizationId"`
	Properties          interface{}            `json:"properties"`
	RulesetID           interface{}            `json:"rulesetId"`
	UpdateTime          interface{}            `json:"updateTime"`
	Violation           StreamExecuteViolation `json:"violation"`
}

// StreamExecuteViolation 违反详情
type StreamExecuteViolation struct {
	Level            int    `json:"level"`
	LocalizedMessage string `json:"localizedMessage"`
	Offset           int    `json:"offset"`
	Start            int    `json:"start"`
	Stop             int    `json:"stop"`
	Text             string `json:"text"`
}

// SQLEAuditResultSummary SQLE 审核结果汇总
type SQLEAuditResultSummary struct {
	AuditLevel string  `json:"audit_level"`
	Score      int32   `json:"score"`
	PassRate   float64 `json:"pass_rate"`
}

// streamExecuteResponseWriter 响应拦截器，用于捕获响应内容
type streamExecuteResponseWriter struct {
	echo.Response
	Buffer   *bytes.Buffer
	original http.ResponseWriter
	status   int
}

func newStreamExecuteResponseWriter(c echo.Context) *streamExecuteResponseWriter {
	buf := new(bytes.Buffer)
	return &streamExecuteResponseWriter{
		Response: *c.Response(),
		Buffer:   buf,
		original: c.Response().Writer,
	}
}

func (w *streamExecuteResponseWriter) Write(b []byte) (int, error) {
	// 如果未设置状态码，则补默认值
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	// 写入 buffer，不立即写给客户端
	return w.Buffer.Write(b)
}

func (w *streamExecuteResponseWriter) WriteHeader(code int) {
	w.status = code
}

// auditSQLReply SQLE 审核响应结构
type auditSQLReply struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *auditResDataV2 `json:"data"`
}

// auditResDataV2 审核结果数据
type auditResDataV2 struct {
	AuditLevel string          `json:"audit_level"`
	Score      int32           `json:"score"`
	PassRate   float64         `json:"pass_rate"`
	SQLResults []auditSQLResV2 `json:"sql_results"`
}

// auditSQLResV2 单个 SQL 审核结果
type auditSQLResV2 struct {
	Number      uint   `json:"number"`
	ExecSQL     string `json:"exec_sql"`
	AuditResult []struct {
		Level           string `json:"level"`
		Message         string `json:"message"`
		ExecutionFailed bool   `json:"execution_failed"`
	} `json:"audit_result"`
	AuditLevel string `json:"audit_level"`
}
