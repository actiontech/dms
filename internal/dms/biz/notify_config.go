package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/actiontech/dms/internal/dms/storage/model"
	"io"
	"net/http"
	"time"

	pkgErr "github.com/actiontech/dms/internal/dms/pkg/errors"
	pkgRand "github.com/actiontech/dms/pkg/rand"
	"github.com/actiontech/dms/pkg/retry"

	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type SMTPConfiguration struct {
	Base

	UID              string
	EnableSMTPNotify bool
	Host             string
	Port             string
	Username         string
	Password         string
	IsSkipVerify     bool
}

func initSMTPConfiguration() (*SMTPConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &SMTPConfiguration{
		UID: uid,
	}, nil
}

type SMTPConfigurationRepo interface {
	UpdateSMTPConfiguration(ctx context.Context, configuration *SMTPConfiguration) error
	GetLastSMTPConfiguration(ctx context.Context) (*SMTPConfiguration, error)
}

type SMTPConfigurationUsecase struct {
	tx   TransactionGenerator
	repo SMTPConfigurationRepo
	log  *utilLog.Helper
}

func NewSMTPConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo SMTPConfigurationRepo) *SMTPConfigurationUsecase {
	return &SMTPConfigurationUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.smtp_configuration")),
	}
}

func (d *SMTPConfigurationUsecase) UpdateSMTPConfiguration(ctx context.Context, host, port, userName, password *string, enableSMTPNotify, isSkipVerify *bool) error {
	smtpC, err := d.repo.GetLastSMTPConfiguration(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到smtp配置,默认生成一个带uid的配置
		smtpC, err = initSMTPConfiguration()
		if err != nil {
			return err
		}
	}
	// patch smtp config
	{
		if host != nil {
			smtpC.Host = *host
		}

		if port != nil {
			smtpC.Port = *port
		}

		if userName != nil {
			smtpC.Username = *userName
		}

		if password != nil {
			smtpC.Password = *password
		}

		if enableSMTPNotify != nil {
			smtpC.EnableSMTPNotify = *enableSMTPNotify
		}

		if isSkipVerify != nil {
			smtpC.IsSkipVerify = *isSkipVerify
		}
	}

	return d.repo.UpdateSMTPConfiguration(ctx, smtpC)
}

func (d *SMTPConfigurationUsecase) TestSMTPConfiguration(ctx context.Context, recipientAddr string) error {
	smtpC, exits, err := d.GetSMTPConfiguration(ctx)
	if err != nil {
		return err
	}
	if !exits {
		return fmt.Errorf("SMTP is not configured")
	}

	if !smtpC.EnableSMTPNotify {
		return fmt.Errorf("SMTP notice is not enabled")
	}

	notifier := &EmailNotifier{uc: d}
	err = notifier.Notify(ctx, TestNotificationSubject, TestNotificationBody, []*User{
		{
			Email: recipientAddr,
		},
	})
	return err
}

func (d *SMTPConfigurationUsecase) GetSMTPConfiguration(ctx context.Context) (smtpc *SMTPConfiguration, exist bool, err error) {
	smtpC, err := d.repo.GetLastSMTPConfiguration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return smtpC, true, nil
}

type WeChatConfiguration struct {
	Base

	UID                string
	EnableWeChatNotify bool
	CorpID             string
	CorpSecret         string
	AgentID            int
	SafeEnabled        bool
	ProxyIP            string
}

func initWeChatConfiguration() (*WeChatConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &WeChatConfiguration{
		UID: uid,
	}, nil
}

type WeChatConfigurationRepo interface {
	UpdateWeChatConfiguration(ctx context.Context, configuration *WeChatConfiguration) error
	GetLastWeChatConfiguration(ctx context.Context) (*WeChatConfiguration, error)
}

type WeChatConfigurationUsecase struct {
	tx   TransactionGenerator
	repo WeChatConfigurationRepo
	log  *utilLog.Helper
}

func NewWeChatConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo WeChatConfigurationRepo) *WeChatConfigurationUsecase {
	return &WeChatConfigurationUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.wechat_configuration")),
	}
}
func (d *WeChatConfigurationUsecase) GetWeChatConfiguration(ctx context.Context) (wechatc *WeChatConfiguration, exist bool, err error) {
	wechatC, err := d.repo.GetLastWeChatConfiguration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return wechatC, true, nil
}

