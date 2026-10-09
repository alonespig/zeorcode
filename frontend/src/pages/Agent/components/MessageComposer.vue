<template>
  <div class="composer-shell">
    <div class="composer">
      <textarea
        ref="textarea"
        v-model="content"
        rows="1"
        maxlength="4000"
        :disabled="running"
        placeholder="描述你想查询或完成的事情…"
        @input="resizeComposer"
        @keydown.enter.exact.prevent="submit"
      ></textarea>
      <button class="send-button" type="button" :disabled="!canSend" aria-label="发送" @click="submit">
        <el-icon v-if="!running"><Promotion /></el-icon>
        <span v-else class="sending-dot"></span>
      </button>
    </div>
    <p>AI 可能会出错，请检查重要信息。</p>
  </div>
</template>

<script setup>
import { computed, nextTick, ref, useTemplateRef } from 'vue'
import { Promotion } from '@element-plus/icons-vue'

const props = defineProps({ running: Boolean })
const emit = defineEmits(['send'])
const content = ref('')
const textarea = useTemplateRef('textarea')
const canSend = computed(() => content.value.trim() && !props.running)

const resizeComposer = () => {
  if (!textarea.value) return
  textarea.value.style.height = 'auto'
  textarea.value.style.height = `${Math.min(textarea.value.scrollHeight, 160)}px`
}

const submit = async () => {
  if (!canSend.value) return
  emit('send', content.value.trim())
  content.value = ''
  await nextTick()
  resizeComposer()
}
</script>

<style scoped>
.composer-shell { width: 100%; margin: 0 auto; padding: 18px 24px 13px; background: linear-gradient(to bottom, rgba(255,255,255,0), #fff 24%); }
.composer {
  width: min(760px, 100%); min-height: 54px; display: flex; align-items: flex-end; gap: 10px; margin: 0 auto; padding: 8px 9px 8px 17px;
  background: #fff; border: 1px solid #dedede; border-radius: 20px; box-shadow: 0 2px 10px rgba(0, 0, 0, .07);
}
.composer:focus-within { border-color: #c7c7c7; box-shadow: 0 2px 12px rgba(0, 0, 0, .1); }
textarea { width: 100%; max-height: 160px; min-height: 36px; padding: 8px 0 5px; resize: none; overflow-y: auto; color: #2f2f2f; font: inherit; font-size: 14px; line-height: 1.5; outline: none; }
textarea::placeholder { color: #999; }
.send-button { width: 36px; height: 36px; flex: 0 0 36px; display: grid; place-items: center; color: #fff; background: #202123; border-radius: 50%; }
.send-button:disabled { color: #aaa; background: #e9e9e9; cursor: not-allowed; }
.composer-shell > p { margin: 7px 2px 0; color: #aaa; font-size: 10px; text-align: center; }
.sending-dot { width: 15px; height: 15px; border: 2px solid rgba(255,255,255,.45); border-top-color: #fff; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) { .composer-shell { padding: 12px 12px 9px; } .composer { border-radius: 16px; } }
</style>
