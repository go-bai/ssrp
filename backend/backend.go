package backend

import (
	"fmt"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type IsAlive struct {
	M       sync.Map // map[host]rtt
	TimeOut int64
}

type Backend struct {
	Url          *url.URL
	Host         string
	ReverseProxy *httputil.ReverseProxy
}

type BackendPool struct {
	Backends []*Backend // 所有节点
	Current  uint64     // 当前轮到的节点索引
	Port     string     // 本地端口
	BackUp   *Backend   // 备用
}

func (p *BackendPool) NextIndex() int {
	return int(atomic.AddUint64(&p.Current, uint64(1)) % uint64(len(p.Backends)))
}

func (i *IsAlive) GetStatus(url string) (int64, bool) {
	v, ok := i.M.Load(url)
	if ok && v.(int64) == 0 {
		return v.(int64), false
	}
	return v.(int64), true
}

func (i *IsAlive) SetStatus(url string, status int64) {
	i.M.Store(url, status)
}

func (p *BackendPool) GetNextPeer(alive *IsAlive) (*Backend, int64) {
	next := p.NextIndex()
	l := len(p.Backends) + next
	for i := next; i < l; i++ {
		idx := i % len(p.Backends)
		if t, ok := alive.GetStatus(p.Backends[idx].Url.Host); ok {
			if i != next {
				atomic.StoreUint64(&p.Current, uint64(idx))
			}
			return p.Backends[idx], t
		}
	}
	return nil, 0
}

func (i *IsAlive) HealthCheck() {
	cnt := 0
	sum := 0
	i.M.Range(func(k, v interface{}) bool {
		sum++
		t, err := isBackAclive(k.(string), i.TimeOut)
		if err != nil {
			fmt.Println("定时检测:", k, err)
			cnt++
			i.SetStatus(k.(string), 0)
		} else {
			i.SetStatus(k.(string), t)
		}
		return true
	})
	fmt.Printf("失联节点 %d/%d\n", cnt, sum)
}

func isBackAclive(host string, to int64) (t int64, err error) {
	start := time.Now().UnixNano() / 1e6
	conn, err := net.DialTimeout("tcp", host, time.Duration(to*1e6))
	end := time.Now().UnixNano() / 1e6
	if err != nil {
		return
	}
	_ = conn.Close()
	return end - start, nil
}
