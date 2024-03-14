package biz

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"

	maskBiz "github.com/actiontech/dms/internal/data_masking/biz"

	"github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/resolver"

	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/labstack/echo/v4"
)

const (
	CbRootUri             = "/sql_query"
	CbGqlApi              = "/api/gql"
	CBUserRole            = "user"
	CloudbeaverCookieName = "cb-session-id"
)

type CloudbeaverCfg struct {
	EnableHttps   bool   `yaml:"enable_https"`
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	AdminUser     string `yaml:"admin_user"`
	AdminPassword string `yaml:"admin_password"`
}

type CloudbeaverUser struct {
	DMSUserID         string `json:"dms_user_id"`
	DMSFingerprint    string `json:"dms_fingerprint"`
	CloudbeaverUserID string `json:"cloudbeaver_user_id"`
}

type CloudbeaverConnection struct {
	DMSDBServiceID          string `json:"dms_db_service_id"`
	Purpose                 string `json:"purpose"`
	DMSUserId               string `json:"dms_user_id"`
	DMSDBServiceFingerprint string `json:"dms_db_service_fingerprint"`
	CloudbeaverConnectionID string `json:"cloudbeaver_connection_id"`
}

func (c CloudbeaverConnection) PrimaryKey() string {
	return getDBPrimaryKey(c.DMSDBServiceID, c.Purpose, c.DMSUserId)
}

type CloudbeaverRepo interface {
	GetCloudbeaverUserByID(ctx context.Context, cloudbeaverUserId string) (*CloudbeaverUser, bool, error)
	UpdateCloudbeaverUserCache(ctx context.Context, u *CloudbeaverUser) error
	GetDbServiceIdByConnectionId(ctx context.Context, connectionId string) (string, error)
	GetAllCloudbeaverConnections(ctx context.Context) ([]*CloudbeaverConnection, error)
	GetCloudbeaverConnectionsByUserIdAndDBServiceIds(ctx context.Context, userId string, dmsDBServiceIds []string) ([]*CloudbeaverConnection, error)
	GetCloudbeaverConnectionsByUserId(ctx context.Context, userId string) ([]*CloudbeaverConnection, error)
	UpdateCloudbeaverConnectionCache(ctx context.Context, u *CloudbeaverConnection) error
	DeleteCloudbeaverConnectionCache(ctx context.Context, dbServiceId, userId, purpose string) error
}

type CloudbeaverUsecase struct {
	graphQl                   cloudbeaver.GraphQLImpl
	cloudbeaverCfg            *CloudbeaverCfg
	log                       *utilLog.Helper
	userUsecase               *UserUsecase
	dbServiceUsecase          *DBServiceUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	dmsConfigUseCase          *DMSConfigUseCase
	dataMaskingUseCase        *maskBiz.DataMaskingUseCase
	repo                      CloudbeaverRepo
	proxyTargetRepo           ProxyTargetRepo
}

func NewCloudbeaverUsecase(log utilLog.Logger, cfg *CloudbeaverCfg,
	userUsecase *UserUsecase,
	dbServiceUsecase *DBServiceUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase,
	dmsConfigUseCase *DMSConfigUseCase,
	dataMaskingUseCase *maskBiz.DataMaskingUseCase,
	cloudbeaverRepo CloudbeaverRepo,
	proxyTargetRepo ProxyTargetRepo) (cu *CloudbeaverUsecase) {
	cu = &CloudbeaverUsecase{
		repo:                      cloudbeaverRepo,
		proxyTargetRepo:           proxyTargetRepo,
		userUsecase:               userUsecase,
		dbServiceUsecase:          dbServiceUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		dmsConfigUseCase:          dmsConfigUseCase,
		dataMaskingUseCase:        dataMaskingUseCase,
		cloudbeaverCfg:            cfg,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.cloudbeaver")),
	}

	return
}

func (cu *CloudbeaverUsecase) GetRootUri() string {
	return CbRootUri
}

func (cu *CloudbeaverUsecase) IsCloudbeaverConfigured() bool {
	if cu.cloudbeaverCfg == nil {
		return false
	}

	return cu.cloudbeaverCfg.Host != "" && cu.cloudbeaverCfg.Port != "" && cu.cloudbeaverCfg.AdminUser != "" && cu.cloudbeaverCfg.AdminPassword != ""
}

func (cu *CloudbeaverUsecase) initialGraphQL() error {
	if cu.IsCloudbeaverConfigured() && cu.graphQl == nil {
		graphQl, graphQlErr := cloudbeaver.NewGraphQL(cu.getGraphQLServerURI())
		if graphQlErr != nil {
			cu.log.Errorf("NewGraphQL err: %v", graphQlErr)

			return fmt.Errorf("initial graphql client err: %v", graphQlErr)
		}

		cu.graphQl = graphQl
	}

	return nil
}

func (cu *CloudbeaverUsecase) getGraphQLServerURI() string {
	protocol := "http"
	if cu.cloudbeaverCfg.EnableHttps {
		protocol = "https"
	}

	return fmt.Sprintf("%v://%v:%v%v%v", protocol, cu.cloudbeaverCfg.Host, cu.cloudbeaverCfg.Port, CbRootUri, CbGqlApi)
}

