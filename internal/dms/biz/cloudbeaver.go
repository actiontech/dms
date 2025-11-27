package biz

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/resolver"
	"github.com/actiontech/dms/internal/pkg/locale"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
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
	DeleteAllCloudbeaverCachesByUserId(ctx context.Context, userId string) error
}

type CloudbeaverUsecase struct {
	graphQl                   cloudbeaver.GraphQLImpl
	cloudbeaverCfg            *CloudbeaverCfg
	log                       *utilLog.Helper
	userUsecase               *UserUsecase
	dbServiceUsecase          *DBServiceUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	dmsConfigUseCase          *DMSConfigUseCase
	dataMaskingUseCase        *DataMaskingUsecase
	cbOperationLogUsecase     *CbOperationLogUsecase
	projectUsecase            *ProjectUsecase
	repo                      CloudbeaverRepo
	proxyTargetRepo           ProxyTargetRepo
}

func NewCloudbeaverUsecase(log utilLog.Logger, cfg *CloudbeaverCfg, userUsecase *UserUsecase, dbServiceUsecase *DBServiceUsecase, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, dmsConfigUseCase *DMSConfigUseCase, dataMaskingUseCase *DataMaskingUsecase, cloudbeaverRepo CloudbeaverRepo, proxyTargetRepo ProxyTargetRepo, cbOperationUseDase *CbOperationLogUsecase, projectUsecase *ProjectUsecase) (cu *CloudbeaverUsecase) {
	cu = &CloudbeaverUsecase{
		repo:                      cloudbeaverRepo,
		proxyTargetRepo:           proxyTargetRepo,
		userUsecase:               userUsecase,
		dbServiceUsecase:          dbServiceUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		dmsConfigUseCase:          dmsConfigUseCase,
		dataMaskingUseCase:        dataMaskingUseCase,
		cbOperationLogUsecase:     cbOperationUseDase,
		projectUsecase:            projectUsecase,
		cloudbeaverCfg:            cfg,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.cloudbeaver")),
	}

	// 启动缓存清理协程
	go cu.startCacheCleanupRoutine()

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

const (
	dmsUserIdKey = "dmsToken"
	CBErrorCode  = "sessionExpired"
)

func (cu *CloudbeaverUsecase) Login() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// 从Cookie中获取DMS token
			var dmsToken string
			for _, cookie := range c.Cookies() {
				if cookie.Name == constant.DMSToken {
					dmsToken = cookie.Value
					break
				}
			}

			// 普通非WebSocket的GET请求跳过token校验和cookie设置逻辑
			if c.Request().Method == http.MethodGet && !c.IsWebSocket() {
				cu.log.Debugf("Cloudbeaver login middleware: Skipping GET request to %s", c.Request().RequestURI)
				return next(c)
			}

			if dmsToken == "" {
				gqlResp := GraphQLResponse{
					Data: nil,
					Errors: []GraphQLError{{
						Extensions: map[string]interface{}{
							"webErrorCode": CBErrorCode,
						},
						Message: "dms user token is empty",
					}},
				}

				cu.log.Errorf("dmsToken is empty")
				return c.JSON(http.StatusOK, gqlResp)
			}

			dmsUserId, err := jwt.ParseUidFromJwtTokenStr(dmsToken)
			if err != nil {
				gqlResp := GraphQLResponse{
					Data: nil,
					Errors: []GraphQLError{{
						Extensions: map[string]interface{}{
							"webErrorCode": CBErrorCode,
						},
						Message: "dms user token expired",
					}},
				}

				cu.log.Errorf("GetUserUidStrFromContext err: %v", err)
				return c.JSON(http.StatusOK, gqlResp)
			}
			// set dmsUserId to context for save ob operation log
			c.Set(dmsUserIdKey, dmsUserId)
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

			// 获取认证 cookies（已包含缓存逻辑）
			cookies, err := cu.getAuthCookies(cloudbeaverUserId, user.Password)
			if err != nil {
				cu.log.Errorf("CloudBeaver authentication failed: %v", err)
				return err
			}

			for _, cookie := range cookies {
				if cookie.Name == CloudbeaverCookieName {
					cu.setCloudbeaverSession(user.UID, dmsToken, cookie.Value)
					SetOrReplaceCBCookieByDMSToken(c, cookie)
				}
			}

			return next(c)
		}
	}
}

