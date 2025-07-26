package tcp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"github.com/zhoudm1743/Netser/core/session"
)

// 定义TCP客户端相关事件
const (
	EventClientConnecting   = "tcp:client_connecting"   // 客户端连接中
	EventClientConnectedTo  = "tcp:client_connected"    // 客户端已连接
	EventClientDisconnected = "tcp:client_disconnected" // 客户端已断开
	EventClientReconnecting = "tcp:client_reconnecting" // 客户端重连中
	EventClientError        = "tcp:client_error"        // 客户端错误
)

// ClientConfig TCP客户端配置
type ClientConfig struct {
	Host                 string        // 服务器主机
	Port                 int           // 服务器端口
	ReadTimeout          time.Duration // 读取超时
	WriteTimeout         time.Duration // 写入超时
	ConnectTimeout       time.Duration // 连接超时
	ReconnectDelay       time.Duration // 重连延迟
	MaxReconnectAttempts int           // 最大重连尝试次数
	AutoReconnect        bool          // 是否自动重连
	BufferSize           int           // 缓冲区大小
}

// Client TCP客户端
type Client struct {
	config         ClientConfig            // 客户端配置
	connection     *Connection             // TCP连接
	eventBus       *event.EventBus         // 事件总线
	sessionService *session.SessionService // 会话服务
	mutex          sync.RWMutex            // 读写锁
	connected      bool                    // 是否已连接
	stopReconnect  chan bool               // 停止重连信号
	reconnecting   bool                    // 是否正在重连
	reconnectCount int                     // 重连次数
	dataHandlers   []func([]byte)          // 数据处理器
}

// DefaultClientConfig 默认客户端配置
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Host:                 "localhost",
		Port:                 8080,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		ConnectTimeout:       10 * time.Second,
		ReconnectDelay:       5 * time.Second,
		MaxReconnectAttempts: 5,
		AutoReconnect:        true,
		BufferSize:           1024,
	}
}

// NewClient 创建新的TCP客户端
func NewClient(config ClientConfig, eventBus *event.EventBus, sessionService *session.SessionService) *Client {
	return &Client{
		config:         config,
		eventBus:       eventBus,
		sessionService: sessionService,
		connected:      false,
		stopReconnect:  make(chan bool),
		reconnecting:   false,
		reconnectCount: 0,
		dataHandlers:   make([]func([]byte), 0),
	}
}

