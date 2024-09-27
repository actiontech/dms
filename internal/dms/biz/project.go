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
	IsFixedBusiness bool
	Business        []Business
	CreateUserUID   string
	CreateTime      time.Time
	Status          ProjectStatus
}

type Business struct {
	Uid  string
	Name string
}

type PreviewProject struct {
	Name     string
	Desc     string
	Business []string
}

func NewProject(createUserUID, name, desc string, isFixedBusiness bool, business []string) (*Project, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}

	businessList := make([]Business, 0)
	for _, b := range business {
		uid, err = pkgRand.GenStrUid()
		if err != nil {
			return nil, err
		}

		businessList = append(businessList, Business{
			Uid:  uid,
			Name: b,
		})
	}

	return &Project{
		UID:             uid,
		Name:            name,
		Desc:            desc,
		Business:        businessList,
		Status:          ProjectStatusActive,
		IsFixedBusiness: isFixedBusiness,
		CreateUserUID:   createUserUID,
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
	UpdateDBServiceBusiness(ctx context.Context, projectUid string, originBusiness string, descBusiness string) error
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
	canViewGlobal, err := d.opPermissionVerifyUsecase.CanViewGlobal(ctx, currentUserUid)
	if err != nil {
		return nil, 0, err
	}

	// filter visible namespce space in advance
	// user can only view his belonging project,sys user can view all project
	if currentUserUid != pkgConst.UIDOfUserSys && !canViewGlobal {
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

func (d *ProjectUsecase) UpdateDBServiceBusiness(ctx context.Context, currentUserUid, projectUid string, originBusiness, descBusiness string) error {
	// 检查项目是否归档/删除
	if err := d.isProjectActive(ctx, projectUid); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}

	// 检查当前用户有项目管理员权限
	if canOpProject, err := d.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	err := d.repo.UpdateDBServiceBusiness(ctx, projectUid, originBusiness, descBusiness)
	if err != nil {
		return fmt.Errorf("update db service business failed: %v", err)
	}

	return nil
}
