package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zhoudm1743/Netser/dto/session"
	"go.etcd.io/bbolt"
)

const (
	// 数据库文件扩展名
	DBFileExtension = ".dm"
	// 数据库目录
	DBDirectory = "message_data"
	// 消息bucket名称
	MessageBucket = "messages"
)

// MessageDB 消息数据库管理器
type MessageDB struct {
	sessionID string
	dbPath    string
	db        *bbolt.DB
	mutex     sync.RWMutex
}

// MessageDBManager 全局数据库管理器
type MessageDBManager struct {
	databases map[string]*MessageDB // sessionID -> MessageDB
	mutex     sync.RWMutex
}

var GlobalMessageDBManager *MessageDBManager

// InitMessageDBManager 初始化消息数据库管理器
func InitMessageDBManager() error {
	// 创建数据库目录
	if err := os.MkdirAll(DBDirectory, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	GlobalMessageDBManager = &MessageDBManager{
		databases: make(map[string]*MessageDB),
	}

	log.Printf("消息数据库管理器初始化成功，数据目录: %s", DBDirectory)
	return nil
}

// GetOrCreateMessageDB 获取或创建消息数据库
func (manager *MessageDBManager) GetOrCreateMessageDB(sessionID string) (*MessageDB, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	log.Printf("GetOrCreateMessageDB: 会话ID=%s", sessionID)

	// 如果已存在，直接返回
	if db, exists := manager.databases[sessionID]; exists {
		log.Printf("数据库已存在，直接返回: %s", sessionID)
		return db, nil
	}

	// 确保数据库目录存在
	if err := os.MkdirAll(DBDirectory, 0755); err != nil {
		log.Printf("创建数据库目录失败: %v", err)
		return nil, fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 创建新的数据库
	dbPath := filepath.Join(DBDirectory, sessionID+DBFileExtension)
	log.Printf("创建新数据库: %s", dbPath)

	boltDB, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.Printf("打开数据库失败: %v", err)
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 创建bucket
	err = boltDB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(MessageBucket))
		return err
	})
	if err != nil {
		boltDB.Close()
		return nil, fmt.Errorf("创建bucket失败: %v", err)
	}

	messageDB := &MessageDB{
		sessionID: sessionID,
		dbPath:    dbPath,
		db:        boltDB,
	}

	manager.databases[sessionID] = messageDB
	log.Printf("数据库创建成功！会话 %s, 路径: %s", sessionID, dbPath)

	return messageDB, nil
}

// CloseSessionDB 关闭并删除会话数据库
func (manager *MessageDBManager) CloseSessionDB(sessionID string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	db, exists := manager.databases[sessionID]
	if !exists {
		return nil // 数据库不存在，无需处理
	}

	// 关闭数据库
	if err := db.db.Close(); err != nil {
		log.Printf("关闭数据库失败: %v", err)
	}

	// 删除数据库文件
	if err := os.Remove(db.dbPath); err != nil {
		log.Printf("删除数据库文件失败: %v", err)
	} else {
		log.Printf("已删除会话 %s 的数据库文件: %s", sessionID, db.dbPath)
	}

	// 从管理器中移除
	delete(manager.databases, sessionID)
	return nil
}

// CloseAllDatabases 关闭并删除所有数据库
func (manager *MessageDBManager) CloseAllDatabases() error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for sessionID, db := range manager.databases {
		// 关闭数据库
		if err := db.db.Close(); err != nil {
			log.Printf("关闭数据库失败 [%s]: %v", sessionID, err)
		}

		// 删除数据库文件
		if err := os.Remove(db.dbPath); err != nil {
			log.Printf("删除数据库文件失败 [%s]: %v", sessionID, err)
		} else {
			log.Printf("已删除会话 %s 的数据库文件", sessionID)
		}
	}

	// 清空管理器
	manager.databases = make(map[string]*MessageDB)

	// 删除整个数据库目录
	if err := os.RemoveAll(DBDirectory); err != nil {
		log.Printf("删除数据库目录失败: %v", err)
	} else {
		log.Printf("已清理所有消息数据库")
	}

	return nil
}

// AddMessage 添加消息到数据库
func (db *MessageDB) AddMessage(record session.MessageRecord) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(MessageBucket))
		if bucket == nil {
			return fmt.Errorf("bucket不存在")
		}

		// 使用时间戳作为key
		key := fmt.Sprintf("%d", record.Timestamp)

		// 序列化消息记录
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("序列化消息失败: %v", err)
		}

		return bucket.Put([]byte(key), data)
	})
}

// GetMessages 获取所有消息
func (db *MessageDB) GetMessages(limit, offset int) ([]session.MessageRecord, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var messages []session.MessageRecord

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(MessageBucket))
		if bucket == nil {
			return nil // bucket不存在，返回空列表
		}

		cursor := bucket.Cursor()
		count := 0
		skipped := 0

		// 按时间戳顺序遍历
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			// 跳过offset条记录
			if skipped < offset {
				skipped++
				continue
			}

			// 达到limit限制
			if limit > 0 && count >= limit {
				break
			}

			var record session.MessageRecord
			if err := json.Unmarshal(value, &record); err != nil {
				log.Printf("反序列化消息失败: %v", err)
				continue
			}

			messages = append(messages, record)
			count++
		}

		return nil
	})

	return messages, err
}

// ClearMessages 清空所有消息
func (db *MessageDB) ClearMessages() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.db.Update(func(tx *bbolt.Tx) error {
		// 删除现有bucket
		if err := tx.DeleteBucket([]byte(MessageBucket)); err != nil {
			// 如果bucket不存在，忽略错误
			if err != bbolt.ErrBucketNotFound {
				return err
			}
		}

		// 重新创建bucket
		_, err := tx.CreateBucket([]byte(MessageBucket))
		return err
	})
}

// GetMessageCount 获取消息总数
func (db *MessageDB) GetMessageCount() (int, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	count := 0

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(MessageBucket))
		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()
		for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
			count++
		}

		return nil
	})

	return count, err
}

// StoreMessageToDB 存储消息到数据库
func StoreMessageToDB(sessionID string, record session.MessageRecord) error {
	if GlobalMessageDBManager == nil {
		return fmt.Errorf("消息数据库管理器未初始化")
	}

	db, err := GlobalMessageDBManager.GetOrCreateMessageDB(sessionID)
	if err != nil {
		return fmt.Errorf("获取数据库失败: %v", err)
	}

	return db.AddMessage(record)
}

// GetMessagesFromDB 从数据库获取消息
func GetMessagesFromDB(sessionID string, limit, offset int) ([]session.MessageRecord, error) {
	if GlobalMessageDBManager == nil {
		return nil, fmt.Errorf("消息数据库管理器未初始化")
	}

	db, err := GlobalMessageDBManager.GetOrCreateMessageDB(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取数据库失败: %v", err)
	}

	return db.GetMessages(limit, offset)
}

// ClearSessionMessagesInDB 清空会话的所有消息
func ClearSessionMessagesInDB(sessionID string) error {
	if GlobalMessageDBManager == nil {
		return fmt.Errorf("消息数据库管理器未初始化")
	}

	db, err := GlobalMessageDBManager.GetOrCreateMessageDB(sessionID)
	if err != nil {
		return fmt.Errorf("获取数据库失败: %v", err)
	}

	return db.ClearMessages()
}