func (cu *CloudbeaverUsecase) Login() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var dmsToken string
			for _, cookie := range c.Cookies() {
				if cookie.Name == constant.DMSToken {
					dmsToken = cookie.Value
					break
				}
			}

			if dmsToken == "" {
				return c.Redirect(http.StatusFound, "/login?target=/sql_query")
			}

			dmsUserId, err := jwt.ParseUidFromJwtTokenStr(dmsToken)
			if err != nil {
				cu.log.Errorf("GetUserUidStrFromContext err: %v", err)
				return errors.New("get user name from token failed")
			}

			if err = cu.initialGraphQL(); err != nil {
				return err
			}

			// 当前用户已经用同一个token登录过CB
			cloudbeaverSessionId := cu.getCloudbeaverSession(dmsUserId, dmsToken)
			if cloudbeaverSessionId != "" {
				cookie := &http.Cookie{Name: CloudbeaverCookieName, Value: cloudbeaverSessionId}
				c.Request().Header.Set("Cookie", fmt.Sprintf("%s=%s", CloudbeaverCookieName, cookie.Value))

				// 根据cookie 获取登录用户
				cloudbeaverActiveUser, err := cu.getActiveUserQuery([]*http.Cookie{cookie})
				if err != nil {
					cu.log.Errorf("getActiveUserQuery err: %v", err)
					return err
				}

				if cloudbeaverActiveUser.User != nil {
					return next(c)
				}
			}

			user, err := cu.userUsecase.GetUser(c.Request().Context(), dmsUserId)
			if err != nil {
				cu.log.Errorf("get user failed: %v", err)
				return fmt.Errorf("get user failed, err: %v", err)
			}

			cu.log.Infof("trigger login cloudbeaver, name: %s", user.Name)

			cloudbeaverUserId := cloudbeaver.GenerateCloudbeaverUserId(user.Name)

			if err = cu.createUserIfNotExist(c.Request().Context(), cloudbeaverUserId, user); err != nil {
				cu.log.Errorf("sync cloudbeaver user %s info failed: %v", user.Name, err)
				return err
			}

			if err = cu.connectManagement(c.Request().Context(), cloudbeaverUserId, user); err != nil {
				cu.log.Errorf("sync cloudbeaver user %s bind instance info failed: %v", user.Name, err)
				return err
			}

			cookies, err := cu.loginCloudbeaverServer(cloudbeaverUserId, user.Password)
			if err != nil {
				cu.log.Errorf("login to cloudbeaver failed: %v", err)
				return err
			}

			for _, cookie := range cookies {
				if cookie.Name == CloudbeaverCookieName {
					cu.setCloudbeaverSession(user.UID, dmsToken, cookie.Value)
					c.Request().AddCookie(cookie)
				}
			}

			return next(c)
		}
	}
}

type responseProcessWriter struct {
	tmp        *bytes.Buffer
	headerCode int
	http.ResponseWriter
}

func (w *responseProcessWriter) WriteHeader(code int) {
	w.headerCode = code
}

func (w *responseProcessWriter) Write(b []byte) (int, error) {
	return w.tmp.Write(b)
}

func (w *responseProcessWriter) Flush() {
	if wf, ok := w.ResponseWriter.(http.Flusher); ok {
		wf.Flush()
	}
}

func (w *responseProcessWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if wh, ok := w.ResponseWriter.(http.Hijacker); ok {
		return wh.Hijack()
	}

	return nil, nil, errors.New("responseProcessWriter assert Hijacker failed")
}

func (cu *CloudbeaverUsecase) getSQLEUrl(ctx context.Context) (string, error) {
	target, err := cu.proxyTargetRepo.GetProxyTargetByName(ctx, cloudbeaver.SQLEProxyName)
	if err != nil {
		return "", err
	}

	return target.URL.String(), nil
}

