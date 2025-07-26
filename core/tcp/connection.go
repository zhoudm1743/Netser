package tcp

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"github.com/zhoudm1743/Netser/core/session"
)

// 定义TCP连接相关的事件
const (
	EventConnected       = "tcp:connected"        // 连接建立
	EventDisconnected    = "tcp:disconnected"     // 连接断开
	EventDataReceived    = "tcp:data_received"    // 收到数据
	EventDataSent        = "tcp:data_sent"        // 发送数据
	EventError           = "tcp:error"            // 连接错误
	EventClientConnected = "tcp:client_connected" // 新客户端连接
)

// Connection 表示TCP连接
type Connection struct {
	ID           string                 // 连接ID
	conn         net.Conn               // 底层TCP连接
	eventBus     *event.EventBus        // 事件总线
	session      *session.Session       // 关联的会话
	closed       bool                   // 是否已关闭
	mutex        sync.RWMutex           // 读写锁
	reader       *bufio.Reader          // 缓冲读取器
	writer       *bufio.Writer          // 缓冲写入器
	hexMode      bool                   // 是否使用HEX模式
	readTimeout  time.Duration          // 读取超时
	writeTimeout time.Duration          // 写入超时
	metadata     map[string]interface{} // 连接元数据
}

// NewConnection 创建新的TCP连接
func NewConnection(conn net.Conn, eventBus *event.EventBus) *Connection {
	connID := fmt.Sprintf("%s-%s-%d",
		conn.LocalAddr().String(),
		conn.RemoteAddr().String(),
		time.Now().UnixNano())

	c := &Connection{
		ID:           connID,
		conn:         conn,
		eventBus:     eventBus,
		closed:       false,
		reader:       bufio.NewReader(conn),
		writer:       bufio.NewWriter(conn),
		hexMode:      false,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		metadata:     make(map[string]interface{}),
	}

	// 发布连接建立事件
	if eventBus != nil {
		eventBus.Publish(EventConnected, event.NewEvent(
			EventConnected,
			map[string]interface{}{
				"connectionId": c.ID,
				"localAddr":    conn.LocalAddr().String(),
				"remoteAddr":   conn.RemoteAddr().String(),
			},
			"tcp_connection",
		))
	}

	return c
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	err := c.conn.Close()

	// 发布连接断开事件
	if c.eventBus != nil {
		c.eventBus.Publish(EventDisconnected, event.NewEvent(
			EventDisconnected,
			map[string]interface{}{
				"connectionId": c.ID,
				"localAddr":    c.conn.LocalAddr().String(),
				"remoteAddr":   c.conn.RemoteAddr().String(),
			},
			"tcp_connection",
		))
	}

	return err
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed
}

// SetSession 设置关联会话
func (c *Connection) SetSession(session *session.Session) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.session = session
}

// GetSession 获取关联会话
func (c *Connection) GetSession() *session.Session {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.session
}

// SetHexMode 设置HEX模式
func (c *Connection) SetHexMode(hexMode bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.hexMode = hexMode
}

// IsHexMode 是否为HEX模式
func (c *Connection) IsHexMode() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.hexMode
}

// SetTimeout 设置超时
func (c *Connection) SetTimeout(read, write time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.readTimeout = read
	c.writeTimeout = write
}

// SetMetadata 设置元数据
func (c *Connection) SetMetadata(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metadata[key] = value
}

// GetMetadata 获取元数据
func (c *Connection) GetMetadata(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.metadata[key]
	return val, ok
}

// LocalAddr 获取本地地址
func (c *Connection) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

// RemoteAddr 获取远程地址
func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

// ReadBytes 读取字节
func (c *Connection) ReadBytes(delim byte) ([]byte, error) {
	if c.IsClosed() {
		return nil, fmt.Errorf("connection closed")
	}

	// 设置读取超时
	if c.readTimeout > 0 {
		err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			return nil, err
		}
	}

	data, err := c.reader.ReadBytes(delim)

	if err != nil {
		// 发布错误事件
		if c.eventBus != nil && err != io.EOF {
			c.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"connectionId": c.ID,
					"error":        err.Error(),
					"operation":    "read",
				},
				"tcp_connection",
			))
		}
		return data, err
	}

	// 发布数据接收事件
	if c.eventBus != nil {
		eventData := map[string]interface{}{
			"connectionId": c.ID,
			"size":         len(data),
		}

		// HEX模式下添加十六进制表示
		if c.hexMode {
			eventData["hexData"] = hex.EncodeToString(data)
		}

		c.eventBus.Publish(EventDataReceived, event.NewEvent(
			EventDataReceived,
			eventData,
			"tcp_connection",
		))
	}

	return data, nil
}

