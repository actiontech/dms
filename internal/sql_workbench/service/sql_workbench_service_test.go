package sql_workbench

import (
	"testing"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func Test_convertDBType(t *testing.T) {
	svc := &SqlWorkbenchService{}
	cases := map[string]struct {
		input    string
		expected string
	}{
		"DM":                 {input: "达梦(DM)", expected: "DM"},
		"MySQL":              {input: "MySQL", expected: "MYSQL"},
		"PostgreSQL":         {input: "PostgreSQL", expected: "POSTGRESQL"},
		"Oracle":             {input: "Oracle", expected: "ORACLE"},
		"SQL Server":         {input: "SQL Server", expected: "SQL_SERVER"},
		"OB Oracle":          {input: "OceanBase For Oracle", expected: "OB_ORACLE"},
		"OB MySQL":           {input: "OceanBase For MySQL", expected: "OB_MYSQL"},
		"Unknown passthrough": {input: "UnknownDB", expected: "UnknownDB"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := svc.convertDBType(tc.input)
			if got != tc.expected {
				t.Errorf("convertDBType(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func Test_SupportDBType(t *testing.T) {
	svc := &SqlWorkbenchService{}
	cases := map[string]struct {
		input    pkgConst.DBType
		expected bool
	}{
		"DM supported":           {input: pkgConst.DBTypeDM, expected: true},
		"MySQL supported":        {input: pkgConst.DBTypeMySQL, expected: true},
		"Oracle supported":       {input: pkgConst.DBTypeOracle, expected: true},
		"OB MySQL supported":     {input: pkgConst.DBTypeOceanBaseMySQL, expected: true},
		"PostgreSQL unsupported": {input: pkgConst.DBTypePostgreSQL, expected: false},
		"SQL Server unsupported": {input: pkgConst.DBTypeSQLServer, expected: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := svc.SupportDBType(tc.input)
			if got != tc.expected {
				t.Errorf("SupportDBType(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
