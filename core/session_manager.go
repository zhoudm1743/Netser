package core

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/dto/session"
)

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// Session 会话结构
type Session struct {
	Info       session.SessionInfo
	Connection io.ReadWriteCloser // 支持TCP和串口连接
	Listener   net.Listener       // 仅用于TCP服务端
	IsActive   bool
	CreatedAt  time.Time
	mutex      sync.RWMutex
}

var GlobalSessionManager = &SessionManager{
	sessions: make(map[string]*Session),
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(info session.SessionInfo) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sess := &Session{
		Info:      info,
		IsActive:  false,
		CreatedAt: time.Now(),
	}

	sm.sessions[info.SessionID] = sess
	return sess
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}
	return sess, nil
}

// RemoveSession 移除会话
func (sm *SessionManager) RemoveSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 关闭连接
	if sess.Connection != nil {
		sess.Connection.Close()
	}
	if sess.Listener != nil {
		sess.Listener.Close()
	}

	delete(sm.sessions, sessionID)

	// 清理会话数据库
	if GlobalMessageDBManager != nil {
		err := GlobalMessageDBManager.CloseSessionDB(sessionID)
		if err != nil {
			log.Printf("清理会话数据库失败: %v", err)
		}
	}

	return nil
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []session.SessionInfo {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessions := make([]session.SessionInfo, 0, len(sm.sessions))
	for _, sess := range sm.sessions {
		sessions = append(sessions, sess.Info)
	}
	return sessions
}

// UpdateSessionStatus 更新会话状态
func (sm *SessionManager) UpdateSessionStatus(sessionID, status string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	sess.Info.Status = status

	// 通知WebSocket客户端状态变化
	if GlobalWebSocketManager != nil {
		GlobalWebSocketManager.NotifySessionStatus(sessionID, status)
	}

	return nil
}

// AddMessage 添加消息记录
func (s *Session) AddMessage(direction, data string, isHex bool) {
	record := session.MessageRecord{
		Direction:  direction,
		Data:       data,
		IsHex:      isHex,
		Timestamp:  time.Now().UnixMilli(),
		ByteLength: len(data),
	}

	// 存储到数据库
	log.Printf("存储消息到数据库: 会话=%s, 方向=%s, 数据=%s", s.Info.SessionID, direction, data)
	err := StoreMessageToDB(s.Info.SessionID, record)
	if err != nil {
		log.Printf("存储消息到数据库失败: %v", err)
	} else {
		log.Printf("消息存储成功")
	}
}

// GetMessages 获取消息记录
func (s *Session) GetMessages(limit, offset int) []session.MessageRecord {
	log.Printf("获取消息: 会话=%s, limit=%d, offset=%d", s.Info.SessionID, limit, offset)
	messages, err := GetMessagesFromDB(s.Info.SessionID, limit, offset)
	if err != nil {
		log.Printf("从数据库获取消息失败: %v", err)
		return []session.MessageRecord{}
	}
	log.Printf("从数据库获取到 %d 条消息", len(messages))
	return messages
}

// ClearMessages 清空消息记录
func (s *Session) ClearMessages() {
	err := ClearSessionMessagesInDB(s.Info.SessionID)
	if err != nil {
		log.Printf("清空数据库消息失败: %v", err)
	}
}
