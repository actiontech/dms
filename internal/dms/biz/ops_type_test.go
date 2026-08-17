package biz

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/stretchr/testify/assert"
)

// fakeOpsTypeRepo 内存实现，供 biz 规则自验。
type fakeOpsTypeRepo struct {
	byUID  map[string]*OpsType
	create []*OpsType
}

func newFakeOpsTypeRepo() *fakeOpsTypeRepo {
	return &fakeOpsTypeRepo{byUID: map[string]*OpsType{}}
}

func (f *fakeOpsTypeRepo) CreateOpsType(_ context.Context, opsType *OpsType) error {
	cp := *opsType
	f.byUID[opsType.UID] = &cp
	f.create = append(f.create, &cp)
	return nil
}

func (f *fakeOpsTypeRepo) UpdateOpsType(_ context.Context, opsTypeUID, opsTypeName string) error {
	ot, ok := f.byUID[opsTypeUID]
	if !ok {
		return fmt.Errorf("ops type not found")
	}
	ot.Name = opsTypeName
	return nil
}

func (f *fakeOpsTypeRepo) DeleteOpsType(_ context.Context, opsTypeUID string) error {
	if _, ok := f.byUID[opsTypeUID]; !ok {
		return fmt.Errorf("ops type not found")
	}
	delete(f.byUID, opsTypeUID)
	return nil
}

func (f *fakeOpsTypeRepo) GetOpsTypeByName(_ context.Context, projectUID, name string) (bool, *OpsType, error) {
	for _, ot := range f.byUID {
		if ot.ProjectUID == projectUID && ot.Name == name {
			cp := *ot
			return true, &cp, nil
		}
	}
	return false, nil, nil
}

func (f *fakeOpsTypeRepo) GetOpsTypeByUID(_ context.Context, uid string) (*OpsType, error) {
	ot, ok := f.byUID[uid]
	if !ok {
		return nil, fmt.Errorf("ops type not found")
	}
	cp := *ot
	return &cp, nil
}

func (f *fakeOpsTypeRepo) ListOpsTypes(_ context.Context, options *ListOpsTypesOption) ([]*OpsType, int64, error) {
	out := make([]*OpsType, 0)
	for _, ot := range f.byUID {
		if ot.ProjectUID == options.ProjectUID {
			cp := *ot
			out = append(out, &cp)
		}
	}
	return out, int64(len(out)), nil
}

func newOpsTypeUsecaseForTest(repo OpsTypeRepo, projectUID string, opRepo OpPermissionVerifyRepo, users map[string]*User) *OpsTypeUsecase {
	logger := utilLog.NewMyLogger(io.Discard)
	projectRepo := &fakeProjectRepoForPassword{
		project: &Project{UID: projectUID, Status: ProjectStatusActive},
	}
	projectUC := &ProjectUsecase{
		repo: projectRepo,
		log:  utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.project.test")),
	}
	if users == nil {
		users = map[string]*User{}
	}
	opUC := NewOpPermissionVerifyUsecase(logger, nil, opRepo, &mockUserRepo{users: users})
	return NewOpsTypeUsecase(repo, logger, projectUC, opUC)
}

