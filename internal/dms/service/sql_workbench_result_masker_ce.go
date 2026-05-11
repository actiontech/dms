//go:build !dms

package service

import (
	"github.com/actiontech/dms/internal/dms/storage"
	sqlresultmasker "github.com/actiontech/dms/internal/sql_workbench/sqlresultmasker"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// NewSQLWorkbenchSQLResultMasker is a no-op in builds without the dms data-masking stack.
func NewSQLWorkbenchSQLResultMasker(_ utilLog.Logger, _ *storage.Storage) (sqlresultmasker.SQLResultMasker, error) {
	return nil, nil
}
