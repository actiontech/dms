//go:build !enterprise

package service

import (
	"context"
	"errors"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	dmsV2 "github.com/actiontech/dms/api/dms/service/v2"
)

var errNotSupportImportDBServices = errors.New("ImportDBServices related functions are enterprise version functions")
var errNotSupportGlobalDBServices = errors.New("GlobalDBServices related functions are enterprise version functions")

func (d *DMSService) importDBServicesOfOneProjectCheck(ctx context.Context, userUid, projectUid, fileContent string) (*dmsV2.ImportDBServicesCheckReply, []byte, error) {
	return nil, nil, errNotSupportImportDBServices
}

func (d *DMSService) importDBServicesOfOneProject(ctx context.Context, req *dmsV1.ImportDBServicesOfOneProjectReq, uid string) error {
	return errNotSupportImportDBServices
}

func (d *DMSService) listGlobalDBServices(ctx context.Context, req *dmsV1.ListGlobalDBServicesReq, currentUserUid string) (reply *dmsV1.ListGlobalDBServicesReply, err error) {
	return nil, errNotSupportGlobalDBServices
}

func (d *DMSService) listGlobalDBServicesTips(ctx context.Context, currentUserUid string) (reply *dmsV1.ListGlobalDBServicesTipsReply, err error) {
	return nil, errNotSupportGlobalDBServices
}
