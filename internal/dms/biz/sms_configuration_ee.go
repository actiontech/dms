//go:build enterprise

package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"github.com/actiontech/dms/pkg/dms-common/pkg/log"
	"github.com/patrickmn/go-cache"
)

const VerifyCodeKey = "verify_code"

var verifyCodeCache = cache.New(cache.NoExpiration, 10 * time.Minute)
var verifyCodeExpirationTime = 5 * time.Minute

type SmsClient struct {
	url           string
	configuration map[string]string
	client        *http.Client
	log           *log.Helper
}

func NewSmsClient(url string, configuration []byte, log *log.Helper) (*SmsClient, error) {
	var configMap map[string]string
	if err := json.Unmarshal(configuration, &configMap); err != nil {
		log.Errorf("unmarshal sms configuration failed: %v", err)
		return nil, fmt.Errorf("unmarshal sms configuration failed: %w", err)
	}

	return &SmsClient{
		url:           url,
		configuration: configMap,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}, nil
}

func (c *SmsClient) SendCode(ctx context.Context, phone, verifyCode string) error {
	c.log.Infof("start to send sms code to phone: %s,sms service url: %s", phone, c.url)

	// 构建请求体
	reqBody := map[string]interface{}{
		"phone":       phone,
		"verify_code": verifyCode,
	}

	// 添加配置中的参数
	for k, v := range c.configuration {
		reqBody[k] = v
	}

	// 发送请求
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.log.Errorf("marshal request body failed: %v, body: %v", err, reqBody)
		return fmt.Errorf("marshal request body failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.log.Errorf("create request failed: %v, url: %s", err, c.url)
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Errorf("send sms request failed: %v, url: %s, body: %s", err, c.url, string(jsonData))
		return fmt.Errorf("send sms request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容用于日志记录
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("read response body failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("sms service returned error status: %d, response: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("sms service returned error status: %d", resp.StatusCode)
	}

	c.log.Infof("successfully sent sms code to phone: %s, response: %s", phone, string(respBody))
	return nil
}

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
	if err := d.sendSmsCode(ctx, recipientPhone); err != nil {
		return fmt.Errorf("send sms failed: %w", err)
	}
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

// 核心发送短信逻辑
func (d *SmsConfigurationUseCase) sendSmsCode(ctx context.Context, phone string) error {

	// 1. 生成4位随机验证码
	code := fmt.Sprintf("%04d", rand.Intn(9000)+1000)

	// 2. 获取短信配置
	smsConfig, exist, err := d.GetSmsConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("get sms configuration failed: %w", err)
	}
	if !exist || !smsConfig.Enable {
		return fmt.Errorf("sms service is not configured or disabled")
	}

	// 3. 创建SMS客户端
	smsClient, err := NewSmsClient(smsConfig.Url, smsConfig.Configuration, d.log)
	if err != nil {
		return fmt.Errorf("create sms client failed: %w", err)
	}

	// 4. 发送验证码
	if err := smsClient.SendCode(ctx, phone, code); err != nil {
		return fmt.Errorf("send sms failed: %w", err)
	}
	// 5. 缓存验证码
	verifyCodeCache.Set(fmt.Sprintf("%s:%s", VerifyCodeKey, phone), code, verifyCodeExpirationTime)
	return nil
}

func (d *SmsConfigurationUseCase) SendSmsCode(ctx context.Context, username string) (*dmsV1.SendSmsCodeReply, error) {
	d.log.Infof("send sms code to user: %s", username)

	// 1. 获取用户电话
	user, exist, err := d.userUsecase.GetUserByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	if !exist {
		return nil, fmt.Errorf("user not exist")
	}
	phone := user.Phone

	// 2. 发送短信
	if err := d.sendSmsCode(ctx, phone); err != nil {
		return nil, err
	}

	return &dmsV1.SendSmsCodeReply{
		Data: dmsV1.SendSmsCodeReplyData{
			IsSmsCodeSentNormally: true,
		},
	}, nil
}

func (d *SmsConfigurationUseCase) VerifySmsCode(code string, username string) *dmsV1.VerifySmsCodeReply {
	d.log.Infof("start to verify sms code for user: %s", username)

	// 1. 获取用户信息
	user, exist, err := d.userUsecase.GetUserByName(context.TODO(), username)
	if err != nil {
		d.log.Errorf("failed to get user info: %v", err)
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally:   false,
				VerifyErrorMessage: fmt.Sprintf("get user failed: %v", err),
			},
		}
	}
	if !exist {
		d.log.Warnf("user not found: %s", username)
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally:   false,
				VerifyErrorMessage: "user not exist",
			},
		}
	}

	// 2. 从缓存获取验证码
	cacheKey := fmt.Sprintf("%s:%s", VerifyCodeKey, user.Phone)
	verifyCodeInCache, exist := verifyCodeCache.Get(cacheKey)
	if !exist {
		d.log.Warnf("verify code expired or not found for phone: %s", user.Phone)
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally:   false,
				VerifyErrorMessage: "验证码已过期",
			},
		}
	}

	// 3. 验证码比对
	if verifyCodeInCache == code {
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally: true,
			},
		}
	}

	d.log.Warnf("verify code failed for user: %s, code not match", username)
	return &dmsV1.VerifySmsCodeReply{
		Data: dmsV1.VerifySmsCodeReplyData{
			IsVerifyNormally:   false,
			VerifyErrorMessage: "验证码错误",
		},
	}
}
