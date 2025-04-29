package service

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	dmsMiddleware "github.com/actiontech/dms/internal/apiserver/middleware"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/pkg/locale"
	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	dmsV2 "github.com/actiontech/dms/pkg/dms-common/api/dms/v2"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	commonLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/go-kratos/kratos/v2/log"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *APIServer) initRouter() error {
	s.echo.GET("/swagger/*", s.DMSController.SwaggerHandler, SwaggerMiddleWare)

	v1 := s.echo.Group(dmsV1.CurrentGroupVersion)
	v2 := s.echo.Group(dmsV2.CurrentGroupVersion)
	// DMS RESTful resource
	{
		{
			v1.GET("/dms/basic_info", s.DMSController.GetBasicInfo)
			v1.GET(biz.PersonalizationUrl, s.DMSController.GetStaticLogo)
			v1.POST("/dms/personalization", s.DMSController.Personalization)
			v1.GET("/dms/db_services/driver_options", s.DMSController.ListDBServiceDriverOption)
			v1.GET("/dms/db_services", s.DeprecatedBy(dmsV2.GroupV2))
			v1.GET("/dms/db_services/tips", s.DMSController.ListGlobalDBServicesTips)
		}
		{
			v2.GET("/dms/db_services", s.DMSController.ListGlobalDBServicesV2)
		}

		dmsProxyV1 := v1.Group(dmsV1.ProxyRouterGroup)
		dmsProxyV1.POST("", s.DMSController.RegisterDMSProxyTarget)

		dmsPluginV1 := v1.Group(dmsV1.PluginRouterGroup)
		dmsPluginV1.POST("", s.DMSController.RegisterDMSPlugin)

		dbServiceV1 := v1.Group(dmsV1.DBServiceRouterGroup)
		{
			dbServiceV1.POST("", s.DeprecatedBy(dmsV2.GroupV2))
			dbServiceV1.GET("", s.DMSController.ListDBServicesV2) // 兼容jet brain插件，临时开放v1接口
			dbServiceV1.GET("/tips", s.DMSController.ListDBServiceTips)
			dbServiceV1.DELETE("/:db_service_uid", s.DMSController.DelDBService)
			dbServiceV1.PUT("/:db_service_uid", s.DeprecatedBy(dmsV2.GroupV2))
			dbServiceV1.POST("/connection", s.DMSController.CheckDBServiceIsConnectable)
			dbServiceV1.POST("/:db_service_uid/connection", s.DMSController.CheckDBServiceIsConnectableById)
			dbServiceV1.POST("/connections", s.DMSController.CheckProjectDBServicesConnections)
			dbServiceV1.POST("/import_check", s.DeprecatedBy(dmsV2.GroupV2))
			dbServiceV1.POST("/import", s.DeprecatedBy(dmsV2.GroupV2))
		}

		dbServiceV2 := v2.Group(dmsV2.DBServiceRouterGroup)
		{
			dbServiceV2.POST("", s.DMSController.AddDBServiceV2)
			dbServiceV2.POST("/import_check", s.DMSController.ImportDBServicesOfOneProjectCheckV2)
			dbServiceV2.POST("/import", s.DMSController.ImportDBServicesOfOneProjectV2)
			dbServiceV2.PUT("/:db_service_uid", s.DMSController.UpdateDBServiceV2)
			dbServiceV2.GET("", s.DMSController.ListDBServicesV2)
		}
		environmentTagV1 := v1.Group(dmsV1.DBEnvironmentTagGroup)
		environmentTagV1.POST("", s.DMSController.CreateEnvironmentTag)
		environmentTagV1.GET("", s.DMSController.ListEnvironmentTags)
		environmentTagV1.PUT("/:environment_tag_uid", s.DMSController.UpdateEnvironmentTag)
		environmentTagV1.DELETE("/:environment_tag_uid", s.DMSController.DeleteEnvironmentTag)

		dbServiceSyncTaskV1 := v1.Group("/dms/db_service_sync_tasks")
		dbServiceSyncTaskV1.GET("/tips", s.DMSController.ListDBServiceSyncTaskTips)
		dbServiceSyncTaskV1.GET("", s.DMSController.ListDBServiceSyncTasks)
		dbServiceSyncTaskV1.POST("", s.DMSController.AddDBServiceSyncTask)
		dbServiceSyncTaskV1.GET("/:db_service_sync_task_uid", s.DMSController.GetDBServiceSyncTask)
		dbServiceSyncTaskV1.PUT("/:db_service_sync_task_uid", s.DMSController.UpdateDBServiceSyncTask)
		dbServiceSyncTaskV1.DELETE("/:db_service_sync_task_uid", s.DMSController.DeleteDBServiceSyncTask)
		dbServiceSyncTaskV1.POST("/:db_service_sync_task_uid/sync", s.DMSController.SyncDBServices)

		userV1 := v1.Group(dmsV1.UserRouterGroup)
		userV1.POST("", s.DMSController.AddUser)
		userV1.GET("", s.DMSController.ListUsers)
		userV1.GET("/:user_uid", s.DMSController.GetUser)
		userV1.DELETE("/:user_uid", s.DMSController.DelUser)
		userV1.PUT("/:user_uid", s.DMSController.UpdateUser)
		userV1.GET(dmsV1.GetUserOpPermissionRouterWithoutPrefix(":user_uid"), s.DMSController.GetUserOpPermission)
		userV1.PUT("", s.DMSController.UpdateCurrentUser)
		userV1.POST("/gen_token", s.DMSController.GenAccessToken)
		userV1.POST("/verify_user_login", s.DMSController.VerifyUserLogin)

		sessionv1 := v1.Group(dmsV1.SessionRouterGroup)
		sessionv1.POST("", s.DMSController.AddSession)
		sessionv1.GET("/user", s.DMSController.GetUserBySession)
		sessionv1.DELETE("", s.DMSController.DelSession)
		sessionv1.POST("/refresh", s.DMSController.RefreshSession)

		userGroupV1 := v1.Group("/dms/user_groups")
		userGroupV1.POST("", s.DMSController.AddUserGroup)
		userGroupV1.GET("", s.DMSController.ListUserGroups)
		userGroupV1.DELETE("/:user_group_uid", s.DMSController.DelUserGroup)
		userGroupV1.PUT("/:user_group_uid", s.DMSController.UpdateUserGroup)

		roleV1 := v1.Group("/dms/roles")
		roleV1.POST("", s.DMSController.AddRole)
		roleV1.GET("", s.DMSController.ListRoles)
		roleV1.DELETE("/:role_uid", s.DMSController.DelRole)
		roleV1.PUT("/:role_uid", s.DMSController.UpdateRole)

		memberV1 := v1.Group(dmsV1.MemberRouterGroup)
		memberV1.POST("", s.DMSController.AddMember)
		memberV1.GET("/tips", s.DMSController.ListMemberTips)
		memberV1.GET("", s.DMSController.ListMembers)
		memberV1.GET(dmsV1.MemberForInternalRouterSuffix, s.DMSController.ListMembersForInternal)
		memberV1.DELETE("/:member_uid", s.DMSController.DelMember)
		memberV1.PUT("/:member_uid", s.DMSController.UpdateMember)

		memberGroupV1 := v1.Group("/dms/projects/:project_uid/member_groups")
		memberGroupV1.GET("", s.DMSController.ListMemberGroups)
		memberGroupV1.GET("/:member_group_uid", s.DMSController.GetMemberGroup)
		memberGroupV1.POST("", s.DMSController.AddMemberGroup)
		memberGroupV1.PUT("/:member_group_uid", s.DMSController.UpdateMemberGroup)
		memberGroupV1.DELETE("/:member_group_uid", s.DMSController.DeleteMemberGroup)

		opPermissionV1 := v1.Group("/dms/op_permissions")
		opPermissionV1.GET("", s.DMSController.ListOpPermissions)

		projectV1 := v1.Group(dmsV2.ProjectRouterGroup)
		projectV1.GET("", s.DMSController.ListProjectsV2) // 兼容jet brain插件，临时开放v1接口
		projectV1.POST("", s.DeprecatedBy(dmsV2.GroupV2))
		projectV1.DELETE("/:project_uid", s.DMSController.DelProject)
		projectV1.PUT("/:project_uid", s.DeprecatedBy(dmsV2.GroupV2))
		projectV1.PUT("/:project_uid/archive", s.DMSController.ArchiveProject)
		projectV1.PUT("/:project_uid/unarchive", s.DMSController.UnarchiveProject)
		projectV1.POST("/import", s.DeprecatedBy(dmsV2.GroupV2))
		projectV1.GET("/import_template", s.DMSController.GetImportProjectsTemplate)
		projectV1.POST("/preview_import", s.DeprecatedBy(dmsV2.GroupV2))
		projectV1.GET("/export", s.DMSController.ExportProjects)
		projectV1.GET("/tips", s.DMSController.GetProjectTips)
		projectV1.GET("/import_db_services_template", s.DMSController.GetImportDBServicesTemplate)
		projectV1.POST("/import_db_services_check", s.DeprecatedBy(dmsV2.GroupV2))
		projectV1.POST("/import_db_services", s.DeprecatedBy(dmsV2.GroupV2))
		projectV1.POST("/db_services_connection", s.DMSController.DBServicesConnection)
		projectV1.POST("/db_services_connections", s.DMSController.CheckGlobalDBServicesConnections)
		projectV1.POST("/business_tags", s.DMSController.CreateBusinessTag)
		projectV1.GET("/business_tags", s.DMSController.ListBusinessTags)
		projectV1.PUT("/business_tags/:business_tag_uid", s.DMSController.UpdateBusinessTag)
		projectV1.DELETE("/business_tags/:business_tag_uid", s.DMSController.DeleteBusinessTag)

		resourceOverviewV1 := v1.Group("/dms/resource_overview")
		resourceOverviewV1.GET("/statistics", s.DMSController.GetResourceOverviewStatistics)
		resourceOverviewV1.GET("/resource_type_distribution", s.DMSController.GetResourceOverviewResourceTypeDistribution)
		resourceOverviewV1.GET("/topology", s.DMSController.GetResourceOverviewTopology)
		resourceOverviewV1.GET("/resource_list", s.DMSController.GetResourceOverviewResourceList)
		resourceOverviewV1.GET("/download", s.DMSController.DownloadResourceOverviewList)

		// oauth2 interface does not require login authentication
		oauth2V1 := v1.Group("/dms/oauth2")
		oauth2V1.GET("/tips", s.DMSController.GetOauth2Tips)
		oauth2V1.GET("/link", s.DMSController.Oauth2Link)
		oauth2V1.GET("/callback", s.DMSController.Oauth2Callback)
		oauth2V1.POST("/user/bind", s.DMSController.BindOauth2User)
		oauth2V1.POST(biz.BackChannelLogoutUri, s.DMSController.BackChannelLogout)

		// company notice
		companyNoticeV1 := v1.Group("/dms/company_notice")
		companyNoticeV1.GET("", s.DMSController.GetCompanyNotice)
		companyNoticeV1.PATCH("", s.DMSController.UpdateCompanyNotice) /* TODO AdminUserAllowed()*/

		configurationV1 := v1.Group("/dms/configurations")
		configurationV1.GET("/login/tips", s.DMSController.GetLoginTips)
		configurationV1.PATCH("/login", s.DMSController.UpdateLoginConfiguration)       /* TODO AdminUserAllowed()*/
		configurationV1.GET("/oauth2", s.DMSController.GetOauth2Configuration)          /* TODO AdminUserAllowed()*/
		configurationV1.PATCH("/oauth2", s.DMSController.UpdateOauth2Configuration)     /* TODO AdminUserAllowed()*/
		configurationV1.GET("/ldap", s.DMSController.GetLDAPConfiguration)              /* TODO AdminUserAllowed()*/
		configurationV1.PATCH("/ldap", s.DMSController.UpdateLDAPConfiguration)         /* TODO AdminUserAllowed()*/
		configurationV1.GET("/smtp", s.DMSController.GetSMTPConfiguration)              /* TODO AdminUserAllowed()*/
		configurationV1.PATCH("/smtp", s.DMSController.UpdateSMTPConfiguration)         /* TODO AdminUserAllowed()*/
		configurationV1.POST("/smtp/test", s.DMSController.TestSMTPConfiguration)       /* TODO AdminUserAllowed()*/
		configurationV1.GET("/wechat", s.DMSController.GetWeChatConfiguration)          /* TODO AdminUserAllowed()*/
		configurationV1.PATCH("/wechat", s.DMSController.UpdateWeChatConfiguration)     /* TODO AdminUserAllowed()*/
		configurationV1.POST("/wechat/test", s.DMSController.TestWeChatConfiguration)   /* TODO AdminUserAllowed()*/
		configurationV1.GET("/feishu", s.DMSController.GetFeishuConfiguration)          /* TODO AdminUserAllowed()*/
		configurationV1.PATCH("/feishu", s.DMSController.UpdateFeishuConfiguration)     /* TODO AdminUserAllowed()*/
		configurationV1.POST("/feishu/test", s.DMSController.TestFeishuConfig)          /* TODO AdminUserAllowed()*/
		configurationV1.GET("/webhook", s.DMSController.GetWebHookConfiguration)        /* TODO AdminUserAllowed()*/
		configurationV1.PATCH("/webhook", s.DMSController.UpdateWebHookConfiguration)   /* TODO AdminUserAllowed()*/
		configurationV1.POST("/webhook/test", s.DMSController.TestWebHookConfiguration) /* TODO AdminUserAllowed()*/
		configurationV1.GET("/sql_query", s.CloudbeaverController.GetSQLQueryConfiguration)
		configurationV1.GET("/sms", s.DMSController.GetSmsConfiguration) /* TODO AdminUserAllowed()*/
		configurationV1.POST("/sms/test", s.DMSController.TestSmsConfiguration)
		configurationV1.PATCH("/sms", s.DMSController.UpdateSmsConfiguration)
		configurationV1.POST("/sms/send_code", s.DMSController.SendSmsCode)
		configurationV1.POST("/sms/verify_code", s.DMSController.VerifySmsCode)

		configurationV1.GET("/license", s.DMSController.GetLicense)            /* TODO AdminUserAllowed()*/
		configurationV1.POST("/license", s.DMSController.SetLicense)           /* TODO AdminUserAllowed()*/
		configurationV1.GET("/license/info", s.DMSController.GetLicenseInfo)   /* TODO AdminUserAllowed()*/
		configurationV1.POST("/license/check", s.DMSController.CheckLicense)   /* TODO AdminUserAllowed()*/
		configurationV1.GET("/license/usage", s.DMSController.GetLicenseUsage) /* TODO AdminUserAllowed()*/
		// notify
		notificationV1 := v1.Group(dmsV1.NotificationRouterGroup)
		notificationV1.POST("", s.DMSController.Notify) /* TODO AdminUserAllowed()*/
		// webhook
		webhookV1 := v1.Group(dmsV1.WebHookRouterGroup)
		webhookV1.POST("", s.DMSController.WebHookSendMessage) /* TODO AdminUserAllowed()*/

		dataExportWorkflowsV1 := v1.Group("/dms/projects/:project_uid/data_export_workflows")
		dataExportWorkflowsV1.POST("", s.DMSController.AddDataExportWorkflow)
		dataExportWorkflowsV1.GET("", s.DMSController.ListDataExportWorkflows)
		dataExportWorkflowsV1.GET("/:data_export_workflow_uid", s.DMSController.GetDataExportWorkflow)
		dataExportWorkflowsV1.POST("/:data_export_workflow_uid/approve", s.DMSController.ApproveDataExportWorkflow)
		dataExportWorkflowsV1.POST("/:data_export_workflow_uid/reject", s.DMSController.RejectDataExportWorkflow)
		dataExportWorkflowsV1.POST("/:data_export_workflow_uid/export", s.DMSController.ExportDataExportWorkflow)
		dataExportWorkflowsV1.POST("/cancel", s.DMSController.CancelDataExportWorkflow)

		dataExportTaskV1 := v1.Group("/dms/projects/:project_uid/data_export_tasks")
		dataExportTaskV1.POST("", s.DMSController.AddDataExportTask)
		dataExportTaskV1.GET("", s.DMSController.BatchGetDataExportTask)
		dataExportTaskV1.GET("/:data_export_task_uid/data_export_task_sqls", s.DMSController.ListDataExportTaskSQLs)
		dataExportTaskV1.GET("/:data_export_task_uid/data_export_task_sqls/download", s.DMSController.DownloadDataExportTaskSQLs)
		dataExportTaskV1.GET("/:data_export_task_uid/download", s.DMSController.DownloadDataExportTask)

		cbOperationLogsV1 := v1.Group("/dms/projects/:project_uid/cb_operation_logs")
		cbOperationLogsV1.GET("", s.DMSController.ListCBOperationLogs)
		cbOperationLogsV1.GET("/export", s.DMSController.ExportCBOperationLogs)
		cbOperationLogsV1.GET("/tips", s.DMSController.GetCBOperationLogTips)

		maskingV1 := v1.Group("/dms/masking")
		maskingV1.GET("/rules", s.DMSController.ListMaskingRules)

		getwayV1 := v1.Group("/dms/gateways")

		getwayV1.POST("", s.DMSController.AddGateway)
		getwayV1.DELETE("/:gateway_id", s.DMSController.DeleteGateway)
		getwayV1.PUT("/:gateway_id", s.DMSController.UpdateGateways)
		getwayV1.GET("/:gateway_id", s.DMSController.GetGateway)
		getwayV1.GET("", s.DMSController.ListGateways)
		getwayV1.GET("/tips", s.DMSController.GetGatewayTips)
		getwayV1.PUT("/", s.DMSController.SyncGateways, s.DMSController.DMS.GatewayUsecase.Broadcast())

		if s.CloudbeaverController.CloudbeaverService.CloudbeaverUsecase.IsCloudbeaverConfigured() {
			cloudbeaverV1 := s.echo.Group(s.CloudbeaverController.CloudbeaverService.CloudbeaverUsecase.GetRootUri())
			targets, err := s.CloudbeaverController.CloudbeaverService.ProxyUsecase.GetCloudbeaverProxyTarget()
			if err != nil {
				return err
			}

			cloudbeaverV1.Use(s.CloudbeaverController.CloudbeaverService.CloudbeaverUsecase.Login())
			cloudbeaverV1.Use(s.CloudbeaverController.CloudbeaverService.CloudbeaverUsecase.GraphQLDistributor())
			cloudbeaverV1.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
				Skipper:  middleware.DefaultSkipper,
				Balancer: middleware.NewRandomBalancer(targets),
			}))
		}
	}

	{
		projectV2 := v2.Group(dmsV2.ProjectRouterGroup)
		projectV2.POST("", s.DMSController.AddProjectV2)
		projectV2.GET("", s.DMSController.ListProjectsV2)
		projectV2.PUT("/:project_uid", s.DMSController.UpdateProjectV2)
		projectV2.POST("/import", s.DMSController.ImportProjectsV2)
		projectV2.POST("/preview_import", s.DMSController.PreviewImportProjectsV2)
		projectV2.POST("/import_db_services_check", s.DMSController.ImportDBServicesOfProjectsCheckV2)
		projectV2.POST("/import_db_services", s.DMSController.ImportDBServicesOfProjectsV2)
	}
	return nil
}

