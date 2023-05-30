package email

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/smtp"
	"sync"
	"sync/atomic"
	"time"
)

/*
现有的go-mail实现连接池 有一些问题
1. 没有Stat()统计方法
2. 没有最大空闲连接的参数
3. 创建连接之后，没有空闲到一定时间自动关闭的功能，即创建之后一直占着连接资源，即使没有发送邮件的请求
*/

var (
	// ErrClosed performs any operation on the closed Client will return this error.
	ErrClosed = errors.New("go-emails: Client is closed")

	// ErrPoolTimeout timed out waiting to get a connection from the connection pool.
	ErrPoolTimeout = errors.New("go-emails: connection pool timeout")
)

// Stats 摘抄自go-reids
type Stats struct {
	// 这些不知道有啥用，暂时先放上
	Hits     uint32 // number of times free connection was found in the pool
	Misses   uint32 // number of times free connection was NOT found in the pool
	Timeouts uint32 // number of times a wait timeout occurred

	TotalConns uint32 // number of total connections in the pool
	IdleConns  uint32 // number of idle connections in the pool
	StaleConns uint32 // number of stale connections removed from the pool
}

type Client struct {
	*smtp.Client
	failCount int
	createdAt time.Time
	usedAt    int64
}

func (cn *Client) UsedAt() time.Time {
	unix := atomic.LoadInt64(&cn.usedAt)
	return time.Unix(unix, 0)
}

func (cn *Client) SetUsedAt(tm time.Time) {
	atomic.StoreInt64(&cn.usedAt, tm.Unix())
}

type Pool struct {
	addr          string
	helloHostname string
	tlsConfig     *tls.Config
	auth          smtp.Auth

	timers sync.Pool

	cfg   *Options
	queue chan struct{}

	connsMu   sync.Mutex
	conns     []*Client // 准确说这是一个smtp的客户端
	idleConns []*Client

	poolSize     int
	idleConnsLen int

	stats Stats

	_closed uint32 // atomic
}

