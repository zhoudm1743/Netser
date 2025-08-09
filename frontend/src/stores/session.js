import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { Greet } from '../../wailsjs/go/main/App'
import { BaseRequest, BaseResponse } from '../dto/base'
import { ElMessage } from 'element-plus'
import { wsManager, WSConnectionState } from '../utils/websocket-manager.js'

export const useSessionStore = defineStore('session', () => {
  // 状态
  const sessions = ref([])
  const selectedSession = ref(null)
  const isLoading = ref(false)
  
  // WebSocket状态
  const wsConnectionState = ref(WSConnectionState.DISCONNECTED)
  const messages = ref([]) // 当前选中会话的消息列表

  // 计算属性
  const selectedSessionId = computed(() => selectedSession.value?.sessionId || '')
  const sessionCount = computed(() => sessions.value.length)
  
  // 根据ID获取会话
  const getSessionById = computed(() => {
    return (sessionId) => sessions.value.find(s => s.sessionId === sessionId)
  })

  // Actions
  
  // 加载所有会话
  const loadSessions = async () => {
    isLoading.value = true
    try {
      const request = new BaseRequest('get_sessions', {})
      const response = await Greet(request.toJson())
      const baseResponse = BaseResponse.fromJson(response)
      
      if (baseResponse.code === 0 && baseResponse.data && baseResponse.data.sessions) {
        sessions.value = baseResponse.data.sessions
      } else {
        sessions.value = []
      }
    } catch (error) {
      console.error('加载会话失败:', error)
      ElMessage.error('加载会话失败')
      sessions.value = []
    } finally {
      isLoading.value = false
    }
  }

  // 创建会话
  const createSession = async (sessionData) => {
    try {
      const request = new BaseRequest('create_session', sessionData)
      const response = await Greet(request.toJson())
      const baseResponse = BaseResponse.fromJson(response)
      
      if (baseResponse.code === 0) {
        await loadSessions() // 重新加载会话列表
        
        // 选中新创建的会话
        if (baseResponse.data && baseResponse.data.sessionId) {
          selectSession(baseResponse.data)
        }
        
        ElMessage.success('会话创建成功')
        return baseResponse.data
      } else {
        ElMessage.error(baseResponse.message || '创建会话失败')
        return null
      }
    } catch (error) {
      console.error('创建会话失败:', error)
      ElMessage.error('创建会话失败')
      return null
    }
  }

  // 删除会话
  const removeSession = async (sessionId) => {
    try {
      const request = new BaseRequest('remove_session', { sessionId })
      const response = await Greet(request.toJson())
      const baseResponse = BaseResponse.fromJson(response)
      
      if (baseResponse.code === 0) {
        // 如果删除的是当前选中的会话，清空选中状态
        if (selectedSession.value && selectedSession.value.sessionId === sessionId) {
          selectedSession.value = null
        }
        
        await loadSessions() // 重新加载会话列表
        ElMessage.success('会话删除成功')
        return true
      } else {
        ElMessage.error(baseResponse.message || '删除会话失败')
        return false
      }
    } catch (error) {
      console.error('删除会话失败:', error)
      ElMessage.error('删除会话失败')
      return false
    }
  }

  // 选择会话
  const selectSession = async (session) => {
    console.log('=== Store 选择会话 ===')
    console.log('选择的会话:', session)
    
    // 取消订阅旧会话
    if (selectedSession.value) {
      await unsubscribeFromSession(selectedSession.value.sessionId)
    }
    
    selectedSession.value = session
    
    // 清空消息列表
    clearMessages()
    
    // 订阅新会话并加载历史消息
    if (session) {
      await loadSessionMessages(session.sessionId)
      await subscribeToSession(session.sessionId)
    }
  }

  // 清空选择的会话
  const clearSelectedSession = () => {
    selectedSession.value = null
  }

  // 更新会话状态
  const updateSessionStatus = (sessionId, status) => {
    console.log('=== updateSessionStatus ===')
    console.log('sessionId:', sessionId)
    console.log('新状态:', status)
    
    const session = sessions.value.find(s => s.sessionId === sessionId)
    console.log('找到的会话:', session)
    
    if (session) {
      const oldStatus = session.status
      session.status = status
      console.log('会话状态更新:', oldStatus, '->', status)
      
      // 如果是当前选中的会话，也更新选中状态
      if (selectedSession.value && selectedSession.value.sessionId === sessionId) {
        const oldSelectedStatus = selectedSession.value.status
        selectedSession.value.status = status
        console.log('选中会话状态更新:', oldSelectedStatus, '->', status)
      }
    } else {
      console.log('警告：未找到要更新的会话')
    }
  }

  // 更新整个会话对象
  const updateSession = (updatedSession) => {
    console.log('=== updateSession ===')
    console.log('更新的会话数据:', updatedSession)
    
    const index = sessions.value.findIndex(s => s.sessionId === updatedSession.sessionId)
    console.log('找到会话索引:', index)
    
    if (index !== -1) {
      const oldSession = sessions.value[index]
      sessions.value[index] = updatedSession
      console.log('会话更新完成:', oldSession.status, '->', updatedSession.status)
      
      // 如果是当前选中的会话，也更新选中状态
      if (selectedSession.value && selectedSession.value.sessionId === updatedSession.sessionId) {
        const oldSelected = selectedSession.value
        selectedSession.value = updatedSession
        console.log('选中会话更新完成:', oldSelected.status, '->', updatedSession.status)
      }
    } else {
      console.log('警告：未找到要更新的会话，sessionId:', updatedSession.sessionId)
    }
  }

  // 连接会话
  const connectSession = async (session) => {
    if (!session) return false
    
    try {
      const request = new BaseRequest('connect', {
        sessionId: session.sessionId,
        sessionData: session
      })
      
      const response = await Greet(request.toJson())
      const baseResponse = BaseResponse.fromJson(response)
      
      if (baseResponse.code === 0) {
        console.log('=== Store connectSession 成功 ===')
        console.log('原始会话:', session)
        console.log('后端响应数据:', baseResponse.data)
        
        // 如果后端返回了更新的会话数据，使用它
        if (baseResponse.data && baseResponse.data.Info) {
          console.log('使用后端返回的会话数据更新')
          // 后端返回的数据结构中，实际的会话信息在 Info 字段中
          updateSession(baseResponse.data.Info)
        } else {
          // 否则手动更新状态
          const newStatus = session.type === 'tcpClient' ? 'connected' : 'listening'
          console.log('手动更新状态为:', newStatus)
          updateSessionStatus(session.sessionId, newStatus)
        }
        
        console.log('更新后的selectedSession:', selectedSession.value)
        console.log('更新后的sessions:', sessions.value)
        
        ElMessage.success('连接成功')
        return true
      } else {
        ElMessage.error(baseResponse.message || '连接失败')
        return false
      }
    } catch (error) {
      console.error('连接失败:', error)
      ElMessage.error('连接失败')
      return false
    }
  }

  // 断开连接
  const disconnectSession = async (session) => {
    if (!session) return false
    
    try {
      const request = new BaseRequest('disconnect', {
        sessionId: session.sessionId
      })
      
      const response = await Greet(request.toJson())
      const baseResponse = BaseResponse.fromJson(response)
      
      if (baseResponse.code === 0) {
        // 如果后端返回了更新的会话数据，使用它
        if (baseResponse.data && baseResponse.data.Info) {
          updateSession(baseResponse.data.Info)
        } else {
          // 否则手动更新状态
          updateSessionStatus(session.sessionId, 'disconnected')
        }
        
        ElMessage.success('断开连接成功')
        return true
      } else {
        ElMessage.error(baseResponse.message || '断开连接失败')
        return false
      }
    } catch (error) {
      console.error('断开连接失败:', error)
      ElMessage.error('断开连接失败')
      return false
    }
  }

  // ========== WebSocket相关方法 ==========

  // 初始化WebSocket连接
  const initWebSocket = async () => {
    console.log('初始化WebSocket连接')
    
    // 设置WebSocket事件监听器
    wsManager.on('state_changed', ({ newState }) => {
      wsConnectionState.value = newState
      console.log('WebSocket状态变化:', newState)
    })

    wsManager.on('tcp_message', (data) => {
      handleTCPMessage(data)
    })

    wsManager.on('session_status', (data) => {
      handleSessionStatusChange(data)
    })

    wsManager.on('connected', () => {
      console.log('WebSocket已连接，重新订阅当前会话')
      if (selectedSession.value) {
        subscribeToSession(selectedSession.value.sessionId)
      }
    })

    wsManager.on('disconnected', () => {
      console.log('WebSocket已断开')
    })

    // 尝试连接
    await wsManager.connect()
  }

  // 处理TCP消息推送
  const handleTCPMessage = (data) => {
    console.log('处理TCP消息推送:', data)
    
    // 只处理当前选中会话的消息
    if (selectedSession.value && data.sessionId === selectedSession.value.sessionId) {
      messages.value.push({
        type: data.direction,
        data: data.content,
        timestamp: data.timestamp,
        isHex: data.isHex,
        byteLength: data.byteLength
      })
      
      console.log('添加新消息到界面')
    }
  }

  // 处理会话状态变化推送
  const handleSessionStatusChange = (data) => {
    console.log('处理会话状态变化:', data)
    updateSessionStatus(data.sessionId, data.status)
  }

  // 订阅会话消息
  const subscribeToSession = async (sessionId) => {
    if (wsManager.isConnected()) {
      await wsManager.subscribeSession(sessionId)
    } else {
      console.log('WebSocket未连接，无法订阅会话')
    }
  }

  // 取消订阅会话消息
  const unsubscribeFromSession = async (sessionId) => {
    if (wsManager.isConnected()) {
      await wsManager.unsubscribeSession(sessionId)
    }
  }

  // 加载会话消息（初始加载）
  const loadSessionMessages = async (sessionId) => {
    try {
      const request = new BaseRequest('get_session_messages', {
        sessionId: sessionId,
        limit: 100,
        offset: 0
      })

      const response = await Greet(request.toJson())
      const baseResponse = BaseResponse.fromJson(response)
      
      if (baseResponse.code !== 0) return

      const messageData = baseResponse.data
      console.log('loadSessionMessages 收到数据:', messageData)
      
      if (messageData && messageData.records) {
        messages.value = messageData.records.map(record => ({
          type: record.direction,
          data: record.data,
          timestamp: record.timestamp,
          isHex: record.isHex || false,
          byteLength: record.byteLength || record.data.length
        }))
        
        console.log(`加载了 ${messages.value.length} 条历史消息`, messages.value)
      } else {
        console.log('没有消息数据或records为空')
        messages.value = []
      }
    } catch (error) {
      console.error('加载会话消息失败:', error)
      messages.value = []
    }
  }

  // 清空消息列表
  const clearMessages = () => {
    messages.value = []
  }

  // 断开WebSocket连接
  const disconnectWebSocket = () => {
    wsManager.disconnect()
  }

  return {
    // 状态
    sessions,
    selectedSession,
    isLoading,
    wsConnectionState,
    messages,
    
    // 计算属性
    selectedSessionId,
    sessionCount,
    getSessionById,
    
    // 方法
    loadSessions,
    createSession,
    removeSession,
    selectSession,
    clearSelectedSession,
    updateSessionStatus,
    updateSession,
    connectSession,
    disconnectSession,
    
    // WebSocket方法
    initWebSocket,
    subscribeToSession,
    unsubscribeFromSession,
    loadSessionMessages,
    clearMessages,
    disconnectWebSocket
  }
}) 