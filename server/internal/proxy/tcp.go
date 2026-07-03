package proxy

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// TCPForwarder TCP转发器
type TCPForwarder struct {
	listener net.Listener
	tunnel   *Tunnel
	client   *Client
	connMap  sync.Map // map[string]net.Conn
}

// NewTCPForwarder 创建TCP转发器
func NewTCPForwarder(tunnel *Tunnel, client *Client) *TCPForwarder {
	return &TCPForwarder{
		tunnel: tunnel,
		client: client,
	}
}

// Start 启动TCP监听
func (f *TCPForwarder) Start() error {
	addr := f.tunnel.LocalHost
	if addr == "" {
		addr = "0.0.0.0"
	}
	addr = addr + ":" + itoa(f.tunnel.LocalPort)

	var err error
	f.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("TCP转发器启动: %s -> %d", addr, f.tunnel.RemotePort)

	go f.acceptLoop()
	return nil
}

// Stop 停止TCP转发器
func (f *TCPForwarder) Stop() {
	if f.listener != nil {
		f.listener.Close()
	}

	f.connMap.Range(func(key, value interface{}) bool {
		conn := value.(net.Conn)
		conn.Close()
		return true
	})
}

// acceptLoop 接受连接循环
func (f *TCPForwarder) acceptLoop() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}

		go f.handleConn(conn)
	}
}

// handleConn 处理单个连接
func (f *TCPForwarder) handleConn(conn net.Conn) {
	connID := generateID()
	f.connMap.Store(connID, conn)
	defer func() {
		f.connMap.Delete(connID)
		conn.Close()
	}()

	// 发送DataOpen消息给客户端
	openMsg := &Message{
		Type:  MsgTypeDataOpen,
		Flag:  MsgFlagNone,
		MsgID: 1,
		Payload: []byte(`{"tunnel_id":"` + f.tunnel.ID + `","connection_id":"` + connID + `","protocol":"tcp","source_addr":"` + conn.RemoteAddr().String() + `"}`),
	}
	f.client.Send <- openMsg.Encode()

	// 读取本地数据并转发
	buf := make([]byte, 32*1024)
	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		// 发送数据给客户端
		dataMsg := &Message{
			Type:  MsgTypeDataTransfer,
			Flag:  MsgFlagNone,
			MsgID: 1,
			Payload: buf[:n],
		}
		f.client.Send <- dataMsg.Encode()
	}

	// 发送关闭消息
	closeMsg := &Message{
		Type:    MsgTypeDataClose,
		Flag:    MsgFlagNone,
		MsgID:   1,
		Payload: []byte(`{"connection_id":"` + connID + `"}`),
	}
	f.client.Send <- closeMsg.Encode()
}

// itoa 整数转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// generateID 生成唯一ID
func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}

// Copy双向拷贝
func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
