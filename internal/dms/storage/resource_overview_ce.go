//go:build !enterprise

package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
)

func (repo *ResourceOverviewRepo) GetResourceList(ctx context.Context, listOptions biz.ListResourceOverviewOption) ([]*biz.ResourceRow, int64, error) {
	return nil, 0, fmt.Errorf("resource overview is enterprise version functions")
}
