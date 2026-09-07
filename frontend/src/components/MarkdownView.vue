<template>
  <div class="markdown-body" v-html="html"></div>
</template>

<script setup>
import { ref, watch } from 'vue'
import MarkdownIt from 'markdown-it';
import DOMPurify from 'dompurify'
import mk from 'markdown-it-katex'
import 'katex/dist/katex.min.css'

const props = defineProps({
  content: {
    type: String,
    default: '',
  }
})

const html = ref('')

const md = new MarkdownIt({
  html: true,        // 允许 HTML
  linkify: true,     // 自动识别链接
  typographer: true  // 美化符号
})

md.use(mk)

watch(
  () => props.content,
  (newVal) => {
    html.value = DOMPurify.sanitize(md.render(newVal || ''))
  },
  { immediate: true, }
)
</script>


<style scoped>
.markdown-body :deep(img) {
  max-width: 100%;
}
</style>