type Options struct {
	PoolSize        int
	PoolTimeout     time.Duration
	MinIdleConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

func NewPool(addr string, auth smtp.Auth, opt *Options) *Pool {
	p := &Pool{
		addr: addr,
		auth: auth,
		cfg:  opt,

		// 创建一个定时器，为Get 等待空闲连接时使用
		timers: sync.Pool{
			New: func() interface{} {
				t := time.NewTimer(time.Hour)
				t.Stop()
				return t
			},
		},

		queue:     make(chan struct{}, opt.PoolSize),
		conns:     make([]*Client, 0, opt.PoolSize),
		idleConns: make([]*Client, 0, opt.PoolSize),
	}

	if host, _, e := net.SplitHostPort(addr); e != nil {
		return nil
	} else {
		p.tlsConfig = &tls.Config{ServerName: host}
	}

	p.connsMu.Lock()
	p.checkMinIdleConns() // 启动时即创建所有的连接，改成按需创建吧。
	p.connsMu.Unlock()

	return p
}

// 疯狂检测连接池空闲连接数量和连接池中的连接数量，不满足就创建
func (p *Pool) checkMinIdleConns() {
	if p.cfg.MinIdleConns == 0 {
		return
	}
	for p.poolSize < p.cfg.PoolSize && p.idleConnsLen < p.cfg.MinIdleConns {
		select {
		case p.queue <- struct{}{}:
			p.poolSize++
			p.idleConnsLen++
			go func() {
				err := p.addIdleConn()
				if err != nil && err != ErrClosed {
					log.Println("error: ", err)
					p.connsMu.Lock()
					p.poolSize--
					p.idleConnsLen--
					p.connsMu.Unlock()
				}
				<-p.queue // freeTurn()
			}()
		default: // 这里意味着p.queue满了，不要阻塞，直接return
			return
		}
	}
}

func (p *Pool) addIdleConn() error {
	cn, err := p.build(context.TODO())
	if err != nil {
		return err
	}

	p.connsMu.Lock()
	defer p.connsMu.Unlock()

	// It is not allowed to add new connections to the closed connection pool.
	if p.closed() {
		_ = cn.Close()
		return ErrClosed
	}

	p.conns = append(p.conns, cn)
	p.idleConns = append(p.idleConns, cn)
	return nil
}

func (p *Pool) build(ctx context.Context) (*Client, error) {
	// 要不要加连接超时
	cl, err := smtp.Dial(p.addr)
	if err != nil {
		return nil, err
	}

	// Is there a custom hostname for doing a HELLO with the SMTP server?
	if p.helloHostname != "" {
		cl.Hello(p.helloHostname)
	}

	c := &Client{Client: cl, createdAt: time.Now()}
	c.SetUsedAt(time.Now())

	if _, err := startTLS(c, p.tlsConfig); err != nil {
		c.Close()
		return nil, err
	}

	if p.auth != nil {
		if _, err := addAuth(c, p.auth); err != nil {
			c.Close()
			return nil, err
		}
	}

	return c, nil
}

func startTLS(c *Client, t *tls.Config) (bool, error) {
	if ok, _ := c.Extension("STARTTLS"); !ok {
		return false, nil
	}

	if err := c.StartTLS(t); err != nil {
		return false, err
	}

	return true, nil
}

func addAuth(c *Client, auth smtp.Auth) (bool, error) {
	if ok, _ := c.Extension("AUTH"); !ok {
		return false, nil
	}

	if err := c.Auth(auth); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Pool) closed() bool {
	return atomic.LoadUint32(&p._closed) == 1
}

func (p *Pool) Get(ctx context.Context) (*Client, error) {
	if p.closed() {
		return nil, ErrClosed
	}

	if err := p.waitTurn(ctx); err != nil {
		return nil, err
	}

	for {
		p.connsMu.Lock()
		// 尝试p.idleConns中拿到一个连接
		cn, err := p.popIdle()
		p.connsMu.Unlock()

		if err != nil {
			return nil, err
		}
		// 没有取到连接，终止循环，在下面的逻辑中新建一个连接
		if cn == nil {
			break
		}
		// 检查连接是否有效，如果无效就关闭它，继续下一轮循环
		// 这里会不会有效率问题
		if !p.isHealthyConn(cn) {
			_ = p.CloseConn(cn)
			continue
		}

		// 当拿到一个健康的可用连接，就增减Hits计数
		atomic.AddUint32(&p.stats.Hits, 1)
		return cn, nil
	}

	atomic.AddUint32(&p.stats.Misses, 1)

	newClient, err := p.build(ctx)
	// 如果获取连接失败，则重新腾出p.queue的一个位置
	// 这里的p.queue中的位置是在waitTurn()中占用的
	if err != nil {
		<-p.queue
		return nil, err
	}

	return newClient, nil
}

func (p *Pool) waitTurn(ctx context.Context) error {
	// queue可以认为是一个大小为pool size的token池，当这个chan还能写入则return
	select {
	case p.queue <- struct{}{}:
		return nil
	default:
	}

	timer := p.timers.Get().(*time.Timer)
	timer.Reset(p.cfg.PoolTimeout)

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		p.timers.Put(timer)
		return ctx.Err()
	case p.queue <- struct{}{}:
		if !timer.Stop() {
			<-timer.C
		}
		p.timers.Put(timer)
		return nil
	// 等待超时
	case <-timer.C:
		p.timers.Put(timer)
		atomic.AddUint32(&p.stats.Timeouts, 1)
		return ErrPoolTimeout
	}
}

func (p *Pool) popIdle() (*Client, error) {
	if p.closed() {
		return nil, ErrClosed
	}
	n := len(p.idleConns)
	if n == 0 {
		return nil, nil
	}

	idx := n - 1
	cn := p.idleConns[idx]
	p.idleConns = p.idleConns[:idx]

	p.idleConnsLen--
	p.checkMinIdleConns()

	return cn, nil
}

func (p *Pool) isHealthyConn(client *Client) bool {
	now := time.Now()

	if p.cfg.ConnMaxLifetime > 0 && now.Sub(client.createdAt) >= p.cfg.ConnMaxLifetime {
		return false
	}
	if p.cfg.ConnMaxIdleTime > 0 && now.Sub(client.UsedAt()) >= p.cfg.ConnMaxIdleTime {
		return false
	}

	// 这里还没有仔细思考 针对email的检查
	//if connCheck(cn.netConn) != nil {
	//	return false
	//}

	client.SetUsedAt(now)

	return true
}

func (p *Pool) Put() error {
	return nil
}

func (p *Pool) Stats() *Stats {
	return &Stats{
		Hits:     atomic.LoadUint32(&p.stats.Hits),
		Misses:   atomic.LoadUint32(&p.stats.Misses),
		Timeouts: atomic.LoadUint32(&p.stats.Timeouts),

		TotalConns: uint32(p.Len()),
		IdleConns:  uint32(p.IdleLen()),
		StaleConns: atomic.LoadUint32(&p.stats.StaleConns),
	}
}

func (p *Pool) Len() int {
	p.connsMu.Lock()
	n := len(p.conns)
	p.connsMu.Unlock()
	return n
}

func (p *Pool) IdleLen() int {
	p.connsMu.Lock()
	n := p.idleConnsLen
	p.connsMu.Unlock()
	return n
}

func (p *Pool) CloseConn(client *Client) error {
	return client.Close()
}
