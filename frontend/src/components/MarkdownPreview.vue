<template>
  <div class="markdown-body" v-html="html"></div>
</template>

<script setup>
import { ref, computed } from 'vue'
import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'
import 'katex/dist/katex.min.css'

import texmath from 'markdown-it-texmath'
import katex from 'katex'

const props = defineProps({
  content: {
    type: String,
    default: ''
  }
})

const md = new MarkdownIt({
  html: false,        // 禁止原始 HTML（防 XSS）
  linkify: true,      // 自动识别链接
  typographer: true   // 排版优化
}).use(texmath, {
  engine: katex,
  delimiters: 'dollars' // 支持 $...$ 和 $$...$$
})

function renderMarkdown(markdown) {
  const rawHtml = md.render(markdown || '')
  return DOMPurify.sanitize(rawHtml)
}

const html = computed(() => {
  return renderMarkdown(props.content)
})
</script>


<style>
.markdown-body {
  min-height: 200px;
  font-family: "Microsoft YaHei", "微软雅黑", "Arial", sans-serif;
  font-size: 14px;       /* 控制大小 */
  line-height: 1.6;      /* 行高 */
}
</style>
