package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgParams "github.com/actiontech/dms/pkg/params"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"gorm.io/gorm"
)

var _ biz.DBServiceSyncTaskRepo = (*DBServiceSyncTaskRepo)(nil)

type DBServiceSyncTaskRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewDBServiceSyncTaskRepo(log utilLog.Logger, s *Storage) *DBServiceSyncTaskRepo {
	return &DBServiceSyncTaskRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.db_service_sync_task"))}
}

func (d *DBServiceSyncTaskRepo) SaveDBServiceSyncTask(ctx context.Context, syncTask *biz.DBServiceSyncTask) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		modelDBServiceSyncTask := toModelDBServiceSyncTask(syncTask)
		if err := tx.WithContext(ctx).Create(modelDBServiceSyncTask).Error; err != nil {
			return fmt.Errorf("failed to save db_service_sync_task: %v", err)
		}
		return nil
	})
}

func (d *DBServiceSyncTaskRepo) GetDBServiceSyncTaskById(ctx context.Context, id string) (*biz.DBServiceSyncTask, error) {
	var dbServiceSyncTask *model.DBServiceSyncTask
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&dbServiceSyncTask, "uid = ?", id).Error; err != nil {
			return fmt.Errorf("failed to get db_service_sync_task: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return toBizDBServiceSyncTask(dbServiceSyncTask), nil
}

func (d *DBServiceSyncTaskRepo) UpdateDBServiceSyncTask(ctx context.Context, syncTask *biz.DBServiceSyncTask) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Omit("created_at").Save(toModelDBServiceSyncTask(syncTask)).Error; err != nil {
			return fmt.Errorf("failed to update db_service_sync_task: %v", err)
		}
		return nil
	})
}


func (d *DBServiceSyncTaskRepo) ListDBServiceSyncTasks(ctx context.Context) ([]*biz.DBServiceSyncTask, error) {
	var items []*model.DBServiceSyncTask
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		if err := tx.WithContext(ctx).Find(&items).Error; err != nil {
			return fmt.Errorf("failed to list db_service_sync_task: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret := make([]*biz.DBServiceSyncTask, 0, len(items))
	// convert model to biz
	for _, item := range items {
		ret = append(ret, toBizDBServiceSyncTask(item))
	}
	return ret, nil
}


func (d *DBServiceSyncTaskRepo) ListDBServiceSyncTaskTips() ([]biz.ListDBServiceSyncTaskTips, error) {
	return DatabaseSourceServiceTips, nil
}

var DatabaseSourceServiceTips = []biz.ListDBServiceSyncTaskTips{
	{
		Type:    pkgConst.DBServiceSourceNameDMP,
		Desc:    "Actiontech DMP",
		DBTypes: []pkgConst.DBType{pkgConst.DBTypeMySQL},
		Params:  pkgParams.Params{
			{
				Key:  "version",
				Desc: "版本(支持DMP5.23.04.0及以上版本)",
				Type: pkgParams.ParamTypeString,
			},
		},
	},
	{
		Type: pkgConst.DBServiceSourceNameExpandService,
		Desc: "数据源同步扩展服务",
		DBTypes: []pkgConst.DBType{
			pkgConst.DBTypeMySQL,
			pkgConst.DBTypePostgreSQL,
			pkgConst.DBTypeTiDB,
			pkgConst.DBTypeSQLServer,
			pkgConst.DBTypeOracle,
			pkgConst.DBTypeDB2,
			pkgConst.DBTypeOceanBaseMySQL,
			pkgConst.DBTypeTDSQLForInnoDB,
			pkgConst.DBTypeGoldenDB,
		},
	},
}


func (d *DBServiceSyncTaskRepo) DeleteDBServiceSyncTask(ctx context.Context, dbServiceSyncTaskUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", dbServiceSyncTaskUid).Delete(&model.DBServiceSyncTask{}).Error; err != nil {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to delete db_service_sync_task: %v", err))
		}
		return nil
	})
}

func (d *DBServiceSyncTaskRepo) UpdateDBServiceSyncTaskByFields(ctx context.Context, dbServiceSyncTaskUid string, fields map[string]interface{}) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DBServiceSyncTask{}).Where("uid = ?", dbServiceSyncTaskUid).Updates(fields).Error; err != nil {
			return fmt.Errorf("failed to update db_service_sync_task: %v", err)
		}

		return nil
	})
}