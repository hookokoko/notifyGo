package sender

import (
	"context"
	"net/smtp"
	"notifyGo/internal"
	"notifyGo/pkg/email"
	"time"
)

type EmailHandler struct{}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func (eh *EmailHandler) Name() string {
	return internal.EMAILNAME
}

func (eh *EmailHandler) Execute(ctx context.Context, task *internal.Task) (err error) {
	client := email.NewClient(&email.ClientConfig{
		Addr: "smtp.qq.com:25",
		Auth: smtp.PlainAuth("", "648646891@qq.com",
			"", "smtp.qq.com"),
		Options: &email.Options{
			PoolSize:        5,
			PoolTimeout:     30 * time.Second,
			MinIdleConns:    0,
			MaxIdleConns:    1,
			ConnMaxIdleTime: 10 * time.Second, // 距离上一次使用时间多久之后标记失效
		},
	})

	defer func() { _ = client.Close() }()

	emailCfg := &email.Email{}
	emailCfg.From = "notifyGo <648646891@qq.com>"
	emailCfg.To = []string{task.MsgReceiver.Value()}
	emailCfg.Text = []byte(task.MsgContent.Content)

	err = client.SendMail(ctx, emailCfg)

	return
}
