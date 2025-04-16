//go:build !enterprise

package storage

import (
	"context"
	"github.com/actiontech/dms/internal/dms/biz"
)

func (repo *ResourceOverviewRepo) GetResourceOverviewTopology(ctx context.Context, listOptions biz.ListResourceOverviewOption) (*biz.ResourceTopology, error) {
	return nil, nil
}
