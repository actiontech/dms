//go:build !enterprise

package biz

import "context"

func ResetDbServiceByAuth(ctx context.Context, activeDBServices []*DBService) ([]*DBService, error) {
	return activeDBServices, nil
}