func (cu *CloudbeaverUsecase) GraphQLDistributor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.Request().RequestURI != path.Join(CbRootUri, CbGqlApi) {
				return next(c)
			}
			// copy request body
			var reqBody = make([]byte, 0)
			if c.Request().Body != nil { // Read
				reqBody, err = io.ReadAll(c.Request().Body)

				if err != nil {
					cu.log.Errorf("read request body err: %v", err)
					return err
				}
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset

			var params *graphql.RawParams
			err = json.Unmarshal(reqBody, &params)
			if err != nil {
				cu.log.Errorf("graphql.RawParams json unmarshal err: %v", err)
				return err
			}

			cloudbeaverHandle, ok := cloudbeaver.GraphQLHandlerRouters[params.OperationName]
			if !ok {
				return next(c)
			}

			if cloudbeaverHandle.Disable {
				message := "this feature is prohibited"
				cu.log.Errorf("%v:%v", message, params.OperationName)
				return c.JSON(http.StatusOK, model.ServerError{
					Message: &message,
				})
			}

			if cloudbeaverHandle.Preprocessing != nil {
				if err = cloudbeaverHandle.Preprocessing(c, params); err != nil {
					cu.log.Error(err)
					return err
				}
			}

			if cloudbeaverHandle.UseLocalHandler {
				ctx := graphql.StartOperationTrace(context.Background())
				if params.OperationName == "asyncSqlExecuteQuery" {
					isEnableSqlAudit, err := cu.isEnableSQLAudit(c.Request().Context(), params)
					if err != nil {
						cu.log.Error(err)
						return err
					}

					if !isEnableSqlAudit {
						return next(c)
					}

					sqleUrl, err := cu.getSQLEUrl(c.Request().Context())
					if err != nil {
						return err
					}

					dbService, err := cu.getDbService(c.Request().Context(), params)
					if err != nil {
						return err
					}

					directAuditReq := cloudbeaver.DirectAuditParams{
						AuditSQLReq: cloudbeaver.AuditSQLReq{
							InstanceType:     dbService.DBType,
							ProjectId:        dbService.ProjectUID,
							RuleTemplateName: dbService.SQLEConfig.RuleTemplateName,
						},
						SQLEAddr: fmt.Sprintf("%s/v2/sql_audit", sqleUrl),
					}

					// pass sqle direct audit params
					ctx = context.WithValue(ctx, cloudbeaver.SQLEDirectAudit, directAuditReq)
				}

				if params.OperationName == "getSqlExecuteTaskResults" {
					isEnableMasking, err := cu.IsEnableDataMasking(c.Request().Context())
					if err != nil {
						cu.log.Error(err)
						return err
					}

					if !isEnableMasking {
						return next(c)
					}
				}

				params.ReadTime = graphql.TraceTiming{
					Start: graphql.Now(),
					End:   graphql.Now(),
				}

				params.Headers = c.Request().Header.Clone()

				var cloudbeaverNext cloudbeaver.Next
				var resWrite *responseProcessWriter
				if !cloudbeaverHandle.NeedModifyRemoteRes {
					cloudbeaverNext = func(c echo.Context) ([]byte, error) {
						return nil, next(c)
					}
				} else {
					cloudbeaverNext = func(c echo.Context) ([]byte, error) {
						resWrite = &responseProcessWriter{tmp: &bytes.Buffer{}, ResponseWriter: c.Response().Writer}
						c.Response().Writer = resWrite

						if err = next(c); err != nil {
							return nil, err
						}

						return resWrite.tmp.Bytes(), nil
					}
				}

				g := resolver.NewExecutableSchema(resolver.Config{
					Resolvers: cloudbeaver.NewResolverImpl(c, cloudbeaverNext, cu.SQLExecuteResultsDataMasking),
				})

				exec := executor.New(g)

				rc, err := exec.CreateOperationContext(ctx, params)
				if err != nil {
					return err
				}
				responses, ctx := exec.DispatchOperation(ctx, rc)

				res := responses(ctx)
				if res.Errors.Error() != "" {
					return res.Errors
				}
				if !cloudbeaverHandle.NeedModifyRemoteRes {
					return nil
				} else {
					header := resWrite.ResponseWriter.Header()
					b, err := json.Marshal(res)
					if err != nil {
						return err
					}
					header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
					_, err = resWrite.ResponseWriter.Write(b)
					return err
				}
			}
			return next(c)
		}
	}
}

func (cu *CloudbeaverUsecase) IsEnableDataMasking(ctx context.Context) (bool, error) {
	return cu.dmsConfigUseCase.IsEnableSQLResultsDataMasking(ctx)
}

func (cu *CloudbeaverUsecase) isEnableSQLAudit(ctx context.Context, params *graphql.RawParams) (bool, error) {
	dbService, err := cu.getDbService(ctx, params)
	if err != nil {
		return false, err
	}

	return dbService.SQLEConfig.SQLQueryConfig.AuditEnabled, nil
}

func (cu *CloudbeaverUsecase) getDbService(ctx context.Context, params *graphql.RawParams) (*DBService, error) {
	var connectionId interface{}
	var connectionIdStr string
	var ok bool

	connectionId, ok = params.Variables["connectionId"]
	if !ok {
		return nil, fmt.Errorf("missing connectionId in %s query", params.OperationName)
	}

	connectionIdStr, ok = connectionId.(string)
	if !ok {
		return nil, fmt.Errorf("connectionId %s convert failed", connectionId)
	}
	dbServiceId, err := cu.repo.GetDbServiceIdByConnectionId(ctx, connectionIdStr)
	if err != nil {
		return nil, err
	}

	return cu.dbServiceUsecase.GetDBService(ctx, dbServiceId)
}

type ActiveUserQueryRes struct {
	User interface{} `json:"user"`
}

func (cu *CloudbeaverUsecase) getActiveUserQuery(cookies []*http.Cookie) (*ActiveUserQueryRes, error) {
	client := cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithCookie(cookies))
	req := cloudbeaver.NewRequest(cu.graphQl.GetActiveUserQuery(), map[string]interface{}{})

	res := &ActiveUserQueryRes{}
	if err := client.Run(context.TODO(), req, res); err != nil {
		return nil, err
	}
	return res, nil
}

