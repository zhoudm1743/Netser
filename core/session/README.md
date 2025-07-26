# Netser 会话管理系统

这个会话管理系统为Netser应用提供了一个完整的用户会话处理解决方案，支持会话创建、管理、存储和事件通知。

## 主要功能

- 会话创建、刷新与销毁
- 会话数据存储与检索
- 会话自动过期与清理
- 多种存储后端支持
- 与事件系统深度集成

## 系统架构

会话管理系统由以下几个主要组件构成：

1. **Session** - 表示单个用户会话的数据结构
2. **SessionManager** - 内存中会话管理器
3. **SessionStorage** - 会话存储接口及实现
   - MemoryStorage - 内存存储实现
   - FileStorage - 文件存储实现
4. **SessionService** - 高级会话服务，整合管理器和存储

## 会话生命周期

1. **创建** - 通过`CreateUserSession`方法创建新会话，分配唯一ID
2. **使用** - 通过ID获取会话，读写会话数据
3. **刷新** - 延长会话有效期
4. **过期/销毁** - 会话到期自动清理或手动销毁

## 会话事件

系统会在关键节点触发以下事件：

- `session:created` - 会话创建时
- `session:destroyed` - 会话被销毁时
- `session:expired` - 会话过期时
- `session:refreshed` - 会话被刷新时
- `session:cleanup` - 清理过期会话时

## 使用示例

### 后端使用示例

```go
// 创建会话
sessionID, _ := app.SessionService.CreateUserSession("user123", map[string]interface{}{
    "role": "admin",
    "name": "管理员",
})

// 获取会话数据
session, _ := app.SessionService.GetSession(sessionID)
role := session.Data["role"].(string)

// 刷新会话
app.SessionService.RefreshSession(sessionID)

// 设置会话数据
app.SessionService.SetSessionData(sessionID, "lastActivity", time.Now())

// 销毁会话
app.SessionService.DestroySession(sessionID)
```

### 前端使用示例

```javascript
// 创建会话
window.go.main.App.CreateSession("user123", {role: "admin", name: "管理员"})
  .then(result => {
    if (result.success) {
      console.log("会话创建成功:", result.sessionId);
      localStorage.setItem("sessionId", result.sessionId);
    }
  });

// 获取会话
const sessionId = localStorage.getItem("sessionId");
window.go.main.App.GetSession(sessionId)
  .then(result => {
    if (result.success) {
      console.log("用户角色:", result.data.role);
    }
  });

// 设置会话数据
window.go.main.App.SetSessionData(sessionId, "lastPage", "dashboard");

// 销毁会话
window.go.main.App.DestroySession(sessionId)
  .then(() => {
    localStorage.removeItem("sessionId");
  });
```

## 会话存储配置

系统默认使用文件存储，保存在`./data/sessions`目录。如果无法创建此目录，将自动降级为内存存储。会话默认有效期为30分钟。

## 安全考虑

- 会话ID使用UUID生成，确保唯一性和安全性
- 系统自动清理过期会话，避免会话泄露
- 会话数据支持任意类型，可存储复杂的用户状态

## 与事件系统集成

会话系统与事件系统深度集成，可以通过订阅相关事件来响应会话状态变化：

```go
app.EventBus.Subscribe(session.EventSessionExpired, func(data interface{}) {
    if evt, ok := data.(*event.Event); ok {
        sessionID := evt.Data.(map[string]interface{})["sessionId"].(string)
        fmt.Printf("会话已过期: %s\n", sessionID)
    }
})
``` 