//go:build !enterprise

package biz

import (
	"context"
)

func (cu *CloudbeaverUsecase) ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService, userId string) ([]*DBService, error) {
	return activeDBServices, nil
}
