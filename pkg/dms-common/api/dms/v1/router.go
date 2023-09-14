package v1

import (
	"fmt"
	"strings"
)

// login config
var (
	JwtSigningKey = []byte("actiontech dms secret")
)

// router
var (
	SessionRouterGroup            = "/dms/sessions"
	UserRouterGroup               = "/dms/users"
	DBServiceRouterGroup          = "/dms/db_services"
	ProxyRouterGroup              = "/dms/proxys"
	PluginRouterGroup             = "/dms/plugins"
	MemberRouterGroup             = "/dms/projects/:project_uid/members"
	ProjectRouterGroup            = "/dms/projects"
	NotificationRouterGroup       = "/dms/notifications"
	WebHookRouterGroup            = "/dms/webhooks"
	MemberForInternalRouterSuffix = "/internal"
	InternalDBServiceRouterGroup  = "/internal/db_services"
)

// api group
var (
	GroupV1             = "/v1"
	CurrentGroupVersion = GroupV1
)

func GetUserOpPermissionRouter(userUid string) string {
	return fmt.Sprintf("%s%s/%s/op_permission", CurrentGroupVersion, UserRouterGroup, userUid)
}

func GetUserOpPermissionRouterWithoutPrefix(userUid string) string {
	router := GetUserOpPermissionRouter(userUid)
	return strings.TrimPrefix(strings.TrimPrefix(router, CurrentGroupVersion), UserRouterGroup)
}

func GetDBServiceRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, DBServiceRouterGroup)
}

func GetUserRouter(userUid string) string {
	return fmt.Sprintf("%s%s/%s", CurrentGroupVersion, UserRouterGroup, userUid)
}

func GetUsersRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, UserRouterGroup)
}

func GetListMembersForInternalRouter(projectId string) string {
	return fmt.Sprintf("%s%s%s", CurrentGroupVersion, strings.Replace(MemberRouterGroup, ":project_uid", projectId, 1), MemberForInternalRouterSuffix)
}

func GetProxyRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, ProxyRouterGroup)
}

func GetPluginRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, PluginRouterGroup)
}

func GetProjectsRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, ProjectRouterGroup)
}

func GetNotificationRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, NotificationRouterGroup)
}

func GetWebHooksRouter() string {
	return fmt.Sprintf("%s%s", CurrentGroupVersion, WebHookRouterGroup)
}

func GetDBConnectionAbleRouter() string {
	return fmt.Sprintf("%s%s/connection", CurrentGroupVersion, InternalDBServiceRouterGroup)
}
