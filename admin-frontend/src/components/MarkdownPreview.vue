<script setup lang="ts">
import { computed } from "vue";
import DOMPurify from "dompurify";
import MarkdownIt from "markdown-it";

defineOptions({ name: "MarkdownPreview" });

const props = withDefaults(defineProps<{ content?: string }>(), {
  content: ""
});

const markdown = new MarkdownIt({
  html: false,
  breaks: true,
  linkify: true,
  typographer: true
});

const html = computed(() => DOMPurify.sanitize(markdown.render(props.content)));
</script>

<template>
  <!-- MarkdownIt 禁止原始 HTML，DOMPurify 再清理一次生成结果。 -->
  <div class="markdown-body" v-html="html" />
</template>

<style scoped>
.markdown-body {
  overflow-wrap: anywhere;
  color: var(--el-text-color-primary);
  font-size: 14px;
  line-height: 1.8;
}
.markdown-body :deep(> :first-child) {
  margin-top: 0;
}
.markdown-body :deep(> :last-child) {
  margin-bottom: 0;
}
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin: 1.5em 0 0.65em;
  color: var(--el-text-color-primary);
  font-weight: 600;
  line-height: 1.35;
}
.markdown-body :deep(h1) {
  padding-bottom: 0.35em;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-size: 1.75em;
}
.markdown-body :deep(h2) {
  padding-bottom: 0.3em;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-size: 1.45em;
}
.markdown-body :deep(h3) {
  font-size: 1.2em;
}
.markdown-body :deep(p),
.markdown-body :deep(ul),
.markdown-body :deep(ol),
.markdown-body :deep(blockquote),
.markdown-body :deep(pre),
.markdown-body :deep(table) {
  margin: 0 0 1em;
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 1.8em;
}
.markdown-body :deep(ul) {
  list-style: disc;
}
.markdown-body :deep(ol) {
  list-style: decimal;
}
.markdown-body :deep(a) {
  color: var(--el-color-primary);
  text-decoration: none;
}
.markdown-body :deep(a:hover) {
  text-decoration: underline;
}
.markdown-body :deep(blockquote) {
  padding: 0.7em 1em;
  border-left: 4px solid var(--el-border-color);
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
}
.markdown-body :deep(code) {
  padding: 0.15em 0.35em;
  border-radius: 3px;
  background: var(--el-fill-color-light);
  font-family: Consolas, Monaco, "Courier New", monospace;
  font-size: 0.9em;
}
.markdown-body :deep(pre) {
  overflow: auto;
  padding: 14px 16px;
  border: 1px solid var(--el-border-color-lighter);
  background: #f6f8fa;
  line-height: 1.6;
}
.markdown-body :deep(pre code) {
  padding: 0;
  background: transparent;
  font-size: 13px;
}
.markdown-body :deep(table) {
  display: block;
  max-width: 100%;
  overflow: auto;
  border-spacing: 0;
  border-collapse: collapse;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  padding: 7px 12px;
  border: 1px solid var(--el-border-color);
}
.markdown-body :deep(th) {
  background: var(--el-fill-color-light);
  font-weight: 600;
}
.markdown-body :deep(img) {
  max-width: 100%;
  height: auto;
}
.markdown-body :deep(hr) {
  margin: 1.5em 0;
  border: 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
