package biz

import (
	"context"
	"encoding/json"
	"testing"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/stretchr/testify/assert"
)

func TestIsConnectivityPluginUnavailable(t *testing.T) {
	// AC-007: plugin missing must be recognized so aggregate does not treat it as dial/auth failure.
	cases := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "plugin_not_found_lower", msg: "open plugin: plugin not found", want: true},
		{name: "plugin_not_found_mixed_case", msg: "Open Plugin: Plugin Not Found", want: true},
		{name: "open_plugin_prefix_only", msg: "open plugin: /path/to/ms.so: no such file", want: true},
		{name: "access_denied_not_plugin", msg: "Error 1045: Access denied for user", want: false},
		{name: "login_failed_mssql", msg: "Login failed for user 'env_mssql_ro'", want: false},
		{name: "empty", msg: "", want: false},
		{name: "unrelated_error", msg: "connection refused", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isConnectivityPluginUnavailable(tc.msg))
		})
	}
}

func TestIsModulePrivilegeAutoCheckSupported(t *testing.T) {
	cases := []struct {
		dbType string
		want   bool
	}{
		{"MySQL", true},
		{"OceanBase For MySQL", true},
		{"SQL Server", true},
		{"PostgreSQL", false},
		{"Oracle", false},
		{"DB2", false},
		{"", false},
		{"  MySQL  ", true},
	}
	for _, tc := range cases {
		t.Run(tc.dbType, func(t *testing.T) {
			assert.Equal(t, tc.want, isModulePrivilegeAutoCheckSupported(tc.dbType))
		})
	}
}

func TestUnsupportedPrivilegeModules_Shape(t *testing.T) {
	mods := unsupportedPrivilegeModules()
	if assert.Len(t, mods, 6) {
		wantOrder := []string{
			"sql_workbench_query",
			"sql_audit_analysis",
			"data_export",
			"sql_deploy",
			"smart_scan",
			"account_privilege_mgmt",
		}
		for i, id := range wantOrder {
			assert.Equal(t, id, mods[i].Module)
			assert.Equal(t, "unsupported_auto_check", mods[i].Status)
			assert.NotEmpty(t, mods[i].ModuleName)
			assert.Empty(t, mods[i].MissingPrivileges)
		}
	}
}

func TestCheckOneDBServicePrivileges_AllowlistUnsupportedShortPath(t *testing.T) {
	// AC-001 boundary: types outside the auto-check allowlist must not become connectivity failures.
	d := &DBServiceUsecase{}
	cases := []string{"PostgreSQL", "Oracle", "DB2", "TiDB", "UnknownType"}
	for _, dbType := range cases {
		t.Run(dbType, func(t *testing.T) {
			got, err := d.checkOneDBServicePrivileges(context.Background(), dmsCommonV1.CheckDbConnectable{
				DBType: dbType,
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "ro",
			})
			assert.NoError(t, err)
			if !assert.NotNil(t, got) {
				return
			}
			assert.Equal(t, dbType, got.DBType)
			assert.Equal(t, "unsupported_auto_check", got.CheckSupport)
			assert.True(t, got.ConnectivityPrecheck.OK, "缺权/暂不支持不得写成连通失败")
			assert.Empty(t, got.ConnectivityPrecheck.ErrorMessage)
			assert.Equal(t, "暂不支持自动检查", got.SummaryMessage)
			assert.Len(t, got.Modules, 6)
			for _, m := range got.Modules {
				assert.Equal(t, "unsupported_auto_check", m.Status)
			}
			raw, err := json.Marshal(got)
			assert.NoError(t, err)
			assert.NotContains(t, string(raw), `"is_connectable"`)
			assert.NotContains(t, string(raw), "CREATE USER")
		})
	}
}

func TestCheckDBServiceHasEnoughPrivileges_UnsupportedBatch(t *testing.T) {
	d := &DBServiceUsecase{}
	got, err := d.CheckDBServiceHasEnoughPrivileges(context.Background(), []dmsCommonV1.CheckDbConnectable{
		{DBType: "PostgreSQL", Host: "h", Port: "1", User: "u"},
		{DBType: "Oracle", Host: "h", Port: "1", User: "u"},
	})
	assert.NoError(t, err)
	if !assert.Len(t, got, 2) {
		return
	}
	assert.Equal(t, "PostgreSQL", got[0].DBType)
	assert.Equal(t, "Oracle", got[1].DBType)
	for _, item := range got {
		assert.Equal(t, "unsupported_auto_check", item.CheckSupport)
		assert.True(t, item.ConnectivityPrecheck.OK)
	}
}
