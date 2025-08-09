<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑会话' : '添加会话'"
    width="500px"
    :before-close="handleClose"
    append-to-body
  >
    <el-form ref="formRef" :model="form" label-width="100px" :rules="rules">
      <el-form-item label="会话名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入会话名称" />
      </el-form-item>

      <el-form-item label="连接类型" prop="type">
        <el-select v-model="form.type" placeholder="请选择连接类型" style="width: 100%">
          <el-option label="TCP服务端" value="tcpServer" />
          <el-option label="TCP客户端" value="tcpClient" />
          <el-option label="串口" value="serial" />
        </el-select>
      </el-form-item>

      <!-- TCP服务端配置 -->
      <template v-if="form.type === 'tcpServer'">
        <el-form-item label="监听端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
      </template>

      <!-- TCP客户端配置 -->
      <template v-if="form.type === 'tcpClient'">
        <el-form-item label="主机地址" prop="host">
          <el-input v-model="form.host" placeholder="请输入主机地址" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
      </template>

      <!-- 串口配置 -->
      <template v-if="form.type === 'serial'">
        <el-form-item label="端口" prop="serialPort">
          <el-select v-model="form.serialPort" placeholder="请选择串口" style="width: 100%">
            <el-option 
              v-for="port in serialPorts" 
              :key="port.path" 
              :label="port.path" 
              :value="port.path" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="波特率" prop="baudRate">
          <el-select v-model="form.baudRate" placeholder="请选择波特率" style="width: 100%">
            <el-option label="4800" value="4800" />
            <el-option label="9600" value="9600" />
            <el-option label="19200" value="19200" />
            <el-option label="38400" value="38400" />
            <el-option label="57600" value="57600" />
            <el-option label="115200" value="115200" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据位" prop="dataBits">
          <el-select v-model="form.dataBits" placeholder="请选择数据位" style="width: 100%">
            <el-option label="5" value="5" />
            <el-option label="6" value="6" />
            <el-option label="7" value="7" />
            <el-option label="8" value="8" />
          </el-select>
        </el-form-item>
        <el-form-item label="停止位" prop="stopBits">
          <el-select v-model="form.stopBits" placeholder="请选择停止位" style="width: 100%">
            <el-option label="1" value="1" />
            <el-option label="1.5" value="1.5" />
            <el-option label="2" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="校验位" prop="parity">
          <el-select v-model="form.parity" placeholder="请选择校验位" style="width: 100%">
            <el-option label="无" value="none" />
            <el-option label="奇校验" value="odd" />
            <el-option label="偶校验" value="even" />
          </el-select>
        </el-form-item>
      </template>

      <!-- 通用配置 -->
      <el-form-item label="十六进制模式">
        <el-switch v-model="form.isHex" />
      </el-form-item>

      <el-form-item label="超时时间(秒)">
        <el-input-number v-model="form.timeout" :min="1" :max="60" />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, defineEmits, defineProps, watch } from 'vue'

const props = defineProps({
  visible: Boolean,
  sessionData: Object,
  isEdit: Boolean
})

const emit = defineEmits(['update:visible', 'submit'])

const dialogVisible = ref(false)
const formRef = ref(null)
const serialPorts = ref([]) // 串口列表

// 表单数据
const form = reactive({
  sessionId: '',
  name: '',
  type: 'tcpClient',
  host: '127.0.0.1',
  port: 8080,
  isHex: false,
  timeout: 5,
  // 串口配置
  serialPort: '',
  baudRate: '9600',
  dataBits: '8',
  stopBits: '1',
  parity: 'none'
})

// 表单验证规则
const rules = {
  name: [{ required: true, message: '请输入会话名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择连接类型', trigger: 'change' }],
  host: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口号', trigger: 'blur' }],
  serialPort: [{ required: true, message: '请选择串口', trigger: 'change' }],
  baudRate: [{ required: true, message: '请选择波特率', trigger: 'change' }]
}

// 监听对话框显示状态
watch(() => props.visible, (val) => {
  dialogVisible.value = val
  if (val && props.isEdit && props.sessionData) {
    // 编辑模式，填充表单数据
    Object.keys(props.sessionData).forEach(key => {
      if (key in form) {
        form[key] = props.sessionData[key]
      }
    })
  } else if (val) {
    // 添加模式，重置表单
    resetForm()
  }
})

// 监听对话框内部状态变化
watch(dialogVisible, (val) => {
  emit('update:visible', val)
})

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  resetForm()
}

// 重置表单
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  
  form.sessionId = ''
  form.name = ''
  form.type = 'tcpClient'
  form.host = '127.0.0.1'
  form.port = 8080
  form.isHex = false
  form.timeout = 5
  form.serialPort = ''
  form.baudRate = '9600'
  form.dataBits = '8'
  form.stopBits = '1'
  form.parity = 'none'
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    // 根据连接类型构建不同的表单数据
    const formData = {
      sessionId: form.sessionId,
      name: form.name,
      type: form.type,
      isHex: form.isHex,
      timeout: form.timeout
    }
    
    if (form.type === 'tcpServer') {
      formData.port = form.port
    } else if (form.type === 'tcpClient') {
      formData.host = form.host
      formData.port = form.port
    } else if (form.type === 'serial') {
      formData.serialPort = form.serialPort
      formData.baudRate = form.baudRate
      formData.dataBits = form.dataBits
      formData.stopBits = form.stopBits
      formData.parity = form.parity
    }
    
    emit('submit', formData)
    dialogVisible.value = false
  } catch (error) {
    console.error('表单验证失败', error)
  }
}

// 加载串口列表（实际应用中需要从后端获取）
const loadSerialPorts = async () => {
  // 模拟串口列表，实际应用中应该从后端获取
  serialPorts.value = [
    { path: 'COM1' },
    { path: 'COM2' },
    { path: 'COM3' }
  ]
}

// 组件挂载时加载串口列表
loadSerialPorts()
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
}
</style> 