package serial

import (
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"go.bug.st/serial"
)

// Port 串口对象
type Port struct {
	settings    *Settings              // 串口设置
	port        serial.Port            // 底层串口对象
	isOpen      bool                   // 是否已打开
	eventBus    *event.EventBus        // 事件总线
	mutex       sync.RWMutex           // 读写锁
	readBuffer  []byte                 // 读取缓冲区
	closeSignal chan struct{}          // 关闭信号
	metadata    map[string]interface{} // 元数据
}

// NewPort 创建新的串口对象
func NewPort(eventBus *event.EventBus) *Port {
	settings := NewDefaultSettings()
	return &Port{
		settings:    settings,
		isOpen:      false,
		eventBus:    eventBus,
		readBuffer:  make([]byte, settings.BufferSize),
		closeSignal: make(chan struct{}),
		metadata:    make(map[string]interface{}),
	}
}

// SetSettings 设置串口参数
func (p *Port) SetSettings(settings *Settings) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.settings = settings.Clone()

	// 如果端口已打开，则重新打开以应用新设置
	if p.isOpen {
		p.Close()
		p.Open()
	}
}

// GetSettings 获取串口设置
func (p *Port) GetSettings() *Settings {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.settings.Clone()
}

// IsHexMode 是否为十六进制模式
func (p *Port) IsHexMode() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.settings.HexMode
}

// SetHexMode 设置十六进制模式
func (p *Port) SetHexMode(hexMode bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.settings.HexMode = hexMode
}

// Open 打开串口
func (p *Port) Open() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isOpen {
		return fmt.Errorf("serial port already open")
	}

	// 配置串口参数
	config := serial.Mode{
		BaudRate: p.settings.BaudRate,
		DataBits: p.settings.DataBits,
		Parity:   mapParity(p.settings.Parity),
		StopBits: mapStopBits(p.settings.StopBits),
	}

	// 打开串口
	port, err := serial.Open(p.settings.PortName, &config)
	if err != nil {
		if p.eventBus != nil {
			p.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"error":     err.Error(),
					"operation": "open",
					"portName":  p.settings.PortName,
				},
				"serial_port",
			))
		}
		return fmt.Errorf("failed to open port %s: %v", p.settings.PortName, err)
	}

	// 设置超时
	if p.settings.ReadTimeout > 0 {
		err = port.SetReadTimeout(time.Duration(p.settings.ReadTimeout) * time.Millisecond)
		if err != nil {
			port.Close()
			return fmt.Errorf("failed to set read timeout: %v", err)
		}
	}

	p.port = port
	p.isOpen = true
	p.closeSignal = make(chan struct{})

	// 发布连接事件
	if p.eventBus != nil {
		p.eventBus.Publish(EventConnected, event.NewEvent(
			EventConnected,
			map[string]interface{}{
				"portName": p.settings.PortName,
				"baudRate": p.settings.BaudRate,
			},
			"serial_port",
		))
	}

	// 启动后台读取
	go p.readLoop()

	return nil
}

// Close 关闭串口
func (p *Port) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isOpen {
		return nil
	}

	// 发送关闭信号
	close(p.closeSignal)

	// 关闭串口
	err := p.port.Close()
	p.isOpen = false

	// 发布断开连接事件
	if p.eventBus != nil {
		p.eventBus.Publish(EventDisconnected, event.NewEvent(
			EventDisconnected,
			map[string]interface{}{
				"portName": p.settings.PortName,
			},
			"serial_port",
		))
	}

	return err
}

// IsOpen 检查串口是否打开
func (p *Port) IsOpen() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isOpen
}

// SetMetadata 设置元数据
func (p *Port) SetMetadata(key string, value interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.metadata[key] = value
}

// GetMetadata 获取元数据
func (p *Port) GetMetadata(key string) (interface{}, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	val, ok := p.metadata[key]
	return val, ok
}