type cloudbeaverSession struct {
	dmsToken             string
	cloudbeaverSessionId string
}

var (
	dmsUserIdCloudbeaverLoginMap = make(map[string]cloudbeaverSession)
	tokenMapMutex                = &sync.Mutex{}
)

func (cu *CloudbeaverUsecase) getCloudbeaverSession(dmsUserId, dmsToken string) string {
	tokenMapMutex.Lock()
	defer tokenMapMutex.Unlock()

	if item, ok := dmsUserIdCloudbeaverLoginMap[dmsUserId]; ok {
		if dmsToken == item.dmsToken {
			return item.cloudbeaverSessionId
		}
	}

	return ""
}

func (cu *CloudbeaverUsecase) UnbindCBSession(token string) {
	tokenMapMutex.Lock()
	defer tokenMapMutex.Unlock()
	delete(dmsUserIdCloudbeaverLoginMap, token)
}

func (cu *CloudbeaverUsecase) setCloudbeaverSession(dmsUserId, dmsToken, cloudbeaverSessionId string) {
	tokenMapMutex.Lock()
	defer tokenMapMutex.Unlock()

	dmsUserIdCloudbeaverLoginMap[dmsUserId] = cloudbeaverSession{
		dmsToken:             dmsToken,
		cloudbeaverSessionId: cloudbeaverSessionId,
	}
}

type UserList struct {
	ListUsers []struct {
		UserID string `json:"userID"`
	} `json:"listUsers"`
}

func (cu *CloudbeaverUsecase) createUserIfNotExist(ctx context.Context, cloudbeaverUserId string, dmsUser *User) error {
	cloudbeaverUser, exist, err := cu.repo.GetCloudbeaverUserByID(ctx, cloudbeaverUserId)
	if err != nil {
		return err
	}

	fingerprint := cu.userUsecase.GetUserFingerprint(dmsUser)
	if exist && cloudbeaverUser.DMSFingerprint == fingerprint {
		return nil
	}

	reservedCloudbeaverUserId := map[string]struct{}{"admin": {}, "user": {}}
	if _, ok := reservedCloudbeaverUserId[cloudbeaverUserId]; ok {
		return fmt.Errorf("this username cannot be used")
	}

	// 使用管理员身份登录
	graphQLClient, err := cu.getGraphQLClientWithRootUser()
	if err != nil {
		return err
	}

	checkExistReq := cloudbeaver.NewRequest(cu.graphQl.IsUserExistQuery(cloudbeaverUserId))

	cloudbeaverUserList := UserList{}
	err = graphQLClient.Run(ctx, checkExistReq, &cloudbeaverUserList)
	if err != nil {
		return fmt.Errorf("check cloudbeaver user exist failed: %v", err)
	}

	// 用户不存在则创建CloudBeaver用户
	if len(cloudbeaverUserList.ListUsers) == 0 {
		// 创建用户
		createUserReq := cloudbeaver.NewRequest(cu.graphQl.CreateUserQuery(), map[string]interface{}{
			"userId": cloudbeaverUserId,
		})
		err = graphQLClient.Run(ctx, createUserReq, &UserList{})
		if err != nil {
			return fmt.Errorf("create cloudbeaver user failed: %v", err)
		}

		// 授予角色(不授予角色的用户无法登录)
		grantUserRoleReq := cloudbeaver.NewRequest(cu.graphQl.GrantUserRoleQuery(), map[string]interface{}{
			"userId": cloudbeaverUserId,
			"roleId": CBUserRole,
			"teamId": CBUserRole,
		})
		err = graphQLClient.Run(ctx, grantUserRoleReq, nil)
		if err != nil {
			return fmt.Errorf("grant cloudbeaver user failed: %v", err)
		}
	}

	// 设置CloudBeaver用户密码
	updatePasswordReq := cloudbeaver.NewRequest(cu.graphQl.UpdatePasswordQuery(), map[string]interface{}{
		"userId": cloudbeaverUserId,
		"credentials": model.JSON{
			"password": strings.ToUpper(aes.Md5(dmsUser.Password)),
		},
	})
	err = graphQLClient.Run(ctx, updatePasswordReq, nil)
	if err != nil {
		return fmt.Errorf("update cloudbeaver user failed: %v", err)
	}

	cloudbeaverUser = &CloudbeaverUser{
		DMSUserID:         dmsUser.UID,
		DMSFingerprint:    fingerprint,
		CloudbeaverUserID: cloudbeaverUserId,
	}

	return cu.repo.UpdateCloudbeaverUserCache(ctx, cloudbeaverUser)
}

