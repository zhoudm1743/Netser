# WebSocket 交互流程设计

## 1. 连接建立流程

### 前端连接WebSocket
```
1. 前端通过现有API获取WebSocket端口
   GET /api/ws_info → {port: 1743, status: "available"}

2. 建立WebSocket连接
   ws://localhost:1743/ws

3. 连接成功后发送认证消息
   {
     "type": "auth",
     "id": "msg_xxx",
     "timestamp": 1640995200000,
     "data": {
       "clientId": "frontend_xxx",
       "version": "1.0",
       "sessionId": null
     }
   }

4. 服务端响应认证结果
   {
     "type": "auth_response", 
     "id": "msg_xxx",
     "code": 0,
     "message": "认证成功",
     "timestamp": 1640995200000,
     "data": {
       "clientId": "frontend_xxx",
       "serverVersion": "1.0"
     }
   }
```

## 2. 会话订阅流程

### 订阅特定会话的消息
```
1. 前端发送订阅消息
   {
     "type": "subscribe",
     "id": "msg_yyy", 
     "timestamp": 1640995200000,
     "data": {
       "sessionId": "tcp_1754711950767119800"
     }
   }

2. 服务端响应订阅结果
   {
     "type": "auth_response",
     "id": "msg_yyy",
     "code": 0,
     "message": "订阅成功",
     "timestamp": 1640995200000,
     "data": {
       "sessionId": "tcp_1754711950767119800",
       "status": "subscribed"
     }
   }
```

## 3. 实时消息推送

### TCP消息推送
```
服务端主动推送TCP消息（无需前端请求）
{
  "type": "tcp_message",
  "timestamp": 1640995200000,
  "data": {
    "sessionId": "tcp_1754711950767119800",
    "direction": "receive",
    "content": "Hello World",
    "isHex": false,
    "byteLength": 11,
    "timestamp": 1640995200000
  }
}
```

### 会话状态变化推送
```
服务端主动推送状态变化
{
  "type": "session_status",
  "timestamp": 1640995200000,
  "data": {
    "sessionId": "tcp_1754711950767119800", 
    "status": "connected",
    "timestamp": 1640995200000
  }
}
```

## 4. 心跳机制

### 心跳检测（每30秒）
```
1. 前端发送心跳
   {
     "type": "ping",
     "id": "msg_ping_xxx",
     "timestamp": 1640995200000
   }

2. 服务端响应心跳
   {
     "type": "pong", 
     "id": "msg_ping_xxx",
     "timestamp": 1640995200000
   }
```

## 5. 错误处理

### 错误消息格式
```
{
  "type": "error",
  "timestamp": 1640995200000,
  "data": {
    "code": 4003,
    "message": "会话不存在",
    "details": "Session tcp_xxx not found",
    "sessionId": "tcp_xxx"
  }
}
```

## 6. 连接管理策略

### 连接复用策略
- 同一个前端页面只维护一个WebSocket连接
- 通过订阅/取消订阅机制管理多个会话
- 连接断开时自动重连（最多重试5次）

### 服务端连接池
- 维护所有活跃的WebSocket连接
- 按sessionId路由消息到对应的连接
- 支持一个会话被多个客户端订阅

## 7. 端口管理

### 端口分配策略
```
优先使用: 1743
如果占用: 1744, 1745, 1746... (递增查找)
最大尝试: 10个端口
失败处理: 返回错误，建议检查防火墙
```

### 端口状态API
```
GET /api/ws_info
Response: {
  "port": 1743,
  "status": "available|occupied|error", 
  "message": "WebSocket服务运行正常"
}
``` 