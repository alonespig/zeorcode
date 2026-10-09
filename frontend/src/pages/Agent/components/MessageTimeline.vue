<template>
  <div ref="scrollElement" class="timeline">
    <div v-if="loading" class="timeline-loading"><span></span><span></span><span></span></div>
    <div v-else-if="!messages.length && !streamingContent" class="welcome">
      <div class="welcome-mark">Z</div>
      <h1>有什么可以帮你？</h1>
      <p>查询平台信息、规划教学任务，或者通过几个简单步骤创建团队作业。</p>
      <div class="suggestions">
        <button type="button" @click="$emit('suggest', '给我的团队创建一个作业')"><span>创建团队作业</span><b>→</b></button>
        <button type="button" @click="$emit('suggest', '帮我规划一周的算法训练内容')"><span>规划一周训练内容</span><b>→</b></button>
        <button type="button" @click="$emit('suggest', '介绍一下这个 OJ 可以完成哪些教学任务')"><span>了解平台教学能力</span><b>→</b></button>
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
.timeline { min-height: 0; flex: 1; overflow-y: auto; padding: 42px 0 24px; scroll-behavior: smooth; }
.timeline-loading { height: 100%; display: flex; align-items: center; justify-content: center; gap: 5px; }
.timeline-loading span, .thinking span { width: 6px; height: 6px; border-radius: 50%; background: #8b8b8b; animation: pulse 1.2s ease-in-out infinite; }
.timeline-loading span:nth-child(2), .thinking span:nth-child(2) { animation-delay: .16s; }
.timeline-loading span:nth-child(3), .thinking span:nth-child(3) { animation-delay: .32s; }
.welcome { width: min(650px, calc(100% - 36px)); margin: 11vh auto 0; text-align: center; }
.welcome-mark { width: 42px; height: 42px; display: grid; place-items: center; margin: 0 auto 20px; color: #fff; border-radius: 12px; background: #202123; font-size: 17px; font-weight: 700; }
.welcome h1 { margin: 0; color: #2f2f2f; font-size: 28px; font-weight: 600; letter-spacing: -.035em; }
.welcome > p { max-width: 520px; margin: 12px auto 29px; color: #777; font-size: 13px; line-height: 1.7; }
.suggestions { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.suggestions button { min-height: 58px; display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 11px 14px; color: #444; border: 1px solid #e5e5e5; border-radius: 12px; background: #fff; font-size: 12px; line-height: 1.5; text-align: left; transition: background-color .15s ease, border-color .15s ease; }
.suggestions button b { color: #aaa; font-size: 15px; font-weight: 400; }
.suggestions button:hover { border-color: #d5d5d5; background: #f7f7f7; }
.message-stack { padding-bottom: 15px; }
.thinking { width: min(760px, calc(100% - 48px)); display: flex; align-items: center; gap: 5px; margin: -8px auto 24px; padding-left: 42px; }
.thinking em { margin-left: 5px; color: #9ca5b4; font-size: 11px; font-style: normal; }
@keyframes pulse { 0%, 60%, 100% { opacity: .35; transform: translateY(0); } 30% { opacity: 1; transform: translateY(-3px); } }
@media (max-width: 760px) { .timeline { padding-top: 34px; } .suggestions { grid-template-columns: 1fr; } .thinking { width: calc(100% - 36px); padding-left: 38px; } }
</style>
