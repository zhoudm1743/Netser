// WebSocket消息类型枚举
export const MessageType = {
  // 基础消息类型
  AUTH: 'auth',                    // 认证消息
  AUTH_RESPONSE: 'auth_response',  // 认证响应
  SUBSCRIBE: 'subscribe',          // 订阅会话消息
  UNSUBSCRIBE: 'unsubscribe',      // 取消订阅
  PING: 'ping',                    // 心跳检测
  PONG: 'pong',                    // 心跳响应
  
  // 业务消息类型
  TCP_MESSAGE: 'tcp_message',      // TCP消息推送
  SESSION_STATUS: 'session_status', // 会话状态变化
  SYSTEM_NOTIFY: 'system_notify',   // 系统通知
  ERROR: 'error'                   // 错误消息
}

// 状态码
export const StatusCode = {
  SUCCESS: 0,            // 成功
  INVALID_MESSAGE: 4001, // 消息格式错误
  AUTH_FAILED: 4002,     // 认证失败
  SESSION_NOT_FOUND: 4003, // 会话不存在
  INTERNAL_ERROR: 5001,  // 内部错误
  SUBSCRIBE_FAILED: 4004 // 订阅失败
}

// WebSocket基础消息类
export class WSBaseMessage {
  constructor(type, data = null, id = null) {
    this.type = type
    this.id = id
    this.timestamp = Date.now()
    this.data = data
  }

  toJSON() {
    return JSON.stringify(this)
  }

  static fromJSON(jsonStr) {
    try {
      const obj = JSON.parse(jsonStr)
      const msg = new WSBaseMessage(obj.type, obj.data, obj.id)
      msg.timestamp = obj.timestamp
      return msg
    } catch (error) {
      throw new Error(`解析WebSocket消息失败: ${error.message}`)
    }
  }
}

// WebSocket响应消息类
export class WSResponseMessage {
  constructor(type, id, code, message, data = null) {
    this.type = type
    this.id = id
    this.code = code
    this.message = message
    this.timestamp = Date.now()
    this.data = data
  }

  toJSON() {
    return JSON.stringify(this)
  }

  static fromJSON(jsonStr) {
    try {
      const obj = JSON.parse(jsonStr)
      const msg = new WSResponseMessage(obj.type, obj.id, obj.code, obj.message, obj.data)
      msg.timestamp = obj.timestamp
      return msg
    } catch (error) {
      throw new Error(`解析WebSocket响应消息失败: ${error.message}`)
    }
  }
}

// 认证数据
export class AuthData {
  constructor(clientId, version, sessionId = null) {
    this.clientId = clientId
    this.version = version
    this.sessionId = sessionId
  }
}

// 订阅数据
export class SubscribeData {
  constructor(sessionId) {
    this.sessionId = sessionId
  }
}

// 取消订阅数据
export class UnsubscribeData {
  constructor(sessionId) {
    this.sessionId = sessionId
  }
}

// TCP消息数据
export class TCPMessageData {
  constructor(sessionId, direction, content, isHex, byteLength, timestamp) {
    this.sessionId = sessionId
    this.direction = direction
    this.content = content
    this.isHex = isHex
    this.byteLength = byteLength
    this.timestamp = timestamp
  }
}

// 会话状态数据
export class SessionStatusData {
  constructor(sessionId, status, timestamp) {
    this.sessionId = sessionId
    this.status = status
    this.timestamp = timestamp
  }
}

// 系统通知数据
export class SystemNotifyData {
  constructor(level, title, message, sessionId = null) {
    this.level = level
    this.title = title
    this.message = message
    this.sessionId = sessionId
  }
}

// 错误数据
export class ErrorData {
  constructor(code, message, details = null, sessionId = null) {
    this.code = code
    this.message = message
    this.details = details
    this.sessionId = sessionId
  }
}

// 获取状态码对应的消息
export function getStatusMessage(code) {
  switch (code) {
    case StatusCode.SUCCESS:
      return '成功'
    case StatusCode.INVALID_MESSAGE:
      return '消息格式错误'
    case StatusCode.AUTH_FAILED:
      return '认证失败'
    case StatusCode.SESSION_NOT_FOUND:
      return '会话不存在'
    case StatusCode.INTERNAL_ERROR:
      return '内部错误'
    case StatusCode.SUBSCRIBE_FAILED:
      return '订阅失败'
    default:
      return '未知错误'
  }
}

// 生成唯一消息ID
export function generateMessageId() {
  return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

// 消息工厂函数
export const WSMessageFactory = {
  // 创建认证消息
  createAuth(clientId, version, sessionId = null) {
    const authData = new AuthData(clientId, version, sessionId)
    return new WSBaseMessage(MessageType.AUTH, authData, generateMessageId())
  },

  // 创建订阅消息
  createSubscribe(sessionId) {
    const subscribeData = new SubscribeData(sessionId)
    return new WSBaseMessage(MessageType.SUBSCRIBE, subscribeData, generateMessageId())
  },

  // 创建取消订阅消息
  createUnsubscribe(sessionId) {
    const unsubscribeData = new UnsubscribeData(sessionId)
    return new WSBaseMessage(MessageType.UNSUBSCRIBE, unsubscribeData, generateMessageId())
  },

  // 创建心跳消息
  createPing() {
    return new WSBaseMessage(MessageType.PING, null, generateMessageId())
  },

  // 创建心跳响应
  createPong() {
    return new WSBaseMessage(MessageType.PONG, null)
  }
} 