func (cu *CloudbeaverUsecase) connectManagement(ctx context.Context, cloudbeaverUserId string, dmsUser *User) error {
	activeDBServices, err := cu.dbServiceUsecase.GetActiveDBServices(ctx, nil)
	if err != nil {
		return err
	}

	if len(activeDBServices) == 0 {
		return cu.clearConnection(ctx)
	}

	isAdmin, err := cu.opPermissionVerifyUsecase.IsUserDMSAdmin(ctx, dmsUser.UID)
	if err != nil {
		return err
	}

	if !isAdmin {
		activeDBServices, err = cu.ResetDbServiceByAuth(ctx, activeDBServices, dmsUser.UID)
		if err != nil {
			return err
		}
		opPermissions, err := cu.opPermissionVerifyUsecase.GetUserOpPermission(ctx, dmsUser.UID)
		if err != nil {
			return err
		}

		// 已配置的项目管理权限和数据源工作台查询权限
		projectIdMap := map[string]struct{}{}
		dbServiceIdMap := map[string]struct{}{}
		for _, opPermission := range opPermissions {
			// project permission
			if opPermission.OpRangeType == OpRangeTypeProject && opPermission.OpPermissionUID == constant.UIDOfOpPermissionProjectAdmin {
				for _, rangeUid := range opPermission.RangeUIDs {
					projectIdMap[rangeUid] = struct{}{}
				}
			}

			// db_service permission
			if opPermission.OpRangeType == OpRangeTypeDBService && opPermission.OpPermissionUID == constant.UIDOfOpPermissionSQLQuery {
				for _, rangeUid := range opPermission.RangeUIDs {
					dbServiceIdMap[rangeUid] = struct{}{}
				}
			}
		}

		var lastActiveDBServices []*DBService
		for _, activeDBService := range activeDBServices {
			if _, ok := projectIdMap[activeDBService.ProjectUID]; ok {
				lastActiveDBServices = append(lastActiveDBServices, activeDBService)
				continue
			}

			if _, ok := dbServiceIdMap[activeDBService.UID]; ok {
				lastActiveDBServices = append(lastActiveDBServices, activeDBService)
			}
		}

		activeDBServices = lastActiveDBServices
	}

	if err = cu.operateConnection(ctx, activeDBServices, dmsUser.UID); err != nil {
		return err
	}

	cloudbeaverUser, exist, err := cu.repo.GetCloudbeaverUserByID(ctx, cloudbeaverUserId)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("cloudbeaver user: %s not eixst", cloudbeaverUserId)
	}
	if err = cu.grantAccessConnection(ctx, cloudbeaverUser, dmsUser, activeDBServices); err != nil {
		return err
	}

	return nil
}

// 判断连接是否唯一的条件：dbServiceId : purpose : userUid
func getDBPrimaryKey(dbUid, purpose, userUid string) string {
	// service.UID:service.AccountPurpose:userId
	return fmt.Sprint(dbUid, ":", purpose, ":", userUid)
}

func (cu *CloudbeaverUsecase) operateConnection(ctx context.Context, activeDBServices []*DBService, userId string) error {
	dbServiceMap := map[string]*DBService{}
	projectMap := map[string]string{}
	for _, service := range activeDBServices {
		dbServiceMap[getDBPrimaryKey(service.UID, service.AccountPurpose, userId)] = service

		project, err := cu.dbServiceUsecase.projectUsecase.GetProject(ctx, service.ProjectUID)
		if err != nil {
			projectMap[service.UID] = "unknown"
			cu.log.Errorf("get db service project %s failed, err: %v", service.ProjectUID, err)
		} else {
			projectMap[service.UID] = project.Name
		}
	}

	//获取当前用户所有已创建的连接
	cloudbeaverConnections, err := cu.repo.GetCloudbeaverConnectionsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	var deleteConnections []*CloudbeaverConnection

	cloudbeaverConnectionMap := map[string]*CloudbeaverConnection{}
	for _, connection := range cloudbeaverConnections {
		// 删除用户关联的连接
		if connection.DMSUserId == userId {
			cloudbeaverConnectionMap[connection.PrimaryKey()] = connection
			if _, ok := dbServiceMap[connection.PrimaryKey()]; !ok {
				deleteConnections = append(deleteConnections, connection)
			}
		}
	}

	createConnections, updateConnections := []*CloudbeaverConnection{}, []*CloudbeaverConnection{}

	for _, dbService := range dbServiceMap {
		if cloudbeaverConnection, ok := cloudbeaverConnectionMap[getDBPrimaryKey(dbService.UID, dbService.AccountPurpose, userId)]; ok {
			if cloudbeaverConnection.DMSDBServiceFingerprint != cu.dbServiceUsecase.GetDBServiceFingerprint(dbService) {
				updateConnections = append(updateConnections, &CloudbeaverConnection{DMSDBServiceID: dbService.UID, Purpose: dbService.AccountPurpose, DMSUserId: userId})
			}
		} else {
			createConnections = append(createConnections, &CloudbeaverConnection{DMSDBServiceID: dbService.UID, Purpose: dbService.AccountPurpose, DMSUserId: userId})
		}
	}

	if len(createConnections) == 0 && len(updateConnections) == 0 && len(deleteConnections) == 0 {
		return nil
	}

	// 获取管理员链接
	cloudbeaverClient, err := cu.getGraphQLClientWithRootUser()
	if err != nil {
		return err
	}

	// 同步实例连接信息
	for _, createConnection := range createConnections {
		if err = cu.createCloudbeaverConnection(ctx, cloudbeaverClient, dbServiceMap[getDBPrimaryKey(createConnection.DMSDBServiceID, createConnection.Purpose, userId)],
			projectMap[createConnection.DMSDBServiceID], userId); err != nil {
			cu.log.Errorf("create connection %v failed: %v", createConnection, err)
		}
	}

	for _, updateConnection := range updateConnections {
		if err = cu.updateCloudbeaverConnection(ctx, cloudbeaverClient, updateConnection.CloudbeaverConnectionID, dbServiceMap[getDBPrimaryKey(updateConnection.DMSDBServiceID, updateConnection.Purpose, userId)], projectMap[updateConnection.DMSDBServiceID], userId); err != nil {
			cu.log.Errorf("update dnServerId %s to connection failed: %v", updateConnection, err)
		}
	}

	for _, deleteConnection := range deleteConnections {
		if err = cu.deleteCloudbeaverConnection(ctx, cloudbeaverClient, deleteConnection.CloudbeaverConnectionID, deleteConnection.DMSDBServiceID, userId, deleteConnection.Purpose); err != nil {
			cu.log.Errorf("delete connection %v failed: %v", deleteConnection, err)
		}
	}

	return nil
}

