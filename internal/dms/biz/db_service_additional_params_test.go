package biz

import (
	"reflect"
	"testing"

	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func TestFilterAdditionalParamsByComponentContract(t *testing.T) {
	databaseName := &dmsCommonV1.AdditionalParam{Name: "database_name", Value: "testdb"}
	defaultSchema := &dmsCommonV1.AdditionalParam{Name: "default_schema", Value: "public"}
	typo := &dmsCommonV1.AdditionalParam{Name: "databas_name", Value: "x"}

	sqleAllow := map[string]struct{}{"database_name": {}, "default_schema": {}}
	provisionAllow := map[string]struct{}{"database_name": {}}
	union := map[string]struct{}{"database_name": {}, "default_schema": {}}

	t.Run("AC-001 sqle keeps contract params including database_name", func(t *testing.T) {
		in := []*dmsCommonV1.AdditionalParam{databaseName, defaultSchema}
		got := filterAdditionalParamsByComponentContract(in, sqleAllow, union)
		if !reflect.DeepEqual(got, in) {
			t.Fatalf("sqle filtered=%v want=%v", namesOf(got), namesOf(in))
		}
	})

	t.Run("AC-001 provision keeps database_name drops default_schema", func(t *testing.T) {
		in := []*dmsCommonV1.AdditionalParam{databaseName, defaultSchema}
		got := filterAdditionalParamsByComponentContract(in, provisionAllow, union)
		if len(got) != 1 || got[0].Name != "database_name" {
			t.Fatalf("provision filtered=%v want=[database_name]", namesOf(got))
		}
	})

	t.Run("AC-001 legacy sqle without database_name strips cross-component catalog", func(t *testing.T) {
		legacySqle := map[string]struct{}{"default_schema": {}}
		in := []*dmsCommonV1.AdditionalParam{databaseName, defaultSchema}
		got := filterAdditionalParamsByComponentContract(in, legacySqle, union)
		if len(got) != 1 || got[0].Name != "default_schema" {
			t.Fatalf("legacy sqle filtered=%v want=[default_schema]", namesOf(got))
		}
	})

	t.Run("AC-004 default_schema retained for sqle", func(t *testing.T) {
		in := []*dmsCommonV1.AdditionalParam{defaultSchema}
		got := filterAdditionalParamsByComponentContract(in, sqleAllow, union)
		if len(got) != 1 || got[0].Value != "public" {
			t.Fatalf("default_schema not retained: %#v", got)
		}
	})

	t.Run("AC-005 unknown typo kept so component can reject", func(t *testing.T) {
		in := []*dmsCommonV1.AdditionalParam{databaseName, typo}
		got := filterAdditionalParamsByComponentContract(in, sqleAllow, union)
		if len(got) != 2 {
			t.Fatalf("typo should not be swallowed, got=%v", namesOf(got))
		}
		foundTypo := false
		for _, p := range got {
			if p.Name == "databas_name" {
				foundTypo = true
			}
		}
		if !foundTypo {
			t.Fatal("expected typo param to remain for explicit failure")
		}
	})

	t.Run("empty union falls back to original slice", func(t *testing.T) {
		in := []*dmsCommonV1.AdditionalParam{databaseName, defaultSchema}
		got := filterAdditionalParamsByComponentContract(in, sqleAllow, nil)
		if !reflect.DeepEqual(got, in) {
			t.Fatalf("fallback failed: got=%v", namesOf(got))
		}
	})
}

func namesOf(params []*dmsCommonV1.AdditionalParam) []string {
	out := make([]string, 0, len(params))
	for _, p := range params {
		if p != nil {
			out = append(out, p.Name)
		}
	}
	return out
}
