package biz

import (
	"context"
	"fmt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

func EnvPrepare(ctx context.Context, logger utilLog.Logger,
	transaction TransactionGenerator,
	config *DMSConfigUseCase,
	opPermissionUsecase *OpPermissionUsecase,
	userUsecase *UserUsecase,
	roleUsecase *RoleUsecase,
	projectUsecase *ProjectUsecase) (err error) {
	log := utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.prepare"))
	// 开启事务
	tx := transaction.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(log, err)
		}
	}()

	// 初始化data operations
	{
		dmsConfig, err := config.GetDMSConfig(tx)
		if nil != err {
			return fmt.Errorf("failed to get dms config: %v", err)
		}

		// 如果找到了，则判断是否需要初始化数据对象操作
		if dmsConfig.NeedInitOpPermissions {
			if err := opPermissionUsecase.InitOpPermissions(tx, initOpPermission()); nil != err {
				return err
			}
			dmsConfig.NeedInitOpPermissions = false
			if err := config.UpdateDMSConfig(tx, dmsConfig); nil != err {
				return fmt.Errorf("failed to update dms config: %v", err)
			}
		}
		if dmsConfig.NeedInitUsers {
			if err := userUsecase.InitUsers(tx); nil != err {
				return err
			}
			dmsConfig.NeedInitUsers = false
			if err := config.UpdateDMSConfig(tx, dmsConfig); nil != err {
				return fmt.Errorf("failed to update dms config: %v", err)
			}
		}
		if dmsConfig.NeedInitRoles {
			if err := roleUsecase.InitRoles(tx); nil != err {
				return err
			}
			dmsConfig.NeedInitRoles = false
			if err := config.UpdateDMSConfig(tx, dmsConfig); nil != err {
				return fmt.Errorf("failed to update dms config: %v", err)
			}
		}
		if dmsConfig.NeedInitProjects {
			if err := projectUsecase.InitProjects(tx); nil != err {
				return err
			}
			dmsConfig.NeedInitProjects = false
			if err := config.UpdateDMSConfig(tx, dmsConfig); nil != err {
				return fmt.Errorf("failed to update dms config: %v", err)
			}
		}
	}
	if err := tx.Commit(log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	return nil
}
