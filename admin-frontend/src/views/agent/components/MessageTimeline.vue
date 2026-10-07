<script setup lang="ts">
import { nextTick, ref, watch } from "vue";
import { MagicStick } from "@element-plus/icons-vue";
import AgentBlockRenderer from "./AgentBlockRenderer.vue";
import type { AgentMessage } from "@/api/admin/agent";

defineOptions({ name: "AgentMessageTimeline" });
const props = defineProps<{
  messages: AgentMessage[];
  pendingRequestId: number;
  pendingActionId: number;
  streamingContent: string;
  toolLabel: string;
  loading: boolean;
  running: boolean;
}>();
const emit = defineEmits<{
  suggest: [content: string];
  answer: [requestId: number, value: unknown];
  approve: [actionId: number, draftVersion: number];
  reject: [actionId: number, draftVersion: number];
}>();
const scrollElement = ref<HTMLElement>();

function formatTime(value?: string) {
  const text = String(value || "");
  return text.length >= 16 ? text.slice(11, 16) : text;
}

watch(
  () => [props.messages.length, props.streamingContent, props.running],
  async () => {
    await nextTick();
    if (scrollElement.value) {
      scrollElement.value.scrollTop = scrollElement.value.scrollHeight;
    }
  },
  { flush: "post" }
);
</script>

<template>
  <div ref="scrollElement" class="timeline">
    <div v-if="loading" class="timeline-loading">
      <el-icon class="is-loading"><MagicStick /></el-icon>正在加载对话
    </div>
    <div v-else-if="!messages.length && !streamingContent" class="welcome">
      <div class="welcome-icon">
        <el-icon><MagicStick /></el-icon>
      </div>
      <h2>今天想完成什么？</h2>
      <p>可以查询平台信息、规划教学任务，也可以通过交互步骤创建团队作业。</p>
      <div class="suggestions">
        <button
          type="button"
          @click="emit('suggest', '给我的团队创建一个作业')"
        >
          <strong>创建团队作业</strong><span>选择团队、知识点和作业设置</span>
        </button>
        <button
          type="button"
          @click="emit('suggest', '帮我规划一周的算法训练内容')"
        >
          <strong>规划训练内容</strong><span>根据知识点整理训练建议</span>
        </button>
        <button
          type="button"
          @click="emit('suggest', '介绍一下这个 OJ 可以完成哪些教学任务')"
        >
          <strong>了解平台能力</strong><span>查看可用的教学与管理能力</span>
        </button>
      </div>
    </div>
    <div v-else class="message-stack">
      <article
        v-for="item in messages"
        :key="item.id"
        :class="['message-row', item.role]"
      >
        <span v-if="item.role === 'assistant'" class="message-avatar">AI</span>
        <div class="message-body">
          <div v-if="item.content" class="message-text">{{ item.content }}</div>
          <AgentBlockRenderer
            v-if="item.blocks?.length"
            :blocks="item.blocks"
            :pending-request-id="pendingRequestId"
            :pending-action-id="pendingActionId"
            :running="running"
            @answer="(requestId, value) => emit('answer', requestId, value)"
            @approve="(actionId, version) => emit('approve', actionId, version)"
            @reject="(actionId, version) => emit('reject', actionId, version)"
          />
          <time v-if="item.createdAt">{{ formatTime(item.createdAt) }}</time>
        </div>
      </article>
      <article v-if="streamingContent" class="message-row assistant">
        <span class="message-avatar">AI</span>
        <div class="message-body">
          <div class="message-text">{{ streamingContent }}</div>
        </div>
      </article>
      <div v-else-if="running" class="thinking">
        <el-icon class="is-loading"><MagicStick /></el-icon
        >{{ toolLabel || "正在处理" }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.timeline {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding: 36px 0 24px;
  scroll-behavior: smooth;
}
.timeline-loading {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.welcome {
  width: min(720px, calc(100% - 36px));
  margin: 7vh auto 0;
  text-align: center;
}
.welcome-icon {
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  margin: 0 auto 16px;
  color: var(--el-color-primary);
  border: 1px solid var(--el-color-primary-light-7);
  background: var(--el-color-primary-light-9);
  font-size: 23px;
}
.welcome h2 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 24px;
}
.welcome > p {
  margin: 12px 0 24px;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
.suggestions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  text-align: left;
}
.suggestions button {
  min-height: 90px;
  padding: 14px;
  border: 1px solid var(--el-border-color-light);
  background: #fff;
  transition: 0.15s;
}
.suggestions button:hover {
  border-color: var(--el-color-primary-light-5);
  background: var(--el-color-primary-light-9);
}
.suggestions strong,
.suggestions span {
  display: block;
}
.suggestions strong {
  color: var(--el-text-color-primary);
  font-size: 13px;
}
.suggestions span {
  margin-top: 7px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
  line-height: 1.5;
}
.message-row {
  width: min(900px, calc(100% - 40px));
  display: flex;
  gap: 11px;
  margin: 0 auto 24px;
}
.message-row.user {
  justify-content: flex-end;
}
.message-avatar {
  width: 30px;
  height: 30px;
  flex: 0 0 30px;
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--el-color-primary);
  font-size: 10px;
  font-weight: 700;
}
.message-body {
  min-width: 0;
  max-width: calc(100% - 41px);
}
.assistant .message-body {
  flex: 1;
}
.message-text {
  color: var(--el-text-color-primary);
  font-size: 14px;
  line-height: 1.75;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.user .message-text {
  max-width: 650px;
  padding: 10px 14px;
  color: #fff;
  background: var(--el-color-primary);
}
time {
  display: block;
  margin-top: 6px;
  color: var(--el-text-color-placeholder);
  font-size: 10px;
}
.user time {
  text-align: right;
}
.thinking {
  width: min(900px, calc(100% - 40px));
  display: flex;
  align-items: center;
  gap: 7px;
  margin: 0 auto;
  padding-left: 41px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
@media (max-width: 720px) {
  .suggestions {
    grid-template-columns: 1fr;
  }
}
</style>
