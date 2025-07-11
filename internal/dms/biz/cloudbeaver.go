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
	"regexp"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/resolver"
	"github.com/actiontech/dms/internal/pkg/locale"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/actiontech/dms/pkg/dms-common/pkg/aes"
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

const dmsUserIdKey = "dmsToken"

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
				return c.JSON(http.StatusUnauthorized, fmt.Sprintf("get token detail failed, err:%v", err))
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

type TaskInfo struct {
	Data struct {
		TaskInfo *model.AsyncTaskInfo `json:"taskInfo"`
	} `json:"data"`
}

var taskIDAssocUid sync.Map
var taskIdAssocMasking sync.Map

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
			if c.Request().RequestURI != path.Join(CbRootUri, CbGqlApi) {
				return next(c)
			}

			// 复制请求体内容
			var reqBody = make([]byte, 0)
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

			//  统一拦截响应
			srw := newSmartResponseWriter(c)
			cloudbeaverResBuf := srw.Buffer
			c.Response().Writer = srw

			defer func() {
				// 能否拦截所有error场景，并让前端重新刷新页面

				// 对响应体做分析
				if handleErrResponse(c, srw, cloudbeaverResBuf.Bytes()) {
					return
				} else {
					// 如果没错误响应，写出原响应内容
					if srw.status != 0 {
						srw.original.WriteHeader(srw.status)
					}
					_, writeErr := srw.original.Write(cloudbeaverResBuf.Bytes())
					if writeErr != nil {
						c.Logger().Error("Failed to write original response:", writeErr)
					}
				}
			}()

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
							RuleTemplateName: dbService.SQLEConfig.RuleTemplateName,
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
				var resWrite *responseProcessWriter
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

						resWrite = &responseProcessWriter{tmp: &bytes.Buffer{}, ResponseWriter: c.Response().Writer}
						c.Response().Writer = resWrite

						if err = next(c); err != nil {
							return nil, err
						}

						if ok && resp.IsSuccess {
							var taskInfo TaskInfo
							err = UnmarshalGraphQLResponse(resWrite.tmp.Bytes(), &taskInfo)
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
							err = cu.UpdateCbOpResult(c, resWrite.tmp, params, ctx)
							if err != nil {
								cu.log.Errorf("update cb operation result err: %v", err)
							}
						}

						return resWrite.tmp.Bytes(), nil
					}
				}

				// 创建GraphQL可执行schema
				g := resolver.NewExecutableSchema(resolver.Config{
					Resolvers: cloudbeaver.NewResolverImpl(c, cloudbeaverNext, cu.dataMaskingUseCase.SQLExecuteResultsDataMasking, enableMasking),
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
					return res.Errors
				}
				if !cloudbeaverHandle.NeedModifyRemoteRes {
					return nil
				} else {
					// 设置响应头
					header := resWrite.ResponseWriter.Header()
					b, err := json.Marshal(res)
					if err != nil {
						return err
					}
					header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
					// 写入响应
					_, err = resWrite.ResponseWriter.Write(b)
					return err
				}
			}
			next(c)
			return
		}
	}
}

const (
	SQLContextNotFoundCode = 508
)

// 示例：分析处理函数
func handleErrResponse(c echo.Context, srw *smartResponseWriter, data []byte) bool {
	// 你可以解析 JSON、做日志记录、统计等操作
	println("Captured response:", string(data))

	// 定义错误匹配正则表达式
	errorRegex := regexp.MustCompile(`"message":\s*"SQL context\s+[^"]+not found"`)

	// 检查是否匹配错误模式
	if errorRegex.Match(data) {
		baseResp := `{"code":%d,"message":"ok","data":[]}`
		// 构建自定义响应
		response := []byte(fmt.Sprintf(baseResp, SQLContextNotFoundCode))
		srw.original.WriteHeader(srw.status)
		_, writeErr := srw.original.Write(response)
		if writeErr != nil {
			c.Logger().Error("Failed to write original response:", writeErr)
		}
		return true
	}
	return false
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

type smartResponseWriter struct {
	echo.Response
	Buffer   *bytes.Buffer
	original http.ResponseWriter
	status   int
}

func newSmartResponseWriter(c echo.Context) *smartResponseWriter {
	buf := new(bytes.Buffer)
	return &smartResponseWriter{
		Response: *c.Response(),
		Buffer:   buf,
		original: c.Response().Writer,
	}
}

func (w *smartResponseWriter) Write(b []byte) (int, error) {
	// 写入 buffer，不立即写给客户端
	return w.Buffer.Write(b)
}

func (w *smartResponseWriter) WriteHeader(code int) {
	w.status = code
}

func (cu *CloudbeaverUsecase) isEnableSQLAudit(dbService *DBService) bool {
	return dbService.SQLEConfig.SQLQueryConfig.AuditEnabled
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

var reservedCloudbeaverUserId = map[string]struct{}{"admin": {}, "user": {}}

func (cu *CloudbeaverUsecase) createUserIfNotExist(ctx context.Context, cloudbeaverUserId string, dmsUser *User) error {
	if _, ok := reservedCloudbeaverUserId[cloudbeaverUserId]; ok {
		return fmt.Errorf("username %s is reserved， cann't be used", cloudbeaverUserId)
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
	} else {
		cloudbeaverUser, exist, err := cu.repo.GetCloudbeaverUserByID(ctx, cloudbeaverUserId)
		if err != nil {
			return err
		}

		if exist && cloudbeaverUser.DMSFingerprint == cu.userUsecase.GetUserFingerprint(dmsUser) {
			return nil
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

	cloudbeaverUser := &CloudbeaverUser{
		DMSUserID:         dmsUser.UID,
		DMSFingerprint:    cu.userUsecase.GetUserFingerprint(dmsUser),
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
	client, err := cu.getGraphQLClient(cloudbeaverUser.CloudbeaverUserID, dmsUser.Password)
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

	//获取当前用户所有已创建的连接
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

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
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
