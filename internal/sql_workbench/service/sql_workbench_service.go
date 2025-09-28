package sql_workbench

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/apiserver/conf"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/dms/storage"
	"github.com/actiontech/dms/internal/sql_workbench/client"
	config "github.com/actiontech/dms/internal/sql_workbench/config"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
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

func (sqlWorkbenchService *SqlWorkbenchService) GetSqlWorkbenchConfiguration() (reply *dmsV1.GetSQLQueryConfigurationReply, err error) {
	sqlWorkbenchService.log.Infof("GetSqlWorkbenchConfiguration")
	defer func() {
		sqlWorkbenchService.log.Infof("GetSqlWorkbenchConfiguration; reply=%v, error=%v", reply, err)
	}()

	return &dmsV1.GetSQLQueryConfigurationReply{
		Data: struct {
			EnableSQLQuery  bool   `json:"enable_sql_query"`
			SQLQueryRootURI string `json:"sql_query_root_uri"`
		}{
			EnableSQLQuery:  sqlWorkbenchService.IsConfigured(),
			SQLQueryRootURI: SQL_WORKBENCH_URL,
		},
	}, nil
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
			var dmsToken string
			for _, cookie := range c.Cookies() {
				if cookie.Name == pkgConst.DMSToken {
					dmsToken = cookie.Value
					break
				}
			}
			dmsUserId, err := jwt.ParseUidFromJwtTokenStr(dmsToken)
			if err != nil {
				return err
			}

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
			err = sqlWorkbenchService.loginSqlWorkbenchUser(c, user)
			if err != nil {
				sqlWorkbenchService.log.Errorf("Failed to login sql workbench user: %v", err)
				return err
			}

			return next(c)
		}
	}
}

// createSqlWorkbenchUser 创建SqlWorkbench用户
func (sqlWorkbenchService *SqlWorkbenchService) createSqlWorkbenchUser(ctx context.Context, dmsUser *biz.User) error {
	cookie, publicKey, err := sqlWorkbenchService.getAdminCookie()
	if err != nil {
		return err
	}

	// 创建用户请求
	sqlWorkbenchUsername := sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name)
	createUserReq := []client.CreateUserRequest{
		{
			AccountName: sqlWorkbenchUsername,
			Name:        sqlWorkbenchUsername,
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
func (sqlWorkbenchService *SqlWorkbenchService) loginSqlWorkbenchUser(c echo.Context, dmsUser *biz.User) error {
	// 获取公钥
	publicKey, err := sqlWorkbenchService.client.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %v", err)
	}

	// 使用SqlWorkbench用户登录
	loginResp, err := sqlWorkbenchService.client.Login(sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name), SQL_WORKBENCH_REAL_PASSWORD, publicKey)
	if err != nil {
		return fmt.Errorf("failed to login sql workbench user: %v", err)
	}

	// 获取组织信息
	orgResp, err := sqlWorkbenchService.client.GetOrganizations(loginResp.Cookie)
	if err != nil {
		return fmt.Errorf("failed to get organizations: %v", err)
	}

	// 设置Cookie到echo.Context中
	c.SetCookie(&http.Cookie{
		Name:     "JSESSIONID",
		Value:    sqlWorkbenchService.client.ExtractCookieValue(loginResp.Cookie, "JSESSIONID"),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 根据实际情况设置
	})

	c.SetCookie(&http.Cookie{
		Name:     "XSRF-TOKEN",
		Value:    sqlWorkbenchService.client.ExtractCookieValue(orgResp.XsrfToken, "XSRF-TOKEN"),
		Path:     "/",
		HttpOnly: false, // XSRF-TOKEN通常需要JavaScript访问
		Secure:   false, // 根据实际情况设置
	})

	sqlWorkbenchService.log.Infof("Successfully logged in sql workbench user %s", dmsUser.Name)
	return nil
}

