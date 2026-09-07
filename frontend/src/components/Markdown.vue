<template>
  <MdEditor v-model="content" height="400px" :toolbars="[
    'bold',
    'italic',
    'strikeThrough',
    'title',
    'link',
    'image',
    'code',
    'preview'
  ]" />
</template>

<script setup>
import { ref } from 'vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const content = ref('')
</script>


<style>
/* 全局字体 */
.md-editor {
    font-family: "Microsoft YaHei", "微软雅黑", "Arial", sans-serif;
    font-size: 14px;       /* 控制大小 */
    line-height: 1.6;      /* 行高 */
}

/* 列表字体与行高一致 */
.md-editor ul li,
.md-editor ol li {
    font-family: "Microsoft YaHei", "微软雅黑", "Arial", sans-serif;
    font-size: 14px;
    line-height: 1.6;
}

/* 代码块 */
.md-editor pre, .md-editor code {
    font-family: "Consolas", "Courier New", monospace;
    font-size: 13px;
}
/* md-editor 工具栏整体 */
.md-editor-toolbar {
  background-color: #f8f9fb;
  border-bottom: 1px solid #e5e7eb;
  padding: 2px 10px;
}

/* 每个工具按钮 */
.md-editor-toolbar-item {
  border-radius: 0px;
  margin: 0 4px;
  transition: all 0.2s ease;
  /* margin: 0 4px;
  padding: 6px;
  transition: all 0.2s ease; */
}

/* hover 效果 */
.md-editor-toolbar-item:hover {
  background-color: #e6f0ff;
  color: #409eff; /* element 主色 */
}

/* 激活态 */
.md-editor-toolbar-item.active {
  background-color: #409eff;
  color: #fff;
}

/* icon 尺寸 */
.md-editor-toolbar-item svg {
  width: 16px;
  height: 16px;
}

.md-editor-toolbar {
  display: flex;
  gap: 26px;
}

.md-editor-toolbar-item {
  background: #fff;
  border: 1px solid #e5e7eb;
}
</style>
