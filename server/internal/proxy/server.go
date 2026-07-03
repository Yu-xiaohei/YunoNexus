package proxy

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yunonexus/server/internal/auth"
	"github.com/yunonexus/server/internal/config"
	"github.com/yunonexus/server/internal/database"
)

// Server 代理服务器
type Server struct {
	cfg      *config.Config
	db       *database.DB
	upgrader websocket.Upgrader
	clients  sync.Map // map[string]*Client
}

// Client 客户端连接
type Client struct {
	ID         string
	UserID     string
	DeviceID   string
	Conn       *websocket.Conn
	Send       chan []byte
	Encryptor  *Encryptor
	Compressor *Compressor
	LastPing   time.Time
	Tunnels    map[string]*Tunnel
	Forwarders map[string]interface{} // 转发器实例
	mu         sync.Mutex
}

// Tunnel 隧道
type Tunnel struct {
	ID       string
	Protocol string
	LocalHost string
	LocalPort int
	RemotePort int
	Status   string
}

// NewServer 创建代理服务器
func NewServer(cfg *config.Config, db *database.DB) *Server {
	return &Server{
		cfg: cfg,
		db:  db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 验证来源域名
				origin := r.Header.Get("Origin")
				// 允许的来源列表（可以从配置中读取）
				allowedOrigins := []string{
					"https://dev.yxhmc.cn",
					"https://yxhmc.cn",
					"http://localhost:3000",
					"http://localhost:5173",
				}
				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}
				// 无来源的请求（如WebSocket客户端）也允许
				return origin == ""
			},
		},
	}
}

// Listen 监听WebSocket连接
func (s *Server) Listen(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("代理服务已启动: %s", addr)
	return server.ListenAndServe()
}

// handleWebSocket 处理WebSocket连接
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &Client{
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Encryptor:  nil, // 认证后初始化
		Compressor: NewCompressor(6),
		LastPing:   time.Now(),
		Tunnels:    make(map[string]*Tunnel),
	}

	go client.writePump()
	go client.readPump(s)
}

// GetOnlineClients 获取在线客户端数
func (s *Server) GetOnlineClients() int {
	count := 0
	s.clients.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// readPump 读取消息
func (c *Client) readPump(server *Server) {
	defer func() {
		server.clients.Delete(c.ID)
		c.Conn.Close()
		log.Printf("客户端断开: %s", c.ID)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		c.LastPing = time.Now()

		msg, err := DecodeMessage(message)
		if err != nil {
			log.Printf("解码消息失败: %v", err)
			continue
		}

		c.handleMessage(server, msg)
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				w.Close()
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理消息
func (c *Client) handleMessage(server *Server, msg *Message) {
	switch msg.Type {
	case MsgTypeAuth:
		c.handleAuth(server, msg)
	case MsgTypeHeartbeat:
		c.handleHeartbeat(msg)
	case MsgTypeDataOpen:
		c.handleDataOpen(msg)
	case MsgTypeDataTransfer:
		c.handleDataTransfer(msg)
	case MsgTypeDataClose:
		c.handleDataClose(msg)
	default:
		log.Printf("未知消息类型: %d", msg.Type)
	}
}

// handleAuth 处理认证
func (c *Client) handleAuth(server *Server, msg *Message) {
	// 解析认证载荷
	var authPayload struct {
		Token    string `json:"token"`
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(msg.Payload, &authPayload); err != nil {
		log.Printf("解析认证载荷失败: %v", err)
		c.sendAuthError(msg.MsgID, "认证数据格式错误")
		return
	}

	// 验证JWT token
	claims, err := auth.ValidateToken(authPayload.Token, server.cfg.JWT.Secret)
	if err != nil {
		log.Printf("JWT验证失败: %v", err)
		c.sendAuthError(msg.MsgID, "认证令牌无效或已过期")
		return
	}

	// 生成安全的随机加密密钥
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Printf("生成加密密钥失败: %v", err)
		c.sendAuthError(msg.MsgID, "系统错误")
		return
	}

	enc, err := NewEncryptor(key)
	if err != nil {
		log.Printf("创建加密器失败: %v", err)
		c.sendAuthError(msg.MsgID, "系统错误")
		return
	}
	c.Encryptor = enc
	c.UserID = claims.UserID
	c.DeviceID = authPayload.DeviceID
	// 使用随机字节生成不可预测的Client ID
	randomID := make([]byte, 16)
	rand.Read(randomID)
	c.ID = fmt.Sprintf("client-%x", randomID)

	server.clients.Store(c.ID, c)

	response := &Message{
		Type:    MsgTypeAuthResponse,
		Flag:    MsgFlagNone,
		MsgID:   msg.MsgID,
		Payload: []byte(`{"success":true,"message":"认证成功","heartbeat_interval":30}`),
	}

	c.Send <- response.Encode()
	log.Printf("客户端 %s (用户 %s) 已连接", c.ID, c.UserID)
}

// sendAuthError 发送认证错误
func (c *Client) sendAuthError(msgID uint32, message string) {
	response := &Message{
		Type:    MsgTypeAuthResponse,
		Flag:    MsgFlagNone,
		MsgID:   msgID,
		Payload: []byte(`{"success":false,"message":"` + message + `"}`),
	}
	c.Send <- response.Encode()
}

// handleHeartbeat 处理心跳
func (c *Client) handleHeartbeat(msg *Message) {
	response := &Message{
		Type:    MsgTypeHeartbeatAck,
		Flag:    MsgFlagNone,
		MsgID:   msg.MsgID,
		Payload: []byte{},
	}
	c.Send <- response.Encode()
}

// handleDataOpen 处理数据通道打开
func (c *Client) handleDataOpen(msg *Message) {
	log.Printf("收到数据通道打开请求")
	// TODO: 实现TCP/UDP转发
}

// handleDataTransfer 处理数据传输
func (c *Client) handleDataTransfer(msg *Message) {
	// TODO: 实现数据转发
}

// handleDataClose 处理数据通道关闭
func (c *Client) handleDataClose(msg *Message) {
	log.Printf("收到数据通道关闭请求")
}
