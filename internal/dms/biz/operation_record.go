package biz

import (
	"context"
	"time"

	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OperationRecordRepo interface {
	SaveOperationRecord(ctx context.Context, record *OperationRecord) error
	ListOperationRecords(ctx context.Context, opt *ListOperationRecordOption) ([]*OperationRecord, uint64, error)
	ExportOperationRecords(ctx context.Context, opt *ListOperationRecordOption) ([]*OperationRecord, error)
}

type OperationRecord struct {
	ID                   uint
	OperationTime        time.Time
	OperationUserName    string
	OperationReqIP       string
	OperationUserAgent   string
	OperationTypeName    string
	OperationAction      string
	OperationProjectName string
	OperationStatus      string
	OperationI18nContent i18nPkg.I18nStr
}

type ListOperationRecordOption struct {
	PageIndex                  uint32
	PageSize                   uint32
	FilterOperateTimeFrom      string
	FilterOperateTimeTo        string
	FilterOperateProjectName   *string
	FuzzySearchOperateUserName string
	FilterOperateTypeName      string
	FilterOperateAction        string
	// 权限相关字段
	CanViewGlobal              bool     // 是否有全局查看权限（admin/sys/全局权限）
	AccessibleProjectNames     []string // 可访问的项目名称列表（项目管理员）
}

type OperationRecordUsecase struct {
	repo OperationRecordRepo
	log  *utilLog.Helper
}

func NewOperationRecordUsecase(logger utilLog.Logger, repo OperationRecordRepo) *OperationRecordUsecase {
	return &OperationRecordUsecase{
		repo: repo,
		log:  utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.operationRecord")),
	}
}
