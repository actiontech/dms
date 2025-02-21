package storage

import (
	"context"
	"fmt"
	"github.com/actiontech/dms/internal/dms/biz"
	"github.com/actiontech/dms/internal/dms/storage/model"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.SmsConfigurationRepo = (*SmsConfigurationRepo)(nil)

type SmsConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewSmsConfigurationRepo(log utilLog.Logger, s *Storage) *SmsConfigurationRepo {
	return &SmsConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.sms_configuration"))}
}

func (d *SmsConfigurationRepo) UpdateSmsConfiguration(ctx context.Context, smsConfiguration *model.SmsConfiguration) error {
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(smsConfiguration).Where("uid = ?", smsConfiguration.UID).Omit("created_at").Save(smsConfiguration).Error; err != nil {
			return fmt.Errorf("failed to save webhook configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *SmsConfigurationRepo) GetLastSmsConfiguration(ctx context.Context) (*model.SmsConfiguration, error) {
	var smsConfiguration *model.SmsConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&smsConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get sms configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return smsConfiguration, nil
}