<template>
  <div class="session-view">
    <!-- 没有选中会话时的空状态 -->
    <div v-if="!selectedSession" class="empty-state">
      <el-empty description="请选择一个会话开始通信" />
    </div>

    <!-- 选中会话时的主视图 -->
    <div v-else class="session-content">
      <!-- 会话信息头部 -->
      <div class="session-header">
        <div class="session-info">
          <h3>{{ selectedSession.name || `会话 ${selectedSession.sessionId.substring(0, 8)}` }}</h3>
          <div class="session-details">
            <el-tag :type="getTypeTagType(selectedSession.type)" size="small">
              {{ getTypeLabel(selectedSession.type) }}
            </el-tag>
            <el-tag 
              :type="selectedSession.status === 'connected' || selectedSession.status === 'listening' ? 'success' : 
                    (selectedSession.status === 'connecting' ? 'warning' : 'info')" 
              size="small"
            >
              {{ getStatusLabel(selectedSession.status) }}
            </el-tag>
            <span class="connection-info">{{ getConnectionInfo(selectedSession) }}</span>
          </div>
        </div>
        <div class="session-actions">
          <el-button 
            :type="isConnected ? 'danger' : 'primary'" 
            :loading="isConnecting"
            :disabled="isConnecting"
            @click="handleConnectionToggle"
          >
            {{ isConnected ? '断开连接' : (isConnecting ? '连接中...' : '连接') }}
          </el-button>
          <el-button @click="handleClearMessages" :disabled="messages.length === 0">
            <el-icon><Delete /></el-icon> 清空消息
          </el-button>
        </div>
      </div>

      <!-- 消息显示区域 -->
      <div class="message-area">
        <div class="message-header">
          <span>消息记录 ({{ messages.length }})</span>
          <div class="message-controls">
            <el-checkbox v-model="autoScroll">自动滚动</el-checkbox>
            <el-checkbox v-model="showTimestamp">显示时间戳</el-checkbox>
            <el-checkbox v-model="hexDisplay">十六进制显示</el-checkbox>
          </div>
        </div>
        
        <el-scrollbar ref="messageScrollbar" class="message-list" height="400px">
          <div v-if="messages.length === 0" class="no-messages">
            <el-empty description="暂无消息记录" />
          </div>
          <div v-else>
            <div 
              v-for="(message, index) in messages" 
              :key="index"
              :class="['message-item', message.type]"
            >
              <div class="message-header-line">
                <el-tag 
                  :type="message.type === 'send' ? 'success' : 'primary'" 
                  size="small"
                >
                  {{ message.type === 'send' ? '发送' : '接收' }}
                </el-tag>
                <span v-if="showTimestamp" class="timestamp">
                  {{ formatTimestamp(message.timestamp) }}
                </span>
                <span class="message-length">{{ message.data.length }} 字节</span>
              </div>
              <div class="message-content">
                <pre>{{ formatMessageData(message.data) }}</pre>
              </div>
            </div>
          </div>
        </el-scrollbar>
      </div>

      <!-- 数据发送区域 -->
      <div class="send-area">
        <div class="send-header">
          <span>发送数据</span>
          <div class="send-controls">
            <el-checkbox v-model="sendAsHex">十六进制发送</el-checkbox>
            <el-checkbox v-model="addLineBreak">添加换行符</el-checkbox>
          </div>
        </div>
        
        <div class="send-input-area">
          <el-input
            v-model="sendData"
            type="textarea"
            :rows="3"
            :placeholder="sendAsHex ? '请输入十六进制数据，如：48656C6C6F' : '请输入要发送的数据'"
            class="send-input"
            @keydown.ctrl.enter="handleSendData"
          />
          <div class="send-actions">
            <el-button 
              type="primary" 
              @click="handleSendData"
              :disabled="!isConnected || !sendData.trim()"
            >
              发送 (Ctrl+Enter)
            </el-button>
            <el-button @click="handleClearSendData">清空</el-button>
          </div>
        </div>

        <!-- 快速发送模板 -->
        <div class="quick-send">
          <div class="quick-send-header">
            <span>快速发送</span>
            <el-button size="small" type="text" @click="showTemplateDialog = true">
              <el-icon><Plus /></el-icon> 添加模板
            </el-button>
          </div>
          <div class="template-list">
            <el-tag
              v-for="(template, index) in sendTemplates"
              :key="index"
              class="template-tag"
              @click="handleUseTemplate(template)"
              closable
              @close="handleRemoveTemplate(index)"
            >
              {{ template.name }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加模板对话框 -->
    <el-dialog
      v-model="showTemplateDialog"
      title="添加发送模板"
      width="400px"
    >
      <el-form :model="templateForm" label-width="80px">
        <el-form-item label="模板名称">
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="模板数据">
          <el-input
            v-model="templateForm.data"
            type="textarea"
            :rows="3"
            placeholder="请输入模板数据"
          />
        </el-form-item>
        <el-form-item label="十六进制">
          <el-switch v-model="templateForm.isHex" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showTemplateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleAddTemplate">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick, defineProps } from 'vue'
import { BaseRequest, BaseResponse } from '../../dto/base'
import { Greet } from '../../../wailsjs/go/main/App'

const props = defineProps({
  selectedSession: {
    type: Object,
    default: null
  }
})

// 响应式数据
const messages = ref([])
const sendData = ref('')
const autoScroll = ref(true)
const showTimestamp = ref(true)
const hexDisplay = ref(false)
const sendAsHex = ref(false)
const addLineBreak = ref(true)
const messageScrollbar = ref(null)
const showTemplateDialog = ref(false)

// 发送模板
const sendTemplates = ref([
  { name: 'Hello', data: 'Hello World', isHex: false },
  { name: 'Test', data: '48656C6C6F', isHex: true }
])

const templateForm = reactive({
  name: '',
  data: '',
  isHex: false
})

// 计算属性
const isConnected = computed(() => {
  return props.selectedSession && 
    (props.selectedSession.status === 'connected' || props.selectedSession.status === 'listening')
})

const isConnecting = computed(() => {
  return props.selectedSession && props.selectedSession.status === 'connecting'
})

// 获取连接类型标签类型
const getTypeTagType = (type) => {
  switch (type) {
    case 'tcpServer':
      return 'primary'
    case 'tcpClient':
      return 'success'
    case 'serial':
      return 'warning'
    default:
      return 'info'
  }
}

// 获取连接类型显示文本
const getTypeLabel = (type) => {
  switch (type) {
    case 'tcpServer':
      return 'TCP服务端'
    case 'tcpClient':
      return 'TCP客户端'
    case 'serial':
      return '串口'
    default:
      return '未知类型'
  }
}

// 获取状态显示文本
const getStatusLabel = (status) => {
  switch (status) {
    case 'connected':
      return '已连接'
    case 'disconnected':
      return '未连接'
    case 'connecting':
      return '连接中'
    case 'listening':
      return '监听中'
    default:
      return '未知状态'
  }
}

// 获取连接信息
const getConnectionInfo = (session) => {
  if (session.type === 'tcpServer') {
    return `端口: ${session.port}`
  } else if (session.type === 'tcpClient') {
    return `${session.host}:${session.port}`
  } else if (session.type === 'serial') {
    return `${session.serialPort} ${session.baudRate}bps`
  }
  return ''
}

// 格式化时间戳
const formatTimestamp = (timestamp) => {
  return new Date(timestamp).toLocaleTimeString()
}

// 格式化消息数据
const formatMessageData = (data) => {
  if (hexDisplay.value) {
    // 转换为十六进制显示
    return data.split('').map(char => 
      char.charCodeAt(0).toString(16).padStart(2, '0').toUpperCase()
    ).join(' ')
  }
  return data
}

// 处理连接切换
const handleConnectionToggle = async () => {
  if (!props.selectedSession) return

  try {
    const action = isConnected.value ? 'disconnect' : 'connect'
    const request = new BaseRequest(action, {
      sessionId: props.selectedSession.sessionId,
      sessionData: props.selectedSession
    })

    const response = await Greet(request.toJson())
    const baseResponse = BaseResponse.fromJson(response)
    
    if (baseResponse.code === 0) {
      // 连接状态改变，添加系统消息
      addSystemMessage(`${action === 'connect' ? '连接' : '断开连接'}${baseResponse.message ? ': ' + baseResponse.message : '成功'}`)
    } else {
      ElMessage.error(baseResponse.message || '操作失败')
    }
  } catch (error) {
    console.error('连接操作失败:', error)
    ElMessage.error('连接操作失败')
  }
}

// 发送数据
const handleSendData = async () => {
  if (!isConnected.value || !sendData.value.trim()) return

  try {
    let dataToSend = sendData.value

    // 处理十六进制数据
    if (sendAsHex.value) {
      // 移除空格并验证十六进制格式
      const hexData = dataToSend.replace(/\s/g, '')
      if (!/^[0-9A-Fa-f]*$/.test(hexData) || hexData.length % 2 !== 0) {
        ElMessage.error('请输入有效的十六进制数据')
        return
      }
      // 转换为字符串
      dataToSend = hexData.match(/.{2}/g).map(hex => String.fromCharCode(parseInt(hex, 16))).join('')
    }

    // 添加换行符
    if (addLineBreak.value) {
      dataToSend += '\n'
    }

    const request = new BaseRequest('send_data', {
      sessionId: props.selectedSession.sessionId,
      data: dataToSend
    })

    const response = await Greet(request.toJson())
    const baseResponse = BaseResponse.fromJson(response)
    
    if (baseResponse.code === 0) {
      // 添加发送消息记录
      addMessage('send', dataToSend)
      sendData.value = ''
    } else {
      ElMessage.error(baseResponse.message || '发送失败')
    }
  } catch (error) {
    console.error('发送数据失败:', error)
    ElMessage.error('发送数据失败')
  }
}

// 清空发送数据
const handleClearSendData = () => {
  sendData.value = ''
}

// 清空消息
const handleClearMessages = () => {
  messages.value = []
}

// 添加消息
const addMessage = (type, data) => {
  messages.value.push({
    type,
    data,
    timestamp: Date.now()
  })

  // 自动滚动到底部
  if (autoScroll.value) {
    nextTick(() => {
      if (messageScrollbar.value) {
        messageScrollbar.value.setScrollTop(messageScrollbar.value.wrapRef.scrollHeight)
      }
    })
  }
}

// 添加系统消息
const addSystemMessage = (message) => {
  messages.value.push({
    type: 'system',
    data: message,
    timestamp: Date.now()
  })
}

// 使用模板
const handleUseTemplate = (template) => {
  sendData.value = template.data
  sendAsHex.value = template.isHex
}

// 添加模板
const handleAddTemplate = () => {
  if (!templateForm.name.trim() || !templateForm.data.trim()) {
    ElMessage.error('请填写完整的模板信息')
    return
  }

  sendTemplates.value.push({
    name: templateForm.name,
    data: templateForm.data,
    isHex: templateForm.isHex
  })

  // 重置表单
  templateForm.name = ''
  templateForm.data = ''
  templateForm.isHex = false
  showTemplateDialog.value = false

  ElMessage.success('模板添加成功')
}

// 删除模板
const handleRemoveTemplate = (index) => {
  sendTemplates.value.splice(index, 1)
}

// 监听选中会话变化
watch(() => props.selectedSession, (newSession, oldSession) => {
  if (newSession && newSession.sessionId !== oldSession?.sessionId) {
    // 切换会话时清空消息
    messages.value = []
    sendData.value = ''
  }
})

// 模拟接收数据（实际应用中应该通过WebSocket或其他方式接收）
const simulateReceiveData = () => {
  if (isConnected.value && Math.random() > 0.7) {
    const sampleData = ['OK', 'ERROR', 'Data received', '48656C6C6F']
    const randomData = sampleData[Math.floor(Math.random() * sampleData.length)]
    addMessage('receive', randomData)
  }
}

// 定期模拟接收数据（仅用于演示）
setInterval(simulateReceiveData, 5000)
</script>

<style scoped>
.session-view {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.session-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #ebeef5;
}

.session-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
}

