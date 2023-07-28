//go:build enterprise

package biz

import (
	"context"
	"errors"
	"fmt"

	pkgConst "github.com/actiontech/dms/internal/dms/pkg/constant"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"

	dmsV1 "github.com/actiontech/dms/pkg/dms-common/api/dms/v1"
)

func (d *NamespaceUsecase) CreateNamespace(ctx context.Context, namespace *Namespace, createUserUID string) (err error) {
	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	err = d.repo.SaveNamespace(tx, namespace)
	if err != nil {
		return fmt.Errorf("save namespaces failed: %v", err)
	}

	// 默认将admin用户加入空间成员，并且为管理员
	_, err = d.memberUsecase.AddUserToNamespaceAdminMember(tx, pkgConst.UIDOfUserAdmin, namespace.UID)
	if err != nil {
		return fmt.Errorf("add admin to namespaces failed: %v", err)
	}
	// 非admin用户创建时,默认将空间创建人加入空间成员，并且为管理员
	if createUserUID != pkgConst.UIDOfUserAdmin {
		_, err = d.memberUsecase.AddUserToNamespaceAdminMember(tx, createUserUID, namespace.UID)
		if err != nil {
			return fmt.Errorf("add create user to namespaces failed: %v", err)
		}
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}

	return nil
}

func (d *NamespaceUsecase) GetNamespace(ctx context.Context, namespaceUid string) (*Namespace, error) {
	return d.repo.GetNamespace(ctx, namespaceUid)
}

func (d *NamespaceUsecase) GetNamespaceByName(ctx context.Context, namespaceName string) (*Namespace, error) {
	return d.repo.GetNamespaceByName(ctx, namespaceName)
}

func (d *NamespaceUsecase) UpdateNamespaceDesc(ctx context.Context, currentUserUid, namespaceUid string, desc *string) (err error) {
	if err := d.checkUserCanUpdateNamespace(ctx, currentUserUid, namespaceUid); err != nil {
		return fmt.Errorf("user can't update namespace: %v", err)
	}

	namespace, err := d.repo.GetNamespace(ctx, namespaceUid)
	if err != nil {
		return fmt.Errorf("get namespace err: %v", err)
	}

	if desc != nil {
		namespace.Desc = *desc
	}

	err = d.repo.UpdateNamespace(ctx, namespace)
	if err != nil {
		return fmt.Errorf("update namespaces desc failed: %v", err)
	}

	return nil
}

func (d *NamespaceUsecase) ArchivedNamespace(ctx context.Context, currentUserUid, namespaceUid string, archived bool) (err error) {
	if err := d.checkUserCanUpdateNamespace(ctx, currentUserUid, namespaceUid); err != nil {
		return fmt.Errorf("user can't update namespace: %v", err)
	}

	namespace, err := d.repo.GetNamespace(ctx, namespaceUid)
	if err != nil {
		return fmt.Errorf("get namespace err: %v", err)
	}

	// 调整空间状态
	var status NamespaceStatus
	if archived {
		status = NamespaceStatusArchived
	} else {
		status = NamespaceStatusActive
	}
	if status == namespace.Status {
		return fmt.Errorf("can't operate namespace current status is %v", status)
	}
	namespace.Status = status

	// plugin check before delete namespace
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, namespaceUid, dmsV1.DataResourceTypeNamespace, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore)
	if err != nil {
		return fmt.Errorf("check before delete namespace failed: %v", err)
	}

	err = d.repo.UpdateNamespace(ctx, namespace)
	if err != nil {
		return fmt.Errorf("update namespaces status failed: %v", err)
	}

	return nil
}

func (d *NamespaceUsecase) DeleteNamespace(ctx context.Context, currentUserUid, namespaceUid string) (err error) {
	// check
	{
		// namespace admin can delete namespace
		isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, namespaceUid)
		if err != nil {
			return fmt.Errorf("check user namespace admin error: %v", err)
		}
		if !isAdmin {
			return fmt.Errorf("user can't update namespace")
		}

		// plugin check before delete namespace
		err = d.pluginUsecase.OperateDataResourceHandle(ctx, namespaceUid, dmsV1.DataResourceTypeNamespace, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore)
		if err != nil {
			return fmt.Errorf("check before delete namespace failed: %v", err)
		}

	}
	err = d.repo.DelNamespace(ctx, namespaceUid)
	if err != nil {
		return err
	}
	// plugin clean unused data after delete namespace
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, namespaceUid, dmsV1.DataResourceTypeNamespace, dmsV1.OperationTypeDelete, dmsV1.OperationTimingAfter)
	if err != nil {
		return err
	}
	return nil
}

func (d *NamespaceUsecase) checkUserCanUpdateNamespace(ctx context.Context, currentUserUid, namespaceUid string) error {
	// namespace admin can update namespace
	isAdmin, err := d.opPermissionVerifyUsecase.IsUserNamespaceAdmin(ctx, currentUserUid, namespaceUid)
	if err != nil {
		return fmt.Errorf("check user namespace admin error: %v", err)
	}
	if !isAdmin {
		return fmt.Errorf("user can't update namespace")
	}
	return nil
}

func (d *NamespaceUsecase) isNamespaceActive(ctx context.Context, namespaceUid string) error {
	namespace, err := d.GetNamespace(ctx, namespaceUid)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("namespace not exist"))
		}
		return err
	}

	if namespace.Status != NamespaceStatusActive {
		return fmt.Errorf("namespace status is : %v", namespace.Status)
	}
	return nil
}
