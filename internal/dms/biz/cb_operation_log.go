package biz

import (
	"context"
	"time"

	"github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// CbOperationLogRepo 定义操作日志的存储接口
type CbOperationLogRepo interface {
	SaveCbOperationLog(ctx context.Context, log *CbOperationLog) error
	UpdateCbOperationLog(ctx context.Context, log *CbOperationLog) error
}

// CbOperationLog 代表操作日志记录
type CbOperationLog struct {
	UID               string
	OpPersonUID       string
	OpTime            *time.Time
	DBServiceUID      string
	OpDetail          string
	OpSessionID       *string
	OpHost            string
	ProjectID         string
	AuditResults      []*AuditResult
	IsAuditPass       *bool
	ExecResult        string
	ExecTotalSec      int64
	ResultSetRowCount int64
}

// ListCbOperationLogOption 用于查询操作日志的选项
type ListCbOperationLogOption struct {
	PageNumber   uint32
	LimitPerPage uint32
	OrderBy      string
	FilterBy     []constant.FilterCondition
}

// CbOperationLogUsecase 定义操作日志的业务逻辑
type CbOperationLogUsecase struct {
	repo CbOperationLogRepo
	log  *utilLog.Helper
}

// NewCbOperationLogUsecase 创建一个新的操作日志业务逻辑实例
func NewCbOperationLogUsecase(logger utilLog.Logger, repo CbOperationLogRepo) *CbOperationLogUsecase {
	return &CbOperationLogUsecase{
		repo: repo,
		log:  utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.cbOperationLog")),
	}
}
