package biz

import (
	"context"
	"fmt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type BusinessTagRepo interface {
	CreateBusinessTag(ctx context.Context, businessTag *BusinessTag) error
	UpdateBusinessTag(ctx context.Context, businessTagName, businessTagUID string) error
	DeleteBusinessTag(ctx context.Context, businessTagUID string) error
	GetBusinessTagByName(ctx context.Context, name string) (*BusinessTag, error)
	GetBusinessTagByUID(ctx context.Context, uid string) (*BusinessTag, error)
	ListBusinessTags(ctx context.Context, options *ListBusinessTagsOption) ([]*BusinessTag, int64, error)
}

type BusinessTagUsecase struct {
	businessTagRepo BusinessTagRepo
	log             *utilLog.Helper
}

func NewBusinessTagUsecase(businessTagRepo BusinessTagRepo, logger utilLog.Logger) *BusinessTagUsecase {
	return &BusinessTagUsecase{
		businessTagRepo: businessTagRepo,
		log:             utilLog.NewHelper(logger, utilLog.WithMessageKey("biz.business_tag")),
	}
}

type BusinessTag struct {
	UID  string
	Name string
}

func (uc *BusinessTagUsecase) newBusinessTag(tagName string) (*BusinessTag, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	if tagName == "" {
		return nil, fmt.Errorf("business tag name is empty")
	}
	return &BusinessTag{
		UID:  uid,
		Name: tagName,
	}, nil
}

func (uc *BusinessTagUsecase) CreateBusinessTag(ctx context.Context, tagName string) error {
	businessTag, err := uc.newBusinessTag(tagName)
	if err != nil {
		uc.log.Errorf("new business tag failed: %v", err)
		return err
	}
	err = uc.businessTagRepo.CreateBusinessTag(ctx, businessTag)
	if err != nil {
		uc.log.Errorf("create business tag failed: %v", err)
		return err
	}
	return nil
}

func (uc *BusinessTagUsecase) UpdateBusinessTag(ctx context.Context, businessTagUID, businessTagName string) error {
	if businessTagUID == "" || businessTagName == "" {
		return fmt.Errorf("business tag name or uid is empty, please check: businessTagID %v, businessTagName %v", businessTagUID, businessTagName)
	}
	_, err := uc.businessTagRepo.GetBusinessTagByUID(ctx, businessTagUID)
	if err != nil {
		uc.log.Errorf("get business tag failed: %v", err)
		return err
	}
	err = uc.businessTagRepo.UpdateBusinessTag(ctx, businessTagUID, businessTagName)
	if err != nil {
		uc.log.Errorf("update business tag failed: %v", err)
		return err
	}
	return nil
}

func (uc *BusinessTagUsecase) DeleteBusinessTag(ctx context.Context, businessTagUID string) error {
	_, err := uc.businessTagRepo.GetBusinessTagByUID(ctx, businessTagUID)
	if err != nil {
		uc.log.Errorf("get business tag failed: %v", err)
		return err
	}
	err = uc.businessTagRepo.DeleteBusinessTag(ctx, businessTagUID)
	if err != nil {
		uc.log.Errorf("delete business tag failed: %v", err)
		return err
	}
	return nil
}

type ListBusinessTagsOption struct {
	Limit  int
	Offset int
}

func (uc *BusinessTagUsecase) ListBusinessTags(ctx context.Context, options *ListBusinessTagsOption) ([]*BusinessTag, int64, error) {
	businessTags, count, err := uc.businessTagRepo.ListBusinessTags(ctx, options)
	if err != nil {
		uc.log.Errorf("list business tags failed: %v", err)
		return nil, 0, err
	}
	return businessTags, count, nil
}

func (uc *BusinessTagUsecase) GetBusinessTagByName(ctx context.Context, tagName string) (*BusinessTag, error) {
	businessTag, err := uc.businessTagRepo.GetBusinessTagByName(ctx, tagName)
	if err != nil {
		uc.log.Errorf("get business tag failed: %v", err)
		return nil, err
	}
	return businessTag, nil
}

func (uc *BusinessTagUsecase) GetBusinessTagByUID(ctx context.Context, uid string) (*BusinessTag, error) {
	businessTag, err := uc.businessTagRepo.GetBusinessTagByUID(ctx, uid)
	if err != nil {
		uc.log.Errorf("get business tag failed: %v", err)
		return nil, err
	}
	return businessTag, nil
}

// LoadBusinessTagForProjects 根据 UID 和名称补全项目的所属业务标签。
// 对于每个项目，如果 BusinessTag 的 Name 为空但 UID 不为空，则通过 UID 查找并填充 Name。
// 如果 BusinessTag 的 Name 不为空但 UID 为空，则通过 Name 查找并填充 UID。
func (uc *BusinessTagUsecase) LoadBusinessTagForProjects(ctx context.Context, projects []*Project) error {
	businessTags, _, err := uc.businessTagRepo.ListBusinessTags(ctx, &ListBusinessTagsOption{Limit: 9999, Offset: 0})
	if err != nil {
		uc.log.Errorf("list business tags failed: %v", err)
		return err
	}
	businessTagUIDMap := make(map[string]*BusinessTag)
	businessTagNameMap := make(map[string]*BusinessTag)
	for _, businessTag := range businessTags {
		businessTagUIDMap[businessTag.UID] = businessTag
		businessTagNameMap[businessTag.Name] = businessTag
	}
	for _, project := range projects {
		if project.BusinessTag.Name == "" && project.BusinessTag.UID != "" {
			if businessTag, ok := businessTagUIDMap[project.BusinessTag.UID]; ok {
				project.BusinessTag.Name = businessTag.Name
				continue
			}
		}
		if project.BusinessTag.Name != "" && project.BusinessTag.UID == "" {
			if businessTag, ok := businessTagNameMap[project.BusinessTag.Name]; ok {
				project.BusinessTag.UID = businessTag.UID
				continue
			}
		}
	}
	return nil
}
