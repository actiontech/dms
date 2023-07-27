package storage

import (
	"context"
	"fmt"

	"github.com/actiontech/dms/internal/dms/biz"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"gorm.io/gorm"
)

var _ biz.SMTPConfigurationRepo = (*SMTPConfigurationRepo)(nil)

type SMTPConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewSMTPConfigurationRepo(log utilLog.Logger, s *Storage) *SMTPConfigurationRepo {
	return &SMTPConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.smtp_configuration"))}
}
func (d *SMTPConfigurationRepo) UpdateSMTPConfiguration(ctx context.Context, smtpC *biz.SMTPConfiguration) error {
	model, err := convertBizSMTPConfiguration(smtpC)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz smtp configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save smtp configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *SMTPConfigurationRepo) GetLastSMTPConfiguration(ctx context.Context) (*biz.SMTPConfiguration, error) {
	var smtpConfiguration *model.SMTPConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&smtpConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get smtp configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModeSMTPConfiguration(smtpConfiguration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model smtp configuration: %w", err))
	}
	return ret, nil
}

var _ biz.WeChatConfigurationRepo = (*WeChatConfigurationRepo)(nil)

type WeChatConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewWeChatConfigurationRepo(log utilLog.Logger, s *Storage) *WeChatConfigurationRepo {
	return &WeChatConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.wechat_configuration"))}
}
func (d *WeChatConfigurationRepo) UpdateWeChatConfiguration(ctx context.Context, wechatC *biz.WeChatConfiguration) error {
	model, err := convertBizWeChatConfiguration(wechatC)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz wechat configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save wechat configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *WeChatConfigurationRepo) GetLastWeChatConfiguration(ctx context.Context) (*biz.WeChatConfiguration, error) {
	var wechatConfiguration *model.WeChatConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&wechatConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get wechat configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModeWeChatConfiguration(wechatConfiguration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model wechat configuration: %w", err))
	}
	return ret, nil
}

var _ biz.WebHookConfigurationRepo = (*WebHookConfigurationRepo)(nil)

type WebHookConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewWebHookConfigurationRepo(log utilLog.Logger, s *Storage) *WebHookConfigurationRepo {
	return &WebHookConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.webhook_configuration"))}
}
func (d *WebHookConfigurationRepo) UpdateWebHookConfiguration(ctx context.Context, webhookC *biz.WebHookConfiguration) error {
	model, err := convertBizWebHookConfiguration(webhookC)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz webhook configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save webhook configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *WebHookConfigurationRepo) GetLastWebHookConfiguration(ctx context.Context) (*biz.WebHookConfiguration, error) {
	var webhookConfiguration *model.WebHookConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Last(&webhookConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get webhook configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModeWebHookConfiguration(webhookConfiguration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model webhook configuration: %w", err))
	}
	return ret, nil
}

var _ biz.IMConfigurationRepo = (*IMConfigurationRepo)(nil)

type IMConfigurationRepo struct {
	*Storage
	log *utilLog.Helper
}

func NewIMConfigurationRepo(log utilLog.Logger, s *Storage) *IMConfigurationRepo {
	return &IMConfigurationRepo{Storage: s, log: utilLog.NewHelper(log, utilLog.WithMessageKey("storage.im_configuration"))}
}
func (d *IMConfigurationRepo) UpdateIMConfiguration(ctx context.Context, imC *biz.IMConfiguration) error {
	model, err := convertBizIMConfiguration(imC)
	if err != nil {
		return pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert biz im configuration: %w", err))
	}
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(model).Where("uid = ?", model.UID).Omit("created_at").Save(model).Error; err != nil {
			return fmt.Errorf("failed to save im configuration: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *IMConfigurationRepo) GetLastIMConfiguration(ctx context.Context, imType biz.ImType) (*biz.IMConfiguration, error) {
	var imConfiguration *model.IMConfiguration
	if err := transaction(d.log, ctx, d.db, func(tx *gorm.DB) error {
		if err := tx.Where("`type` = ?", imType).Last(&imConfiguration).Error; err != nil {
			return fmt.Errorf("failed to get im configuration: %v", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	ret, err := convertModeIMConfiguration(imConfiguration)
	if err != nil {
		return nil, pkgErr.WrapStorageErr(d.log, fmt.Errorf("failed to convert model im configuration: %w", err))
	}
	return ret, nil
}
