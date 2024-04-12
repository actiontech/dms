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
	return nil, errNotSupportProject
}

func (d *DMSService) previewImportProjects(ctx context.Context, uid string, file string, err error) (*dmsV1.PreviewImportProjectsReply, error) {
	return nil, errNotSupportProject
}

func (d *DMSService) exportProjects(ctx context.Context, uid string, req *dmsV1.ExportProjectsReq) ([]byte, error) {
	return nil, errNotSupportProject
}
