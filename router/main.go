package router

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zhoudm1743/Netser/core"
	"github.com/zhoudm1743/Netser/dto"
	"github.com/zhoudm1743/Netser/dto/session"
)

func Handle(ctx context.Context, data string) (string, error) {
	fmt.Printf("=== 收到请求 ===\n")
	fmt.Printf("请求数据: %s\n", data)

	request := dto.BaseRequest{}
	err := request.Unmarshal(data)
	if err != nil {
		fmt.Printf("数据解析失败: %v\n", err)
		return "", fmt.Errorf("数据解析失败: %v", err)
	}

	fmt.Printf("解析后请求: %+v\n", request)

	switch request.Name {
	case "get_version":
		return dto.Success("1.0.1"), nil

	case "minimize":
		runtime.WindowMinimise(ctx)
		return dto.Success(nil, "窗口最小化成功"), nil

	case "maximize":
		if runtime.WindowIsMaximised(ctx) {
			runtime.WindowUnmaximise(ctx)
		} else {
			runtime.WindowMaximise(ctx)
		}
		return dto.Success(nil, "窗口状态切换成功"), nil

	case "close":
		runtime.Quit(ctx)
		return dto.Success(nil, "应用程序关闭"), nil

	case "connect":
		return handleConnect(request.Data)

	case "disconnect":
		return handleDisconnect(request.Data)

	case "send_data":
		return handleSendData(request.Data)

	case "create_session":
		return handleCreateSession(request.Data)

	case "get_sessions":
		return handleGetSessions()

	case "remove_session":
		return handleRemoveSession(request.Data)

	case "get_session_messages":
		return handleGetSessionMessages(request.Data)

	case "clear_session_messages":
		return handleClearSessionMessages(request.Data)

	case "get_ws_info":
		return handleGetWSInfo()

	case "get_serial_ports":
		return handleGetSerialPorts()

	default:
		return dto.Error("未知的请求类型: " + request.Name), nil
	}
}

// handleConnect 处理连接请求
func handleConnect(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var connectData struct {
		SessionID   string              `json:"sessionId"`
		SessionData session.SessionInfo `json:"sessionData"`
	}

	err = json.Unmarshal(dataBytes, &connectData)
	if err != nil {
		return dto.Error("连接数据解析失败"), nil
	}

	// 如果没有直接的sessionId，从sessionData中获取
	sessionID := connectData.SessionID
	if sessionID == "" {
		sessionID = connectData.SessionData.SessionID
	}

	fmt.Printf("连接会话ID: %s, 类型: %s\n", sessionID, connectData.SessionData.Type)

	// 根据会话类型执行不同的连接逻辑
	switch connectData.SessionData.Type {
	case "tcpClient":
		err = core.GlobalTCPManager.ConnectTCP(
			sessionID,
			connectData.SessionData.Host,
			connectData.SessionData.Port,
			5, // 默认超时5秒
		)
	case "tcpServer":
		err = core.GlobalTCPManager.ListenTCP(
			sessionID,
			connectData.SessionData.Port,
		)
	case "serial":
		// 对于串口，暂时使用默认参数
		err = core.GlobalSerialManager.ConnectSerial(
			sessionID,
			connectData.SessionData.SerialPort, // 串口名称
			9600,                               // 默认波特率
			8,                                  // 默认数据位
			1,                                  // 默认停止位
			"none",                             // 默认无校验
		)
	default:
		return dto.Error("不支持的会话类型"), nil
	}

	if err != nil {
		return dto.Error(fmt.Sprintf("连接失败: %v", err)), nil
	}

	// 更新会话状态
	var newStatus string
	if connectData.SessionData.Type == "tcpClient" {
		newStatus = "connected"
	} else if connectData.SessionData.Type == "tcpServer" {
		newStatus = "listening"
	} else if connectData.SessionData.Type == "serial" {
		newStatus = "connected"
	}

	core.GlobalSessionManager.UpdateSessionStatus(sessionID, newStatus)

	// 获取更新后的会话信息
	updatedSession, _ := core.GlobalSessionManager.GetSession(sessionID)

	return dto.Success(updatedSession, "连接成功"), nil
}

// handleDisconnect 处理断开连接请求
func handleDisconnect(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var disconnectData struct {
		SessionID string `json:"sessionId"`
	}

	err = json.Unmarshal(dataBytes, &disconnectData)
	if err != nil {
		return dto.Error("断开连接数据解析失败"), nil
	}

	err = core.GlobalTCPManager.DisconnectTCP(disconnectData.SessionID)
	if err != nil {
		return dto.Error(fmt.Sprintf("断开连接失败: %v", err)), nil
	}

	// 更新会话状态为断开连接
	core.GlobalSessionManager.UpdateSessionStatus(disconnectData.SessionID, "disconnected")

	// 获取更新后的会话信息
	updatedSession, _ := core.GlobalSessionManager.GetSession(disconnectData.SessionID)

	return dto.Success(updatedSession, "断开连接成功"), nil
}

