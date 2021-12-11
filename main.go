package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-bai/ssrp/backend"
	"github.com/go-bai/ssrp/config"
)

var backendPools []backend.BackendPool
var urlStatus backend.IsAlive

type myHandler struct {
	backendPool *backend.BackendPool
}

func (p *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer, t := p.backendPool.GetNextPeer(&urlStatus)
	if peer != nil {
		fmt.Printf("|| 本地端口: %-5s || 远程地址: %-21s || Host: %-15s || 主负载节点: %s %dms\n", p.backendPool.Port, r.RemoteAddr, peer.Host, peer.Url, t)
		r.Host = peer.Host
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	} else {
		// 如果有备用节点
		if p.backendPool.BackUp != nil {
			fmt.Printf("|| 本地端口: %-5s || 远程地址: %-21s || Host: %-15s || 备用节点: %s\n", p.backendPool.Port, r.RemoteAddr, p.backendPool.BackUp.Host, p.backendPool.BackUp.Url)
			// r.Host = p.backendPool.BackUp.Host
			p.backendPool.BackUp.ReverseProxy.ServeHTTP(w, r)
			return
		}
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func main() {
	config := config.Config{}
	err := config.Read()
	if err != nil {
		panic(err)
	}
	urlStatus.TimeOut = config.TimeOut
	for _, bd := range config.Backends {
		bdpool := backend.BackendPool{
			Current: 0,
			Port:    bd.Port,
		}
		host := bd.Host
		// 备用节点
		if bd.BackUp != "" {
			uu, err := url.Parse(bd.BackUp)
			if err != nil {
				log.Panic(err)
			}
			reverseproxy := httputil.NewSingleHostReverseProxy(uu)
			reverseproxy.Transport = &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: time.Duration(urlStatus.TimeOut) * time.Millisecond,
				}).DialContext,
			}
			reverseproxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) {
				fmt.Printf("[%s] %s\n", uu.Host, e.Error())
			}
			bdpool.BackUp = &backend.Backend{
				Url:          uu,
				Host:         host,
				ReverseProxy: reverseproxy,
			}
		}
		// 主节点
		for _, u := range bd.Urls {
			uu, err := url.Parse(u)
			if err != nil {
				log.Panic(err)
			}
			urlStatus.SetStatus(uu.Host, 10) // 将主节点存入 sync.map
			reverseproxy := httputil.NewSingleHostReverseProxy(uu)
			reverseproxy.Transport = &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: time.Duration(urlStatus.TimeOut) * time.Millisecond,
				}).DialContext,
			}
			reverseproxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) {
				fmt.Printf("|| 本地端口: %-5s || 远程地址: %-21s || Host: %-15s || 备用节点: %s\n", bd.Port, r.RemoteAddr, bd.Host, bd.BackUp)
				bdpool.BackUp.ReverseProxy.ServeHTTP(rw, r)
			}
			bdpool.Backends = append(bdpool.Backends, &backend.Backend{
				Url:          uu,
				Host:         host,
				ReverseProxy: reverseproxy,
			})
		}
		backendPools = append(backendPools, bdpool)
	}
	for i, bd := range backendPools {
		handler := &myHandler{&backendPools[i]}
		server := http.Server{
			Addr:    fmt.Sprintf(":%s", bd.Port),
			Handler: handler,
		}
		go listenAndServe(&server)
	}

	fmt.Println("ssrp 启动成功")

	t := time.NewTicker(time.Second * time.Duration(config.HealthCheckInterval))
	for range t.C {
		urlStatus.HealthCheck()
	}
}

func listenAndServe(s *http.Server) {
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
