package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
)

// 会话相关事件
const (
	EventSessionCreated   = "session:created"
	EventSessionDestroyed = "session:destroyed"
	EventSessionExpired   = "session:expired"
	EventSessionRefreshed = "session:refreshed"
)

// SessionStorage 会话存储接口
type SessionStorage interface {
	Save(session *Session) error
	Load(sessionID string) (*Session, error)
	Delete(sessionID string) error
	GetAll() ([]*Session, error)
}

// MemoryStorage 内存会话存储
type MemoryStorage struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewMemoryStorage 创建新的内存存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		sessions: make(map[string]*Session),
	}
}

// Save 保存会话到内存
func (ms *MemoryStorage) Save(session *Session) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.sessions[session.ID] = session
	return nil
}

// Load 从内存加载会话
func (ms *MemoryStorage) Load(sessionID string) (*Session, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	session, exists := ms.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}

	return session, nil
}

// Delete 从内存删除会话
func (ms *MemoryStorage) Delete(sessionID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.sessions[sessionID]; !exists {
		return fmt.Errorf("会话不存在: %s", sessionID)
	}

	delete(ms.sessions, sessionID)
	return nil
}

// GetAll 获取所有会话
func (ms *MemoryStorage) GetAll() ([]*Session, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	sessions := make([]*Session, 0, len(ms.sessions))
	for _, session := range ms.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// FileStorage 文件会话存储
type FileStorage struct {
	directory string
	mutex     sync.Mutex
	eventBus  *event.EventBus // 事件总线
}

// NewFileStorage 创建新的文件存储
func NewFileStorage(directory string, eventBus *event.EventBus) (*FileStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("创建会话目录失败: %v", err)
	}

	return &FileStorage{
		directory: directory,
		eventBus:  eventBus,
	}, nil
}

// sessionFilePath 获取会话文件路径
func (fs *FileStorage) sessionFilePath(sessionID string) string {
	return filepath.Join(fs.directory, fmt.Sprintf("session_%s.json", sessionID))
}

// Save 保存会话到文件
func (fs *FileStorage) Save(session *Session) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("序列化会话失败: %v", err)
	}

	err = ioutil.WriteFile(fs.sessionFilePath(session.ID), data, 0644)
	if err != nil {
		return fmt.Errorf("写入会话文件失败: %v", err)
	}

	// 发布会话创建事件
	if fs.eventBus != nil {
		fs.eventBus.Publish(EventSessionCreated, event.NewEvent(
			EventSessionCreated,
			map[string]interface{}{
				"sessionId": session.ID,
				"userId":    session.UserID,
				"expiresAt": session.ExpiresAt,
			},
			"session_storage",
		))
	}

	return nil
}

// Load 从文件加载会话
func (fs *FileStorage) Load(sessionID string) (*Session, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	data, err := ioutil.ReadFile(fs.sessionFilePath(sessionID))
	if err != nil {
		return nil, fmt.Errorf("读取会话文件失败: %v", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("反序列化会话失败: %v", err)
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		// 发布会话过期事件
		if fs.eventBus != nil {
			fs.eventBus.Publish(EventSessionExpired, event.NewEvent(
				EventSessionExpired,
				map[string]interface{}{
					"sessionId": session.ID,
					"userId":    session.UserID,
					"expiresAt": session.ExpiresAt,
				},
				"session_storage",
			))
		}

		// 删除过期会话
		os.Remove(fs.sessionFilePath(sessionID))
		return nil, fmt.Errorf("会话已过期: %s", sessionID)
	}

	return &session, nil
}

// Delete 从文件删除会话
func (fs *FileStorage) Delete(sessionID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 先加载会话以获取用户信息
	data, err := ioutil.ReadFile(fs.sessionFilePath(sessionID))

	// 即使文件不存在也继续尝试删除
	if err == nil {
		var session Session
		if json.Unmarshal(data, &session) == nil && fs.eventBus != nil {
			// 发布会话销毁事件
			fs.eventBus.Publish(EventSessionDestroyed, event.NewEvent(
				EventSessionDestroyed,
				map[string]interface{}{
					"sessionId": sessionID,
					"userId":    session.UserID,
				},
				"session_storage",
			))
		}
	}

	err = os.Remove(fs.sessionFilePath(sessionID))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除会话文件失败: %v", err)
	}

	return nil
}

// GetAll 获取所有会话
func (fs *FileStorage) GetAll() ([]*Session, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	files, err := ioutil.ReadDir(fs.directory)
	if err != nil {
		return nil, fmt.Errorf("读取会话目录失败: %v", err)
	}

	var sessions []*Session
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(fs.directory, file.Name()))
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		// 跳过过期会话
		if time.Now().After(session.ExpiresAt) {
			// 删除过期会话文件
			os.Remove(filepath.Join(fs.directory, file.Name()))
			continue
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}
