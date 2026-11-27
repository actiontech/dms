package service

import (
	"encoding/json"
	"testing"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/stretchr/testify/assert"
)

func TestMapPrivilegeCheckResults_FieldMapping(t *testing.T) {
	// AC-001: privilege Reply carries module statuses + precheck; missing priv ≠ connect failure shape.
	in := []*biz.CheckDBServicePrivileges{
		{
			DBType:       "MySQL",
			CheckSupport: "supported",
			ConnectivityPrecheck: biz.ConnectivityPrecheck{
				OK:           true,
				ErrorMessage: "",
			},
			Modules: []biz.PrivilegeModuleResult{
				{
					Module:     "sql_workbench_query",
					ModuleName: "SQL 工作台查询",
					Status:     "available",
				},
				{
					Module:     "account_privilege_mgmt",
					ModuleName: "账号与权限管理",
					Status:     "unavailable",
					MissingPrivileges: []biz.MissingPrivilegeItem{
						{Privilege: "CREATE USER", ObjectScope: "*.*", Note: "缺少管理权限"},
					},
					Message: "账号管理不可用",
				},
			},
			SummaryMessage: "部分模块可用，不影响创建/保存",
		},
		{
			DBType:       "PostgreSQL",
			CheckSupport: "unsupported_auto_check",
			ConnectivityPrecheck: biz.ConnectivityPrecheck{
				OK: true,
			},
			Modules: []biz.PrivilegeModuleResult{
				{Module: "sql_workbench_query", ModuleName: "SQL 工作台查询", Status: "unsupported_auto_check"},
			},
			SummaryMessage: "暂不支持自动检查",
		},
	}

	out := mapPrivilegeCheckResults(in)
	if !assert.Len(t, out, 2) {
		return
	}

	assert.Equal(t, "MySQL", out[0].DBType)
	assert.Equal(t, "supported", out[0].CheckSupport)
	assert.True(t, out[0].ConnectivityPrecheck.OK)
	assert.Empty(t, out[0].ConnectivityPrecheck.ErrorMessage)
	if assert.Len(t, out[0].Modules, 2) {
		assert.Equal(t, "available", out[0].Modules[0].Status)
		assert.Equal(t, "unavailable", out[0].Modules[1].Status)
		if assert.Len(t, out[0].Modules[1].MissingPrivileges, 1) {
			assert.Equal(t, "CREATE USER", out[0].Modules[1].MissingPrivileges[0].Privilege)
			assert.Equal(t, "*.*", out[0].Modules[1].MissingPrivileges[0].ObjectScope)
			assert.Equal(t, "缺少管理权限", out[0].Modules[1].MissingPrivileges[0].Note)
		}
		assert.Equal(t, "账号管理不可用", out[0].Modules[1].Message)
	}
	assert.Equal(t, "部分模块可用，不影响创建/保存", out[0].SummaryMessage)

	assert.Equal(t, "PostgreSQL", out[1].DBType)
	assert.Equal(t, "unsupported_auto_check", out[1].CheckSupport)
	assert.True(t, out[1].ConnectivityPrecheck.OK)

	reply := dmsV1.CheckDBServicesPrivilegesReply{Data: out}
	raw, err := json.Marshal(reply)
	assert.NoError(t, err)
	body := string(raw)
	assert.Contains(t, body, `"check_support"`)
	assert.Contains(t, body, `"connectivity_precheck"`)
	assert.Contains(t, body, `"modules"`)
	assert.Contains(t, body, `"missing_privileges"`)
	assert.NotContains(t, body, `"is_connectable"`)
	assert.NotContains(t, body, `"CheckDBServicesPrivileges"`)
}
