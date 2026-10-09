<template>
  <aside class="agent-sidebar">
    <div class="sidebar-topbar">
      <button type="button" class="back-button" :aria-label="returnLabel" @click="$emit('back')">
        <el-icon><ArrowLeft /></el-icon>
      </button>
      <div class="sidebar-brand">
        <div class="agent-symbol" aria-hidden="true">Z</div>
        <strong>ZeorCode AI</strong>
      </div>
    </div>

    <button class="new-chat" type="button" @click="$emit('create')">
      <el-icon><EditPen /></el-icon>
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

    <div class="sidebar-note">AI 生成的内容可能不准确，请检查重要信息。</div>
  </aside>
</template>

<script setup>
import dayjs from 'dayjs'
import { ArrowLeft, Delete, EditPen } from '@element-plus/icons-vue'

defineProps({
  conversations: { type: Array, default: () => [] },
  activeId: { type: Number, default: 0 },
  loading: Boolean,
  returnLabel: { type: String, default: '返回' },
})

defineEmits(['back', 'create', 'select', 'archive'])

const displayTime = (value) => {
  const date = dayjs(value)
  if (!date.isValid()) return ''
  return date.isSame(dayjs(), 'day') ? date.format('HH:mm') : date.format('MM-DD')
}
</script>

<style scoped>
.agent-sidebar {
  width: 260px;
  flex: 0 0 260px;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 10px 10px 14px;
  background: #f7f7f8;
  border-right: 1px solid #e9e9ea;
}

.sidebar-topbar { height: 42px; display: flex; align-items: center; gap: 8px; padding: 0 4px; }
.back-button { width: 34px; height: 34px; flex: 0 0 34px; display: grid; place-items: center; border-radius: 8px; color: #555; font-size: 17px; }
.back-button:hover { background: #ececee; color: #111; }
.sidebar-brand { min-width: 0; display: flex; align-items: center; gap: 8px; }
.sidebar-brand strong { overflow: hidden; color: #202123; font-size: 14px; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }

.agent-symbol {
  width: 25px; height: 25px; flex: 0 0 25px; display: grid; place-items: center;
  border-radius: 7px; color: #fff; background: #202123; font-size: 12px; font-weight: 700;
}

.new-chat {
  width: 100%; height: 40px; display: flex; align-items: center; gap: 9px; margin-top: 9px; padding: 0 12px;
  color: #2f2f2f; font-size: 13px; font-weight: 500; background: transparent; border-radius: 8px;
  transition: background-color .15s ease;
}
.new-chat:hover { background: #e9e9eb; }

.conversation-label { padding: 25px 10px 8px; color: #8f8f8f; font-size: 11px; font-weight: 500; }
.conversation-list { min-height: 100px; flex: 1; overflow-y: auto; }
.conversation-item {
  position: relative; width: 100%; min-height: 44px; padding: 7px 38px 7px 11px; text-align: left;
  border-radius: 8px; color: #424242; transition: background-color .15s ease;
}
.conversation-item:hover { background: #ececee; color: #202123; }
.conversation-item.active { color: #202123; background: #e8e8ea; }
.conversation-title { display: block; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; font-size: 14px; }
.conversation-time { display: block; margin-top: 2px; color: #9b9b9b; font-size: 10px; }
.archive-button {
  position: absolute; top: 10px; right: 9px; display: none; width: 26px; height: 26px;
  align-items: center; justify-content: center; color: #777; border-radius: 6px;
}
.conversation-item:hover .archive-button { display: flex; }
.archive-button:hover { color: #d84848; background: #dddde0; }
.empty-history { padding: 25px 8px; color: #999; text-align: center; font-size: 12px; }
.sidebar-note { padding: 13px 10px 0; border-top: 1px solid #e3e3e4; color: #999; font-size: 10px; line-height: 1.6; }

@media (max-width: 760px) {
  .agent-sidebar { width: 100%; flex-basis: auto; flex-direction: row; align-items: center; padding: 7px 10px; border-right: 0; border-bottom: 1px solid #e9e9ea; }
  .sidebar-topbar { flex: 1; }
  .conversation-label, .conversation-list, .sidebar-note { display: none; }
  .new-chat { width: auto; margin: 0; padding: 0 10px; }
}
</style>
