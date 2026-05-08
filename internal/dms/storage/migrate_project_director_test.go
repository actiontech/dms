package storage

import (
	"fmt"
	"testing"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
)

// TestMigrateProjectDirectorToSystemAdmin tests the migration function
// using the existing MySQL test infrastructure (test_util.go).
//
// Test matrix from design.md 5.1.4:
//   - only_700001: user only holds 700001 -> migrated to 700017
//   - both_700001_700017: user holds both -> only 700017 remains
//   - repeat_migration: already migrated -> no change, no error
//   - migrated_user_bwp: after migration -> BusinessWritePermission=true (default)
func TestMigrateProjectDirectorToSystemAdmin(t *testing.T) {
	s := GetTestStorage(t)
	db := s.db

	// Clean up test data before and after
	cleanup := func() {
		db.Exec("DELETE FROM user_op_permissions WHERE user_uid LIKE 'test_migrate_%'")
		db.Exec("DELETE FROM users WHERE uid LIKE 'test_migrate_%'")
		db.Exec("DELETE FROM op_permissions WHERE uid IN (?, ?)",
			pkgConst.UIDOfOpPermissionCreateProject,
			pkgConst.UIDOfOpPermissionGlobalManagement)
	}
	cleanup()
	t.Cleanup(cleanup)

	// Ensure required op_permissions records exist
	db.Exec("INSERT IGNORE INTO op_permissions (uid, name, created_at, updated_at) VALUES (?, '项目总监', NOW(), NOW())",
		pkgConst.UIDOfOpPermissionCreateProject)
	db.Exec("INSERT IGNORE INTO op_permissions (uid, name, created_at, updated_at) VALUES (?, '系统管理员', NOW(), NOW())",
		pkgConst.UIDOfOpPermissionGlobalManagement)

	cases := []struct {
		name           string
		setupSQL       []string
		wantCount700001 int64
		wantCount700017 int64
		desc           string
	}{
		{
			name: "only_700001",
			setupSQL: []string{
				fmt.Sprintf("INSERT INTO users (uid, name, password, user_authentication_type, stat, created_at, updated_at) VALUES ('test_migrate_u1', 'user1', '', 'dms', 0, NOW(), NOW())"),
				fmt.Sprintf("INSERT INTO user_op_permissions (user_uid, op_permission_uid) VALUES ('test_migrate_u1', '%s')", pkgConst.UIDOfOpPermissionCreateProject),
			},
			wantCount700001: 0,
			wantCount700017: 1,
			desc:            "User only holding 700001 should be migrated to 700017",
		},
		{
			name: "both_700001_700017",
			setupSQL: []string{
				fmt.Sprintf("INSERT INTO users (uid, name, password, user_authentication_type, stat, created_at, updated_at) VALUES ('test_migrate_u2', 'user2', '', 'dms', 0, NOW(), NOW())"),
				fmt.Sprintf("INSERT INTO user_op_permissions (user_uid, op_permission_uid) VALUES ('test_migrate_u2', '%s')", pkgConst.UIDOfOpPermissionCreateProject),
				fmt.Sprintf("INSERT INTO user_op_permissions (user_uid, op_permission_uid) VALUES ('test_migrate_u2', '%s')", pkgConst.UIDOfOpPermissionGlobalManagement),
			},
			wantCount700001: 0,
			wantCount700017: 1,
			desc:            "User holding both 700001 and 700017 should only retain 700017",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean before each case
			db.Exec("DELETE FROM user_op_permissions WHERE user_uid LIKE 'test_migrate_%'")
			db.Exec("DELETE FROM users WHERE uid LIKE 'test_migrate_%'")

			// Setup
			for _, sql := range tc.setupSQL {
				if err := db.Exec(sql).Error; err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Execute migration
			if err := migrateProjectDirectorToSystemAdmin(db); err != nil {
				t.Fatalf("migration failed: %v", err)
			}

			// Verify 700001 count
			var count700001 int64
			db.Raw("SELECT COUNT(*) FROM user_op_permissions WHERE user_uid LIKE 'test_migrate_%' AND op_permission_uid = ?",
				pkgConst.UIDOfOpPermissionCreateProject).Scan(&count700001)
			if count700001 != tc.wantCount700001 {
				t.Errorf("%s: want 700001 count=%d, got=%d", tc.desc, tc.wantCount700001, count700001)
			}

			// Verify 700017 count
			var count700017 int64
			db.Raw("SELECT COUNT(*) FROM user_op_permissions WHERE user_uid LIKE 'test_migrate_%' AND op_permission_uid = ?",
				pkgConst.UIDOfOpPermissionGlobalManagement).Scan(&count700017)
			if count700017 != tc.wantCount700017 {
				t.Errorf("%s: want 700017 count=%d, got=%d", tc.desc, tc.wantCount700017, count700017)
			}
		})
	}

	// Test repeat_migration: run migration again on already-migrated data
	t.Run("repeat_migration", func(t *testing.T) {
		// Data from previous tests is already migrated, run again
		if err := migrateProjectDirectorToSystemAdmin(db); err != nil {
			t.Fatalf("repeat migration should not fail: %v", err)
		}

		var count700001 int64
		db.Raw("SELECT COUNT(*) FROM user_op_permissions WHERE user_uid LIKE 'test_migrate_%' AND op_permission_uid = ?",
			pkgConst.UIDOfOpPermissionCreateProject).Scan(&count700001)
		if count700001 != 0 {
			t.Errorf("repeat migration: want 700001 count=0, got=%d", count700001)
		}
	})

	// Test migrated_user_bwp: verify BWP default value after migration
	t.Run("migrated_user_bwp", func(t *testing.T) {
		var bwp bool
		err := db.Raw("SELECT business_write_permission FROM users WHERE uid = 'test_migrate_u1'").Scan(&bwp).Error
		if err != nil {
			t.Fatalf("failed to query BWP: %v", err)
		}
		if !bwp {
			t.Errorf("migrated user BWP should be true (default), got false")
		}
	})
}
