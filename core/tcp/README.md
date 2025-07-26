# Netser TCP模块

TCP模块提供了一个功能完整的TCP服务端和客户端实现，与事件系统和会话系统深度集成，并支持十六进制模式的数据读写。

## 主要特性

- **TCP服务端**：支持多客户端连接管理
- **TCP客户端**：支持自动重连机制
- **事件驱动**：所有网络事件都通过事件系统通知
- **会话集成**：TCP连接可以与会话系统关联
- **HEX模式**：支持十六进制数据的读写和格式化
- **并发安全**：所有操作都是线程安全的

## 主要组件

### 1. TCP连接 (Connection)

提供对TCP连接的封装，包含读写方法和状态管理。

```go
// 创建新连接
conn := NewConnection(tcpConn, eventBus)

// 设置超时
conn.SetTimeout(readTimeout, writeTimeout)

// 设置HEX模式
conn.SetHexMode(true)

// 读写数据
data, err := conn.Read(1024)
n, err := conn.Write([]byte("Hello"))

// 读写HEX数据
hexStr, err := conn.ReadHex(8)
n, err := conn.WriteHex("48656C6C6F") // "Hello"的十六进制
```

### 2. TCP服务端 (Server)

提供TCP服务端功能，支持多客户端连接管理。

```go
// 创建配置
config := DefaultServerConfig()
config.Port = 8080
config.MaxConnections = 100

// 创建服务端
server := NewServer(config, eventBus, sessionService)

// 注册数据处理器
server.RegisterHandler("echo", func(conn *Connection, data []byte) {
    conn.Write(data) // 回显数据
})

// 启动服务器
err := server.Start()

// 广播数据
count := server.BroadcastData([]byte("系统消息"))
// 或广播HEX数据
count := server.BroadcastHex("48656C6C6F")

// 停止服务器
server.Stop()
```

### 3. TCP客户端 (Client)

提供TCP客户端功能，支持自动重连。

```go
// 创建配置
config := DefaultClientConfig()
config.Host = "example.com"
config.Port = 8080
config.AutoReconnect = true

// 创建客户端
client := NewClient(config, eventBus, sessionService)

// 注册数据处理器
client.RegisterDataHandler(func(data []byte) {
    fmt.Println("收到数据:", BytesToHex(data))
})

// 连接服务器
err := client.Connect()

// 发送数据
n, err := client.Send([]byte("Hello"))
// 或发送HEX数据
n, err := client.SendHex("48656C6C6F")

// 断开连接
client.Disconnect()
```

### 4. 十六进制工具 (HexUtils)

提供十六进制数据处理工具。

```go
// 字节数组与十六进制字符串转换
hexStr := BytesToHex(data)
data, err := HexToBytes("48656C6C6F")

// 格式化十六进制字符串
formatted := FormatHexString("48656C6C6F", 2) // "4865 6C6C 6F"

// 十六进制转储
dump := HexDump(data, 16)

// 带通配符的十六进制模式匹配
pattern := "48 ?? 6C"  // 第2个字节为通配符
indices := FindHexPattern(data, pattern)
```

## 事件集成

TCP模块与事件系统深度集成，会触发以下事件：

### 服务端事件

- `tcp:server_started` - 服务器启动
- `tcp:server_stopped` - 服务器停止
- `tcp:connection_rejected` - 连接被拒绝
- `tcp:client_connected` - 客户端连接

### 客户端事件

- `tcp:client_connecting` - 客户端连接中
- `tcp:client_connected` - 客户端已连接
- `tcp:client_disconnected` - 客户端断开连接
- `tcp:client_reconnecting` - 客户端重连中
- `tcp:client_error` - 客户端错误

### 连接事件

- `tcp:connected` - 连接建立
- `tcp:disconnected` - 连接断开
- `tcp:data_received` - 接收到数据
- `tcp:data_sent` - 发送数据
- `tcp:error` - 连接错误

## 会话集成

TCP模块可以将TCP连接与会话系统关联：

- 服务端为每个客户端连接创建会话
- 客户端连接可以关联到用户会话
- 通过会话可以持久化TCP连接状态

## 使用示例

### 创建TCP服务器

```go
// 创建事件总线和会话服务
eventBus := event.NewEventBus()
sessionService := session.NewSessionService(30*time.Minute, session.NewMemoryStorage(), eventBus)

// 创建服务器配置
config := DefaultServerConfig()
config.Port = 9000

// 创建并启动服务器
server := NewServer(config, eventBus, sessionService)
server.Start()

// 订阅客户端连接事件
eventBus.Subscribe(EventClientConnected, func(data interface{}) {
    if evt, ok := data.(*event.Event); ok {
        fmt.Printf("新客户端连接: %s\n", evt.Data.(map[string]interface{})["remoteAddr"])
    }
})

// 注册数据处理器
server.RegisterHandler("data_handler", func(conn *Connection, data []byte) {
    // 如果开启了HEX模式，显示十六进制数据
    if conn.IsHexMode() {
        fmt.Printf("收到HEX数据: %s\n", BytesToHex(data))
    } else {
        fmt.Printf("收到数据: %s\n", string(data))
    }
    
    // 回复数据
    conn.Write(data)
})
```

### 创建TCP客户端

```go
// 创建事件总线和会话服务
eventBus := event.NewEventBus()
sessionService := session.NewSessionService(30*time.Minute, session.NewMemoryStorage(), eventBus)

// 创建客户端配置
config := DefaultClientConfig()
config.Host = "localhost"
config.Port = 9000
config.AutoReconnect = true

// 创建客户端
client := NewClient(config, eventBus, sessionService)

// 订阅连接事件
eventBus.Subscribe(EventClientConnectedTo, func(data interface{}) {
    if evt, ok := data.(*event.Event); ok {
        fmt.Println("已连接到服务器")
    }
})

// 注册数据处理器
client.RegisterDataHandler(func(data []byte) {
    fmt.Printf("收到响应: %s\n", string(data))
})

// 连接服务器
err := client.Connect()
if err != nil {
    fmt.Printf("连接失败: %v\n", err)
    return
}

// 发送数据
client.SendString("Hello Server!")

// 发送十六进制数据
client.SetHexMode(true)
client.SendHex("48656C6C6F20576F726C6421") // "Hello World!"
```

## 十六进制模式的使用

TCP模块支持十六进制模式，方便调试和处理二进制协议：

```go
// 设置HEX模式
connection.SetHexMode(true)

// 发送十六进制数据
connection.WriteHex("FF00A1B2C3")

// 格式化十六进制数据
hex := "48656C6C6F20576F726C6421" // "Hello World!"
fmt.Println(FormatHexString(hex, 4)) // "48656C6C 6F20576F 726C6421"

// 十六进制转储
data := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}
fmt.Println(HexDump(data, 16))
// 输出:
// 00000000: 48 65 6c 6c 6f                  |Hello|
```

## 安全注意事项

- TCP连接默认没有加密，敏感数据应该使用额外的加密层
- 所有远程数据都应当被视为不可信，进行适当的验证和清理
- 设置适当的连接超时和重连策略，避免资源泄漏 