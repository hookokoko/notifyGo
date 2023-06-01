package email

import (
	"context"
	"net/smtp"
	"sync"
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
		{
			name: "qq email",
			addr: "smtp.qq.com:25",
			auth: smtp.PlainAuth("", "648646891@qq.com", "", "smtp.qq.com"),
			args: args{opt: &Options{
				PoolSize:     3,
				MinIdleConns: 2,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool(tt.addr, tt.auth, tt.args.opt)
			//time.Sleep(3 * time.Second) // 保证统计的时候，执行创建pool的goroutine已经执行完毕
			t.Logf("%+v\n", p.Stats()) // 能否再加上当前goroutine的统计？
		})
	}
}

func TestPool_Get(t *testing.T) {
	var p *Pool
	tests := []struct {
		name    string
		wantErr bool
		before  func()
		after   func()
	}{
		{
			name: "basic",
			before: func() {
				p = NewPool(
					"smtp.qq.com:25",
					smtp.PlainAuth("", "648646891@qq.com",
						"", "smtp.qq.com"),
					&Options{
						PoolSize:     3,
						MinIdleConns: 1,
						PoolTimeout:  30 * time.Second,
						//MaxIdleConns: 1,
					},
				)
			},
			after: func() {
				//time.Sleep(3 * time.Second)
				_ = p.Close()
				t.Logf("%+v\n", p.Stats())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			//time.Sleep(3 * time.Second)
			t.Logf("new: %+v\n", p.Stats())
			wg := sync.WaitGroup{}
			for idx := 0; idx < 20; idx++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					//time.Sleep(1 * time.Second)
					t.Logf("idx[%d] begin: %+v\n", idx, p.Stats())
					err := p.SendMail(context.TODO(), &Email{
						From:    "notifyGo <648646891@qq.com>",
						To:      []string{"ch_haokun@163.com"},
						Subject: "Awesome Subject",
						Text:    []byte("Text Body is, of course, supported!"),
						HTML:    []byte("<h1>Fancy HTML is supported, too!</h1>"),
					})
					if err != nil {
						t.Log(err)
					}
					//time.Sleep(1 * time.Second)
					t.Logf("idx[%d] end: %+v\n", idx, p.Stats())
				}(idx)
			}
			wg.Wait()
			tt.after()
		})
	}
}

//respects max size
//目的：测试 Redis 客户端连接池的连接数量是否受到最大值限制。
//功能：对 Redis 进行 1000 次 Ping 操作，并检查连接池与空闲连接池的连接数是否小于等于 10 个。
//
//respects max size on multi
//目的：测试 Redis 客户端连接池在 Redis 事务中连接数量是否受到最大值限制。
//功能：在 Redis 事务中执行 1000 次 Ping 操作，并检查连接池与空闲连接池的连接数是否小于等于 10 个。
//
//respects max size on pipelines
//目的：测试 Redis 客户端连接池在 Redis 管道中连接数量是否受到最大值限制。
//功能：在 Redis 管道中执行 1000 次 Ping 操作，并检查连接池与空闲连接池的连接数是否小于等于 10 个。
//
//removes broken connections
//目的：测试 Redis 客户端连接池是否会删除已经断开的连接。
//功能：在连接池中获取一个连接，将其设置为“坏连接”，然后进行 Ping 操作，重复进行该操作两次，以检查连接池是否删除了已经断开的连接。
//
//reuses connections
//目的：测试 Redis 客户端连接池是否会复用连接。
//功能：对 Redis 进行 100 次 Ping 操作，然后检查连接池与空闲连接池中的连接数和连接池的统计信息（命中次数、未命中次数、超时次数）是否正确。这个测试用例中也要保证连接池中的 ConnMaxIdleTime 设置得足够大，否则超时的连接将会被删除并新建一条连接，影响测试结果。

func TestPool(t *testing.T) {
	opt := Options{
		PoolSize:        0,
		PoolTimeout:     0,
		MinIdleConns:    0,
		MaxIdleConns:    0,
		ConnMaxIdleTime: 0,
		ConnMaxLifetime: 0,
	}
	t.Run("respects max size", func(t *testing.T) {
		perform(1000, func(id int) {
			val, err := client.Ping(ctx).Result()
			if err != nil {
				t.Errorf("Failed to Ping client, error: %v", err)
			}
			if val != "PONG" {
				t.Errorf("Expected Ping result to be 'PONG', got '%s'", val)
			}
		})

		pool := client.Pool()
		if pool.Len() > 10 || pool.IdleLen() > 10 {
			t.Errorf("Pool size exceeds maximum limit of 10, actual: %d/%d", pool.Len(), pool.IdleLen())
		}
		if pool.Len() != pool.IdleLen() {
			t.Errorf("Pool size does not match idle size, pool: %d, idle: %d", pool.Len(), pool.IdleLen())
		}
	})
}
