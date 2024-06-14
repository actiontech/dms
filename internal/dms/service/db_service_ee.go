//go:build enterprise

package service

import (
	"context"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) importDBServicesOfOneProjectCheck(ctx context.Context, userUid, projectUid, fileContent string) (*dmsV1.ImportDBServicesCheckReply, []byte, error) {
	dbs, resultContent, err := d.DBServiceUsecase.ImportDBServicesOfOneProjectCheck(ctx, userUid, projectUid, fileContent)
	if err != nil {
		return nil, nil, err
	}
	if resultContent != nil {
		return nil, resultContent, nil
	}

	ret := d.convertBizDBServiceArgs2ImportDBService(dbs)

	return &dmsV1.ImportDBServicesCheckReply{Data: ret}, nil, nil
}

func (d *DMSService) importDBServicesOfOneProject(ctx context.Context, req *dmsV1.ImportDBServicesOfOneProjectReq, uid string) error {
	ret := d.convertImportDBService2BizDBService(req.DBServices)
	return d.DBServiceUsecase.ImportDBServicesOfOneProject(ctx, ret, uid, req.ProjectUid)
}
