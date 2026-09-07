<template>
  <el-card shadow="never">
    <template #header>
      <div>
        <div class="text-base font-semibold text-gray-800">系统通知</div>
        <div class="mt-0.5 text-xs text-gray-400">发送后立即推送给全体用户</div>
      </div>
    </template>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="64px" style="max-width: 560px">
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" maxlength="60" show-word-limit placeholder="通知标题" />
      </el-form-item>
      <el-form-item label="内容" prop="content">
        <el-input v-model="form.content" type="textarea" :rows="4" maxlength="500" show-word-limit
          placeholder="通知内容" />
      </el-form-item>
      <el-form-item label="链接">
        <el-input v-model="form.link" placeholder="选填，站内相对路径，如 /contest/5" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="sending" @click="send">发送给全体用户</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { broadcastNotification } from '@/api/notification'

const formRef = ref(null)
const sending = ref(false)
const form = ref({ title: '', content: '', link: '' })
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }],
}

const send = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  try {
    await ElMessageBox.confirm('确定向全体用户发送这条系统通知？', '确认', { type: 'warning' })
  } catch {
    return
  }
  sending.value = true
  try {
    await broadcastNotification(form.value)
    ElMessage.success('已发送')
    form.value = { title: '', content: '', link: '' }
  } catch (err) {
    console.error(err)
  } finally {
    sending.value = false
  }
}
</script>
