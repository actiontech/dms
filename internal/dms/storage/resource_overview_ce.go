//go:build !enterprise

package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
)

func (repo *ResourceOverviewRepo) GetResourceList(ctx context.Context, listOptions biz.ListResourceOverviewOption) ([]*biz.ResourceRow, error) {
	return nil, fmt.Errorf("resource overview is enterprise version functions")
}