// Connect 连接到服务器
func (c *Client) Connect() error {
	c.mutex.Lock()
	if c.connected {
		c.mutex.Unlock()
		return fmt.Errorf("client already connected")
	}

	// 发布连接中事件
	if c.eventBus != nil {
		c.eventBus.Publish(EventClientConnecting, event.NewEvent(
			EventClientConnecting,
			map[string]interface{}{
				"host": c.config.Host,
				"port": c.config.Port,
			},
			"tcp_client",
		))
	}

	c.mutex.Unlock()

	// 创建连接
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// 设置连接超时
	dialer := net.Dialer{Timeout: c.config.ConnectTimeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		// 发布错误事件
		if c.eventBus != nil {
			c.eventBus.Publish(EventClientError, event.NewEvent(
				EventClientError,
				map[string]interface{}{
					"error":     err.Error(),
					"operation": "connect",
				},
				"tcp_client",
			))
		}

		// 如果配置了自动重连，启动重连过程
		if c.config.AutoReconnect {
			go c.reconnect()
		}

		return fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	// 创建连接对象
	connection := NewConnection(conn, c.eventBus)
	connection.SetTimeout(c.config.ReadTimeout, c.config.WriteTimeout)

	c.mutex.Lock()
	c.connection = connection
	c.connected = true
	c.reconnectCount = 0 // 重置重连计数
	c.mutex.Unlock()

	// 发布连接成功事件
	if c.eventBus != nil {
		c.eventBus.Publish(EventClientConnectedTo, event.NewEvent(
			EventClientConnectedTo,
			map[string]interface{}{
				"connectionId": connection.ID,
				"localAddr":    conn.LocalAddr().String(),
				"remoteAddr":   conn.RemoteAddr().String(),
			},
			"tcp_client",
		))
	}

	// 创建会话
	if c.sessionService != nil {
		sessionData := map[string]interface{}{
			"connectionId": connection.ID,
			"remoteAddr":   conn.RemoteAddr().String(),
			"localAddr":    conn.LocalAddr().String(),
			"connectedAt":  time.Now(),
			"type":         "tcp_client",
		}

		session, err := c.sessionService.CreateUserSession("tcp_client", sessionData)
		if err == nil {
			connection.SetSession(session)
		}
	}

	// 启动接收循环
	go c.readLoop()

	return nil
}

// readLoop 接收数据循环
func (c *Client) readLoop() {
	c.mutex.RLock()
	if c.connection == nil || c.connection.IsClosed() {
		c.mutex.RUnlock()
		return
	}
	conn := c.connection
	bufferSize := c.config.BufferSize
	c.mutex.RUnlock()

	buffer := make([]byte, bufferSize)

	for {
		c.mutex.RLock()
		if c.connection == nil || c.connection.IsClosed() {
			c.mutex.RUnlock()
			break
		}
		c.mutex.RUnlock()

		// 读取数据
		n, err := conn.conn.Read(buffer)
		if err != nil {
			c.mutex.RLock()
			connected := c.connected
			c.mutex.RUnlock()

			if connected {
				// 处理断开连接
				c.handleDisconnect(err)
			}
			break
		}

		// 处理接收到的数据
		data := make([]byte, n)
		copy(data, buffer[:n])

		// 调用所有数据处理器
		c.mutex.RLock()
		handlers := make([]func([]byte), len(c.dataHandlers))
		copy(handlers, c.dataHandlers)
		c.mutex.RUnlock()

		for _, handler := range handlers {
			go handler(data)
		}
	}
}

// handleDisconnect 处理断开连接
func (c *Client) handleDisconnect(err error) {
	c.mutex.Lock()

	if !c.connected {
		c.mutex.Unlock()
		return
	}

	// 设置连接状态
	c.connected = false

	// 保存连接ID以便在事件中使用
	var connectionId string
	if c.connection != nil {
		connectionId = c.connection.ID
		c.connection.Close()
	}
	c.mutex.Unlock()

	// 发布断开连接事件
	if c.eventBus != nil {
		c.eventBus.Publish(EventClientDisconnected, event.NewEvent(
			EventClientDisconnected,
			map[string]interface{}{
				"connectionId": connectionId,
				"error":        err.Error(),
			},
			"tcp_client",
		))
	}

	// 如果配置了自动重连，启动重连过程
	if c.config.AutoReconnect {
		go c.reconnect()
	}
}

// reconnect 尝试重新连接
func (c *Client) reconnect() {
	c.mutex.Lock()

	// 检查是否已经在重连
	if c.reconnecting {
		c.mutex.Unlock()
		return
	}

	c.reconnecting = true
	c.mutex.Unlock()

	// 退出时重置重连状态
	defer func() {
		c.mutex.Lock()
		c.reconnecting = false
		c.mutex.Unlock()
	}()

	for {
		c.mutex.Lock()
		// 检查重连次数
		if c.config.MaxReconnectAttempts > 0 && c.reconnectCount >= c.config.MaxReconnectAttempts {
			// 超过最大重连次数
			c.mutex.Unlock()

			// 发布重连失败事件
			if c.eventBus != nil {
				c.eventBus.Publish(EventClientError, event.NewEvent(
					EventClientError,
					map[string]interface{}{
						"error":     "max reconnect attempts reached",
						"operation": "reconnect",
					},
					"tcp_client",
				))
			}
			return
		}

		// 增加重连计数
		c.reconnectCount++
		count := c.reconnectCount
		c.mutex.Unlock()

		// 发布重连事件
		if c.eventBus != nil {
			c.eventBus.Publish(EventClientReconnecting, event.NewEvent(
				EventClientReconnecting,
				map[string]interface{}{
					"host":    c.config.Host,
					"port":    c.config.Port,
					"attempt": count,
				},
				"tcp_client",
			))
		}

		// 等待重连延迟
		select {
		case <-time.After(c.config.ReconnectDelay):
			// 尝试重新连接
			err := c.Connect()
			if err == nil {
				// 连接成功
				return
			}
		case <-c.stopReconnect:
			// 收到停止信号
			return
		}
	}
}

// Disconnect 断开连接
func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 停止自动重连
	if c.reconnecting {
		c.stopReconnect <- true
	}

	if !c.connected || c.connection == nil {
		return fmt.Errorf("client not connected")
	}

	c.connected = false
	connectionId := c.connection.ID
	err := c.connection.Close()
	c.connection = nil

	// 发布断开连接事件
	if c.eventBus != nil {
		c.eventBus.Publish(EventClientDisconnected, event.NewEvent(
			EventClientDisconnected,
			map[string]interface{}{
				"connectionId": connectionId,
				"error":        "manual disconnect",
			},
			"tcp_client",
		))
	}

	return err
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected && c.connection != nil && !c.connection.IsClosed()
}

// Send 发送数据
func (c *Client) Send(data []byte) (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected || c.connection == nil {
		return 0, fmt.Errorf("client not connected")
	}

	return c.connection.Write(data)
}

// SendHex 发送十六进制数据
func (c *Client) SendHex(hexStr string) (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected || c.connection == nil {
		return 0, fmt.Errorf("client not connected")
	}

	return c.connection.WriteHex(hexStr)
}

// SendString 发送字符串
func (c *Client) SendString(s string) (int, error) {
	return c.Send([]byte(s))
}

// SetHexMode 设置HEX模式
func (c *Client) SetHexMode(hexMode bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected && c.connection != nil {
		c.connection.SetHexMode(hexMode)
	}
}

// RegisterDataHandler 注册数据处理器
func (c *Client) RegisterDataHandler(handler func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dataHandlers = append(c.dataHandlers, handler)
}

// GetLocalAddr 获取本地地址
func (c *Client) GetLocalAddr() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.connected && c.connection != nil {
		return c.connection.LocalAddr()
	}

	return ""
}

// GetRemoteAddr 获取远程地址
func (c *Client) GetRemoteAddr() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.connected && c.connection != nil {
		return c.connection.RemoteAddr()
	}

	return fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
}

// GetConnection 获取连接对象
func (c *Client) GetConnection() *Connection {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connection
}
