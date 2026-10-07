<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessageBox } from "element-plus";
import { Promotion } from "@element-plus/icons-vue";
import ConversationSidebar from "./components/ConversationSidebar.vue";
import MessageTimeline from "./components/MessageTimeline.vue";
import { useAgentConversation } from "./useAgentConversation";

defineOptions({ name: "AdminAgent" });

const keyword = ref("");
const content = ref("");
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
  toolLabel,
  loadConversations,
  selectConversation,
  createConversation,
  archiveConversation,
  sendMessage,
  answerInteraction,
  approveAction,
  rejectAction
} = useAgentConversation();

async function initialize() {
  const list = await loadConversations();
  if (list.length) await selectConversation(list[0].id);
}

async function handleArchive(id: number) {
  try {
    await ElMessageBox.confirm(
      "归档后该对话将从最近对话中隐藏，确定继续吗？",
      "归档对话",
      {
        confirmButtonText: "归档",
        cancelButtonText: "取消",
        type: "warning"
      }
    );
    await archiveConversation(id);
  } catch {
    // 用户取消归档。
  }
}

async function handleApprove(actionId: number, draftVersion: number) {
  try {
    await ElMessageBox.confirm(
      "确认后将按预览内容创建作业并写入平台数据，请再次核对信息。",
      "确认执行",
      {
        confirmButtonText: "确认创建",
        cancelButtonText: "返回检查",
        type: "warning"
      }
    );
    await approveAction(actionId, draftVersion);
  } catch {
    // 用户取消执行。
  }
}

function submit() {
  const value = content.value.trim();
  if (!value || running.value) return;
  content.value = "";
  void sendMessage(value);
}

onMounted(() => {
  void initialize();
});
</script>

<template>
  <div class="agent-page">
    <ConversationSidebar
      v-model:keyword="keyword"
      :conversations="conversations"
      :active-id="current?.id"
      :loading="listLoading"
      @create="createConversation"
      @select="selectConversation"
      @archive="handleArchive"
      @search="loadConversations"
    />
    <main class="agent-main">
      <header class="conversation-header">
        <div>
          <h3>{{ current?.title || "新对话" }}</h3>
          <p>管理员 AI 助手 · 写入操作需要确认后才会执行</p>
        </div>
        <el-tag v-if="running" type="primary" effect="plain">处理中</el-tag>
      </header>
      <MessageTimeline
        :messages="messages"
        :pending-request-id="pendingRequestId"
        :pending-action-id="pendingActionId"
        :streaming-content="streamingContent"
        :tool-label="toolLabel"
        :loading="detailLoading"
        :running="running"
        @suggest="sendMessage"
        @answer="answerInteraction"
        @approve="handleApprove"
        @reject="rejectAction"
      />
      <div class="composer-shell">
        <div class="composer">
          <el-input
            v-model="content"
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 5 }"
            maxlength="4000"
            resize="none"
            :disabled="running"
            placeholder="描述你想查询或完成的事情…"
            @keydown.enter.exact.prevent="submit"
          />
          <el-button
            type="primary"
            :icon="Promotion"
            circle
            :loading="running"
            :disabled="!content.trim() || running"
            aria-label="发送"
            @click="submit"
          />
        </div>
        <p>Enter 发送，Shift + Enter 换行</p>
      </div>
    </main>
  </div>
</template>

<style scoped>
.agent-page {
  height: calc(100vh - 126px);
  min-height: 620px;
  display: flex;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  background: #fff;
}
.agent-main {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}
.conversation-header {
  min-height: 66px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 11px 22px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.conversation-header h3 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 15px;
}
.conversation-header p {
  margin: 5px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}
.composer-shell {
  width: min(900px, calc(100% - 40px));
  margin: 0 auto;
  padding: 12px 0 15px;
}
.composer {
  display: flex;
  align-items: flex-end;
  gap: 9px;
  padding: 7px 8px 7px 14px;
  border: 1px solid var(--el-border-color);
  background: #fff;
  box-shadow: 0 5px 18px rgb(31 45 61 / 7%);
}
.composer:focus-within {
  border-color: var(--el-color-primary-light-3);
}
.composer :deep(.el-textarea__inner) {
  padding: 7px 0;
  border: 0;
  box-shadow: none;
}
.composer-shell > p {
  margin: 6px 0 0;
  color: var(--el-text-color-placeholder);
  font-size: 10px;
  text-align: center;
}
@media (max-width: 860px) {
  .agent-page {
    height: calc(100vh - 110px);
  }
}
</style>