func TestOpsTypeUsecase_CreateUpdatePermissionAndDuplicate(t *testing.T) {
	const projectUID = "2088162858587131904" // ops-type-proj-a
	const projectAdminUID = "proj-admin-1"
	const memberUID = "proj-member-1"

	repo := newFakeOpsTypeRepo()
	opRepo := &mockOpPermissionVerifyRepo{
		projectPermissions: map[string]map[string]map[string]bool{
			projectAdminUID: {
				projectUID: {pkgConst.UIDOfOpPermissionProjectAdmin: true},
			},
			memberUID: {
				projectUID: {},
			},
		},
		globalPermissions: map[string][]*OpPermission{},
	}
	users := map[string]*User{
		pkgConst.UIDOfUserAdmin: {UID: pkgConst.UIDOfUserAdmin, BusinessWritePermission: true},
		projectAdminUID:         {UID: projectAdminUID},
		memberUID:               {UID: memberUID},
	}
	uc := newOpsTypeUsecaseForTest(repo, projectUID, opRepo, users)
	ctx := context.Background()

	t.Run("project_admin_create_ok", func(t *testing.T) {
		err := uc.CreateOpsType(ctx, projectUID, projectAdminUID, "  数据修改  ")
		assert.NoError(t, err)
		exist, got, err := uc.GetOpsTypeByName(ctx, projectUID, "数据修改")
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, "数据修改", got.Name)
		assert.Equal(t, projectUID, got.ProjectUID)
	})

	t.Run("duplicate_name_rejected", func(t *testing.T) {
		err := uc.CreateOpsType(ctx, projectUID, projectAdminUID, "数据修改")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "already exists"), "err=%v", err)
	})

	t.Run("duplicate_after_trim_rejected", func(t *testing.T) {
		err := uc.CreateOpsType(ctx, projectUID, projectAdminUID, "\t数据修改\n")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "already exists"), "err=%v", err)
	})

	t.Run("member_create_rejected", func(t *testing.T) {
		err := uc.CreateOpsType(ctx, projectUID, memberUID, "服务维护")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not project admin"), "err=%v", err)
	})

	t.Run("platform_admin_create_ok", func(t *testing.T) {
		err := uc.CreateOpsType(ctx, projectUID, pkgConst.UIDOfUserAdmin, "服务维护")
		assert.NoError(t, err)
	})

	t.Run("project_admin_rename_ok", func(t *testing.T) {
		exist, got, err := uc.GetOpsTypeByName(ctx, projectUID, "服务维护")
		assert.NoError(t, err)
		assert.True(t, exist)
		err = uc.UpdateOpsType(ctx, projectUID, projectAdminUID, got.UID, "  配置修改  ")
		assert.NoError(t, err)
		exist, renamed, err := uc.GetOpsTypeByName(ctx, projectUID, "配置修改")
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, got.UID, renamed.UID)
	})

	t.Run("rename_to_existing_rejected", func(t *testing.T) {
		exist, got, err := uc.GetOpsTypeByName(ctx, projectUID, "配置修改")
		assert.NoError(t, err)
		assert.True(t, exist)
		err = uc.UpdateOpsType(ctx, projectUID, projectAdminUID, got.UID, "数据修改")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "already exists"), "err=%v", err)
	})

	t.Run("member_update_rejected", func(t *testing.T) {
		exist, got, err := uc.GetOpsTypeByName(ctx, projectUID, "配置修改")
		assert.NoError(t, err)
		assert.True(t, exist)
		err = uc.UpdateOpsType(ctx, projectUID, memberUID, got.UID, "其它名称")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not project admin"), "err=%v", err)
	})

	t.Run("member_delete_rejected", func(t *testing.T) {
		exist, got, err := uc.GetOpsTypeByName(ctx, projectUID, "配置修改")
		assert.NoError(t, err)
		assert.True(t, exist)
		err = uc.DeleteOpsType(ctx, projectUID, memberUID, got.UID)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not project admin"), "err=%v", err)
	})

	t.Run("list_readable_without_write_perm", func(t *testing.T) {
		list, count, err := uc.ListOpsTypes(ctx, &ListOpsTypesOption{ProjectUID: projectUID, Limit: 100, Offset: 0})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
		assert.NotEmpty(t, list)
	})

	t.Run("project_admin_delete_ok", func(t *testing.T) {
		exist, got, err := uc.GetOpsTypeByName(ctx, projectUID, "配置修改")
		assert.NoError(t, err)
		assert.True(t, exist)
		err = uc.DeleteOpsType(ctx, projectUID, projectAdminUID, got.UID)
		assert.NoError(t, err)
		exist, _, err = uc.GetOpsTypeByName(ctx, projectUID, "配置修改")
		assert.NoError(t, err)
		assert.False(t, exist)
	})
}

