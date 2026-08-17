package biz

import (
	"context"
	"fmt"
	"strings"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

// OpsTypeRepo 运维类型字典存储接口（对标 EnvironmentTagRepo）。
type OpsTypeRepo interface {
	CreateOpsType(ctx context.Context, opsType *OpsType) error
	UpdateOpsType(ctx context.Context, opsTypeUID, opsTypeName string) error
	DeleteOpsType(ctx context.Context, opsTypeUID string) error
	GetOpsTypeByName(ctx context.Context, projectUID, name string) (bool, *OpsType, error)
	GetOpsTypeByUID(ctx context.Context, uid string) (*OpsType, error)
	ListOpsTypes(ctx context.Context, options *ListOpsTypesOption) ([]*OpsType, int64, error)
}

// OpsType 项目级运维类型字典项。
type OpsType struct {
	UID        string
	Name       string
	ProjectUID string
}

// ListOpsTypesOption 按项目分页列举运维类型。
type ListOpsTypesOption struct {
	Limit      int
	Offset     int
	ProjectUID string
}

// OpsTypeUsecase 运维类型字典业务规则（权限 / 重名对齐 EnvironmentTagUsecase）。
type OpsTypeUsecase struct {
	opsTypeRepo               OpsTypeRepo
	projectUsecase            *ProjectUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	log                       *utilLog.Helper
}

func NewOpsTypeUsecase(opsTypeRepo OpsTypeRepo, logger utilLog.Logger,
	projectUsecase *ProjectUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase) *OpsTypeUsecase {
	return &OpsTypeUsecase{
		opsTypeRepo:               opsTypeRepo,
		projectUsecase:            projectUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		log:                       utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.ops_type")),
	}
}

// 默认运维类型名称（存库中文规范名；展示 i18n 见 service.localizeOpsTypeDisplayName）。
var defaultOpsTypeNames = []string{
	"数据修改",
	"数据提取",
	"服务维护",
	"配置修改",
}

func (uc *OpsTypeUsecase) newOpsType(projectUID, name string) (*OpsType, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	if name == "" || projectUID == "" {
		return nil, fmt.Errorf("ops type name or project is empty")
	}
	return &OpsType{
		UID:        uid,
		Name:       name,
		ProjectUID: projectUID,
	}, nil
}

// InitDefaultOpsTypes 新建项目预置默认运维类型（对标 InitDefaultEnvironmentTags）。
func (uc *OpsTypeUsecase) InitDefaultOpsTypes(ctx context.Context, projectUID, currentUserUID string) (err error) {
	for _, name := range defaultOpsTypeNames {
		err = uc.CreateOpsType(ctx, projectUID, currentUserUID, name)
		if err != nil {
			uc.log.Errorf("create ops type failed: %v", err)
			return fmt.Errorf("create ops type failed: %w", err)
		}
	}
	return nil
}

// ensureDefaultOpsTypes 空字典幂等补种：已存在则跳过；并发下唯一索引冲突后按名复检。
func (uc *OpsTypeUsecase) ensureDefaultOpsTypes(ctx context.Context, projectUID string) error {
	for _, name := range defaultOpsTypeNames {
		exist, _, err := uc.opsTypeRepo.GetOpsTypeByName(ctx, projectUID, name)
		if err != nil {
			uc.log.Errorf("get ops type by name failed: %v", err)
			return err
		}
		if exist {
			continue
		}
		opsType, err := uc.newOpsType(projectUID, name)
		if err != nil {
			uc.log.Errorf("new ops type failed: %v", err)
			return err
		}
		if err = uc.opsTypeRepo.CreateOpsType(ctx, opsType); err != nil {
			existAfter, _, getErr := uc.opsTypeRepo.GetOpsTypeByName(ctx, projectUID, name)
			if getErr == nil && existAfter {
				continue
			}
			uc.log.Errorf("create ops type failed: %v", err)
			return err
		}
	}
	return nil
}

func (uc *OpsTypeUsecase) CreateOpsType(ctx context.Context, projectUID, currentUserUID, name string) error {
	name = strings.TrimSpace(name)

	if err := uc.projectUsecase.isProjectActive(ctx, projectUID); err != nil {
		return fmt.Errorf("create ops type error: %v", err)
	}

	if canOpProject, err := uc.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUID, projectUID, false); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	exist, _, err := uc.GetOpsTypeByName(ctx, projectUID, name)
	if err != nil {
		uc.log.Errorf("get ops type by name failed: %v", err)
		return err
	}
	if exist {
		return fmt.Errorf("the ops type %s already exists in the current project", name)
	}

	opsType, err := uc.newOpsType(projectUID, name)
	if err != nil {
		uc.log.Errorf("new ops type failed: %v", err)
		return err
	}
	if err = uc.opsTypeRepo.CreateOpsType(ctx, opsType); err != nil {
		uc.log.Errorf("create ops type failed: %v", err)
		return err
	}
	return nil
}