// Read 读取指定长度的数据
func (c *Connection) Read(size int) ([]byte, error) {
	if c.IsClosed() {
		return nil, fmt.Errorf("connection closed")
	}

	// 设置读取超时
	if c.readTimeout > 0 {
		err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			return nil, err
		}
	}

	data := make([]byte, size)
	n, err := io.ReadFull(c.reader, data)
	if err != nil {
		// 发布错误事件
		if c.eventBus != nil && err != io.EOF {
			c.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"connectionId": c.ID,
					"error":        err.Error(),
					"operation":    "read",
				},
				"tcp_connection",
			))
		}
		return data[:n], err
	}

	// 发布数据接收事件
	if c.eventBus != nil {
		eventData := map[string]interface{}{
			"connectionId": c.ID,
			"size":         n,
		}

		// HEX模式下添加十六进制表示
		if c.hexMode {
			eventData["hexData"] = hex.EncodeToString(data[:n])
		}

		c.eventBus.Publish(EventDataReceived, event.NewEvent(
			EventDataReceived,
			eventData,
			"tcp_connection",
		))
	}

	return data, nil
}

// ReadString 读取一行数据
func (c *Connection) ReadString(delim byte) (string, error) {
	if c.IsClosed() {
		return "", fmt.Errorf("connection closed")
	}

	// 设置读取超时
	if c.readTimeout > 0 {
		err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			return "", err
		}
	}

	data, err := c.reader.ReadString(delim)

	if err != nil {
		// 发布错误事件
		if c.eventBus != nil && err != io.EOF {
			c.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"connectionId": c.ID,
					"error":        err.Error(),
					"operation":    "read",
				},
				"tcp_connection",
			))
		}
		return data, err
	}

	// 发布数据接收事件
	if c.eventBus != nil {
		eventData := map[string]interface{}{
			"connectionId": c.ID,
			"size":         len(data),
		}

		// HEX模式下添加十六进制表示
		if c.hexMode {
			eventData["hexData"] = hex.EncodeToString([]byte(data))
		}

		c.eventBus.Publish(EventDataReceived, event.NewEvent(
			EventDataReceived,
			eventData,
			"tcp_connection",
		))
	}

	return data, nil
}

// Write 写入数据
func (c *Connection) Write(data []byte) (int, error) {
	if c.IsClosed() {
		return 0, fmt.Errorf("connection closed")
	}

	// 设置写入超时
	if c.writeTimeout > 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
		if err != nil {
			return 0, err
		}
	}

	n, err := c.writer.Write(data)
	if err != nil {
		// 发布错误事件
		if c.eventBus != nil {
			c.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"connectionId": c.ID,
					"error":        err.Error(),
					"operation":    "write",
				},
				"tcp_connection",
			))
		}
		return n, err
	}

	// 刷新缓冲区
	err = c.writer.Flush()
	if err != nil {
		return n, err
	}

	// 发布数据发送事件
	if c.eventBus != nil {
		eventData := map[string]interface{}{
			"connectionId": c.ID,
			"size":         n,
		}

		// HEX模式下添加十六进制表示
		if c.hexMode {
			eventData["hexData"] = hex.EncodeToString(data[:n])
		}

		c.eventBus.Publish(EventDataSent, event.NewEvent(
			EventDataSent,
			eventData,
			"tcp_connection",
		))
	}

	return n, nil
}

// WriteHex 写入十六进制字符串
func (c *Connection) WriteHex(hexStr string) (int, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		// 发布错误事件
		if c.eventBus != nil {
			c.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"connectionId": c.ID,
					"error":        "invalid hex string: " + err.Error(),
					"operation":    "write_hex",
				},
				"tcp_connection",
			))
		}
		return 0, fmt.Errorf("invalid hex string: %v", err)
	}

	return c.Write(data)
}

// WriteString 写入字符串
func (c *Connection) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

// ReadHex 读取十六进制数据
func (c *Connection) ReadHex(size int) (string, error) {
	data, err := c.Read(size)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}
