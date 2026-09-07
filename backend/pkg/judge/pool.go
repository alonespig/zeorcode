package judge

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// node 池中一台判题机 + 健康标记
type node struct {
	client  *Client
	healthy atomic.Bool
}

// Pool 多台 go-judge 的负载均衡池：按提交轮询挑一台，跳过不健康的实例。
// 一次提交的 编译→运行→删除 必须落在同一台（编译产物 FileID 存在该实例上），
// 所以负载均衡粒度是"每提交一台"，而非每个 HTTP 请求。
type Pool struct {
	nodes   []*node
	counter uint64
	stop    chan struct{}
}

// NewPool 用一组沙箱地址构造池（自动去空白/跳过空项）。
func NewPool(urls []string) *Pool {
	p := &Pool{stop: make(chan struct{})}
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		nd := &node{client: NewClient(u)}
		nd.healthy.Store(true) // 初始乐观，健康检查会很快校正
		p.nodes = append(p.nodes, nd)
	}
	return p
}

// Size 池中实例数
func (p *Pool) Size() int { return len(p.nodes) }

// URLs 池中所有地址（日志用）
func (p *Pool) URLs() []string {
	out := make([]string, 0, len(p.nodes))
	for _, nd := range p.nodes {
		out = append(out, nd.client.baseURL)
	}
	return out
}

// Next 轮询取下一台健康实例；若全都不健康，兜底返回轮询到的那台（fail-open，让它去试）。
func (p *Pool) Next() *Client {
	n := len(p.nodes)
	if n == 0 {
		return nil
	}
	start := atomic.AddUint64(&p.counter, 1)
	for i := 0; i < n; i++ {
		nd := p.nodes[int((start+uint64(i))%uint64(n))]
		if nd.healthy.Load() {
			return nd.client
		}
	}
	return p.nodes[int(start%uint64(n))].client
}

// MarkDown 把某台标记为不健康（调用方检测到网络不可达时用）；下次健康检查会重新探活恢复。
func (p *Pool) MarkDown(c *Client) {
	for _, nd := range p.nodes {
		if nd.client == c {
			nd.healthy.Store(false)
			return
		}
	}
}

// StartHealthCheck 启动后台定时探活协程，interval<=0 用默认 15s。
func (p *Pool) StartHealthCheck(interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	go func() {
		p.checkAll()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-p.stop:
				return
			case <-t.C:
				p.checkAll()
			}
		}
	}()
}

// Stop 停止健康检查协程（进程退出时调用一次）
func (p *Pool) Stop() { close(p.stop) }

func (p *Pool) checkAll() {
	for _, nd := range p.nodes {
		nd.healthy.Store(nd.client.ping())
	}
}

// ping 探活：能拿到任何 HTTP 响应就算活（哪怕 404，也说明进程/端口在）。
func (c *Client) ping() bool {
	hc := &http.Client{Timeout: 3 * time.Second}
	resp, err := hc.Get(c.baseURL + "/version")
	if err != nil {
		resp, err = hc.Get(c.baseURL)
		if err != nil {
			return false
		}
	}
	_ = resp.Body.Close()
	return true
}

// IsUnavailable 判断错误是否为"判题机不可达"（网络/连接层，net/http 会包成 *url.Error）；
// 用户代码编译失败之类的业务错误（fmt.Errorf）不会命中，可据此只对前者做失败转移。
func IsUnavailable(err error) bool {
	var ue *url.Error
	return errors.As(err, &ue)
}
