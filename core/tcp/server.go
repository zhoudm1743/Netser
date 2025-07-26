package tcp

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"github.com/zhoudm1743/Netser/core/session"
)

// ServerConfig TCP服务器配置
type ServerConfig struct {
	Host             string        // 监听主机
	Port             int           // 监听端口
	MaxConnections   int           // 最大连接数
	ReadTimeout      time.Duration // 读取超时
	WriteTimeout     time.Duration // 写入超时
	HandshakeTimeout time.Duration // 握手超时
	BufferSize       int           // 缓冲区大小
}

// Server TCP服务端
type Server struct {
	config         ServerConfig                         // 服务器配置
	listener       net.Listener                         // TCP监听器
	eventBus       *event.EventBus                      // 事件总线
	sessionService *session.SessionService              // 会话服务
	connections    map[string]*Connection               // 连接管理
	mutex          sync.RWMutex                         // 读写锁
	running        bool                                 // 是否运行中
	handlers       map[string]func(*Connection, []byte) // 数据处理器
}

// DefaultServerConfig 默认服务器配置
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:             "0.0.0.0",
		Port:             8080,
		MaxConnections:   100,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		HandshakeTimeout: 10 * time.Second,
		BufferSize:       1024,
	}
}

// NewServer 创建新的TCP服务器
func NewServer(config ServerConfig, eventBus *event.EventBus, sessionService *session.SessionService) *Server {
	return &Server{
		config:         config,
		eventBus:       eventBus,
		sessionService: sessionService,
		connections:    make(map[string]*Connection),
		running:        false,
		handlers:       make(map[string]func(*Connection, []byte)),
	}
}

// RegisterHandler 注册数据处理器
func (s *Server) RegisterHandler(eventType string, handler func(*Connection, []byte)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handlers[eventType] = handler
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return fmt.Errorf("server already running")
	}

	// 创建监听器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.mutex.Unlock()
		return fmt.Errorf("failed to start server on %s: %v", addr, err)
	}

	s.listener = listener
	s.running = true
	s.mutex.Unlock()

	// 发布服务器启动事件
	if s.eventBus != nil {
		s.eventBus.Publish("tcp:server_started", event.NewEvent(
			"tcp:server_started",
			map[string]interface{}{
				"address": addr,
				"time":    time.Now().Format(time.RFC3339),
			},
			"tcp_server",
		))
	}

	// 接受连接
	go s.acceptConnections()

	return nil
}

