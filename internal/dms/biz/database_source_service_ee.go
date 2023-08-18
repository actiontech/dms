//go:build enterprise

package biz

import (
	"context"
	"fmt"
	"time"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

func (d *DatabaseSourceServiceUsecase) ListDatabaseSourceServices(ctx context.Context, conditions []pkgConst.FilterCondition, namespaceId string, currentUserId string) ([]*DatabaseSourceServiceParams, error) {
	// 只允许系统用户查询所有数据源,同步数据到其他服务(provision)
	// 检查空间是否归档/删除
	if namespaceId != "" {
		if err := d.namespaceUsecase.isNamespaceActive(ctx, namespaceId); err != nil {
			return nil, fmt.Errorf("list database_source_service error: %v", err)
		}
	} else if currentUserId != pkgConst.UIDOfUserSys {
		return nil, fmt.Errorf("list database_source_service error: namespace is empty")
	}

	services, err := d.repo.ListDatabaseSourceServices(ctx, conditions)
	if err != nil {
		return nil, fmt.Errorf("list database_source_service failed: %w", err)
	}

	return services, nil
}

func (d *DatabaseSourceServiceUsecase) AddDatabaseSourceService(ctx context.Context, params *DatabaseSourceServiceParams, currentUserId string) (string, error) {
	// 检查空间是否归档/删除
	if err := d.namespaceUsecase.isNamespaceActive(ctx, params.NamespaceUID); err != nil {
		return "", fmt.Errorf("create database_source_service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserId, params.NamespaceUID); err != nil {
		return "", fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return "", fmt.Errorf("user is not namespace admin")
	}

	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return "", err
	}

	params.UID = uid

	return uid, d.repo.SaveDatabaseSourceService(ctx, params)
}

func (d *DatabaseSourceServiceUsecase) UpdateDatabaseSourceService(ctx context.Context, databaseSourceServiceId string, params *DatabaseSourceServiceParams, currentUserId string) error {
	// 检查空间是否归档/删除
	if err := d.namespaceUsecase.isNamespaceActive(ctx, params.NamespaceUID); err != nil {
		return fmt.Errorf("update database_source_service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserId, params.NamespaceUID); err != nil {
		return fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not namespace admin")
	}

	databaseSourceService, err := d.repo.GetDatabaseSourceServiceById(ctx, databaseSourceServiceId)
	if err != nil {
		return fmt.Errorf("get database_source_service failed: %v", err)
	}

	if databaseSourceService.DbType != params.DbType {
		return fmt.Errorf("update database_source_service type is unsupported")
	}

	params.UID = databaseSourceServiceId
	params.NamespaceUID = databaseSourceService.NamespaceUID
	params.LastSyncErr = databaseSourceService.LastSyncErr
	params.LastSyncSuccessTime = databaseSourceService.LastSyncSuccessTime

	return d.repo.UpdateDatabaseSourceService(ctx, params)
}

func (d *DatabaseSourceServiceUsecase) DeleteDatabaseSourceService(ctx context.Context, databaseSourceServiceId, currentUserId string) error {
	databaseSourceService, err := d.repo.GetDatabaseSourceServiceById(ctx, databaseSourceServiceId)
	if err != nil {
		return fmt.Errorf("get database_source_service failed: %v", err)
	}
	// 检查空间是否归档/删除
	if err = d.namespaceUsecase.isNamespaceActive(ctx, databaseSourceService.NamespaceUID); err != nil {
		return fmt.Errorf("delete database_source_service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserId, databaseSourceService.NamespaceUID); err != nil {
		return fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not namespace admin")
	}

	//todo: currently only database_source_services data is deleted
	return d.repo.DeleteDatabaseSourceService(ctx, databaseSourceServiceId)
}

func (d *DatabaseSourceServiceUsecase) ListDatabaseSourceServiceTips(ctx context.Context) ([]ListDatabaseSourceServiceTipsParams, error) {
	ret := []ListDatabaseSourceServiceTipsParams{
		{
			Source:  pkgConst.DBServiceSourceNameDMP,
			DbTypes: []pkgConst.DBType{pkgConst.DBTypeMySQL},
		},
	}

	return ret, nil
}

func (d *DatabaseSourceServiceUsecase) SyncDatabaseSourceService(ctx context.Context, databaseSourceServiceId, currentUserId string) error {
	databaseSourceService, err := d.repo.GetDatabaseSourceServiceById(ctx, databaseSourceServiceId)
	if err != nil {
		return fmt.Errorf("get database_source_service failed: %v", err)
	}
	// 检查空间是否归档/删除
	if err = d.namespaceUsecase.isNamespaceActive(ctx, databaseSourceService.NamespaceUID); err != nil {
		return fmt.Errorf("sync database_source_service error: %v", err)
	}
	// 检查当前用户有空间管理员权限
	if isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserId, databaseSourceService.NamespaceUID); err != nil {
		return fmt.Errorf("check user is namespace admin failed: %v", err)
	} else if !isAdmin {
		return fmt.Errorf("user is not namespace admin")
	}

	databaseSourceName, err := pkgConst.ParseDBServiceSource(databaseSourceService.Source)
	if err != nil {
		return err
	}

	databaseSourceImpl, err := GetDatabaseSourceImpl(databaseSourceName)
	if err != nil {
		return err
	}

	// sync database source
	syncErr := databaseSourceImpl.SyncDatabaseSource(ctx, databaseSourceService, d, currentUserId)
	fields := make(map[string]interface{})
	if syncErr != nil {
		fields[string(DatabaseSourceServiceFieldLastSyncErr)] = syncErr.Error()
	} else {
		currentTime := time.Now()
		fields[string(DatabaseSourceServiceFieldLastSyncErr)] = ""
		fields[string(DatabaseSourceServiceFieldLastSyncSuccessTime)] = &currentTime
	}

	if err = d.repo.UpdateSyncDatabaseSourceService(ctx, databaseSourceService.UID, fields); err != nil {
		return fmt.Errorf("update sync database source err: %v, sync err: %v", err, syncErr)
	}

	return nil
}
