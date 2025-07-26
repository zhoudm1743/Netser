# Netser 事件驱动系统

这个事件驱动系统为Netser应用提供了一种松耦合的组件通信机制。通过发布-订阅模式，不同的模块可以在不直接依赖的情况下进行通信。

## 主要特性

- 基于发布-订阅模式的事件总线
- 支持异步事件处理
- 提供标准化的事件类型定义
- 与Wails应用程序深度集成
- 支持前端和后端之间的事件通信

## 使用方法

### 在Go后端使用

```go
// 发布事件
app.EventBus.Publish(event.EventDataUpdated, event.NewEvent(
    event.EventDataUpdated,
    map[string]interface{}{"id": 123, "status": "completed"},
    "dataService",
))

// 订阅事件
app.EventBus.Subscribe(event.EventDataUpdated, func(data interface{}) {
    if evt, ok := data.(*event.Event); ok {
        // 处理事件数据
        fmt.Printf("Received event: %s from %s\n", evt.Name, evt.Source)
    }
})
```

### 在前端JavaScript中使用

```javascript
// 订阅事件
window.go.main.App.SubscribeToEvent("data:updated", (eventData) => {
  console.log("Data updated:", eventData);
  // 更新UI等操作
});

// 发布事件
window.go.main.App.PublishEvent("ui:changed", { componentId: "userList", action: "filter" }, "userComponent");

// 取消订阅
window.go.main.App.UnsubscribeFromEvent("data:updated", callbackReference);
```

## 标准事件类型

系统预定义了一系列标准事件类型，包括：

- `app:startup` - 应用程序启动
- `app:shutdown` - 应用程序关闭
- `user:login` - 用户登录
- `user:logout` - 用户登出
- `data:loaded` - 数据加载完成
- `data:saved` - 数据保存完成
- `data:updated` - 数据更新
- `ui:changed` - UI变更
- `ui:refresh` - UI刷新

## 最佳实践

1. 为特定功能域创建专用的事件名称（如 `user:`, `data:`, `ui:` 等前缀）
2. 在事件数据中包含足够的上下文信息
3. 对于关键操作，考虑使用同步处理方式
4. 避免过度使用事件系统导致"事件风暴"
5. 为调试目的，考虑实现事件日志记录 