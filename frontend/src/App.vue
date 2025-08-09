<script setup>
import { BaseRequest, BaseResponse } from './dto/base'
import { ref, onMounted } from 'vue'
import { Greet } from '../wailsjs/go/main/App'
import SessionList from './components/session/SessionList.vue'
import SessionView from './components/session/SessionView.vue'
import { useSessionStore } from './stores/session'

const version = ref('')
const sessionStore = useSessionStore()

// 应用启动时初始化
onMounted(async () => {
  // 先初始化WebSocket连接
  await sessionStore.initWebSocket()
  // 然后加载会话列表
  await sessionStore.loadSessions()
})

const getVersion = () => {
  console.log('=== 测试getVersion ===')
  const request = new BaseRequest('get_version', {})
  console.log('发送请求:', request.toJson())
  Greet(request.toJson()).then(response => {
    console.log('收到响应:', response)
    const baseResponse = BaseResponse.fromJson(response)
    console.log('解析后响应:', baseResponse)
    version.value = baseResponse.data
    console.log('版本设置为:', version.value)
  }).catch(error => {
    console.error('请求失败:', error)
  })
}

const minimize = () => {
  const request = new BaseRequest('minimize', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
    console.log(response)
  })
}

const maximize = () => {
  const request = new BaseRequest('maximize', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
    console.log(response)
  })
}

const close = () => {
  const request = new BaseRequest('close', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
    console.log(response)
  })
}


</script>

<template>
  <div class="container">
    <header>
      <h3>Netser {{ version }}</h3>
      <div class="test-area">
        <el-button @click="getVersion" size="small">测试连接</el-button>
      </div>
      <div class="control">
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
    height: calc(100% - 60px);
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
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  h3 {
    font-size: 24px;
    font-weight: 600;
    color: #333;
    margin-left: 20px;
  }
}
</style>
