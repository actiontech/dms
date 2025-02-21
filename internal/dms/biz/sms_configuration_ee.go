//go:build enterprise

package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	"github.com/actiontech/dms/internal/cache"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"math/rand"
	"strconv"
)

const VerifyCodeKey = "verify_code"

func (d *SmsConfigurationUseCase) UpdateSmsConfiguration(ctx context.Context, enable *bool, url *string, smsType *string, configuration *map[string]string) error {
	d.log.Infof("update sms configuration")
	smsConfiguration, err := d.repo.GetLastSmsConfiguration(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到sms配置,默认生成一个带uid的配置
		smsConfiguration, err = initSmsConfiguration()
		if err != nil {
			return err
		}
	}
	if configuration != nil {
		jsonConfiguration, _ := json.Marshal(*configuration)
		smsConfiguration.Configuration = jsonConfiguration
	}
	if url != nil {
		smsConfiguration.Url = *url
	}
	if smsType != nil {
		smsConfiguration.Type = *smsType
	}
	if enable != nil {
		smsConfiguration.Enable = *enable
	}
	return d.repo.UpdateSmsConfiguration(ctx, smsConfiguration)
}

func (d *SmsConfigurationUseCase) TestSmsConfiguration(ctx context.Context, recipientPhone string) error {
	// TODO: 查询sms配置并根据配置发送测试消息到手机
	return nil
}

func (d *SmsConfigurationUseCase) GetSmsConfiguration(ctx context.Context) (smsConfiguration *model.SmsConfiguration, exist bool, err error) {
	smsConfiguration, err = d.repo.GetLastSmsConfiguration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return smsConfiguration, true, nil
}

func (d *SmsConfigurationUseCase) SendSmsCode(ctx context.Context, username string) (reply *dmsV1.SendSmsCodeReply, err error) {
	d.log.Infof("send sms code")
	// 1. 生成4位的随机数
	code := strconv.Itoa(rand.Intn(9000) + 1000)
	// 2. 调用短信接口发送随机数
	_, exist, err := d.GetSmsConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.New("sms configuration not exist")
	}
	// TODO: 等到有真正的短信平台后进行对接, 这里暂时通过SendErrorMessage返回code,代替接收短信
	// 3. 发送成功后将随机数保存到缓存，设置五分钟过期时间。发送失败返回失败原因。
	err = cache.Set(fmt.Sprintf("%s:%s", VerifyCodeKey, username), []byte(code))
	return &dmsV1.SendSmsCodeReply{
		Data: dmsV1.SendSmsCodeReplyData{
			IsSmsCodeSentNormally: true,
			SendErrorMessage: code,
		},
	}, nil
}

func (d *SmsConfigurationUseCase) VerifySmsCode(request *dmsV1.VerifySmsCodeReq, username string) (reply *dmsV1.VerifySmsCodeReply) {
	d.log.Infof("verify sms code")
	verifyCodeBytes, err := cache.Get(fmt.Sprintf("%s:%s", VerifyCodeKey, username))
	if err != nil {
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally: false,
				VerifyErrorMessage: "验证码已过期",
			},
		}
	}
	verifyCodeInCache := string(verifyCodeBytes)
	if verifyCodeInCache == request.Code {
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally: true,
			},
		}
	}
	return &dmsV1.VerifySmsCodeReply{
		Data: dmsV1.VerifySmsCodeReplyData{
			IsVerifyNormally: false,
			VerifyErrorMessage: "验证码错误",
		},
	}
}