package proxy

import (
	"log"
	"net"
	"sync"
	"time"
)

// UDPForwarder UDP转发器
type UDPForwarder struct {
	conn     *net.UDPConn
	tunnel   *Tunnel
	client   *Client
	connMap  sync.Map
}

// NewUDPForwarder 创建UDP转发器
func NewUDPForwarder(tunnel *Tunnel, client *Client) *UDPForwarder {
	return &UDPForwarder{
		tunnel: tunnel,
		client: client,
	}
}

// Start 启动UDP监听
func (f *UDPForwarder) Start() error {
	addr := f.tunnel.LocalHost
	if addr == "" {
		addr = "0.0.0.0"
	}
	addr = addr + ":" + itoa(f.tunnel.LocalPort)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	f.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	log.Printf("UDP转发器启动: %s -> %d", addr, f.tunnel.RemotePort)

	go f.readLoop()
	return nil
}

// Stop 停止UDP转发器
func (f *UDPForwarder) Stop() {
	if f.conn != nil {
		f.conn.Close()
	}
}

// readLoop 读取UDP数据循环
func (f *UDPForwarder) readLoop() {
	buf := make([]byte, 65535)
	for {
		f.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, remoteAddr, err := f.conn.ReadFromUDP(buf)
		if err != nil {
			return
		}

		addrKey := remoteAddr.String()

		// 获取或创建本地连接
		var localConn *net.UDPConn
		if v, ok := f.connMap.Load(addrKey); ok {
			localConn = v.(*net.UDPConn)
		} else {
			// 连接到本地服务
			localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+itoa(f.tunnel.LocalPort))
			if err != nil {
				log.Printf("解析UDP地址失败: %v", err)
				continue
			}
			localConn, err = net.DialUDP("udp", nil, localAddr)
			if err != nil {
				log.Printf("UDP连接本地服务失败: %v", err)
				continue
			}
			f.connMap.Store(addrKey, localConn)

			go f.handleLocalResponse(addrKey, remoteAddr, localConn)
		}

		// 转发数据到本地
		if _, err := localConn.Write(buf[:n]); err != nil {
			log.Printf("UDP转发数据失败: %v", err)
		}
	}
}

// handleLocalResponse 处理本地服务响应
func (f *UDPForwarder) handleLocalResponse(addrKey string, remoteAddr *net.UDPAddr, localConn *net.UDPConn) {
	defer func() {
		f.connMap.Delete(addrKey)
		localConn.Close()
	}()

	buf := make([]byte, 65535)
	for {
		localConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := localConn.Read(buf)
		if err != nil {
			return
		}

		// 转发响应给客户端
		if _, err := f.conn.WriteToUDP(buf[:n], remoteAddr); err != nil {
			log.Printf("UDP回传数据失败: %v", err)
		}
	}
}
