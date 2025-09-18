package storage

import (
	"github.com/actiontech/dms/internal/dms/biz"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var _ biz.OdcRepo = (*OdcRepo)(nil)

type OdcRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOdcRepo(log utilLog.Logger, s *Storage) *OdcRepo {
	return &OdcRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.odc"))}
}










