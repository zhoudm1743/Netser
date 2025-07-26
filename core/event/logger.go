package event

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// EventLogger 事件日志记录器
type EventLogger struct {
	eventBus  *EventBus
	logger    *log.Logger
	enabled   bool
	logFile   *os.File
	mutex     sync.Mutex
	interests map[string]bool // 需要记录的事件类型
}

// NewEventLogger 创建一个新的事件日志记录器
func NewEventLogger(eventBus *EventBus, logFilePath string) (*EventLogger, error) {
	var logFile *os.File
	var logger *log.Logger
	var err error

	if logFilePath != "" {
		// 创建或打开日志文件
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("无法打开日志文件: %v", err)
		}
		logger = log.New(logFile, "EVENT: ", log.Ldate|log.Ltime)
	} else {
		// 默认使用标准输出
		logger = log.New(os.Stdout, "EVENT: ", log.Ldate|log.Ltime)
	}

	el := &EventLogger{
		eventBus:  eventBus,
		logger:    logger,
		enabled:   true,
		logFile:   logFile,
		interests: make(map[string]bool),
	}

	// 默认记录所有事件
	eventBus.Subscribe("*", el.logEvent)

	return el, nil
}

// logEvent 记录事件
func (el *EventLogger) logEvent(data interface{}) {
	el.mutex.Lock()
	defer el.mutex.Unlock()

	if !el.enabled {
		return
	}

	if evt, ok := data.(*Event); ok {
		// 检查是否需要记录此类型的事件
		if len(el.interests) > 0 {
			if _, exists := el.interests[evt.Name]; !exists {
				return
			}
		}

		el.logger.Printf("[%s] 来源: %s, 数据: %v",
			evt.Name, evt.Source, evt.Data)
	}
}

// Enable 启用日志记录
func (el *EventLogger) Enable() {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.enabled = true
}

// Disable 禁用日志记录
func (el *EventLogger) Disable() {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.enabled = false
}

// SetInterests 设置需要记录的事件类型
func (el *EventLogger) SetInterests(eventNames []string) {
	el.mutex.Lock()
	defer el.mutex.Unlock()

	// 清除现有的关注列表
	el.interests = make(map[string]bool)

	// 添加新的关注事件
	for _, name := range eventNames {
		el.interests[name] = true
	}
}

// Close 关闭日志记录器
func (el *EventLogger) Close() error {
	el.mutex.Lock()
	defer el.mutex.Unlock()

	// 取消事件订阅
	el.eventBus.Unsubscribe("*", el.logEvent)

	// 如果使用文件，则关闭
	if el.logFile != nil {
		return el.logFile.Close()
	}

	return nil
}
