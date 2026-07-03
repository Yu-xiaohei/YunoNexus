package proxy

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// HTTPForwarder HTTP转发器
type HTTPForwarder struct {
	server *http.Server
	tunnel *Tunnel
	client *Client
}

// NewHTTPForwarder 创建HTTP转发器
func NewHTTPForwarder(tunnel *Tunnel, client *Client) *HTTPForwarder {
	return &HTTPForwarder{
		tunnel: tunnel,
		client: client,
	}
}

// Start 启动HTTP监听
func (f *HTTPForwarder) Start() error {
	addr := f.tunnel.LocalHost
	if addr == "" {
		addr = "0.0.0.0"
	}
	addr = ":" + itoa(f.tunnel.LocalPort)

	// 创建反向代理
	target := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:" + itoa(f.tunnel.LocalPort),
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Transport
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	f.server = &http.Server{
		Addr:    addr,
		Handler: proxy,
	}

	log.Printf("HTTP转发器启动: %s", addr)

	go func() {
		if err := f.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP服务错误: %v", err)
		}
	}()

	return nil
}

// Stop 停止HTTP转发器
func (f *HTTPForwarder) Stop() {
	if f.server != nil {
		f.server.Close()
	}
}
