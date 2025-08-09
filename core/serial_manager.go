package core

import (
	"fmt"
	"log"
	"time"

	"github.com/zhoudm1743/Netser/dto/session"
	"go.bug.st/serial"
)

// SerialManager 串口管理器
type SerialManager struct{}

var GlobalSerialManager = &SerialManager{}

// GetSerialPorts 获取可用串口列表
func (sm *SerialManager) GetSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("获取串口列表失败: %v", err)
	}

	log.Printf("发现 %d 个串口: %v", len(ports), ports)
	return ports, nil
}

// ConnectSerial 连接串口
func (sm *SerialManager) ConnectSerial(sessionID string, portName string, baudRate, dataBits, stopBits int, parity string) error {
	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("会话不存在: %v", err)
	}

	if sess.Connection != nil {
		return fmt.Errorf("串口已连接")
	}

	// 设置串口参数
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: getStopBits(stopBits),
		Parity:   getParity(parity),
	}

	log.Printf("连接串口: %s, 波特率: %d, 数据位: %d, 停止位: %d, 校验: %s",
		portName, baudRate, dataBits, stopBits, parity)

	// 打开串口
	port, err := serial.Open(portName, mode)
	if err != nil {
		return fmt.Errorf("打开串口失败: %v", err)
	}

	// 保存连接 - serial.Port实现了io.ReadWriteCloser接口
	sess.Connection = port
	sess.IsActive = true

	log.Printf("串口 %s 连接成功", portName)

	// 启动接收数据的goroutine
	go sm.handleSerialReceive(sess)

	return nil
}

// DisconnectSerial 断开串口连接
func (sm *SerialManager) DisconnectSerial(sessionID string) error {
	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("会话不存在: %v", err)
	}

	if sess.Connection == nil {
		return fmt.Errorf("串口未连接")
	}

	// 关闭连接
	sess.Connection.Close()
	sess.Connection = nil
	sess.IsActive = false

	log.Printf("串口连接已断开: %s", sessionID)
	return nil
}

// SendSerialData 发送串口数据
func (sm *SerialManager) SendSerialData(sessionID, data string, isHex bool) (*session.MessageRecord, error) {
	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if sess.Connection == nil {
		return nil, fmt.Errorf("串口未连接")
	}

	var sendData []byte
	if isHex {
		// 处理十六进制数据
		// 简单实现：移除空格并转换
		cleanHex := ""
		for _, char := range data {
			if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
				cleanHex += string(char)
			}
		}

		// 转换十六进制字符串为字节
		for i := 0; i < len(cleanHex); i += 2 {
			if i+1 >= len(cleanHex) {
				break
			}
			var b byte
			fmt.Sscanf(cleanHex[i:i+2], "%02x", &b)
			sendData = append(sendData, b)
		}
	} else {
		sendData = []byte(data)
	}

	// 发送数据
	_, err = sess.Connection.Write(sendData)
	if err != nil {
		return nil, fmt.Errorf("发送数据失败: %v", err)
	}

	// 创建消息记录
	record := session.MessageRecord{
		Direction:  "send",
		Data:       data,
		IsHex:      isHex,
		Timestamp:  time.Now().UnixMilli(),
		ByteLength: len(data),
	}

	// 记录发送的消息
	sess.AddMessage("send", data, isHex)

	// 通知WebSocket客户端
	if GlobalWebSocketManager != nil {
		GlobalWebSocketManager.NotifyTCPMessage(sessionID, "send", data, isHex, len(data))
	}

	return &record, nil
}

// handleSerialReceive 处理串口接收数据
func (sm *SerialManager) handleSerialReceive(sess *Session) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("串口接收处理异常: %v", r)
		}
	}()

	buffer := make([]byte, 1024)

	for sess.IsActive && sess.Connection != nil {
		// 设置读取超时
		if closer, ok := sess.Connection.(interface{ SetReadTimeout(time.Duration) error }); ok {
			closer.SetReadTimeout(100 * time.Millisecond)
		}

		n, err := sess.Connection.Read(buffer)
		if err != nil {
			// 超时错误是正常的，继续循环
			if err.Error() == "timeout" {
				continue
			}
			log.Printf("串口读取错误: %v", err)
			break
		}

		if n > 0 {
			data := string(buffer[:n])
			log.Printf("串口收到数据 [%s]: %s (%d字节)", sess.Info.SessionID, data, n)

			// 记录接收的消息
			sess.AddMessage("receive", data, false)

			// 通知WebSocket客户端
			if GlobalWebSocketManager != nil {
				GlobalWebSocketManager.NotifyTCPMessage(sess.Info.SessionID, "receive", data, false, n)
			}
		}
	}

	log.Printf("串口接收处理结束: %s", sess.Info.SessionID)
}

// 辅助函数：转换停止位
func getStopBits(stopBits int) serial.StopBits {
	switch stopBits {
	case 1:
		return serial.OneStopBit
	case 2:
		return serial.TwoStopBits
	default:
		return serial.OneStopBit
	}
}

// 辅助函数：转换校验位
func getParity(parity string) serial.Parity {
	switch parity {
	case "odd":
		return serial.OddParity
	case "even":
		return serial.EvenParity
	case "none":
		fallthrough
	default:
		return serial.NoParity
	}
}
