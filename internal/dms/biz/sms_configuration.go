package biz

import (
	"context"

	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	pkgRand "github.com/actiontech/dms/pkg/rand"
)

func initSmsConfiguration() (*model.SmsConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &model.SmsConfiguration{
		Model: model.Model{
			UID: uid,
		},
	}, nil
}

type SmsConfigurationRepo interface {
	UpdateSmsConfiguration(ctx context.Context, configuration *model.SmsConfiguration) error
	GetLastSmsConfiguration(ctx context.Context) (*model.SmsConfiguration, error)
}

type SmsConfigurationUseCase struct {
	tx          TransactionGenerator
	repo        SmsConfigurationRepo
	userUsecase *UserUsecase
	log         *utilLog.Helper
}

func NewSmsConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo SmsConfigurationRepo, userUsecase *UserUsecase) *SmsConfigurationUseCase {
	return &SmsConfigurationUseCase{
		tx:          tx,
		repo:        repo,
		userUsecase: userUsecase,
		log:         utilLog.NewHelper(log, utilLog.WithMessageKey("biz.webhook_configuration")),
	}
}
