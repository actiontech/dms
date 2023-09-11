package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.NamespaceRepo = (*NamespaceRepo)(nil)

type NamespaceRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewNamespaceRepo(log utilLog.Logger, s *Storage) *NamespaceRepo {
	return &NamespaceRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.namespace"))}
}

func (d *NamespaceRepo) SaveNamespace(ctx context.Context, u *biz.Namespace) error {
	model, err := convertBizNamespace(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz namespace: %v", err))
	}

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to save namespace: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *NamespaceRepo) ListNamespaces(ctx context.Context, opt *biz.ListNamespacesOption, currentUserUid string) (namespaces []*biz.Namespace, total int64, err error) {
	var models []*model.Namespace

	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {

		// find models
		{
			db := tx.WithContext(ctx).Order(opt.OrderBy)
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			db = db.Limit(int(opt.LimitPerPage)).Offset(int(opt.LimitPerPage * (uint32(fixPageIndices(opt.PageNumber)))))
			if err := db.Find(&models).Error; err != nil {
				return fmt.Errorf("failed to list namespaces: %v", err)
			}
		}

		// find total
		{
			db := tx.WithContext(ctx).Model(&model.Namespace{})
			for _, f := range opt.FilterBy {
				db = gormWhere(db, f)
			}
			if err := db.Count(&total).Error; err != nil {
				return fmt.Errorf("failed to count namespaces: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, 0, err
	}

	// convert model to biz
	for _, model := range models {
		ds, err := convertModelNamespace(model)
		if err != nil {
			return nil, 0, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model namespaces: %v", err))
		}
		// ds.CreateUserName = model.UserName
		namespaces = append(namespaces, ds)
	}
	return namespaces, total, nil
}

func (d *NamespaceRepo) GetNamespace(ctx context.Context, namespaceUid string) (*biz.Namespace, error) {
	var namespace *model.Namespace
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&namespace, "uid = ?", namespaceUid).Error; err != nil {
			return fmt.Errorf("failed to get namespace: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelNamespace(namespace)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model namespace: %v", err))
	}
	return ret, nil
}

func (d *NamespaceRepo) GetNamespaceByName(ctx context.Context, namespaceName string) (*biz.Namespace, error) {
	var namespace *model.Namespace
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.First(&namespace, "name = ?", namespaceName).Error; err != nil {
			return fmt.Errorf("failed to get namespace by name: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	ret, err := convertModelNamespace(namespace)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model namespace: %v", err))
	}
	return ret, nil
}

func (d *NamespaceRepo) UpdateNamespace(ctx context.Context, u *biz.Namespace) error {
	_, err := d.GetNamespace(ctx, u.UID)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("namespace not exist"))
		}
		return err
	}

	namespace, err := convertBizNamespace(u)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz namespace: %v", err))
	}

	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.Namespace{}).Where("uid = ?", u.UID).Omit("created_at").Save(namespace).Error; err != nil {
			return fmt.Errorf("failed to update namespace: %v", err)
		}
		return nil
	})

}

func (d *NamespaceRepo) DelNamespace(ctx context.Context, namespaceUid string) error {
	return transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("uid = ?", namespaceUid).Delete(&model.Namespace{}).Error; err != nil {
			return fmt.Errorf("failed to delete namespace: %v", err)
		}
		return nil
	})
}

func (d *NamespaceRepo) IsNamespaceActive(ctx context.Context, namespaceUid string) error {
	namespace, err := d.GetNamespace(ctx, namespaceUid)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("namespace not exist"))
		}
		return err
	}

	if namespace.Status != biz.NamespaceStatusActive {
		return fmt.Errorf("namespace status is : %v", namespace.Status)
	}
	return nil
}