// handleSendData 处理发送数据请求
func handleSendData(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var sendData struct {
		SessionID string `json:"sessionId"`
		Data      string `json:"data"`
		IsHex     bool   `json:"isHex"`
	}

	err = json.Unmarshal(dataBytes, &sendData)
	if err != nil {
		return dto.Error("发送数据解析失败"), nil
	}

	// 获取会话信息以确定类型
	sess, err := core.GlobalSessionManager.GetSession(sendData.SessionID)
	if err != nil {
		return dto.Error(fmt.Sprintf("会话不存在: %v", err)), nil
	}

	var record *session.MessageRecord

	// 根据会话类型选择不同的发送方式
	switch sess.Info.Type {
	case "tcpClient", "tcpServer":
		record, err = core.GlobalTCPManager.SendTCPData(sendData.SessionID, sendData.Data, sendData.IsHex)
	case "serial":
		record, err = core.GlobalSerialManager.SendSerialData(sendData.SessionID, sendData.Data, sendData.IsHex)
	default:
		return dto.Error("不支持的会话类型"), nil
	}

	if err != nil {
		return dto.Error(fmt.Sprintf("发送数据失败: %v", err)), nil
	}

	return dto.Success(record, "数据发送成功"), nil
}

// handleCreateSession 处理创建会话请求
func handleCreateSession(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var sessionData struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Host    string `json:"host"`
		Port    int    `json:"port"`
		IsHex   bool   `json:"isHex"`
		Timeout int    `json:"timeout"`
	}

	err = json.Unmarshal(dataBytes, &sessionData)
	if err != nil {
		return dto.Error("会话数据解析失败"), nil
	}

	sessionID, err := core.GlobalTCPManager.CreateTCPSession(
		sessionData.Name,
		sessionData.Type,
		sessionData.Host,
		sessionData.Port,
		sessionData.IsHex,
		sessionData.Timeout,
	)

	if err != nil {
		return dto.Error(fmt.Sprintf("创建会话失败: %v", err)), nil
	}

	// 获取创建的会话信息
	sess, err := core.GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		return dto.Error("获取会话信息失败"), nil
	}

	return dto.Success(sess.Info, "会话创建成功"), nil
}

// handleGetSessions 处理获取会话列表请求
func handleGetSessions() (string, error) {
	sessions := core.GlobalSessionManager.GetAllSessions()

	response := session.SessionListResponse{
		Sessions: sessions,
	}

	return dto.Success(response, "获取会话列表成功"), nil
}

// handleRemoveSession 处理移除会话请求
func handleRemoveSession(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var removeData struct {
		SessionID string `json:"sessionId"`
	}

	err = json.Unmarshal(dataBytes, &removeData)
	if err != nil {
		return dto.Error("移除会话数据解析失败"), nil
	}

	err = core.GlobalSessionManager.RemoveSession(removeData.SessionID)
	if err != nil {
		return dto.Error(fmt.Sprintf("移除会话失败: %v", err)), nil
	}

	return dto.Success(nil, "会话移除成功"), nil
}

// handleGetSessionMessages 处理获取会话消息请求
func handleGetSessionMessages(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var messageData struct {
		SessionID string `json:"sessionId"`
		Limit     int    `json:"limit"`
		Offset    int    `json:"offset"`
	}

	err = json.Unmarshal(dataBytes, &messageData)
	if err != nil {
		return dto.Error("消息数据解析失败"), nil
	}

	// 如果没有sessionId，返回空消息列表
	if messageData.SessionID == "" {
		response := session.SessionHistoryResponse{
			SessionID: "",
			Records:   []session.MessageRecord{},
			Total:     0,
		}
		return dto.Success(response, "无会话ID"), nil
	}

	sess, err := core.GlobalSessionManager.GetSession(messageData.SessionID)
	if err != nil {
		return dto.Error("会话不存在"), nil
	}

	messages := sess.GetMessages(messageData.Limit, messageData.Offset)

	// 获取消息总数
	var total int
	if core.GlobalMessageDBManager != nil {
		db, err := core.GlobalMessageDBManager.GetOrCreateMessageDB(messageData.SessionID)
		if err == nil {
			total, _ = db.GetMessageCount()
		}
	}

	response := session.SessionHistoryResponse{
		SessionID: messageData.SessionID,
		Records:   messages,
		Total:     total,
	}

	return dto.Success(response, "获取消息记录成功"), nil
}

// handleClearSessionMessages 处理清空会话消息请求
func handleClearSessionMessages(data any) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return dto.Error("数据格式错误"), nil
	}

	var clearData struct {
		SessionID string `json:"sessionId"`
	}

	err = json.Unmarshal(dataBytes, &clearData)
	if err != nil {
		return dto.Error("清空消息数据解析失败"), nil
	}

	sess, err := core.GlobalSessionManager.GetSession(clearData.SessionID)
	if err != nil {
		return dto.Error("会话不存在"), nil
	}

	sess.ClearMessages()

	return dto.Success(nil, "消息记录清空成功"), nil
}

// handleGetWSInfo 处理获取WebSocket信息请求
func handleGetWSInfo() (string, error) {
	if core.GlobalWebSocketManager == nil {
		return dto.Error("WebSocket服务未初始化"), nil
	}

	wsInfo := core.GlobalWebSocketManager.GetInfo()
	return dto.Success(wsInfo, "获取WebSocket信息成功"), nil
}

// handleGetSerialPorts 处理获取串口列表请求
func handleGetSerialPorts() (string, error) {
	ports, err := core.GlobalSerialManager.GetSerialPorts()
	if err != nil {
		return dto.Error(fmt.Sprintf("获取串口列表失败: %v", err)), nil
	}

	response := map[string]interface{}{
		"ports": ports,
	}

	return dto.Success(response, "获取串口列表成功"), nil
}
