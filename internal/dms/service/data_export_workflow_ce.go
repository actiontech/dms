//go:build !dms

package service

import (
	"context"
	"errors"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) DownloadOriginalDataExportWorkflow(ctx context.Context, req *dmsV1.DownloadOriginalDataExportWorkflowReq, currentUserUid string) (string, []byte, error) {
	return "", nil, errors.New("export original data is an enterprise version function")
}

func (d *DMSService) fillGetDataExportUnmaskingWorkflowSummary(_ context.Context, _ string, _ *dmsV1.GetDataExportWorkflow) {}