// SetOrReplaceCBCookieByDMSToken sets or replaces a specific cookie in the request header of an echo.Context.
//
// Example:
//
//	cookie = token=abc123
//
//	Replace existing cookie:
//	  before: "sessionid=xyz789; token=oldval; lang=zh"
//	  after:  "sessionid=xyz789; token=abc123; lang=zh"
//
//	Set new cookie (when "token" does not exist):
//	  before: "sessionid=xyz789; lang=zh"
//	  after:  "sessionid=xyz789; lang=zh; token=abc123"
func SetOrReplaceCBCookieByDMSToken(c echo.Context, cookie *http.Cookie) {
	req := c.Request()

	// Get the original "Cookie" header, e.g., "a=1; b=2"
	original := req.Header.Get("Cookie")
	pairs := []string{}
	found := false

	// Split the cookie string into individual name-value pairs
	for _, segment := range strings.Split(original, ";") {
		pair := strings.SplitN(strings.TrimSpace(segment), "=", 2)
		if len(pair) != 2 {
			continue // Skip malformed cookie segments
		}

		if pair[0] == cookie.Name {
			// Replace the value if the target cookie is found
			pairs = append(pairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
			found = true
		} else {
			// Preserve other cookies
			pairs = append(pairs, fmt.Sprintf("%s=%s", pair[0], pair[1]))
		}
	}

	// If the target cookie was not found, append it as a new entry
	if !found {
		pairs = append(pairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	// Set the updated "Cookie" header back to the request
	req.Header.Set("Cookie", strings.Join(pairs, "; "))
}

func (cu *CloudbeaverUsecase) getSQLEUrl(ctx context.Context) (string, error) {
	target, err := cu.proxyTargetRepo.GetProxyTargetByName(ctx, cloudbeaver.SQLEProxyName)
	if err != nil {
		return "", err
	}

	return target.URL.String(), nil
}

type TaskInfo struct {
	Data struct {
		TaskInfo *model.AsyncTaskInfo `json:"taskInfo"`
	} `json:"data"`
}

var (
	taskIDAssocUid     sync.Map
	taskIdAssocMasking sync.Map
)

func (cu *CloudbeaverUsecase) buildTaskIdAssocDataMasking(raw []byte, enableMasking bool) error {
	var taskInfo TaskInfo

	if err := UnmarshalGraphQLResponse(raw, &taskInfo); err != nil {
		cu.log.Errorf("extract task id err: %v", err)

		return fmt.Errorf("extract task id err: %v", err)
	}

	taskIdAssocMasking.Store(taskInfo.Data.TaskInfo.ID, enableMasking)

	return nil
}

// TODO 这个函数太大了，需要找时间拆分一下
// GraphQLDistributor 返回一个Echo中间件函数，用于分发和处理CloudBeaver的GraphQL请求
func (cu *CloudbeaverUsecase) GraphQLDistributor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// 检查请求URI是否匹配CloudBeaver的GraphQL API路径
			gqlPath := path.Join(CbRootUri, CbGqlApi)

			if c.Request().RequestURI != gqlPath {
				return next(c)
			}

			// 复制请求体内容
			reqBody := make([]byte, 0)
			if c.Request().Body != nil { // 读取请求体
				reqBody, err = io.ReadAll(c.Request().Body)
				if err != nil {
					cu.log.Errorf("read request body err: %v", err)
					return err
				}
			}
			// 重置请求体以便后续处理
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))

			// 解析GraphQL请求参数
			var params *graphql.RawParams
			err = json.Unmarshal(reqBody, &params)
			if err != nil {
				cu.log.Errorf("graphql.RawParams json unmarshal err: %v", err)
				return err
			}

			// 根据操作名称查找对应的处理器
			cloudbeaverHandle, ok := cloudbeaver.GraphQLHandlerRouters[params.OperationName]
			if !ok {
				return next(c)
			}
			// 如果该操作被禁用，返回错误响应
			if cloudbeaverHandle.Disable {
				message := "this feature is prohibited"
				cu.log.Errorf("%v:%v", message, params.OperationName)
				return c.JSON(http.StatusOK, model.ServerError{
					Message: &message,
				})
			}

			// 执行预处理函数（如果存在）
			if cloudbeaverHandle.Preprocessing != nil {
				if err = cloudbeaverHandle.Preprocessing(c, params); err != nil {
					cu.log.Error(err)
					return err
				}
			}

			// 只有当使用本地处理器或需要特殊处理时才创建smartResponseWriter
			var srw *ResponseInterceptor
			var cloudbeaverResBuf *bytes.Buffer
			var needSmartWriter bool = cloudbeaverHandle.UseLocalHandler ||
				params.OperationName == "navNodeChildren"

			if needSmartWriter {
				srw = newSmartResponseWriter(c)
				cloudbeaverResBuf = srw.Buffer
				c.Response().Writer = srw
				defer func() {
					if !needSmartWriter || srw == nil {
						return
					}
					cu.handleNavNodeChildrenOperation(params, c, cloudbeaverResBuf)
					respBytesBuf := cloudbeaverResBuf.Bytes()

					length := c.Response().Header().Get("content-length")
					actualLength := strconv.Itoa(len(respBytesBuf))
					if length != actualLength {
						cu.log.Warnf("Response content-length mismatch, header: %s, actual: %s", length, actualLength)
						// 在WriteHeader之前设置正确的content-length
						c.Response().Header().Set("content-length", actualLength)
					}

					if srw.status != 0 {
						srw.original.WriteHeader(srw.status)
					}
					_, writeErr := srw.original.Write(respBytesBuf)
					if writeErr != nil {
						c.Logger().Error("Failed to write original response:", writeErr)
					}
				}()
			}

			// 使用本地处理方法
			if cloudbeaverHandle.UseLocalHandler {
				ctx := graphql.StartOperationTrace(c.Request().Context())

				var dbService *DBService
				if params.OperationName == "asyncReadDataFromContainer" {
					dbService, err = cu.getDbService(c.Request().Context(), params)
					if err != nil {
						cu.log.Error(err)
						return err
					}

					if err = next(c); err != nil {
						return err
					}

					// 构建任务ID与数据脱敏的关联
					return cu.buildTaskIdAssocDataMasking(cloudbeaverResBuf.Bytes(), dbService.IsMaskingSwitch)
				}

				// 处理异步SQL执行查询请求
				if params.OperationName == "asyncSqlExecuteQuery" {
					dbService, err = cu.getDbService(c.Request().Context(), params)
					if err != nil {
						cu.log.Error(err)
						return err
					}

					// 如果未启用SQL审计
					if !cu.isEnableSQLAudit(dbService) {

						if err = next(c); err != nil {
							return err
						}

						// 保存未启用SQL审计的日志
						err := cu.SaveCbLogSqlAuditNotEnable(c, dbService, params, cloudbeaverResBuf)
						if err != nil {
							cu.log.Error(err)
						}

						// 构建任务ID与数据脱敏的关联
						return cu.buildTaskIdAssocDataMasking(cloudbeaverResBuf.Bytes(), dbService.IsMaskingSwitch)
					}

					// 获取SQLE服务地址
					sqleUrl, err := cu.getSQLEUrl(c.Request().Context())
					if err != nil {
						return err
					}

					// 构建直接审计请求参数
					directAuditReq := cloudbeaver.DirectAuditParams{
						AuditSQLReq: cloudbeaver.AuditSQLReq{
							InstanceType:     dbService.DBType,
							ProjectId:        dbService.ProjectUID,
							RuleTemplateName: dbService.SQLEConfig.SQLQueryConfig.RuleTemplateName,
						},
						SQLEAddr:                         fmt.Sprintf("%s/v2/sql_audit", sqleUrl),
						AllowQueryWhenLessThanAuditLevel: dbService.GetAllowQueryWhenLessThanAuditLevel(),
					}

					// 将SQLE直接审计参数传递到上下文中
					ctx = context.WithValue(ctx, cloudbeaver.SQLEDirectAudit, directAuditReq)
				}

				// 处理批量更新结果请求
				if params.OperationName == "updateResultsDataBatch" {

					if err = next(c); err != nil {
						return err
					}

					// 保存UI操作日志
					if err := cu.SaveUiOp(c, cloudbeaverResBuf, params); err != nil {
						cu.log.Errorf("save ui op err: %v", err)
						return nil
					}

					return nil
				}

				// 处理异步批量更新结果请求
				if params.OperationName == "asyncUpdateResultsDataBatch" {
					dbService, err := cu.getDbService(c.Request().Context(), params)
					if err != nil {
						cu.log.Error(err)
						return err
					}

					if err = next(c); err != nil {
						return err
					}

					// 构建任务ID与数据脱敏的关联
					return cu.buildTaskIdAssocDataMasking(cloudbeaverResBuf.Bytes(), dbService.IsMaskingSwitch)
				}

				// 处理获取异步任务信息请求
				if params.OperationName == "getAsyncTaskInfo" {
					if err = next(c); err != nil {
						return err
					}

					// 从任务ID关联中获取用户ID
					cbUid, exist := taskIDAssocUid.Load(params.Variables["taskId"])
					if !exist {
						return nil
					}
					cbUidStr, ok := cbUid.(string)
					if !ok {
						return nil
					}

					// 获取操作日志
					operationLog, err := cu.cbOperationLogUsecase.GetCbOperationLogByID(ctx, cbUidStr)
					if err != nil {
						cu.log.Errorf("get cb operation log by id %s failed: %v", cbUidStr, err)
						return nil
					}

					var taskInfo TaskInfo
					if err := UnmarshalGraphQLResponse(cloudbeaverResBuf.Bytes(), &taskInfo); err != nil {
						cu.log.Errorf("extract task id err: %v", err)
						return nil
					}
					task := taskInfo.Data.TaskInfo
					if task.Running {
						return nil
					}
					if task.Error != nil {
						operationLog.ExecResult = *task.Error.Message
					}
					if task.TaskResult != nil {
						operationLog.ExecResult = fmt.Sprintf("%s", task.TaskResult)
					}

					// 更新操作日志
					err = cu.cbOperationLogUsecase.UpdateCbOperationLog(ctx, operationLog)
					if err != nil {
						cu.log.Error(err)
						return nil
					}
					return nil
				}

				enableMasking := false
				// 处理获取SQL执行任务结果请求
				if params.OperationName == "getSqlExecuteTaskResults" {
					// 检查是否需要数据脱敏
					taskIdAssocMaskingVal, exist := taskIdAssocMasking.LoadAndDelete(params.Variables["taskId"])
					if !exist {
						msg := fmt.Sprintf("task id %v assoc masking val does not exist", params.Variables["taskId"])
						return c.JSON(http.StatusOK, model.ServerError{Message: &msg})
					}

					enableMasking, ok = taskIdAssocMaskingVal.(bool)
					if !ok {
						msg := fmt.Sprintf("task id %v assoc masking val is not bool", params.Variables["taskId"])
						return c.JSON(http.StatusOK, model.ServerError{Message: &msg})
					}
				}

				// 设置GraphQL请求的读取时间
				params.ReadTime = graphql.TraceTiming{
					Start: graphql.Now(),
					End:   graphql.Now(),
				}

				// 克隆请求头
				params.Headers = c.Request().Header.Clone()

				var cloudbeaverNext cloudbeaver.Next
				var resp cloudbeaver.AuditResults
				// 如果不需要修改远程响应
				if !cloudbeaverHandle.NeedModifyRemoteRes {
					cloudbeaverNext = func(c echo.Context) ([]byte, error) {
						resp, ok = c.Get(cloudbeaver.AuditResultKey).(cloudbeaver.AuditResults)
						if ok && !resp.IsSuccess {
							err = cu.SaveCbOpLog(c, dbService, params, resp.Results, resp.IsSuccess, nil)
							if err != nil {
								cu.log.Errorf("save cb operation log err: %v", err)
							}

							return nil, c.JSON(http.StatusOK, convertToResp(ctx, resp))
						}

						// 判断是否需要通过工单执行（非 DQL 语句）
						if cu.shouldExecuteByWorkflow(dbService, resp.Results) {
							return cu.executeNonDQLByWorkflow(ctx, c, dbService, params, resp)
						}

						if err = next(c); err != nil {
							return nil, err
						}

						if ok && resp.IsSuccess {
							var taskInfo TaskInfo
							err = UnmarshalGraphQLResponse(cloudbeaverResBuf.Bytes(), &taskInfo)
							if err != nil {
								cu.log.Errorf("extract task id err: %v", err)
							} else {
								err = cu.SaveCbOpLog(c, dbService, params, resp.Results, resp.IsSuccess, &taskInfo.Data.TaskInfo.ID)
								if err != nil {
									cu.log.Errorf("save cb operation log err: %v", err)
								}
							}
						}

						if params.OperationName == "getSqlExecuteTaskResults" {
							err = cu.UpdateCbOpResult(c, cloudbeaverResBuf, params, ctx)
							if err != nil {
								cu.log.Errorf("update cb operation result err: %v", err)
							}
						}

						if params.OperationName == "asyncSqlExecuteQuery" {
							if err := cu.buildTaskIdAssocDataMasking(cloudbeaverResBuf.Bytes(), dbService.IsMaskingSwitch); err != nil {
								return nil, err
							}
						}

						return nil, nil
					}
				} else { // 需要修改远程响应
					cloudbeaverNext = func(c echo.Context) ([]byte, error) {
						resp, ok = c.Get(cloudbeaver.AuditResultKey).(cloudbeaver.AuditResults)
						if ok && !resp.IsSuccess {
							err = cu.SaveCbOpLog(c, dbService, params, resp.Results, resp.IsSuccess, nil)
							if err != nil {
								cu.log.Errorf("save cb operation log err: %v", err)
							}

							return nil, c.JSON(http.StatusOK, convertToResp(ctx, resp))
						}

						if err = next(c); err != nil {
							return nil, err
						}

						if ok && resp.IsSuccess {
							var taskInfo TaskInfo
							err = UnmarshalGraphQLResponse(cloudbeaverResBuf.Bytes(), &taskInfo)
							if err != nil {
								cu.log.Errorf("extract task id err: %v", err)
							} else {
								err = cu.SaveCbOpLog(c, dbService, params, resp.Results, resp.IsSuccess, &taskInfo.Data.TaskInfo.ID)
								if err != nil {
									cu.log.Errorf("save cb operation log err: %v", err)
								}
							}
						}
						// 处理SQL执行结果
						if params.OperationName == "getSqlExecuteTaskResults" {
							err = cu.UpdateCbOpResult(c, cloudbeaverResBuf, params, ctx)
							if err != nil {
								cu.log.Errorf("update cb operation result err: %v", err)
							}
						}

						return cloudbeaverResBuf.Bytes(), nil
					}
				}

				// 创建GraphQL可执行schema
				g := resolver.NewExecutableSchema(resolver.Config{
					Resolvers: cloudbeaver.NewResolverImpl(c, cloudbeaverNext, cu.dataMaskingUseCase.SQLExecuteResultsDataMasking, enableMasking),
					Directives: resolver.DirectiveRoot{
						Since: func(ctx context.Context, obj any, next graphql.Resolver, version string) (res any, err error) {
							// @since directive implementation
							// This directive is used to mark fields that are available since a specific version
							// For CloudBeaver integration, we'll always allow access to these fields
							// as they are part of the current schema
							return next(ctx)
						},
					},
				})

				// 创建GraphQL执行器
				exec := executor.New(g)

				// 创建操作上下文
				rc, err := exec.CreateOperationContext(ctx, params)
				if err != nil {
					return err
				}
				// 分发操作
				responses, ctx := exec.DispatchOperation(ctx, rc)

				// 获取响应
				res := responses(ctx)
				if res.Errors.Error() != "" {
					cu.log.Errorf("GraphQL error: %v", res.Errors)
					return res.Errors
				}
				if !cloudbeaverHandle.NeedModifyRemoteRes {
					return nil
				} else {
					// 设置响应头
					b, err := json.Marshal(res)
					if err != nil {
						return err
					}
					header := srw.original.Header()
					// 重写响应内容
					header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
					cloudbeaverResBuf.Reset()
					// 写入新内容
					cloudbeaverResBuf.Write(b)
					return err
				}
			}
			return next(c)
		}
	}
}