// syncDatasources 同步DMS数据源到SqlWorkbench
func (sqlWorkbenchService *SqlWorkbenchService) syncDatasources(ctx context.Context, dmsUser *biz.User, sqlWorkbenchUser *biz.SqlWorkbenchUser) error {
	// 获取用户有权限访问的数据源
	activeDBServices, err := sqlWorkbenchService.getUserAccessibleDBServices(ctx, dmsUser)
	if err != nil {
		return fmt.Errorf("failed to get user accessible db services: %v", err)
	}

	// 获取当前用户Cookie
	userCookie, organizationId, err := sqlWorkbenchService.getUserCookie(dmsUser)
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

// getAdminCookie 获取SqlWorkbench管理员Cookie
func (sqlWorkbenchService *SqlWorkbenchService) getAdminCookie() (string, string, error) {
	// 获取公钥
	publicKey, err := sqlWorkbenchService.client.GetPublicKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to get public key: %v", err)
	}

	// 使用管理员账号登录
	loginResp, err := sqlWorkbenchService.client.Login(sqlWorkbenchService.cfg.AdminUser, sqlWorkbenchService.cfg.AdminPassword, publicKey)
	if err != nil {
		return "", publicKey, fmt.Errorf("failed to login as admin: %v", err)
	}

	// 获取组织信息
	orgResp, err := sqlWorkbenchService.client.GetOrganizations(loginResp.Cookie)
	if err != nil {
		return "", publicKey, fmt.Errorf("failed to get organizations: %v", err)
	}

	// 合并Cookie
	return sqlWorkbenchService.client.MergeCookies(orgResp.XsrfToken, loginResp.Cookie), publicKey, nil
}

// getUserCookie 获取当前用户Cookie
func (sqlWorkbenchService *SqlWorkbenchService) getUserCookie(dmsUser *biz.User) (string, int64, error) {
	// 获取公钥
	publicKey, err := sqlWorkbenchService.client.GetPublicKey()
	if err != nil {
		return "", 0, fmt.Errorf("failed to get public key: %v", err)
	}

	// 使用当前用户账号登录
	loginResp, err := sqlWorkbenchService.client.Login(sqlWorkbenchService.generateSqlWorkbenchUsername(dmsUser.Name), SQL_WORKBENCH_REAL_PASSWORD, publicKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to login as user: %v", err)
	}

	// 获取组织信息
	orgResp, err := sqlWorkbenchService.client.GetOrganizations(loginResp.Cookie)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get organizations: %v", err)
	}

	// 检查是否有足够的组织
	if len(orgResp.Data.Contents) < 2 {
		return "", 0, fmt.Errorf("insufficient organizations, expected at least 2, got %d", len(orgResp.Data.Contents))
	}

	// 合并Cookie
	return sqlWorkbenchService.client.MergeCookies(orgResp.XsrfToken, loginResp.Cookie), orgResp.Data.Contents[1].ID, nil
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

// buildCreateDatasourceRequest 构建创建数据源请求
func (sqlWorkbenchService *SqlWorkbenchService) buildCreateDatasourceRequest(ctx context.Context, dbService *biz.DBService, sqlWorkbenchUser *biz.SqlWorkbenchUser, environmentID int64) (client.CreateDatasourceRequest, error) {
	datasourceName, err := sqlWorkbenchService.buildDatasourceName(ctx, dbService)
	if err != nil {
		return client.CreateDatasourceRequest{}, err
	}

	return client.CreateDatasourceRequest{
		CreatorID: sqlWorkbenchUser.SqlWorkbenchUserId,
		Type:      sqlWorkbenchService.convertDBType(dbService.DBType),
		Name:      datasourceName,
		Username:  dbService.User,
		Password:  dbService.Password,
		Host:      dbService.Host,
		Port:      dbService.Port,
		SSLConfig: client.SSLConfig{
			Enabled: false,
		},
		EnvironmentID: environmentID,
	}, nil
}

// buildUpdateDatasourceRequest 构建更新数据源请求
func (sqlWorkbenchService *SqlWorkbenchService) buildUpdateDatasourceRequest(ctx context.Context, dbService *biz.DBService, environmentID int64) (client.UpdateDatasourceRequest, error) {
	datasourceName, err := sqlWorkbenchService.buildDatasourceName(ctx, dbService)
	if err != nil {
		return client.UpdateDatasourceRequest{}, err
	}

	return client.UpdateDatasourceRequest{
		Type:     sqlWorkbenchService.convertDBType(dbService.DBType),
		Name:     &datasourceName,
		Username: dbService.User,
		Password: &dbService.Password,
		Host:     dbService.Host,
		Port:     dbService.Port,
		SSLConfig: client.SSLConfig{
			Enabled: false,
		},
		EnvironmentID: environmentID,
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
