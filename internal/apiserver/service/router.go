package service

import (
	"fmt"
	"strings"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	pkgLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	dmsMiddleware "github.com/actiontech/dms/internal/apiserver/middleware"
	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	commonLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	echojwt "github.com/labstack/echo-jwt/v4"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *APIServer) initRouter() error {
	v1 := s.echo.Group(dmsV1.CurrentGroupVersion)

	// DMS RESTful resource
	{
		v1.GET("/dms/basic_info", s.DMSController.GetBasicInfo)
		v1.GET(biz.PersonalizationUrl, s.DMSController.GetStaticLogo)
		v1.POST("/dms/personalization", s.DMSController.Personalization)
		v1.GET("/dms/db_services/driver_options", s.DMSController.ListDBServiceDriverOption)

		dmsProxyV1 := v1.Group(dmsV1.ProxyRouterGroup)
		dmsProxyV1.POST("", s.DMSController.RegisterDMSProxyTarget)

		dmsPluginV1 := v1.Group(dmsV1.PluginRouterGroup)
		dmsPluginV1.POST("", s.DMSController.RegisterDMSPlugin)

		dbServiceV1 := v1.Group(dmsV1.DBServiceRouterGroup)
		dbServiceV1.POST("", s.DMSController.AddDBService)
		dbServiceV1.GET("", s.DMSController.ListDBServices)
		dbServiceV1.GET("/tips", s.DMSController.ListDBServiceTips)
		dbServiceV1.DELETE("/:db_service_uid", s.DMSController.DelDBService)
		dbServiceV1.PUT("/:db_service_uid", s.DMSController.UpdateDBService)
		dbServiceV1.POST("/connection", s.DMSController.CheckDBServiceIsConnectable)
		dbServiceV1.POST("/:db_service_uid/connection", s.DMSController.CheckDBServiceIsConnectableById)

		DatabaseSourceServiceV1 := v1.Group("/dms/projects/:project_uid/database_source_services")
		DatabaseSourceServiceV1.GET("/tips", s.DMSController.ListDatabaseSourceServiceTips)
		DatabaseSourceServiceV1.POST("/:database_source_service_uid/sync", s.DMSController.SyncDatabaseSourceService)
		DatabaseSourceServiceV1.GET("", s.DMSController.ListDatabaseSourceServices)
		DatabaseSourceServiceV1.GET("/:database_source_service_uid", s.DMSController.GetDatabaseSourceService)
		DatabaseSourceServiceV1.POST("", s.DMSController.AddDatabaseSourceService)
		DatabaseSourceServiceV1.PUT("/:database_source_service_uid", s.DMSController.UpdateDatabaseSourceService)
		DatabaseSourceServiceV1.DELETE("/:database_source_service_uid", s.DMSController.DeleteDatabaseSourceService)

		userV1 := v1.Group(dmsV1.UserRouterGroup)
		userV1.POST("", s.DMSController.AddUser)
		userV1.GET("", s.DMSController.ListUsers)
		userV1.GET("/:user_uid", s.DMSController.GetUser)
		userV1.DELETE("/:user_uid", s.DMSController.DelUser)
		userV1.PUT("/:user_uid", s.DMSController.UpdateUser)
		userV1.GET(dmsV1.GetUserOpPermissionRouterWithoutPrefix(":user_uid"), s.DMSController.GetUserOpPermission)
		userV1.PUT("", s.DMSController.UpdateCurrentUser)

		sessionv1 := v1.Group(dmsV1.SessionRouterGroup)
		sessionv1.POST("", s.DMSController.AddSession)
		sessionv1.GET("/user", s.DMSController.GetUserBySession)
		sessionv1.DELETE("", s.DMSController.DelSession)

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

		projectV1 := v1.Group(dmsV1.ProjectRouterGroup)
		projectV1.GET("", s.DMSController.ListProjects)
		projectV1.POST("", s.DMSController.AddProject)
		projectV1.DELETE("/:project_uid", s.DMSController.DelProject)
		projectV1.PUT("/:project_uid", s.DMSController.UpdateProject)
		projectV1.PUT("/:project_uid/archive", s.DMSController.ArchiveProject)
		projectV1.PUT("/:project_uid/unarchive", s.DMSController.UnarchiveProject)

		// oauth2 interface does not require login authentication
		oauth2V1 := v1.Group("/dms/oauth2")
		oauth2V1.GET("/tips", s.DMSController.GetOauth2Tips)
		oauth2V1.GET("/link", s.DMSController.Oauth2Link)
		oauth2V1.GET("/callback", s.DMSController.Oauth2Callback)
		oauth2V1.POST("/user/bind", s.DMSController.BindOauth2User)

		// company notice
		companyNoticeV1 := v1.Group("/dms/company_notice")
		companyNoticeV1.GET("", s.DMSController.GetCompanyNotice)
		companyNoticeV1.PATCH("", s.DMSController.UpdateCompanyNotice) /* TODO AdminUserAllowed()*/

		configurationV1 := v1.Group("/dms/configurations")
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

		configurationV1.GET("/license", s.DMSController.GetLicense)          /* TODO AdminUserAllowed()*/
		configurationV1.POST("/license", s.DMSController.SetLicense)         /* TODO AdminUserAllowed()*/
		configurationV1.GET("/license/info", s.DMSController.GetLicenseInfo) /* TODO AdminUserAllowed()*/
		configurationV1.POST("/license/check", s.DMSController.CheckLicense) /* TODO AdminUserAllowed()*/

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
		dataExportTaskV1.GET("/:data_export_task_uid/download", s.DMSController.DownloadDataExportTask)

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
	return nil
}

func (s *APIServer) installMiddleware() error {
	// Middleware
	s.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `${time_custom} ECHO id:${id}, remote_ip:${remote_ip}, ` +
			`host:${host}, method:${method}, uri:${uri}, user_agent:${user_agent}, ` +
			`status:${status}, error:${error}, latency:${latency}, latency_human:${latency_human}` +
			`, bytes_in:${bytes_in}, bytes_out:${bytes_out}` + "\n",
		CustomTimeFormat: pkgLog.LogTimeLayout,
	}))

	s.echo.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Request().RequestURI, dmsV1.GroupV1)
		},
		Handler: func(context echo.Context, req []byte, reply []byte) {
			userUid, _ := jwt.GetUserUidStrFromContext(context)
			commonLog.NewHelper(s.logger).Log(commonLog.LevelDebug, "middleware.uri", context.Request().RequestURI, "user_id", userUid, "req", string(req), "reply", string(reply))
		},
	}))

	s.echo.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: middleware.Skipper(func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().URL.Path, s.CloudbeaverController.CloudbeaverService.CloudbeaverUsecase.GetRootUri()) {
				return true
			}

			return strings.Contains(c.Request().URL.Path, "/swagger")
		}),
		Root:   "static",
		Index:  "index.html",
		HTML5:  true,
		Browse: false,
	}))

	s.echo.Use(echojwt.WithConfig(echojwt.Config{
		Skipper: middleware.Skipper(func(c echo.Context) bool {
			logger := log.NewHelper(log.With(pkgLog.NewKLogWrapper(s.logger), "middleware", "jwt"))
			if strings.HasSuffix(c.Request().RequestURI, dmsV1.SessionRouterGroup) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/oauth2" /* TODO 使用统一方法skip */) ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/personalization/logo") ||
				strings.HasPrefix(c.Request().RequestURI, "/v1/dms/configurations/license" /* TODO 使用统一方法skip */) ||
				!strings.HasPrefix(c.Request().RequestURI, dmsV1.CurrentGroupVersion) {
				logger.Debugf("skipper url jwt check: %v", c.Request().RequestURI)
				return true
			}
			return false
		}),
		SigningKey:  dmsV1.JwtSigningKey,
		TokenLookup: "cookie:dms-token,header:Authorization:Bearer ", // tell the middleware where to get token: from cookie and header,
	}))

	s.echo.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper:  s.DMSController.DMS.DmsProxyUsecase.GetEchoProxySkipper(),
		Balancer: s.DMSController.DMS.DmsProxyUsecase.GetEchoProxyBalancer(),
		Rewrite:  s.DMSController.DMS.DmsProxyUsecase.GetEchoProxyRewrite(),
	}))

	s.echo.Use(dmsMiddleware.LicenseAdapter(s.DMSController.DMS.LicenseUsecase))

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
