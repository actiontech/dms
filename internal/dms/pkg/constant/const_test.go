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
		// Issue #2868: GaussDB (PostgreSQL 协议) 与 GaussDB for MySQL (MySQL 协议) 是两个独立产品
		"GaussDB":           true, // PostgreSQL 协议 GaussDB / openGauss, 走 opengauss-connector-go-pq 驱动
		"GaussDB for MySQL": true, // 华为云 GaussDB(for MySQL), 走 MySQL 驱动
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
			wantDBType: DBTypePolarDBForMySQL,
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
		// Issue #2868: 拆分 GaussDB / GaussDB for MySQL 为两个独立产品
		"GaussDB":            {input: "GaussDB", expected: DBTypeGaussDB},
		"GaussDB for MySQL":  {input: "GaussDB for MySQL", expected: DBTypeGaussDBForMySQL},
		"HANA":               {input: "HANA", expected: DBTypeHANA},
		// PolarDB-MySQL 新增 (Issue #826)
		"PolarDB For MySQL":  {input: "PolarDB For MySQL", expected: DBTypePolarDBForMySQL},
		// "PolarDB" 单独不应匹配
		"PolarDB only":       {input: "PolarDB", expectError: true},
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