func convertToResp(ctx context.Context, resp cloudbeaver.AuditResults) interface{} {
	var messages []string
	var executionFailedMessage []string
	langTag := locale.Bundle.GetLangTagFromCtx(ctx)
	for _, sqlResult := range resp.Results {
		for _, audit := range sqlResult.AuditResult {
			msg := audit.GetAuditMsgByLangTag(langTag)
			if audit.ExecutionFailed {
				executionFailedMessage = append(executionFailedMessage, msg)
			} else {
				messages = append(messages, msg)
			}
		}
	}

	messageStr := strings.Join(messages, ",")
	executionFailedMessageStr := strings.Join(executionFailedMessage, ",")
	name := "SQL Audit Failed"

	return struct {
		Data struct {
			TaskInfo model.AsyncTaskInfo `json:"taskInfo"`
		} `json:"data"`
	}{
		struct {
			TaskInfo model.AsyncTaskInfo `json:"taskInfo"`
		}{
			TaskInfo: model.AsyncTaskInfo{
				Name:    &name,
				Running: false,
				Status:  &resp.SQL,
				Error: &model.ServerError{
					Message:                &messageStr,
					ExecutionFailedMessage: &executionFailedMessageStr,
					StackTrace:             &messageStr,
				},
			},
		},
	}
}

func newResp(ctx context.Context, name, errCode, msg string) interface{} {
	return struct {
		Data struct {
			TaskInfo model.AsyncTaskInfo `json:"taskInfo"`
		} `json:"data"`
	}{
		struct {
			TaskInfo model.AsyncTaskInfo `json:"taskInfo"`
		}{
			TaskInfo: model.AsyncTaskInfo{
				Name:    &name,
				Running: false,
				Error: &model.ServerError{
					Message:    &msg,
					ErrorCode:  &errCode,
					StackTrace: &msg,
				},
			},
		},
	}
}

// ResponseInterceptor 拦截HTTP响应，允许在响应发送到客户端之前进行修改或日志记录
type ResponseInterceptor struct {
	echo.Response
	Buffer   *bytes.Buffer
	original http.ResponseWriter
	status   int
	hijacked bool // 添加劫持状态标记
}

func newSmartResponseWriter(c echo.Context) *ResponseInterceptor {
	buf := new(bytes.Buffer)
	return &ResponseInterceptor{
		Response: *c.Response(),
		Buffer:   buf,
		original: c.Response().Writer,
		hijacked: false,
	}
}

// 实现Hijacker接口
func (w *ResponseInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	if hijacker, ok := w.original.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support Hijack")
}

func (w *ResponseInterceptor) Write(b []byte) (int, error) {
	// 检查连接是否已被劫持
	if w.hijacked {
		return len(b), nil
	}

	// 如果未设置状态码，则补默认值
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	// 写入 buffer，不立即写给客户端
	return w.Buffer.Write(b)
}

func (w *ResponseInterceptor) WriteHeader(code int) {
	// 检查连接是否已被劫持
	if w.hijacked {
		return
	}
	w.status = code
}

