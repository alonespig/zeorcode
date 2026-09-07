<template>
  <aside class="agent-sidebar">
    <div class="sidebar-heading">
      <div class="agent-symbol" aria-hidden="true">
        <span></span><span></span><span></span>
      </div>
      <div>
        <strong>AI 助手</strong>
        <p>教学与平台任务</p>
      </div>
    </div>

    <button class="new-chat" type="button" @click="$emit('create')">
      <el-icon><Plus /></el-icon>
      新建对话
    </button>

    <div class="conversation-label">最近对话</div>
    <div v-loading="loading" class="conversation-list">
      <button
        v-for="item in conversations"
        :key="item.id"
        type="button"
        :class="['conversation-item', { active: item.id === activeId }]"
        @click="$emit('select', item.id)"
      >
        <span class="conversation-title">{{ item.title }}</span>
        <span class="conversation-time">{{ displayTime(item.updatedAt) }}</span>
        <span
          class="archive-button"
          role="button"
          tabindex="0"
          aria-label="归档对话"
          @click.stop="$emit('archive', item.id)"
          @keydown.enter.stop="$emit('archive', item.id)"
        >
          <el-icon><Delete /></el-icon>
        </span>
      </button>
      <div v-if="!loading && !conversations.length" class="empty-history">
        还没有历史对话
      </div>
    </div>

    <div class="sidebar-note">
      AI 生成的内容可能不准确，涉及数据写入时请仔细确认。
    </div>
  </aside>
</template>

<script setup>
import dayjs from 'dayjs'
import { Delete, Plus } from '@element-plus/icons-vue'

defineProps({
  conversations: { type: Array, default: () => [] },
  activeId: { type: Number, default: 0 },
  loading: Boolean,
})

defineEmits(['create', 'select', 'archive'])

const displayTime = (value) => {
  const date = dayjs(value)
  if (!date.isValid()) return ''
  return date.isSame(dayjs(), 'day') ? date.format('HH:mm') : date.format('MM-DD')
}
</script>

<style scoped>
.agent-sidebar {
  width: 254px;
  flex: 0 0 254px;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 24px 16px 18px;
  background: #f7f9fc;
  border-right: 1px solid #e6ebf2;
}

.sidebar-heading { display: flex; align-items: center; gap: 11px; padding: 0 8px 23px; }
.sidebar-heading strong { display: block; color: #172033; font-size: 18px; line-height: 1.3; }
.sidebar-heading p { margin: 3px 0 0; color: #8a94a7; font-size: 12px; }

.agent-symbol {
  width: 36px; height: 36px; display: flex; align-items: flex-end; justify-content: center; gap: 3px;
  padding: 8px 7px; border-radius: 8px; background: linear-gradient(145deg, #3378f6, #5e9bff);
  box-shadow: 0 5px 14px rgba(51, 120, 246, .2);
}
.agent-symbol span { width: 4px; border-radius: 4px; background: #fff; }
.agent-symbol span:nth-child(1) { height: 9px; opacity: .78; }
.agent-symbol span:nth-child(2) { height: 16px; }
.agent-symbol span:nth-child(3) { height: 12px; opacity: .9; }

.new-chat {
  width: 100%; height: 40px; display: flex; align-items: center; justify-content: center; gap: 7px;
  color: #fff; font-size: 14px; font-weight: 600; background: #367bf5; border-radius: 6px;
  box-shadow: 0 4px 12px rgba(54, 123, 245, .16); transition: .18s ease;
}
.new-chat:hover { background: #286de8; transform: translateY(-1px); }

.conversation-label { padding: 24px 8px 9px; color: #9aa3b4; font-size: 12px; font-weight: 600; letter-spacing: .04em; }
.conversation-list { min-height: 100px; flex: 1; overflow-y: auto; }
.conversation-item {
  position: relative; width: 100%; min-height: 52px; padding: 9px 38px 8px 11px; text-align: left;
  border-radius: 6px; color: #4e596d; transition: background .15s, color .15s;
}
.conversation-item:hover { background: #eef3fb; color: #1e2a3d; }
.conversation-item.active { color: #215fc9; background: #e8f0ff; }
.conversation-title { display: block; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; font-size: 14px; }
.conversation-time { display: block; margin-top: 4px; color: #a3acbc; font-size: 11px; }
.archive-button {
  position: absolute; top: 16px; right: 11px; display: none; width: 25px; height: 25px;
  align-items: center; justify-content: center; color: #8c96a8; border-radius: 4px;
}
.conversation-item:hover .archive-button { display: flex; }
.archive-button:hover { color: #e45656; background: rgba(228, 86, 86, .08); }
.empty-history { padding: 25px 8px; color: #a1aaba; text-align: center; font-size: 13px; }
.sidebar-note { padding: 15px 8px 0; border-top: 1px solid #e7ebf2; color: #9aa3b1; font-size: 11px; line-height: 1.65; }

@media (max-width: 760px) {
  .agent-sidebar { width: 100%; flex-basis: auto; padding: 13px 14px; border-right: 0; border-bottom: 1px solid #e6ebf2; }
  .sidebar-heading, .conversation-label, .conversation-list, .sidebar-note { display: none; }
  .new-chat { width: 130px; margin-left: auto; }
}
</style>
