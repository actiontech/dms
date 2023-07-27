package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"

	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.TransactionGenerator = (*TransactionGenerator)(nil)

func NewTXGenerator() *TransactionGenerator {
	return &TransactionGenerator{}
}

type TransactionGenerator struct{}

func (s TransactionGenerator) BeginTX(ctx context.Context) biz.RepoTX {
	return &RepoTX{ctx, nil}
}

type RepoTX struct {
	context.Context
	*gorm.DB
}

func (tx *RepoTX) Begin() {
	tx.DB = tx.DB.WithContext(tx.Context).Begin()
}

func (tx *RepoTX) Commit(log *utilLog.Helper) error {
	err := tx.DB.WithContext(tx.Context).Commit().Error
	if nil != err {
		return pkgErr.WrapStorageErr(log, fmt.Errorf("tx commit error: %v", err))
	}
	return nil
}

func (tx *RepoTX) RollbackWithError(log *utilLog.Helper, originalErr error) error {
	if tx.DB == nil {
		errMsg := fmt.Sprintf("tx rollback error: `tx db is nil`, original error: `%v`", originalErr)
		log.Errorf(errMsg)
		return pkgErr.WrapStorageErr(log, fmt.Errorf(errMsg))
	}
	err := tx.DB.WithContext(tx.Context).Rollback().Error
	if nil != err {
		errMsg := fmt.Sprintf("tx rollback error: `%v`, original error: `%v`", err, originalErr)
		log.Errorf(errMsg)
		return pkgErr.WrapStorageErr(log, fmt.Errorf(errMsg))
	}
	log.Errorf("tx rollback seccess, original error: %v", originalErr)
	return originalErr
}

func transaction(log *utilLog.Helper, ctx context.Context, db *gorm.DB, work func(tx *gorm.DB) error) (err error) {

	ctxWithTx, ok := ctx.(*RepoTX)

	if !ok {
		if err := db.WithContext(ctx).Transaction(work); err != nil {
			return pkgErr.WrapStorageErr(log, err)
		}

	} else {
		txCtx := ctxWithTx.Context

		if ctxWithTx.DB == nil {
			ctxWithTx.DB = db.WithContext(txCtx).Begin()
			if err := ctxWithTx.DB.Error; err != nil {
				return pkgErr.WrapStorageErr(log, err)
			}
		}

		if err := work(ctxWithTx.DB); err != nil {
			return pkgErr.WrapStorageErr(log, err)
		}
	}
	return nil
}
