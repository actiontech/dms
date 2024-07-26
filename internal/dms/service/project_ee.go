//go:build enterprise

package service

import (
	"context"
	"fmt"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	dmsCommonV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *DMSService) importProjects(ctx context.Context, uid string, req *dmsV1.ImportProjectsReq) error {
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
		project, err := biz.NewProject(uid, p.Name, p.Desc, true, p.Business)
		if err != nil {
			return nil, fmt.Errorf("create project failed: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (d *DMSService) getImportProjectsTemplate(ctx context.Context, uid string) ([]byte, error) {
	content, err := d.ProjectUsecase.GetImportProjectsTemplate(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get import projects template failed: %w", err)
	}

	return content, nil
}

func (d *DMSService) getProjectTips(ctx context.Context, uid string, req *dmsV1.GetProjectTipsReq, err error) (*dmsV1.GetProjectTipsReply, error) {
	projects, err := d.ProjectUsecase.GetProjectTips(ctx, uid, req.ProjectUid)
	if err != nil {
		return nil, fmt.Errorf("get project tips failed: %w", err)
	}

	resp := make([]*dmsV1.ProjectTips, len(projects))
	for i, p := range projects {
		businessList := make([]string, 0)
		for _, business := range p.Business {
			businessList = append(businessList, business.Name)
		}

		resp[i] = &dmsV1.ProjectTips{
			IsFixedBusiness: p.IsFixedBusiness,
			Business:        businessList,
		}

		business, err := d.DBServiceUsecase.GetBusiness(ctx, req.ProjectUid)
		if err != nil {
			d.log.Errorf("get business from db service failed: %v", err)
			continue
		}
		resp[i].Business = append(resp[i].Business, business...)

	}

	return &dmsV1.GetProjectTipsReply{
		Data: resp,
	}, nil
}

func (d *DMSService) previewImportProjects(ctx context.Context, uid string, file string, err error) (*dmsV1.PreviewImportProjectsReply, error) {
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

func (d *DMSService) exportProjects(ctx context.Context, uid string, req *dmsV1.ExportProjectsReq) ([]byte, error) {
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
		PageNumber:   0,
		LimitPerPage: 99999,
		OrderBy:      orderBy,
		FilterBy:     filterBy,
	}

	return d.ProjectUsecase.ExportProjects(ctx, uid, listOption)
}

func (d *DMSService) getImportDBServicesTemplate(ctx context.Context, uid string) ([]byte, error) {
	content, err := d.ProjectUsecase.GetImportDBServicesTemplate(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get import projects template failed: %w", err)
	}
	return content, nil
}

func (d *DMSService) importDBServicesOfProjectsCheck(ctx context.Context, userUid, fileContent string) (*dmsV1.ImportDBServicesCheckReply, []byte, error) {
	dbs, resultContent, err := d.DBServiceUsecase.ImportDBServicesOfProjectsCheck(ctx, userUid, fileContent)
	if err != nil {
		return nil, nil, err
	}
	if resultContent != nil {
		return nil, resultContent, nil
	}

	ret := d.convertBizDBServiceArgs2ImportDBService(dbs)

	return &dmsV1.ImportDBServicesCheckReply{Data: ret}, nil, nil
}

func (d *DMSService) importDBServicesOfProjects(ctx context.Context, req *dmsV1.ImportDBServicesOfProjectsReq, uid string) error {
	ret := d.convertImportDBService2BizDBService(req.DBServices)
	return d.DBServiceUsecase.ImportDBServicesOfProjects(ctx, ret, uid)
}

func (d *DMSService) dbServicesConnection(ctx context.Context, req *dmsV1.DBServiceConnectionReq, uid string) (*dmsV1.DBServicesConnectionReply, error) {
	items, err := d.DBServiceUsecase.DBServicesConnection(ctx, req.DBServices)
	if err != nil {
		return nil, err
	}
	return &dmsV1.DBServicesConnectionReply{Data: items}, nil
}
