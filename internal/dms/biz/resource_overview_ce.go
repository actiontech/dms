//go:build !enterprise
package biz

import (
	"context"
	"errors"
)

var ErrNotSupportResourceOverview = errors.New("resource overview related functions are enterprise version functions")

func (uc *ResourceOverviewUsecase) GetResourceOverviewVisibility(ctx context.Context, currentUserUid string) (visibility ResourceOverviewVisibility, managedProjects []string, err error) {
	return visibility, managedProjects, ErrNotSupportResourceOverview
}
