package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"github.com/zhoudm1743/Netser/core/session"
)

// App struct
type App struct {
	ctx            context.Context
	EventBus       *event.EventBus         // 事件总线
	logger         *event.EventLogger      // 事件日志记录器
	SessionService *session.SessionService // 会话服务
}

// NewApp creates a new App application struct
func NewApp() *App {
	// 创建一个新的App实例
	app := &App{
		EventBus: event.NewEventBus(),
	}

	// 初始化事件日志器
	logger, err := event.NewEventLogger(app.EventBus, "events.log")
	if err != nil {
		fmt.Printf("无法初始化事件日志器: %v\n", err)
	} else {
		app.logger = logger
		// 只记录应用程序生命周期相关事件
		logger.SetInterests([]string{
			event.EventAppStartup,
			event.EventAppShutdown,
			event.EventUIChanged,
			session.EventSessionCreated,
			session.EventSessionDestroyed,
			session.EventSessionExpired,
		})
	}

	// 初始化会话服务
	var storage session.SessionStorage
	fileStorage, err := session.NewFileStorage("./data/sessions", app.EventBus)
	if err != nil {
		fmt.Printf("无法初始化会话存储: %v\n", err)
		// 降级为内存存储
		storage = session.NewMemoryStorage()
	} else {
		storage = fileStorage
	}

	app.SessionService = session.NewSessionService(30*time.Minute, storage, app.EventBus)

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 发布应用程序启动事件
	a.EventBus.Publish(event.EventAppStartup, event.NewEvent(
		event.EventAppStartup,
		nil,
		"app",
	))
}

// shutdown 在应用程序关闭时调用
func (a *App) shutdown(ctx context.Context) {
	// 关闭会话服务
	if a.SessionService != nil {
		a.SessionService.Shutdown()
	}

	// 发布应用程序关闭事件
	a.EventBus.Publish(event.EventAppShutdown, event.NewEvent(
		event.EventAppShutdown,
		nil,
		"app",
	))

	// 关闭日志记录器
	if a.logger != nil {
		a.logger.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	// 发布用户交互事件
	a.EventBus.Publish(event.EventUIChanged, event.NewEvent(
		event.EventUIChanged,
		map[string]interface{}{"name": name},
		"greet",
	))

	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// SubscribeToEvent 允许前端订阅事件
func (a *App) SubscribeToEvent(eventName string, callback func(interface{})) {
	a.EventBus.Subscribe(eventName, callback)
}

// UnsubscribeFromEvent 允许前端取消订阅事件
func (a *App) UnsubscribeFromEvent(eventName string, callback func(interface{})) {
	a.EventBus.Unsubscribe(eventName, callback)
}

// PublishEvent 允许前端发布事件
func (a *App) PublishEvent(eventName string, data interface{}, source string) {
	a.EventBus.Publish(eventName, event.NewEvent(
		eventName,
		data,
		source,
	))
}

// CreateSession 创建用户会话
func (a *App) CreateSession(userID string, userData map[string]interface{}) map[string]interface{} {
	session, err := a.SessionService.CreateUserSession(userID, userData)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":   true,
		"sessionId": session.ID,
		"expiresAt": session.ExpiresAt,
	}
}

// GetSession 获取会话信息
func (a *App) GetSession(sessionID string) map[string]interface{} {
	session, err := a.SessionService.GetSession(sessionID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success":   true,
		"sessionId": session.ID,
		"userId":    session.UserID,
		"expiresAt": session.ExpiresAt,
		"data":      session.Data,
	}
}

// RefreshSession 刷新会话
func (a *App) RefreshSession(sessionID string) map[string]interface{} {
	err := a.SessionService.RefreshSession(sessionID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	session, _ := a.SessionService.GetSession(sessionID)

	return map[string]interface{}{
		"success":   true,
		"expiresAt": session.ExpiresAt,
	}
}

// DestroySession 销毁会话
func (a *App) DestroySession(sessionID string) map[string]interface{} {
	err := a.SessionService.DestroySession(sessionID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// SetSessionData 设置会话数据
func (a *App) SetSessionData(sessionID, key string, value interface{}) map[string]interface{} {
	err := a.SessionService.SetSessionData(sessionID, key, value)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// GetSessionData 获取会话数据
func (a *App) GetSessionData(sessionID, key string) map[string]interface{} {
	value, err := a.SessionService.GetSessionData(sessionID, key)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"value":   value,
	}
}

// GetActiveSessionCount 获取活跃会话数
func (a *App) GetActiveSessionCount() int {
	return a.SessionService.GetActiveSessionCount()
}
