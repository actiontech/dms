package biz

import (
	"context"
	"io"
	"strings"
	"testing"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type fakeDBServiceRepoForPassword struct {
	svc     *DBService
	updated *DBService
}

func (f *fakeDBServiceRepoForPassword) SaveDBServices(context.Context, []*DBService) error {
	return nil
}
func (f *fakeDBServiceRepoForPassword) GetDBServicesByIds(context.Context, []string) ([]*DBService, error) {
	return nil, nil
}
func (f *fakeDBServiceRepoForPassword) ListDBServices(context.Context, *ListDBServicesOption) ([]*DBService, int64, error) {
	return nil, 0, nil
}
func (f *fakeDBServiceRepoForPassword) DelDBService(context.Context, string) error { return nil }
func (f *fakeDBServiceRepoForPassword) GetDBService(_ context.Context, _ string) (*DBService, error) {
	return f.svc, nil
}
func (f *fakeDBServiceRepoForPassword) GetDBServices(context.Context, []pkgConst.FilterCondition) ([]*DBService, error) {
	return nil, nil
}
func (f *fakeDBServiceRepoForPassword) CheckDBServiceExist(context.Context, []string) (bool, error) {
	return true, nil
}
func (f *fakeDBServiceRepoForPassword) UpdateDBService(_ context.Context, dbService *DBService) error {
	f.updated = dbService
	return nil
}
func (f *fakeDBServiceRepoForPassword) CountDBService(context.Context) ([]DBTypeCount, error) {
	return nil, nil
}
func (f *fakeDBServiceRepoForPassword) GetBusinessByProjectUID(context.Context, string) ([]string, error) {
	return nil, nil
}
func (f *fakeDBServiceRepoForPassword) GetFieldDistinctValue(context.Context, DBServiceField, interface{}) error {
	return nil
}

type fakeProjectRepoForPassword struct {
	project *Project
}

func (f *fakeProjectRepoForPassword) SaveProject(context.Context, *Project) error { return nil }
func (f *fakeProjectRepoForPassword) BatchSaveProjects(context.Context, []*Project) error {
	return nil
}
func (f *fakeProjectRepoForPassword) ListProjects(context.Context, *ListProjectsOption, string) ([]*Project, int64, error) {
	return nil, 0, nil
}
func (f *fakeProjectRepoForPassword) GetProject(context.Context, string) (*Project, error) {
	return f.project, nil
}
func (f *fakeProjectRepoForPassword) GetProjectByName(context.Context, string) (*Project, error) {
	return f.project, nil
}
func (f *fakeProjectRepoForPassword) GetProjectByNames(context.Context, []string) ([]*Project, error) {
	return []*Project{f.project}, nil
}
func (f *fakeProjectRepoForPassword) UpdateProject(context.Context, *Project) error { return nil }
func (f *fakeProjectRepoForPassword) DelProject(context.Context, string) error       { return nil }
func (f *fakeProjectRepoForPassword) UpdateDBServiceBusiness(context.Context, string, string, string) error {
	return nil
}

type fakeEnvTagRepoForPassword struct {
	tag *EnvironmentTag
}

func (f *fakeEnvTagRepoForPassword) CreateEnvironmentTag(context.Context, *EnvironmentTag) error {
	return nil
}
func (f *fakeEnvTagRepoForPassword) UpdateEnvironmentTag(context.Context, string, string, string) error {
	return nil
}
func (f *fakeEnvTagRepoForPassword) DeleteEnvironmentTag(context.Context, string) error { return nil }
func (f *fakeEnvTagRepoForPassword) GetEnvironmentTagByName(context.Context, string, string) (bool, *EnvironmentTag, error) {
	return true, f.tag, nil
}
func (f *fakeEnvTagRepoForPassword) GetEnvironmentTagByUID(context.Context, string) (*EnvironmentTag, error) {
	return f.tag, nil
}
func (f *fakeEnvTagRepoForPassword) ListEnvironmentTags(context.Context, *ListEnvironmentTagsOption) ([]*EnvironmentTag, int64, error) {
	return nil, 0, nil
}

func newDBServiceUsecaseForEmptyPasswordTest(repo *fakeDBServiceRepoForPassword) *DBServiceUsecase {
	logger := utilLog.NewMyLogger(io.Discard)
	projectRepo := &fakeProjectRepoForPassword{
		project: &Project{UID: "700300", Status: ProjectStatusActive},
	}
	projectUC := &ProjectUsecase{
		repo: projectRepo,
		log:  utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.project.test")),
	}
	opUC := NewOpPermissionVerifyUsecase(logger, nil, &mockOpPermissionVerifyRepo{}, &mockUserRepo{users: map[string]*User{}})
	envUC := &EnvironmentTagUsecase{
		environmentTagRepo: &fakeEnvTagRepoForPassword{
			tag: &EnvironmentTag{UID: "env-1", Name: "prod"},
		},
		log: utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.env.test")),
	}
	pluginUC := &PluginUsecase{registeredPlugins: nil}
	return NewDBServiceUsecase(logger, repo, nil, pluginUC, opUC, projectUC, nil, envUC)
}

func TestUpdateDBServiceByArgs_EmptyPasswordAllowed(t *testing.T) {
	repo := &fakeDBServiceRepoForPassword{
		svc: &DBService{
			UID:        "ds-1",
			Name:       "gbase8a",
			DBType:     "GBase-8a",
			Host:       "10.186.16.126",
			Port:       "5258",
			User:       "root",
			Password:   "old-secret",
			ProjectUID: "700300",
		},
	}
	uc := newDBServiceUsecaseForEmptyPasswordTest(repo)
	empty := ""
	err := uc.UpdateDBServiceByArgs(context.Background(), "ds-1", &BizDBServiceArgs{
		DBType:            "GBase-8a",
		Host:              "10.186.16.126",
		Port:              "5258",
		User:              "root",
		Password:          &empty,
		EnvironmentTagUID: "env-1",
	}, pkgConst.UIDOfUserAdmin)
	if err != nil {
		t.Fatalf("expected Update with password=\"\" to succeed (not \"password can't be empty\"), got: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected UpdateDBService to be called")
	}
	if repo.updated.Password != "" {
		t.Fatalf("expected stored password to be empty string, got %q", repo.updated.Password)
	}
}

func TestUpdateDBServiceByArgs_MissingHostStillRejected(t *testing.T) {
	repo := &fakeDBServiceRepoForPassword{
		svc: &DBService{
			UID:        "ds-1",
			Name:       "gbase8a",
			DBType:     "GBase-8a",
			Host:       "10.186.16.126",
			Port:       "5258",
			User:       "root",
			Password:   "old-secret",
			ProjectUID: "700300",
		},
	}
	uc := newDBServiceUsecaseForEmptyPasswordTest(repo)
	empty := ""
	err := uc.UpdateDBServiceByArgs(context.Background(), "ds-1", &BizDBServiceArgs{
		DBType:            "GBase-8a",
		Host:              "",
		Port:              "5258",
		User:              "root",
		Password:          &empty,
		EnvironmentTagUID: "env-1",
	}, pkgConst.UIDOfUserAdmin)
	if err == nil {
		t.Fatal("expected missing Host to fail Update")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected host-related error, got: %v", err)
	}
	if strings.Contains(err.Error(), "password can't be empty") {
		t.Fatalf("must not fail on legacy empty-password check: %v", err)
	}
}
