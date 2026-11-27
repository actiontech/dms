package catalogschema

import "testing"

const testDBType = "SQL Server"

func TestIsSchemaLevelDBType(t *testing.T) {
	if !IsSchemaLevelDBType("SQL Server") {
		t.Fatalf("SQL Server should be schema-level")
	}
	if !IsSchemaLevelDBType("sql server") {
		t.Fatalf("match should be case-insensitive")
	}
	if IsSchemaLevelDBType("MySQL") {
		t.Fatalf("MySQL should not be schema-level")
	}
}

func TestBuildFullCatalogSchemaName(t *testing.T) {
	if got := BuildFullCatalogSchemaName("TestDB", "dbo"); got != "TestDB.dbo" {
		t.Fatalf("got %q, want TestDB.dbo", got)
	}
	if got := BuildFullCatalogSchemaName("TestDB", ""); got != "TestDB.dbo" {
		t.Fatalf("empty schema should default to dbo, got %q", got)
	}
	if got := BuildFullCatalogSchemaName("", "public"); got != "public" {
		t.Fatalf("empty catalog should return schema as-is, got %q", got)
	}
}

func TestSchemaNameMatchesFilter(t *testing.T) {
	filter := map[string]struct{}{"TestDB": {}}
	if !SchemaNameMatchesFilter(testDBType, filter, "TestDB.dbo") {
		t.Fatalf("whole-db filter should match its schema")
	}
	if SchemaNameMatchesFilter(testDBType, filter, "OtherDB.dbo") {
		t.Fatalf("whole-db filter should not match a different db")
	}
	if !SchemaNameMatchesFilter(testDBType, map[string]struct{}{}, "AnyDB.dbo") {
		t.Fatalf("empty filter should match everything")
	}
	if !SchemaNameMatchesFilter(testDBType, map[string]struct{}{"TestDB.dbo": {}}, "TestDB.dbo") {
		t.Fatalf("exact filter should match")
	}
}

func TestSplitCatalogSchema(t *testing.T) {
	cases := []struct {
		name    string
		dbType  string
		in      string
		catalog string
		schema  string
		whole   bool
	}{
		{"whole db", testDBType, "TestDB", "TestDB", "", true},
		{"db.schema", testDBType, "TestDB.dbo", "TestDB", "dbo", false},
		{"dotted catalog", testDBType, "my.db.dbo", "my.db", "dbo", false},
		{"empty", testDBType, "", "", "", false},
		{"non schema-level", "MySQL", "TestDB.dbo", "", "TestDB.dbo", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitCatalogSchema(tc.dbType, tc.in)
			if got.Catalog != tc.catalog || got.Schema != tc.schema || got.WholeDB != tc.whole {
				t.Fatalf("SplitCatalogSchema(%q, %q) = %+v, want catalog=%q schema=%q whole=%v",
					tc.dbType, tc.in, got, tc.catalog, tc.schema, tc.whole)
			}
		})
	}
}

func TestConnectionCatalogForMergedName(t *testing.T) {
	if got := ConnectionCatalogForMergedName(testDBType, "TestDB.dbo"); got != "TestDB" {
		t.Fatalf("got %q, want TestDB", got)
	}
	if got := ConnectionCatalogForMergedName(testDBType, "TestDB"); got != "TestDB" {
		t.Fatalf("whole-db name should return catalog, got %q", got)
	}
	if got := ConnectionCatalogForMergedName("MySQL", "somedb"); got != "" {
		t.Fatalf("non schema-level should return empty, got %q", got)
	}
}
