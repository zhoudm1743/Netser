<script setup>
import { BaseRequest, BaseResponse } from './dto/base'
import { ref } from 'vue'
import { Greet } from '../wailsjs/go/main/App'
const version = ref('')

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
    <template #header>
      <div class="card-header">
        <span>会话列表</span>
      </div>
    </template>
    <p v-for="o in 4" :key="o" class="text item">{{ 'List item ' + o }}</p>
  </el-card>
      </div>
      <div class="content-right">
        <el-card style="width: 100%;height: 100%;">
          <router-view />
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
