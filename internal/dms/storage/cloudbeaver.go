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

func (cr *CloudbeaverRepo) GetCloudbeaverConnectionByDMSDBServiceIds(ctx context.Context, dmsDBServiceIds []string) ([]*biz.CloudbeaverConnection, error) {
	var cloudbeaverConnections []*model.CloudbeaverConnectionCache
	err := transaction(cr.log, ctx, cr.db, func(tx *gorm.DB) error {
		if err := tx.Find(&cloudbeaverConnections, dmsDBServiceIds).Error; err != nil {
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
