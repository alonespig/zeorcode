<template>
  <div v-if="error" class="agent-error" role="alert">
    <div class="agent-error__mark">!</div>
    <h2>AI 助手加载失败</h2>
    <p>当前对话中有无法展示的数据，请重新加载后再试。</p>
    <button type="button" @click="$emit('retry')">重新加载</button>
  </div>
  <slot v-else />
</template>

<script setup>
import { onErrorCaptured, shallowRef } from 'vue'

defineEmits(['retry'])

const error = shallowRef(null)

onErrorCaptured((capturedError, instance, info) => {
  error.value = capturedError
  console.error('[AgentErrorBoundary]', info, capturedError, instance)
  return false
})
</script>

<style scoped>
.agent-error {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  color: #475569;
  background: #f8fafc;
  text-align: center;
}

.agent-error__mark {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  margin-bottom: 14px;
  border: 1px solid #fecaca;
  border-radius: 50%;
  color: #dc2626;
  background: #fef2f2;
  font-size: 22px;
  font-weight: 700;
}

.agent-error h2 {
  margin: 0;
  color: #1e293b;
  font-size: 18px;
}

.agent-error p {
  margin: 8px 0 18px;
  color: #64748b;
  font-size: 13px;
}

.agent-error button {
  height: 36px;
  padding: 0 18px;
  border-radius: 5px;
  color: #fff;
  background: #367bf5;
  font-size: 13px;
}
</style>
