package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"gorm.io/gorm"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

var _ biz.SqlWorkbenchUserRepo = (*SqlWorkbenchRepo)(nil)

type SqlWorkbenchRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewSqlWorkbenchRepo(log utilLog.Logger, s *Storage) *SqlWorkbenchRepo {
	return &SqlWorkbenchRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.sql_workbench"))}
}

func (sr *SqlWorkbenchRepo) GetSqlWorkbenchUserByDMSUserID(ctx context.Context, dmsUserID string) (*biz.SqlWorkbenchUser, bool, error) {
	var user model.SqlWorkbenchUserCache
	err := transaction(sr.log, ctx, sr.db, func(tx *gorm.DB) error {
		if err := tx.First(&user, "dms_user_id = ?", dmsUserID).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, false, nil
	}

	return convertModelSqlWorkbenchUser(&user), true, nil
}

func (sr *SqlWorkbenchRepo) SaveSqlWorkbenchUserCache(ctx context.Context, user *biz.SqlWorkbenchUser) error {
	return transaction(sr.log, ctx, sr.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(convertBizSqlWorkbenchUser(user)).Error; err != nil {
			return fmt.Errorf("failed to save sql workbench user cache: %v", err)
		}
		return nil
	})
}

func convertModelSqlWorkbenchUser(user *model.SqlWorkbenchUserCache) *biz.SqlWorkbenchUser {
	return &biz.SqlWorkbenchUser{
		DMSUserID:            user.DMSUserID,
		SqlWorkbenchUserId:   user.SqlWorkbenchUserId,
		SqlWorkbenchUsername: user.SqlWorkbenchUsername,
	}
}

func convertBizSqlWorkbenchUser(user *biz.SqlWorkbenchUser) *model.SqlWorkbenchUserCache {
	return &model.SqlWorkbenchUserCache{
		DMSUserID:            user.DMSUserID,
		SqlWorkbenchUserId:   user.SqlWorkbenchUserId,
		SqlWorkbenchUsername: user.SqlWorkbenchUsername,
	}
}

// SqlWorkbenchDatasourceRepo 实现
var _ biz.SqlWorkbenchDatasourceRepo = (*SqlWorkbenchDatasourceRepo)(nil)

type SqlWorkbenchDatasourceRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewSqlWorkbenchDatasourceRepo(log utilLog.Logger, s *Storage) *SqlWorkbenchDatasourceRepo {
	return &SqlWorkbenchDatasourceRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.sql_workbench_datasource"))}
}

func (sr *SqlWorkbenchDatasourceRepo) GetSqlWorkbenchDatasourceByDMSDBServiceID(ctx context.Context, dmsDBServiceID, dmsUserID, purpose string) (*biz.SqlWorkbenchDatasource, bool, error) {
	var datasource model.SqlWorkbenchDatasourceCache
	err := transaction(sr.log, ctx, sr.db, func(tx *gorm.DB) error {
		if err := tx.First(&datasource, "dms_db_service_id = ? AND dms_user_id = ? AND purpose = ?", dmsDBServiceID, dmsUserID, purpose).Error; err != nil {
			return fmt.Errorf("failed to get sql workbench datasource: %v", err)
		}
		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	return convertModelSqlWorkbenchDatasource(&datasource), true, nil
}

func (sr *SqlWorkbenchDatasourceRepo) SaveSqlWorkbenchDatasourceCache(ctx context.Context, datasource *biz.SqlWorkbenchDatasource) error {
	return transaction(sr.log, ctx, sr.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(convertBizSqlWorkbenchDatasource(datasource)).Error; err != nil {
			return fmt.Errorf("failed to save sql workbench datasource cache: %v", err)
		}
		return nil
	})
}

func (sr *SqlWorkbenchDatasourceRepo) DeleteSqlWorkbenchDatasourceCache(ctx context.Context, dmsDBServiceID, dmsUserID, purpose string) error {
	return transaction(sr.log, ctx, sr.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("dms_db_service_id = ? AND dms_user_id = ? AND purpose = ?", dmsDBServiceID, dmsUserID, purpose).Delete(&model.SqlWorkbenchDatasourceCache{}).Error; err != nil {
			return fmt.Errorf("failed to delete sql workbench datasource cache: %v", err)
		}
		return nil
	})
}

func (sr *SqlWorkbenchDatasourceRepo) GetSqlWorkbenchDatasourcesByUserID(ctx context.Context, dmsUserID string) ([]*biz.SqlWorkbenchDatasource, error) {
	var datasources []model.SqlWorkbenchDatasourceCache
	err := transaction(sr.log, ctx, sr.db, func(tx *gorm.DB) error {
		if err := tx.Where("dms_user_id = ?", dmsUserID).Find(&datasources).Error; err != nil {
			return fmt.Errorf("failed to get sql workbench datasources by user id: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	var result []*biz.SqlWorkbenchDatasource
	for _, ds := range datasources {
		result = append(result, convertModelSqlWorkbenchDatasource(&ds))
	}
	return result, nil
}

func convertModelSqlWorkbenchDatasource(datasource *model.SqlWorkbenchDatasourceCache) *biz.SqlWorkbenchDatasource {
	return &biz.SqlWorkbenchDatasource{
		DMSDBServiceID:           datasource.DMSDBServiceID,
		DMSUserID:                datasource.DMSUserID,
		DMSDBServiceFingerprint:  datasource.DMSDBServiceFingerprint,
		SqlWorkbenchDatasourceID: datasource.SqlWorkbenchDatasourceID,
		Purpose:                  datasource.Purpose,
	}
}

func convertBizSqlWorkbenchDatasource(datasource *biz.SqlWorkbenchDatasource) *model.SqlWorkbenchDatasourceCache {
	return &model.SqlWorkbenchDatasourceCache{
		DMSDBServiceID:           datasource.DMSDBServiceID,
		DMSUserID:                datasource.DMSUserID,
		DMSDBServiceFingerprint:  datasource.DMSDBServiceFingerprint,
		SqlWorkbenchDatasourceID: datasource.SqlWorkbenchDatasourceID,
		Purpose:                  datasource.Purpose,
	}
}