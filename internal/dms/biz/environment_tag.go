package biz

import (
	"context"
	"fmt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type EnvironmentTagRepo interface {
	CreateEnvironmentTag(ctx context.Context, environmentTag *EnvironmentTag) error
	UpdateEnvironmentTag(ctx context.Context, environmentTagName, environmentTagUID string) error
	DeleteEnvironmentTag(ctx context.Context, environmentTagUID string) error
	GetEnvironmentTagByName(ctx context.Context, projectUid, name string) (bool, *EnvironmentTag, error)
	GetEnvironmentTagByUID(ctx context.Context, uid string) (*EnvironmentTag, error)
	ListEnvironmentTags(ctx context.Context, options *ListEnvironmentTagsOption) ([]*EnvironmentTag, int64, error)
}

type EnvironmentTagUsecase struct {
	environmentTagRepo        EnvironmentTagRepo
	projectUsecase            *ProjectUsecase
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase
	log                       *utilLog.Helper
}

func NewEnvironmentTagUsecase(environmentTagRepo EnvironmentTagRepo, logger utilLog.Logger,
	projectUsecase *ProjectUsecase,
	opPermissionVerifyUsecase *OpPermissionVerifyUsecase) *EnvironmentTagUsecase {
	return &EnvironmentTagUsecase{
		environmentTagRepo:        environmentTagRepo,
		projectUsecase:            projectUsecase,
		opPermissionVerifyUsecase: opPermissionVerifyUsecase,
		log:                       utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.environment_tag")),
	}
}

type EnvironmentTag struct {
	UID        string
	Name       string
	ProjectUID string
}

func (uc *EnvironmentTagUsecase) newEnvironmentTag(projectUid, tagName string) (*EnvironmentTag, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	if tagName == "" {
		return nil, fmt.Errorf("environment tag name or project is empty")
	}
	return &EnvironmentTag{
		UID:        uid,
		Name:       tagName,
		ProjectUID: projectUid,
	}, nil
}

func (uc *EnvironmentTagUsecase) InitDefaultEnvironmentTags(ctx context.Context, projectUid, currentUserUid string) (err error) {
	defaultEnvironmentTags := []string{
		"DEV",
		"TEST",
		"UAT",
		"PROD",
	}

	for _, environmentTag := range defaultEnvironmentTags {
		err = uc.CreateEnvironmentTag(ctx, projectUid, currentUserUid, environmentTag)
		if err != nil {
			uc.log.Errorf("create environment tag failed: %v", err)
			return fmt.Errorf("create environment tag failed: %w", err)
		}
	}
	return nil
}

func (uc *EnvironmentTagUsecase) CreateEnvironmentTag(ctx context.Context, projectUid, currentUserUid, tagName string) error {
	// 检查项目是否归档/删除
	if err := uc.projectUsecase.isProjectActive(ctx, projectUid); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}

	// 检查当前用户有项目管理员权限
	if canOpProject, err := uc.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}
	// 校验环境标签名称
	exist, _, err := uc.GetEnvironmentTagByName(ctx, projectUid, tagName)
	if err != nil {
		uc.log.Errorf("get environment tag by name failed: %v", err)
		return err
	}
	if exist {
		return fmt.Errorf("the tag %s already exists in the current project", tagName)
	}
	environmentTag, err := uc.newEnvironmentTag(projectUid, tagName)
	if err != nil {
		uc.log.Errorf("new environment tag failed: %v", err)
		return err
	}
	err = uc.environmentTagRepo.CreateEnvironmentTag(ctx, environmentTag)
	if err != nil {
		uc.log.Errorf("create environment tag failed: %v", err)
		return err
	}
	return nil
}

func (uc *EnvironmentTagUsecase) UpdateEnvironmentTag(ctx context.Context, projectUid, currentUserUid, environmentTagUID, environmentTagName string) error {
	// 检查项目是否归档/删除
	if err := uc.projectUsecase.isProjectActive(ctx, projectUid); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}

	// 检查当前用户有项目管理员权限
	if canOpProject, err := uc.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	if environmentTagUID == "" || environmentTagName == "" {
		return fmt.Errorf("environment tag name or uid is empty, please check: id %v, name %v", environmentTagUID, environmentTagName)
	}
	_, err := uc.environmentTagRepo.GetEnvironmentTagByUID(ctx, environmentTagUID)
	if err != nil {
		uc.log.Errorf("get environment tag failed: %v", err)
		return err
	}
	err = uc.environmentTagRepo.UpdateEnvironmentTag(ctx, environmentTagUID, environmentTagName)
	if err != nil {
		uc.log.Errorf("update environment tag failed: %v", err)
		return err
	}
	return nil
}

func (uc *EnvironmentTagUsecase) DeleteEnvironmentTag(ctx context.Context, projectUid, currentUserUid, environmentTagUID string) error {
	// 检查项目是否归档/删除
	if err := uc.projectUsecase.isProjectActive(ctx, projectUid); err != nil {
		return fmt.Errorf("update db service error: %v", err)
	}

	// 检查当前用户有项目管理员权限
	if canOpProject, err := uc.opPermissionVerifyUsecase.CanOpProject(ctx, currentUserUid, projectUid); err != nil {
		return fmt.Errorf("check user is project admin or golobal op permission failed: %v", err)
	} else if !canOpProject {
		return fmt.Errorf("user is not project admin or golobal op permission user")
	}

	_, err := uc.environmentTagRepo.GetEnvironmentTagByUID(ctx, environmentTagUID)
	if err != nil {
		uc.log.Errorf("get environment tag failed: %v", err)
		return err
	}
	err = uc.environmentTagRepo.DeleteEnvironmentTag(ctx, environmentTagUID)
	if err != nil {
		uc.log.Errorf("delete environment tag failed: %v", err)
		return err
	}
	return nil
}

type ListEnvironmentTagsOption struct {
	Limit      int
	Offset     int
	ProjectUID string
}

func (uc *EnvironmentTagUsecase) ListEnvironmentTags(ctx context.Context, options *ListEnvironmentTagsOption) ([]*EnvironmentTag, int64, error) {
	environmentTags, count, err := uc.environmentTagRepo.ListEnvironmentTags(ctx, options)
	if err != nil {
		uc.log.Errorf("list environment tags failed: %v", err)
		return nil, 0, err
	}
	return environmentTags, count, nil
}

func (uc *EnvironmentTagUsecase) GetEnvironmentTagByName(ctx context.Context, projectUid, tagName string) (bool, *EnvironmentTag, error) {
	exist, environmentTag, err := uc.environmentTagRepo.GetEnvironmentTagByName(ctx, projectUid, tagName)
	if err != nil {
		uc.log.Errorf("get environment tag failed: %v", err)
		return false, nil, err
	}
	return exist, environmentTag, nil
}

func (uc *EnvironmentTagUsecase) GetEnvironmentTagByUID(ctx context.Context, uid string) (*EnvironmentTag, error) {
	environmentTag, err := uc.environmentTagRepo.GetEnvironmentTagByUID(ctx, uid)
	if err != nil {
		uc.log.Errorf("get environment tag failed: %v", err)
		return nil, err
	}
	return environmentTag, nil
}
