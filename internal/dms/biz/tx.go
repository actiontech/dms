package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type TransactionGenerator interface {
	BeginTX(ctx context.Context) RepoTX
}

type RepoTX interface {
	context.Context
	Commit(log *utilLog.Helper) error
	// RollbackWithError wraps original error into rollback error if there is one
	RollbackWithError(log *utilLog.Helper, originalErr error) error
}
