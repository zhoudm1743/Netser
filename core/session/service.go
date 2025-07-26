package session

import (
	"fmt"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
)

// SessionService 提供会话管理服务
type SessionService struct {
	manager     *SessionManager
	storage     SessionStorage
	eventBus    *event.EventBus
	cleanupTick time.Duration
	stopCleanup chan bool
}

// NewSessionService 创建新的会话服务
func NewSessionService(timeout time.Duration, storage SessionStorage, eventBus *event.EventBus) *SessionService {
	service := &SessionService{
		manager:     NewSessionManager(timeout),
		storage:     storage,
		eventBus:    eventBus,
		cleanupTick: 5 * time.Minute, // 默认每5分钟清理一次
		stopCleanup: make(chan bool),
	}

	// 加载存储的会话
	service.loadSessions()

	// 启动定时清理
	go service.startCleanupTask()

	return service
}

// loadSessions 从存储加载会话
func (ss *SessionService) loadSessions() {
	sessions, err := ss.storage.GetAll()
	if err != nil {
		fmt.Printf("加载会话失败: %v\n", err)
		return
	}

	for _, session := range sessions {
		ss.manager.sessions[session.ID] = session
	}
}

// startCleanupTask 启动会话清理任务
func (ss *SessionService) startCleanupTask() {
	ticker := time.NewTicker(ss.cleanupTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := ss.manager.CleanupExpiredSessions()
			if count > 0 && ss.eventBus != nil {
				ss.eventBus.Publish("session:cleanup", event.NewEvent(
					"session:cleanup",
					map[string]interface{}{"expiredCount": count},
					"session_service",
				))
			}
		case <-ss.stopCleanup:
			return
		}
	}
}

// Shutdown 关闭会话服务
func (ss *SessionService) Shutdown() {
	ss.stopCleanup <- true
}

// CreateUserSession 创建用户会话
func (ss *SessionService) CreateUserSession(userID string, userData map[string]interface{}) (*Session, error) {
	session := ss.manager.CreateSession(userID)

	// 添加用户数据
	for key, value := range userData {
		session.Data[key] = value
	}

	// 保存到存储
	if err := ss.storage.Save(session); err != nil {
		return nil, fmt.Errorf("保存会话失败: %v", err)
	}

	return session, nil
}

// GetSession 获取会话
func (ss *SessionService) GetSession(sessionID string) (*Session, error) {
	session, exists := ss.manager.GetSession(sessionID)
	if exists {
		return session, nil
	}

	// 从存储中尝试加载
	session, err := ss.storage.Load(sessionID)
	if err != nil {
		return nil, err
	}

	// 加入内存
	ss.manager.sessions[session.ID] = session

	return session, nil
}

// RefreshSession 刷新会话
func (ss *SessionService) RefreshSession(sessionID string) error {
	if !ss.manager.RefreshSession(sessionID) {
		return fmt.Errorf("刷新会话失败: %s", sessionID)
	}

	session, exists := ss.manager.GetSession(sessionID)
	if !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	// 更新存储
	if err := ss.storage.Save(session); err != nil {
		return fmt.Errorf("保存刷新会话失败: %v", err)
	}

	// 发布会话刷新事件
	if ss.eventBus != nil {
		ss.eventBus.Publish(EventSessionRefreshed, event.NewEvent(
			EventSessionRefreshed,
			map[string]interface{}{
				"sessionId": session.ID,
				"userId":    session.UserID,
				"expiresAt": session.ExpiresAt,
			},
			"session_service",
		))
	}

	return nil
}

// DestroySession 销毁会话
func (ss *SessionService) DestroySession(sessionID string) error {
	// 从内存删除
	ss.manager.DestroySession(sessionID)

	// 从存储删除
	return ss.storage.Delete(sessionID)
}

// SetSessionData 设置会话数据
func (ss *SessionService) SetSessionData(sessionID string, key string, value interface{}) error {
	if !ss.manager.SetSessionData(sessionID, key, value) {
		return fmt.Errorf("设置会话数据失败: %s", sessionID)
	}

	// 更新存储
	session, exists := ss.manager.GetSession(sessionID)
	if !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	return ss.storage.Save(session)
}

// GetSessionData 获取会话数据
func (ss *SessionService) GetSessionData(sessionID string, key string) (interface{}, error) {
	value, exists := ss.manager.GetSessionData(sessionID, key)
	if !exists {
		return nil, fmt.Errorf("获取会话数据失败: %s.%s", sessionID, key)
	}

	return value, nil
}

// GetActiveSessionCount 获取活跃会话数
func (ss *SessionService) GetActiveSessionCount() int {
	return len(ss.manager.GetAllSessions())
}

// GetSessionsByUserID 获取用户的所有会话
func (ss *SessionService) GetSessionsByUserID(userID string) []*Session {
	allSessions := ss.manager.GetAllSessions()
	userSessions := make([]*Session, 0)

	for _, session := range allSessions {
		if session.UserID == userID {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions
}

// DestroyUserSessions 销毁用户的所有会话
func (ss *SessionService) DestroyUserSessions(userID string) int {
	userSessions := ss.GetSessionsByUserID(userID)
	count := 0

	for _, session := range userSessions {
		if err := ss.DestroySession(session.ID); err == nil {
			count++
		}
	}

	return count
}
