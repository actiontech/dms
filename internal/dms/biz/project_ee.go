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

func (d *ProjectUsecase) CreateProject(ctx context.Context, project *Project, createUserUID string) (err error) {
	tx := d.tx.BeginTX(ctx)
	defer func() {
		if err != nil {
			err = tx.RollbackWithError(d.log, err)
		}
	}()

	err = d.repo.SaveProject(tx, project)
	if err != nil {
		return fmt.Errorf("save projects failed: %v", err)
	}

	// 默认将admin用户加入空间成员，并且为管理员
	_, err = d.memberUsecase.AddUserToProjectAdminMember(tx, pkgConst.UIDOfUserAdmin, project.UID)
	if err != nil {
		return fmt.Errorf("add admin to projects failed: %v", err)
	}
	// 非admin用户创建时,默认将空间创建人加入空间成员，并且为管理员
	if createUserUID != pkgConst.UIDOfUserAdmin {
		_, err = d.memberUsecase.AddUserToProjectAdminMember(tx, createUserUID, project.UID)
		if err != nil {
			return fmt.Errorf("add create user to projects failed: %v", err)
		}
	}

	if err := tx.Commit(d.log); err != nil {
		return fmt.Errorf("commit tx failed: %v", err)
	}
	// plugin handle after create project
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, project.UID, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeCreate, dmsV1.OperationTimingAfter)
	if err != nil {
		return fmt.Errorf("plugin handle after create project failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) GetProjectByName(ctx context.Context, projectName string) (*Project, error) {
	return d.repo.GetProjectByName(ctx, projectName)
}

func (d *ProjectUsecase) UpdateProjectDesc(ctx context.Context, currentUserUid, projectUid string, desc *string) (err error) {
	if err := d.checkUserCanUpdateProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("user can't update project: %v", err)
	}

	project, err := d.repo.GetProject(ctx, projectUid)
	if err != nil {
		return fmt.Errorf("get project err: %v", err)
	}

	if desc != nil {
		project.Desc = *desc
	}

	err = d.repo.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("update projects desc failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) ArchivedProject(ctx context.Context, currentUserUid, projectUid string, archived bool) (err error) {
	if err := d.checkUserCanUpdateProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("user can't update project: %v", err)
	}

	project, err := d.repo.GetProject(ctx, projectUid)
	if err != nil {
		return fmt.Errorf("get project err: %v", err)
	}

	// 调整空间状态
	var status ProjectStatus
	if archived {
		status = ProjectStatusArchived
	} else {
		status = ProjectStatusActive
	}
	if status == project.Status {
		return fmt.Errorf("can't operate project current status is %v", status)
	}
	project.Status = status

	// plugin check before delete project
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, projectUid, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore)
	if err != nil {
		return fmt.Errorf("check before delete project failed: %v", err)
	}

	err = d.repo.UpdateProject(ctx, project)
	if err != nil {
		return fmt.Errorf("update projects status failed: %v", err)
	}

	return nil
}

func (d *ProjectUsecase) DeleteProject(ctx context.Context, currentUserUid, projectUid string) (err error) {
	// check
	{
		// project admin can delete project
		isAdmin, err := d.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, projectUid)
		if err != nil {
			return fmt.Errorf("check user project admin error: %v", err)
		}
		if !isAdmin {
			return fmt.Errorf("user can't update project")
		}

		// plugin check before delete project
		err = d.pluginUsecase.OperateDataResourceHandle(ctx, projectUid, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingTypeBefore)
		if err != nil {
			return fmt.Errorf("check before delete project failed: %v", err)
		}

	}
	err = d.repo.DelProject(ctx, projectUid)
	if err != nil {
		return err
	}
	// plugin clean unused data after delete project
	err = d.pluginUsecase.OperateDataResourceHandle(ctx, projectUid, dmsV1.DataResourceTypeProject, dmsV1.OperationTypeDelete, dmsV1.OperationTimingAfter)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectUsecase) checkUserCanUpdateProject(ctx context.Context, currentUserUid, projectUid string) error {
	// project admin can update project
	isAdmin, err := d.opPermissionVerifyUsecase.IsUserProjectAdmin(ctx, currentUserUid, projectUid)
	if err != nil {
		return fmt.Errorf("check user project admin error: %v", err)
	}
	if !isAdmin {
		return fmt.Errorf("user can't update project")
	}
	return nil
}

func (d *ProjectUsecase) isProjectActive(ctx context.Context, projectUid string) error {
	project, err := d.GetProject(ctx, projectUid)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return pkgErr.WrapStorageErr(d.log, fmt.Errorf("project not exist"))
		}
		return err
	}

	if project.Status != ProjectStatusActive {
		return fmt.Errorf("project status is : %v", project.Status)
	}
	return nil
}
