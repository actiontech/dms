package biz

import (
	"context"
	"testing"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/stretchr/testify/assert"
)

// mockTx implements TransactionGenerator for testing
type mockTx struct{}

func (m *mockTx) BeginTX(_ context.Context) RepoTX {
	return &mockRepoTx{}
}

// mockRepoTx implements RepoTX for testing
type mockRepoTx struct {
	context.Context
}

func (m *mockRepoTx) Commit(_ *utilLog.Helper) error                    { return nil }
func (m *mockRepoTx) RollbackWithError(_ *utilLog.Helper, err error) error { return err }
func (m *mockRepoTx) Deadline() (time.Time, bool)                      { return time.Time{}, false }
func (m *mockRepoTx) Done() <-chan struct{}                             { return nil }
func (m *mockRepoTx) Err() error                                       { return nil }
func (m *mockRepoTx) Value(_ any) any                                  { return nil }

// mockUserRepoForUpdate extends mockUserRepo to capture updated user
type mockUserRepoForUpdate struct {
	mockUserRepo
	updatedUser *User
}

func (m *mockUserRepoForUpdate) UpdateUser(_ context.Context, u *User) error {
	m.updatedUser = u
	return nil
}

func (m *mockUserRepoForUpdate) ReplaceUserGroupsInUser(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockUserRepoForUpdate) ReplaceOpPermissionsInUser(_ context.Context, _ string, _ []string) error {
	return nil
}

// mockUserGroupRepoForTest implements UserGroupRepo for testing
type mockUserGroupRepoForTest struct{}

func (m *mockUserGroupRepoForTest) SaveUserGroup(context.Context, *UserGroup) error { return nil }
func (m *mockUserGroupRepoForTest) UpdateUserGroup(context.Context, *UserGroup) error { return nil }
func (m *mockUserGroupRepoForTest) CheckUserGroupExist(_ context.Context, _ []string) (bool, error) {
	return true, nil
}
func (m *mockUserGroupRepoForTest) ListUserGroups(context.Context, *ListUserGroupsOption) ([]*UserGroup, int64, error) {
	return nil, 0, nil
}
func (m *mockUserGroupRepoForTest) DelUserGroup(context.Context, string) error { return nil }
func (m *mockUserGroupRepoForTest) GetUserGroup(context.Context, string) (*UserGroup, error) {
	return nil, nil
}
func (m *mockUserGroupRepoForTest) AddUserToUserGroup(context.Context, string, string) error {
	return nil
}
func (m *mockUserGroupRepoForTest) ReplaceUsersInUserGroup(context.Context, string, []string) error {
	return nil
}
func (m *mockUserGroupRepoForTest) GetUsersInUserGroup(context.Context, string) ([]*User, error) {
	return nil, nil
}

// mockOpPermissionRepoForTest implements OpPermissionRepo for testing
type mockOpPermissionRepoForTest struct {
	permissions map[string]*OpPermission
}

func (m *mockOpPermissionRepoForTest) SaveOpPermission(context.Context, *OpPermission) error {
	return nil
}
func (m *mockOpPermissionRepoForTest) UpdateOpPermission(context.Context, *OpPermission) error {
	return nil
}
func (m *mockOpPermissionRepoForTest) CheckOpPermissionExist(_ context.Context, _ []string) (bool, error) {
	return true, nil
}
func (m *mockOpPermissionRepoForTest) ListOpPermissions(context.Context, *ListOpPermissionsOption) ([]*OpPermission, int64, error) {
	return nil, 0, nil
}
func (m *mockOpPermissionRepoForTest) DelOpPermission(context.Context, string) error { return nil }
func (m *mockOpPermissionRepoForTest) GetOpPermission(_ context.Context, uid string) (*OpPermission, error) {
	if p, ok := m.permissions[uid]; ok {
		return p, nil
	}
	return &OpPermission{UID: uid, RangeType: "global"}, nil
}

