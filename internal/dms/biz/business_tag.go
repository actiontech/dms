package biz

import (
	"context"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

type BusinessTagRepo interface {
	CreateBusinessTag(ctx context.Context, businessTag *BusinessTag) error
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
	UID             string
	BusinessTagName string
}

func (uc *BusinessTagUsecase) newBusinessTag(businessTag string) (*BusinessTag, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &BusinessTag{
		UID:             uid,
		BusinessTagName: businessTag,
	}, nil
}

func (uc *BusinessTagUsecase) CreateBusinessTag(ctx context.Context, businessTagName string) error {
	businessTag, err := uc.newBusinessTag(businessTagName)
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
