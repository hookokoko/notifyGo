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

type Client struct {
	*Pool
	clientCfg *ClientConfig
}

type ClientConfig struct {
	addr string
	auth smtp.Auth
	*Options
}

func NewClient(cfg *ClientConfig) *Client {
	p := NewPool(cfg.addr, cfg.auth, cfg.Options)
	return &Client{Pool: p}
}

func (cp *Client) SendMail(ctx context.Context, e *Email) error {
	cli, err := cp.Pool.Get(ctx)
	defer func() {
		_ = cp.Pool.Put(cli)
	}()
	if err != nil {
		return err
	}
	return cli.sendMail(ctx, e)
}

func (cp *Client) Ping(ctx context.Context) error {
	// 考虑优先级队列，防止老的等待连接被饿死
	cli, err := cp.Pool.Get(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = cp.Pool.Put(cli)
	}()
	// TODO 统计下执行时间
	return cli.Noop()
}

type SMTPClient struct {
	*smtp.Client
	failCount int
	createdAt time.Time
	usedAt    int64
	pooled    bool
}

func (sc *SMTPClient) UsedAt() time.Time {
	unix := atomic.LoadInt64(&sc.usedAt)
	return time.Unix(unix, 0)
}

func (sc *SMTPClient) SetUsedAt(tm time.Time) {
	atomic.StoreInt64(&sc.usedAt, tm.Unix())
}

func (sc *SMTPClient) startTLS(t *tls.Config) (bool, error) {
	if ok, _ := sc.Extension("STARTTLS"); !ok {
		return false, nil
	}

	if err := sc.StartTLS(t); err != nil {
		return false, err
	}

	return true, nil
}

func (sc *SMTPClient) addAuth(auth smtp.Auth) (bool, error) {
	if ok, _ := sc.Extension("AUTH"); !ok {
		return false, nil
	}

	if err := sc.Auth(auth); err != nil {
		return false, err
	}

	return true, nil
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
	conns     []*SMTPClient // 准确说这是一个smtp的客户端
	idleConns []*SMTPClient

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
		conns:     make([]*SMTPClient, 0, opt.PoolSize),
		idleConns: make([]*SMTPClient, 0, opt.PoolSize),
	}

	if host, _, e := net.SplitHostPort(addr); e != nil {
		return nil
	} else {
		p.tlsConfig = &tls.Config{ServerName: host}
	}

	p.connsMu.Lock()
	p.checkMinIdleConns() // 保证池子里的连接数量能够达到MinIdleConns的要求
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
func (p *Pool) build() (*SMTPClient, error) {
	// 要不要加连接超时
	cl, err := smtp.Dial(p.addr)
	if err != nil {
		return nil, err
	}

	// Is there a custom hostname for doing a HELLO with the SMTP server?
	if p.helloHostname != "" {
		cl.Hello(p.helloHostname)
	}

	c := &SMTPClient{Client: cl, createdAt: time.Now()}
	c.SetUsedAt(time.Now())

	if _, err := c.startTLS(p.tlsConfig); err != nil {
		c.Close()
		return nil, err
	}

	if p.auth != nil {
		if _, err := c.addAuth(p.auth); err != nil {
			c.Close()
			return nil, err
		}
	}

	return c, nil
}

func (p *Pool) newConn(ctx context.Context) (*SMTPClient, error) {
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
	// 要不是put执行不到这个if分支
	if p.poolSize >= p.cfg.PoolSize {
		c.pooled = false
	} else {
		p.poolSize++
	}

	return c, nil
}

func (p *Pool) closed() bool {
	return atomic.LoadUint32(&p._closed) == 1
}

// Get 需要考虑优先级队列，防止老的等待连接被饿死
func (p *Pool) Get(ctx context.Context) (*SMTPClient, error) {
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
		// 没有取到连接，准确说是没有取到空闲连接，终止循环，在循环外的逻辑中新建一个连接
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
	// idle中没有空闲连接了，misses 意味着要再创建新的连接
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
		//log.Println("不用等待pool，直接有空位")
		return nil
	default:
	}

	timer := p.timers.Get().(*time.Timer)
	timer.Reset(p.cfg.PoolTimeout)
	//log.Println("等待pool空闲位置...")
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		p.timers.Put(timer)
		//log.Println("等待pool空闲位置...不好意思ctx超时了")
		return ctx.Err()
	case p.queue <- struct{}{}:
		if !timer.Stop() {
			<-timer.C
		}
		p.timers.Put(timer)
		//log.Println("等待pool空闲位置...nice等到了")
		return nil
	// 等待超时
	case <-timer.C:
		p.timers.Put(timer)
		atomic.AddUint32(&p.stats.Timeouts, 1)
		//log.Println("等待pool空闲位置...bad等待pool超时了")
		return ErrPoolTimeout
	}
}

func (p *Pool) popIdle() (*SMTPClient, error) {
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

func (p *Pool) isHealthyConn(client *SMTPClient) bool {
	now := time.Now()

	if p.cfg.ConnMaxLifetime > 0 && now.Sub(client.createdAt) >= p.cfg.ConnMaxLifetime {
		return false
	}
	if p.cfg.ConnMaxIdleTime > 0 && now.Sub(client.UsedAt()) >= p.cfg.ConnMaxIdleTime {
		return false
	}

	if err := client.Noop(); err != nil {
		return false
	}

	client.SetUsedAt(now)

	return true
}

func (p *Pool) Put(c *SMTPClient) error {
	p.connsMu.Lock()
	// 放回的时候，pool_size是不会发生变化的。除非要remove掉这个连接。
	// 如果MaxIdleConns设置为0，标识无限空闲连接数
	// p.idleConnsLen >= p.cfg.MaxIdleConns的情况是不是同时put多次(加锁了应该不是)。min_idle_size > max_idle_size的情况？
	// p.cfg.MaxIdleConns == 0 且 p.cfg.MinIdleConns == 0 的情况，也会将 连接放到idleConns里，这种情况是不是违背了MinIdleConns的概念？但是效果是不影响的，即连接不会重新创建
	// p.cfg.MaxIdleConns == 1 且 p.cfg.MinIdleConns == 0 的情况，同上
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
	// 不管是放到idleconn中，还是remove掉，都会腾出一个p.queue的位置
	<-p.queue
	return nil
}

func (p *Pool) removeConn(client *SMTPClient) {
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

	// 测试发现一般不会走到这里，走到这个函数的条件是设置了max_idle_size > 0 并且 len(idleconn) >= max_idle_size
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

func (p *Pool) CloseConn(client *SMTPClient) error {
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

func (c *SMTPClient) sendMail(ctx context.Context, e *Email) error {
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
