package cloudbeaver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/actiontech/dms/internal/pkg/cloudbeaver/model"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver/resolver"

	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/labstack/echo/v4"
)

type Next func(c echo.Context) ([]byte, error)

type ResolverImpl struct {
	*resolver.Resolver
	Ctx  echo.Context
	Next Next

	// SQLExecuteResultsHandlerFn 为对SQL结果集的处理方法，具体处理逻辑为业务行为，由外部biz层定义后传入
	SQLExecuteResultsHandlerFn SQLExecuteResultsHandler
}

func NewResolverImpl(ctx echo.Context, next Next, SQLExecuteResultsHandlerFn SQLExecuteResultsHandler) *ResolverImpl {
	return &ResolverImpl{
		Ctx:                        ctx,
		Next:                       next,
		SQLExecuteResultsHandlerFn: SQLExecuteResultsHandlerFn,
	}
}

func (r *ResolverImpl) Mutation() resolver.MutationResolver {
	return &MutationResolverImpl{
		Ctx:                        r.Ctx,
		Next:                       r.Next,
		SQLExecuteResultsHandlerFn: r.SQLExecuteResultsHandlerFn,
	}
}

// Query returns generated.QueryResolver implementation.
func (r *ResolverImpl) Query() resolver.QueryResolver {
	return &QueryResolverImpl{
		Ctx:  r.Ctx,
		Next: r.Next,
	}
}

type SQLExecuteResultsHandler func(ctx context.Context, result *model.SQLExecuteInfo) error

type MutationResolverImpl struct {
	*resolver.MutationResolverImpl
	Ctx  echo.Context
	Next Next

	// SQLExecuteResultsHandlerFn 为对SQL结果集的处理方法，具体处理逻辑为业务行为，由外部biz层定义后传入
	SQLExecuteResultsHandlerFn SQLExecuteResultsHandler
}

type QueryResolverImpl struct {
	*resolver.QueryResolverImpl
	Ctx  echo.Context
	Next Next
}

