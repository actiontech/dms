package biz

import (
	"context"
	"fmt"
	"testing"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/stretchr/testify/assert"
)

// mockUserRepo implements UserRepo for testing
type mockUserRepo struct {
	users map[string]*User
}

func (m *mockUserRepo) GetUser(_ context.Context, uid string) (*User, error) {
	if u, ok := m.users[uid]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("user %s not found", uid)
}

// Unused methods to satisfy interface
func (m *mockUserRepo) SaveUser(context.Context, *User) error                          { return nil }
func (m *mockUserRepo) UpdateUser(context.Context, *User) error                        { return nil }
func (m *mockUserRepo) CheckUserExist(context.Context, []string) (bool, error)         { return false, nil }
func (m *mockUserRepo) ListUsers(context.Context, *ListUsersOption) ([]*User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) CountUsers(context.Context, []pkgConst.FilterCondition) (int64, error) {
	return 0, nil
}
func (m *mockUserRepo) DelUser(context.Context, string) error                          { return nil }
func (m *mockUserRepo) GetUserIncludeDeleted(context.Context, string) (*User, error)   { return nil, nil }
func (m *mockUserRepo) GetUserByName(context.Context, string) (*User, error)           { return nil, nil }
func (m *mockUserRepo) AddUserToUserGroup(context.Context, string, string) error       { return nil }
func (m *mockUserRepo) DelUserFromAllUserGroups(context.Context, string) error         { return nil }
func (m *mockUserRepo) ReplaceUserGroupsInUser(context.Context, string, []string) error { return nil }
func (m *mockUserRepo) ReplaceOpPermissionsInUser(context.Context, string, []string) error {
	return nil
}
func (m *mockUserRepo) GetUserGroupsByUser(context.Context, string) ([]*UserGroup, error) {
	return nil, nil
}
func (m *mockUserRepo) GetOpPermissionsByUser(context.Context, string) ([]*OpPermission, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUserByThirdPartyUserID(context.Context, string) (*User, error) {
	return nil, nil
}
func (m *mockUserRepo) SaveAccessToken(context.Context, *AccessTokenInfo) error { return nil }
func (m *mockUserRepo) GetAccessTokenByUser(context.Context, string) (*AccessTokenInfo, error) {
	return nil, nil
}
func (m *mockUserRepo) RecordLoginSession(context.Context, string, string) error { return nil }
func (m *mockUserRepo) GetLatestLoginSession(context.Context, string) (string, bool, error) {
	return "", false, nil
}

// mockOpPermissionVerifyRepo implements OpPermissionVerifyRepo for testing
type mockOpPermissionVerifyRepo struct {
	// projectPermissions maps userUID -> projectUID -> opPermissionUID -> has
	projectPermissions map[string]map[string]map[string]bool
	// globalPermissions maps userUID -> list of OpPermission
	globalPermissions map[string][]*OpPermission
	// projectMembers stores project members for ListUsersOpPermissionInProject
	projectMembers map[string][]ListMembersOpPermissionItem
}

func (m *mockOpPermissionVerifyRepo) IsUserHasOpPermissionInProject(_ context.Context, userUid, projectUid, opPermissionUid string) (bool, error) {
	if projects, ok := m.projectPermissions[userUid]; ok {
		if perms, ok := projects[projectUid]; ok {
			return perms[opPermissionUid], nil
		}
	}
	return false, nil
}

func (m *mockOpPermissionVerifyRepo) GetUserGlobalOpPermission(_ context.Context, userUid string) ([]*OpPermission, error) {
	return m.globalPermissions[userUid], nil
}

func (m *mockOpPermissionVerifyRepo) ListUsersOpPermissionInProject(_ context.Context, projectUid string, _ *ListMembersOpPermissionOption) ([]ListMembersOpPermissionItem, int64, error) {
	items := m.projectMembers[projectUid]
	return items, int64(len(items)), nil
}

// Unused methods
func (m *mockOpPermissionVerifyRepo) GetUserOpPermissionInProject(context.Context, string, string) ([]OpPermissionWithOpRange, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) GetOneOpPermissionInProject(context.Context, string, string, string) ([]OpPermissionWithOpRange, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) GetUserProjectOpPermissionInProject(context.Context, string, string) ([]OpPermissionWithOpRange, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) GetUserOpPermission(context.Context, string) ([]OpPermissionWithOpRange, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) GetUserProjectOpPermission(context.Context, string) ([]OpPermissionWithOpRange, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) GetUserProjectWithOpPermissions(context.Context, string) ([]ProjectOpPermissionWithOpRange, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) GetUserProject(context.Context, string) ([]*Project, error) {
	return nil, nil
}
func (m *mockOpPermissionVerifyRepo) ListUsersInProject(context.Context, string) ([]ListMembersOpPermissionItem, error) {
	return nil, nil
}

// noopLogger satisfies utilLog.Logger for testing
type noopLogger struct{}

func (n *noopLogger) Log(_ utilLog.Level, _ ...interface{}) error { return nil }

func newTestOpPermissionVerifyUsecase(userRepo UserRepo, opRepo OpPermissionVerifyRepo) *OpPermissionVerifyUsecase {
	return NewOpPermissionVerifyUsecase(&noopLogger{}, nil, opRepo, userRepo)
}

// TestCanOpGlobal covers design.md 5.1.1 test matrix for CanOpGlobal function.
func TestCanOpGlobal(t *testing.T) {
	cases := []struct {
		name            string
		userUID         string
		bwp             bool
		isBusinessWrite bool
		want            bool
	}{
		{
			name:            "admin_bwp_on_business",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             true,
			isBusinessWrite: true,
			want:            true,
		},
		{
			name:            "admin_bwp_off_business",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             false,
			isBusinessWrite: true,
			want:            false,
		},
		{
			name:            "admin_bwp_off_resource",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             false,
			isBusinessWrite: false,
			want:            true,
		},
		{
			name:            "sysadmin_bwp_on_business",
			userUID:         pkgConst.UIDOfUserSys,
			bwp:             true,
			isBusinessWrite: true,
			want:            true,
		},
		{
			name:            "sysadmin_bwp_off_business",
			userUID:         pkgConst.UIDOfUserSys,
			bwp:             false,
			isBusinessWrite: true,
			want:            false,
		},
		{
			name:            "sysadmin_bwp_off_resource",
			userUID:         pkgConst.UIDOfUserSys,
			bwp:             false,
			isBusinessWrite: false,
			want:            true,
		},
		{
			name:            "normal_user_no_global_permission",
			userUID:         "normal_user_1",
			bwp:             true,
			isBusinessWrite: true,
			want:            false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				users: map[string]*User{
					tc.userUID: {UID: tc.userUID, BusinessWritePermission: tc.bwp},
				},
			}
			opRepo := &mockOpPermissionVerifyRepo{
				globalPermissions: map[string][]*OpPermission{},
			}
			uc := newTestOpPermissionVerifyUsecase(userRepo, opRepo)

			got, err := uc.CanOpGlobal(context.Background(), tc.userUID, tc.isBusinessWrite)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got, "case: %s", tc.name)
		})
	}
}

