<template>
  <article :class="['message-row', message.role]">
    <div v-if="message.role === 'assistant'" class="assistant-avatar" aria-hidden="true">
      <span></span><span></span><span></span>
    </div>
    <div class="message-content">
      <div v-if="message.content" class="message-text">
        <MarkdownView :content="message.content" />
      </div>
      <BlockRenderer
        v-if="message.blocks?.length"
        :blocks="message.blocks"
        :pending-request-id="pendingRequestId"
        :pending-action-id="pendingActionId"
        :running="running"
        @answer="(requestId, value) => $emit('answer', requestId, value)"
        @approve="(actionId, version) => $emit('approve', actionId, version)"
        @reject="(actionId, version) => $emit('reject', actionId, version)"
      />
      <time v-if="message.createdAt">{{ formatTime(message.createdAt) }}</time>
    </div>
  </article>
</template>

<script setup>
import MarkdownView from '@/components/MarkdownView.vue'
import BlockRenderer from './BlockRenderer.vue'

defineProps({
  message: { type: Object, required: true },
  pendingRequestId: { type: Number, default: 0 },
  pendingActionId: { type: Number, default: 0 },
  running: Boolean,
})
defineEmits(['answer', 'approve', 'reject'])

const formatTime = (value) => {
  const text = String(value || '')
  return text.length >= 16 ? text.slice(11, 16) : text
}
</script>

<style scoped>
.message-row { width: min(880px, calc(100% - 40px)); display: flex; gap: 12px; margin: 0 auto 27px; }
.message-row.user { justify-content: flex-end; }
.assistant-avatar { width: 30px; height: 30px; flex: 0 0 30px; display: flex; align-items: flex-end; justify-content: center; gap: 2px; padding: 7px 6px; border-radius: 6px; background: #367bf5; }
.assistant-avatar span { width: 3px; border-radius: 3px; background: #fff; }
.assistant-avatar span:nth-child(1) { height: 7px; opacity: .78; }
.assistant-avatar span:nth-child(2) { height: 13px; }
.assistant-avatar span:nth-child(3) { height: 10px; opacity: .9; }
.message-content { min-width: 0; max-width: calc(100% - 42px); }
.assistant .message-content { flex: 1; }
.message-text { color: #334057; font-size: 14px; line-height: 1.8; }
.user .message-text { max-width: 650px; padding: 10px 14px; color: #fff; border-radius: 7px 7px 2px 7px; background: #367bf5; line-height: 1.65; }
.message-text :deep(.markdown-body > :first-child) { margin-top: 0; }
.message-text :deep(.markdown-body > :last-child) { margin-bottom: 0; }
.message-text :deep(p) { margin: 0 0 8px; }
.user .message-text :deep(*) { color: inherit; }
time { display: block; margin-top: 7px; color: #adb4c1; font-size: 10px; }
.user time { text-align: right; }
@media (max-width: 760px) { .message-row { width: calc(100% - 24px); } .message-content { max-width: calc(100% - 38px); } }
</style>
