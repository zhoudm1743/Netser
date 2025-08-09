<script setup>
import { BaseRequest, BaseResponse } from './dto/base'
import { ref, onMounted, watch } from 'vue'
import { Greet } from '../wailsjs/go/main/App'
import SessionList from './components/session/SessionList.vue'
import SessionView from './components/session/SessionView.vue'
import { useSessionStore } from './stores/session'
import { BottomLeft, FullScreen, Close } from '@element-plus/icons-vue'

const version = ref('')
const sessionStore = useSessionStore()
const isDarkMode = ref(false) // 主题模式状态，默认白天模式

// 应用启动时初始化
onMounted(async () => {
  // 从本地存储读取主题设置
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    isDarkMode.value = true
  }
  
  // 先初始化WebSocket连接
  await sessionStore.initWebSocket()
  // 然后加载会话列表
  await sessionStore.loadSessions()
  // 获取版本信息
  getVersion()
})

// 监听主题变化
watch(isDarkMode, (newVal) => {
  const body = document.body
  if (newVal) {
    body.classList.add('dark-mode')
    body.classList.remove('light-mode')
    localStorage.setItem('theme', 'dark')
  } else {
    body.classList.add('light-mode')
    body.classList.remove('dark-mode')
    localStorage.setItem('theme', 'light')
  }
}, { immediate: true })

const getVersion = () => {
  const request = new BaseRequest('get_version', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
    version.value = baseResponse.data
  }).catch(error => {
    console.error('获取版本失败:', error)
  })
}

const minimize = () => {
  const request = new BaseRequest('minimize', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
  })
}

const maximize = () => {
  const request = new BaseRequest('maximize', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
  })
}

const close = () => {
  const request = new BaseRequest('close', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
  })
}


</script>

<template>
  <div class="container">
    <header>
      <h3>Netser {{ version }}</h3>
      <div class="test-area">
       
      </div>
      <div class="control">
        <div class="theme-switch">
          <el-switch
            v-model="isDarkMode"
            active-color="#2c3e50"
            inactive-color="#f39c12"
            size="large"
          />
        </div>
        <el-button-group>
        <el-button @click="minimize">
          <el-icon><BottomLeft /></el-icon>
        </el-button>
        <el-button @click="maximize">
          <el-icon><FullScreen /></el-icon>
        </el-button>
        <el-button @click="close">
          <el-icon><Close /></el-icon>
        </el-button>
      </el-button-group>
      </div>
    </header>
    <div class="content">
      <div class="content-left">
        <el-card style="max-width: 300px;height: 100%;">
          <SessionList />
        </el-card>
      </div>
      <div class="content-right">
        <el-card style="width: 100%;height: 100%;">
          <SessionView />
        </el-card>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  .content {
    padding: 10px;
    width: 100%;
    height: calc(100% - 70px);
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 10px;
    .content-left {
      width: 260px;
      height: 100%;
      // background-color: #fff;
    }
    .content-right {
      width: calc(100% - 270px);
      height: 100%;
      // background-color: #fff;
    }
  }
}
header {
  width: 100%;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  h3 {
    font-size: 24px;
    font-weight: 600;
    color: var(--text-primary);
    margin-left: 20px;
    transition: color 0.3s ease;
  }
}

.control {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-right: 20px;
  height: 100%;
}

.theme-switch {
  display: flex;
  align-items: center;
  height: 32px;
}
</style>