func (cu *CloudbeaverUsecase) handleNavNodeChildrenOperation(params *graphql.RawParams, c echo.Context, responseBuf *bytes.Buffer) {
	if params.OperationName != "navNodeChildren" {
		return
	}

	parentPath, ok := params.Variables["parentPath"].(string)
	if !ok || parentPath != "resource://g_GlobalConfiguration" {
		return
	}

	currentUserUid := c.Get(dmsUserIdKey).(string)
	var navNodeChildrenResponse cloudbeaver.NavNodeChildrenResponse
	if err := UnmarshalGraphQLResponseNavNodeChildren(responseBuf.Bytes(), &navNodeChildrenResponse); err != nil {
		cu.log.Errorf("Cloudbeaver extract navNodeChildren response err: %v", err)
		return
	}

	cu.log.Debugf("user %s get connections navNodeChildren: %v", currentUserUid, navNodeChildrenResponse.Data.NavNodeChildren)
	connections, err := cu.repo.GetCloudbeaverConnectionsByUserId(c.Request().Context(), currentUserUid)
	if err != nil {
		cu.log.Errorf("get cloudbeaver connections by user id %s failed: %v", currentUserUid, err)
		return
	}

	if len(connections) != len(navNodeChildrenResponse.Data.NavNodeChildren) &&
		len(navNodeChildrenResponse.Data.NavNodeChildren) > 0 {
		// 根据connections的值来过滤navNodeChildren，删除多余的信息
		originalCount := len(navNodeChildrenResponse.Data.NavNodeChildren)
		filteredNavNodeChildren := make([]cloudbeaver.NavNodeInfo, 0)

		// 创建connections的ID映射，提高查找效率
		connectionIDMap := make(map[string]bool)
		for _, connection := range connections {
			if connection != nil {
				connectionIDMap[connection.CloudbeaverConnectionID] = true
			}
		}

		// 只保留在connections中存在的navNode
		for _, navNode := range navNodeChildrenResponse.Data.NavNodeChildren {
			if connectionIDMap[strings.TrimPrefix(navNode.ID, "database://")] {
				filteredNavNodeChildren = append(filteredNavNodeChildren, navNode)
			}
		}

		// 保持完整的 GraphQL 响应结构，只修改 data.navNodeChildren 部分
		var originalResponse map[string]interface{}
		if err := json.Unmarshal(responseBuf.Bytes(), &originalResponse); err != nil {
			cu.log.Errorf("failed to unmarshal original response: %v", err)
			return
		}

		// 更新 data.navNodeChildren
		if data, ok := originalResponse["data"].(map[string]interface{}); ok {
			data["navNodeChildren"] = filteredNavNodeChildren
		}

		// 重新序列化并更新responseBuf
		updatedResponseBytes, err := json.Marshal(originalResponse)
		if err != nil {
			cu.log.Errorf("failed to marshal updated response: %v", err)
			return
		}

		// 清空原buffer并写入新的响应
		responseBuf.Reset()
		responseBuf.Write(updatedResponseBytes)

		cu.log.Warnf("user %s filtered navNodeChildren from %d to %d based on connections",
			currentUserUid, originalCount, len(filteredNavNodeChildren))
	}
}

func (cu *CloudbeaverUsecase) isEnableSQLAudit(dbService *DBService) bool {
	if dbService.SQLEConfig == nil || dbService.SQLEConfig.SQLQueryConfig == nil {
		return false
	}
	return dbService.SQLEConfig.AuditEnabled && dbService.SQLEConfig.SQLQueryConfig.AuditEnabled
}

func (cu *CloudbeaverUsecase) isEnableWorkflowExec(dbService *DBService) bool {
	if dbService.SQLEConfig == nil || dbService.SQLEConfig.SQLQueryConfig == nil {
		return false
	}
	return dbService.SQLEConfig.AuditEnabled && dbService.SQLEConfig.SQLQueryConfig.WorkflowExecEnabled
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
	client := cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithCookie(cookies), cloudbeaver.WithLogger(cu.log))
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

type authCache struct {
	cookies   []*http.Cookie
	expiresAt time.Time
}

var (
	dmsUserIdCloudbeaverLoginMap = make(map[string]cloudbeaverSession)
	tokenMapMutex                = &sync.Mutex{}

	// 统一认证缓存
	authCacheMap = make(map[string]*authCache)
	authMutex    = &sync.Mutex{}

	// 缓存过期时间：5分钟
	authCacheExpiry = 5 * time.Minute
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

// getAuthCache 获取认证缓存
func (cu *CloudbeaverUsecase) getAuthCache(username string) []*http.Cookie {
	authMutex.Lock()
	defer authMutex.Unlock()

	cache, exists := authCacheMap[username]
	if !exists || time.Now().After(cache.expiresAt) {
		// 缓存不存在或已过期，清理
		delete(authCacheMap, username)
		return nil
	}

	cu.log.Debugf("Using cached auth for user: %s", username)
	return cache.cookies
}

// setAuthCache 设置认证缓存
func (cu *CloudbeaverUsecase) setAuthCache(username string, cookies []*http.Cookie) {
	authMutex.Lock()
	defer authMutex.Unlock()

	authCacheMap[username] = &authCache{
		cookies:   cookies,
		expiresAt: time.Now().Add(authCacheExpiry),
	}

	cu.log.Debugf("Cached auth for user: %s, expires at: %v", username, authCacheMap[username].expiresAt)
}

// getAuthCookies 获取认证 cookies（复用缓存逻辑）
func (cu *CloudbeaverUsecase) getAuthCookies(username, password string) ([]*http.Cookie, error) {
	// 尝试从缓存获取认证信息
	cachedCookies := cu.getAuthCache(username)
	if cachedCookies != nil {
		cu.log.Debugf("Using cached authentication for user: %s", username)
		return cachedCookies, nil
	}

	// 缓存未命中，执行登录
	cu.log.Debugf("Cache miss, performing CloudBeaver login for user: %s", username)
	cookies, err := cu.loginCloudbeaverServer(username, password)
	if err != nil {
		return nil, err
	}

	// 缓存认证信息
	cu.setAuthCache(username, cookies)
	return cookies, nil
}

// clearExpiredAuthCache 清理过期的认证缓存
func (cu *CloudbeaverUsecase) clearExpiredAuthCache() {
	authMutex.Lock()
	defer authMutex.Unlock()

	now := time.Now()
	for user, cache := range authCacheMap {
		if now.After(cache.expiresAt) {
			delete(authCacheMap, user)
			cu.log.Debugf("Cleared expired auth cache for user: %s", user)
		}
	}
}

// startCacheCleanupRoutine 启动缓存清理协程
func (cu *CloudbeaverUsecase) startCacheCleanupRoutine() {
	ticker := time.NewTicker(authCacheExpiry / 2) // 每2.5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		cu.clearExpiredAuthCache()
	}
}

type UserList struct {
	ListUsers []struct {
		UserID string `json:"userID"`
	} `json:"listUsers"`
}

var reservedCloudbeaverUserId = map[string]struct{}{"admin": {}, "user": {}}

func (cu *CloudbeaverUsecase) createUserIfNotExist(ctx context.Context, cloudbeaverUserId string, dmsUser *User) error {
	cu.log.Infof("Creating CloudBeaver user if not exist: %s for DMS user: %s", cloudbeaverUserId, dmsUser.UID)

	if _, ok := reservedCloudbeaverUserId[cloudbeaverUserId]; ok {
		cu.log.Errorf("Username %s is reserved and cannot be used", cloudbeaverUserId)
		return fmt.Errorf("username %s is reserved， cann't be used", cloudbeaverUserId)
	}

	// 使用管理员身份登录
	graphQLClient, err := cu.getGraphQLClientWithRootUser()
	if err != nil {
		cu.log.Errorf("Failed to get GraphQL client with root user for creating user %s: %v", cloudbeaverUserId, err)
		return err
	}

	userExist, err := cu.checkCloudBeaverUserExist(ctx, graphQLClient, cloudbeaverUserId)
	if err != nil {
		cu.log.Errorf("Failed to check if CloudBeaver user %s exists: %v", cloudbeaverUserId, err)
		return fmt.Errorf("check cloudbeaver user exist failed: %v", err)
	}

	// 用户不存在则创建CloudBeaver用户
	if !userExist {
		cu.log.Infof("CloudBeaver user %s does not exist, creating new user", cloudbeaverUserId)

		// 创建用户
		createUserReq := cloudbeaver.NewRequest(cu.graphQl.CreateUserQuery(), map[string]interface{}{
			"userId": cloudbeaverUserId,
		})
		err = graphQLClient.Run(ctx, createUserReq, &UserList{})
		if err != nil {
			cu.log.Errorf("Failed to create CloudBeaver user %s: %v", cloudbeaverUserId, err)
			return fmt.Errorf("create cloudbeaver user failed: %v", err)
		}

		cu.log.Infof("Successfully created CloudBeaver user: %s", cloudbeaverUserId)

		// 授予角色(不授予角色的用户无法登录)
		grantUserRoleReq := cloudbeaver.NewRequest(cu.graphQl.GrantUserRoleQuery(), map[string]interface{}{
			"userId": cloudbeaverUserId,
			"roleId": CBUserRole,
			"teamId": CBUserRole,
		})
		err = graphQLClient.Run(ctx, grantUserRoleReq, nil)
		if err != nil {
			cu.log.Errorf("Failed to grant role to CloudBeaver user %s: %v", cloudbeaverUserId, err)
			return fmt.Errorf("grant cloudbeaver user failed: %v", err)
		}

		cu.log.Infof("Successfully granted role to CloudBeaver user: %s", cloudbeaverUserId)
	} else {
		cu.log.Debugf("CloudBeaver user %s already exists, checking fingerprint", cloudbeaverUserId)

		cloudbeaverUser, exist, err := cu.repo.GetCloudbeaverUserByID(ctx, cloudbeaverUserId)
		if err != nil {
			cu.log.Errorf("Failed to get CloudBeaver user %s from cache: %v", cloudbeaverUserId, err)
			return err
		}

		if exist && cloudbeaverUser.DMSFingerprint == cu.userUsecase.GetUserFingerprint(dmsUser) {
			cu.log.Debugf("CloudBeaver user %s fingerprint matches, no update needed", cloudbeaverUserId)
			return nil
		}
	}

	// 设置CloudBeaver用户密码
	cu.log.Infof("Updating CloudBeaver user password for user: %s", cloudbeaverUserId)

	updatePasswordReq := cloudbeaver.NewRequest(cu.graphQl.UpdatePasswordQuery(), map[string]interface{}{
		"userId": cloudbeaverUserId,
		"credentials": model.JSON{
			"password": strings.ToUpper(aes.Md5(dmsUser.Password)),
		},
	})
	err = graphQLClient.Run(ctx, updatePasswordReq, nil)
	if err != nil {
		cu.log.Errorf("Failed to update CloudBeaver user password for user %s: %v", cloudbeaverUserId, err)
		return fmt.Errorf("update cloudbeaver user failed: %v", err)
	}

	cu.log.Infof("Successfully updated CloudBeaver user password for user: %s", cloudbeaverUserId)

	cloudbeaverUser := &CloudbeaverUser{
		DMSUserID:         dmsUser.UID,
		DMSFingerprint:    cu.userUsecase.GetUserFingerprint(dmsUser),
		CloudbeaverUserID: cloudbeaverUserId,
	}

	return cu.repo.UpdateCloudbeaverUserCache(ctx, cloudbeaverUser)
}

