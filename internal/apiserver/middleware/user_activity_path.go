package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/pkg/dms-common/api/jwt"
	"github.com/labstack/echo/v4"
)

const maxUserAgentLen = 255

var (
	snowflakeSegmentPattern = regexp.MustCompile(`^[0-9]{15,20}$`)
	numericSegmentPattern   = regexp.MustCompile(`^[0-9]+$`)
)

type modulePrefixRule struct {
	ModuleCode string
	Prefixes   []string
}

var defaultModulePrefixRules = []modulePrefixRule{
	{ModuleCode: "AUTH", Prefixes: []string{"/v1/dms/sessions", "/v1/dms/oauth2/", "/v1/dms/users/verify_user_login", "/v1/dms/configurations/login"}},
	{ModuleCode: "USER_ROLE", Prefixes: []string{"/v1/dms/users", "/v1/dms/user_groups", "/v1/dms/roles", "/v1/dms/op_permissions", "/sqle/v1/user_tips", "/sqle/v2/user_tips", "/sqle/v3/user_tips"}},
	{ModuleCode: "PROJECT", Prefixes: []string{"/v1/dms/projects", "/v2/dms/projects", "/sqle/v1/projects/", "/sqle/v2/projects/", "/sqle/v3/projects/"}},
	{ModuleCode: "DB_SERVICE", Prefixes: []string{"/v1/dms/db_services", "/v2/dms/db_services", "/v1/dms/projects/", "/v2/dms/projects/", "/v1/dms/db_service_sync_tasks", "/sqle/v1/projects/", "/sqle/v2/projects/", "/sqle/v3/projects/"}},
	{ModuleCode: "WORKFLOW", Prefixes: []string{"/sqle/v1/projects/", "/sqle/v2/projects/", "/sqle/v3/projects/", "/sqle/v1/dashboard/workflows", "/sqle/v2/dashboard/workflows", "/sqle/v1/workflows/"}},
	{ModuleCode: "DATA_EXPORT", Prefixes: []string{"/v1/dms/projects/", "/v1/dms/dashboard/data_export_workflows", "/sqle/v1/dashboard/data_export_workflows", "/sqle/v2/dashboard/data_export_workflows"}},
	{ModuleCode: "SQL_AUDIT", Prefixes: []string{"/sqle/v1/sql_audit", "/sqle/v2/sql_audit", "/sqle/v1/audit_files", "/sqle/v2/audit_files", "/sqle/v1/sql_analysis", "/sqle/v1/sql_lineage_analysis", "/sqle/v1/tasks/audits/", "/sqle/v2/tasks/audits/", "/sqle/v1/task_groups/audit"}},
	{ModuleCode: "RULE", Prefixes: []string{"/sqle/v1/rule_templates", "/sqle/v2/rule_templates", "/sqle/v1/custom_rules", "/sqle/v1/rules", "/sqle/v1/rule_knowledge", "/sqle/v1/knowledge_bases", "/sqle/v1/rule_template_tips"}},
	{ModuleCode: "AUDIT_PLAN", Prefixes: []string{"/sqle/v1/audit_plan_metas", "/sqle/v1/audit_plan_types", "/sqle/v1/projects/", "/sqle/v2/projects/", "/sqle/v3/projects/"}},
	{ModuleCode: "SQL_MANAGE", Prefixes: []string{"/sqle/v1/dashboard/sql_manage", "/sqle/v2/dashboard/sql_manage", "/sqle/v1/dashboard/sql_manages", "/sqle/v2/dashboard/sql_manages"}},
	{ModuleCode: "SQL_OPTIMIZE", Prefixes: []string{"/sqle/v1/projects/", "/sqle/v2/projects/", "/sqle/v3/projects/"}},
	{ModuleCode: "SQL_VERSION", Prefixes: []string{"/sqle/v1/projects/", "/sqle/v2/projects/", "/sqle/v3/projects/"}},
	{ModuleCode: "WORKBENCH", Prefixes: []string{"/sql_query/", "/odc_query/", "/v1/dms/configurations/sql_query", "/v1/dms/projects/"}},
	{ModuleCode: "SYS_CONFIG", Prefixes: []string{"/v1/dms/configurations/", "/sqle/v1/configurations/", "/sqle/v2/configurations/", "/v1/dms/company_notice", "/sqle/v1/company_notice", "/v1/dms/notifications", "/v1/dms/webhooks", "/v1/dms/personalization", "/v1/dms/masking/", "/v1/dms/gateways", "/sqle/v1/system/", "/sqle/v2/system/"}},
	{ModuleCode: "DASHBOARD", Prefixes: []string{"/sqle/v1/dashboard", "/sqle/v2/dashboard", "/sqle/v1/statistic/", "/v1/dms/resource_overview/", "/sqle/v1/ai_hub/"}},
	{ModuleCode: "AUDIT_LOG", Prefixes: []string{"/v1/dms/operation_records"}},
	{ModuleCode: "PROVISION", Prefixes: []string{"/provision/v1/auth/", "/provision/v1/plugin/"}},
	{ModuleCode: "INTERNAL", Prefixes: []string{"/v1/dms/proxys", "/v1/dms/plugins", "/sqle/v1/internal/", "/v1/internal/"}},
}

var staticAssetExtensions = map[string]struct{}{
	".js": {}, ".mjs": {}, ".css": {}, ".map": {},
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".svg": {}, ".ico": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".eot": {},
	".html": {}, ".htm": {}, ".txt": {}, ".xml": {},
}

var userActivityAPITrackPrefixes = []string{
	"/v1/", "/sqle/", "/sql_query/", "/odc_query/", "/provision/",
}