func SwaggerMiddleWare(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// swagger 请求分为两种,一种是swagger html页面请求,一种是swagger json请求. eg:
		// swagger/index.html 获取html
		// swagger/dms/doc.yaml 获取json
		hasPkPrefix := strings.HasPrefix(c.Request().RequestURI, "/swagger/index.html?urls.primaryName=")
		if hasPkPrefix {
			// 为了避免404
			c.Request().RequestURI = "/swagger/index.html"
		}

		return next(c)
	}
}

// 检查 reply 是否是 Gzip 数据
func isGzip(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// 解码 Gzip 数据
func decodeGzip(data []byte) string {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Sprintf("Gzip decode error: %v", err)
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Sprintf("Read Gzip data error: %v", err)
	}
	return string(decoded)
}

func (s *APIServer) installMiddleware() error {
	// Middleware

	s.echo.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Request().RequestURI, dmsV1.GroupV1)
		},
		Handler: func(context echo.Context, req []byte, reply []byte) {
			userUid, _ := jwt.GetUserUidStrFromContext(context)

			// 将请求转为字符串
			reqStr := string(req)
			// 尝试解码 reply（gzip 格式）
			var replyStr string
			if isGzip(reply) {
				replyStr = decodeGzip(reply)
			} else {
				replyStr = string(reply)
			}

			commonLog.NewHelper(s.logger).Log(
				commonLog.LevelDebug,
				"middleware.uri", context.Request().RequestURI,
				"user_id", userUid,
				"req", reqStr, // 输出处理后的请求数据
				"reply", replyStr, // 输出处理后的响应数据
			)
		},
	}))

	s.echo.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: middleware.DefaultSkipper,
		Level:   5,
	}))

	s.echo.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: middleware.Skipper(func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().URL.Path, s.CloudbeaverController.CloudbeaverService.CloudbeaverUsecase.GetRootUri()) {
				return true
			}
			if strings.HasPrefix(c.Request().URL.Path, "/provision/v") ||
				strings.HasPrefix(c.Request().URL.Path, "/sqle/v") {
				return true
			}

			return strings.Contains(c.Request().URL.Path, "/swagger")
		}),
		Root:   "static",
		Index:  "index.html",
		HTML5:  true,
		Browse: false,
	}))
	s.echo.Any("", echo.NotFoundHandler)
	s.echo.Any("/*", echo.NotFoundHandler)

	s.echo.Use(dmsMiddleware.JWTTokenAdapter())

	s.echo.Use(echojwt.WithConfig(echojwt.Config{
		Skipper: middleware.Skipper(func(c echo.Context) bool {
			logger := log.NewHelper(log.With(pkgLog.NewKLogWrapper(s.logger), "middleware", "jwt"))
			if strings.HasSuffix(c.Request().RequestURI, dmsV1.SessionRouterGroup) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/sessions/refresh" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/oauth2" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/configurations/login/tips" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/personalization/logo") ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/configurations/license" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/users/verify_user_login" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/configurations/sms/send_code" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/configurations/sms/verify_code" /* TODO 使用统一方法skip */) ||
				!(strings.HasPrefix(c.Request().RequestURI, dmsV1.CurrentGroupVersion) ||
					strings.HasPrefix(c.Request().RequestURI, dmsV2.CurrentGroupVersion)) {
				logger.Debugf("skipper url jwt check: %v", c.Request().RequestURI)
				return true
			}
			return false
		}),
		SigningKey:  dmsV1.JwtSigningKey,
		TokenLookup: "cookie:dms-token,header:Authorization:Bearer ", // tell the middleware where to get token: from cookie and header,
	}))

	s.echo.Use(dmsMiddleware.LicenseAdapter(s.DMSController.DMS.LicenseUsecase))

	s.echo.Use(s.DMSController.DMS.AuthAccessTokenUseCase.CheckLatestAccessToken())

	s.echo.Use(s.DMSController.DMS.Oauth2ConfigurationUsecase.CheckBackChannelLogoutEvent())

	s.echo.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper:  s.DMSController.DMS.DmsProxyUsecase.GetEchoProxySkipper(),
		Balancer: s.DMSController.DMS.DmsProxyUsecase.GetEchoProxyBalancer(),
		Rewrite:  s.DMSController.DMS.DmsProxyUsecase.GetEchoProxyRewrite(),
	}))

	s.echo.Use(locale.Bundle.EchoMiddlewareByCustomFunc(
		s.DMSController.DMS.UserUsecase.GetUserLanguageByEchoCtx,
		i18nPkg.GetLangByAcceptLanguage,
	))

	return nil
}

func (s *APIServer) installController() error {

	cloudbeaverController, err := NewCloudbeaverController(s.logger, s.opts)
	if nil != err {
		return fmt.Errorf("failed to create CloudbeaverController: %v", err)
	}

	DMSController, err := NewDMSController(s.logger, s.opts, cloudbeaverController.CloudbeaverService)
	if nil != err {
		return fmt.Errorf("failed to create DMSController: %v", err)
	}

	s.DMSController = DMSController
	s.CloudbeaverController = cloudbeaverController

	// s.AuthController.RegisterPlugin(s.DMSController.GetRegisterPluginFn())
	return nil
}

func (s *APIServer) Shutdown() error {
	if err := s.echo.Close(); nil != err {
		return fmt.Errorf("failed to close echo: %v", err)
	}
	if err := s.DMSController.shutdownCallback(); nil != err {
		return fmt.Errorf("failed to shutdown dmsController: %v", err)
	}
	return nil
}

// DeprecatedBy is a controller used to mark deprecated and used to replace the original controller.
func (s *APIServer) DeprecatedBy(version string) func(echo.Context) error {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf(
			"the API has been deprecated, please using the %s version", version))
	}
}
