package biz

import (
	"context"

	pkgRand "github.com/actiontech/dms/pkg/rand"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type CompanyNotice struct {
	Base

	UID               string
	NoticeStr         string
	ReadByCurrentUser bool
}

func initCompanyNotice() (*CompanyNotice, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &CompanyNotice{
		UID: uid,
	}, nil
}

type CompanyNoticeRepo interface {
	UpdateCompanyNotice(ctx context.Context, configuration *CompanyNotice) error
	GetCompanyNotice(ctx context.Context) (*CompanyNotice, error)
}

type CompanyNoticeUsecase struct {
	tx          TransactionGenerator
	repo        CompanyNoticeRepo
	userUsecase *UserUsecase
	log         *utilLog.Helper
}

func NewCompanyNoticeUsecase(log utilLog.Logger, tx TransactionGenerator, repo CompanyNoticeRepo, userUsecase *UserUsecase) *CompanyNoticeUsecase {
	return &CompanyNoticeUsecase{
		tx:          tx,
		repo:        repo,
		userUsecase: userUsecase,
		log:         utilLog.NewHelper(log, utilLog.WithMessageKey("biz.company_notice")),
	}
}
