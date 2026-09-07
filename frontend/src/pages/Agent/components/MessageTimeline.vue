<template>
  <div ref="scrollElement" class="timeline">
    <div v-if="loading" class="timeline-loading"><span></span><span></span><span></span></div>
    <div v-else-if="!messages.length && !streamingContent" class="welcome">
      <div class="welcome-mark"><Sparkles /></div>
      <h1>今天想完成什么？</h1>
      <p>我可以协助查询平台信息、梳理教学任务，也可以通过交互步骤帮你创建团队作业。</p>
      <div class="suggestions">
        <button type="button" @click="$emit('suggest', '给我的团队创建一个作业')">创建团队作业</button>
        <button type="button" @click="$emit('suggest', '帮我规划一周的算法训练内容')">规划训练内容</button>
        <button type="button" @click="$emit('suggest', '介绍一下这个 OJ 可以完成哪些教学任务')">了解平台能力</button>
      </div>
    </div>
    <div v-else class="message-stack">
      <MessageItem
        v-for="message in messages"
        :key="message.id"
        :message="message"
        :pending-request-id="pendingRequestId"
        :pending-action-id="pendingActionId"
        :running="running"
        @answer="(requestId, value) => $emit('answer', requestId, value)"
        @approve="(actionId, version) => $emit('approve', actionId, version)"
        @reject="(actionId, version) => $emit('reject', actionId, version)"
      />
      <MessageItem
        v-if="streamingContent"
        :message="{ id: 'streaming', role: 'assistant', content: streamingContent, blocks: [] }"
      />
      <div v-else-if="running" class="thinking">
        <span></span><span></span><span></span><em>正在处理</em>
      </div>
    </div>
  </div>
</template>

<script setup>
import { nextTick, ref, watch } from 'vue'
import { MagicStick as Sparkles } from '@element-plus/icons-vue'
import MessageItem from './MessageItem.vue'

const props = defineProps({
  messages: { type: Array, default: () => [] },
  pendingRequestId: { type: Number, default: 0 },
  pendingActionId: { type: Number, default: 0 },
  streamingContent: { type: String, default: '' },
  loading: Boolean,
  running: Boolean,
})
defineEmits(['suggest', 'answer', 'approve', 'reject'])
const scrollElement = ref(null)

watch(
  () => [props.messages.length, props.streamingContent, props.running],
  async () => {
    await nextTick()
    if (scrollElement.value) scrollElement.value.scrollTop = scrollElement.value.scrollHeight
  },
  { flush: 'post' },
)
</script>

<style scoped>
.timeline { min-height: 0; flex: 1; overflow-y: auto; padding: 52px 0 28px; scroll-behavior: smooth; }
.timeline-loading { height: 100%; display: flex; align-items: center; justify-content: center; gap: 5px; }
.timeline-loading span, .thinking span { width: 6px; height: 6px; border-radius: 50%; background: #85a9e8; animation: pulse 1.2s ease-in-out infinite; }
.timeline-loading span:nth-child(2), .thinking span:nth-child(2) { animation-delay: .16s; }
.timeline-loading span:nth-child(3), .thinking span:nth-child(3) { animation-delay: .32s; }
.welcome { width: min(660px, calc(100% - 36px)); margin: 7vh auto 0; text-align: center; }
.welcome-mark { width: 48px; height: 48px; display: grid; place-items: center; margin: 0 auto 19px; color: #367bf5; border: 1px solid #d7e4fb; border-radius: 10px; background: #f2f7ff; font-size: 23px; }
.welcome h1 { margin: 0; color: #202b3d; font-size: 25px; font-weight: 650; letter-spacing: -.02em; }
.welcome > p { max-width: 540px; margin: 13px auto 25px; color: #7d8799; font-size: 14px; line-height: 1.75; }
.suggestions { display: grid; grid-template-columns: repeat(3, 1fr); gap: 9px; }
.suggestions button { min-height: 64px; padding: 10px 12px; color: #556277; border: 1px solid #e0e6ee; border-radius: 7px; background: #fff; font-size: 13px; line-height: 1.5; transition: .15s; }
.suggestions button:hover { color: #2869d5; border-color: #9ebef5; background: #f8faff; transform: translateY(-1px); }
.message-stack { padding-bottom: 15px; }
.thinking { width: min(838px, calc(100% - 82px)); display: flex; align-items: center; gap: 5px; margin: -8px auto 24px; padding-left: 42px; }
.thinking em { margin-left: 5px; color: #9ca5b4; font-size: 11px; font-style: normal; }
@keyframes pulse { 0%, 60%, 100% { opacity: .35; transform: translateY(0); } 30% { opacity: 1; transform: translateY(-3px); } }
@media (max-width: 760px) { .timeline { padding-top: 34px; } .suggestions { grid-template-columns: 1fr; } .thinking { width: calc(100% - 36px); padding-left: 38px; } }
</style>