// 删除DMS已知待删除的连接
func (cu *CloudbeaverUsecase) clearConnection(ctx context.Context) error {
	cloudbeaverConnections, err := cu.repo.GetAllCloudbeaverConnections(ctx)
	if err != nil {
		return err
	}

	// 获取管理员链接
	cloudbeaverClient, err := cu.getGraphQLClientWithRootUser()
	if err != nil {
		return err
	}

	for _, item := range cloudbeaverConnections {
		if err = cu.deleteCloudbeaverConnection(ctx, cloudbeaverClient, item.CloudbeaverConnectionID, item.DMSDBServiceID, "", ""); err != nil {
			cu.log.Errorf("delete dbServerId %s to connection failed: %v", item.DMSDBServiceID, err)

			return fmt.Errorf("delete dbServerId %s to connection failed: %v", item.DMSDBServiceID, err)
		}
	}

	return nil
}

func (cu *CloudbeaverUsecase) grantAccessConnection(ctx context.Context, cloudbeaverUser *CloudbeaverUser, dmsUser *User, activeDBServices []*DBService) error {
	if cloudbeaverUser.DMSFingerprint != cu.userUsecase.GetUserFingerprint(dmsUser) {
		return fmt.Errorf("user information is not synchronized, unable to update connection information")
	}

	// 清空绑定能访问的数据库连接
	if len(activeDBServices) == 0 {
		return cu.bindUserAccessConnection(ctx, []*CloudbeaverConnection{}, cloudbeaverUser.CloudbeaverUserID)
	}

	dbServiceIds := make([]string, 0, len(activeDBServices))
	for _, dbService := range activeDBServices {
		dbServiceIds = append(dbServiceIds, dbService.UID)
	}
	cloudbeaverConnections, err := cu.repo.GetCloudbeaverConnectionsByUserIdAndDBServiceIds(ctx, dmsUser.UID, dbServiceIds)
	if err != nil {
		return err
	}

	// 从缓存中获取需要同步的CloudBeaver实例
	cloudbeaverConnectionMap := map[string]*CloudbeaverConnection{}
	for _, cloudbeaverConnection := range cloudbeaverConnections {
		cloudbeaverConnectionMap[cloudbeaverConnection.CloudbeaverConnectionID] = cloudbeaverConnection
	}

	// 获取用户当前实例列表
	connResp := &struct {
		Connections []*struct {
			Id string `json:"id"`
		} `json:"connections"`
	}{}

	client, err := cu.getGraphQLClient(cloudbeaverUser.CloudbeaverUserID, dmsUser.Password)
	if err != nil {
		return err
	}

	err = client.Run(ctx, cloudbeaver.NewRequest(cu.graphQl.GetUserConnectionsQuery(), nil), connResp)
	if err != nil {
		return err
	}

	if len(connResp.Connections) != len(cloudbeaverConnections) {
		return cu.bindUserAccessConnection(ctx, cloudbeaverConnections, cloudbeaverUser.CloudbeaverUserID)
	}

	for _, connection := range connResp.Connections {
		if _, ok := cloudbeaverConnectionMap[connection.Id]; !ok {
			return cu.bindUserAccessConnection(ctx, cloudbeaverConnections, cloudbeaverUser.CloudbeaverUserID)
		}
	}

	return nil
}

func (cu *CloudbeaverUsecase) bindUserAccessConnection(ctx context.Context, cloudbeaverDBServices []*CloudbeaverConnection, cloudBeaverUserID string) error {
	var cloudbeaverConnectionIds = make([]string, 0, len(cloudbeaverDBServices))
	for _, service := range cloudbeaverDBServices {
		cloudbeaverConnectionIds = append(cloudbeaverConnectionIds, service.CloudbeaverConnectionID)
	}

	cloudbeaverConnReq := cloudbeaver.NewRequest(cu.graphQl.SetUserConnectionsQuery(), map[string]interface{}{
		"userId":      cloudBeaverUserID,
		"connections": cloudbeaverConnectionIds,
	})

	rootClient, err := cu.getGraphQLClientWithRootUser()
	if err != nil {
		return err
	}

	return rootClient.Run(ctx, cloudbeaverConnReq, nil)
}

