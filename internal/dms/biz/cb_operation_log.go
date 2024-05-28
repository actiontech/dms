package biz

import (
	"context"
	"time"

	"github.com/actiontech/dms/internal/dms/pkg/constant"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type CbOperationLogType string

const (
	CbOperationLogTypeSql CbOperationLogType = "SQL"
)

// CbOperationLogRepo 定义操作日志的存储接口
type CbOperationLogRepo interface {
	GetCbOperationLogByID(ctx context.Context, uid string) (*CbOperationLog, error)
	SaveCbOperationLog(ctx context.Context, log *CbOperationLog) error
	UpdateCbOperationLog(ctx context.Context, log *CbOperationLog) error
	ListCbOperationLogs(ctx context.Context, opt *ListCbOperationLogOption) ([]*CbOperationLog, int64, error)
	CleanCbOperationLogOpTimeBefore(ctx context.Context, t time.Time) (int64, error)
}

// CbOperationLog 代表操作日志记录
type CbOperationLog struct {
	UID               string
	OpPersonUID       string
	OpTime            *time.Time
	DBServiceUID      string
	OpType            CbOperationLogType
	OpDetail          string
	OpSessionID       *string
	OpHost            string
	ProjectID         string
	AuditResults      []*AuditResult
	IsAuditPass       *bool
	ExecResult        string
	ExecTotalSec      int64
	ResultSetRowCount int64

	User      *User
	DbService *DBService
}

func (c CbOperationLog) GetOpTime() time.Time {
	if c.OpTime != nil {
		return *c.OpTime
	}
	return time.Time{}
}

func (c CbOperationLog) GetSessionID() string {
	if c.OpSessionID != nil {
		return *c.OpSessionID
	}
	return ""
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
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	repo                      CbOperationLogRepo
	dmsProxyTargetRepo        ProxyTargetRepo
	log                       *utilLog.Helper
}

// NewCbOperationLogUsecase 创建一个新的操作日志业务逻辑实例
func NewCbOperationLogUsecase(logger utilLog.Logger, repo CbOperationLogRepo, opPermissionVerifyUsecase *OpPermissionVerifyUsecase, proxyTargetRepo ProxyTargetRepo) *CbOperationLogUsecase {
	return &CbOperationLogUsecase{
		repo:                      repo,
		log:                       utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.cbOperationLog")),
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		dmsProxyTargetRepo:        proxyTargetRepo,
	}
}
