<script setup>
import { BaseRequest, BaseResponse } from './dto/base'
import { ref } from 'vue'
import { Greet } from '../wailsjs/go/main/App'
import SessionList from './components/session/SessionList.vue'
import SessionView from './components/session/SessionView.vue'

const version = ref('')
const selectedSession = ref(null)

const getVersion = () => {
  const request = new BaseRequest('get_version', {})
  Greet(request.toJson()).then(response => {
    const baseResponse = BaseResponse.fromJson(response)
    version.value = baseResponse.data
    console.log(response)
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

const handleSessionSelect = (session) => {
  console.log('选中会话:', session)
  selectedSession.value = session
}

const handleSessionConnect = (session) => {
  console.log('连接会话:', session)
  // 在这里可以更新会话状态
  if (selectedSession.value && selectedSession.value.sessionId === session.sessionId) {
    selectedSession.value.status = 'connecting'
  }
}

const handleSessionDisconnect = (session) => {
  console.log('断开会话:', session)
  // 在这里可以更新会话状态
  if (selectedSession.value && selectedSession.value.sessionId === session.sessionId) {
    selectedSession.value.status = 'disconnected'
  }
}

const handleSessionDelete = (session) => {
  console.log('删除会话:', session)
  // 如果删除的是当前选中的会话，清空选中状态
  if (selectedSession.value && selectedSession.value.sessionId === session.sessionId) {
    selectedSession.value = null
  }
}
</script>

<template>
  <div class="container">
    <header>
      <h3>Netser</h3>
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
          <session-list 
            @select="handleSessionSelect"
            @connect="handleSessionConnect"
            @disconnect="handleSessionDisconnect"
            @delete="handleSessionDelete"
          />
        </el-card>
      </div>
      <div class="content-right">
        <el-card style="width: 100%;height: 100%;">
          <session-view :selected-session="selectedSession" />
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
