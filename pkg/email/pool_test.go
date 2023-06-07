package email

import (
	"context"
	"fmt"
	"net/smtp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 目前想到的测试用例
//1. respects max size
//目的：测试客户端连接池的连接数量是否受到最大值限制。
//功能：进行 1000 次 Ping 操作，并检查连接池与空闲连接池的连接数是否小于等于 10 个。

//2. removes broken connections
//目的：测试客户端连接池是否会删除已经断开的连接。
//功能：在连接池中获取一个连接，将其设置为“坏连接”，然后进行 Ping 操作，重复进行该操作两次，以检查连接池是否删除了已经断开的连接。

//3. reuses connections
//目的：测试客户端连接池是否会复用连接。
//功能：进行 100 次 Ping 操作，然后检查连接池与空闲连接池中的连接数和连接池的统计信息（命中次数、未命中次数、超时次数）是否正确。
//这个测试用例中也要保证连接池中的 ConnMaxIdleTime 设置得足够大，否则超时的连接将会被删除并新建一条连接，影响测试结果。

func TestPool(t *testing.T) {
	t.Run("大量请求，大量等待超时", func(t *testing.T) {
		ctx := context.TODO()
		poolOpt := &Options{
			PoolSize:        500,
			PoolTimeout:     60 * time.Second,
			ConnMaxLifetime: time.Minute, // 距离创建时间多久之后失效标记
			ConnMaxIdleTime: time.Second, // 距离上一次使用时间多久之后标记失效
		}
		client := NewClient(&ClientConfig{
			Addr: "smtp.qq.com:25",
			Auth: smtp.PlainAuth("", "648646891@qq.com",
				"", "smtp.qq.com"),
			Options: poolOpt,
		})

		defer client.Close()

		perform(1000, func(id int) {
			err := client.Ping(ctx)
			if err != nil {
				t.Errorf("[%d]Failed to Ping client, error: %v", id, err)
			}
		})

		fmt.Printf("%+v\n", client.Pool.Stats())
	})

	//创建了一个连接，在连接最大存活时间不限制的情况下能一直复用这个连接
	t.Run("reuse connection", func(t *testing.T) {
		ctx := context.TODO()
		poolOpt := &Options{
			PoolSize:    1,
			PoolTimeout: 30 * time.Second,
			//MinIdleConns:    1,
			ConnMaxLifetime: 0,                // 距离创建时间多久之后失效标记，等于0说明一直不失效
			ConnMaxIdleTime: 10 * time.Second, // 距离上一次使用时间多久之后标记失效
		}
		client := NewClient(&ClientConfig{
			Addr: "smtp.qq.com:25",
			Auth: smtp.PlainAuth("", "648646891@qq.com",
				"", "smtp.qq.com"),
			Options: poolOpt,
		})

		defer client.Close()

		perform(100, func(id int) {
			err := client.Ping(ctx)
			if err != nil {
				t.Errorf("[%d]Failed to Ping client, error: %v", id, err)
			}
		})

		fmt.Printf("%+v\n", client.Pool.Stats())
		fmt.Println(client.Pool.Len())
		fmt.Println(client.Pool.IdleLen())
	})

	// MinIdleConns 比 MaxIdleConns大的情况
	t.Run("test3", func(t *testing.T) {
		ctx := context.TODO()
		poolOpt := &Options{
			PoolSize:        5,
			PoolTimeout:     30 * time.Second,
			MinIdleConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 0,                // 距离创建时间多久之后失效标记，等于0说明一直不失效
			ConnMaxIdleTime: 10 * time.Second, // 距离上一次使用时间多久之后标记失效
		}
		client := NewClient(&ClientConfig{
			Addr: "smtp.qq.com:25",
			Auth: smtp.PlainAuth("", "648646891@qq.com",
				"", "smtp.qq.com"),
			Options: poolOpt,
		})

		defer client.Close()

		perform(100, func(id int) {
			err := client.Ping(ctx)
			if err != nil {
				t.Errorf("[%d]Failed to Ping client, error: %v", id, err)
			}
		})

		fmt.Printf("%+v\n", client.Pool.Stats())
		fmt.Println(client.Pool.Len())
		fmt.Println(client.Pool.IdleLen())
		assert.Equal(t, 3, client.Pool.stats.StaleConns)
	})

	// MinIdleConns==0 和 MaxIdleConns==1的情况
	t.Run("test3", func(t *testing.T) {
		ctx := context.TODO()
		poolOpt := &Options{
			PoolSize:        5,
			PoolTimeout:     30 * time.Second,
			MinIdleConns:    0,
			MaxIdleConns:    1,
			ConnMaxIdleTime: 10 * time.Second, // 距离上一次使用时间多久之后标记失效
		}
		client := NewClient(&ClientConfig{
			Addr: "smtp.qq.com:25",
			Auth: smtp.PlainAuth("", "648646891@qq.com",
				"", "smtp.qq.com"),
			Options: poolOpt,
		})

		defer client.Close()

		perform(100, func(id int) {
			err := client.Ping(ctx)
			if err != nil {
				t.Errorf("[%d]Failed to Ping client, error: %v", id, err)
			}
		})

		fmt.Printf("%+v\n", client.Pool.Stats())
		fmt.Println(client.Pool.Len())
		fmt.Println(client.Pool.IdleLen())
	})
}

func perform(num int, fun func(i int)) {
	wg := &sync.WaitGroup{}
	for i := 0; i < num; i++ {
		newI := i
		go func() {
			defer wg.Done()
			wg.Add(1)
			fun(newI)
		}()
	}
	wg.Wait()
}
