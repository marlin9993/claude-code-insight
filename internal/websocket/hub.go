package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应限制）
	},
}

// Message WebSocket消息
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client WebSocket客户端
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

// Hub WebSocket连接中心
type Hub struct {
	clients    []*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewHub 创建新的 Hub
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run 启动 Hub 事件循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients = append(h.clients, client)
			h.mu.Unlock()
			log.Printf("WebSocket 客户端连接，当前连接数: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			for i, c := range h.clients {
				if c == client {
					h.clients = append(h.clients[:i], h.clients[i+1:]...)
					break
				}
			}
			h.mu.Unlock()
			close(client.Send)
			log.Printf("WebSocket 客户端断开，当前连接数: %d", len(h.clients))

		case message := <-h.broadcast:
			// 收集需要断开的客户端（在持有锁时）
			h.mu.RLock()
			var disconnectedClients []*Client
			for _, client := range h.clients {
				select {
				case client.Send <- message:
					// 消息发送成功
				default:
					// 客户端 Send 通道已满或已关闭，标记为需要断开
					disconnectedClients = append(disconnectedClients, client)
				}
			}
			h.mu.RUnlock()

			// 释放锁后再断开客户端连接，避免在持有锁时关闭通道
			for _, client := range disconnectedClients {
				client.Disconnect()
			}
		}
	}
}

// Broadcast 广播消息给所有客户端
func (h *Hub) Broadcast(msgType string, data interface{}) error {
	message := Message{
		Type: msgType,
		Data: data,
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- jsonData
	return nil
}

// HandleWebSocket 处理 WebSocket 连接
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}

	client := &Client{
		Hub:  h,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	h.register <- client

	// 读取协程
	go client.readPump()
	// 写入协程
	go client.writePump()
}

// readPump 读取 WebSocket 消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			break
		}
	}
}

// writePump 写入 WebSocket 消息
func (c *Client) writePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Disconnect 安全地断开客户端连接
func (c *Client) Disconnect() {
	// 通过 unregister channel 让 Hub 优雅地处理断开
	c.Hub.unregister <- c
}

// GetClientCount 获取当前连接的客户端数量
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
