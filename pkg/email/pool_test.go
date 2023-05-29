package email

import (
	"net/smtp"
	"testing"
	"time"

	"github.com/jordan-wright/email"
	"golang.org/x/sync/errgroup"
)

func TestMail(t *testing.T) {
	eg := errgroup.Group{}
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			e := email.NewEmail()
			e.From = "notifyGo <648646891@qq.com>"
			e.To = []string{"ch_haokun@163.com"}
			e.Bcc = []string{"hooko@tju.edu.cn"}
			e.Cc = []string{"hookokoko@126.com"}
			e.Subject = "Awesome Subject"
			e.Text = []byte("Text Body is, of course, supported!")
			e.HTML = []byte("<h1>Fancy HTML is supported, too!</h1>")
			//err := e.Send("smtp.qq.com:25",
			//	smtp.PlainAuth("", "648646891@qq.com", "mmlryfcwupktbehd", "smtp.qq.com"))
			err := e.Send("smtp.gmail.com:587",
				smtp.PlainAuth("", "648646891@qq.com", "mmlryfcwupktbehd", "smtp.gmail.com"))
			return err
		})
	}
	err := eg.Wait()
	t.Log(err)
}

func TestPoolMail(t *testing.T) {
	e := email.NewEmail()
	e.From = "648646891@qq.com"
	eg := errgroup.Group{}
	p, err := email.NewPool(
		"smtp.gmail.com:587",
		1,
		smtp.PlainAuth("", "test@gmail.com", "password123", "smtp.gmail.com"))

	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			err = p.Send(e, time.Hour)
			return err
		})
	}

	err = eg.Wait()
	t.Log(err)
}

func TestNewPool(t *testing.T) {
	type args struct {
		opt *Options
	}
	tests := []struct {
		name string
		addr string
		auth smtp.Auth
		args args
		want *Pool
	}{
		{
			name: "新建pool，不创建空闲连接",
			args: args{opt: &Options{
				PoolSize: 3,
			}},
		},
		//{
		//	name: "新建pool，创建空闲连接",
		//	addr: "smtp.gmail.com:587",
		//	auth: smtp.PlainAuth("", "648646891@qq.com", "mmlryfcwupktbehd", "smtp.gmail.com"),
		//	args: args{opt: &Options{
		//		PoolSize:     3,
		//		MinIdleConns: 3,
		//	}},
		//},
		{
			name: "qq email",
			addr: "smtp.qq.com:25",
			auth: smtp.PlainAuth("", "648646891@qq.com", "mmlryfcwupktbehd", "smtp.qq.com"),
			args: args{opt: &Options{
				PoolSize:     3,
				MinIdleConns: 3,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool(tt.addr, tt.auth, tt.args.opt)
			time.Sleep(3 * time.Second) // 保证统计的时候，执行创建pool的goroutine已经执行完毕
			t.Logf("%+v\n", p.Stats())  // 能否再加上当前goroutine的统计？
		})
	}
}