func (r *QueryResolverImpl) ActiveUser(ctx context.Context) (*model.UserInfo, error) {
	data, err := r.Next(r.Ctx)
	if err != nil {
		return nil, err
	}

	resp := &struct {
		Data struct {
			User *model.UserInfo `json:"user"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}

	if resp.Data.User != nil && resp.Data.User.DisplayName != nil {
		*resp.Data.User.DisplayName = RemoveCloudbeaverUserIdPrefix(*resp.Data.User.DisplayName)
	}

	return resp.Data.User, err
}

type ContextKey string

const (
	UsernamePrefix               = "dms-"
	SQLEProxyAddrName ContextKey = "sqle_addr"
	SQLEProxyName                = "sqle"
)

func GenerateCloudbeaverUserId(name string) string {
	return UsernamePrefix + name
}

func RemoveCloudbeaverUserIdPrefix(name string) string {
	return strings.TrimPrefix(name, UsernamePrefix)
}

type auditSQLReq struct {
	InstanceType string `json:"instance_type" form:"instance_type" example:"MySQL" valid:"required"`
	// 调用方不应该关心SQL是否被完美的拆分成独立的条目, 拆分SQL由SQLE实现
	SQLContent string `json:"sql_content" form:"sql_content" example:"select * from t1; select * from t2;" valid:"required"`
	SQLType    string `json:"sql_type" form:"sql_type" example:"sql" enums:"sql,mybatis," valid:"omitempty,oneof=sql mybatis"`
}

type AuditResult struct {
	Level    string `json:"level" example:"warn"`
	Message  string `json:"message" example:"避免使用不必要的内置函数md5()"`
	RuleName string `json:"rule_name"`
}

type AuditSQLResV2 struct {
	Number      uint          `json:"number"`
	ExecSQL     string        `json:"exec_sql"`
	AuditResult []AuditResult `json:"audit_result"`
	AuditLevel  string        `json:"audit_level"`
}

type AuditResDataV2 struct {
	AuditLevel string          `json:"audit_level" enums:"normal,notice,warn,error,"`
	Score      int32           `json:"score"`
	PassRate   float64         `json:"pass_rate"`
	SQLResults []AuditSQLResV2 `json:"sql_results"`
}

type auditSQLReply struct {
	Code    int             `json:"code" example:"0"`
	Message string          `json:"message" example:"ok"`
	Data    *AuditResDataV2 `json:"data"`
}

// AuditSQL todo: this is a provisional programme that will need to be adjusted at a later stage
func (r *MutationResolverImpl) AuditSQL(ctx context.Context, sql string, connectionID string) (auditSuccess bool, result []AuditSQLResV2, err error) {
	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}
	req := auditSQLReq{
		InstanceType: "MySQL",
		SQLContent:   sql,
		SQLType:      "sql",
	}
	reply := &auditSQLReply{}

	addr := ctx.Value(SQLEProxyAddrName)
	sqleAddr, ok := addr.(string)
	if !ok {
		return false, nil, fmt.Errorf("ctx.value %s failed", SQLEProxyAddrName)
	}

	if sqleAddr == "" {
		return false, nil, fmt.Errorf("%s is empty", SQLEProxyAddrName)
	}

	if err = pkgHttp.POST(ctx, sqleAddr, header, req, reply); err != nil {
		return false, nil, err
	}
	if reply.Code != 0 {
		return false, nil, fmt.Errorf("reply code(%v) error: %v", reply.Code, reply.Message)
	}

	if reply.Data.PassRate == 0 {
		return false, reply.Data.SQLResults, nil
	}

	return true, nil, nil
}

func (r *MutationResolverImpl) AsyncSQLExecuteQuery(ctx context.Context, projectID *string, connectionID string, contextID string, sql string, resultID *string, filter *model.SQLDataFilter, dataFormat *model.ResultDataFormat, readLogs *bool) (*model.AsyncTaskInfo, error) {
	success, results, err := r.AuditSQL(ctx, sql, connectionID)
	if err != nil {
		return nil, err
	}

	if !success {
		var messages []string
		for _, sqlResult := range results {
			for _, audit := range sqlResult.AuditResult {
				messages = append(messages, audit.Message)
			}
		}

		messageStr := strings.Join(messages, ",")
		name := "SQL Audit Failed"
		return nil, r.Ctx.JSON(http.StatusOK, struct {
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
					Status:  &sql,
					Error: &model.ServerError{
						Message:    &messageStr,
						StackTrace: &messageStr,
					},
				},
			},
		})
	}

	_, err = r.Next(r.Ctx)
	if err != nil {
		return nil, err
	}

	return nil, err
}

func (r *MutationResolverImpl) AsyncSQLExecuteResults(ctx context.Context, taskID string) (*model.SQLExecuteInfo, error) {

	data, err := r.Next(r.Ctx)
	if err != nil {
		return nil, err
	}
	
	resp := &struct {
		Data struct {
			Result *model.SQLExecuteInfo `json:"result"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal sql execute info: %v", err)
	}

	if resp.Data.Result != nil {
		if err := r.SQLExecuteResultsHandlerFn(ctx, resp.Data.Result); err != nil {
			return nil, fmt.Errorf("failed to handle sql result: %v", err)
		}
	}

	return resp.Data.Result, err
}

type gqlBehavior struct {
	UseLocalHandler     bool
	NeedModifyRemoteRes bool
	Disable             bool
	// 预处理主要用于在真正使用前处理前端传递的参数, 比如需要接收int, 但收到float, 则可以在此处调整参数类型
	Preprocessing func(ctx echo.Context, params *graphql.RawParams) error
}

var GraphQLHandlerRouters = map[string] /* gql operation name */ gqlBehavior{
	"asyncSqlExecuteQuery": {
		UseLocalHandler:     true,
		NeedModifyRemoteRes: false,
		Preprocessing: func(ctx echo.Context, params *graphql.RawParams) (err error) {
			// json中没有int类型, 这将导致执行json.Unmarshal()时int会被当作float64, 从而导致后面出现类型错误的异常
			if filter, ok := params.Variables["filter"].(map[string]interface{}); ok {
				if filter["limit"] != nil {
					params.Variables["filter"].(map[string]interface{})["limit"], err = strconv.Atoi(fmt.Sprintf("%v", params.Variables["filter"].(map[string]interface{})["limit"]))
				}
			}
			return err
		},
	},
	"getSqlExecuteTaskResults": {
		UseLocalHandler:     true,
		NeedModifyRemoteRes: true,
	},
	"getActiveUser": {
		UseLocalHandler:     true,
		NeedModifyRemoteRes: true,
	}, "authLogout": {
		Disable: true,
	}, "authLogin": {
		Disable: true,
	}, "configureServer": {
		Disable: true,
	}, "createUser": {
		Disable: true,
	}, "setUserCredentials": {
		Disable: true,
	}, "enableUser": {
		Disable: true,
	}, "grantUserRole": {
		Disable: true,
	}, "setConnections": {
		Disable: true,
	}, "saveUserMetaParameters": {
		Disable: true,
	}, "deleteUser": {
		Disable: true,
	}, "createRole": {
		Disable: true,
	}, "updateRole": {
		Disable: true,
	}, "deleteRole": {
		Disable: true,
	}, "authChangeLocalPassword": {
		Disable: true,
	},
}