.session-details {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.connection-info {
  margin-left: 8px;
}

.session-actions {
  display: flex;
  gap: 8px;
}

.message-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  border-bottom: 1px solid #ebeef5;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
}

.message-controls {
  display: flex;
  gap: 16px;
}

.message-list {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: #fafafa;
}

.no-messages {
  padding: 40px;
  text-align: center;
}

.message-item {
  padding: 8px 12px;
  border-bottom: 1px solid #ebeef5;
}

.message-item:last-child {
  border-bottom: none;
}

.message-item.send {
  background-color: #f0f9ff;
}

.message-item.receive {
  background-color: #f6ffed;
}

.message-item.system {
  background-color: #fffbe6;
}

.message-header-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 12px;
}

.timestamp {
  color: #909399;
}

.message-length {
  color: #909399;
  margin-left: auto;
}

.message-content {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.4;
}

.message-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.send-area {
  padding: 16px;
}

.send-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
}

.send-controls {
  display: flex;
  gap: 16px;
}

.send-input-area {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.send-input {
  flex: 1;
}

.send-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.quick-send {
  border-top: 1px solid #ebeef5;
  padding-top: 16px;
}

.quick-send-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
}

.template-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.template-tag {
  cursor: pointer;
}

.template-tag:hover {
  background-color: #ecf5ff;
}
</style>
