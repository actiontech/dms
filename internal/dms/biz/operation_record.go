package biz

import (
	"context"
	"strconv"
	"time"

	"github.com/actiontech/dms/pkg/dms-common/i18nPkg"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type OperationRecordRepo interface {
	SaveOperationRecord(ctx context.Context, record *OperationRecord) error
	ListOperationRecords(ctx context.Context, opt *ListOperationRecordOption) ([]*OperationRecord, uint64, error)
	ExportOperationRecords(ctx context.Context, opt *ListOperationRecordOption) ([]*OperationRecord, error)
	CleanOperationRecordOpTimeBefore(ctx context.Context, t time.Time) (rowsAffected int64, err error)
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
	CanViewGlobal          bool     // 是否有全局查看权限（admin/sys/全局权限）
	AccessibleProjectNames []string // 可访问的项目名称列表（项目管理员）
}

type OperationRecordUsecase struct {
	repo                  OperationRecordRepo
	systemVariableUsecase *SystemVariableUsecase
	log                   *utilLog.Helper
}

func NewOperationRecordUsecase(logger utilLog.Logger, repo OperationRecordRepo, svu *SystemVariableUsecase) *OperationRecordUsecase {
	return &OperationRecordUsecase{
		repo:                  repo,
		systemVariableUsecase: svu,
		log:                   utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.operationRecord")),
	}
}

func (u *OperationRecordUsecase) DoClean() {
	if u.systemVariableUsecase == nil {
		u.log.Errorf("failed to clean operation record when get systemVariableUsecase")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	variables, err := u.systemVariableUsecase.GetSystemVariables(ctx)
	if err != nil {
		u.log.Errorf("failed to clean operation record when get expired duration: %v", err)
		return
	}

	operationRecordExpiredHoursVar, ok := variables[SystemVariableOperationRecordExpiredHours]
	if !ok {
		u.log.Debugf("system variable %s not found, using default value", SystemVariableOperationRecordExpiredHours)
		operationRecordExpiredHoursVar = SystemVariable{
			Key:   SystemVariableOperationRecordExpiredHours,
			Value: strconv.Itoa(DefaultOperationRecordExpiredHours),
		}
	}

	operationRecordExpiredHours, err := strconv.Atoi(operationRecordExpiredHoursVar.Value)
	if err != nil {
		u.log.Errorf("failed to parse operation_record_expired_hours value: %v", err)
		return
	}

	if operationRecordExpiredHours <= 0 {
		u.log.Errorf("got OperationRecordExpiredHours: %d", operationRecordExpiredHours)
		return
	}

	cleanTime := time.Now().Add(time.Duration(-operationRecordExpiredHours) * time.Hour)
	rowsAffected, err := u.repo.CleanOperationRecordOpTimeBefore(ctx, cleanTime)
	if err != nil {
		u.log.Errorf("failed to clean operation record: %v", err)
		return
	}
	u.log.Infof("OperationRecord regular cleaned rows: %d operation time before: %s", rowsAffected, cleanTime.Format("2006-01-02 15:04:05"))
}

func (u *OperationRecordUsecase) GetLog() *utilLog.Helper {
	return u.log
}
