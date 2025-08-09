import { ElMessage } from 'element-plus'
import { 
  MessageType, 
  StatusCode, 
  WSMessageFactory, 
  WSBaseMessage,
  WSResponseMessage,
  generateMessageId 
} from '../dto/websocket.js'
import { BaseRequest } from '../dto/base.js'
import { Greet } from '../../wailsjs/go/main/App.js'

// WebSocket连接状态
export const WSConnectionState = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting', 
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting',
  FAILED: 'failed'
}

// WebSocket管理器类
export class WebSocketManager {
  constructor() {
    this.ws = null
    this.wsUrl = null
    this.state = WSConnectionState.DISCONNECTED
    this.clientId = `frontend_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    
    // 重连配置
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectInterval = 1000 // 1秒
    this.reconnectTimer = null
    
    // 心跳配置
    this.heartbeatInterval = 30000 // 30秒
    this.heartbeatTimer = null
    this.lastPongTime = 0
    
    // 消息处理
    this.messageHandlers = new Map()
    this.pendingRequests = new Map() // 待响应的请求
    
    // 事件监听器
    this.eventListeners = new Map()
    
    // 当前订阅的会话
    this.subscribedSessions = new Set()
    
    // 初始化消息处理器
    this.initMessageHandlers()
  }

  // 初始化消息处理器
  initMessageHandlers() {
    this.messageHandlers.set(MessageType.AUTH_RESPONSE, this.handleAuthResponse.bind(this))
    this.messageHandlers.set(MessageType.SUBSCRIBE, this.handleSubscribeResponse.bind(this))
    this.messageHandlers.set(MessageType.UNSUBSCRIBE, this.handleUnsubscribeResponse.bind(this))
    this.messageHandlers.set(MessageType.PONG, this.handlePong.bind(this))
    this.messageHandlers.set(MessageType.TCP_MESSAGE, this.handleTCPMessage.bind(this))
    this.messageHandlers.set(MessageType.SESSION_STATUS, this.handleSessionStatus.bind(this))
    this.messageHandlers.set(MessageType.SYSTEM_NOTIFY, this.handleSystemNotify.bind(this))
    this.messageHandlers.set(MessageType.ERROR, this.handleError.bind(this))
  }

  // 获取WebSocket端口信息
  async getWebSocketInfo() {
    try {
      const request = new BaseRequest('get_ws_info', {})
      const response = await Greet(request.toJson())
      const baseResponse = JSON.parse(response)
      
      if (baseResponse.code === 0 && baseResponse.data) {
        return baseResponse.data
      } else {
        throw new Error(baseResponse.message || 'Failed to get WebSocket info')
      }
    } catch (error) {
      console.error('获取WebSocket信息失败:', error)
      throw error
    }
  }

  // 连接WebSocket
  async connect() {
    console.log('connect() 方法被调用，当前状态:', this.state)
    
    if (this.state === WSConnectionState.CONNECTING || this.state === WSConnectionState.CONNECTED) {
      console.log('WebSocket已经连接或正在连接中')
      return true
    }

    try {
      this.setState(WSConnectionState.CONNECTING)
      console.log('状态已设置为 CONNECTING')
      
      // 获取WebSocket端口信息
      console.log('正在获取WebSocket端口信息...')
      const wsInfo = await this.getWebSocketInfo()
      console.log('获取到WebSocket信息:', wsInfo)
      this.wsUrl = `ws://localhost:${wsInfo.port}/ws`
      
      console.log(`连接WebSocket: ${this.wsUrl}`)
      
      // 建立WebSocket连接
      this.ws = new WebSocket(this.wsUrl)
      console.log('WebSocket对象已创建')
      
      // 设置事件处理器
      this.ws.onopen = this.onOpen.bind(this)
      this.ws.onmessage = this.onMessage.bind(this)
      this.ws.onclose = this.onClose.bind(this)
      this.ws.onerror = this.onError.bind(this)
      console.log('WebSocket事件处理器已设置')
      
      return true
    } catch (error) {
      console.error('WebSocket连接失败:', error)
      this.setState(WSConnectionState.FAILED)
      ElMessage.error(`WebSocket连接失败: ${error.message}`)
      return false
    }
  }

  // 断开连接
  disconnect() {
    console.log('主动断开WebSocket连接')
    this.stopReconnect()
    this.stopHeartbeat()
    
    if (this.ws) {
      this.ws.close(1000, 'User disconnected')
      this.ws = null
    }
    
    this.setState(WSConnectionState.DISCONNECTED)
    this.subscribedSessions.clear()
  }

  // 发送消息
  sendMessage(message) {
    if (this.state !== WSConnectionState.CONNECTED) {
      console.error('WebSocket未连接，无法发送消息')
      return false
    }

    try {
      const jsonData = message.toJSON()
      this.ws.send(jsonData)
      console.log('发送WebSocket消息:', message.type, message.id)
      return true
    } catch (error) {
      console.error('发送WebSocket消息失败:', error)
      return false
    }
  }

  // 发送带响应的请求
  async sendRequest(message, timeout = 10000) {
    return new Promise((resolve, reject) => {
      console.log('sendRequest: 准备发送消息:', message.id, message.type)
      
      if (!this.sendMessage(message)) {
        console.error('sendRequest: 发送消息失败')
        reject(new Error('Failed to send message'))
        return
      }

      console.log('sendRequest: 消息已发送，等待响应...')

      // 设置请求超时
      const timeoutId = setTimeout(() => {
        console.error('sendRequest: 请求超时，消息ID:', message.id)
        this.pendingRequests.delete(message.id)
        reject(new Error('Request timeout'))
      }, timeout)

      // 保存待响应的请求
      this.pendingRequests.set(message.id, { resolve, reject, timeoutId })
      console.log('sendRequest: 已保存待响应请求，当前数量:', this.pendingRequests.size)
    })
  }

  // 订阅会话
  async subscribeSession(sessionId) {
    if (this.subscribedSessions.has(sessionId)) {
      console.log(`会话 ${sessionId} 已经订阅`)
      return true
    }

    try {
      const message = WSMessageFactory.createSubscribe(sessionId)
      const response = await this.sendRequest(message)
      
      if (response.code === StatusCode.SUCCESS) {
        this.subscribedSessions.add(sessionId)
        console.log(`订阅会话成功: ${sessionId}`)
        this.emit('session_subscribed', { sessionId })
        return true
      } else {
        console.error(`订阅会话失败: ${response.message}`)
        ElMessage.error(`订阅失败: ${response.message}`)
        return false
      }
    } catch (error) {
      console.error('订阅会话失败:', error)
      ElMessage.error(`订阅失败: ${error.message}`)
      return false
    }
  }

  // 取消订阅会话
  async unsubscribeSession(sessionId) {
    if (!this.subscribedSessions.has(sessionId)) {
      console.log(`会话 ${sessionId} 未订阅`)
      return true
    }

    try {
      const message = WSMessageFactory.createUnsubscribe(sessionId)
      const response = await this.sendRequest(message)
      
      if (response.code === StatusCode.SUCCESS) {
        this.subscribedSessions.delete(sessionId)
        console.log(`取消订阅会话成功: ${sessionId}`)
        this.emit('session_unsubscribed', { sessionId })
        return true
      } else {
        console.error(`取消订阅会话失败: ${response.message}`)
        return false
      }
    } catch (error) {
      console.error('取消订阅会话失败:', error)
      return false
    }
  }

  // WebSocket连接打开
  async onOpen(event) {
    console.log('WebSocket连接已建立')
    this.setState(WSConnectionState.CONNECTED)
    this.reconnectAttempts = 0
    
    // 发送认证消息
    try {
      console.log('准备发送认证消息，客户端ID:', this.clientId)
      const authMessage = WSMessageFactory.createAuth(this.clientId, '1.0')
      console.log('创建认证消息:', authMessage)
      console.log('认证消息JSON:', authMessage.toJSON())
      
      const response = await this.sendRequest(authMessage)
      console.log('收到认证响应:', response)
      
      if (response.code === StatusCode.SUCCESS) {
        console.log('WebSocket认证成功')
        this.startHeartbeat()
        this.emit('connected')
        ElMessage.success('WebSocket连接成功')
      } else {
        console.error('WebSocket认证失败:', response.message)
        this.ws.close()
      }
    } catch (error) {
      console.error('WebSocket认证失败:', error)
      this.ws.close()
    }
  }

  // WebSocket接收消息
  onMessage(event) {
    try {
      console.log('收到WebSocket原始消息:', event.data)
      
      // 尝试解析为ResponseMessage或BaseMessage
      let message
      try {
        message = WSResponseMessage.fromJSON(event.data)
        console.log('解析为ResponseMessage:', message.type, message.code)
      } catch (e) {
        message = WSBaseMessage.fromJSON(event.data)
        console.log('解析为BaseMessage:', message.type)
      }
      
      // 处理响应消息
      if (message.id && this.pendingRequests.has(message.id)) {
        console.log('找到对应的待响应请求:', message.id)
        const { resolve, timeoutId } = this.pendingRequests.get(message.id)
        clearTimeout(timeoutId)
        this.pendingRequests.delete(message.id)
        console.log('响应处理完成，剩余待响应请求:', this.pendingRequests.size)
        resolve(message)
        return
      } else if (message.id) {
        console.warn('收到未匹配的响应消息:', message.id, '当前待响应请求:', Array.from(this.pendingRequests.keys()))
      }
      
      // 处理推送消息
      const handler = this.messageHandlers.get(message.type)
      if (handler) {
        handler(message)
      } else {
        console.warn('未知的消息类型:', message.type)
      }
    } catch (error) {
      console.error('处理WebSocket消息失败:', error)
    }
  }

  // WebSocket连接关闭
  onClose(event) {
    console.log('WebSocket连接已关闭:', event.code, event.reason)
    this.stopHeartbeat()
    this.ws = null
    
    // 清理待响应的请求
    for (const [id, { reject, timeoutId }] of this.pendingRequests) {
      clearTimeout(timeoutId)
      reject(new Error('Connection closed'))
    }
    this.pendingRequests.clear()
    
    // 临时禁用重连，专注于调试第一次连接
    this.setState(WSConnectionState.DISCONNECTED)
    console.log('WebSocket连接关闭，暂时不重连以便调试')
    
    this.emit('disconnected', { code: event.code, reason: event.reason })
  }

  // WebSocket连接错误
  onError(event) {
    console.error('WebSocket连接错误:', event)
    this.emit('error', event)
  }

  // 开始重连
  startReconnect() {
    if (this.reconnectTimer || this.reconnectAttempts >= this.maxReconnectAttempts) {
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectInterval * Math.pow(2, this.reconnectAttempts - 1) // 指数退避
    
    console.log(`第 ${this.reconnectAttempts} 次重连尝试，${delay}ms 后开始`)
    
    this.reconnectTimer = setTimeout(async () => {
      this.reconnectTimer = null
      const success = await this.connect()
      
      if (!success && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.startReconnect()
      } else if (!success) {
        console.error('WebSocket重连失败，已达到最大重试次数')
        this.setState(WSConnectionState.FAILED)
        ElMessage.error('WebSocket连接失败，请刷新页面重试')
      }
    }, delay)
  }

  // 停止重连
  stopReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.reconnectAttempts = 0
  }

  // 开始心跳
  startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      const pingMessage = WSMessageFactory.createPing()
      this.sendMessage(pingMessage)
    }, this.heartbeatInterval)
  }

  // 停止心跳
  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  // 设置连接状态
  setState(newState) {
    const oldState = this.state
    this.state = newState
    console.log(`WebSocket状态变化: ${oldState} -> ${newState}`)
    this.emit('state_changed', { oldState, newState })
  }

  // 事件发射器
  emit(event, data) {
    const listeners = this.eventListeners.get(event) || []
    listeners.forEach(listener => {
      try {
        listener(data)
      } catch (error) {
        console.error(`事件监听器错误 [${event}]:`, error)
      }
    })
  }

  // 添加事件监听器
  on(event, listener) {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, [])
    }
    this.eventListeners.get(event).push(listener)
  }

  // 移除事件监听器
  off(event, listener) {
    const listeners = this.eventListeners.get(event)
    if (listeners) {
      const index = listeners.indexOf(listener)
      if (index > -1) {
        listeners.splice(index, 1)
      }
    }
  }

  // ========== 消息处理器 ==========

  handleAuthResponse(message) {
    console.log('认证响应:', message)
  }

  handleSubscribeResponse(message) {
    console.log('订阅响应:', message)
  }

  handleUnsubscribeResponse(message) {
    console.log('取消订阅响应:', message)
  }

  handlePong(message) {
    this.lastPongTime = Date.now()
    console.log('收到心跳响应')
  }

  handleTCPMessage(message) {
    console.log('收到TCP消息推送:', message.data)
    this.emit('tcp_message', message.data)
  }

  handleSessionStatus(message) {
    console.log('收到会话状态变化:', message.data)
    this.emit('session_status', message.data)
  }

  handleSystemNotify(message) {
    console.log('收到系统通知:', message.data)
    
    const { level, title, message: msg } = message.data
    if (level === 'error') {
      ElMessage.error(`${title}: ${msg}`)
    } else if (level === 'warning') {
      ElMessage.warning(`${title}: ${msg}`)
    } else {
      ElMessage.info(`${title}: ${msg}`)
    }
    
    this.emit('system_notify', message.data)
  }

  handleError(message) {
    console.error('收到错误消息:', message.data)
    const { message: errorMsg } = message.data
    ElMessage.error(`WebSocket错误: ${errorMsg}`)
    this.emit('websocket_error', message.data)
  }

  // ========== 公共方法 ==========

  // 获取连接状态
  getState() {
    return this.state
  }

  // 是否已连接
  isConnected() {
    return this.state === WSConnectionState.CONNECTED
  }

  // 获取订阅的会话列表
  getSubscribedSessions() {
    return Array.from(this.subscribedSessions)
  }
}

// 全局WebSocket管理器实例
export const wsManager = new WebSocketManager() 