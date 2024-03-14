package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"

	"gorm.io/gorm"
)

var _ biz.CloudbeaverRepo = (*CloudbeaverRepo)(nil)

type CloudbeaverRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewCloudbeaverRepo(log utilLog.Logger, s *Storage) *CloudbeaverRepo {
	return &CloudbeaverRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.cloudbeaver"))}
}

func (cr *CloudbeaverRepo) GetCloudbeaverUserByID(ctx context.Context, cloudbeaverUserId string) (*biz.CloudbeaverUser, bool, error) {
	var user model.CloudbeaverUserCache
	err := transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.Find(&user, "cloudbeaver_user_id = ?", cloudbeaverUserId).Error; err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}
		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}

		return nil, false, err
	}

	return convertModelCloudbeaverUser(&user), true, nil
}

func (cr *CloudbeaverRepo) UpdateCloudbeaverUserCache(ctx context.Context, u *biz.CloudbeaverUser) error {
	return transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(convertBizCloudbeaverUser(u)).Error; err != nil {
			return fmt.Errorf("failed to update user: %v", err)
		}
		return nil
	})
}

func (cr *CloudbeaverRepo) GetDbServiceIdByConnectionId(ctx context.Context, connectionId string) (string, error) {
	var cloudbeaverConnection model.CloudbeaverConnectionCache
	err := transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.Where("cloudbeaver_connection_id = ?", connectionId).First(&cloudbeaverConnection).Error; err != nil {
			return fmt.Errorf("failed to get cloudbeaver db_service_id: %v", err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return cloudbeaverConnection.DMSDBServiceID, nil
}

func (cr *CloudbeaverRepo) GetAllCloudbeaverConnections(ctx context.Context) ([]*biz.CloudbeaverConnection, error) {
	var cloudbeaverConnections []*model.CloudbeaverConnectionCache
	err := transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.Find(&cloudbeaverConnections).Error; err != nil {
			return fmt.Errorf("failed to get cloudbeaver db service: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return convertModelCloudbeaverConnection(cloudbeaverConnections), nil
}

func (cr *CloudbeaverRepo) GetCloudbeaverConnectionsByUserIdAndDBServiceIds(ctx context.Context, userId string, dmsDBServiceIds []string) ([]*biz.CloudbeaverConnection, error) {
	var cloudbeaverConnections []*model.CloudbeaverConnectionCache
	err := transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.Where("dms_user_id = ? and dms_db_service_id in (?)", userId, dmsDBServiceIds).Find(&cloudbeaverConnections).Error; err != nil {
			return fmt.Errorf("failed to get cloudbeaver db service: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return convertModelCloudbeaverConnection(cloudbeaverConnections), nil
}

func (cr *CloudbeaverRepo) GetCloudbeaverConnectionsByUserId(ctx context.Context, userId string) ([]*biz.CloudbeaverConnection, error) {
	var cloudbeaverConnections []*model.CloudbeaverConnectionCache
	err := transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.Where("dms_user_id = ?", userId).Find(&cloudbeaverConnections).Error; err != nil {
			return fmt.Errorf("failed to get cloudbeaver db service: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return convertModelCloudbeaverConnection(cloudbeaverConnections), nil
}

func (cr *CloudbeaverRepo) UpdateCloudbeaverConnectionCache(ctx context.Context, u *biz.CloudbeaverConnection) error {
	return transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(convertBizCloudbeaverConnection(u)).Error; err != nil {
			return fmt.Errorf("failed to update cloudbeaver db Service: %v", err)
		}
		return nil
	})
}

func (cr *CloudbeaverRepo) DeleteCloudbeaverConnectionCache(ctx context.Context, dbServiceId, userId, purpose string) error {
	return transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		db := tx.WithContext(ctx).Where("dms_db_service_id = ?", dbServiceId)
		if len(userId) > 0 {
			db = db.Where("dms_user_id = ?", userId)
		}
		if len(purpose) > 0 {
			db = db.Where("purpose = ?", purpose)
		}
		if err := db.Delete(&model.CloudbeaverConnectionCache{}).Error; err != nil {
			return fmt.Errorf("failed to delete cloudbeaver db Service: %v", err)
		}
		return nil
	})
}
