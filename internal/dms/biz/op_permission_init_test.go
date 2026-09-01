package biz

import (
	"context"
	"io"
	"testing"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/stretchr/testify/assert"
)

// fakeOpPermissionRepo 表驱动 InitOpPermissions：已存在 skip / 缺失插入。
type fakeOpPermissionRepo struct {
	byUID map[string]*OpPermission
	saved []string
}

func newFakeOpPermissionRepo(existing ...*OpPermission) *fakeOpPermissionRepo {
	m := make(map[string]*OpPermission)
	for _, op := range existing {
		cp := *op
		m[op.UID] = &cp
	}
	return &fakeOpPermissionRepo{byUID: m}
}

func (f *fakeOpPermissionRepo) SaveOpPermission(_ context.Context, op *OpPermission) error {
	cp := *op
	f.byUID[op.UID] = &cp
	f.saved = append(f.saved, op.UID)
	return nil
}
func (f *fakeOpPermissionRepo) UpdateOpPermission(context.Context, *OpPermission) error { return nil }
func (f *fakeOpPermissionRepo) CheckOpPermissionExist(context.Context, []string) (bool, error) {
	return true, nil
}
func (f *fakeOpPermissionRepo) ListOpPermissions(context.Context, *ListOpPermissionsOption) ([]*OpPermission, int64, error) {
	return nil, 0, nil
}
func (f *fakeOpPermissionRepo) DelOpPermission(context.Context, string) error { return nil }
func (f *fakeOpPermissionRepo) GetOpPermission(_ context.Context, uid string) (*OpPermission, error) {
	if op, ok := f.byUID[uid]; ok {
		cp := *op
		return &cp, nil
	}
	return nil, pkgErr.ErrStorageNoData
}

// TestInitOpPermissions_PrivilegeApplyAudit_Idempotent 对应方案 B-privilege_apply_audit_op_permission：
// 已存在 UID skip；缺失则插入；重复 Init 不产生第二行。
func TestInitOpPermissions_PrivilegeApplyAudit_Idempotent(t *testing.T) {
	logger := utilLog.NewMyLogger(io.Discard)
	auditSeed := &OpPermission{
		UID:       pkgConst.UIdOfOpPermissionPrivilegeApplyAudit,
		Name:      "提权申请审批",
		RangeType: OpRangeTypeDBService,
	}

	cases := []struct {
		name          string
		existing      []*OpPermission
		candidates    []*OpPermission
		wantSavedUIDs []string
		wantPresent   []string
	}{
		{
			name:          "missing_privilege_apply_audit_inserts",
			existing:      nil,
			candidates:    []*OpPermission{auditSeed},
			wantSavedUIDs: []string{pkgConst.UIdOfOpPermissionPrivilegeApplyAudit},
			wantPresent:   []string{pkgConst.UIdOfOpPermissionPrivilegeApplyAudit},
		},
		{
			name: "existing_uid_skips_insert",
			existing: []*OpPermission{{
				UID:       pkgConst.UIdOfOpPermissionPrivilegeApplyAudit,
				Name:      "提权申请审批",
				RangeType: OpRangeTypeDBService,
			}},
			candidates:    []*OpPermission{auditSeed},
			wantSavedUIDs: nil,
			wantPresent:   []string{pkgConst.UIdOfOpPermissionPrivilegeApplyAudit},
		},
		{
			name: "table_driven_mix_skip_and_insert",
			existing: []*OpPermission{{
				UID:       pkgConst.UIDOfOpPermissionProjectAdmin,
				Name:      "项目管理",
				RangeType: OpRangeTypeProject,
			}},
			candidates: []*OpPermission{
				{UID: pkgConst.UIDOfOpPermissionProjectAdmin, Name: "项目管理", RangeType: OpRangeTypeProject},
				auditSeed,
			},
			wantSavedUIDs: []string{pkgConst.UIdOfOpPermissionPrivilegeApplyAudit},
			wantPresent: []string{
				pkgConst.UIDOfOpPermissionProjectAdmin,
				pkgConst.UIdOfOpPermissionPrivilegeApplyAudit,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeOpPermissionRepo(tc.existing...)
			uc := NewOpPermissionUsecase(logger, nil, repo, nil)
			assert.NoError(t, uc.InitOpPermissions(context.Background(), tc.candidates))
			assert.Equal(t, tc.wantSavedUIDs, repo.saved)

			// 再次 Init：幂等，不再 Save
			savedBefore := len(repo.saved)
			assert.NoError(t, uc.InitOpPermissions(context.Background(), tc.candidates))
			assert.Equal(t, savedBefore, len(repo.saved), "second Init must not insert again")

			for _, uid := range tc.wantPresent {
				_, err := repo.GetOpPermission(context.Background(), uid)
				assert.NoError(t, err, "uid %s should exist", uid)
			}
		})
	}
}