// TestIsUserProjectAdmin covers design.md 5.1.1 test matrix for IsUserProjectAdmin.
// When admin has BWP=false and isBusinessWrite=true, admin identity is bypassed
// and it falls through to check project-level authorization.
func TestIsUserProjectAdmin(t *testing.T) {
	const testProjectUID = "project_1"

	cases := []struct {
		name              string
		userUID           string
		bwp               bool
		isBusinessWrite   bool
		hasProjectAdmin   bool // whether user has ProjectAdmin permission in project
		want              bool
	}{
		{
			name:            "admin_bwp_on_business_write",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             true,
			isBusinessWrite: true,
			hasProjectAdmin: false,
			want:            true,
		},
		{
			name:            "admin_bwp_off_business_write_no_project_auth",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             false,
			isBusinessWrite: true,
			hasProjectAdmin: false,
			want:            false,
		},
		{
			name:            "admin_bwp_off_business_write_with_project_auth",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             false,
			isBusinessWrite: true,
			hasProjectAdmin: true,
			want:            true,
		},
		{
			name:            "admin_bwp_off_resource_config",
			userUID:         pkgConst.UIDOfUserAdmin,
			bwp:             false,
			isBusinessWrite: false,
			hasProjectAdmin: false,
			want:            true,
		},
		{
			name:            "sysadmin_bwp_off_business_fallthrough_with_project",
			userUID:         pkgConst.UIDOfUserSys,
			bwp:             false,
			isBusinessWrite: true,
			hasProjectAdmin: true,
			want:            true,
		},
		{
			name:            "sysadmin_bwp_off_business_fallthrough_no_project",
			userUID:         pkgConst.UIDOfUserSys,
			bwp:             false,
			isBusinessWrite: true,
			hasProjectAdmin: false,
			want:            false,
		},
		{
			name:            "normal_user_with_project_admin",
			userUID:         "normal_user_1",
			bwp:             true,
			isBusinessWrite: true,
			hasProjectAdmin: true,
			want:            true,
		},
		{
			name:            "normal_user_without_project_admin",
			userUID:         "normal_user_1",
			bwp:             true,
			isBusinessWrite: true,
			hasProjectAdmin: false,
			want:            false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				users: map[string]*User{
					tc.userUID: {UID: tc.userUID, BusinessWritePermission: tc.bwp},
				},
			}
			projectPerms := make(map[string]map[string]map[string]bool)
			if tc.hasProjectAdmin {
				projectPerms[tc.userUID] = map[string]map[string]bool{
					testProjectUID: {pkgConst.UIDOfOpPermissionProjectAdmin: true},
				}
			}
			opRepo := &mockOpPermissionVerifyRepo{
				projectPermissions: projectPerms,
				globalPermissions:  map[string][]*OpPermission{},
			}
			uc := newTestOpPermissionVerifyUsecase(userRepo, opRepo)

			got, err := uc.IsUserProjectAdmin(context.Background(), tc.userUID, testProjectUID, tc.isBusinessWrite)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got, "case: %s", tc.name)
		})
	}
}