func (uc *OpsTypeUsecase) UpdateOpsType(ctx context.Context, projectUID, currentUserUID, opsTypeUID, name string) error {
	name = strings.TrimSpace(name)

	if err := uc.projectUsecase.isProjectActive(ctx, projectUID); err != nil {
		return fmt.Errorf("update ops type error: %v", err)
	}

	if canOpProject, err := uc.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUID, projectUID, false); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	if opsTypeUID == "" || name == "" {
		return fmt.Errorf("ops type name or uid is empty, please check: id %v, name %v", opsTypeUID, name)
	}

	if _, err := uc.opsTypeRepo.GetOpsTypeByUID(ctx, opsTypeUID); err != nil {
		uc.log.Errorf("get ops type failed: %v", err)
		return err
	}

	exist, existing, err := uc.GetOpsTypeByName(ctx, projectUID, name)
	if err != nil {
		uc.log.Errorf("get ops type by name failed: %v", err)
		return err
	}
	if exist && existing.UID != opsTypeUID {
		return fmt.Errorf("the ops type %s already exists in the current project", name)
	}

	if err = uc.opsTypeRepo.UpdateOpsType(ctx, opsTypeUID, name); err != nil {
		uc.log.Errorf("update ops type failed: %v", err)
		return err
	}
	return nil
}

func (uc *OpsTypeUsecase) DeleteOpsType(ctx context.Context, projectUID, currentUserUID, opsTypeUID string) error {
	if err := uc.projectUsecase.isProjectActive(ctx, projectUID); err != nil {
		return fmt.Errorf("delete ops type error: %v", err)
	}

	if canOpProject, err := uc.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUID, projectUID, false); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	if _, err := uc.opsTypeRepo.GetOpsTypeByUID(ctx, opsTypeUID); err != nil {
		uc.log.Errorf("get ops type failed: %v", err)
		return err
	}
	if err := uc.opsTypeRepo.DeleteOpsType(ctx, opsTypeUID); err != nil {
		uc.log.Errorf("delete ops type failed: %v", err)
		return err
	}
	return nil
}

func (uc *OpsTypeUsecase) ListOpsTypes(ctx context.Context, options *ListOpsTypesOption) ([]*OpsType, int64, error) {
	opsTypes, count, err := uc.opsTypeRepo.ListOpsTypes(ctx, options)
	if err != nil {
		uc.log.Errorf("list ops types failed: %v", err)
		return nil, 0, err
	}
	// 存量空字典：首次读取幂等补种默认 4 项（D9）。
	if count == 0 && options != nil && options.ProjectUID != "" {
		if err = uc.ensureDefaultOpsTypes(ctx, options.ProjectUID); err != nil {
			uc.log.Errorf("ensure default ops types failed: %v", err)
			return nil, 0, err
		}
		opsTypes, count, err = uc.opsTypeRepo.ListOpsTypes(ctx, options)
		if err != nil {
			uc.log.Errorf("list ops types after seed failed: %v", err)
			return nil, 0, err
		}
	}
	return opsTypes, count, nil
}

func (uc *OpsTypeUsecase) GetOpsTypeByName(ctx context.Context, projectUID, name string) (bool, *OpsType, error) {
	name = strings.TrimSpace(name)
	exist, opsType, err := uc.opsTypeRepo.GetOpsTypeByName(ctx, projectUID, name)
	if err != nil {
		uc.log.Errorf("get ops type failed: %v", err)
		return false, nil, err
	}
	return exist, opsType, nil
}

func (uc *OpsTypeUsecase) GetOpsTypeByUID(ctx context.Context, uid string) (*OpsType, error) {
	opsType, err := uc.opsTypeRepo.GetOpsTypeByUID(ctx, uid)
	if err != nil {
		uc.log.Errorf("get ops type failed: %v", err)
		return nil, err
	}
	return opsType, nil
}