func (d *WeChatConfigurationUsecase) UpdateWeChatConfiguration(ctx context.Context, enableWeChatNotify, safeEnabled *bool, agentID *int, corpID, corpSecret, proxyIP *string) error {
	wechatC, err := d.repo.GetLastWeChatConfiguration(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到wechat配置,默认生成一个带uid的配置
		wechatC, err = initWeChatConfiguration()
		if err != nil {
			return err
		}
	}
	if enableWeChatNotify != nil {
		wechatC.EnableWeChatNotify = *enableWeChatNotify
	}

	if corpID != nil {
		wechatC.CorpID = *corpID
	}
	if corpSecret != nil {
		wechatC.CorpSecret = *corpSecret
	}
	if agentID != nil {
		wechatC.AgentID = *agentID
	}
	if proxyIP != nil {
		wechatC.ProxyIP = *proxyIP
	}

	if safeEnabled != nil {
		wechatC.SafeEnabled = *safeEnabled
	}

	return d.repo.UpdateWeChatConfiguration(ctx, wechatC)
}

func (d *WeChatConfigurationUsecase) TestWeChatConfiguration(ctx context.Context, wechatID string) error {
	smtpC, exits, err := d.GetWeChatConfiguration(ctx)
	if err != nil {
		return err
	}
	if !exits {
		return fmt.Errorf("WeChat is not configured")
	}

	if !smtpC.EnableWeChatNotify {
		return fmt.Errorf("WeChat notice is not enabled")
	}

	notifier := &WeChatNotifier{uc: d}
	err = notifier.Notify(ctx, TestNotificationSubject, TestNotificationBody, []*User{
		{
			Name: wechatID,
			WxID: wechatID,
		},
	})
	return err
}

type ImType string

const (
	// ImTypeDingTalk = "dingTalk"
	ImTypeFeishu = "feishu"
	ImTypeUnknow = "unknow"
)

// Instant messaging config
type IMConfiguration struct {
	Base

	UID         string
	AppKey      string
	AppSecret   string
	IsEnable    bool
	ProcessCode string
	Type        ImType
}

func initIMConfiguration(im_type ImType) (*IMConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &IMConfiguration{
		UID:  uid,
		Type: im_type,
	}, nil
}

type IMConfigurationRepo interface {
	UpdateIMConfiguration(ctx context.Context, configuration *IMConfiguration) error
	GetLastIMConfiguration(ctx context.Context, imType ImType) (*IMConfiguration, error)
}

type IMConfigurationUsecase struct {
	tx   TransactionGenerator
	repo IMConfigurationRepo
	log  *utilLog.Helper
}

func NewIMConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo IMConfigurationRepo) *IMConfigurationUsecase {
	return &IMConfigurationUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.im_configuration")),
	}
}

func (d *IMConfigurationUsecase) UpdateIMConfiguration(ctx context.Context, isFeishuNotificationEnabled *bool, appID, appSecret *string) error {
	feishuC, err := d.repo.GetLastIMConfiguration(ctx, ImTypeFeishu)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到im配置,默认生成一个带uid的配置
		feishuC, err = initIMConfiguration(ImTypeFeishu)
		if err != nil {
			return err
		}
	}

	if appID != nil {
		feishuC.AppKey = *appID
	}
	if appSecret != nil {
		feishuC.AppSecret = *appSecret
	}
	if isFeishuNotificationEnabled != nil {
		feishuC.IsEnable = *isFeishuNotificationEnabled
	}

	return d.repo.UpdateIMConfiguration(ctx, feishuC)
}

func (d *IMConfigurationUsecase) GetIMConfiguration(ctx context.Context, imType ImType) (imc *IMConfiguration, exist bool, err error) {
	imC, err := d.repo.GetLastIMConfiguration(ctx, imType)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return imC, true, nil
}

func (d *IMConfigurationUsecase) TestFeishuConfiguration(ctx context.Context, users []*User) error {
	imC, exits, err := d.GetIMConfiguration(ctx, ImTypeFeishu)
	if err != nil {
		return err
	}
	if !exits {
		return fmt.Errorf("feishu is not configured")
	}

	if !imC.IsEnable {
		return fmt.Errorf("feishu notice is not enabled")
	}

	notifier := &FeishuNotifier{uc: d}
	err = notifier.Notify(ctx, TestNotificationSubject, TestNotificationBody, users)
	return err
}

type WebHookConfiguration struct {
	Base

	UID                  string
	Enable               bool
	MaxRetryTimes        int
	RetryIntervalSeconds int
	Token                string
	URL                  string
}

