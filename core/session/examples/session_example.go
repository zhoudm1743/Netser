package main

import (
	"fmt"
	"time"

	"github.com/zhoudm1743/Netser/core/event"
	"github.com/zhoudm1743/Netser/core/session"
)

func main() {
	// 创建事件总线
	eventBus := event.NewEventBus()

	// 创建事件日志记录器
	logger, err := event.NewEventLogger(eventBus, "session_events.log")
	if err != nil {
		fmt.Printf("无法初始化事件日志器: %v\n", err)
	} else {
		logger.SetInterests([]string{
			session.EventSessionCreated,
			session.EventSessionDestroyed,
			session.EventSessionExpired,
			session.EventSessionRefreshed,
		})
	}

	// 订阅会话创建事件
	eventBus.Subscribe(session.EventSessionCreated, func(data interface{}) {
		if evt, ok := data.(*event.Event); ok {
			sessionID := evt.Data.(map[string]interface{})["sessionId"].(string)
			userID := evt.Data.(map[string]interface{})["userId"].(string)
			fmt.Printf("会话已创建: %s (用户: %s)\n", sessionID, userID)
		}
	})

	// 订阅会话销毁事件
	eventBus.Subscribe(session.EventSessionDestroyed, func(data interface{}) {
		if evt, ok := data.(*event.Event); ok {
			sessionID := evt.Data.(map[string]interface{})["sessionId"].(string)
			fmt.Printf("会话已销毁: %s\n", sessionID)
		}
	})

	// 创建内存存储
	storage := session.NewMemoryStorage()

	// 创建会话服务 (使用短超时以便测试过期)
	sessionService := session.NewSessionService(10*time.Second, storage, eventBus)

	// 创建用户会话
	userSession, err := sessionService.CreateUserSession("user123", map[string]interface{}{
		"role": "admin",
		"name": "测试管理员",
	})
	if err != nil {
		fmt.Printf("创建会话失败: %v\n", err)
		return
	}

	fmt.Printf("会话创建成功: %s (过期时间: %v)\n", userSession.ID, userSession.ExpiresAt)

	// 获取会话数据
	retrievedSession, err := sessionService.GetSession(userSession.ID)
	if err != nil {
		fmt.Printf("获取会话失败: %v\n", err)
	} else {
		role := retrievedSession.Data["role"].(string)
		name := retrievedSession.Data["name"].(string)
		fmt.Printf("会话数据: 角色=%s, 名称=%s\n", role, name)
	}

	// 设置更多会话数据
	err = sessionService.SetSessionData(userSession.ID, "lastActivity", time.Now().Format(time.RFC3339))
	if err != nil {
		fmt.Printf("设置会话数据失败: %v\n", err)
	} else {
		fmt.Println("成功设置会话数据")
	}

	// 刷新会话
	err = sessionService.RefreshSession(userSession.ID)
	if err != nil {
		fmt.Printf("刷新会话失败: %v\n", err)
	} else {
		refreshedSession, _ := sessionService.GetSession(userSession.ID)
		fmt.Printf("会话已刷新，新过期时间: %v\n", refreshedSession.ExpiresAt)
	}

	// 等待5秒
	fmt.Println("等待5秒...")
	time.Sleep(5 * time.Second)

	// 获取会话计数
	count := sessionService.GetActiveSessionCount()
	fmt.Printf("活跃会话数: %d\n", count)

	// 销毁会话
	fmt.Println("正在销毁会话...")
	err = sessionService.DestroySession(userSession.ID)
	if err != nil {
		fmt.Printf("销毁会话失败: %v\n", err)
	}

	// 尝试获取已销毁的会话
	_, err = sessionService.GetSession(userSession.ID)
	if err != nil {
		fmt.Printf("预期错误: %v\n", err)
	}

	// 创建另一个测试会话并等待它过期
	expireSession, _ := sessionService.CreateUserSession("expireTest", map[string]interface{}{
		"testData": "将过期的测试会话",
	})
	fmt.Printf("创建测试过期会话: %s\n", expireSession.ID)

	// 等待12秒使会话过期
	fmt.Println("等待12秒让会话过期...")
	time.Sleep(12 * time.Second)

	// 尝试获取已过期的会话
	_, err = sessionService.GetSession(expireSession.ID)
	if err != nil {
		fmt.Printf("预期的过期错误: %v\n", err)
	}

	// 关闭会话服务
	sessionService.Shutdown()

	// 等待1秒确保所有事件处理完毕
	time.Sleep(1 * time.Second)
	logger.Close()

	fmt.Println("会话示例运行完成")
}
