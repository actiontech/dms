package sql_workbench

import (
	"testing"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgParams "github.com/actiontech/dms/pkg/params"
)

func Test_convertDBType(t *testing.T) {
	svc := &SqlWorkbenchService{}
	cases := map[string]struct {
		input    string
		expected string
	}{
		"DM":                  {input: "达梦(DM)", expected: "DM"},
		"MySQL":               {input: "MySQL", expected: "MYSQL"},
		"PostgreSQL":          {input: "PostgreSQL", expected: "POSTGRESQL"},
		"Oracle":              {input: "Oracle", expected: "ORACLE"},
		"SQL Server":          {input: "SQL Server", expected: "SQL_SERVER"},
		"OB Oracle":           {input: "OceanBase For Oracle", expected: "OB_ORACLE"},
		"OB MySQL":            {input: "OceanBase For MySQL", expected: "OB_MYSQL"},
		"TiDB":                {input: "TiDB", expected: "TIDB"},
		"TDSQL For InnoDB":    {input: "TDSQL For InnoDB", expected: "MYSQL"},
		"GoldenDB":            {input: "GoldenDB", expected: "MYSQL"},
		"PolarDB For MySQL":   {input: "PolarDB For MySQL", expected: "MYSQL"},
		"MongoDB":             {input: "MongoDB", expected: "MONGODB"},
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
		"DM supported":                {input: pkgConst.DBTypeDM, expected: true},
		"MySQL supported":             {input: pkgConst.DBTypeMySQL, expected: true},
		"Oracle supported":            {input: pkgConst.DBTypeOracle, expected: true},
		"OB MySQL supported":          {input: pkgConst.DBTypeOceanBaseMySQL, expected: true},
		"TiDB supported":              {input: pkgConst.DBTypeTiDB, expected: true},
		"TDSQL supported":             {input: pkgConst.DBTypeTDSQLForInnoDB, expected: true},
		"GoldenDB supported":          {input: pkgConst.DBTypeGoldenDB, expected: true},
		"MongoDB supported":           {input: pkgConst.DBTypeMongoDB, expected: true},
		"PostgreSQL unsupported":      {input: pkgConst.DBTypePostgreSQL, expected: false},
		"SQL Server unsupported":      {input: pkgConst.DBTypeSQLServer, expected: false},
		"PolarDB For MySQL supported": {input: pkgConst.DBTypePolarDBForMySQL, expected: true},
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

func Test_buildMongoDatasourceOptions(t *testing.T) {
	defaultDB := "appdb"
	defaultSchema, propertiesValue, jdbcParams := buildMongoDatasourceOptions(&biz.DBService{
		DBType: string(pkgConst.DBTypeMongoDB),
		Host:   "127.0.0.1",
		Port:   "27017",
		AdditionalParams: pkgParams.Params{
			&pkgParams.Param{Key: mongoDefaultDatabaseParam, Value: defaultDB, Type: pkgParams.ParamTypeString},
			&pkgParams.Param{Key: mongoAuthDatabaseParam, Value: "admin", Type: pkgParams.ParamTypeString},
			&pkgParams.Param{Key: mongoAuthMechanismParam, Value: "SCRAM-SHA-256", Type: pkgParams.ParamTypeString},
			&pkgParams.Param{Key: mongoReplicaSetParam, Value: "rs0", Type: pkgParams.ParamTypeString},
			&pkgParams.Param{Key: mongoTLSEnabledParam, Value: "true", Type: pkgParams.ParamTypeBool},
			&pkgParams.Param{Key: mongoDirectConnectionParam, Value: "true", Type: pkgParams.ParamTypeBool},
			&pkgParams.Param{Key: mongoTLSSkipVerifyParam, Value: "true", Type: pkgParams.ParamTypeBool},
		},
	})
	if defaultSchema == nil || *defaultSchema != defaultDB {
		t.Fatalf("unexpected default schema: %#v", defaultSchema)
	}
	properties, ok := propertiesValue.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected properties type: %T", propertiesValue)
	}
	if properties["authenticationDatabase"] != "admin" {
		t.Fatalf("unexpected auth db: %#v", properties)
	}
	if properties["tlsEnabled"] != true {
		t.Fatalf("unexpected tlsEnabled: %#v", properties)
	}
	if jdbcParams["directConnection"] != true || jdbcParams["tlsInsecure"] != true {
		t.Fatalf("unexpected jdbc params: %#v", jdbcParams)
	}
}

