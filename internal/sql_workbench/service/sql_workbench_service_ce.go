//go:build !enterprise

package sql_workbench

import (
	"context"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (sqlWorkbenchService *SqlWorkbenchService) ResetDbServiceByAuth(ctx context.Context, activeDBServices []*biz.DBService, userId string) ([]*biz.DBService, error) {
	return activeDBServices, nil
}
