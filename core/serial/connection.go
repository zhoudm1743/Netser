package serial

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"github.com/zhoudm1743/Netser/core/session"
)

// Connection 表示串口连接
type Connection struct {
	ID             string                 // 连接ID
	port           *Port                  // 关联的串口
	eventBus       *event.EventBus        // 事件总线
	session        *session.Session       // 关联的会话
	closed         bool                   // 是否已关闭
	mutex          sync.RWMutex           // 读写锁
	readBuffer     *bufio.Reader          // 缓冲读取器
	writeBuffer    *bufio.Writer          // 缓冲写入器
	hexMode        bool                   // 是否使用HEX模式
	readTimeout    time.Duration          // 读取超时
	writeTimeout   time.Duration          // 写入超时
	metadata       map[string]interface{} // 连接元数据
	onDataReceived func([]byte)           // 数据接收回调
}

// NewConnection 创建新的串口连接
func NewConnection(port *Port, eventBus *event.EventBus) *Connection {
	connID := fmt.Sprintf("serial-%s-%d", port.GetSettings().PortName, time.Now().UnixNano())

	// 读写缓冲
	// 注意：实际的读写操作由Port对象处理
	readBuffer := bufio.NewReader(nil)
	writeBuffer := bufio.NewWriter(nil)

	settings := port.GetSettings()

	c := &Connection{
		ID:           connID,
		port:         port,
		eventBus:     eventBus,
		closed:       false,
		readBuffer:   readBuffer,
		writeBuffer:  writeBuffer,
		hexMode:      settings.HexMode,
		readTimeout:  time.Duration(settings.ReadTimeout) * time.Millisecond,
		writeTimeout: time.Duration(settings.WriteTimeout) * time.Millisecond,
		metadata:     make(map[string]interface{}),
	}

	// 订阅串口数据接收事件
	if eventBus != nil {
		eventBus.Subscribe(EventDataReceived, func(data interface{}) {
			if event, ok := data.(*event.Event); ok {
				c.handleDataReceived(event)
			}
		})
	}

	return c
}

// 处理数据接收事件
func (c *Connection) handleDataReceived(event *event.Event) {
	if event.Data == nil {
		return
	}

	// 检查是否为此连接的事件
	eventData, ok := event.Data.(map[string]interface{})
	if !ok {
		return
	}

	// 检查端口名称是否匹配
	portName, ok := eventData["portName"].(string)
	if !ok || portName != c.port.GetSettings().PortName {
		return
	}

	// 提取数据
	var data []byte
	if hexData, ok := eventData["hexData"].(string); ok && c.hexMode {
		// 十六进制数据处理
		data, _ = hex.DecodeString(hexData)
	} else {
		// TODO: 从eventData中获取原始数据
		// 注意：这里需要根据实际情况进行调整
	}

	// 调用回调函数
	if c.onDataReceived != nil && len(data) > 0 {
		c.onDataReceived(data)
	}
}

// SetOnDataReceived 设置数据接收回调
func (c *Connection) SetOnDataReceived(callback func([]byte)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onDataReceived = callback
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.port.Close()
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed || !c.port.IsOpen()
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
	c.port.SetHexMode(hexMode)
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

	// 更新串口设置
	settings := c.port.GetSettings()
	settings.ReadTimeout = int(read / time.Millisecond)
	settings.WriteTimeout = int(write / time.Millisecond)
	c.port.SetSettings(settings)
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

// GetPortName 获取端口名称
func (c *Connection) GetPortName() string {
	settings := c.port.GetSettings()
	return settings.PortName
}

// GetBaudRate 获取波特率
func (c *Connection) GetBaudRate() int {
	settings := c.port.GetSettings()
	return settings.BaudRate
}

// Write 写入数据
func (c *Connection) Write(data []byte) (int, error) {
	if c.IsClosed() {
		return 0, fmt.Errorf("connection closed")
	}

	return c.port.Write(data)
}

// WriteString 写入字符串
func (c *Connection) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

// WriteHex 写入十六进制字符串
func (c *Connection) WriteHex(hexStr string) (int, error) {
	if c.IsClosed() {
		return 0, fmt.Errorf("connection closed")
	}

	return c.port.WriteHex(hexStr)
}

// ReadBytes 读取指定字节数
// 注意：这个方法并不是真正的阻塞读，而是从最近收到的数据中读取
// 实际数据接收是通过事件和回调完成的
func (c *Connection) ReadBytes(delim byte) ([]byte, error) {
	if c.IsClosed() {
		return nil, fmt.Errorf("connection closed")
	}

	// 这里需要根据实际情况实现
	// 一般来说，串口通信会通过事件回调处理数据，而不是同步读取
	return nil, fmt.Errorf("not implemented: use OnDataReceived callback instead")
}

// Read 读取指定长度的数据
// 注意：这个方法并不是真正的阻塞读，而是从最近收到的数据中读取
func (c *Connection) Read(size int) ([]byte, error) {
	if c.IsClosed() {
		return nil, fmt.Errorf("connection closed")
	}

	// 同上，需要通过事件回调处理数据
	return nil, fmt.Errorf("not implemented: use OnDataReceived callback instead")
}

// ReadString 读取一行数据
func (c *Connection) ReadString(delim byte) (string, error) {
	if c.IsClosed() {
		return "", fmt.Errorf("connection closed")
	}

	// 同上，需要通过事件回调处理数据
	return "", fmt.Errorf("not implemented: use OnDataReceived callback instead")
}

// ReadHex 读取十六进制数据
func (c *Connection) ReadHex(size int) (string, error) {
	data, err := c.Read(size)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// ApplySettings 应用设置
func (c *Connection) ApplySettings(settings *Settings) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 更新连接设置
	c.hexMode = settings.HexMode
	c.readTimeout = time.Duration(settings.ReadTimeout) * time.Millisecond
	c.writeTimeout = time.Duration(settings.WriteTimeout) * time.Millisecond

	// 更新串口设置
	c.port.SetSettings(settings)

	return nil
}