func (cu *CloudbeaverUsecase) createCloudbeaverConnection(ctx context.Context, client *cloudbeaver.Client, dbService *DBService, project, userId string) error {
	params, err := cu.GenerateCloudbeaverConnectionParams(dbService, project, dbService.AccountPurpose)
	if err != nil {
		return fmt.Errorf("%s unsupported", dbService.DBType)
	}

	// 添加实例
	req := cloudbeaver.NewRequest(cu.graphQl.CreateConnectionQuery(), params)
	resp := struct {
		Connection struct {
			ID string `json:"id"`
		} `json:"connection"`
	}{}

	err = client.Run(ctx, req, &resp)
	if err != nil {
		return err
	}

	// 同步缓存
	return cu.repo.UpdateCloudbeaverConnectionCache(ctx, &CloudbeaverConnection{
		DMSDBServiceID:          dbService.UID,
		DMSUserId:               userId,
		DMSDBServiceFingerprint: cu.dbServiceUsecase.GetDBServiceFingerprint(dbService),
		Purpose:                 dbService.AccountPurpose,
		CloudbeaverConnectionID: resp.Connection.ID,
	})
}

// UpdateCloudbeaverConnection 更新完毕后会同步缓存
func (cu *CloudbeaverUsecase) updateCloudbeaverConnection(ctx context.Context, client *cloudbeaver.Client, cloudbeaverConnectionId string, dbService *DBService, project, userId string) error {
	params, err := cu.GenerateCloudbeaverConnectionParams(dbService, project, dbService.AccountPurpose)
	if err != nil {
		return fmt.Errorf("%s unsupported", dbService.DBType)
	}

	config, ok := params["config"].(map[string]interface{})
	if !ok {
		return errors.New("assert connection params failed")
	}

	config["connectionId"] = cloudbeaverConnectionId
	params["config"] = config
	req := cloudbeaver.NewRequest(cu.graphQl.UpdateConnectionQuery(), params)
	resp := struct {
		Connection struct {
			ID string `json:"id"`
		} `json:"connection"`
	}{}

	err = client.Run(ctx, req, &resp)
	if err != nil {
		return err
	}

	return cu.repo.UpdateCloudbeaverConnectionCache(ctx, &CloudbeaverConnection{
		DMSDBServiceID:          dbService.UID,
		DMSUserId:               userId,
		DMSDBServiceFingerprint: cu.dbServiceUsecase.GetDBServiceFingerprint(dbService),
		Purpose:                 dbService.AccountPurpose,
		CloudbeaverConnectionID: resp.Connection.ID,
	})
}

func (cu *CloudbeaverUsecase) deleteCloudbeaverConnection(ctx context.Context, client *cloudbeaver.Client, cloudbeaverConnectionId, dbServiceId, userId, purpose string) error {
	variables := make(map[string]interface{})
	variables["connectionId"] = cloudbeaverConnectionId
	variables["projectId"] = cloudbeaverProjectId

	req := cloudbeaver.NewRequest(cu.graphQl.DeleteConnectionQuery(), variables)
	resp := struct {
		DeleteConnection bool `json:"deleteConnection"`
	}{}

	if err := client.Run(ctx, req, &resp); err != nil {
		return err
	}

	return cu.repo.DeleteCloudbeaverConnectionCache(ctx, dbServiceId, userId, purpose)
}

func (cu *CloudbeaverUsecase) generateCommonCloudbeaverConfigParams(dbService *DBService, project, purpose string) map[string]interface{} {
	name := fmt.Sprintf("%s:%s", project, dbService.Name)
	if purpose != "" {
		name = fmt.Sprintf("%s:%s", name, purpose)
	}
	return map[string]interface{}{
		"configurationType": "MANUAL",
		"name":              name,
		"template":          false,
		"host":              dbService.Host,
		"port":              dbService.Port,
		"databaseName":      nil,
		"description":       nil,
		"authModelId":       "native",
		"saveCredentials":   true,
		"credentials": map[string]interface{}{
			"userName":     dbService.User,
			"userPassword": dbService.Password,
		},
	}
}

const cloudbeaverProjectId = "g_GlobalConfiguration"

func (cu *CloudbeaverUsecase) GenerateCloudbeaverConnectionParams(dbService *DBService, project string, purpose string) (map[string]interface{}, error) {
	var err error
	config := cu.generateCommonCloudbeaverConfigParams(dbService, project, purpose)

	dbType, err := constant.ParseDBType(dbService.DBType)
	if err != nil {
		return nil, err
	}
	switch dbType {
	case constant.DBTypeMySQL, constant.DBTypeTDSQLForInnoDB:
		err = cu.fillMySQLParams(config)
	case constant.DBTypeTiDB:
		err = cu.fillTiDBParams(config)
	case constant.DBTypePostgreSQL:
		err = cu.fillPGSQLParams(config)
	case constant.DBTypeSQLServer:
		err = cu.fillMSSQLParams(config)
	case constant.DBTypeOracle:
		err = cu.fillOracleParams(dbService, config)
	case constant.DBTypeDB2:
		err = cu.fillDB2Params(dbService, config)
	case constant.DBTypeOceanBaseMySQL:
		err = cu.fillOceanBaseParams(dbService, config)
	case constant.DBTypeGoldenDB:
		err = cu.fillGoldenDBParams(config)
	default:
		return nil, fmt.Errorf("temporarily unsupported instance types")
	}

	resp := map[string]interface{}{
		"projectId": cloudbeaverProjectId,
		"config":    config,
	}
	return resp, err
}

