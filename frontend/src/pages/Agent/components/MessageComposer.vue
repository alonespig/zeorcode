<template>
  <div class="composer-shell">
    <div class="composer">
      <textarea
        ref="textarea"
        v-model="content"
        rows="1"
        maxlength="4000"
        :disabled="running"
        placeholder="描述你想查询或完成的事情…"
        @input="resizeComposer"
        @paste="handlePaste"
        @keydown.enter.exact.prevent="submit"
      ></textarea>
      <button class="send-button" type="button" :disabled="!canSend" aria-label="发送" @click="submit">
        <el-icon v-if="!running"><Promotion /></el-icon>
        <span v-else class="sending-dot"></span>
      </button>
    </div>
    <p>AI 可能会出错，请检查重要信息。</p>
  </div>
</template>

<script setup>
import { computed, nextTick, ref, useTemplateRef } from 'vue'
import { ElMessage } from 'element-plus'
import { Promotion } from '@element-plus/icons-vue'
import TurndownService from 'turndown'
import { gfm } from 'turndown-plugin-gfm'

const MAX_LENGTH = 4000

// 粘贴时优先把剪贴板里的 HTML 转成 Markdown，保留标题、列表、代码块等结构
const turndown = new TurndownService({
  headingStyle: 'atx',
  bulletListMarker: '-',
  codeBlockStyle: 'fenced',
  emDelimiter: '*',
  strongDelimiter: '**',
  hr: '---',
})
turndown.use(gfm)
turndown.addRule('dropImages', { filter: 'img', replacement: () => '' })
turndown.addRule('dropTags', { filter: ['script', 'style'], replacement: () => '' })
turndown.addRule('fencedCode', {
  filter: (node) => node.nodeName === 'PRE' && node.firstChild?.nodeName === 'CODE',
  replacement: (_, node) => {
    const code = node.firstChild
    const lang = (code.getAttribute('class') || '').match(/language-([\w-]+)/)?.[1] || ''
    const body = code.textContent.replace(/\n$/, '')
    return '\n\n```' + lang + '\n' + body + '\n```\n\n'
  },
})

const props = defineProps({ running: Boolean })
const emit = defineEmits(['send'])
const content = ref('')
const textarea = useTemplateRef('textarea')
const canSend = computed(() => content.value.trim() && !props.running)

const resizeComposer = () => {
  if (!textarea.value) return
  textarea.value.style.height = 'auto'
  textarea.value.style.height = `${Math.min(textarea.value.scrollHeight, 160)}px`
}

const insertText = (text) => {
  const el = textarea.value
  if (!el) return
  const start = el.selectionStart ?? content.value.length
  const end = el.selectionEnd ?? start
  const value = content.value
  let next = value.slice(0, start) + text + value.slice(end)
  if (next.length > MAX_LENGTH) {
    next = next.slice(0, MAX_LENGTH)
    ElMessage.warning(`消息不能超过 ${MAX_LENGTH} 个字符，已截断`)
  }
  const caret = Math.min(start + text.length, MAX_LENGTH)
  content.value = next
  nextTick(() => {
    el.focus()
    el.setSelectionRange(caret, caret)
    resizeComposer()
  })
}

const handlePaste = (event) => {
  const el = textarea.value
  const clipboard = event.clipboardData
  if (!el || !clipboard) return
  const html = clipboard.getData('text/html')
  if (!html) return

  let markdown = ''
  try {
    markdown = turndown.turndown(html).trim()
  } catch {
    markdown = ''
  }
  if (!markdown) return

  event.preventDefault()
  insertText(markdown)
}

const submit = async () => {
  if (!canSend.value) return
  emit('send', content.value.trim())
  content.value = ''
  await nextTick()
  resizeComposer()
}
</script>

<style scoped>
.composer-shell { width: 100%; margin: 0 auto; padding: 18px 24px 13px; background: linear-gradient(to bottom, rgba(255,255,255,0), #fff 24%); }
.composer {
  width: min(760px, 100%); min-height: 54px; display: flex; align-items: flex-end; gap: 10px; margin: 0 auto; padding: 8px 9px 8px 17px;
  background: #fff; border: 1px solid #dedede; border-radius: 20px; box-shadow: 0 2px 10px rgba(0, 0, 0, .07);
}
.composer:focus-within { border-color: #c7c7c7; box-shadow: 0 2px 12px rgba(0, 0, 0, .1); }
textarea { width: 100%; max-height: 160px; min-height: 36px; padding: 8px 0 5px; resize: none; overflow-y: auto; color: #2f2f2f; font: inherit; font-size: 14px; line-height: 1.5; outline: none; }
textarea::placeholder { color: #999; }
.send-button { width: 36px; height: 36px; flex: 0 0 36px; display: grid; place-items: center; color: #fff; background: #202123; border-radius: 50%; }
.send-button:disabled { color: #aaa; background: #e9e9e9; cursor: not-allowed; }
.composer-shell > p { margin: 7px 2px 0; color: #aaa; font-size: 10px; text-align: center; }
.sending-dot { width: 15px; height: 15px; border: 2px solid rgba(255,255,255,.45); border-top-color: #fff; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) { .composer-shell { padding: 12px 12px 9px; } .composer { border-radius: 16px; } }
</style>
