package core

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	wsProtocol "github.com/zhoudm1743/Netser/dto/websocket"
)

const (
	// 心跳间隔
	pingPeriod = 30 * time.Second
	// 读取超时
	pongWait = 60 * time.Second
	// 写入超时
	writeWait = 10 * time.Second
	// 最大消息大小 (64KB)
	maxMessageSize = 64 * 1024
)

// readPump 读取消息协程
func (c *WSClient) readPump() {
	defer func() {
		c.Manager.removeClient(c)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastPing = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 处理收到的消息
		c.handleMessage(messageBytes)
	}
}

// writePump 写入消息协程
func (c *WSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 发送通道关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket写入错误: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket心跳发送失败: %v", err)
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *WSClient) handleMessage(messageBytes []byte) {
	log.Printf("客户端 %s 收到原始消息: %s", c.ID, string(messageBytes))

	// 解析基础消息
	message, err := wsProtocol.ParseMessage(messageBytes)
	if err != nil {
		log.Printf("解析WebSocket消息失败: %v", err)
		c.sendError(wsProtocol.StatusInvalidMessage, "消息格式错误", err.Error())
		return
	}

	log.Printf("收到客户端 %s 的消息: %s (ID: %s)", c.ID, message.Type, message.ID)

	switch message.Type {
	case wsProtocol.MsgTypeAuth:
		c.handleAuth(message)
	case wsProtocol.MsgTypeSubscribe:
		c.handleSubscribe(message)
	case wsProtocol.MsgTypeUnsubscribe:
		c.handleUnsubscribe(message)
	case wsProtocol.MsgTypePing:
		c.handlePing(message)
	default:
		c.sendError(wsProtocol.StatusInvalidMessage, "未知消息类型", string(message.Type))
	}
}

// handleAuth 处理认证消息
func (c *WSClient) handleAuth(message *wsProtocol.BaseMessage) {
	// 解析认证数据
	authDataBytes, err := json.Marshal(message.Data)
	if err != nil {
		c.sendError(wsProtocol.StatusInvalidMessage, "认证数据格式错误", "")
		return
	}

	var authData wsProtocol.AuthData
	err = json.Unmarshal(authDataBytes, &authData)
	if err != nil {
		c.sendError(wsProtocol.StatusAuthFailed, "认证数据解析失败", "")
		return
	}

	// 简单验证（生产环境需要更严格的认证）
	if authData.ClientID == "" {
		c.sendError(wsProtocol.StatusAuthFailed, "客户端ID不能为空", "")
		return
	}

	// 更新客户端ID
	c.Manager.mutex.Lock()
	delete(c.Manager.clients, c.ID)
	c.ID = authData.ClientID
	c.Manager.clients[c.ID] = c
	c.Manager.mutex.Unlock()

	// 发送认证成功响应
	responseData := map[string]interface{}{
		"clientId":      authData.ClientID,
		"serverVersion": "1.0",
	}

	response := wsProtocol.NewResponseMessage(
		wsProtocol.MsgTypeAuthResponse,
		message.ID,
		wsProtocol.StatusSuccess,
		"认证成功",
		responseData,
	)

	log.Printf("发送认证响应给客户端 %s", c.ID)
	c.sendMessage(response)
	log.Printf("客户端 %s 认证成功", c.ID)
}

