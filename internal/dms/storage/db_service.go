package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.DBServiceRepo = (*DBServiceRepo)(nil)

type DBServiceRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewDBServiceRepo(log utilLog.Logger, s *Storage) *DBServiceRepo {
	return &DBServiceRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.db_service"))}
}

func (d *DBServiceRepo) SaveDBService(ctx context.Context, ds *biz.DBService) error {
	model, err := convertBizDBService(ds)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz db service: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to save db service: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *DBServiceRepo) ListDBServices(ctx context.Context, opt *biz.ListDBServicesOption) (services []*biz.DBService, total int64, err error) {

	var models []*model.DBService
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			db = gormWheres(ctx, db, opt.FilterBy)
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber))))).Find(&models)
			if err := db.Error; err != nil {
				return fmt.Errorf("failed to list db service: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.DBService{})
			db = gormWheres(ctx, db, opt.FilterBy)
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count db service: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelDBService(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model db services: %w", err))
		}
		services = append(services, ds)
	}
	return services, total, nil
}

func (d *DBServiceRepo) GetDBServicesByIds(ctx context.Context, dbServiceIds []string) (services []*biz.DBService, err error) {
	var items []*model.DBService
	if err = transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		db := tx.WithContext(ctx).Find(&items, dbServiceIds)

		if err = db.Error; err != nil {
			return fmt.Errorf("failed to list db service: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// convert model to biz
	for _, item := range items {
		ds, err := convertModelDBService(item)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model db services: %w", err))
		}
		services = append(services, ds)
	}

	return services, nil
}

func (d *DBServiceRepo) DelDBService(ctx context.Context, dbServiceUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", dbServiceUid).Delete(&model.DBService{}).Error; err != nil {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to delete db service: %v", err))
		}

		var memberRoleItems []*model.MemberRoleOpRange
		if err := tx.WithContext(ctx).Where("op_range_type = ? and find_in_set(?, range_uids)", "db_service", dbServiceUid).Find(&memberRoleItems).Error; err != nil {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to find member_role_op_range err: %v", err))
		}

		for _, item := range memberRoleItems {
			var dbServiceIds []string
			for _, uid := range strings.Split(item.RangeUIDs, ",") {
				if uid != dbServiceUid {
					dbServiceIds = append(dbServiceIds, uid)
				}
			}

			item.RangeUIDs = strings.Join(dbServiceIds, ",")

			if err := tx.WithContext(ctx).Model(&model.MemberRoleOpRange{}).Where("member_uid = ? and role_uid = ? and op_range_type = ?", item.MemberUID, item.RoleUID, item.OpRangeType).Update("range_uids", item.RangeUIDs).Error; err != nil {
				return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to update member_role_op_range err: %v", err))
			}
		}

		var memberGroupRoleItems []*model.MemberGroupRoleOpRange
		if err := tx.WithContext(ctx).Where("op_range_type = ? and find_in_set(?, range_uids)", "db_service", dbServiceUid).Find(&memberGroupRoleItems).Error; err != nil {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to find member_group_role_op_range err: %v", err))
		}

		for _, item := range memberGroupRoleItems {
			var dbServiceIds []string
			for _, uid := range strings.Split(item.RangeUIDs, ",") {
				if uid != dbServiceUid {
					dbServiceIds = append(dbServiceIds, uid)
				}
			}

			item.RangeUIDs = strings.Join(dbServiceIds, ",")

			if err := tx.WithContext(ctx).Model(&model.MemberGroupRoleOpRange{}).Where("member_group_uid = ? and role_uid = ? and op_range_type = ?", item.MemberGroupUID, item.RoleUID, item.OpRangeType).Update("range_uids", item.RangeUIDs).Error; err != nil {
				return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to update member_group_role_op_range err: %v", err))
			}
		}

		return nil
	})
}

func (d *DBServiceRepo) GetDBServices(ctx context.Context, conditions []pkgConst.FilterCondition) (services []*biz.DBService, err error) {
	var models []*model.DBService
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		// find models
		{
			db := tx.WithContext(ctx)
			for _, f := range conditions {
				db = gormWhere(db, f)
			}
			db = db.Find(&models)
			if err := db.Error; err != nil {
				return fmt.Errorf("failed to list db service: %v", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelDBService(model)
		if err != nil {
			return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model db services: %w", err))
		}
		services = append(services, ds)
	}
	return services, nil
}

func (d *DBServiceRepo) GetDBService(ctx context.Context, dbServiceUid string) (*biz.DBService, error) {
	var dbService *model.DBService
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&dbService, "uid = ?", dbServiceUid).Error; err != nil {
			return fmt.Errorf("failed to get db service: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModelDBService(dbService)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model db service: %w", err))
	}
	return ret, nil
}

func (d *DBServiceRepo) CheckDBServiceExist(ctx context.Context, dbServiceUids []string) (exists bool, err error) {
	var count int64
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DBService{}).Where("uid in (?)", dbServiceUids).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check dbService exist: %v", err)
		}
		return nil
	}); err != nil {
		return false, err
	}

	if count != int64(len(dbServiceUids)) {
		return false, nil
	}
	return true, nil
}

func (d *DBServiceRepo) UpdateDBService(ctx context.Context, dbService *biz.DBService) error {

	db, err := convertBizDBService(dbService)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz dbService: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DBService{}).Where("uid = ?", db.UID).Omit("created_at").Save(db).Error; err != nil {
			return fmt.Errorf("failed to update dbService: %v", err)
		}
		return nil
	})

}

func (d *DBServiceRepo) CountDBService(ctx context.Context) ([]biz.DBTypeCount, error) {
	type Result struct {
		DBType string `json:"db_type"`
		Count  int64  `json:"count"`
	}
	var results []Result

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.DBService{}).Select("db_type, count(*) as count").Group("db_type").Find(&results).Error; err != nil {
			return fmt.Errorf("failed to count dbService: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	dbTypeCounts := make([]biz.DBTypeCount, 0, len(results))
	for _, result := range results {
		dbTypeCounts = append(dbTypeCounts, biz.DBTypeCount{
			DBType: result.DBType,
			Count:  result.Count,
		})
	}
	return dbTypeCounts, nil
}

func (d *DBServiceRepo) GetBusiness(ctx context.Context, projectUid string) ([]string, error) {
	businessList := make([]string, 0)
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Raw(`
SELECT DISTINCT business
FROM db_services
WHERE project_uid = ?;
	`, projectUid).Find(&businessList).Error; err != nil {
			return fmt.Errorf("failed to get business: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return businessList, nil
}