func TestOpsTypeUsecase_InitAndListSeed(t *testing.T) {
	const projectUID = "seed-proj-uid"
	const projectAdminUID = "seed-admin-1"

	expected := []string{"数据修改", "数据提取", "服务维护", "配置修改"}

	opRepo := &mockOpPermissionVerifyRepo{
		projectPermissions: map[string]map[string]map[string]bool{
			projectAdminUID: {
				projectUID: {pkgConst.UIDOfOpPermissionProjectAdmin: true},
			},
		},
		globalPermissions: map[string][]*OpPermission{},
	}
	users := map[string]*User{
		projectAdminUID: {UID: projectAdminUID},
	}
	ctx := context.Background()

	t.Run("init_default_on_create_project", func(t *testing.T) {
		repo := newFakeOpsTypeRepo()
		uc := newOpsTypeUsecaseForTest(repo, projectUID, opRepo, users)
		err := uc.InitDefaultOpsTypes(ctx, projectUID, projectAdminUID)
		assert.NoError(t, err)
		list, count, err := repo.ListOpsTypes(ctx, &ListOpsTypesOption{ProjectUID: projectUID, Limit: 100, Offset: 0})
		assert.NoError(t, err)
		assert.Equal(t, int64(4), count)
		gotNames := make([]string, 0, len(list))
		for _, ot := range list {
			gotNames = append(gotNames, ot.Name)
		}
		assert.ElementsMatch(t, expected, gotNames)
	})

	t.Run("list_empty_seeds_then_idempotent", func(t *testing.T) {
		repo := newFakeOpsTypeRepo()
		uc := newOpsTypeUsecaseForTest(repo, projectUID, opRepo, users)

		list1, count1, err := uc.ListOpsTypes(ctx, &ListOpsTypesOption{ProjectUID: projectUID, Limit: 100, Offset: 0})
		assert.NoError(t, err)
		assert.Equal(t, int64(4), count1)
		gotNames := make([]string, 0, len(list1))
		for _, ot := range list1 {
			gotNames = append(gotNames, ot.Name)
		}
		assert.ElementsMatch(t, expected, gotNames)

		list2, count2, err := uc.ListOpsTypes(ctx, &ListOpsTypesOption{ProjectUID: projectUID, Limit: 100, Offset: 0})
		assert.NoError(t, err)
		assert.Equal(t, int64(4), count2)
		assert.Equal(t, count1, count2)
		assert.Len(t, list2, 4)
		assert.Equal(t, 4, len(repo.create), "second list must not create again")
	})

	t.Run("ensure_tolerates_concurrent_duplicate", func(t *testing.T) {
		repo := &racingOpsTypeRepo{fakeOpsTypeRepo: newFakeOpsTypeRepo()}
		uc := newOpsTypeUsecaseForTest(repo, projectUID, opRepo, users)
		// 预先写入一项，并让 Create 对同名返回 duplicate（模拟并发赢家已插入）。
		pre, err := uc.newOpsType(projectUID, "数据修改")
		assert.NoError(t, err)
		assert.NoError(t, repo.fakeOpsTypeRepo.CreateOpsType(ctx, pre))
		repo.failNextCreateFor = "数据提取"

		err = uc.ensureDefaultOpsTypes(ctx, projectUID)
		assert.NoError(t, err)
		_, count, err := repo.ListOpsTypes(ctx, &ListOpsTypesOption{ProjectUID: projectUID, Limit: 100, Offset: 0})
		assert.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})
}

// racingOpsTypeRepo 模拟唯一索引竞态：指定名称首次 Create 失败但随后 Get 可见。
type racingOpsTypeRepo struct {
	*fakeOpsTypeRepo
	failNextCreateFor string
}

func (r *racingOpsTypeRepo) CreateOpsType(ctx context.Context, opsType *OpsType) error {
	if r.failNextCreateFor != "" && opsType.Name == r.failNextCreateFor {
		name := r.failNextCreateFor
		r.failNextCreateFor = ""
		// 另一并发已写入
		cp := *opsType
		cp.UID = "race-winner-" + name
		r.byUID[cp.UID] = &cp
		return fmt.Errorf("Error 1062: Duplicate entry for key project_uid_ops_type_name")
	}
	return r.fakeOpsTypeRepo.CreateOpsType(ctx, opsType)
}
