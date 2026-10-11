<template>
  <div class="agent-page">
    <ConversationSidebar
      :conversations="conversations"
      :active-id="current?.id || 0"
      :loading="listLoading"
      :return-label="returnLabel"
      @back="$emit('back')"
      @create="handleCreate"
      @select="handleSelect"
      @archive="handleArchive"
    />
    <main class="agent-main">
      <header class="chat-header">
        <div class="chat-title">
          <strong>{{ current?.title || '新对话' }}</strong>
          <span><i></i> 教学任务助手</span>
        </div>
      </header>
      <MessageTimeline
        :messages="messages"
        :pending-request-id="pendingRequestId"
        :pending-action-id="pendingActionId"
        :streaming-content="streamingContent"
        :loading="detailLoading"
        :running="running"
        @suggest="sendMessage"
        @answer="answerInteraction"
        @approve="approveAction"
        @reject="rejectAction"
        @cancel="cancelWorkflow"
      />
      <MessageComposer :running="running" @send="sendMessage" />
    </main>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { ElMessageBox } from 'element-plus'
import ConversationSidebar from './components/ConversationSidebar.vue'
import MessageComposer from './components/MessageComposer.vue'
import MessageTimeline from './components/MessageTimeline.vue'
import { useAgentConversation } from '@/composables/agent/useAgentConversation'

const props = defineProps({
  returnLabel: { type: String, default: '返回' },
  startNewConversation: Boolean,
})
const emit = defineEmits(['back', 'new-conversation-started'])

const {
  conversations,
  current,
  messages,
  pendingRequestId,
  pendingActionId,
  listLoading,
  detailLoading,
  running,
  streamingContent,
  loadConversations,
  loadConversation,
  createConversation,
  archiveConversation,
  sendMessage,
  answerInteraction,
  approveAction,
  rejectAction,
} = useAgentConversation()

const handleCreate = async () => {
  if (!running.value) await createConversation()
}

const handleSelect = async (id) => {
  if (!running.value && id !== current.value?.id) await loadConversation(id)
}

const handleArchive = async (id) => {
  if (running.value) return
  try {
    await ElMessageBox.confirm('归档后将从最近对话中隐藏，确定继续吗？', '归档对话', {
      confirmButtonText: '归档',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await archiveConversation(id)
    if (conversations.value.length) await loadConversation(conversations.value[0].id)
  } catch {
    // 用户取消归档，无需提示。
  }
}

const cancelWorkflow = () => {
  if (!running.value) sendMessage('取消')
}

onMounted(async () => {
  try {
    if (props.startNewConversation) {
      await createConversation()
      emit('new-conversation-started')
      return
    }
    const list = await loadConversations()
    if (list.length) await loadConversation(list[0].id)
  } catch (error) {
    console.error('[Agent] 初始化对话失败', error)
  }
})
</script>

<style scoped>
.agent-page {
  --agent-bg: #fff;
  width: 100%; height: 100%; min-height: 0;
  display: flex; overflow: hidden; box-sizing: border-box;
  color: #202123; background: var(--agent-bg);
  font-family: "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
}
.agent-main { min-width: 0; min-height: 0; flex: 1; display: flex; flex-direction: column; overflow: hidden; background: #fff; }
.chat-header {
  height: 56px; flex: 0 0 56px; display: flex; align-items: center; padding: 0 22px;
  border-bottom: 1px solid #ececec; background: rgba(255, 255, 255, .94);
}
.chat-title { min-width: 0; display: flex; align-items: center; gap: 12px; }
.chat-title strong { max-width: min(520px, 48vw); overflow: hidden; color: #2f2f2f; font-size: 14px; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.chat-title span { display: inline-flex; align-items: center; gap: 6px; color: #8b8b8b; font-size: 11px; }
.chat-title i { width: 6px; height: 6px; border-radius: 50%; background: #10a37f; }
@media (max-width: 760px) {
  .agent-page {
    flex-direction: column;
  }
  .chat-header { height: 48px; flex-basis: 48px; padding: 0 14px; }
  .chat-title span { display: none; }
}
</style>
