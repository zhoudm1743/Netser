package core

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	wsProtocol "github.com/zhoudm1743/Netser/dto/websocket"
)

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	port      int                        // WebSocket服务端口
	server    *http.Server               // HTTP服务器
	upgrader  websocket.Upgrader         // WebSocket升级器
	clients   map[string]*WSClient       // 客户端连接池 clientId -> WSClient
	sessions  map[string]map[string]bool // 会话订阅映射 sessionId -> {clientId: true}
	mutex     sync.RWMutex               // 读写锁
	ctx       context.Context            // 上下文
	cancel    context.CancelFunc         // 取消函数
	isRunning bool                       // 运行状态
}

// WSClient WebSocket客户端
type WSClient struct {
	ID            string            // 客户端ID
	Conn          *websocket.Conn   // WebSocket连接
	Send          chan []byte       // 发送通道
	Manager       *WebSocketManager // 管理器引用
	LastPing      time.Time         // 最后心跳时间
	Subscriptions map[string]bool   // 订阅的会话列表
	mutex         sync.RWMutex      // 读写锁
}

var GlobalWebSocketManager *WebSocketManager

// InitWebSocketManager 初始化WebSocket管理器
func InitWebSocketManager() error {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &WebSocketManager{
		port: 0, // 稍后分配
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有来源（开发环境）
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:   make(map[string]*WSClient),
		sessions:  make(map[string]map[string]bool),
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}

	// 尝试分配端口
	port, err := manager.allocatePort()
	if err != nil {
		cancel()
		return fmt.Errorf("无法分配WebSocket端口: %v", err)
	}
	manager.port = port

	GlobalWebSocketManager = manager
	return nil
}

// allocatePort 分配可用端口
func (wm *WebSocketManager) allocatePort() (int, error) {
	startPort := 1743
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if wm.isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("无法找到可用端口，尝试了端口 %d-%d", startPort, startPort+maxAttempts-1)
}

// isPortAvailable 检查端口是否可用
func (wm *WebSocketManager) isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// StartServer 启动WebSocket服务器
func (wm *WebSocketManager) StartServer() error {
	if wm.isRunning {
		return fmt.Errorf("WebSocket服务器已在运行")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wm.handleWebSocket)

	wm.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", wm.port),
		Handler: mux,
	}

	go func() {
		log.Printf("WebSocket服务器启动，端口: %d", wm.port)
		if err := wm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket服务器错误: %v", err)
		}
	}()

	wm.isRunning = true
	return nil
}

// StopServer 停止WebSocket服务器
func (wm *WebSocketManager) StopServer() error {
	if !wm.isRunning {
		return nil
	}

	wm.cancel()

	// 关闭所有客户端连接
	wm.mutex.Lock()
	for _, client := range wm.clients {
		close(client.Send)
		client.Conn.Close()
	}
	wm.clients = make(map[string]*WSClient)
	wm.sessions = make(map[string]map[string]bool)
	wm.mutex.Unlock()

	// 停止HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := wm.server.Shutdown(ctx)
	wm.isRunning = false

	return err
}

// GetPort 获取WebSocket端口
func (wm *WebSocketManager) GetPort() int {
	return wm.port
}

// GetStatus 获取WebSocket服务状态
func (wm *WebSocketManager) GetStatus() string {
	if wm.isRunning {
		return "available"
	}
	return "stopped"
}

// GetInfo 获取WebSocket信息
func (wm *WebSocketManager) GetInfo() map[string]interface{} {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	return map[string]interface{}{
		"port":         wm.port,
		"status":       wm.GetStatus(),
		"clientCount":  len(wm.clients),
		"sessionCount": len(wm.sessions),
		"message":      "WebSocket服务运行正常",
	}
}

// handleWebSocket 处理WebSocket连接
func (wm *WebSocketManager) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &WSClient{
		ID:            fmt.Sprintf("client_%d", time.Now().UnixNano()),
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Manager:       wm,
		LastPing:      time.Now(),
		Subscriptions: make(map[string]bool),
	}

	wm.mutex.Lock()
	wm.clients[client.ID] = client
	wm.mutex.Unlock()

	log.Printf("新客户端连接: %s", client.ID)

	// 启动读写协程
	go client.readPump()
	go client.writePump()
}

// BroadcastToSession 向订阅特定会话的客户端广播消息
func (wm *WebSocketManager) BroadcastToSession(sessionID string, message []byte) {
	wm.mutex.RLock()
	clientsMap, exists := wm.sessions[sessionID]
	if !exists {
		wm.mutex.RUnlock()
		return
	}

	// 复制客户端列表避免长时间锁定
	clientIDs := make([]string, 0, len(clientsMap))
	for clientID := range clientsMap {
		clientIDs = append(clientIDs, clientID)
	}
	wm.mutex.RUnlock()

	// 向所有订阅的客户端发送消息
	for _, clientID := range clientIDs {
		wm.mutex.RLock()
		client, exists := wm.clients[clientID]
		wm.mutex.RUnlock()

		if exists {
			select {
			case client.Send <- message:
			default:
				// 发送通道满，异步关闭客户端避免死锁
				log.Printf("客户端 %s 发送通道满，关闭连接", clientID)
				go wm.removeClient(client)
			}
		}
	}
}

// NotifyTCPMessage 通知TCP消息
func (wm *WebSocketManager) NotifyTCPMessage(sessionID, direction, content string, isHex bool, byteLength int) {
	msgData := wsProtocol.TCPMessageData{
		SessionID:  sessionID,
		Direction:  direction,
		Content:    content,
		IsHex:      isHex,
		ByteLength: byteLength,
		Timestamp:  time.Now().UnixMilli(),
	}

	message := wsProtocol.NewBaseMessage(wsProtocol.MsgTypeTCPMessage, msgData)
	jsonData, err := message.ToJSON()
	if err != nil {
		log.Printf("序列化TCP消息失败: %v", err)
		return
	}

	wm.BroadcastToSession(sessionID, []byte(jsonData))
}

// NotifySessionStatus 通知会话状态变化
func (wm *WebSocketManager) NotifySessionStatus(sessionID, status string) {
	msgData := wsProtocol.SessionStatusData{
		SessionID: sessionID,
		Status:    status,
		Timestamp: time.Now().UnixMilli(),
	}

	message := wsProtocol.NewBaseMessage(wsProtocol.MsgTypeSessionStatus, msgData)
	jsonData, err := message.ToJSON()
	if err != nil {
		log.Printf("序列化会话状态消息失败: %v", err)
		return
	}

	wm.BroadcastToSession(sessionID, []byte(jsonData))
}

// removeClient 移除客户端
func (wm *WebSocketManager) removeClient(client *WSClient) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()

	// 从客户端列表中移除
	delete(wm.clients, client.ID)

	// 从会话订阅中移除
	for sessionID := range client.Subscriptions {
		if clientsMap, exists := wm.sessions[sessionID]; exists {
			delete(clientsMap, client.ID)
			if len(clientsMap) == 0 {
				delete(wm.sessions, sessionID)
			}
		}
	}

	close(client.Send)
	client.Conn.Close()
	log.Printf("客户端 %s 已断开连接", client.ID)
}
