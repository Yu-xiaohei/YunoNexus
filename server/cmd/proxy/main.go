package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yunonexus/server/internal/config"
	"github.com/yunonexus/server/internal/database"
	"github.com/yunonexus/server/internal/proxy"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 创建代理服务器
	proxyServer, err := proxy.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("代理服务器创建失败: %v", err)
	}

	// 优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 启动代理服务器
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Proxy.Port)
		log.Printf("YunoNexus 代理服务已启动: %s", addr)
		
		if err := proxyServer.Listen(addr); err != nil {
			log.Fatalf("代理服务器启动失败: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("正在关闭代理服务器...")

	proxyServer.Shutdown()
	log.Println("代理服务器已关闭")
}

// Server 代理服务器
type Server struct {
	cfg      *config.Config
	db       *database.DB
	upgrader websocket.Upgrader
	clients  sync.Map
}

// NewServer 创建新的代理服务器
func NewServer(cfg *config.Config, db *database.DB) (*Server, error) {
	return &Server{
		cfg: cfg,
		db:  db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源（生产环境应限制）
			},
		},
	}, nil
}

// Listen 监听WebSocket连接
func (s *Server) Listen(addr string) error {
	http.HandleFunc("/ws", s.handleWebSocket)
	return http.ListenAndServe(addr, nil)
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		server: s,
		send:   make(chan []byte, 256),
	}

	s.clients.Store(client, true)

	go client.writePump()
	go client.readPump()
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	s.clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		client.conn.Close()
		return true
	})
}

// Client 客户端连接
type Client struct {
	conn   *websocket.Conn
	server *Server
	send   chan []byte
	userID string
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.server.clients.Delete(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// 处理消息...
		log.Printf("收到消息: %s", message)
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
