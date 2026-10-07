<script setup lang="ts">
import dayjs from "dayjs";
import { Delete, Plus, Search } from "@element-plus/icons-vue";
import type { AgentConversationItem } from "@/api/admin/agent";

defineOptions({ name: "AgentConversationSidebar" });
defineProps<{
  conversations: AgentConversationItem[];
  activeId?: number;
  loading: boolean;
}>();
const emit = defineEmits<{
  create: [];
  select: [id: number];
  archive: [id: number];
  search: [keyword: string];
}>();

const keyword = defineModel<string>("keyword", { default: "" });
const displayTime = (value: string) => {
  const date = dayjs(value);
  if (!date.isValid()) return "";
  return date.isSame(dayjs(), "day")
    ? date.format("HH:mm")
    : date.format("MM-DD");
};
</script>

<template>
  <aside class="conversation-sidebar">
    <div class="sidebar-title">
      <span class="assistant-mark">AI</span>
      <div>
        <strong>管理助手</strong>
        <p>查询平台信息并完成教学任务</p>
      </div>
    </div>
    <el-button type="primary" class="create-button" @click="emit('create')">
      <el-icon><Plus /></el-icon>新建对话
    </el-button>
    <el-input
      v-model="keyword"
      class="search-input"
      clearable
      placeholder="搜索对话"
      :prefix-icon="Search"
      @keyup.enter="emit('search', keyword.trim())"
      @clear="emit('search', '')"
    />
    <div class="section-label">最近对话</div>
    <div v-loading="loading" class="conversation-list">
      <button
        v-for="item in conversations"
        :key="item.id"
        type="button"
        :class="['conversation-item', { active: item.id === activeId }]"
        @click="emit('select', item.id)"
      >
        <span class="conversation-title">{{ item.title }}</span>
        <span class="conversation-time">{{ displayTime(item.updatedAt) }}</span>
        <el-button
          class="archive-button"
          text
          :icon="Delete"
          aria-label="归档对话"
          @click.stop="emit('archive', item.id)"
        />
      </button>
      <el-empty
        v-if="!loading && !conversations.length"
        :image-size="52"
        description="暂无对话"
      />
    </div>
    <p class="sidebar-note">
      AI 生成的内容可能不准确，执行数据写入前请核对预览信息。
    </p>
  </aside>
</template>

<style scoped>
.conversation-sidebar {
  width: 260px;
  min-width: 260px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 20px 14px 14px;
  border-right: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-extra-light);
}
.sidebar-title {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 6px 18px;
}
.sidebar-title strong {
  color: var(--el-text-color-primary);
  font-size: 17px;
}
.sidebar-title p {
  margin: 3px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}
.assistant-mark {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 6px;
  background: var(--el-color-primary);
  font-size: 13px;
  font-weight: 700;
}
.create-button {
  width: 100%;
}
.search-input {
  margin-top: 12px;
}
.section-label {
  padding: 20px 8px 8px;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
.conversation-list {
  min-height: 120px;
  flex: 1;
  overflow-y: auto;
}
.conversation-item {
  position: relative;
  width: 100%;
  min-height: 52px;
  padding: 9px 40px 8px 11px;
  color: var(--el-text-color-regular);
  border-radius: 4px;
  text-align: left;
}
.conversation-item:hover {
  background: var(--el-fill-color-light);
}
.conversation-item.active {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
.conversation-title {
  display: block;
  overflow: hidden;
  font-size: 13px;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.conversation-time {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-placeholder);
  font-size: 11px;
}
.archive-button {
  position: absolute;
  top: 9px;
  right: 5px;
  visibility: hidden;
  color: var(--el-text-color-secondary);
}
.conversation-item:hover .archive-button {
  visibility: visible;
}
.archive-button:hover {
  color: var(--el-color-danger);
}
.sidebar-note {
  margin: 12px 5px 0;
  padding-top: 12px;
  color: var(--el-text-color-placeholder);
  border-top: 1px solid var(--el-border-color-lighter);
  font-size: 11px;
  line-height: 1.6;
}
@media (max-width: 860px) {
  .conversation-sidebar {
    width: 210px;
    min-width: 210px;
  }
}
</style>
