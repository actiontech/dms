// Package catalogschema 提供 SQL Server「合并层」catalog/schema 名称的解析、拼接与
// 过滤工具。SQL Server 的库/模式为两级结构（database.schema），DMS 在选库入口将其
// 合并为单层名称（整库 "db" 或 "db.schema"）呈现；本包统一维护这一层名称语义，供
// dms-ee 结构采集/脱敏与 provision 建账户授权两侧共享，避免两端各自实现导致行为漂移。
package catalogschema

import "strings"

// dbTypeSQLServer 与 dms internal 常量 DBTypeSQLServer、provision driver.DBTypeSQLServer
// 取值保持一致；此处以字符串常量内联，避免公共包反向依赖各仓库的 internal 常量。
const dbTypeSQLServer = "SQL Server"

// defaultSchemaName 是 SQL Server 库下的默认 schema。
const defaultSchemaName = "dbo"

// CatalogSchemaSplit 表示一个合并层名称拆分后的结果。
// WholeDB 为 true 表示该名称指向整库（无显式 schema）。
type CatalogSchemaSplit struct {
	Catalog string
	Schema  string
	WholeDB bool
}

// IsSchemaLevelDBType 标识需要 database+schema 合并层的数据库类型。
func IsSchemaLevelDBType(dbType string) bool {
	return strings.EqualFold(dbType, dbTypeSQLServer)
}

// BuildFullCatalogSchemaName 合并 catalog 与 schema 为 "catalog.schema"；schema 为空时默认 dbo。
func BuildFullCatalogSchemaName(catalog, schemaName string) string {
	if catalog == "" {
		return schemaName
	}
	if schemaName == "" {
		schemaName = defaultSchemaName
	}
	return catalog + "." + schemaName
}

// SplitCatalogSchema 将合并名拆回 catalog/schema；无点表示整库作用域。
//
// 合并名由 BuildFullCatalogSchemaName 以 "catalog.schema" 生成。因数据库名与
// schema 名理论上都可能含 "."，无法从字符串完美还原；这里以最后一个 "." 作为
// 分隔（LastIndex），即假定 schema 段不含 "."（SQL Server 默认 dbo 等均满足），
// 从而正确兼容数据库名本身含 "." 的情形（如 "my.db.dbo" → catalog="my.db"）。
func SplitCatalogSchema(dbType, name string) CatalogSchemaSplit {
	if !IsSchemaLevelDBType(dbType) || name == "" {
		return CatalogSchemaSplit{Schema: name}
	}
	if idx := strings.LastIndex(name, "."); idx > 0 {
		return CatalogSchemaSplit{
			Catalog: name[:idx],
			Schema:  name[idx+1:],
		}
	}
	return CatalogSchemaSplit{
		Catalog: name,
		WholeDB: true,
	}
}

// SchemaNameMatchesFilter 判断 schema 名是否命中任务过滤（支持整库 db 匹配其下 db.schema）。
func SchemaNameMatchesFilter(dbType string, filter map[string]struct{}, schemaName string) bool {
	if len(filter) == 0 {
		return true
	}
	if _, ok := filter[schemaName]; ok {
		return true
	}
	if !IsSchemaLevelDBType(dbType) {
		return false
	}
	for selected := range filter {
		selectedSplit := SplitCatalogSchema(dbType, selected)
		if !selectedSplit.WholeDB {
			continue
		}
		nameSplit := SplitCatalogSchema(dbType, schemaName)
		if nameSplit.Catalog == selectedSplit.Catalog {
			return true
		}
	}
	return false
}

// ConnectionCatalogForMergedName 返回 JDBC/连接串使用的 catalog（database）名。
func ConnectionCatalogForMergedName(dbType, mergedName string) string {
	split := SplitCatalogSchema(dbType, mergedName)
	if split.Catalog != "" {
		return split.Catalog
	}
	return ""
}
