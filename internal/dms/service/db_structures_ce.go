//go:build !dms

package service

import (
	"context"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
)

func (d *DMSService) ListTableColumns(ctx context.Context, req *dmsV1.ListTableColumnsReq) (*dmsV1.ListTableColumnsReply, error) {
	return nil, errNotSupportDataMasking
}