// acceptConnections 接受客户端连接
func (s *Server) acceptConnections() {
	for {
		s.mutex.RLock()
		if !s.running || s.listener == nil {
			s.mutex.RUnlock()
			break
		}
		listener := s.listener
		s.mutex.RUnlock()

		// 接受连接
		conn, err := listener.Accept()
		if err != nil {
			// 检查服务器是否已关闭
			s.mutex.RLock()
			running := s.running
			s.mutex.RUnlock()
			if !running {
				break
			}

			// 发布错误事件
			if s.eventBus != nil {
				s.eventBus.Publish(EventError, event.NewEvent(
					EventError,
					map[string]interface{}{
						"error":     err.Error(),
						"operation": "accept",
					},
					"tcp_server",
				))
			}

			// 短暂暂停避免CPU占用过高
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 检查是否超过最大连接数
		s.mutex.RLock()
		connectionCount := len(s.connections)
		maxConnections := s.config.MaxConnections
		s.mutex.RUnlock()

		if maxConnections > 0 && connectionCount >= maxConnections {
			// 达到最大连接数，拒绝连接
			conn.Close()

			// 发布拒绝连接事件
			if s.eventBus != nil {
				s.eventBus.Publish("tcp:connection_rejected", event.NewEvent(
					"tcp:connection_rejected",
					map[string]interface{}{
						"remoteAddr": conn.RemoteAddr().String(),
						"reason":     "max connections reached",
					},
					"tcp_server",
				))
			}
			continue
		}

		// 创建连接并启动处理
		go s.handleConnection(conn)
	}
}

// handleConnection 处理新连接
func (s *Server) handleConnection(conn net.Conn) {
	// 创建Connection对象
	connection := NewConnection(conn, s.eventBus)

	// 设置超时
	connection.SetTimeout(s.config.ReadTimeout, s.config.WriteTimeout)

	// 注册连接
	s.mutex.Lock()
	s.connections[connection.ID] = connection
	s.mutex.Unlock()

	// 发布客户端连接事件
	if s.eventBus != nil {
		s.eventBus.Publish(EventClientConnected, event.NewEvent(
			EventClientConnected,
			map[string]interface{}{
				"connectionId": connection.ID,
				"remoteAddr":   conn.RemoteAddr().String(),
				"localAddr":    conn.LocalAddr().String(),
				"time":         time.Now().Format(time.RFC3339),
			},
			"tcp_server",
		))
	}

	// 创建会话
	if s.sessionService != nil {
		sessionData := map[string]interface{}{
			"connectionId": connection.ID,
			"remoteAddr":   conn.RemoteAddr().String(),
			"connectedAt":  time.Now(),
			"type":         "tcp",
		}

		session, err := s.sessionService.CreateUserSession(conn.RemoteAddr().String(), sessionData)
		if err == nil {
			connection.SetSession(session)
		}
	}

	// 启动接收循环
	buffer := make([]byte, s.config.BufferSize)
	for {
		// 检查连接是否已关闭
		if connection.IsClosed() {
			break
		}

		// 读取数据
		n, err := conn.Read(buffer)
		if err != nil {
			// 处理连接关闭
			s.removeConnection(connection)
			break
		}

		// 处理接收到的数据
		data := buffer[:n]

		// 调用所有注册的处理器
		s.mutex.RLock()
		handlers := make([]func(*Connection, []byte), 0, len(s.handlers))
		for _, handler := range s.handlers {
			handlers = append(handlers, handler)
		}
		s.mutex.RUnlock()

		// 在单独的goroutine中执行处理器，避免阻塞读取循环
		for _, handler := range handlers {
			go handler(connection, data)
		}
	}
}

// removeConnection 移除连接
func (s *Server) removeConnection(connection *Connection) {
	if connection == nil {
		return
	}

	connection.Close()

	s.mutex.Lock()
	delete(s.connections, connection.ID)
	s.mutex.Unlock()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("server not running")
	}

	// 关闭监听器
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return fmt.Errorf("error closing listener: %v", err)
		}
		s.listener = nil
	}

	// 关闭所有连接
	for _, conn := range s.connections {
		conn.Close()
	}

	// 清空连接
	s.connections = make(map[string]*Connection)
	s.running = false

	// 发布服务器停止事件
	if s.eventBus != nil {
		s.eventBus.Publish("tcp:server_stopped", event.NewEvent(
			"tcp:server_stopped",
			map[string]interface{}{
				"time": time.Now().Format(time.RFC3339),
			},
			"tcp_server",
		))
	}

	return nil
}

// IsRunning 检查服务器是否运行中
func (s *Server) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// GetConnections 获取所有连接
func (s *Server) GetConnections() []*Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	connections := make([]*Connection, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}

	return connections
}

// GetConnection 获取指定ID的连接
func (s *Server) GetConnection(connectionID string) (*Connection, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	conn, exists := s.connections[connectionID]
	return conn, exists
}

// GetConnectionCount 获取连接数量
func (s *Server) GetConnectionCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.connections)
}

// BroadcastData 广播数据到所有连接
func (s *Server) BroadcastData(data []byte) int {
	s.mutex.RLock()
	connections := s.GetConnections()
	s.mutex.RUnlock()

	count := 0
	for _, conn := range connections {
		if !conn.IsClosed() {
			_, err := conn.Write(data)
			if err == nil {
				count++
			}
		}
	}

	return count
}

// BroadcastHex 广播十六进制数据到所有连接
func (s *Server) BroadcastHex(hexStr string) int {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0
	}
	return s.BroadcastData(data)
}

// GetAddress 获取服务器地址
func (s *Server) GetAddress() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.listener != nil {
		return s.listener.Addr().String()
	}

	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}