func (cu *CloudbeaverUsecase) checkCloudBeaverUserExist(ctx context.Context, graphQLClient *cloudbeaver.Client, cloudbeaverUserId string) (bool, error) {
	if graphQLClient == nil {
		return false, fmt.Errorf("graphQLClient is nil")
	}

	checkExistReq := cloudbeaver.NewRequest(cu.graphQl.IsUserExistQuery(cloudbeaverUserId))
	cloudbeaverUserList := UserList{}
	err := graphQLClient.Run(ctx, checkExistReq, &cloudbeaverUserList)
	if err != nil {
		return false, err
	}
	for _, cloudBeaverUser := range cloudbeaverUserList.ListUsers {
		if cloudBeaverUser.UserID == cloudbeaverUserId {
			return true, nil
		}
	}
	return false, nil
}

func (cu *CloudbeaverUsecase) connectManagement(ctx context.Context, cloudbeaverUserId string, dmsUser *User) error {
	activeDBServices, err := cu.dbServiceUsecase.GetActiveDBServices(ctx, nil)
	if err != nil {
		return err
	}

	if len(activeDBServices) == 0 {
		return cu.clearConnection(ctx)
	}

	hasGlobalOpPermission, err := cu.opPermissionVerifyUsecase.CanOpGlobal(ctx, dmsUser.UID)
	if err != nil {
		return err
	}

	if !hasGlobalOpPermission {
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

	cloudbeaverUser, exist, err := cu.repo.GetCloudbeaverUserByID(ctx, cloudbeaverUserId)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("cloudbeaver user: %s not eixst", cloudbeaverUserId)
	}

	if err = cu.operateConnection(ctx, cloudbeaverUser, dmsUser, activeDBServices); err != nil {
		return err
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

type UserConnectionsResp struct {
	Connections []*struct {
		Id       string `json:"id"`
		Template bool   `json:"template"`
	} `json:"connections"`
}

// 获取用户当前数据库连接ID
func (cu *CloudbeaverUsecase) getUserConnectionIds(ctx context.Context, cloudbeaverUser *CloudbeaverUser, dmsUser *User) ([]string, error) {
	// 使用3次重试，因为这个接口有时会出现不稳定的情况
	client, err := cu.getGraphQLClient(cloudbeaverUser.CloudbeaverUserID, dmsUser.Password, 3)
	if err != nil {
		return nil, err
	}

	var userConnectionsResp UserConnectionsResp

	variables := map[string]interface{}{"projectId": cloudbeaverProjectId}
	err = client.Run(ctx, cloudbeaver.NewRequest(cu.graphQl.GetUserConnectionsQuery(), variables), &userConnectionsResp)
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0, len(userConnectionsResp.Connections))
	for _, connection := range userConnectionsResp.Connections {
		if !connection.Template {
			ret = append(ret, connection.Id)
		}
	}

	return ret, nil
}

func (cu *CloudbeaverUsecase) operateConnection(ctx context.Context, cloudbeaverUser *CloudbeaverUser, dmsUser *User, activeDBServices []*DBService) error {
	dbServiceMap := map[string]*DBService{}
	projectMap := map[string]string{}
	for _, service := range activeDBServices {
		dbServiceMap[getDBPrimaryKey(service.UID, service.AccountPurpose, dmsUser.UID)] = service

		project, err := cu.dbServiceUsecase.projectUsecase.GetProject(ctx, service.ProjectUID)
		if err != nil {
			projectMap[service.UID] = "unknown"
			cu.log.Errorf("get db service project %s failed, err: %v", service.ProjectUID, err)
		} else {
			projectMap[service.UID] = project.Name
		}
	}

	// 获取当前用户所有已创建的连接
	localCloudbeaverConnections, err := cu.repo.GetCloudbeaverConnectionsByUserId(ctx, dmsUser.UID)
	if err != nil {
		return err
	}

	// cloudbeaver连接数为空则重置缓存
	if userConnectionIds, err := cu.getUserConnectionIds(ctx, cloudbeaverUser, dmsUser); err != nil {
		return err
	} else if len(userConnectionIds) == 0 {
		localCloudbeaverConnections = []*CloudbeaverConnection{}
	}

	var deleteConnections []*CloudbeaverConnection

	cloudbeaverConnectionMap := map[string]*CloudbeaverConnection{}
	for _, connection := range localCloudbeaverConnections {
		// 删除用户关联的连接
		if connection.DMSUserId == dmsUser.UID {
			cloudbeaverConnectionMap[connection.PrimaryKey()] = connection
			if _, ok := dbServiceMap[connection.PrimaryKey()]; !ok {
				deleteConnections = append(deleteConnections, connection)
			}
		}
	}

	var createConnections []*CloudbeaverConnection
	var updateConnections []*CloudbeaverConnection

	for _, dbService := range dbServiceMap {
		if cloudbeaverConnection, ok := cloudbeaverConnectionMap[getDBPrimaryKey(dbService.UID, dbService.AccountPurpose, dmsUser.UID)]; ok {
			if cloudbeaverConnection.DMSDBServiceFingerprint != cu.dbServiceUsecase.GetDBServiceFingerprint(dbService) {
				updateConnections = append(updateConnections, &CloudbeaverConnection{DMSDBServiceID: dbService.UID, CloudbeaverConnectionID: cloudbeaverConnection.CloudbeaverConnectionID, Purpose: dbService.AccountPurpose, DMSUserId: dmsUser.UID})
			}
		} else {
			createConnections = append(createConnections, &CloudbeaverConnection{DMSDBServiceID: dbService.UID, Purpose: dbService.AccountPurpose, DMSUserId: dmsUser.UID})
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
		if err = cu.createCloudbeaverConnection(ctx, cloudbeaverClient, dbServiceMap[getDBPrimaryKey(createConnection.DMSDBServiceID, createConnection.Purpose, dmsUser.UID)],
			projectMap[createConnection.DMSDBServiceID], dmsUser.UID); err != nil {
			cu.log.Errorf("create connection %v failed: %v", createConnection, err)
		}
	}

	for _, updateConnection := range updateConnections {
		if err = cu.updateCloudbeaverConnection(ctx, cloudbeaverClient, updateConnection.CloudbeaverConnectionID, dbServiceMap[getDBPrimaryKey(updateConnection.DMSDBServiceID, updateConnection.Purpose, dmsUser.UID)], projectMap[updateConnection.DMSDBServiceID], dmsUser.UID); err != nil {
			cu.log.Errorf("update dnServerId %s to connection failed: %v", updateConnection, err)
		}
	}

	for _, deleteConnection := range deleteConnections {
		if err = cu.deleteCloudbeaverConnection(ctx, cloudbeaverClient, deleteConnection.CloudbeaverConnectionID, deleteConnection.DMSDBServiceID, dmsUser.UID, deleteConnection.Purpose); err != nil {
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
	localCloudbeaverConnections, err := cu.repo.GetCloudbeaverConnectionsByUserIdAndDBServiceIds(ctx, dmsUser.UID, dbServiceIds)
	if err != nil {
		return err
	}

	// 从缓存中获取需要同步的CloudBeaver实例
	cloudbeaverConnectionMap := map[string]*CloudbeaverConnection{}
	for _, connection := range localCloudbeaverConnections {
		cloudbeaverConnectionMap[connection.CloudbeaverConnectionID] = connection
	}

	cloudbeaverConnectionIds, err := cu.getUserConnectionIds(ctx, cloudbeaverUser, dmsUser)
	if err != nil {
		return err
	}

	if len(cloudbeaverConnectionIds) != len(localCloudbeaverConnections) {
		return cu.bindUserAccessConnection(ctx, localCloudbeaverConnections, cloudbeaverUser.CloudbeaverUserID)
	}

	for _, connectionId := range cloudbeaverConnectionIds {
		if _, ok := cloudbeaverConnectionMap[connectionId]; !ok {
			return cu.bindUserAccessConnection(ctx, localCloudbeaverConnections, cloudbeaverUser.CloudbeaverUserID)
		}
	}

	return nil
}

func (cu *CloudbeaverUsecase) bindUserAccessConnection(ctx context.Context, cloudbeaverDBServices []*CloudbeaverConnection, cloudBeaverUserID string) error {
	cu.log.Infof("Binding CloudBeaver connections to user: %s, connection count: %d", cloudBeaverUserID, len(cloudbeaverDBServices))

	cloudbeaverConnectionIds := make([]string, 0, len(cloudbeaverDBServices))
	for _, service := range cloudbeaverDBServices {
		cloudbeaverConnectionIds = append(cloudbeaverConnectionIds, service.CloudbeaverConnectionID)
	}

	cu.log.Debugf("CloudBeaver connection IDs for user %s: %v", cloudBeaverUserID, cloudbeaverConnectionIds)

	cloudbeaverConnReq := cloudbeaver.NewRequest(cu.graphQl.SetUserConnectionsQuery(), map[string]interface{}{
		"userId":      cloudBeaverUserID,
		"connections": cloudbeaverConnectionIds,
	})

	rootClient, err := cu.getGraphQLClientWithRootUser()
	if err != nil {
		cu.log.Errorf("Failed to get GraphQL client with root user for binding connections to user %s: %v", cloudBeaverUserID, err)
		return err
	}

	err = rootClient.Run(ctx, cloudbeaverConnReq, nil)
	if err != nil {
		cu.log.Errorf("Failed to bind CloudBeaver connections to user %s: %v", cloudBeaverUserID, err)
		return err
	}

	cu.log.Infof("Successfully bound %d CloudBeaver connections to user: %s", len(cloudbeaverConnectionIds), cloudBeaverUserID)
	return nil
}

func (cu *CloudbeaverUsecase) createCloudbeaverConnection(ctx context.Context, client *cloudbeaver.Client, dbService *DBService, project, userId string) error {
	cu.log.Infof("Creating CloudBeaver connection for DB service: %s, project: %s, user: %s, purpose: %s",
		dbService.UID, project, userId, dbService.AccountPurpose)

	params, err := cu.GenerateCloudbeaverConnectionParams(dbService, project, dbService.AccountPurpose)
	if err != nil {
		cu.log.Errorf("Failed to generate connection params for DB service %s: %v", dbService.UID, err)
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
		cu.log.Errorf("Failed to create CloudBeaver connection for DB service %s: %v", dbService.UID, err)
		return err
	}

	cu.log.Infof("Successfully created CloudBeaver connection: %s for DB service: %s", resp.Connection.ID, dbService.UID)

	// 同步缓存
	err = cu.repo.UpdateCloudbeaverConnectionCache(ctx, &CloudbeaverConnection{
		DMSDBServiceID:          dbService.UID,
		DMSUserId:               userId,
		DMSDBServiceFingerprint: cu.dbServiceUsecase.GetDBServiceFingerprint(dbService),
		Purpose:                 dbService.AccountPurpose,
		CloudbeaverConnectionID: resp.Connection.ID,
	})
	if err != nil {
		cu.log.Errorf("Failed to update CloudBeaver connection cache for DB service %s: %v", dbService.UID, err)
	}
	return err
}

// UpdateCloudbeaverConnection 更新完毕后会同步缓存
func (cu *CloudbeaverUsecase) updateCloudbeaverConnection(ctx context.Context, client *cloudbeaver.Client, cloudbeaverConnectionId string, dbService *DBService, project, userId string) error {
	cu.log.Infof("Updating CloudBeaver connection: %s for DB service: %s, project: %s, user: %s, purpose: %s",
		cloudbeaverConnectionId, dbService.UID, project, userId, dbService.AccountPurpose)

	params, err := cu.GenerateCloudbeaverConnectionParams(dbService, project, dbService.AccountPurpose)
	if err != nil {
		cu.log.Errorf("Failed to generate connection params for DB service %s: %v", dbService.UID, err)
		return fmt.Errorf("%s unsupported", dbService.DBType)
	}

	config, ok := params["config"].(map[string]interface{})
	if !ok {
		cu.log.Errorf("Failed to assert connection params for DB service %s", dbService.UID)
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
		cu.log.Errorf("Failed to update CloudBeaver connection %s for DB service %s: %v", cloudbeaverConnectionId, dbService.UID, err)
		return err
	}

	cu.log.Infof("Successfully updated CloudBeaver connection: %s for DB service: %s", resp.Connection.ID, dbService.UID)

	err = cu.repo.UpdateCloudbeaverConnectionCache(ctx, &CloudbeaverConnection{
		DMSDBServiceID:          dbService.UID,
		DMSUserId:               userId,
		DMSDBServiceFingerprint: cu.dbServiceUsecase.GetDBServiceFingerprint(dbService),
		Purpose:                 dbService.AccountPurpose,
		CloudbeaverConnectionID: resp.Connection.ID,
	})
	if err != nil {
		cu.log.Errorf("Failed to update CloudBeaver connection cache for DB service %s: %v", dbService.UID, err)
	}
	return err
}

func (cu *CloudbeaverUsecase) deleteCloudbeaverConnection(ctx context.Context, client *cloudbeaver.Client, cloudbeaverConnectionId, dbServiceId, userId, purpose string) error {
	cu.log.Infof("Deleting CloudBeaver connection: %s for DB service: %s, user: %s, purpose: %s",
		cloudbeaverConnectionId, dbServiceId, userId, purpose)

	variables := make(map[string]interface{})
	variables["connectionId"] = cloudbeaverConnectionId
	variables["projectId"] = cloudbeaverProjectId

	req := cloudbeaver.NewRequest(cu.graphQl.DeleteConnectionQuery(), variables)
	resp := struct {
		DeleteConnection bool `json:"deleteConnection"`
	}{}

	if err := client.Run(ctx, req, &resp); err != nil {
		cu.log.Errorf("Failed to delete CloudBeaver connection %s for DB service %s: %v", cloudbeaverConnectionId, dbServiceId, err)
		return err
	}

	cu.log.Infof("Successfully deleted CloudBeaver connection: %s for DB service: %s", cloudbeaverConnectionId, dbServiceId)

	err := cu.repo.DeleteCloudbeaverConnectionCache(ctx, dbServiceId, userId, purpose)
	if err != nil {
		cu.log.Errorf("Failed to delete CloudBeaver connection cache for DB service %s: %v", dbServiceId, err)
	}
	return err
}

func (cu *CloudbeaverUsecase) generateCommonCloudbeaverConfigParams(dbService *DBService, project, purpose string) map[string]interface{} {
	name := fmt.Sprintf("%s:%s", project, dbService.Name)
	if purpose != "" {
		name = fmt.Sprintf("%s:%s", name, purpose)
	}
	return map[string]interface{}{
		"configurationType": "MANUAL",
		"name":              name,
		// "template":          false,
		"host":            dbService.Host,
		"port":            dbService.Port,
		"databaseName":    nil,
		"description":     nil,
		"authModelId":     "native",
		"saveCredentials": true,
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
	case constant.DBTypePostgreSQL, constant.DBTypeTBase:
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
	// https://github.com/actiontech/dms-ee/issues/276
	config["properties"] = map[string]interface{}{"allowPublicKeyRetrieval": "TRUE"}
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

	providerProperties := map[string]interface{}{
		"@dbeaver-sid-service@": "SERVICE",
		"oracle.logon-as":       "Normal",
	}
	// sys用户不能用normal角色登陆，根据用户名做特殊处理
	if inst.User == "sys" {
		providerProperties["oracle.logon-as"] = "SYSDBA"
	}

	config["providerProperties"] = providerProperties

	// 默认关闭timezoneAsRegion，防止连接Oracle11g报错
	config["properties"] = map[string]interface{}{
		"oracle.jdbc.timezoneAsRegion": false,
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
	// 获取租户, 在用户名处指定连接的租户: user@tenant
	tenant := inst.AdditionalParams.GetParam("tenant_name").String()
	if !strings.Contains(inst.User, "@") && tenant != "" {
		inst.User = fmt.Sprintf("%s@%s", inst.User, tenant)
	}
	// OB MySQL本身支持使用MySQL客户端连接, 此处复用MySQL driver
	// https://github.com/actiontech/dms/issues/365
	config["driverId"] = "mysql:mysql8"
	return nil
}

func (cu *CloudbeaverUsecase) fillGoldenDBParams(config map[string]interface{}) error {
	config["driverId"] = "mysql:mysql8"
	return nil
}

func (cu *CloudbeaverUsecase) getGraphQLClientWithRootUser() (*cloudbeaver.Client, error) {
	adminUser := cu.cloudbeaverCfg.AdminUser

	cookies, err := cu.getAuthCookies(adminUser, cu.cloudbeaverCfg.AdminPassword)
	if err != nil {
		return nil, err
	}

	return cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithCookie(cookies), cloudbeaver.WithLogger(cu.log)), nil
}

// 这个客户端会用指定用户操作, 请求会直接发到CB
// maxRetries 为可选参数，默认为0（不重试）
func (cu *CloudbeaverUsecase) getGraphQLClient(username, password string, maxRetries ...int) (*cloudbeaver.Client, error) {
	cookies, err := cu.getAuthCookies(username, password)
	if err != nil {
		return nil, err
	}

	opts := []cloudbeaver.ClientOption{
		cloudbeaver.WithCookie(cookies),
		cloudbeaver.WithLogger(cu.log),
	}

	// 如果提供了重试次数参数，则添加重试配置
	if len(maxRetries) > 0 && maxRetries[0] > 0 {
		opts = append(opts, cloudbeaver.WithMaxRetries(maxRetries[0]))
	}

	return cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), opts...), nil
}

func (cu *CloudbeaverUsecase) loginCloudbeaverServer(user, pwd string) (cookie []*http.Cookie, err error) {
	client := cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithHttpResHandler(
		func(response *http.Response) {
			if response != nil {
				cookie = response.Cookies()
			}
		}), cloudbeaver.WithLogger(cu.log))
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

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

func (g *GraphQLResponse) Bytes() ([]byte, error) {
	return json.Marshal(g)
}

type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []map[string]int       `json:"locations"`
	Path       []interface{}          `json:"path"`
	Extensions map[string]interface{} `json:"extensions"`
}

func UnmarshalGraphQLResponse(body []byte, taskInfo *TaskInfo) error {
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return err // 真正 JSON 格式错误时才报错
	}

	if len(gqlResp.Errors) > 0 {
		// GraphQL 执行错误
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	// 再解析 Data 成真正结构
	if err := json.Unmarshal(body, taskInfo); err != nil {
		return err
	}
	if taskInfo == nil || taskInfo.Data.TaskInfo == nil {
		return fmt.Errorf("GraphQL error: %v", gqlResp)
	}
	return nil
}

type ExecutionContextInfo struct {
	ID             string  `json:"id"`
	ProjectID      string  `json:"projectId"`
	ConnectionID   string  `json:"connectionId"`
	DefaultCatalog *string `json:"defaultCatalog"`
	DefaultSchema  *string `json:"defaultSchema"`
}

type ExecutionContextListRes struct {
	Contexts []ExecutionContextInfo `json:"contexts"`
}

func (cu *CloudbeaverUsecase) getContextSchema(c echo.Context, connectionId, contextId string) (string, error) {
	if contextId == "" {
		return "", nil
	}

	var cookies []*http.Cookie
	for _, cookie := range c.Cookies() {
		if cookie.Name == CloudbeaverCookieName {
			cookies = append(cookies, cookie)
		}
	}

	if len(cookies) == 0 {
		return "", fmt.Errorf("no cloudbeaver session cookie found")
	}

	query := `
query executionContextList($projectId: ID, $connectionId: ID, $contextId: ID) {
  contexts: sqlListContexts(
    projectId: $projectId
    connectionId: $connectionId
    contextId: $contextId
  ) {
    ...ExecutionContextInfo
  }
}

fragment ExecutionContextInfo on SQLContextInfo {
  id
  projectId
  connectionId
  defaultCatalog
  defaultSchema
}
`

	variables := map[string]interface{}{
		"projectId":    cloudbeaverProjectId,
		"connectionId": connectionId,
		"contextId":    contextId,
	}

	client := cloudbeaver.NewGraphQlClient(cu.getGraphQLServerURI(), cloudbeaver.WithCookie(cookies))
	client.Log = func(s string) {
		cu.log.Debugf("getContextSchema CB GraphQL: %s", s)
	}
	req := cloudbeaver.NewRequest(query, variables)
	req.SetOperationName("executionContextList")

	var res ExecutionContextListRes
	if err := client.Run(c.Request().Context(), req, &res); err != nil {
		cu.log.Errorf("query execution context failed: %v, connectionId: %s, contextId: %s", err, connectionId, contextId)
		return "", fmt.Errorf("query execution context failed: %v", err)
	}

	cu.log.Debugf("execution context response: %+v", res)

	if len(res.Contexts) == 0 {
		cu.log.Warnf("no contexts found in response, connectionId: %s, contextId: %s", connectionId, contextId)
		return "", nil
	}

	contextInfo := res.Contexts[0]
	if contextInfo.DefaultSchema != nil && *contextInfo.DefaultSchema != "" {
		return *contextInfo.DefaultSchema, nil
	}

	if contextInfo.DefaultCatalog != nil && *contextInfo.DefaultCatalog != "" {
		return *contextInfo.DefaultCatalog, nil
	}

	return "", nil
}

// checkWorkflowPermission 校验用户是否有数据源上的"创建、审批、上线工单"的权限
// 权限可以是项目级别的（项目管理员）或数据源级别的（直接针对数据源的工单权限）
func (cu *CloudbeaverUsecase) checkWorkflowPermission(ctx context.Context, userUid string, dbService *DBService) (bool, error) {
	if userUid == constant.UIDOfUserAdmin {
		return true, nil
	}
	opPermissions, err := cu.opPermissionVerifyUsecase.GetUserOpPermissionInProject(ctx, userUid, dbService.ProjectUID)
	if err != nil {
		return false, fmt.Errorf("get user op permission in project err: %v", err)
	}

	// 需要检查的三个工单权限
	requiredPermissions := map[string]struct{}{
		constant.UIDOfOpPermissionCreateWorkflow:  {},
		constant.UIDOfOpPermissionAuditWorkflow:   {},
		constant.UIDOfOpPermissionExecuteWorkflow: {},
	}

	// 当前数据源已拥有的工单权限
	dbServicePermissions := make(map[string]struct{})

	for _, opPermission := range opPermissions {
		// 项目管理员权限
		if opPermission.OpRangeType == OpRangeTypeProject && opPermission.OpPermissionUID == constant.UIDOfOpPermissionProjectAdmin {
			return true, nil
		}

		// 数据源级别的工单权限，只关注当前数据源
		if opPermission.OpRangeType == OpRangeTypeDBService {
			// 检查是否是必需的工单权限
			if _, isRequired := requiredPermissions[opPermission.OpPermissionUID]; isRequired {
				// 检查是否包含当前数据源
				for _, rangeUid := range opPermission.RangeUIDs {
					if rangeUid == dbService.UID {
						dbServicePermissions[opPermission.OpPermissionUID] = struct{}{}
						// 如果已收集到所有必需的权限，提前返回
						if len(dbServicePermissions) == len(requiredPermissions) {
							return true, nil
						}
						break
					}
				}
			}
		}
	}

	// 检查是否拥有所有必需的工单权限
	return len(dbServicePermissions) == len(requiredPermissions), nil
}

// shouldExecuteByWorkflow 判断是否需要通过工单执行（非 DQL 语句）
func (cu *CloudbeaverUsecase) shouldExecuteByWorkflow(dbService *DBService, auditResults []cloudbeaver.AuditSQLResV2) bool {
	if !cu.isEnableSQLAudit(dbService) || !cu.isEnableWorkflowExec(dbService) {
		return false
	}

	for _, result := range auditResults {
		if result.SQLType != "" && result.SQLType != "dql" {
			return true
		}
	}
	return false
}

// workflowExecParams 工单执行所需的参数
type workflowExecParams struct {
	contextIdStr    string
	connectionIdStr string
	query           string
	instanceSchema  string
}

// getWorkflowExecParams 从 GraphQL 参数中提取工单执行所需的参数
func (cu *CloudbeaverUsecase) getWorkflowExecParams(c echo.Context, params *graphql.RawParams) (*workflowExecParams, error) {
	contextIdStr, ok := params.Variables["contextId"].(string)
	if !ok {
		return nil, fmt.Errorf("missing contextId in params")
	}

	connectionIdStr, ok := params.Variables["connectionId"].(string)
	if !ok {
		return nil, fmt.Errorf("missing connectionId in params")
	}

	query, ok := params.Variables["query"].(string)
	if !ok {
		return nil, fmt.Errorf("missing query in params")
	}

	instanceSchema, err := cu.getContextSchema(c, connectionIdStr, contextIdStr)
	if err != nil {
		return nil, fmt.Errorf("get context schema failed: %v", err)
	}

	return &workflowExecParams{
		contextIdStr:    contextIdStr,
		connectionIdStr: connectionIdStr,
		query:           query,
		instanceSchema:  instanceSchema,
	}, nil
}

// executeNonDQLByWorkflow 通过工单执行非 DQL 语句
func (cu *CloudbeaverUsecase) executeNonDQLByWorkflow(ctx context.Context, c echo.Context, dbService *DBService, params *graphql.RawParams, auditResults cloudbeaver.AuditResults) ([]byte, error) {
	// 1. 获取当前用户 ID
	currentUserUid, _ := c.Get(dmsUserIdKey).(string)
	if currentUserUid == "" {
		return nil, c.JSON(http.StatusOK, newResp(ctx, "get dms user id", "failed", "get dms user id failed"))
	}

	// 2. 校验工单权限
	hasPermission, err := cu.checkWorkflowPermission(ctx, currentUserUid, dbService)
	if err != nil {
		return nil, c.JSON(http.StatusOK, newResp(ctx, "check workflow permission", "failed", err.Error()))
	}
	if !hasPermission {
		return nil, c.JSON(http.StatusOK, newResp(ctx, "check workflow permission", "failed", "用户没有数据源上的创建、审批、上线工单权限"))
	}

	// 3. 获取项目信息
	project, err := cu.projectUsecase.GetProject(ctx, dbService.ProjectUID)
	if err != nil {
		return nil, c.JSON(http.StatusOK, newResp(ctx, "get project", "failed", err.Error()))
	}

	// 4. 提取工单执行参数
	execParams, err := cu.getWorkflowExecParams(c, params)
	if err != nil {
		return nil, c.JSON(http.StatusOK, newResp(ctx, "get workflow exec params", "failed", err.Error()))
	}

	// 5. 执行工单上线流程
	workflowRes, err := cu.AutoCreateAndExecuteWorkflow(ctx, project.Name, dbService, execParams.query, execParams.instanceSchema)
	if err != nil {
		err = fmt.Errorf("auto create and execute workflow failed: %w", err)
		cu.log.Error(err)
		return nil, c.JSON(http.StatusOK, newResp(ctx, "auto create and execute workflow", "failed", err.Error()))
	}
	cu.log.Infof("auto create and execute workflow, workflow_id: %s, status: %s", workflowRes.Data.WorkflowID, workflowRes.Data.WorkFlowStatus)

	// 6. 判断执行结果
	isExecFailed := !strings.Contains(workflowRes.Data.WorkFlowStatus, "finished")

	// 7. 保存操作日志
	if err := cu.SaveCbOpLogForWorkflow(c, dbService, params, auditResults.Results, auditResults.IsSuccess, workflowRes.Data.WorkflowID, isExecFailed); err != nil {
		cu.log.Errorf("save cb operation log for workflow failed: %v", err)
	}

	// 8. 返回执行结果
	if isExecFailed {
		return nil, c.JSON(http.StatusOK, newResp(ctx, "auto create and execute workflow", "workflow_failed",
			fmt.Sprintf("workflow_id:%s, workflow_status:%s", workflowRes.Data.WorkflowID, workflowRes.Data.WorkFlowStatus)))
	}

	return nil, c.JSON(http.StatusOK, newResp(ctx, "auto create and execute workflow", "workflow_success",
		fmt.Sprintf("workflow_id:%s, workflow_status:%s", workflowRes.Data.WorkflowID, workflowRes.Data.WorkFlowStatus)))
}

type InstanceForCreatingTask struct {
	InstanceName   string `json:"instance_name"`
	InstanceSchema string `json:"instance_schema"`
}

type AutoCreateAndExecuteWorkflowReq struct {
	Instances       []*InstanceForCreatingTask `json:"instances"`
	ExecMode        string                     `json:"exec_mode"`
	FileOrderMethod string                     `json:"file_order_method"`
	Sql             string                     `json:"sql"`
	Subject         string                     `json:"workflow_subject"`
	Desc            string                     `json:"desc"`
}

type AutoCreateAndExecuteWorkflowRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		WorkflowID     string `json:"workflow_id"`
		WorkFlowStatus string `json:"workflow_status"`
	} `json:"data"`
}

