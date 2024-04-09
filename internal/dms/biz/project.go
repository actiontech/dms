package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type ProjectStatus string

const (
	ProjectStatusArchived ProjectStatus = "archived"
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusUnknown  ProjectStatus = "unknown"
)

type Project struct {
	Base

	UID             string
	Name            string
	Desc            string
	Business        []string
	IsFixedBusiness bool
	CreateUserUID   string
	CreateTime      time.Time
	Status          ProjectStatus
}

func NewProject(createUserUID, name, desc string, business []string) (*Project, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &Project{
		UID:           uid,
		Name:          name,
		Desc:          desc,
		Business:      business,
		Status:        ProjectStatusActive,
		CreateUserUID: createUserUID,
	}, nil
}

func initProjects() []*Project {
	return []*Project{
		{
			UID:           pkgConst.UIDOfProjectDefault,
			Name:          "default",
			Desc:          "default project",
			Status:        ProjectStatusActive,
			CreateUserUID: pkgConst.UIDOfUserAdmin,
		},
	}
}

type ProjectRepo interface {
	SaveProject(ctx context.Context, project *Project) error
	BatchSaveProjects(ctx context.Context, projects []*Project) error
	ListProjects(ctx context.Context, opt *ListProjectsOption, currentUserUID string) (projects []*Project, total int64, err error)
	GetProject(ctx context.Context, projectUid string) (*Project, error)
	GetProjectByName(ctx context.Context, projectName string) (*Project, error)
	UpdateProject(ctx context.Context, u *Project) error
	DelProject(ctx context.Context, projectUid string) error
}

type ProjectUsecase struct {
	tx                        TransactionGenerator
	repo                      ProjectRepo
	memberUsecase             *MemberUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	pluginUsecase             *PluginUsecase
	log                       *utilLog.Helper
}

func NewProjectUsecase(log utilLog.Logger, tx TransactionGenerator, repo ProjectRepo, memberUsecase *MemberUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase, pluginUsecase *PluginUsecase) *ProjectUsecase {
	return &ProjectUsecase{
		tx:                        tx,
		repo:                      repo,
		log:                       utilLog.NewHelper(log, utilLog.WithMessageKey("biz.project")),
		memberUsecase:             memberUsecase,
		pluginUsecase:             pluginUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
	}
}

type ListProjectsOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      ProjectField
	FilterBy     []pkgConst.FilterCondition
}

func (d *ProjectUsecase) ListProject(ctx context.Context, option *ListProjectsOption, currentUserUid string) (projects []*Project, total int64, err error) {
	// filter visible namespce space in advance
	// user can only view his belonging project,sys user can view all project
	if currentUserUid != pkgConst.UIDOfUserSys {
		projects, err := d.opPermissionVerifyUsecase.GetUserProject(ctx, currentUserUid)
		if err != nil {
			return nil, 0, err
		}
		canViewableId := make([]string, 0)
		for _, project := range projects {
			canViewableId = append(canViewableId, project.UID)
		}
		option.FilterBy = append(option.FilterBy, pkgConst.FilterCondition{
			Field:    string(ProjectFieldUID),
			Operator: pkgConst.FilterOperatorIn,
			Value:    canViewableId,
		})

	}

	projects, total, err = d.repo.ListProjects(ctx, option, currentUserUid)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects failed: %v", err)
	}

	return projects, total, nil
}

func (d *ProjectUsecase) ListProjectTips(ctx context.Context, currentUserUid string) ([]*Project, error) {
	return d.opPermissionVerifyUsecase.GetUserProject(ctx, currentUserUid)
}

func (d *ProjectUsecase) InitProjects(ctx context.Context) (err error) {
	for _, n := range initProjects() {
		_, err := d.GetProject(ctx, n.UID)
		// already exist
		if err == nil {
			continue
		}

		// error, return directly
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return fmt.Errorf("failed to get project: %v", err)
		}

		// not exist, then create it.
		err = d.repo.SaveProject(ctx, n)
		if err != nil {
			return fmt.Errorf("save projects failed: %v", err)
		}

		_, err = d.memberUsecase.AddUserToProjectAdminMember(ctx, pkgConst.UIDOfUserAdmin, n.UID)
		if err != nil {
			return fmt.Errorf("add admin to projects failed: %v", err)
		}
	}
	d.log.Debug("init project success")
	return nil
}

func (d *ProjectUsecase) GetProject(ctx context.Context, projectUid string) (*Project, error) {
	return d.repo.GetProject(ctx, projectUid)
}
