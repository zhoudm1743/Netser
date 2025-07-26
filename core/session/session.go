package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session 表示一个用户会话
type Session struct {
	ID        string                 // 会话唯一标识符
	UserID    string                 // 关联的用户ID
	CreatedAt time.Time              // 会话创建时间
	ExpiresAt time.Time              // 会话过期时间
	Data      map[string]interface{} // 会话数据
	isValid   bool                   // 会话是否有效
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session // 会话映射表
	mutex    sync.RWMutex        // 读写锁
	timeout  time.Duration       // 会话超时时间
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager(timeout time.Duration) *SessionManager {
	if timeout == 0 {
		// 默认30分钟超时
		timeout = 30 * time.Minute
	}

	return &SessionManager{
		sessions: make(map[string]*Session),
		timeout:  timeout,
	}
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(userID string) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.timeout),
		Data:      make(map[string]interface{}),
		isValid:   true,
	}

	sm.sessions[session.ID] = session
	return session
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) || !session.isValid {
		return nil, false
	}

	return session, true
}

// RefreshSession 刷新会话过期时间
func (sm *SessionManager) RefreshSession(sessionID string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists || !session.isValid {
		return false
	}

	session.ExpiresAt = time.Now().Add(sm.timeout)
	return true
}

// DestroySession 销毁会话
func (sm *SessionManager) DestroySession(sessionID string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.isValid = false
		delete(sm.sessions, sessionID)
		return true
	}

	return false
}

// CleanupExpiredSessions 清理过期会话
func (sm *SessionManager) CleanupExpiredSessions() int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	count := 0

	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) || !session.isValid {
			delete(sm.sessions, id)
			count++
		}
	}

	return count
}

// SetSessionData 设置会话数据
func (sm *SessionManager) SetSessionData(sessionID, key string, value interface{}) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists || !session.isValid {
		return false
	}

	session.Data[key] = value
	return true
}

// GetSessionData 获取会话数据
func (sm *SessionManager) GetSessionData(sessionID, key string) (interface{}, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists || !session.isValid {
		return nil, false
	}

	value, ok := session.Data[key]
	return value, ok
}

// GetAllSessions 获取所有有效会话
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	now := time.Now()
	validSessions := make([]*Session, 0)

	for _, session := range sm.sessions {
		if !now.After(session.ExpiresAt) && session.isValid {
			validSessions = append(validSessions, session)
		}
	}

	return validSessions
}
