//go:build !dms

package storage

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type SensitiveDataDiscoveryTaskRepo struct{}

func NewSensitiveDataDiscoveryTaskRepo(_ utilLog.Logger, _ *Storage) *SensitiveDataDiscoveryTaskRepo {
	return &SensitiveDataDiscoveryTaskRepo{}
}

func (r *SensitiveDataDiscoveryTaskRepo) CheckMaskingTaskExist(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (r *SensitiveDataDiscoveryTaskRepo) ListMaskingTaskStatus(_ context.Context, _ []string) (map[string]bool, error) {
	return map[string]bool{}, nil
}
