package core

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/zhoudm1743/Netser/dto/session"
)

// TCPManager TCP连接管理器
type TCPManager struct{}

var GlobalTCPManager = &TCPManager{}

// ConnectTCP TCP客户端连接
func (tm *TCPManager) ConnectTCP(sessionID, host string, port int, timeout int) error {
	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 更新状态为连接中
	GlobalSessionManager.UpdateSessionStatus(sessionID, "connecting")

	// 建立连接
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if err != nil {
		GlobalSessionManager.UpdateSessionStatus(sessionID, "disconnected")
		return fmt.Errorf("连接失败: %v", err)
	}

	sess.Connection = conn
	sess.IsActive = true
	GlobalSessionManager.UpdateSessionStatus(sessionID, "connected")

	// 启动接收数据的协程
	go tm.handleTCPReceive(sess)

	return nil
}

// ListenTCP TCP服务端监听
func (tm *TCPManager) ListenTCP(sessionID string, port int) error {
	fmt.Printf("开始监听TCP，会话ID: %s, 端口: %d\n", sessionID, port)

	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		fmt.Printf("获取会话失败: %v\n", err)
		return err
	}

	// 更新状态为连接中
	GlobalSessionManager.UpdateSessionStatus(sessionID, "connecting")
	fmt.Printf("会话状态更新为连接中\n")

	// 开始监听
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("监听失败: %v\n", err)
		GlobalSessionManager.UpdateSessionStatus(sessionID, "disconnected")
		return fmt.Errorf("监听失败: %v", err)
	}

	sess.Listener = listener
	sess.IsActive = true
	GlobalSessionManager.UpdateSessionStatus(sessionID, "listening")
	fmt.Printf("TCP监听成功，端口: %d, 状态更新为listening\n", port)

	// 启动接受连接的协程
	go tm.handleTCPAccept(sess)

	return nil
}

// DisconnectTCP 断开TCP连接
func (tm *TCPManager) DisconnectTCP(sessionID string) error {
	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	sess.IsActive = false

	// 关闭连接
	if sess.Connection != nil {
		sess.Connection.Close()
		sess.Connection = nil
	}

	// 关闭监听器
	if sess.Listener != nil {
		sess.Listener.Close()
		sess.Listener = nil
	}

	GlobalSessionManager.UpdateSessionStatus(sessionID, "disconnected")
	return nil
}

// SendTCPData 发送TCP数据
func (tm *TCPManager) SendTCPData(sessionID, data string, isHex bool) (*session.MessageRecord, error) {
	sess, err := GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if sess.Connection == nil {
		return nil, fmt.Errorf("连接未建立")
	}

	var sendData []byte
	if isHex {
		// 处理十六进制数据
		cleanHex := strings.ReplaceAll(data, " ", "")
		sendData, err = hex.DecodeString(cleanHex)
		if err != nil {
			return nil, fmt.Errorf("十六进制数据格式错误: %v", err)
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

// handleTCPReceive 处理TCP接收数据
func (tm *TCPManager) handleTCPReceive(sess *Session) {
	defer func() {
		if sess.Connection != nil {
			sess.Connection.Close()
			sess.Connection = nil
		}
		sess.IsActive = false
		GlobalSessionManager.UpdateSessionStatus(sess.Info.SessionID, "disconnected")
	}()

	buffer := make([]byte, 4096)
	for sess.IsActive {
		if sess.Connection == nil {
			break
		}

		// 设置读取超时（仅对TCP连接有效）
		if tcpConn, ok := sess.Connection.(net.Conn); ok {
			tcpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		}

		n, err := sess.Connection.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // 超时继续循环
			}
			if err != io.EOF {
				fmt.Printf("读取数据错误: %v\n", err)
			}
			break
		}

		if n > 0 {
			data := string(buffer[:n])
			sess.AddMessage("receive", data, false)

			// 通知WebSocket客户端
			if GlobalWebSocketManager != nil {
				GlobalWebSocketManager.NotifyTCPMessage(sess.Info.SessionID, "receive", data, false, len(data))
			}
		}
	}
}

// handleTCPAccept 处理TCP服务端接受连接
func (tm *TCPManager) handleTCPAccept(sess *Session) {
	defer func() {
		if sess.Listener != nil {
			sess.Listener.Close()
			sess.Listener = nil
		}
		sess.IsActive = false
		GlobalSessionManager.UpdateSessionStatus(sess.Info.SessionID, "disconnected")
	}()

	for sess.IsActive {
		if sess.Listener == nil {
			break
		}

		conn, err := sess.Listener.Accept()
		if err != nil {
			if sess.IsActive {
				fmt.Printf("接受连接错误: %v\n", err)
			}
			break
		}

		// 如果已有连接，关闭旧连接
		if sess.Connection != nil {
			sess.Connection.Close()
		}

		sess.Connection = conn
		GlobalSessionManager.UpdateSessionStatus(sess.Info.SessionID, "connected")

		// 启动处理这个连接的接收数据协程
		go tm.handleTCPReceive(sess)
	}
}

// CreateTCPSession 创建TCP会话
func (tm *TCPManager) CreateTCPSession(name, sessionType, host string, port int, isHex bool, timeout int) (string, error) {
	sessionID := fmt.Sprintf("tcp_%d", time.Now().UnixNano())

	info := session.SessionInfo{
		SessionID:   sessionID,
		Type:        sessionType,
		Name:        name,
		Status:      "disconnected",
		Host:        host,
		Port:        port,
		Protocol:    "tcp",
		IsHex:       isHex,
		ConnectTime: 0,
	}

	GlobalSessionManager.CreateSession(info)
	return sessionID, nil
}