func shouldSkipUserActivityPath(path string) bool {
	if path == "" || path == "/" {
		return true
	}
	skipPrefixes := []string{
		"/logo",
		"/static/",
		"/assets/",
		"/fonts/",
		"/favicon",
		"/manifest",
		"/swagger",
		"/v1/dms/sessions/refresh",
		"/v1/dms/basic_info",
		"/v1/dms/personalization/",
		"/v1/dms/statistic/user_activity/",
	}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	if strings.Contains(path, "/swagger") {
		return true
	}
	if hasStaticAssetExtension(path) {
		return true
	}
	return false
}

func hasStaticAssetExtension(path string) bool {
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		if _, ok := staticAssetExtensions[strings.ToLower(path[idx:])]; ok {
			return true
		}
	}
	return false
}

func isTrackableUserActivityPath(path string) bool {
	if shouldSkipUserActivityPath(path) {
		return false
	}
	for _, prefix := range userActivityAPITrackPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func extractTokenFromRequest(c echo.Context) string {
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	if auth != "" {
		auth = strings.TrimSpace(auth)
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			return strings.TrimSpace(auth[7:])
		}
		return auth
	}
	if cookie, err := c.Request().Cookie("dms-token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func parseUserUIDFromRequest(c echo.Context) (string, bool) {
	tokenStr := extractTokenFromRequest(c)
	if tokenStr == "" {
		return "", false
	}
	uid, err := jwt.ParseUidFromJwtTokenStr(tokenStr)
	if err != nil || uid == "" {
		return "", false
	}
	if uid == constant.UIDOfUserSys {
		return "", false
	}
	return uid, true
}

func normalizeRequestPath(path string) string {
	if path == "" {
		return path
	}
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i, segment := range segments {
		if segment == "" {
			continue
		}
		switch {
		case snowflakeSegmentPattern.MatchString(segment):
			segments[i] = ":uid"
		case numericSegmentPattern.MatchString(segment):
			segments[i] = ":id"
		default:
			segments[i] = segment
		}
	}
	return "/" + strings.Join(segments, "/")
}

func resolveModuleCode(normalizedPath string) string {
	normalizedPath = strings.ToLower(normalizedPath)
	if strings.Contains(normalizedPath, "/workflows") ||
		strings.Contains(normalizedPath, "/workflow_template") {
		return "WORKFLOW"
	}
	if strings.Contains(normalizedPath, "/sql_audit_records") ||
		strings.HasSuffix(normalizedPath, "/sql_audit") ||
		strings.Contains(normalizedPath, "/audit_files") ||
		strings.Contains(normalizedPath, "/sql_analysis") ||
		strings.Contains(normalizedPath, "/tasks/audits/") {
		return "SQL_AUDIT"
	}
	if strings.Contains(normalizedPath, "/audit_plans") ||
		strings.Contains(normalizedPath, "/instance_audit_plans") ||
		strings.Contains(normalizedPath, "/audit_plan_metas") ||
		strings.Contains(normalizedPath, "/audit_plan_types") {
		return "AUDIT_PLAN"
	}
	if strings.Contains(normalizedPath, "/sql_manages") {
		return "SQL_MANAGE"
	}
	if strings.Contains(normalizedPath, "/sql_optimization_records") {
		return "SQL_OPTIMIZE"
	}
	if strings.Contains(normalizedPath, "/sql_versions") ||
		strings.Contains(normalizedPath, "/pipelines") ||
		strings.Contains(normalizedPath, "/database_comparison") {
		return "SQL_VERSION"
	}
	if strings.Contains(normalizedPath, "/data_export_workflows") ||
		strings.Contains(normalizedPath, "/data_export_tasks") {
		return "DATA_EXPORT"
	}
	if strings.Contains(normalizedPath, "/cb_operation_logs") ||
		strings.HasPrefix(normalizedPath, "/sql_query/") ||
		strings.HasPrefix(normalizedPath, "/odc_query/") {
		return "WORKBENCH"
	}
	if strings.Contains(normalizedPath, "/db_services") ||
		strings.Contains(normalizedPath, "/instances/") ||
		strings.Contains(normalizedPath, "/environment_tags") ||
		strings.Contains(normalizedPath, "/db_service_sync_tasks") {
		return "DB_SERVICE"
	}
	if strings.Contains(normalizedPath, "/members") ||
		strings.Contains(normalizedPath, "/member_groups") ||
		strings.Contains(normalizedPath, "/business_tags") ||
		(strings.Contains(normalizedPath, "/projects/") && strings.Contains(normalizedPath, "/statistic")) {
		return "PROJECT"
	}
	if strings.Contains(normalizedPath, "/rule_templates") ||
		strings.Contains(normalizedPath, "/custom_rules") ||
		strings.Contains(normalizedPath, "/rule_knowledge") ||
		strings.Contains(normalizedPath, "/knowledge_bases") {
		return "RULE"
	}

	for _, rule := range defaultModulePrefixRules {
		for _, prefix := range rule.Prefixes {
			if strings.HasPrefix(normalizedPath, strings.ToLower(prefix)) {
				return rule.ModuleCode
			}
		}
	}
	return "OTHER"
}

func extractProjectUIDFromPath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(segments)-1; i++ {
		if segments[i] == "projects" && snowflakeSegmentPattern.MatchString(segments[i+1]) {
			return segments[i+1]
		}
	}
	return ""
}

func truncateUserAgent(ua string) string {
	if len(ua) <= maxUserAgentLen {
		return ua
	}
	return ua[:maxUserAgentLen]
}

func isTrackableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead:
		return true
	default:
		return false
	}
}
