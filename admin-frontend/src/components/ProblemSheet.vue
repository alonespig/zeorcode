<script setup lang="ts">
import { reactive } from "vue";
import { ElMessage } from "element-plus";
import type { ProblemForEdit } from "@/api/admin/problems";
import { renderMarkdown } from "@/utils/markdown";

defineOptions({ name: "ProblemSheet" });

defineProps<{
  problem: ProblemForEdit | null;
}>();

const copiedKeys = reactive(new Set<string>());
const COPIED_DURATION_MS = 1500;

async function copyText(key: string, text?: string) {
  if (copiedKeys.has(key)) return;
  const value = String(text ?? "");
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value);
    } else {
      const ta = document.createElement("textarea");
      ta.value = value;
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
    }
    copiedKeys.add(key);
    window.setTimeout(() => copiedKeys.delete(key), COPIED_DURATION_MS);
  } catch {
    ElMessage.error("复制失败");
  }
}
</script>

<template>
  <article class="problem-sheet">
    <h1 class="problem-title">{{ problem?.name }}</h1>

    <div class="problem-limits">
      <span>
        时间：<span class="font-mono">{{ problem?.timeLimit }} ms</span>
      </span>
      <span>
        空间：<span class="font-mono">{{ problem?.memoryLimit }} MB</span>
      </span>
    </div>

    <section class="problem-section">
      <h2>题目描述</h2>
      <div
        class="markdown-body"
        v-html="renderMarkdown(problem?.description)"
      />
    </section>

    <section class="problem-section">
      <h2>输入格式</h2>
      <div
        class="markdown-body"
        v-html="renderMarkdown(problem?.inputFormat)"
      />
    </section>

    <section class="problem-section">
      <h2>输出格式</h2>
      <div
        class="markdown-body"
        v-html="renderMarkdown(problem?.outputFormat)"
      />
    </section>

    <section v-if="problem?.samples?.length" class="problem-section">
      <h2>样例</h2>
      <div
        v-for="(sample, index) in problem?.samples ?? []"
        :key="index"
        class="sample-item"
      >
        <div class="sample-grid">
          <div class="sample-column">
            <div class="sample-heading">
              <span>输入 #{{ index + 1 }}</span>
              <button
                type="button"
                class="copy-btn"
                :class="{ 'is-copied': copiedKeys.has(`in-${index}`) }"
                @click="copyText(`in-${index}`, sample?.input)"
              >
                {{ copiedKeys.has(`in-${index}`) ? "已复制" : "复制" }}
              </button>
            </div>
            <pre>{{ sample?.input ?? "" }}</pre>
          </div>

          <div class="sample-column">
            <div class="sample-heading">
              <span>输出 #{{ index + 1 }}</span>
              <button
                type="button"
                class="copy-btn"
                :class="{ 'is-copied': copiedKeys.has(`out-${index}`) }"
                @click="copyText(`out-${index}`, sample?.output)"
              >
                {{ copiedKeys.has(`out-${index}`) ? "已复制" : "复制" }}
              </button>
            </div>
            <pre>{{ sample?.output ?? "" }}</pre>
          </div>
        </div>

        <template v-if="sample?.explain">
          <h3 class="sample-explain-title">样例说明</h3>
          <div class="markdown-body" v-html="renderMarkdown(sample.explain)" />
        </template>
      </div>
    </section>

    <section v-if="problem?.hint" class="problem-section">
      <h2>数据范围与提示</h2>
      <div class="markdown-body" v-html="renderMarkdown(problem.hint)" />
    </section>
  </article>
</template>

<style scoped>
.problem-sheet {
  padding: 20px 32px;
  color: #1f2937;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  box-shadow: 0 1px 3px rgb(0 0 0 / 10%);
}

.problem-title {
  margin: 0;
  font-size: 22px;
  font-weight: 500;
  color: #1f2937;
  text-align: center;
}

.problem-limits {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-top: 8px;
  font-size: 12px;
}

.problem-limits > span {
  padding: 4px 12px;
  color: #6b7280;
  background: #fef3c7;
  border-radius: 4px;
}

.problem-section {
  margin-top: 24px;
}

.problem-section h2 {
  margin: 0 0 12px;
  font-size: 20px;
  font-weight: 500;
  color: #3b82f6;
}

.markdown-body {
  font-size: 14px;
  line-height: 1.5;
  color: #1f2937;
  overflow-wrap: anywhere;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin: 0;
  font-size: inherit;
  font-weight: inherit;
  line-height: inherit;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 2em;
  margin: 0.5em 0;
  list-style: revert;
}

.markdown-body :deep(ul) {
  list-style-type: disc;
}

.markdown-body :deep(ol) {
  list-style-type: decimal;
}

.markdown-body :deep(li) {
  margin: 0.2em 0;
}

.markdown-body :deep(pre),
.markdown-body :deep(table) {
  display: block;
  width: max-content;
  max-width: 100%;
  overflow-x: auto;
}

.markdown-body :deep(img) {
  max-width: 100%;
  height: auto;
}

.markdown-body :deep(.katex-display) {
  overflow: auto hidden;
}

.sample-item + .sample-item {
  margin-top: 20px;
}

.sample-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.sample-column {
  min-width: 0;
}

.sample-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #4b5563;
}

.copy-btn {
  padding: 2px 8px;
  font-size: 12px;
  line-height: 1.5;
  color: #3b82f6;
  cursor: pointer;
  background: transparent;
  border: 1px solid #3b82f6;
  border-radius: 4px;
  transition:
    background-color 0.15s,
    border-color 0.15s,
    color 0.15s;
}

.copy-btn:hover {
  background: #eff6ff;
}

.copy-btn.is-copied {
  color: #16a34a;
  background: #f0fdf4;
  border-color: #16a34a;
}

.sample-column pre {
  min-height: 84px;
  padding: 12px;
  margin: 0;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #1f2937;
  white-space: pre-wrap;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
}

.sample-column:nth-child(2) pre {
  background: transparent;
  border-color: #d1d5db;
}

.sample-explain-title {
  margin: 12px 0 4px;
  font-size: 14px;
  font-weight: 500;
  color: #4b5563;
}

@media (width <= 768px) {
  .problem-sheet {
    padding: 0 4px 16px;
  }

  .problem-title {
    font-size: 18px;
  }

  .sample-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
