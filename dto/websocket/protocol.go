package websocket

import (
	"encoding/json"
	"time"
)

// MessageType 消息类型枚举
type MessageType string

const (
	// 基础消息类型
	MsgTypeAuth         MessageType = "auth"          // 认证消息
	MsgTypeAuthResponse MessageType = "auth_response" // 认证响应
	MsgTypeSubscribe    MessageType = "subscribe"     // 订阅会话消息
	MsgTypeUnsubscribe  MessageType = "unsubscribe"   // 取消订阅
	MsgTypePing         MessageType = "ping"          // 心跳检测
	MsgTypePong         MessageType = "pong"          // 心跳响应

	// 业务消息类型
	MsgTypeTCPMessage    MessageType = "tcp_message"    // TCP消息推送
	MsgTypeSessionStatus MessageType = "session_status" // 会话状态变化
	MsgTypeSystemNotify  MessageType = "system_notify"  // 系统通知
	MsgTypeError         MessageType = "error"          // 错误消息
)

// StatusCode 状态码
type StatusCode int

const (
	StatusSuccess         StatusCode = 0    // 成功
	StatusInvalidMessage  StatusCode = 4001 // 消息格式错误
	StatusAuthFailed      StatusCode = 4002 // 认证失败
	StatusSessionNotFound StatusCode = 4003 // 会话不存在
	StatusInternalError   StatusCode = 5001 // 内部错误
	StatusSubscribeFailed StatusCode = 4004 // 订阅失败
)

// BaseMessage WebSocket基础消息结构
type BaseMessage struct {
	Type      MessageType `json:"type"`           // 消息类型
	ID        string      `json:"id,omitempty"`   // 消息ID（用于请求-响应匹配）
	Timestamp int64       `json:"timestamp"`      // 时间戳（毫秒）
	Data      interface{} `json:"data,omitempty"` // 消息数据
}

// ResponseMessage 响应消息结构
type ResponseMessage struct {
	Type      MessageType `json:"type"`              // 消息类型
	ID        string      `json:"id,omitempty"`      // 对应请求的消息ID
	Code      StatusCode  `json:"code"`              // 状态码
	Message   string      `json:"message,omitempty"` // 状态描述
	Timestamp int64       `json:"timestamp"`         // 时间戳（毫秒）
	Data      interface{} `json:"data,omitempty"`    // 响应数据
}

// AuthData 认证数据
type AuthData struct {
	ClientID  string `json:"clientId"`  // 客户端ID
	Version   string `json:"version"`   // 协议版本
	SessionID string `json:"sessionId"` // 要订阅的会话ID
}

// SubscribeData 订阅数据
type SubscribeData struct {
	SessionID string `json:"sessionId"` // 会话ID
}

// UnsubscribeData 取消订阅数据
type UnsubscribeData struct {
	SessionID string `json:"sessionId"` // 会话ID
}

// TCPMessageData TCP消息数据
type TCPMessageData struct {
	SessionID  string `json:"sessionId"`  // 会话ID
	Direction  string `json:"direction"`  // 方向: send/receive
	Content    string `json:"content"`    // 消息内容
	IsHex      bool   `json:"isHex"`      // 是否为十六进制
	ByteLength int    `json:"byteLength"` // 字节长度
	Timestamp  int64  `json:"timestamp"`  // 时间戳（毫秒）
}

// SessionStatusData 会话状态数据
type SessionStatusData struct {
	SessionID string `json:"sessionId"` // 会话ID
	Status    string `json:"status"`    // 状态: connected/disconnected/listening/connecting
	Timestamp int64  `json:"timestamp"` // 时间戳（毫秒）
}

// SystemNotifyData 系统通知数据
type SystemNotifyData struct {
	Level     string `json:"level"`               // 级别: info/warning/error
	Title     string `json:"title"`               // 标题
	Message   string `json:"message"`             // 消息内容
	SessionID string `json:"sessionId,omitempty"` // 相关会话ID（可选）
}

// ErrorData 错误数据
type ErrorData struct {
	Code      StatusCode `json:"code"`                // 错误码
	Message   string     `json:"message"`             // 错误消息
	Details   string     `json:"details,omitempty"`   // 错误详情
	SessionID string     `json:"sessionId,omitempty"` // 相关会话ID（可选）
}

// NewBaseMessage 创建基础消息
func NewBaseMessage(msgType MessageType, data interface{}) *BaseMessage {
	return &BaseMessage{
		Type:      msgType,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

// NewResponseMessage 创建响应消息
func NewResponseMessage(msgType MessageType, id string, code StatusCode, message string, data interface{}) *ResponseMessage {
	return &ResponseMessage{
		Type:      msgType,
		ID:        id,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

// ToJSON 转换为JSON字符串
func (m *BaseMessage) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSON 转换为JSON字符串
func (r *ResponseMessage) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseMessage 解析消息
func ParseMessage(data []byte) (*BaseMessage, error) {
	var msg BaseMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetStatusMessage 获取状态码对应的消息
func GetStatusMessage(code StatusCode) string {
	switch code {
	case StatusSuccess:
		return "成功"
	case StatusInvalidMessage:
		return "消息格式错误"
	case StatusAuthFailed:
		return "认证失败"
	case StatusSessionNotFound:
		return "会话不存在"
	case StatusInternalError:
		return "内部错误"
	case StatusSubscribeFailed:
		return "订阅失败"
	default:
		return "未知错误"
	}
}
