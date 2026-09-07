<template>
  <div class="agent-page">
    <ConversationSidebar
      :conversations="conversations"
      :active-id="current?.id || 0"
      :loading="listLoading"
      @create="handleCreate"
      @select="handleSelect"
      @archive="handleArchive"
    />
    <main class="agent-main">
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

onMounted(async () => {
  try {
    const list = await loadConversations()
    if (list.length) await loadConversation(list[0].id)
  } catch (error) {
    console.error('[Agent] 初始化对话失败', error)
  }
})
</script>

<style scoped>
.agent-page {
  --agent-bg: #f5f7fa;
  width: 100%; height: 100%; min-height: 0;
  display: flex; overflow: hidden; box-sizing: border-box;
  color: #29354a; background: var(--agent-bg);
  font-family: Inter, "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
}
.agent-main { min-width: 0; flex: 1; display: flex; flex-direction: column; background: #fff; }
@media (max-width: 760px) {
  .agent-page {
    flex-direction: column;
  }
}
</style>
