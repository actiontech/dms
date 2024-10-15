package service

import (
	"context"
	"errors"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	"github.com/go-openapi/strfmt"
)

func (d *DMSService) ListProjects(ctx context.Context, req *dmsCommonV1.ListProjectReq, currentUserUid string) (reply *dmsCommonV1.ListProjectReply, err error) {
	var orderBy biz.ProjectField
	switch req.OrderBy {
	case dmsCommonV1.ProjectOrderByName:
		orderBy = biz.ProjectFieldName
	default:
		orderBy = biz.ProjectFieldName
	}

	filterBy := make([]pkgConst.FilterCondition, 0)
	if req.FilterByName != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.ProjectFieldName),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByName,
		})
	}
	if req.FilterByUID != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.ProjectFieldUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    req.FilterByUID,
		})
	}

	listOption := &biz.ListProjectsOption{
		PageNumber:   req.PageIndex,
		LimitPerPage: req.PageSize,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	projects, total, err := d.ProjectUsecase.ListProject(ctx, listOption, currentUserUid)
	if nil != err {
		return nil, err
	}

	ret := make([]*dmsCommonV1.ListProject, len(projects))
	for i, n := range projects {
		ret[i] = &dmsCommonV1.ListProject{
			ProjectUid:      n.UID,
			Name:            n.Name,
			Archived:        (n.Status == biz.ProjectStatusArchived),
			Desc:            n.Desc,
			IsFixedBusiness: n.IsFixedBusiness,
			CreateTime:      strfmt.DateTime(n.CreateTime),
		}
		user, err := d.UserUsecase.GetUser(ctx, n.CreateUserUID)
		if err != nil {
			d.log.Errorf("get user error: %v", err)
			continue
		}
		ret[i].CreateUser = dmsCommonV1.UidWithName{
			Uid:  n.UID,
			Name: user.Name,
		}

		business, err := d.DBServiceUsecase.GetBusiness(ctx, n.UID)
		if err != nil {
			d.log.Errorf("get business error: %v", err)
			continue
		}

		for _, b := range n.Business {
			isBusinessInUse := isStrInSlice(b.Name, business)
			ret[i].Business = append(ret[i].Business, dmsCommonV1.Business{
				Id:     b.Uid,
				Name:   b.Name,
				IsUsed: isBusinessInUse,
			})
		}
	}

	return &dmsCommonV1.ListProjectReply{
		Data: ret, Total: total,
	}, nil
}

func isStrInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (d *DMSService) AddProject(ctx context.Context, currentUserUid string, req *dmsV1.AddProjectReq) (reply *dmsV1.AddProjectReply, err error) {
	// check
	{
		// check current user has enough permission
		if canCreateProject, err := d.OpPermissionVerifyUsecase.CanCreateProject(ctx, currentUserUid); err != nil {
			return nil, err
		} else if !canCreateProject {
			return nil, fmt.Errorf("current user can't create project")
		}

		// check project is exist
		_, err := d.ProjectUsecase.GetProjectByName(ctx, req.Project.Name)
		if err == nil {
			return nil, fmt.Errorf("project %v is exist", req.Project.Name)
		}
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, fmt.Errorf("failed to get project by name: %w", err)
		}
	}

	project, err := biz.NewProject(currentUserUid, req.Project.Name, req.Project.Desc, req.Project.ProjectPriority, req.Project.IsFixedBusiness, req.Project.Business)
	if err != nil {
		return nil, err
	}
	err = d.ProjectUsecase.CreateProject(ctx, project, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("create  project failed: %w", err)
	}

	return &dmsV1.AddProjectReply{
		Data: struct {
			//  project UID
			Uid string `json:"uid"`
		}{Uid: project.UID},
	}, nil
}

func (d *DMSService) DeleteProject(ctx context.Context, currentUserUid string, req *dmsV1.DelProjectReq) (err error) {
	err = d.ProjectUsecase.DeleteProject(ctx, currentUserUid, req.ProjectUid)
	if err != nil {
		return fmt.Errorf("delete  project failed: %w", err)
	}

	return nil
}

func (d *DMSService) UpdateProjectDesc(ctx context.Context, currentUserUid string, req *dmsV1.UpdateProjectReq) (err error) {
	err = d.ProjectUsecase.UpdateProjectDesc(ctx, currentUserUid, req.ProjectUid, req.Project.Desc)
	if err != nil {
		return fmt.Errorf("update project failed: %w", err)
	}

	return nil
}

func (d *DMSService) UpdateProject(ctx context.Context, currentUserUid string, req *dmsV1.UpdateProjectReq) (err error) {
	err = d.ProjectUsecase.UpdateProject(ctx, currentUserUid, req.ProjectUid, req.Project.Desc, req.Project.IsFixedBusiness, req.Project.Business)
	if err != nil {
		return fmt.Errorf("update project failed: %w", err)
	}

	return nil
}

func (d *DMSService) ArchivedProject(ctx context.Context, currentUserUid string, req *dmsV1.ArchiveProjectReq) (err error) {
	err = d.ProjectUsecase.ArchivedProject(ctx, currentUserUid, req.ProjectUid, true)
	if err != nil {
		return fmt.Errorf("archived project failed: %w", err)
	}

	return nil
}

func (d *DMSService) UnarchiveProject(ctx context.Context, currentUserUid string, req *dmsV1.UnarchiveProjectReq) (err error) {
	err = d.ProjectUsecase.ArchivedProject(ctx, currentUserUid, req.ProjectUid, false)
	if err != nil {
		return fmt.Errorf("unarchive project failed: %w", err)
	}

	return nil
}

func (d *DMSService) ImportProjects(ctx context.Context, uid string, req *dmsV1.ImportProjectsReq) error {
	return d.importProjects(ctx, uid, req)
}

func (d *DMSService) GetImportProjectsTemplate(ctx context.Context, uid string) ([]byte, error) {
	return d.getImportProjectsTemplate(ctx, uid)
}

func (d *DMSService) GetProjectTips(ctx context.Context, uid string, req *dmsV1.GetProjectTipsReq) (reply *dmsV1.GetProjectTipsReply, err error) {
	return d.getProjectTips(ctx, uid, req, err)
}

func (d *DMSService) PreviewImportProjects(ctx context.Context, uid string, file string) (reply *dmsV1.PreviewImportProjectsReply, err error) {
	return d.previewImportProjects(ctx, uid, file, err)
}

func (d *DMSService) ExportProjects(ctx context.Context, uid string, req *dmsV1.ExportProjectsReq) ([]byte, error) {
	return d.exportProjects(ctx, uid, req)
}
func (d *DMSService) GetImportDBServicesTemplate(ctx context.Context, uid string) ([]byte, error) {
	return d.getImportDBServicesTemplate(ctx, uid)
}

func (d *DMSService) ImportDBServicesOfProjectsCheck(ctx context.Context, userUid, fileContent string) (*dmsV1.ImportDBServicesCheckReply, []byte, error) {
	return d.importDBServicesOfProjectsCheck(ctx, userUid, fileContent)
}

func (d *DMSService) ImportDBServicesOfProjects(ctx context.Context, req *dmsV1.ImportDBServicesOfProjectsReq, uid string) error {
	return d.importDBServicesOfProjects(ctx, req, uid)
}

func (d *DMSService) DBServicesConnection(ctx context.Context, req *dmsV1.DBServiceConnectionReq, uid string) (*dmsV1.DBServicesConnectionReply, error) {
	return d.dbServicesConnection(ctx, req, uid)
}
