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
			ProjectUid: n.UID,
			Name:       n.Name,
			Archived:   (n.Status == biz.ProjectStatusArchived),
			Desc:       n.Desc,
			CreateTime: strfmt.DateTime(n.CreateTime),
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

	}

	return &dmsCommonV1.ListProjectReply{
		Data: ret, Total: total,
	}, nil
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

	project, err := biz.NewProject(currentUserUid, req.Project.Name, req.Project.Desc, req.Project.Business)
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
	projects, err := convertImportReqToBiz(req, uid)
	if err != nil {
		return fmt.Errorf("convert req to biz failed: %w", err)
	}

	err = d.ProjectUsecase.ImportProjects(ctx, uid, projects)
	if err != nil {
		return fmt.Errorf("import projects failed: %w", err)
	}

	return nil
}

func convertImportReqToBiz(req *dmsV1.ImportProjectsReq, uid string) ([]*biz.Project, error) {
	projects := make([]*biz.Project, 0, len(req.Projects))
	for _, p := range req.Projects {
		project, err := biz.NewProject(uid, p.Name, p.Desc, p.Business)
		if err != nil {
			return nil, fmt.Errorf("create project failed: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (d *DMSService) GetImportProjectsTemplate(ctx context.Context, uid string) ([]byte, error) {
	content, err := d.ProjectUsecase.GetImportProjectsTemplate(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get import projects template failed: %w", err)
	}

	return content, nil
}

func (d *DMSService) GetProjectTips(ctx context.Context, uid string, req *dmsV1.GetProjectTipsReq) (reply *dmsV1.GetProjectTipsReply, err error) {
	projects, err := d.ProjectUsecase.GetProjectTips(ctx, uid, req.ProjectUid)
	if err != nil {
		return nil, fmt.Errorf("get project tips failed: %w", err)
	}

	resp := make([]*dmsV1.ProjectTips, len(projects))
	for i, p := range projects {
		resp[i] = &dmsV1.ProjectTips{
			ProjectUid:      p.UID,
			IsFixedBusiness: p.IsFixedBusiness,
			Business:        p.Business,
		}
	}

	return &dmsV1.GetProjectTipsReply{
		Data: resp,
	}, nil
}

func (d *DMSService) PreviewImportProjects(ctx context.Context, uid string, file string) (reply *dmsV1.PreviewImportProjectsReply, err error) {
	projects, err := d.ProjectUsecase.PreviewImportProjects(ctx, uid, file)
	if err != nil {
		return nil, fmt.Errorf("preview import projects failed: %w", err)
	}

	resp := make([]*dmsV1.PreviewImportProjects, len(projects))
	for i, p := range projects {
		resp[i] = &dmsV1.PreviewImportProjects{
			Name:     p.Name,
			Desc:     p.Desc,
			Business: p.Business,
		}
	}

	return &dmsV1.PreviewImportProjectsReply{
		Data: resp,
	}, nil
}
