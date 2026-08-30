package biz

import pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"

// FilterDBServicesBySQLWorkbenchAccess applies the shared SQL workbench access rules:
// project admins can access all active DB services in the project, and users with
// SQL query permission on a DB service can access that specific service.
func FilterDBServicesBySQLWorkbenchAccess(dbServices []*DBService, opPermissions []OpPermissionWithOpRange) []*DBService {
	projectIDMap := make(map[string]struct{})
	dbServiceIDMap := make(map[string]struct{})

	for _, opPermission := range opPermissions {
		if opPermission.OpRangeType == OpRangeTypeProject && opPermission.OpPermissionUID == pkgConst.UIDOfOpPermissionProjectAdmin {
			for _, rangeUID := range opPermission.RangeUIDs {
				projectIDMap[rangeUID] = struct{}{}
			}
		}

		if opPermission.OpRangeType == OpRangeTypeDBService && opPermission.OpPermissionUID == pkgConst.UIDOfOpPermissionSQLQuery {
			for _, rangeUID := range opPermission.RangeUIDs {
				dbServiceIDMap[rangeUID] = struct{}{}
			}
		}
	}

	filteredDBServices := make([]*DBService, 0, len(dbServices))
	for _, dbService := range dbServices {
		if _, hasProjectPermission := projectIDMap[dbService.ProjectUID]; hasProjectPermission {
			filteredDBServices = append(filteredDBServices, dbService)
			continue
		}

		if _, hasDBServicePermission := dbServiceIDMap[dbService.UID]; hasDBServicePermission {
			filteredDBServices = append(filteredDBServices, dbService)
		}
	}

	return filteredDBServices
}
