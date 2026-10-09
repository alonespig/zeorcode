<template>
  <article :class="['message-row', message.role]">
    <div v-if="message.role === 'assistant'" class="assistant-avatar" aria-hidden="true">
      Z
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
.message-row { width: min(760px, calc(100% - 40px)); display: flex; gap: 14px; margin: 0 auto 30px; }
.message-row.user { justify-content: flex-end; }
.assistant-avatar { width: 28px; height: 28px; flex: 0 0 28px; display: grid; place-items: center; border-radius: 8px; color: #fff; background: #202123; font-size: 11px; font-weight: 700; }
.message-content { min-width: 0; max-width: calc(100% - 42px); }
.assistant .message-content { flex: 1; }
.message-text { color: #2f2f2f; font-size: 14px; line-height: 1.8; }
.user .message-text { max-width: 620px; padding: 10px 15px; color: #2f2f2f; border-radius: 18px 18px 4px 18px; background: #f4f4f4; line-height: 1.65; }
.message-text :deep(.markdown-body > :first-child) { margin-top: 0; }
.message-text :deep(.markdown-body > :last-child) { margin-bottom: 0; }
.message-text :deep(p) { margin: 0 0 8px; }
.user .message-text :deep(*) { color: inherit; }
time { display: block; margin-top: 7px; color: #aaa; font-size: 10px; }
.user time { text-align: right; }
@media (max-width: 760px) { .message-row { width: calc(100% - 24px); } .message-content { max-width: calc(100% - 38px); } }
</style>
