package storage

import (
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"gorm.io/gorm"
)

// migrateProjectDirectorToSystemAdmin migrates users holding the "project director"
// role (op_permission_uid=700001) to the "system administrator" role (700017).
//
// The migration is idempotent: repeated execution matches zero rows and produces no errors.
//
// Step 1: For users who hold both 700001 and 700017 simultaneously, delete the 700001 record
//         to avoid unique constraint violations.
// Step 2: For users who only hold 700001, update it to 700017.
//
// data-upgrade: B-20260508_project_director_to_system_admin
func migrateProjectDirectorToSystemAdmin(db *gorm.DB) error {
	// Step 1: delete duplicate 700001 records where user already has 700017
	err := db.Exec(`
		DELETE FROM user_op_permissions
		WHERE op_permission_uid = ?
		  AND user_uid IN (
		    SELECT user_uid FROM (
		      SELECT user_uid FROM user_op_permissions WHERE op_permission_uid = ?
		    ) AS tmp
		  )
	`, pkgConst.UIDOfOpPermissionCreateProject, pkgConst.UIDOfOpPermissionGlobalManagement).Error
	if err != nil {
		return fmt.Errorf("failed to delete duplicate project director records: %w", err)
	}

	// Step 2: migrate remaining 700001 to 700017
	err = db.Exec(`
		UPDATE user_op_permissions
		SET op_permission_uid = ?
		WHERE op_permission_uid = ?
	`, pkgConst.UIDOfOpPermissionGlobalManagement, pkgConst.UIDOfOpPermissionCreateProject).Error
	if err != nil {
		return fmt.Errorf("failed to migrate project director to system admin: %w", err)
	}

	return nil
}
