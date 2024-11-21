//go:build enterprise

package biz

import (
	"context"
	"time"

	"github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	"github.com/actiontech/dms/internal/pkg/cloudbeaver"
	v1Base "github.com/actiontech/dms/pkg/dms-common/api/base/v1"
	pkgHttp "github.com/actiontech/dms/pkg/dms-common/pkg/http"
)

func (cu *CbOperationLogUsecase) GetCbOperationLogByID(ctx context.Context, uid string) (*CbOperationLog, error) {
	operationLog, err := cu.repo.GetCbOperationLogByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return operationLog, nil
}

func (u *CbOperationLogUsecase) SaveCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return u.repo.SaveCbOperationLog(ctx, log)
}

func (u *CbOperationLogUsecase) UpdateCbOperationLog(ctx context.Context, log *CbOperationLog) error {
	return u.repo.UpdateCbOperationLog(ctx, log)
}

func (u *CbOperationLogUsecase) ListCbOperationLog(ctx context.Context, option *ListCbOperationLogOption, currentUid string, filterPersonID string, projectUid string) ([]*CbOperationLog, int64, error) {
	// 只有管理员和拥有全局管理/浏览权限的用户可以查看所有操作日志, 其他用户只能查看自己的操作日志
	if currentUid != pkgConst.UIDOfUserSys {
		if canViewProject, err := u.opPermissionVerifyUsecase.CanViewProject(ctx, currentUid, projectUid); err != nil {
			return nil, 0, err
		} else if canViewProject {
			// do nothing,skip to next,because admin or global view user can view all operation logs
		} else if filterPersonID != "" && currentUid != filterPersonID {
			// 查询操作用户参数不是当前用户，无权看其他用户的操作记录
			return nil, 0, nil
		} else if filterPersonID == "" {
			// 查询操作用户参数为空时只能看自己的
			option.FilterBy = append(option.FilterBy, constant.FilterCondition{
				Field:    string(CbOperationLogFieldOpPersonUID),
				Operator: constant.FilterOperatorEqual,
				Value:    currentUid,
			})
		}
	}

	operationLogs, count, err := u.repo.ListCbOperationLogs(ctx, option)
	if err != nil {
		return nil, 0, err
	}

	return operationLogs, count, nil
}

func (u *CbOperationLogUsecase) DoClean() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// request SQLE to get cb_operation_logs_expired_hours
	target, err := u.dmsProxyTargetRepo.GetProxyTargetByName(ctx, cloudbeaver.SQLEProxyName)
	if err != nil {
		u.log.Error("CbOperationLogUsecase DoClean GetProxyTargetByName err:", err)
		return
	}
	url := target.URL.String() + "/v1/configurations/system_variables"

	header := map[string]string{
		"Authorization": pkgHttp.DefaultDMSToken,
	}

	reply := &struct {
		v1Base.GenericResp
		Data struct {
			CbOperationLogsExpiredHours int `json:"cb_operation_logs_expired_hours"`
		} `json:"data"`
	}{}
	err = pkgHttp.Get(ctx, url, header, nil, reply)
	if err != nil {
		u.log.Errorf("failed to clean CB operation log when get expired duration: %v", err)
		return
	} else if reply.Code != 0 {
		u.log.Errorf("failed to clean CB operation log, sqle reply code: %d", reply.Code)
		return
	} else if reply.Data.CbOperationLogsExpiredHours <= 0 {
		u.log.Debugf("got CbOperationLogsExpiredHours: %d", reply.Data.CbOperationLogsExpiredHours)
		return
	}

	cleanTime := time.Now().Add(time.Duration(-reply.Data.CbOperationLogsExpiredHours) * time.Hour)
	rowsAffected, err := u.repo.CleanCbOperationLogOpTimeBefore(ctx, cleanTime)
	if err != nil {
		u.log.Errorf("failed to clean CB operation log: %v", err)
		return
	}
	u.log.Infof("CbOperationLog regular cleaned rows: %d operation time before: %s", rowsAffected, cleanTime.Format("2006-01-02 15:04:05"))
	return
}

func (u *CbOperationLogUsecase) CountOperationLogs(ctx context.Context, option *ListCbOperationLogOption, currentUid string, filterPersonID string, projectUid string) (int64, error) {
	// 只有管理员和拥有全局管理/浏览权限的用户可以查看所有操作日志, 其他用户只能查看自己的操作日志
	if currentUid != pkgConst.UIDOfUserSys {
		if canViewProject, err := u.opPermissionVerifyUsecase.CanViewProject(ctx, currentUid, projectUid); err != nil {
			return 0, err
		} else if canViewProject {
			// do nothing,skip to next,because admin or global view user can view all operation logs
		} else if currentUid != filterPersonID {
			return 0, nil
		}
	}

	count, err := u.repo.CountOperationLogs(ctx, option)
	if err != nil {
		return 0, err
	}

	return count, nil
}
