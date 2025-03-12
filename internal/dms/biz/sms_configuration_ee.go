//go:build enterprise

package biz

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"

	dmsV1 "github.com/actiontech/dms/api/dms/service/v1"
	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

const VerifyCodeKey = "verify_code"
var verifyCodeCache = cache.New(cache.NoExpiration, 10 * time.Minute)

const VerifyCodeTimeSlot = 300 // 300秒 = 5分
var (
	ErrTooFrequentMinute = errors.New("请求过于频繁，请1分钟后再试")
)

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

// GenerateStatelessCode 生成无状态验证码
// key: 用于生成验证码的密钥
// 返回4位数字验证码
func GenerateStatelessCode(key string) string {
	// 获取当前时间戳并按5分钟间隔取整
	timeSlot := time.Now().Unix() / VerifyCodeTimeSlot 

	// 将时间戳转换为字节数组
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeSlot))

	// 使用HMAC-SHA256计算哈希
	h := hmac.New(sha256.New, []byte(key))
	h.Write(timeBytes)
	hash := h.Sum(nil)

	// 使用哈希的最后4个字节生成一个数字
	num := binary.BigEndian.Uint32(hash[len(hash)-4:])
	
	// 取模得到4位数字（1000-9999）
	code := (num % 9000) + 1000

	return fmt.Sprintf("%04d", code)
}

// ValidateStatelessCode 验证无状态验证码
// key: 用于生成验证码的密钥
// inputCode: 用户输入的验证码
// 返回验证是否成功
func ValidateStatelessCode(key, inputCode string) bool {
	// 生成当前时间段的验证码
	currentCode := GenerateStatelessCode(key)
	
	// 检查当前时间段的验证码
	if currentCode == inputCode {
		return true
	}

	// 检查上一个时间段的验证码（处理时间边界情况）
	timeSlot := time.Now().Unix()/VerifyCodeTimeSlot - 1
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeSlot))
	
	h := hmac.New(sha256.New, []byte(key))
	h.Write(timeBytes)
	hash := h.Sum(nil)
	
	prevNum := binary.BigEndian.Uint32(hash[len(hash)-4:])
	prevCode := fmt.Sprintf("%04d", (prevNum%9000)+1000)

	return prevCode == inputCode
}

// checkSmsRateLimit 检查是否可以发送短信
// 如果缓存中存在key，则表示最近已发送过短信，返回错误
// 如果缓存中不存在key，则可以发送短信，并将key存入缓存
func checkSmsRateLimit(key string, phone string) error {
    // 生成缓存key
    cacheKey := fmt.Sprintf("%s:%s", key, phone)
    
    // 检查是否存在未过期的缓存
    if _, found := verifyCodeCache.Get(cacheKey); found {
        return ErrTooFrequentMinute
    }
    
    // 如果不存在缓存，则添加缓存（有效期1分钟）
    verifyCodeCache.Set(cacheKey, true, 1*time.Minute)
    
    return nil
}

// 修改后的发送短信逻辑
func (d *SmsConfigurationUseCase) sendSmsCode(ctx context.Context, phone string) error {
	// 1. 检查发送频率
	if err := checkSmsRateLimit(VerifyCodeKey, phone); err != nil {
		return err
	}

	// 2. 生成4位随机验证码
	code := GenerateStatelessCode(VerifyCodeKey + phone) // 加入phone作为额外的熵源

	// 3. 获取短信配置
	smsConfig, exist, err := d.GetSmsConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("get sms configuration failed: %w", err)
	}
	if !exist || !smsConfig.Enable {
		return fmt.Errorf("sms service is not configured or disabled")
	}

	// 4. 创建SMS客户端
	smsClient, err := NewSmsClient(smsConfig.Url, smsConfig.Configuration, d.log)
	if err != nil {
		return fmt.Errorf("create sms client failed: %w", err)
	}

	// 5. 发送验证码
	if err := smsClient.SendCode(ctx, phone, code); err != nil {
		return fmt.Errorf("send sms failed: %w", err)
	}

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

// 验证短信验证码
func (d *SmsConfigurationUseCase) VerifySmsCode(  inputCode,username string) 	*dmsV1.VerifySmsCodeReply {
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

	// 使用相同的key和phone生成验证码进行比对
	if !ValidateStatelessCode(VerifyCodeKey+user.Phone, inputCode) {
		return &dmsV1.VerifySmsCodeReply{
			Data: dmsV1.VerifySmsCodeReplyData{
				IsVerifyNormally:   false,
				VerifyErrorMessage: "验证码错误或者已过期",
			},
		}
	}
	return &dmsV1.VerifySmsCodeReply{
		Data: dmsV1.VerifySmsCodeReplyData{
			IsVerifyNormally:   true,
			VerifyErrorMessage: "",
		},
	}
}
