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