func (cu *CloudbeaverUsecase) AutoCreateAndExecuteWorkflow(ctx context.Context, projectName string, dbService *DBService, sql string, instanceSchema string) (*AutoCreateAndExecuteWorkflowRes, error) {
	sqleUrl, err := cu.getSQLEUrl(ctx)
	if err != nil {
		return nil, fmt.Errorf("get sqle url failed: %v", err)
	}

	project, err := cu.projectUsecase.GetProject(ctx, dbService.ProjectUID)
	if err != nil {
		return nil, fmt.Errorf("get project failed: %v", err)
	}

	if projectName == "" {
		projectName = project.Name
	}

	instances := []*InstanceForCreatingTask{
		{
			InstanceName:   dbService.Name,
			InstanceSchema: instanceSchema,
		},
	}

	req := AutoCreateAndExecuteWorkflowReq{
		Instances:       instances,
		ExecMode:        "sqls",
		FileOrderMethod: "",
		Sql:             sql,
		Subject:         fmt.Sprintf("工作台工单_%s_%s", dbService.Name, time.Now().Format("20060102150405")),
		Desc:            "通过工作台执行非DQL类型的SQL时，自动创建的工单",
	}

	instancesJSON, err := json.Marshal(req.Instances)
	if err != nil {
		return nil, fmt.Errorf("marshal instances failed: %v", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	fields := map[string]string{
		"instances":         string(instancesJSON),
		"exec_mode":         req.ExecMode,
		"file_order_method": req.FileOrderMethod,
		"sql":               req.Sql,
		"workflow_subject":  req.Subject,
		"desc":              req.Desc,
	}

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			writer.Close()
			return nil, fmt.Errorf("write field %s failed: %v", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer failed: %v", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/workflows/auto_create_and_execute", sqleUrl, projectName)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", pkgHttp.DefaultDMSToken)

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request sqle failed: %v", err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sqle request failed with status %d: %s", resp.StatusCode, string(result))
	}

	var reply AutoCreateAndExecuteWorkflowRes
	if err := json.Unmarshal(result, &reply); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %v", err)
	}

	if reply.Code != 0 {
		return nil, fmt.Errorf("sqle returned error code %d: %s", reply.Code, reply.Message)
	}

	return &reply, nil
}

func UnmarshalGraphQLResponseNavNodeChildren(body []byte, resp *cloudbeaver.NavNodeChildrenResponse) error {
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return err // 真正 JSON 格式错误时才报错
	}

	if len(gqlResp.Errors) > 0 {
		// GraphQL 执行错误
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	// 再解析 Data 成真正结构
	if err := json.Unmarshal(body, resp); err != nil {
		return err
	}
	if resp == nil || resp.Data.NavNodeChildren == nil {
		return fmt.Errorf("GraphQL error: %v", gqlResp)
	}
	return nil
}