// TestGetCanOpDBUsers covers design.md 5.1.1 and 5.1.2 test matrix for GetCanOpDBUsers.
// BWP=false system administrators should only be included via project-level authorization
// when isBusinessWrite=true.
func TestGetCanOpDBUsers(t *testing.T) {
	const (
		testProjectUID   = "project_1"
		testDBServiceUID = "db_1"
	)
	opPermExportApproval := pkgConst.UIDOfOpPermissionExportApprovalReject

	cases := []struct {
		name            string
		members         []ListMembersOpPermissionItem
		users           map[string]*User
		isBusinessWrite bool
		wantUserUIDs    []string
	}{
		{
			name: "all_admins_bwp_on",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:  pkgConst.UIDOfUserAdmin,
					UserName: "admin",
					OpPermissions: []OpPermissionWithOpRange{
						{OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin},
					},
				},
			},
			users: map[string]*User{
				pkgConst.UIDOfUserAdmin: {UID: pkgConst.UIDOfUserAdmin, BusinessWritePermission: true},
			},
			isBusinessWrite: true,
			wantUserUIDs:    []string{pkgConst.UIDOfUserAdmin},
		},
		{
			name: "admin_bwp_off_no_project_auth",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:  pkgConst.UIDOfUserAdmin,
					UserName: "admin",
					OpPermissions: []OpPermissionWithOpRange{
						{OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin},
					},
				},
			},
			users: map[string]*User{
				pkgConst.UIDOfUserAdmin: {UID: pkgConst.UIDOfUserAdmin, BusinessWritePermission: false},
			},
			isBusinessWrite: true,
			// Admin has ProjectAdmin but BWP=false skips admin privilege;
			// userCanOpDBWithoutAdminPrivilege skips ProjectAdmin -> not included
			wantUserUIDs: []string{},
		},
		{
			name: "admin_bwp_off_with_explicit_db_permission",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:  pkgConst.UIDOfUserAdmin,
					UserName: "admin",
					OpPermissions: []OpPermissionWithOpRange{
						{OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin},
						{
							OpPermissionUID: opPermExportApproval,
							OpRangeType:     OpRangeType(dmsV1.OpRangeTypeDBService),
							RangeUIDs:       []string{testDBServiceUID},
						},
					},
				},
			},
			users: map[string]*User{
				pkgConst.UIDOfUserAdmin: {UID: pkgConst.UIDOfUserAdmin, BusinessWritePermission: false},
			},
			isBusinessWrite: true,
			// Admin has BWP=false but also has explicit DB permission -> included via project auth
			wantUserUIDs: []string{pkgConst.UIDOfUserAdmin},
		},
		{
			name: "mixed_admin_and_normal_user",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:  pkgConst.UIDOfUserAdmin,
					UserName: "admin",
					OpPermissions: []OpPermissionWithOpRange{
						{OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin},
					},
				},
				{
					UserUid:  "normal_user_1",
					UserName: "normal",
					OpPermissions: []OpPermissionWithOpRange{
						{
							OpPermissionUID: opPermExportApproval,
							OpRangeType:     OpRangeType(dmsV1.OpRangeTypeDBService),
							RangeUIDs:       []string{testDBServiceUID},
						},
					},
				},
			},
			users: map[string]*User{
				pkgConst.UIDOfUserAdmin: {UID: pkgConst.UIDOfUserAdmin, BusinessWritePermission: false},
				"normal_user_1":         {UID: "normal_user_1", BusinessWritePermission: true},
			},
			isBusinessWrite: true,
			// Only normal user has explicit DB permission; admin BWP=false without explicit auth
			wantUserUIDs: []string{"normal_user_1"},
		},
		{
			name: "admin_bwp_off_resource_config_keeps_admin_privilege",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:  pkgConst.UIDOfUserAdmin,
					UserName: "admin",
					OpPermissions: []OpPermissionWithOpRange{
						{OpPermissionUID: pkgConst.UIDOfOpPermissionProjectAdmin},
					},
				},
			},
			users: map[string]*User{
				pkgConst.UIDOfUserAdmin: {UID: pkgConst.UIDOfUserAdmin, BusinessWritePermission: false},
			},
			isBusinessWrite: false,
			// Resource config: BWP doesn't affect, admin privilege applies
			wantUserUIDs: []string{pkgConst.UIDOfUserAdmin},
		},
		{
			name: "normal_user_with_project_auth",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:  "normal_user_1",
					UserName: "normal",
					OpPermissions: []OpPermissionWithOpRange{
						{
							OpPermissionUID: opPermExportApproval,
							OpRangeType:     OpRangeType(dmsV1.OpRangeTypeDBService),
							RangeUIDs:       []string{testDBServiceUID},
						},
					},
				},
			},
			users: map[string]*User{
				"normal_user_1": {UID: "normal_user_1", BusinessWritePermission: true},
			},
			isBusinessWrite: true,
			wantUserUIDs:    []string{"normal_user_1"},
		},
		{
			name: "normal_user_no_project_auth",
			members: []ListMembersOpPermissionItem{
				{
					UserUid:       "normal_user_1",
					UserName:      "normal",
					OpPermissions: []OpPermissionWithOpRange{},
				},
			},
			users: map[string]*User{
				"normal_user_1": {UID: "normal_user_1", BusinessWritePermission: true},
			},
			isBusinessWrite: true,
			wantUserUIDs:    []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &mockUserRepo{users: tc.users}
			opRepo := &mockOpPermissionVerifyRepo{
				projectMembers: map[string][]ListMembersOpPermissionItem{
					testProjectUID: tc.members,
				},
				globalPermissions: map[string][]*OpPermission{},
			}
			uc := newTestOpPermissionVerifyUsecase(userRepo, opRepo)

			got, err := uc.GetCanOpDBUsers(
				context.Background(),
				testProjectUID,
				testDBServiceUID,
				[]string{opPermExportApproval},
				tc.isBusinessWrite,
			)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.wantUserUIDs, got, "case: %s", tc.name)
		})
	}
}
