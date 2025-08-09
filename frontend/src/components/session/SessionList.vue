<template>
  <div class="session-list">
    <div class="session-header">
      <h3>会话列表</h3>
      <div class="session-actions">
        <el-button type="primary" size="small" @click="handleAddSession">
          <el-icon><Plus /></el-icon> 添加
        </el-button>
      </div>
    </div>
    
    <el-empty v-if="sessions.length === 0" description="暂无会话" />
    
    <el-scrollbar height="calc(100vh - 160px)" v-else>
      <div 
        v-for="session in sessions" 
        :key="session.sessionId" 
        class="session-item"
        :class="{ 'session-active': activeSessionId === session.sessionId }"
        @click="handleSessionSelect(session)"
      >
        <div class="session-info">
          <div class="session-name">
            {{ session.name || `会话 ${session.sessionId.substring(0, 6)}` }}
          </div>
          <div class="session-type">
            <el-tag size="small" :type="getTypeTagType(session.type)">
              {{ getTypeLabel(session.type) }}
            </el-tag>
          </div>
          <div class="session-status">
            <el-tag 
              size="small" 
              :type="session.status === 'connected' ? 'success' : 
                    (session.status === 'connecting' ? 'warning' : 'info')"
            >
              {{ getStatusLabel(session.status) }}
            </el-tag>
          </div>
          <div class="session-details">
            {{ getSessionDetails(session) }}
          </div>
        </div>
        <div class="session-actions">
          <el-dropdown trigger="click" @command="(command) => handleCommand(command, session)">
            <el-button size="small" type="text">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item :disabled="isSessionBusy(session)" command="connect">
                  {{ session.status === 'connected' ? '断开连接' : '连接' }}
                </el-dropdown-item>
                <el-dropdown-item command="edit">编辑</el-dropdown-item>
                <el-dropdown-item command="rename">重命名</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-scrollbar>
    
    <!-- 会话表单对话框 -->
    <session-dialog
      v-model:visible="dialogVisible"
      :session-data="currentSession"
      :is-edit="isEdit"
      @submit="handleDialogSubmit"
    />
    
    <!-- 重命名对话框 -->
    <el-dialog
      v-model="renameDialogVisible"
      title="重命名会话"
      width="400px"
    >
      <el-input v-model="newSessionName" placeholder="请输入新的会话名称" />
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="renameDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleRenameConfirm">确认</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, defineEmits } from 'vue'
import SessionDialog from './SessionDialog.vue'

const emit = defineEmits(['select', 'connect', 'disconnect', 'delete'])

// 会话列表数据
const sessions = ref([
  // 示例数据，实际使用时应从后端获取
  // {
  //   sessionId: '1',
  //   name: 'TCP服务端示例',
  //   type: 'tcpServer',
  //   status: 'disconnected',
  //   port: 8080,
  //   isHex: false
  // }
])

const activeSessionId = ref('')
const dialogVisible = ref(false)
const renameDialogVisible = ref(false)
const isEdit = ref(false)
const currentSession = ref(null)
const newSessionName = ref('')

// 类型标签颜色
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

// 获取会话详情文本
const getSessionDetails = (session) => {
  if (session.type === 'tcpServer') {
    return `端口: ${session.port}`
  } else if (session.type === 'tcpClient') {
    return `${session.host}:${session.port}`
  } else if (session.type === 'serial') {
    return `${session.serialPort} ${session.baudRate}bps`
  }
  return ''
}

// 判断会话是否处于繁忙状态
const isSessionBusy = (session) => {
  return session.status === 'connecting'
}

// 选择会话
const handleSessionSelect = (session) => {
  activeSessionId.value = session.sessionId
  emit('select', session)
}

// 添加会话
const handleAddSession = () => {
  isEdit.value = false
  currentSession.value = null
  dialogVisible.value = true
}

// 处理下拉菜单命令
const handleCommand = (command, session) => {
  switch (command) {
    case 'connect':
      if (session.status === 'connected' || session.status === 'listening') {
        emit('disconnect', session)
      } else {
        emit('connect', session)
      }
      break
    case 'edit':
      isEdit.value = true
      currentSession.value = { ...session }
      dialogVisible.value = true
      break
    case 'rename':
      currentSession.value = { ...session }
      newSessionName.value = session.name || ''
      renameDialogVisible.value = true
      break
    case 'delete':
      handleDeleteSession(session)
      break
  }
}

// 处理表单对话框提交
const handleDialogSubmit = (formData) => {
  if (isEdit.value) {
    // 编辑现有会话
    const index = sessions.value.findIndex(s => s.sessionId === formData.sessionId)
    if (index !== -1) {
      sessions.value[index] = { ...formData }
    }
  } else {
    // 添加新会话
    const newSession = {
      ...formData,
      sessionId: `session_${Date.now()}`,
      status: 'disconnected'
    }
    sessions.value.push(newSession)
    activeSessionId.value = newSession.sessionId
    emit('select', newSession)
  }
}

// 确认重命名
const handleRenameConfirm = () => {
  if (currentSession.value && newSessionName.value) {
    const index = sessions.value.findIndex(s => s.sessionId === currentSession.value.sessionId)
    if (index !== -1) {
      sessions.value[index].name = newSessionName.value
    }
  }
  renameDialogVisible.value = false
}

// 删除会话
const handleDeleteSession = (session) => {
  // 确认删除
  ElMessageBox.confirm(
    `确定要删除会话 "${session.name || session.sessionId}" 吗？`,
    '删除会话',
    {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    emit('delete', session)
    const index = sessions.value.findIndex(s => s.sessionId === session.sessionId)
    if (index !== -1) {
      sessions.value.splice(index, 1)
    }
    if (activeSessionId.value === session.sessionId) {
      activeSessionId.value = sessions.value.length > 0 ? sessions.value[0].sessionId : ''
    }
  }).catch(() => {
    // 取消删除
  })
}

// 提供方法给外部调用
defineExpose({
  updateSessionStatus: (sessionId, status) => {
    const index = sessions.value.findIndex(s => s.sessionId === sessionId)
    if (index !== -1) {
      sessions.value[index].status = status
    }
  },
  updateSessions: (newSessions) => {
    sessions.value = newSessions
  },
  addSession: (session) => {
    sessions.value.push(session)
  }
})
</script>

<style scoped>
.session-list {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 10px;
  margin-bottom: 10px;
}

.session-header h3 {
  margin: 0;
  font-size: 16px;
}

.session-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 5px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.session-item:hover {
  background-color: #f5f7fa;
}

.session-active {
  background-color: #ecf5ff;
}

.session-info {
  flex: 1;
}

.session-name {
  font-weight: bold;
  margin-bottom: 5px;
}

.session-type, .session-status {
  margin-bottom: 5px;
}

.session-details {
  color: #909399;
  font-size: 12px;
}

.session-actions {
  display: flex;
  align-items: center;
}
</style> 