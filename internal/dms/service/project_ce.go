//go:build !enterprise

package service

import (
	"context"
	"errors"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

var errNotSupportProject = errors.New("project related functions are enterprise version functions")

func (d *DMSService) importProjects(ctx context.Context, uid string, req *dmsV1.ImportProjectsReq) error {
	return errNotSupportProject
}

func (d *DMSService) getImportProjectsTemplate(ctx context.Context, uid string) ([]byte, error) {
	return nil, errNotSupportProject
}

func (d *DMSService) getProjectTips(ctx context.Context, uid string, req *dmsV1.GetProjectTipsReq, err error) (*dmsV1.GetProjectTipsReply, error) {
	business, err := d.DBServiceUsecase.GetBusiness(ctx, req.ProjectUid)
	if err != nil {
		return nil, err
	}
	return &dmsV1.GetProjectTipsReply{
		Data: []*dmsV1.ProjectTips{
			{Business: business},
		},
	}, nil
}

func (d *DMSService) previewImportProjects(ctx context.Context, uid string, file string, err error) (*dmsV1.PreviewImportProjectsReply, error) {
	return nil, errNotSupportProject
}

func (d *DMSService) exportProjects(ctx context.Context, uid string, req *dmsV1.ExportProjectsReq) ([]byte, error) {
	return nil, errNotSupportProject
}

func (d *DMSService) getImportDBServicesTemplate(ctx context.Context, uid string) ([]byte, error) {
	return nil, errNotSupportProject
}

func (d *DMSService) importDBServicesOfProjectsCheck(ctx context.Context, userUid, fileContent string) (*dmsV1.ImportDBServicesCheckReply, []byte, error) {
	return nil, nil, errNotSupportProject
}

func (d *DMSService) importDBServicesOfProjects(ctx context.Context, req *dmsV1.ImportDBServicesOfProjectsReq, uid string) error {
	return errNotSupportProject
}

func (d *DMSService) dbServicesConnection(ctx context.Context, req *dmsV1.DBServiceConnectionReq, uid string) (*dmsV1.DBServicesConnectionReply, error) {
	return nil, errNotSupportProject
}