// TestUpdateUserBusinessWritePermission covers design.md 5.1.5 test matrix
// for the BusinessWritePermission field handling in UpdateUser.
func TestUpdateUserBusinessWritePermission(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	cases := []struct {
		name             string
		targetUserUID    string
		initialBWP       bool
		opPermissionUIDs []string // permissions to assign (determines if user is/becomes admin)
		argsBWP          *bool    // BWP value in UpdateUserArgs
		expectedBWP      bool
	}{
		{
			name:             "set_bwp_false",
			targetUserUID:    pkgConst.UIDOfUserAdmin,
			initialBWP:       true,
			opPermissionUIDs: []string{},
			argsBWP:          boolPtr(false),
			expectedBWP:      false,
		},
		{
			name:             "set_bwp_true",
			targetUserUID:    pkgConst.UIDOfUserAdmin,
			initialBWP:       false,
			opPermissionUIDs: []string{},
			argsBWP:          boolPtr(true),
			expectedBWP:      true,
		},
		{
			name:             "role_switch_reset",
			targetUserUID:    "normal_user_1",
			initialBWP:       false,
			opPermissionUIDs: []string{}, // no global management -> not admin
			argsBWP:          boolPtr(false),
			expectedBWP:      true, // reset to true because not admin
		},
		{
			name:             "non_admin_bwp_ignored",
			targetUserUID:    "normal_user_1",
			initialBWP:       true,
			opPermissionUIDs: []string{}, // no global management -> not admin
			argsBWP:          boolPtr(false),
			expectedBWP:      true, // ignored, stays true
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup user repo with current user (admin performing the operation) and target user
			userRepo := &mockUserRepoForUpdate{
				mockUserRepo: mockUserRepo{
					users: map[string]*User{
						pkgConst.UIDOfUserAdmin: {
							UID:                     pkgConst.UIDOfUserAdmin,
							Name:                    "admin",
							Password:                "admin_pass",
							BusinessWritePermission: true,
						},
						tc.targetUserUID: {
							UID:                     tc.targetUserUID,
							Name:                    "target_user",
							Password:                "user_pass",
							BusinessWritePermission: tc.initialBWP,
						},
					},
				},
			}

			opPermVerifyRepo := &mockOpPermissionVerifyRepo{
				globalPermissions:  map[string][]*OpPermission{},
				projectPermissions: map[string]map[string]map[string]bool{},
			}

			opPermRepo := &mockOpPermissionRepoForTest{
				permissions: map[string]*OpPermission{
					pkgConst.UIDOfOpPermissionGlobalManagement: {
						UID:       pkgConst.UIDOfOpPermissionGlobalManagement,
						RangeType: "global",
					},
				},
			}

			opPermVerifyUsecase := NewOpPermissionVerifyUsecase(&noopLogger{}, nil, opPermVerifyRepo, userRepo)
			opPermUsecase := NewOpPermissionUsecase(&noopLogger{}, nil, opPermRepo, nil)

			userUsecase := &UserUsecase{
				tx:                        &mockTx{},
				repo:                      userRepo,
				userGroupRepo:             &mockUserGroupRepoForTest{},
				opPermissionUsecase:       opPermUsecase,
				OpPermissionVerifyUsecase: opPermVerifyUsecase,
				log:                       utilLog.NewHelper(&noopLogger{}, utilLog.WithMessageKey("test")),
			}

			args := &UpdateUserArgs{
				UserUID:                 tc.targetUserUID,
				OpPermissionUIDs:        tc.opPermissionUIDs,
				BusinessWritePermission: tc.argsBWP,
			}

			err := userUsecase.UpdateUser(context.Background(), pkgConst.UIDOfUserAdmin, args)
			assert.NoError(t, err, "case: %s", tc.name)
			assert.NotNil(t, userRepo.updatedUser, "case: %s - user should have been updated", tc.name)
			assert.Equal(t, tc.expectedBWP, userRepo.updatedUser.BusinessWritePermission, "case: %s", tc.name)
		})
	}
}

// TestUpdateUserBWPWithGlobalManagementPermission tests that a user being assigned
// global management permission is treated as system administrator for BWP purposes.
func TestUpdateUserBWPWithGlobalManagementPermission(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	userRepo := &mockUserRepoForUpdate{
		mockUserRepo: mockUserRepo{
			users: map[string]*User{
				pkgConst.UIDOfUserAdmin: {
					UID:                     pkgConst.UIDOfUserAdmin,
					Name:                    "admin",
					Password:                "admin_pass",
					BusinessWritePermission: true,
				},
				"user_with_global_mgmt": {
					UID:                     "user_with_global_mgmt",
					Name:                    "global_mgmt_user",
					Password:                "user_pass",
					BusinessWritePermission: true,
				},
			},
		},
	}

	opPermVerifyRepo := &mockOpPermissionVerifyRepo{
		globalPermissions: map[string][]*OpPermission{
			"user_with_global_mgmt": {
				{UID: pkgConst.UIDOfOpPermissionGlobalManagement},
			},
		},
		projectPermissions: map[string]map[string]map[string]bool{},
	}

	opPermRepo := &mockOpPermissionRepoForTest{
		permissions: map[string]*OpPermission{
			pkgConst.UIDOfOpPermissionGlobalManagement: {
				UID:       pkgConst.UIDOfOpPermissionGlobalManagement,
				RangeType: "global",
			},
		},
	}

	opPermVerifyUsecase := NewOpPermissionVerifyUsecase(&noopLogger{}, nil, opPermVerifyRepo, userRepo)
	opPermUsecase := NewOpPermissionUsecase(&noopLogger{}, nil, opPermRepo, nil)

	userUsecase := &UserUsecase{
		tx:                        &mockTx{},
		repo:                      userRepo,
		userGroupRepo:             &mockUserGroupRepoForTest{},
		opPermissionUsecase:       opPermUsecase,
		OpPermissionVerifyUsecase: opPermVerifyUsecase,
		log:                       utilLog.NewHelper(&noopLogger{}, utilLog.WithMessageKey("test")),
	}

	// User with global management permission setting BWP=false should be accepted
	args := &UpdateUserArgs{
		UserUID:                 "user_with_global_mgmt",
		OpPermissionUIDs:        []string{pkgConst.UIDOfOpPermissionGlobalManagement},
		BusinessWritePermission: boolPtr(false),
	}

	err := userUsecase.UpdateUser(context.Background(), pkgConst.UIDOfUserAdmin, args)
	assert.NoError(t, err)
	assert.NotNil(t, userRepo.updatedUser)
	assert.Equal(t, false, userRepo.updatedUser.BusinessWritePermission,
		"user with global management permission should be able to set BWP=false")
}