func (cu *CloudbeaverUsecase) fillMySQLParams(config map[string]interface{}) error {
	config["driverId"] = "mysql:mysql8"
	return nil
}

func (cu *CloudbeaverUsecase) fillTiDBParams(config map[string]interface{}) error {
	config["driverId"] = "mysql:tidb"
	return nil
}

func (cu *CloudbeaverUsecase) fillMSSQLParams(config map[string]interface{}) error {
	config["driverId"] = "sqlserver:microsoft"
	config["authModelId"] = "sqlserver_database"
	return nil
}

func (cu *CloudbeaverUsecase) fillPGSQLParams(config map[string]interface{}) error {
	config["driverId"] = "postgresql:postgres-jdbc"
	config["databaseName"] = "postgres"
	config["providerProperties"] = map[string]interface{}{
		"@dbeaver-show-non-default-db@": true,
		"@dbeaver-show-unavailable-db@": true,
		"@dbeaver-show-template-db@":    true,
	}
	return nil
}

func (cu *CloudbeaverUsecase) fillOracleParams(inst *DBService, config map[string]interface{}) error {
	serviceName := inst.AdditionalParams.GetParam("service_name")
	if serviceName == nil {
		return fmt.Errorf("the service name of oracle cannot be empty")
	}

	config["driverId"] = "oracle:oracle_thin"
	config["authModelId"] = "oracle_native"
	config["databaseName"] = serviceName.Value
	config["providerProperties"] = map[string]interface{}{
		"@dbeaver-sid-service@": "SID",
		"oracle.logon-as":       "Normal",
	}
	return nil
}

func (cu *CloudbeaverUsecase) fillDB2Params(inst *DBService, config map[string]interface{}) error {
	dbName := inst.AdditionalParams.GetParam("database_name")
	if dbName == nil {
		return fmt.Errorf("the database name of DB2 cannot be empty")
	}

	config["driverId"] = "db2:db2"
	config["databaseName"] = dbName.Value
	return nil
}

func (cu *CloudbeaverUsecase) fillOceanBaseParams(inst *DBService, config map[string]interface{}) error {
	tenant := inst.AdditionalParams.GetParam("tenant_name")
	if tenant == nil {
		return fmt.Errorf("the tenant name of oceanbase cannot be empty")
	}

	config["driverId"] = "oceanbase:alipay_oceanbase"
	config["authModelId"] = "oceanbase_native"

	credentials := config["credentials"]
	credentialConfig, ok := credentials.(map[string]interface{})
	if !ok {
		return errors.New("assert oceanbase connection params failed")
	}
	credentialConfig["userName"] = fmt.Sprintf("%v@%v", inst.User, tenant)
	config["credentials"] = credentialConfig
	return nil
}

func (cu *CloudbeaverUsecase) fillGoldenDBParams(config map[string]interface{}) error {
	config["driverId"] = "mysql:mysql8"
	return nil
}

func (cu *CloudbeaverUsecase) getGraphQLClientWithRootUser() (*cloudbeaver.Client, error) {
	cookies, err := cu.loginCloudbeaverServer(cu.cloudbeaverCfg.AdminUser, cu.cloudbeaverCfg.AdminPassword)
	if err != nil {
		return nil, err
	}

	return cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithCookie(cookies)), nil
}

// 这个客户端会用指定用户操作, 请求会直接发到CB
func (cu *CloudbeaverUsecase) getGraphQLClient(username, password string) (*cloudbeaver.Client, error) {
	cookies, err := cu.loginCloudbeaverServer(username, password)
	if err != nil {
		return nil, err
	}

	return cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithCookie(cookies)), nil
}

func (cu *CloudbeaverUsecase) loginCloudbeaverServer(user, pwd string) (cookie []*http.Cookie, err error) {
	client := cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithHttpResHandler(
		func(response *http.Response) {
			if response != nil {
				cookie = response.Cookies()
			}
		}))
	req := cloudbeaver.NewRequest(cu.graphQl.LoginQuery(), map[string]interface{}{
		"credentials": model.JSON{
			"user":     user,
			"password": strings.ToUpper(aes.Md5(pwd)), // the password is an all-caps md5-32 string
		},
	})

	res := struct {
		AuthInfo struct {
			AuthId interface{} `json:"authId"`
		} `json:"authInfo"`
	}{}
	if err = client.Run(context.TODO(), req, &res); err != nil {
		return cookie, fmt.Errorf("cloudbeaver login failed: %v,req: %v,res %v", err, req, res)
	}

	return cookie, nil
}
