package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
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

// ---------- TODO 完善NewClient

type ClientWithPool struct {
	*Pool
	clientCfg *ClientOptions
}

type ClientOptions struct {
	addr string
	auth smtp.Auth
	*Options
}

func NewClientWithPool(opt *ClientOptions) *ClientWithPool {
	p := NewPool(opt.addr, opt.auth, opt.Options)
	return &ClientWithPool{Pool: p}
}

func (cp *ClientWithPool) SendMail(ctx context.Context, e *Email) error {
	cli, err := cp.Pool.Get(ctx)
	defer func() {
		_ = cp.Pool.Put(cli)
	}()
	if err != nil {
		return err
	}
	return cli.sendMail(ctx, e)
}

// ----------------------

type Client struct {
	*smtp.Client
	failCount int
	createdAt time.Time
	usedAt    int64
	pooled    bool
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
	// 初始创建pool时，最多创建idleConnsLen个连接
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
	c, err := p.build()
	if err != nil {
		return err
	}

	p.connsMu.Lock()
	defer p.connsMu.Unlock()

	// It is not allowed to add new connections to the closed connection pool.
	if p.closed() {
		_ = c.Close()
		return ErrClosed
	}

	p.conns = append(p.conns, c)
	p.idleConns = append(p.idleConns, c)
	return nil
}

// 单纯的创建一个连接
func (p *Pool) build() (*Client, error) {
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

func (p *Pool) newConn(ctx context.Context) (*Client, error) {
	c, err := p.build()
	if err != nil {
		return nil, err
	}

	p.connsMu.Lock()
	defer p.connsMu.Unlock()

	// It is not allowed to add new connections to the closed connection pool.
	if p.closed() {
		_ = c.Close()
		return nil, ErrClosed
	}

	p.conns = append(p.conns, c)

	// If pool is full remove the cn on next Put.
	if p.poolSize >= p.cfg.PoolSize {
		c.pooled = false
	} else {
		p.poolSize++
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

	// 这里就是等待获取连接的条件是否满足，满足就往下走，不满足按照一定的策略等待
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

	newClient, err := p.newConn(ctx)
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
		log.Println("不用等待pool，直接有空位")
		return nil
	default:
	}

	timer := p.timers.Get().(*time.Timer)
	timer.Reset(p.cfg.PoolTimeout)
	log.Println("等待pool空闲位置...")
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		p.timers.Put(timer)
		log.Println("等待pool空闲位置...不好意思ctx超时了")
		return ctx.Err()
	case p.queue <- struct{}{}:
		if !timer.Stop() {
			<-timer.C
		}
		p.timers.Put(timer)
		log.Println("等待pool空闲位置...nice等到了")
		return nil
	// 等待超时
	case <-timer.C:
		p.timers.Put(timer)
		atomic.AddUint32(&p.stats.Timeouts, 1)
		log.Println("等待pool空闲位置...bad等待pool超时了")
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
	// 暂时还不确定这个函数是否符合要求
	if err := client.Noop(); err != nil {
		return false
	}

	client.SetUsedAt(now)

	return true
}

func (p *Pool) Put(c *Client) error {
	p.connsMu.Lock()
	// 如果MaxIdleConns设置为0，标识无限空闲连接数？
	if p.cfg.MaxIdleConns == 0 || p.idleConnsLen < p.cfg.MaxIdleConns {
		err := c.Reset()
		// 重置失败，直接关闭
		if err != nil {
			log.Println("put时，重置连接失败", err)
			_ = c.Close()
			p.connsMu.Unlock()
			return err
		}
		p.idleConns = append(p.idleConns, c)
		p.idleConnsLen++
	} else {
		p.removeConn(c)
	}

	p.connsMu.Unlock()

	// freeTurn()
	<-p.queue
	log.Println("put放进去了一个位置")
	return nil
}

func (p *Pool) removeConn(client *Client) {
	for i, c := range p.conns {
		if c == client {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			p.poolSize--
			p.checkMinIdleConns()
			break
		}
	}
	// 关闭移除的连接
	_ = client.Close()

	atomic.AddUint32(&p.stats.StaleConns, 1)
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

func (p *Pool) SendMail(ctx context.Context, e *Email) error {
	c, err := p.Get(ctx)
	if c == nil || err != nil {
		return fmt.Errorf("Err: 获取邮件客户端连接失败: %v", err)
	}
	defer func() {
		_ = p.Put(c)
	}()

	return c.sendMail(ctx, e)
}

func (c *Client) sendMail(ctx context.Context, e *Email) error {
	recipients, err := addressLists(e.To, e.Cc, e.Bcc)
	if err != nil {
		return err
	}

	msg, err := e.Bytes()
	if err != nil {
		return err
	}

	from, err := emailOnly(e.From)
	if err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}

	for _, recip := range recipients {
		if err = c.Rcpt(recip); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}

	err = w.Close()

	return err
}

func emailOnly(full string) (string, error) {
	addr, err := mail.ParseAddress(full)
	if err != nil {
		return "", err
	}
	return addr.Address, nil
}

func addressLists(lists ...[]string) ([]string, error) {
	length := 0
	for _, lst := range lists {
		length += len(lst)
	}
	combined := make([]string, 0, length)

	for _, lst := range lists {
		for _, full := range lst {
			addr, err := emailOnly(full)
			if err != nil {
				return nil, err
			}
			combined = append(combined, addr)
		}
	}

	return combined, nil
}

func (p *Pool) Close() error {
	if !atomic.CompareAndSwapUint32(&p._closed, 0, 1) {
		return ErrClosed
	}

	var firstErr error
	p.connsMu.Lock()
	for _, cn := range p.conns {
		if err := cn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.conns = nil
	p.poolSize = 0
	p.idleConns = nil
	p.idleConnsLen = 0
	p.connsMu.Unlock()

	return firstErr
}
