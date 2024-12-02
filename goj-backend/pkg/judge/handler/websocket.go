package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// 全局WebSocket管理器实例
var wsManager *WebSocketManager

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	connections sync.Map     // userID -> []websocket.Conn
	maxConns    int          // 最大连接数
	connCount   atomic.Int32 // 当前连接数
}

// InitWebSocketManager 初始化WebSocket管理器
func InitWebSocketManager() {
	wsManager = NewWebSocketManager()
}

// GetWebSocketManager 获取WebSocket管理器实例
func GetWebSocketManager() *WebSocketManager {
	return wsManager
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{}
}

// AddConnection 添加连接
func (m *WebSocketManager) AddConnection(userID uint, conn *websocket.Conn) error {
	// 检查连接数限制
	if m.connCount.Load() >= int32(m.maxConns) {
		return errors.New("达到最大连接数限制")
	}

	m.connections.Store(userID, conn)
	m.connCount.Add(1)
	return nil
}

// RemoveConnection 移除连接
func (m *WebSocketManager) RemoveConnection(userID uint) {
	if _, exists := m.connections.LoadAndDelete(userID); exists {
		m.connCount.Add(-1)
	}
}

// SendToUser 发送消息给指定用户
func (m *WebSocketManager) SendToUser(userID uint, msg WebSocketMessage) error {
	log.Printf("[WebSocket] Sending message to user %d: %+v", userID, msg)
	if conn, ok := m.connections.Load(userID); ok {
		if wsConn, ok := conn.(*websocket.Conn); ok {
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[WebSocket] Failed to marshal message: %v", err)
				return err
			}
			if err := wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[WebSocket] Failed to send message: %v", err)
				m.RemoveConnection(userID)
				return err
			}
			log.Printf("[WebSocket] Message sent successfully to user %d", userID)
			return nil
		}
		log.Printf("[WebSocket] Invalid connection type for user %d", userID)
		return fmt.Errorf("invalid connection type")
	}
	log.Printf("[WebSocket] No connection found for user %d", userID)
	return fmt.Errorf("no connection found")
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uint) {
	log.Printf("[WebSocket] 收到新的WebSocket连接请求")
	log.Printf("[WebSocket] 请求URL: %s", r.URL.String())
	log.Printf("[WebSocket] 请求头: %+v", r.Header)

	// 检查token
	token := r.URL.Query().Get("token")
	log.Printf("[WebSocket] 收到token: %s", token)

	if wsManager == nil {
		log.Printf("[WebSocket] 错误: WebSocket管理器未初始化")
		http.Error(w, "WebSocket服务未准备就绪", http.StatusServiceUnavailable)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			log.Printf("[WebSocket] 检查Origin: %s", origin)
			return true // 开发环境允许所有来源
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] 升级连接失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[WebSocket] 连接成功建立, 用户ID: %d", userID)

	// 将连接添加到管理器
	wsManager.AddConnection(userID, conn)
	defer wsManager.RemoveConnection(userID)

	// 发送连接成功消息
	welcomeMsg := WebSocketMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"message": "WebSocket连接成功",
			"userId":  userID,
		},
	}
	if err := wsManager.SendToUser(userID, welcomeMsg); err != nil {
		log.Printf("[WebSocket] 发送欢迎消息失败: %v", err)
	}

	// 保持连接活跃
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WebSocket] 读取消息失败: %v", err)
			break
		}
		log.Printf("[WebSocket] 收到消息: type=%d, payload=%s", messageType, string(p))

		// 处理心跳消息
		if string(p) == "ping" {
			if err := conn.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				log.Printf("[WebSocket] 发送pong失败: %v", err)
				break
			}
		}
	}
}

// BroadcastToUser 向指定用户广播消息
func (m *WebSocketManager) BroadcastToUser(userID uint, msg interface{}) error {
	log.Printf("[WebSocket] Broadcasting message to user %d: %+v", userID, msg)
	return m.SendToUser(userID, WebSocketMessage{
		Type: "judge_result",
		Data: msg,
	})
}
