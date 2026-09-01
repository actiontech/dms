package biz

import (
	"regexp"
	"strings"
)

var (
	privilegeDeniedVendorCodes = map[string]struct{}{
		"1044": {},
		"1142": {},
		"1227": {},
	}

	privilegeDeniedMessagePatterns = []string{
		"access denied for user",
		"command denied to user",
	}

	sqlActionPattern = regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|CREATE|DROP|ALTER|TRUNCATE|GRANT|REVOKE)\b`)
	tableFromPattern = regexp.MustCompile(`(?i)(?:FROM|INTO|UPDATE|TABLE)\s+[` + "`" + `"]?([a-zA-Z0-9_]+)[` + "`" + `"]?(?:\.[` + "`" + `"]?([a-zA-Z0-9_]+)[` + "`" + `"]?)?`)
)

// PrivilegeDeniedContext 缺权结构化上下文
type PrivilegeDeniedContext struct {
	PrivilegeDenied  bool              `json:"privilege_denied"`
	ProjectUID       string            `json:"project_uid"`
	DBServiceUID     string            `json:"db_service_uid"`
	DBServiceName    string            `json:"db_service_name,omitempty"`
	DBAccountUID     string            `json:"db_account_uid"`
	DBAccountName    string            `json:"db_account_name"`
	RawSQL           string            `json:"raw_sql"`
	ErrorMessage     string            `json:"error_message"`
	VendorCode       *string           `json:"vendor_code"`
	SQLState         *string           `json:"sql_state"`
	RequestedObjects []PrivilegeObject `json:"requested_objects,omitempty"`
	RequestedActions []string          `json:"requested_actions,omitempty"`
}

// IsPrivilegeDeniedError 判断是否为库账号缺权类错误
func IsPrivilegeDeniedError(errorMessage string, vendorCode string) bool {
	code := strings.TrimSpace(vendorCode)
	if code != "" {
		if _, ok := privilegeDeniedVendorCodes[code]; ok {
			return true
		}
	}
	msg := strings.ToLower(strings.TrimSpace(errorMessage))
	if msg == "" {
		return false
	}
	for _, pattern := range privilegeDeniedMessagePatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// ParseSQLObjectsActions 尽力从 SQL 解析对象与动作
func ParseSQLObjectsActions(sql string) (objects []PrivilegeObject, actions []string) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return nil, nil
	}

	actionSet := make(map[string]struct{})
	for _, match := range sqlActionPattern.FindAllString(trimmed, -1) {
		action := strings.ToUpper(match)
		actionSet[action] = struct{}{}
	}
	for action := range actionSet {
		actions = append(actions, action)
	}

	objectSet := make(map[string]PrivilegeObject)
	for _, match := range tableFromPattern.FindAllStringSubmatch(trimmed, -1) {
		if len(match) < 2 {
			continue
		}
		obj := PrivilegeObject{ObjectType: "table"}
		if match[2] != "" {
			obj.Schema = match[1]
			obj.ObjectName = match[2]
		} else {
			obj.ObjectName = match[1]
		}
		key := obj.Schema + "." + obj.ObjectName
		objectSet[key] = obj
	}
	for _, obj := range objectSet {
		objects = append(objects, obj)
	}
	return objects, actions
}
