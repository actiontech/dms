//go:build !enterprise

package storage

import (
	"context"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (repo *ResourceOverviewRepo) GetResourceList(ctx context.Context, listOptions ListResourceOverviewOption) ([]*biz.ResourceRow, error) {
	return nil, nil
}
