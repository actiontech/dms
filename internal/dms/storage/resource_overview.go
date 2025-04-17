package storage

import (
	"github.com/actiontech/dms/internal/dms/biz"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var _ biz.ResourceOverviewRepo = (*ResourceOverviewRepo)(nil)

type ResourceOverviewRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewResourceOverviewRepo(log utilLog.Logger, s *Storage) *ResourceOverviewRepo {
	return &ResourceOverviewRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.resource_overview"))}
}
