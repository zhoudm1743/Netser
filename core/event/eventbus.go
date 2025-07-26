package event

import (
	"sync"
)

// EventHandler 事件处理函数类型
type EventHandler func(data interface{})

// EventBus 事件总线结构
type EventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewEventBus 创建一个新的事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventName string, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if _, exists := eb.handlers[eventName]; !exists {
		eb.handlers[eventName] = []EventHandler{}
	}
	eb.handlers[eventName] = append(eb.handlers[eventName], handler)
}

// Unsubscribe 取消订阅事件
func (eb *EventBus) Unsubscribe(eventName string, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if handlers, exists := eb.handlers[eventName]; exists {
		for i, h := range handlers {
			if &h == &handler {
				eb.handlers[eventName] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(eventName string, data interface{}) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	if handlers, exists := eb.handlers[eventName]; exists {
		for _, handler := range handlers {
			go handler(data)
		}
	}
}

// HasSubscribers 检查是否有订阅者
func (eb *EventBus) HasSubscribers(eventName string) bool {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	handlers, exists := eb.handlers[eventName]
	return exists && len(handlers) > 0
}
