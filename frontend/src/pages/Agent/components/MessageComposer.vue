<template>
  <div class="composer-shell">
    <div class="composer">
      <textarea
        v-model="content"
        rows="1"
        maxlength="4000"
        :disabled="running"
        placeholder="描述你想查询或完成的事情…"
        @keydown.enter.exact.prevent="submit"
      ></textarea>
      <button class="send-button" type="button" :disabled="!canSend" aria-label="发送" @click="submit">
        <el-icon v-if="!running"><Promotion /></el-icon>
        <span v-else class="sending-dot"></span>
      </button>
    </div>
    <p>Enter 发送 · Shift + Enter 换行</p>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { Promotion } from '@element-plus/icons-vue'

const props = defineProps({ running: Boolean })
const emit = defineEmits(['send'])
const content = ref('')
const canSend = computed(() => content.value.trim() && !props.running)

const submit = () => {
  if (!canSend.value) return
  emit('send', content.value.trim())
  content.value = ''
}
</script>

<style scoped>
.composer-shell { width: min(880px, calc(100% - 40px)); margin: 0 auto; padding: 13px 0 17px; }
.composer {
  min-height: 52px; display: flex; align-items: flex-end; gap: 10px; padding: 8px 8px 8px 17px;
  background: #fff; border: 1px solid #dfe5ee; border-radius: 8px; box-shadow: 0 7px 24px rgba(34, 54, 86, .08);
}
.composer:focus-within { border-color: #77a5f8; box-shadow: 0 0 0 3px rgba(54, 123, 245, .1), 0 7px 24px rgba(34, 54, 86, .08); }
textarea { width: 100%; max-height: 132px; min-height: 34px; padding: 7px 0 5px; resize: none; overflow-y: auto; color: #263248; font: inherit; font-size: 14px; line-height: 1.5; outline: none; }
textarea::placeholder { color: #a5adbb; }
.send-button { width: 36px; height: 36px; flex: 0 0 36px; display: grid; place-items: center; color: #fff; background: #367bf5; border-radius: 6px; }
.send-button:disabled { color: #aeb7c7; background: #edf1f6; cursor: not-allowed; }
.composer-shell > p { margin: 7px 2px 0; color: #adb4c1; font-size: 11px; text-align: center; }
.sending-dot { width: 15px; height: 15px; border: 2px solid rgba(255,255,255,.45); border-top-color: #fff; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) { .composer-shell { width: calc(100% - 24px); padding-bottom: 10px; } }
</style>
