package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.OpsTypeRepo = (*OpsTypeRepo)(nil)

type OpsTypeRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewOpsTypeRepo(log utilLog.Logger, s *Storage) *OpsTypeRepo {
	return &OpsTypeRepo{
		Storage: s,
		log:     utilLog.NewHelper(log, utilLog.WithMessageKey("storage.ops_type")),
	}
}

func (repo *OpsTypeRepo) toModel(opsType *biz.OpsType) *model.OpsType {
	return &model.OpsType{
		OpsTypeName: opsType.Name,
		Model:       model.Model{UID: opsType.UID},
		ProjectUID:  opsType.ProjectUID,
	}
}

func (repo *OpsTypeRepo) toBiz(opsType *model.OpsType) *biz.OpsType {
	return &biz.OpsType{
		Name:       opsType.OpsTypeName,
		UID:        opsType.UID,
		ProjectUID: opsType.ProjectUID,
	}
}

func (repo *OpsTypeRepo) CreateOpsType(ctx context.Context, opsType *biz.OpsType) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(repo.toModel(opsType)).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to create ops type: %v", err))
		}
		return nil
	})
}

func (repo *OpsTypeRepo) UpdateOpsType(ctx context.Context, opsTypeUID, opsTypeName string) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.OpsType{}).Where("uid = ?", opsTypeUID).Updates(map[string]interface{}{
			"ops_type_name": opsTypeName,
		}).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to update ops type: %v", err))
		}
		return nil
	})
}

func (repo *OpsTypeRepo) DeleteOpsType(ctx context.Context, opsTypeUID string) error {
	return transaction(repo.log, ctx, repo.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", opsTypeUID).Delete(&model.OpsType{}).Error; err != nil {
			return pkgErr.WrapStorageErr(repo.log, fmt.Errorf("failed to delete ops type: %v", err))
		}
		return nil
	})
}

func (repo *OpsTypeRepo) GetOpsTypeByName(ctx context.Context, projectUID, name string) (bool, *biz.OpsType, error) {
	var opsType model.OpsType
	if err := repo.db.WithContext(ctx).Where("ops_type_name = ? AND project_uid = ?", name, projectUID).First(&opsType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to get ops type by name: %w", err)
	}
	return true, repo.toBiz(&opsType), nil
}

func (repo *OpsTypeRepo) GetOpsTypeByUID(ctx context.Context, uid string) (*biz.OpsType, error) {
	var opsType model.OpsType
	if err := repo.db.WithContext(ctx).Where("uid = ?", uid).First(&opsType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgErr.ErrStorageNoData
		}
		return nil, fmt.Errorf("failed to get ops type by uid: %w", err)
	}
	return repo.toBiz(&opsType), nil
}

func (repo *OpsTypeRepo) ListOpsTypes(ctx context.Context, options *biz.ListOpsTypesOption) ([]*biz.OpsType, int64, error) {
	var opsTypes []*model.OpsType
	db := repo.db.WithContext(ctx)

	query := db.Model(&model.OpsType{}).Where("project_uid = ?", options.ProjectUID)
	if options.Limit >= 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset >= 0 {
		query = query.Offset(options.Offset)
	}

	if err := query.Find(&opsTypes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ops types: %w", err)
	}

	var count int64
	if err := repo.db.WithContext(ctx).Model(&model.OpsType{}).Where("project_uid = ?", options.ProjectUID).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ops types: %w", err)
	}

	bizOpsTypes := make([]*biz.OpsType, 0, len(opsTypes))
	for _, opsType := range opsTypes {
		bizOpsTypes = append(bizOpsTypes, repo.toBiz(opsType))
	}

	return bizOpsTypes, count, nil
}
