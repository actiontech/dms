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
		"GaussDB":             {input: "GaussDB", expected: "GAUSSDB"},
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
		"MongoDB unsupported":         {input: pkgConst.DBTypeMongoDB, expected: false},
		"PostgreSQL supported":        {input: pkgConst.DBTypePostgreSQL, expected: true},
		"SQL Server unsupported":      {input: pkgConst.DBTypeSQLServer, expected: false},
		"PolarDB For MySQL supported": {input: pkgConst.DBTypePolarDBForMySQL, expected: true},
		"GaussDB supported":           {input: pkgConst.DBTypeGaussDB, expected: true},
		"GaussDBForMySQL unsupported": {input: pkgConst.DBTypeGaussDBForMySQL, expected: false},
		"empty string unsupported":    {input: pkgConst.DBType(""), expected: false},
		"unknown type unsupported":    {input: pkgConst.DBType("UnknownDBType"), expected: false},
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

func Test_SupportDBType_GaussDB_PG_family_consistency(t *testing.T) {
	svc := &SqlWorkbenchService{}
	// CR-13: design §1.2 decision-3 locks PG family (PostgreSQL + GaussDB)
	// must be whitelisted together; SQL workbench routing assumes the pair.
	if got := svc.SupportDBType(pkgConst.DBTypePostgreSQL); !got {
		t.Errorf("PostgreSQL must be supported (CR-13 / EARS-1.2)")
	}
	if got := svc.SupportDBType(pkgConst.DBTypeGaussDB); !got {
		t.Errorf("GaussDB must be supported (EARS-1.2 / decision-3)")
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
	if propertiesValue != nil {
		t.Fatalf("expected nil properties, got %#v", propertiesValue)
	}
	if jdbcParams["authSource"] != "admin" {
		t.Fatalf("unexpected authSource: %#v", jdbcParams["authSource"])
	}
	if jdbcParams["authMechanism"] != "SCRAM-SHA-256" {
		t.Fatalf("unexpected authMechanism: %#v", jdbcParams["authMechanism"])
	}
	if jdbcParams["replicaSet"] != "rs0" {
		t.Fatalf("unexpected replicaSet: %#v", jdbcParams["replicaSet"])
	}
	if jdbcParams["tls"] != "true" {
		t.Fatalf("unexpected tls: %#v", jdbcParams["tls"])
	}
	if jdbcParams["directConnection"] != true || jdbcParams["tlsInsecure"] != true {
		t.Fatalf("unexpected jdbc params: %#v", jdbcParams)
	}
}

func Test_buildMongoDatasourceOptions_tlsOnly(t *testing.T) {
	_, propertiesValue, jdbcParams := buildMongoDatasourceOptions(&biz.DBService{
		DBType: string(pkgConst.DBTypeMongoDB),
		AdditionalParams: pkgParams.Params{
			&pkgParams.Param{Key: mongoTLSEnabledParam, Value: "true", Type: pkgParams.ParamTypeBool},
		},
	})
	if propertiesValue != nil {
		t.Fatalf("expected nil properties, got %#v", propertiesValue)
	}
	if jdbcParams["tls"] != "true" {
		t.Fatalf("expected tls in jdbcUrlParameters when only tls is configured, got %#v", jdbcParams)
	}
}

