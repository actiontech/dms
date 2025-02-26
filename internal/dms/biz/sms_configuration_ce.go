//go:build !enterprise

package biz

import (
	"context"
	"errors"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/dms/storage/model"
)

var errNotSupportSms = errors.New("sms related functions are enterprise version functions")

func (d *SmsConfigurationUseCase) UpdateSmsConfiguration(ctx context.Context, enable *bool, url *string, smsType *string, configuration *map[string]string) error {

	return errNotSupportSms
}

func (d *SmsConfigurationUseCase) TestSmsConfiguration(ctx context.Context, recipientPhone string) error {
	return errNotSupportSms
}

func (d *SmsConfigurationUseCase) GetSmsConfiguration(ctx context.Context) (smsConfiguration *model.SmsConfiguration, exist bool, err error) {
	return nil, false, errNotSupportSms
}

func (d *SmsConfigurationUseCase) SendSmsCode(ctx context.Context, username string) (reply *dmsV1.SendSmsCodeReply, err error) {
	return nil, errNotSupportSms
}

func (d *SmsConfigurationUseCase) VerifySmsCode(code string, username string) (reply *dmsV1.VerifySmsCodeReply) {
	return nil
}