// Write 写入数据
func (p *Port) Write(data []byte) (int, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.isOpen {
		return 0, fmt.Errorf("serial port not open")
	}

	// go.bug.st/serial 库的 Port 接口不支持 SetWriteDeadline
	// 因此我们只能依赖于操作系统底层的超时机制

	// 写入数据
	n, err := p.port.Write(data)

	// 发布数据发送事件
	if err == nil && p.eventBus != nil {
		eventData := map[string]interface{}{
			"portName": p.settings.PortName,
			"size":     n,
		}

		// 如果是十六进制模式，添加十六进制表示
		if p.settings.HexMode {
			eventData["hexData"] = hex.EncodeToString(data[:n])
		}

		p.eventBus.Publish(EventDataSent, event.NewEvent(
			EventDataSent,
			eventData,
			"serial_port",
		))
	} else if err != nil && p.eventBus != nil {
		p.eventBus.Publish(EventError, event.NewEvent(
			EventError,
			map[string]interface{}{
				"error":     err.Error(),
				"operation": "write",
				"portName":  p.settings.PortName,
			},
			"serial_port",
		))
	}

	return n, err
}

// WriteString 写入字符串
func (p *Port) WriteString(s string) (int, error) {
	return p.Write([]byte(s))
}

// WriteHex 写入十六进制字符串
func (p *Port) WriteHex(hexStr string) (int, error) {
	// 解码十六进制字符串
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		if p.eventBus != nil {
			p.eventBus.Publish(EventError, event.NewEvent(
				EventError,
				map[string]interface{}{
					"error":     "invalid hex string: " + err.Error(),
					"operation": "write_hex",
					"portName":  p.settings.PortName,
				},
				"serial_port",
			))
		}
		return 0, fmt.Errorf("invalid hex string: %v", err)
	}

	return p.Write(data)
}

// 后台读取循环
func (p *Port) readLoop() {
	buffer := make([]byte, p.settings.BufferSize)

	for {
		select {
		case <-p.closeSignal:
			return
		default:
			// 检查串口是否开启
			if !p.IsOpen() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 读取数据
			n, err := p.port.Read(buffer)

			// 如果有数据，处理并发布事件
			if n > 0 && p.eventBus != nil {
				eventData := map[string]interface{}{
					"portName": p.settings.PortName,
					"size":     n,
				}

				// 如果是十六进制模式，添加十六进制表示
				if p.settings.HexMode {
					eventData["hexData"] = hex.EncodeToString(buffer[:n])
				}

				p.eventBus.Publish(EventDataReceived, event.NewEvent(
					EventDataReceived,
					eventData,
					"serial_port",
				))
			}

			// 处理错误
			if err != nil && err != io.EOF && p.eventBus != nil {
				p.eventBus.Publish(EventError, event.NewEvent(
					EventError,
					map[string]interface{}{
						"error":     err.Error(),
						"operation": "read",
						"portName":  p.settings.PortName,
					},
					"serial_port",
				))

				// 如果是严重错误，关闭串口
				// go.bug.st/serial 库没有定义ErrTimeout常量，使用标准io错误处理
				if err != io.EOF && err != io.ErrShortWrite && err != io.ErrShortBuffer {
					p.mutex.Lock()
					p.isOpen = false
					p.port.Close()
					p.mutex.Unlock()
					return
				}
			}

			// 读取间隔
			if p.settings.ReadInterval > 0 {
				time.Sleep(time.Duration(p.settings.ReadInterval) * time.Millisecond)
			}
		}
	}
}

// GetAvailablePorts 获取可用的串口列表
func GetAvailablePorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}
	return ports, nil
}

// 辅助函数：映射校验位
func mapParity(parity int) serial.Parity {
	switch parity {
	case ParityNone:
		return serial.NoParity
	case ParityOdd:
		return serial.OddParity
	case ParityEven:
		return serial.EvenParity
	case ParityMark:
		return serial.MarkParity
	case ParitySpace:
		return serial.SpaceParity
	default:
		return serial.NoParity
	}
}

// 辅助函数：映射停止位
func mapStopBits(stopBits int) serial.StopBits {
	switch stopBits {
	case StopBits1:
		return serial.OneStopBit
	case StopBits15:
		return serial.OnePointFiveStopBits
	case StopBits2:
		return serial.TwoStopBits
	default:
		return serial.OneStopBit
	}
}
