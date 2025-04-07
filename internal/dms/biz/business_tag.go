package biz

import (
	"context"
	"fmt"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type BusinessTagRepo interface {
	CreateBusinessTag(ctx context.Context, businessTag *BusinessTag) error
	GetBusinessTagByName(ctx context.Context, name string) (*BusinessTag, error)
	ListBusinessTags(ctx context.Context) ([]*BusinessTag, error)
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

func (uc *BusinessTagUsecase) GetBusinessTagByName(ctx context.Context, tagName string) (*BusinessTag, error) {
	businessTag, err := uc.businessTagRepo.GetBusinessTagByName(ctx, tagName)
	if err != nil {
		uc.log.Errorf("get business tag failed: %v", err)
		return nil, err
	}
	return businessTag, nil
}

func (uc *BusinessTagUsecase) LoadBusinessTagForProjects(ctx context.Context, projects []*Project) error {
	businessTags, err := uc.businessTagRepo.ListBusinessTags(ctx)
	if err != nil {
		uc.log.Errorf("list business tags failed: %v", err)
		return err
	}
	businessTagMap := make(map[string]*BusinessTag)
	for _, businessTag := range businessTags {
		businessTagMap[businessTag.UID] = businessTag
	}
	for _, project := range projects {
		if businessTag, ok := businessTagMap[project.BusinessTag.UID]; ok {
			project.BusinessTag.Name = businessTag.Name
		}
	}
	return nil
}
