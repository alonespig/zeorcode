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
        @cancel="$emit('cancel')"
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
defineEmits(['answer', 'approve', 'reject', 'cancel'])

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
.message-text { color: #2f2f2f; font-size: 15px; line-height: 1.75; }
.user .message-text { max-width: 620px; padding: 10px 15px; color: #23405f; border: 1px solid #d9eaff; border-radius: 18px 18px 4px 18px; background: #edf6ff; line-height: 1.65; }
.message-text :deep(.markdown-body) { min-width: 0; overflow-wrap: anywhere; }
.message-text :deep(.markdown-body > :first-child) { margin-top: 0; }
.message-text :deep(.markdown-body > :last-child) { margin-bottom: 0; }
.message-text :deep(p) { margin: 0 0 12px; }
.message-text :deep(h1),
.message-text :deep(h2),
.message-text :deep(h3),
.message-text :deep(h4) { margin: 24px 0 10px; color: #202123; line-height: 1.4; font-weight: 650; }
.message-text :deep(h1) { font-size: 20px; }
.message-text :deep(h2) { padding-bottom: 7px; border-bottom: 1px solid #ececec; font-size: 18px; }
.message-text :deep(h3) { font-size: 16px; }
.message-text :deep(h4) { font-size: 15px; }
.message-text :deep(ul),
.message-text :deep(ol) { margin: 8px 0 14px; padding-left: 1.6em; }
.message-text :deep(li) { margin: 5px 0; }
.message-text :deep(strong) { color: #202123; font-weight: 650; }
.message-text :deep(a) { color: #1769e0; text-decoration: underline; text-underline-offset: 3px; }
.message-text :deep(code) { padding: 2px 5px; border-radius: 4px; color: #b42318; background: #f3f4f6; font-family: Consolas, "SFMono-Regular", monospace; font-size: .9em; }
.message-text :deep(pre) { width: 100%; max-width: 100%; margin: 12px 0 16px; padding: 14px 16px; overflow-x: auto; border: 1px solid #e5e7eb; border-radius: 6px; background: #f7f8fa; box-sizing: border-box; }
.message-text :deep(pre code) { padding: 0; color: inherit; background: transparent; font-size: 13px; line-height: 1.65; }
.message-text :deep(blockquote) { margin: 12px 0; padding: 8px 14px; border-left: 3px solid #9dbbea; color: #5f6875; background: #f7f9fc; }
.message-text :deep(blockquote p) { margin: 0; }
.message-text :deep(table) { width: 100%; max-width: 100%; margin: 14px 0; overflow-x: auto; border-collapse: collapse; }
.message-text :deep(th),
.message-text :deep(td) { padding: 8px 12px; border: 1px solid #e2e5e9; text-align: left; }
.message-text :deep(th) { background: #f7f8fa; font-weight: 600; }
.message-text :deep(hr) { margin: 20px 0; border: 0; border-top: 1px solid #e5e7eb; }
.user .message-text :deep(*) { color: inherit; }
time { display: block; margin-top: 7px; color: #aaa; font-size: 10px; }
.user time { text-align: right; }
@media (max-width: 760px) { .message-row { width: calc(100% - 24px); } .message-content { max-width: calc(100% - 38px); } }
</style>