// handleSubscribe 处理订阅消息
func (c *WSClient) handleSubscribe(message *wsProtocol.BaseMessage) {
	// 解析订阅数据
	subscribeDataBytes, err := json.Marshal(message.Data)
	if err != nil {
		c.sendError(wsProtocol.StatusInvalidMessage, "订阅数据格式错误", "")
		return
	}

	var subscribeData wsProtocol.SubscribeData
	err = json.Unmarshal(subscribeDataBytes, &subscribeData)
	if err != nil {
		c.sendError(wsProtocol.StatusSubscribeFailed, "订阅数据解析失败", "")
		return
	}

	sessionID := subscribeData.SessionID
	if sessionID == "" {
		c.sendError(wsProtocol.StatusSubscribeFailed, "会话ID不能为空", "")
		return
	}

	// 检查会话是否存在
	_, err = GlobalSessionManager.GetSession(sessionID)
	if err != nil {
		c.sendError(wsProtocol.StatusSessionNotFound, "会话不存在", sessionID)
		return
	}

	// 添加订阅
	c.Manager.mutex.Lock()
	if c.Manager.sessions[sessionID] == nil {
		c.Manager.sessions[sessionID] = make(map[string]bool)
	}
	c.Manager.sessions[sessionID][c.ID] = true
	c.Subscriptions[sessionID] = true
	c.Manager.mutex.Unlock()

	// 发送订阅成功响应
	responseData := map[string]interface{}{
		"sessionId": sessionID,
		"status":    "subscribed",
	}

	response := wsProtocol.NewResponseMessage(
		wsProtocol.MsgTypeSubscribe,
		message.ID,
		wsProtocol.StatusSuccess,
		"订阅成功",
		responseData,
	)

	c.sendMessage(response)
	log.Printf("客户端 %s 订阅会话 %s", c.ID, sessionID)
}

// handleUnsubscribe 处理取消订阅消息
func (c *WSClient) handleUnsubscribe(message *wsProtocol.BaseMessage) {
	// 解析取消订阅数据
	unsubscribeDataBytes, err := json.Marshal(message.Data)
	if err != nil {
		c.sendError(wsProtocol.StatusInvalidMessage, "取消订阅数据格式错误", "")
		return
	}

	var unsubscribeData wsProtocol.UnsubscribeData
	err = json.Unmarshal(unsubscribeDataBytes, &unsubscribeData)
	if err != nil {
		c.sendError(wsProtocol.StatusInvalidMessage, "取消订阅数据解析失败", "")
		return
	}

	sessionID := unsubscribeData.SessionID
	if sessionID == "" {
		c.sendError(wsProtocol.StatusInvalidMessage, "会话ID不能为空", "")
		return
	}

	// 移除订阅
	c.Manager.mutex.Lock()
	if clientsMap, exists := c.Manager.sessions[sessionID]; exists {
		delete(clientsMap, c.ID)
		if len(clientsMap) == 0 {
			delete(c.Manager.sessions, sessionID)
		}
	}
	delete(c.Subscriptions, sessionID)
	c.Manager.mutex.Unlock()

	// 发送取消订阅成功响应
	responseData := map[string]interface{}{
		"sessionId": sessionID,
		"status":    "unsubscribed",
	}

	response := wsProtocol.NewResponseMessage(
		wsProtocol.MsgTypeUnsubscribe,
		message.ID,
		wsProtocol.StatusSuccess,
		"取消订阅成功",
		responseData,
	)

	c.sendMessage(response)
	log.Printf("客户端 %s 取消订阅会话 %s", c.ID, sessionID)
}

// handlePing 处理心跳消息
func (c *WSClient) handlePing(message *wsProtocol.BaseMessage) {
	// 发送心跳响应
	response := wsProtocol.NewResponseMessage(
		wsProtocol.MsgTypePong,
		message.ID,
		wsProtocol.StatusSuccess,
		"pong",
		nil,
	)

	c.sendMessage(response)
}

// sendMessage 发送消息
func (c *WSClient) sendMessage(message *wsProtocol.ResponseMessage) {
	jsonData, err := message.ToJSON()
	if err != nil {
		log.Printf("序列化响应消息失败: %v", err)
		return
	}

	log.Printf("发送响应消息给客户端 %s: %s", c.ID, jsonData)

	select {
	case c.Send <- []byte(jsonData):
	default:
		log.Printf("客户端 %s 发送通道满，关闭连接", c.ID)
		go c.Manager.removeClient(c)
	}
}

// sendError 发送错误消息
func (c *WSClient) sendError(code wsProtocol.StatusCode, message, details string) {
	errorData := wsProtocol.ErrorData{
		Code:    code,
		Message: message,
		Details: details,
	}

	errorMessage := wsProtocol.NewBaseMessage(wsProtocol.MsgTypeError, errorData)
	jsonData, err := errorMessage.ToJSON()
	if err != nil {
		log.Printf("序列化错误消息失败: %v", err)
		return
	}

	select {
	case c.Send <- []byte(jsonData):
	default:
		log.Printf("客户端 %s 发送通道满，无法发送错误消息", c.ID)
	}
}
