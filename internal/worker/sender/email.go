package sender

import (
	"context"
	"fmt"
	"net/smtp"
	"notifyGo/internal"

	"github.com/jordan-wright/email"
)

type EmailHandler struct {
	client *email.Email
}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func (eh *EmailHandler) Name() string {
	return internal.EMAILNAME
}

func (eh *EmailHandler) Execute(ctx context.Context, task *internal.Task) (err error) {
	fmt.Printf("send email success, %v\n", *task)

	e := email.NewEmail()
	e.From = "notifyGo <648646891@qq.com>"
	e.To = []string{"ch_haokun@163.com"}
	e.Bcc = []string{"hooko@tju.edu.cn"}
	e.Cc = []string{"hookokoko@126.com"}
	e.Subject = "Awesome Subject"
	e.Text = []byte("Text Body is, of course, supported!")
	e.HTML = []byte("<h1>Fancy HTML is supported, too!</h1>")
	err = e.Send("smtp.qq.com:25",
		smtp.PlainAuth("", "648646891@qq.com", "mmlryfcwupktbehd", "smtp.qq.com"))

	return
}
