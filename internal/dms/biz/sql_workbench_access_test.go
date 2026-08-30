package biz

import (
	"testing"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

func TestFilterDBServicesBySQLWorkbenchAccess(t *testing.T) {
	dbServices := []*DBService{
		{UID: "db-1", ProjectUID: "project-1", Name: "alpha"},
		{UID: "db-2", ProjectUID: "project-1", Name: "beta"},
		{UID: "db-3", ProjectUID: "project-2", Name: "gamma"},
	}

	cases := map[string]struct {
		opPermissions []OpPermissionWithOpRange
		wantUIDs      []string
	}{
		"no workbench permission returns empty": {
			opPermissions: nil,
			wantUIDs:      nil,
		},
		"project admin can access all services in project": {
			opPermissions: []OpPermissionWithOpRange{{
				OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin,
				OpRangeType:     OpRangeTypeProject,
				RangeUIDs:       []string{"project-1"},
			}},
			wantUIDs: []string{"db-1", "db-2"},
		},
		"sql query permission grants specific datasource only": {
			opPermissions: []OpPermissionWithOpRange{{
				OpPermissionUID: pkgConst.UIDOfOpPermissionSQLQuery,
				OpRangeType:     OpRangeTypeDBService,
				RangeUIDs:       []string{"db-3"},
			}},
			wantUIDs: []string{"db-3"},
		},
		"project admin and datasource permission combine": {
			opPermissions: []OpPermissionWithOpRange{
				{
					OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin,
					OpRangeType:     OpRangeTypeProject,
					RangeUIDs:       []string{"project-1"},
				},
				{
					OpPermissionUID: pkgConst.UIDOfOpPermissionSQLQuery,
					OpRangeType:     OpRangeTypeDBService,
					RangeUIDs:       []string{"db-3"},
				},
			},
			wantUIDs: []string{"db-1", "db-2", "db-3"},
		},
		"unrelated permission does not grant workbench access": {
			opPermissions: []OpPermissionWithOpRange{{
				OpPermissionUID: pkgConst.UIDOfOpPermissionCreateWorkflow,
				OpRangeType:     OpRangeTypeProject,
				RangeUIDs:       []string{"project-1"},
			}},
			wantUIDs: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			filtered := FilterDBServicesBySQLWorkbenchAccess(dbServices, tc.opPermissions)
			gotUIDs := make([]string, 0, len(filtered))
			for _, dbService := range filtered {
				gotUIDs = append(gotUIDs, dbService.UID)
			}

			if len(gotUIDs) != len(tc.wantUIDs) {
				t.Fatalf("filtered count = %d, want %d, got %v", len(gotUIDs), len(tc.wantUIDs), gotUIDs)
			}
			for idx := range tc.wantUIDs {
				if gotUIDs[idx] != tc.wantUIDs[idx] {
					t.Fatalf("filtered[%d] = %s, want %s (all=%v)", idx, gotUIDs[idx], tc.wantUIDs[idx], gotUIDs)
				}
			}
		})
	}
}
