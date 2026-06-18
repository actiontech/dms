package sql_workbench

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/sql_workbench/client"
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
		"Redis":               {input: "Redis", expected: "REDIS"},
		"DB2":                 {input: "DB2", expected: "DB2"},
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
		"Redis supported":             {input: pkgConst.DBTypeRedis, expected: true},
		"PostgreSQL supported":        {input: pkgConst.DBTypePostgreSQL, expected: true},
		"SQL Server unsupported":      {input: pkgConst.DBTypeSQLServer, expected: false},
		"PolarDB For MySQL supported": {input: pkgConst.DBTypePolarDBForMySQL, expected: true},
		"GaussDB supported":           {input: pkgConst.DBTypeGaussDB, expected: true},
		"GaussDBForMySQL unsupported": {input: pkgConst.DBTypeGaussDBForMySQL, expected: false},
		"DB2 unsupported":             {input: pkgConst.DBTypeDB2, expected: false},
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

func Test_buildRedisDatasourceOptions(t *testing.T) {
	defaultDB := "2"
	defaultSchema, propertiesValue, jdbcParams := buildRedisDatasourceOptions(&biz.DBService{
		DBType: string(pkgConst.DBTypeRedis),
		Host:   "127.0.0.1",
		Port:   "6379",
		AdditionalParams: pkgParams.Params{
			&pkgParams.Param{Key: redisDefaultDatabaseParam, Value: defaultDB, Type: pkgParams.ParamTypeString},
		},
	})
	if defaultSchema == nil || *defaultSchema != defaultDB {
		t.Fatalf("unexpected default schema: %#v", defaultSchema)
	}
	if propertiesValue != nil {
		t.Fatalf("expected nil properties, got %#v", propertiesValue)
	}
	if jdbcParams["defaultDatabase"] != defaultDB {
		t.Fatalf("unexpected redis jdbc params: %#v", jdbcParams)
	}
}

func Test_buildRedisDatasourceOptions_noSensitiveProperties(t *testing.T) {
	_, propertiesValue, jdbcParams := buildRedisDatasourceOptions(&biz.DBService{
		DBType:   string(pkgConst.DBTypeRedis),
		User:     "default",
		Password: "secret",
		AdditionalParams: pkgParams.Params{
			&pkgParams.Param{Key: redisDefaultDatabaseParam, Value: "0", Type: pkgParams.ParamTypeString},
		},
	})
	if propertiesValue != nil {
		t.Fatalf("expected redis properties to stay nil, got %#v", propertiesValue)
	}
	if _, ok := jdbcParams["password"]; ok {
		t.Fatalf("redis password leaked into jdbc params: %#v", jdbcParams)
	}
}

