package service

import (
	"context"
	"errors"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	dmsV2 "github.com/actiontech/dms/api/dms/service/v2"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
	dmsCommonV2 "github.com/actiontech/dms/pkg/dms-common/api/dms/v2"
	"github.com/go-openapi/strfmt"
)

func (d *DMSService) ListProjects(ctx context.Context, req *dmsCommonV2.ListProjectReq, currentUserUid string) (reply *dmsCommonV2.ListProjectReply, err error) {
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
	if req.FilterByProjectPriority != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.ProjectFieldPriority),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    dmsCommonV1.ToPriorityNum(req.FilterByProjectPriority),
		})
	}

	if len(req.FilterByProjectUids) > 0 {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.ProjectFieldUID),
			Operator: pkgConst.FilterOperatorIn,
			Value:    req.FilterByProjectUids,
		})
	}
	if req.FilterByDesc != "" {
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.ProjectFieldDesc),
			Operator: pkgConst.FilterOperatorContains,
			Value:    req.FilterByDesc,
		})
	}

	if req.FilterByBusinessTag != "" {
		businessTag, err := d.BusinessTagUsecase.GetBusinessTagByName(ctx, req.FilterByBusinessTag)
		if err != nil {
			d.log.Errorf("get business tag failed: %v", err)
			return nil, err
		}
		filterBy = append(filterBy, pkgConst.FilterCondition{
			Field:    string(biz.ProjectFieldBusinessTagUID),
			Operator: pkgConst.FilterOperatorEqual,
			Value:    businessTag.UID,
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

	err = d.BusinessTagUsecase.LoadBusinessTagForProjects(ctx, projects)
	if err != nil {
		return nil, err
	}
	ret := make([]*dmsCommonV2.ListProject, len(projects))
	for i, n := range projects {
		ret[i] = &dmsCommonV2.ListProject{
			ProjectUid: n.UID,
			Name:       n.Name,
			Archived:   (n.Status == biz.ProjectStatusArchived),
			Desc:       n.Desc,
			CreateTime: strfmt.DateTime(n.CreateTime),
			BusinessTag: &dmsCommonV2.BusinessTag{
				UID:  n.BusinessTag.UID,
				Name: n.BusinessTag.Name,
			},
			ProjectPriority: n.Priority,
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

	return &dmsCommonV2.ListProjectReply{
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

func (d *DMSService) AddProject(ctx context.Context, currentUserUid string, req *dmsV2.AddProjectReq) (reply *dmsV2.AddProjectReply, err error) {
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
		// check business tag is exist
		if req.Project.BusinessTag == nil || req.Project.BusinessTag.UID == "" {
			return nil, fmt.Errorf("business tag is empty")
		}
	}

	project, err := biz.NewProject(currentUserUid, req.Project.Name, req.Project.Desc, req.Project.ProjectPriority, req.Project.BusinessTag.UID)
	if err != nil {
		return nil, err
	}
	err = d.ProjectUsecase.CreateProject(ctx, project, currentUserUid)
	if err != nil {
		return nil, fmt.Errorf("create  project failed: %w", err)
	}

	return &dmsV2.AddProjectReply{
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

func (d *DMSService) UpdateProject(ctx context.Context, currentUserUid string, req *dmsV2.UpdateProjectReq) (err error) {
	err = d.ProjectUsecase.UpdateProject(ctx, currentUserUid, req.ProjectUid, req.Project.Desc, req.Project.ProjectPriority, req.Project.BusinessTag.UID)
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

func (d *DMSService) ImportProjects(ctx context.Context, uid string, req *dmsV2.ImportProjectsReq) error {
	return d.importProjects(ctx, uid, req)
}

func (d *DMSService) GetImportProjectsTemplate(ctx context.Context, uid string) ([]byte, error) {
	return d.getImportProjectsTemplate(ctx, uid)
}

func (d *DMSService) PreviewImportProjects(ctx context.Context, uid string, file string) (reply *dmsV2.PreviewImportProjectsReply, err error) {
	return d.previewImportProjects(ctx, uid, file)
}

func (d *DMSService) ExportProjects(ctx context.Context, uid string, req *dmsV1.ExportProjectsReq) ([]byte, error) {
	return d.exportProjects(ctx, uid, req)
}
func (d *DMSService) GetImportDBServicesTemplate(ctx context.Context, uid string) ([]byte, error) {
	return d.getImportDBServicesTemplate(ctx, uid)
}

func (d *DMSService) ImportDBServicesOfProjectsCheck(ctx context.Context, userUid, fileContent string) (*dmsV2.ImportDBServicesCheckReply, []byte, error) {
	return d.importDBServicesOfProjectsCheck(ctx, userUid, fileContent)
}

func (d *DMSService) ImportDBServicesOfProjects(ctx context.Context, req *dmsV2.ImportDBServicesOfProjectsReq, uid string) error {
	return d.importDBServicesOfProjects(ctx, req, uid)
}

func (d *DMSService) DBServicesConnection(ctx context.Context, req *dmsV1.DBServiceConnectionReq, uid string) (*dmsV1.DBServicesConnectionReply, error) {
	return d.dbServicesConnection(ctx, req, uid)
}
