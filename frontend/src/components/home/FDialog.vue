<template>
  <el-dialog
    v-model="visible"
    :title="title"
    :width="width"
    :before-close="handleClose"
    :close-on-click-modal="false"
  >
    <!-- 内容区域 -->
    <slot />

    <!-- 底部 -->
    <template #footer>
      <div class="footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="onConfirm">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: Boolean,
  title: String,
  width: {
    type: String,
    default: '500px'
  }
})

const emit = defineEmits(['update:modelValue', 'confirm', 'close'])

const visible = computed({
  get: () => props.modelValue,
  set: val => emit('update:modelValue', val)
})

const onConfirm = () => {
  emit('confirm')
}

const handleClose = () => {
  emit('close')
  visible.value = false
}
</script>

<style scoped>
.footer {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 40px;
}
</style>