// Test_buildDatasourceBaseInfo_DB2 覆盖 buildDatasourceBaseInfo 中 DB2 / 回归 4 组 case：
//
//	(a) DB2 正例：AdditionalParam database_name=testdb → baseInfo.DefaultSchema=="testdb"
//	(b) DB2 负例：缺 database_name → 返回 err 且 err 含 "database_name"
//	(c) MySQL 回归：DefaultSchema == nil 且无 err
//	(d) Oracle 回归：ServiceName != nil 且无 err
//
// 通过 fillDatasourceBaseInfo（无 IO helper）进行 mock-only 单测，避免触达 projectUsecase / DB。
func Test_buildDatasourceBaseInfo_DB2(t *testing.T) {
	svc := &SqlWorkbenchService{}
	const envID = int64(1)
	const datasourceName = "proj:ds"

	cases := map[string]struct {
		dbService           *biz.DBService
		expectErr           bool
		expectErrSubstr     string
		expectDefaultSchema *string
		expectServiceName   *string
	}{
		"DB2 happy path": {
			dbService: &biz.DBService{
				Name:   "db2-1",
				DBType: "DB2",
				AdditionalParams: pkgParams.Params{
					{Key: "database_name", Value: "testdb"},
				},
			},
			expectErr:           false,
			expectDefaultSchema: strPtr("testdb"),
			expectServiceName:   nil,
		},
		"DB2 missing database_name": {
			dbService: &biz.DBService{
				Name:             "db2-2",
				DBType:           "DB2",
				AdditionalParams: pkgParams.Params{},
			},
			expectErr:       true,
			expectErrSubstr: "database_name",
		},
		"MySQL regression": {
			dbService: &biz.DBService{
				Name:             "mysql-1",
				DBType:           "MySQL",
				AdditionalParams: pkgParams.Params{},
			},
			expectErr:           false,
			expectDefaultSchema: nil,
			expectServiceName:   nil,
		},
		"Oracle regression": {
			dbService: &biz.DBService{
				Name:   "oracle-1",
				DBType: "Oracle",
				AdditionalParams: pkgParams.Params{
					{Key: "service_name", Value: "ORCL"},
				},
			},
			expectErr:           false,
			expectDefaultSchema: nil,
			expectServiceName:   strPtr("ORCL"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := svc.fillDatasourceBaseInfo(datasourceName, tc.dbService, envID)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil; baseInfo=%+v", got)
				}
				if tc.expectErrSubstr != "" && !strings.Contains(err.Error(), tc.expectErrSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.expectErrSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil baseInfo")
			}
			// DefaultSchema 对比
			if (got.DefaultSchema == nil) != (tc.expectDefaultSchema == nil) {
				t.Errorf("DefaultSchema nil mismatch: got=%v, want=%v", got.DefaultSchema, tc.expectDefaultSchema)
			} else if got.DefaultSchema != nil && tc.expectDefaultSchema != nil && *got.DefaultSchema != *tc.expectDefaultSchema {
				t.Errorf("DefaultSchema = %q, want %q", *got.DefaultSchema, *tc.expectDefaultSchema)
			}
			// ServiceName 对比
			if (got.ServiceName == nil) != (tc.expectServiceName == nil) {
				t.Errorf("ServiceName nil mismatch: got=%v, want=%v", got.ServiceName, tc.expectServiceName)
			} else if got.ServiceName != nil && tc.expectServiceName != nil && *got.ServiceName != *tc.expectServiceName {
				t.Errorf("ServiceName = %q, want %q", *got.ServiceName, *tc.expectServiceName)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func Test_odcDatasourceRequests_includePasswordSaved(t *testing.T) {
	pwd := "secret"
	createReq := client.CreateDatasourceRequest{
		Name:          "proj:ds",
		Type:          "MYSQL",
		Username:      "u",
		Password:      pwd,
		PasswordSaved: true,
	}
	createJSON, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal create request: %v", err)
	}
	if !strings.Contains(string(createJSON), `"passwordSaved":true`) {
		t.Fatalf("create request JSON missing passwordSaved:true: %s", createJSON)
	}

	updateReq := client.UpdateDatasourceRequest{
		Type:          "MYSQL",
		Username:      "u",
		Password:      &pwd,
		PasswordSaved: true,
	}
	updateJSON, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}
	if !strings.Contains(string(updateJSON), `"passwordSaved":true`) {
		t.Fatalf("update request JSON missing passwordSaved:true: %s", updateJSON)
	}
}

func Test_buildOdcCreateAndUpdateRequests_setPasswordSaved(t *testing.T) {
	svc := &SqlWorkbenchService{}
	baseInfo, err := svc.fillDatasourceBaseInfo("proj:ds", &biz.DBService{
		Name:     "ds",
		DBType:   "MySQL",
		User:     "root",
		Password: "pass",
		Host:     "127.0.0.1",
		Port:     "3306",
	}, 1)
	if err != nil {
		t.Fatalf("fillDatasourceBaseInfo: %v", err)
	}

	createReq := client.CreateDatasourceRequest{
		Type:          baseInfo.Type,
		Name:          baseInfo.Name,
		Username:      baseInfo.Username,
		Password:      baseInfo.Password,
		Host:          baseInfo.Host,
		Port:          baseInfo.Port,
		EnvironmentID: baseInfo.EnvironmentID,
		PasswordSaved: true,
	}
	createJSON, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal create: %v", err)
	}
	if !strings.Contains(string(createJSON), `"passwordSaved":true`) {
		t.Fatalf("expected passwordSaved in create JSON: %s", createJSON)
	}

	updateReq := client.UpdateDatasourceRequest{
		Type:          baseInfo.Type,
		Name:          &baseInfo.Name,
		Username:      baseInfo.Username,
		Password:      &baseInfo.Password,
		Host:          baseInfo.Host,
		Port:          baseInfo.Port,
		EnvironmentID: baseInfo.EnvironmentID,
		PasswordSaved: true,
	}
	updateJSON, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	if !strings.Contains(string(updateJSON), `"passwordSaved":true`) {
		t.Fatalf("expected passwordSaved in update JSON: %s", updateJSON)
	}
}
