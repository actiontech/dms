package constant

import (
	"testing"
)

func TestCheckDBTypeIfDataExportSupported_NewTypes(t *testing.T) {
	// 验证新增的数据源应在白名单中
	newTypes := map[string]bool{
		"TiDB":              true,
		"TDSQL For InnoDB":  true,
		"GoldenDB":          true,
		"TBase":             true,
		"GaussDB for MySQL": true, // GaussDB/openGauss: ParseDBType 的输入值是 "GaussDB for MySQL"
		"DB2":               true,
		"HANA":              true,
		"PolarDB For MySQL": true,
	}
	for dbType, expectedSupported := range newTypes {
		t.Run(dbType, func(t *testing.T) {
			got := CheckDBTypeIfDataExportSupported(dbType)
			if got != expectedSupported {
				t.Errorf("CheckDBTypeIfDataExportSupported(%q) = %v, want %v", dbType, got, expectedSupported)
			}
		})
	}
}

func TestCheckDBTypeIfDataExportSupported_ExistingTypes(t *testing.T) {
	// 回归测试: 验证已有的 7 种数据源仍在白名单中
	existingTypes := map[string]bool{
		"MySQL":               true,
		"PostgreSQL":          true,
		"Oracle":              true,
		"SQL Server":          true,
		"OceanBase For MySQL": true,
		"Hive":                true,
		"DM":                  true,
	}
	for dbType, expectedSupported := range existingTypes {
		t.Run(dbType, func(t *testing.T) {
			got := CheckDBTypeIfDataExportSupported(dbType)
			if got != expectedSupported {
				t.Errorf("CheckDBTypeIfDataExportSupported(%q) = %v, want %v", dbType, got, expectedSupported)
			}
		})
	}
}

func TestParseDBType_PolarDB(t *testing.T) {
	cases := map[string]struct {
		input       string
		wantDBType  DBType
		wantErr     bool
	}{
		"valid PolarDB For MySQL": {
			input:      "PolarDB For MySQL",
			wantDBType: DBTypePolarDBMySQL,
			wantErr:    false,
		},
		"invalid lowercase polardb": {
			input:      "polardb for mysql",
			wantDBType: "",
			wantErr:    true,
		},
		"invalid partial match": {
			input:      "PolarDB",
			wantDBType: "",
			wantErr:    true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseDBType(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseDBType(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
				return
			}
			if got != tc.wantDBType {
				t.Errorf("ParseDBType(%q) = %v, want %v", tc.input, got, tc.wantDBType)
			}
		})
	}
}

func TestCheckDBTypeIfDataExportSupported_UnsupportedTypes(t *testing.T) {
	// 验证未支持的类型返回 false
	unsupportedTypes := map[string]bool{
		"MongoDB":   false,
		"Redis":     false,
		"UnknownDB": false,
		"":          false,
	}
	for dbType, expectedSupported := range unsupportedTypes {
		t.Run(dbType, func(t *testing.T) {
			got := CheckDBTypeIfDataExportSupported(dbType)
			if got != expectedSupported {
				t.Errorf("CheckDBTypeIfDataExportSupported(%q) = %v, want %v", dbType, got, expectedSupported)
			}
		})
	}
}

func TestParseDBType(t *testing.T) {
	tests := map[string]struct {
		input       string
		expected    DBType
		expectError bool
	}{
		"MySQL":              {input: "MySQL", expected: DBTypeMySQL},
		"TDSQL For InnoDB":   {input: "TDSQL For InnoDB", expected: DBTypeTDSQLForInnoDB},
		"TiDB":               {input: "TiDB", expected: DBTypeTiDB},
		"PostgreSQL":         {input: "PostgreSQL", expected: DBTypePostgreSQL},
		"Oracle":             {input: "Oracle", expected: DBTypeOracle},
		"DB2":                {input: "DB2", expected: DBTypeDB2},
		"SQL Server":         {input: "SQL Server", expected: DBTypeSQLServer},
		"OceanBase For MySQL": {input: "OceanBase For MySQL", expected: DBTypeOceanBaseMySQL},
		"GoldenDB":           {input: "GoldenDB", expected: DBTypeGoldenDB},
		"TBase":              {input: "TBase", expected: DBTypeTBase},
		"Hive":               {input: "Hive", expected: DBTypeHive},
		"DM":                 {input: "DM", expected: DBTypeDM},
		"GaussDB for MySQL":  {input: "GaussDB for MySQL", expected: DBTypeGaussDB},
		"HANA":               {input: "HANA", expected: DBTypeHANA},
		"invalid type":       {input: "UnknownDB", expectError: true},
		"empty string":       {input: "", expectError: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseDBType(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("ParseDBType(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDBType(%q) unexpected error: %v", tc.input, err)
				return
			}
			if got != tc.expected {
				t.Errorf("ParseDBType(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// TestParseDBType_GaussDB_Aliases 覆盖 Issue #2868 / compat-RISK-1 的 GaussDB 系列别名:
//   - "GaussDB" / "openGauss" / "GaussDB / openGauss" 是本次新增
//   - "GaussDB for MySQL" 是 ADR-004 历史命名 (回归保留)
//
// 四种字符串均严格字面量匹配 (无 ToLower/ToUpper 归一化), 因此提供小写/拼写差异作为反向 case。
func TestParseDBType_GaussDB_Aliases(t *testing.T) {
	positive := map[string]string{
		"alias_GaussDB":               "GaussDB",
		"alias_openGauss":             "openGauss",
		"alias_GaussDB_slash_openGauss": "GaussDB / openGauss",
		"alias_GaussDB_for_MySQL":     "GaussDB for MySQL",
	}
	for name, input := range positive {
		t.Run(name, func(t *testing.T) {
			got, err := ParseDBType(input)
			if err != nil {
				t.Fatalf("ParseDBType(%q) unexpected error: %v", input, err)
			}
			if got != DBTypeGaussDB {
				t.Errorf("ParseDBType(%q) = %q, want %q", input, got, DBTypeGaussDB)
			}
		})
	}

	// 反向 case: ParseDBType 字面量敏感, 大小写/空格差异均不映射到 DBTypeGaussDB
	negative := map[string]string{
		"lowercase_gaussdb":              "gaussdb",
		"uppercase_GAUSSDB":              "GAUSSDB",
		"no_space_around_slash":          "GaussDB/openGauss",
		"upper_GAUSSDB_FOR_MYSQL":        "GAUSSDB FOR MYSQL",
		"random_string_not_in_whitelist": "NotADBType",
	}
	for name, input := range negative {
		t.Run(name, func(t *testing.T) {
			got, err := ParseDBType(input)
			if err == nil {
				t.Errorf("ParseDBType(%q) expected error (not GaussDB), got %q", input, got)
			}
		})
	}
}
