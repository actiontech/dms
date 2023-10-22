package biz

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/actiontech/dms/pkg/im/feishu"

	larkIm "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"gopkg.in/chanxuehong/wechat.v1/corp"
	"gopkg.in/chanxuehong/wechat.v1/corp/message/send"
	"gopkg.in/gomail.v2"
)

const (
	TestNotificationSubject     = "DMS notification test"
	TestNotificationBody        = "This is a DMS test notification\nIf you receive this message, it only means that the message can be pushed"
	TestNotificationWebhookBody = `{"msg_type":"text","content":{"text":"DMS notification test\n This is a DMS test notification\nIf you receive this message, it only means that the message can be pushed"}}`
)

var Notifiers = []Notifier{}

type Notifier interface {
	Notify(ctx context.Context, notificationSubject, notificationBody string, users []*User) error
}

func Init(smtp *SMTPConfigurationUsecase, wechat *WeChatConfigurationUsecase, im *IMConfigurationUsecase) {
	Notifiers = append(Notifiers, &EmailNotifier{uc: smtp}, &WeChatNotifier{uc: wechat}, &FeishuNotifier{uc: im})
}

type EmailNotifier struct {
	uc *SMTPConfigurationUsecase
}

func (n *EmailNotifier) Notify(ctx context.Context, notificationSubject, notificationBody string, users []*User) error {
	if len(users) == 0 {
		return nil
	}

	var emails []string
	for _, user := range users {
		if user.Email != "" {
			emails = append(emails, user.Email)
		}
	}

	// no user has configured email, don't send.
	if len(emails) == 0 {
		return nil
	}

	smtpC, exist, err := n.uc.GetSMTPConfiguration(ctx)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	if !smtpC.EnableSMTPNotify {
		return nil
	}

	message := gomail.NewMessage()
	message.SetHeader("From", smtpC.Username)
	message.SetHeader("To", emails...)
	message.SetHeader("Subject", notificationSubject)
	message.SetBody("text/html", strings.Replace(notificationBody, "\n", "<br/>\n", -1))

	port, _ := strconv.Atoi(smtpC.Port)
	dialer := gomail.NewDialer(smtpC.Host, port, smtpC.Username, smtpC.Password)
	if smtpC.IsSkipVerify {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("send email to %v error: %v", emails, err)
	}
	return nil
}

type WeChatNotifier struct {
	uc *WeChatConfigurationUsecase
}

func (w *WeChatNotifier) Notify(ctx context.Context, notificationSubject, notificationBody string, users []*User) error {
	// workflow has been finished.
	if len(users) == 0 {
		return nil
	}
	wechatUsers := map[string] /*user name*/ string /*wechat id*/ {}
	for _, user := range users {
		if user.WxID != "" {
			wechatUsers[user.Name] = user.WxID
		}
	}

	// no user has configured wechat, don't send.
	if len(wechatUsers) == 0 {
		return nil
	}

	wechatC, exist, err := w.uc.GetWeChatConfiguration(ctx)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	if !wechatC.EnableWeChatNotify {
		return nil
	}

	client := generateWeChatClient(wechatC)
	safe := 0
	if wechatC.SafeEnabled {
		safe = 1
	}
	errs := []string{}
	for name, id := range wechatUsers {
		req := &send.Text{
			MessageHeader: send.MessageHeader{
				ToUser:  id,
				MsgType: "text",
				AgentId: int64(wechatC.AgentID),
				Safe:    &safe,
			},
		}
		req.Text.Content = fmt.Sprintf("%v \n\n %v", notificationSubject, notificationBody)
		_, err := client.SendText(req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("send message to %v failed, error: %v", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%v", strings.Join(errs, "\n"))
	}
	return nil
}

func generateWeChatClient(conf *WeChatConfiguration) *send.Client {
	proxy := http.ProxyFromEnvironment
	if conf.ProxyIP != "" {
		proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(conf.ProxyIP)
		}
	}
	var transport http.RoundTripper = &http.Transport{
		Proxy: proxy,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	accessTokenServer := corp.NewDefaultAccessTokenServer(conf.CorpID, conf.CorpSecret, httpClient)
	return send.NewClient(accessTokenServer, httpClient)
}

type FeishuNotifier struct {
	uc *IMConfigurationUsecase
}

func (f *FeishuNotifier) Notify(ctx context.Context, notificationSubject, notificationBody string, users []*User) error {
	// workflow has been finished.
	if len(users) == 0 {
		return nil
	}

	cfg, exist, err := f.uc.GetIMConfiguration(ctx, ImTypeFeishu)
	if err != nil {
		return fmt.Errorf("get im config failed: %v", err)
	}
	if !exist {
		return nil
	}

	if !cfg.IsEnable {
		return nil
	}

	// 通过邮箱、手机从飞书获取用户ids
	var emails, mobiles []string
	userCount := 0
	for _, u := range users {
		if u.Email == "" && u.Phone == "" {
			continue
		}
		if u.Email != "" {
			emails = append(emails, u.Email)
		}
		if u.Phone != "" {
			mobiles = append(mobiles, u.Phone)
		}
		userCount++
		if userCount == feishu.MaxCountOfIdThatUsedToFindUser {
			break
		}
	}

	client := feishu.NewFeishuClient(cfg.AppKey, cfg.AppSecret)
	feishuUsers, err := client.GetUsersByEmailOrMobileWithLimitation(emails, mobiles)
	if err != nil {
		return fmt.Errorf("get user_ids from feishu failed: %v", err)
	}
	if len(feishuUsers) == 0 {
		return fmt.Errorf("there is no notify user find from feishu")
	}

	content, err := BuildFeishuMessageBody(notificationSubject, notificationBody)
	if err != nil {
		return fmt.Errorf("convert content failed: %v", err)
	}
	errMsgs := []string{}
	for id, u := range feishuUsers {
		f.uc.log.Infof("send message to feishu, email=%v, phone=%v, userId=%v", u.Email, u.Mobile, id)
		if err = client.SendMessage(feishu.FeishuReceiverIdTypeUserId, id, feishu.FeishuSendMessageMsgTypePost, content); err != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("send message to feishu failed: %v; email=%v; phone=%v", err, u.Email, u.Mobile))
		}
	}
	if len(errMsgs) > 0 {
		return fmt.Errorf(strings.Join(errMsgs, "\n"))
	}
	return nil
}

func BuildFeishuMessageBody(notificationSubject, notificationBody string) (string, error) {
	zhCnPostText := &larkIm.MessagePostText{Text: notificationBody}
	zhCnMessagePostContent := &larkIm.MessagePostContent{Title: notificationSubject, Content: [][]larkIm.MessagePostElement{{zhCnPostText}}}
	messagePostText := &larkIm.MessagePost{ZhCN: zhCnMessagePostContent}
	content, err := messagePostText.String()
	if err != nil {
		return "", err
	}
	return content, nil
}