func initWebHookConfiguration() (*WebHookConfiguration, error) {
	uid, err := pkgRand.GenStrUid()
	if err != nil {
		return nil, err
	}
	return &WebHookConfiguration{
		UID: uid,
	}, nil
}

type WebHookConfigurationRepo interface {
	UpdateWebHookConfiguration(ctx context.Context, configuration *WebHookConfiguration) error
	GetLastWebHookConfiguration(ctx context.Context) (*WebHookConfiguration, error)
}

type WebHookConfigurationUsecase struct {
	tx   TransactionGenerator
	repo WebHookConfigurationRepo
	log  *utilLog.Helper
}

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
	tx   TransactionGenerator
	repo SmsConfigurationRepo
	log  *utilLog.Helper
}

func NewWebHookConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo WebHookConfigurationRepo) *WebHookConfigurationUsecase {
	return &WebHookConfigurationUsecase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.webhook_configuration")),
	}
}

func (d *WebHookConfigurationUsecase) UpdateWebHookConfiguration(ctx context.Context, enable *bool, maxRetryTimes, retryIntervalSeconds *int, token, url *string) error {
	webhookC, err := d.repo.GetLastWebHookConfiguration(ctx)
	if err != nil {
		if !errors.Is(err, pkgErr.ErrStorageNoData) {
			return err
		}
		// 查询不到webhook配置,默认生成一个带uid的配置
		webhookC, err = initWebHookConfiguration()
		if err != nil {
			return err
		}
	}
	if enable != nil {
		webhookC.Enable = *enable
	}
	if maxRetryTimes != nil {
		webhookC.MaxRetryTimes = *maxRetryTimes
	}
	if retryIntervalSeconds != nil {
		webhookC.RetryIntervalSeconds = *retryIntervalSeconds
	}
	if token != nil {
		webhookC.Token = *token
	}
	if url != nil {
		webhookC.URL = *url
	}
	return d.repo.UpdateWebHookConfiguration(ctx, webhookC)
}

func (d *WebHookConfigurationUsecase) GetWebHookConfiguration(ctx context.Context) (webhookc *WebHookConfiguration, exist bool, err error) {
	webhookC, err := d.repo.GetLastWebHookConfiguration(ctx)
	if err != nil {
		if errors.Is(err, pkgErr.ErrStorageNoData) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return webhookC, true, nil
}

func (d *WebHookConfigurationUsecase) TestWebHookConfiguration(ctx context.Context) error {
	webhookC, exits, err := d.GetWebHookConfiguration(ctx)
	if err != nil {
		return err
	}
	if !exits {
		return fmt.Errorf("webhook is not configured")
	}

	if !webhookC.Enable {
		return fmt.Errorf("webhook notice is not enabled")
	}

	return d.webhookSendRequest(ctx, "hello")
}

func NewSmsConfigurationUsecase(log utilLog.Logger, tx TransactionGenerator, repo SmsConfigurationRepo) *SmsConfigurationUseCase {
	return &SmsConfigurationUseCase{
		tx:   tx,
		repo: repo,
		log:  utilLog.NewHelper(log, utilLog.WithMessageKey("biz.webhook_configuration")),
	}
}

func (d *SmsConfigurationUseCase) UpdateSmsConfiguration(ctx context.Context, enable *bool, url *string, smsType *string, configuration *map[string]string) error {
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

func (d *WebHookConfigurationUsecase) SendWebHookMessage(ctx context.Context, triggerEventType string /*TODO validate TriggerEventType*/, message string) error {
	return d.webhookSendRequest(ctx, message)
}

func (d *WebHookConfigurationUsecase) webhookSendRequest(ctx context.Context, message string) (err error) {
	webhookC, exist, err := d.GetWebHookConfiguration(ctx)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	if !webhookC.Enable {
		return fmt.Errorf("webhook notice is not enabled")
	}

	if webhookC.URL == "" {
		return fmt.Errorf("url is missing, please check webhook config")
	}

	req, err := http.NewRequest(http.MethodPost, webhookC.URL, bytes.NewBuffer([]byte(message)))
	if err != nil {
		return
	}
	if webhookC.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", webhookC.Token))
	}

	doneChan := make(chan struct{})
	return retry.Do(func() error {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("response status_code(%v) body(%s)", resp.StatusCode, respBytes)
	}, doneChan,
		retry.Delay(time.Duration(webhookC.RetryIntervalSeconds)*time.Second),
		retry.Attempts(uint(webhookC.MaxRetryTimes)))

}
