<script setup>
import { computed, onScopeDispose, shallowRef } from "vue";
import hljs from "@/utils/highlight";
import "highlight.js/styles/github.css";

const props = defineProps({
  code: { type: String, default: "" },
  language: { type: String, default: "plaintext" },
  title: { type: String, default: "" },
  variant: {
    type: String,
    default: "light",
    validator: (value) => ["light", "dark"].includes(value),
  },
  embedded: { type: Boolean, default: false },
});

const copied = shallowRef(false);
let copiedTimer = null;

const highlightedCode = computed(() => {
  if (!props.code) return "";
  try {
    return hljs.highlight(props.code, { language: props.language }).value;
  } catch {
    return hljs.highlightAuto(props.code).value;
  }
});

const lineCount = computed(() => props.code ? props.code.split("\n").length : 0);

async function copy() {
  try {
    await navigator.clipboard.writeText(props.code);
    copied.value = true;
    window.clearTimeout(copiedTimer);
    copiedTimer = window.setTimeout(() => { copied.value = false; }, 1500);
  } catch {
    copied.value = false;
  }
}

onScopeDispose(() => window.clearTimeout(copiedTimer));
</script>

<template>
  <div class="code-card" :class="[`code-card--${variant}`, { 'code-card--embedded': embedded }]">
    <div class="code-head">
      <strong class="code-title">{{ title || "源代码" }}</strong>
      <div class="code-tools">
        <span class="lang">{{ language }}</span>
        <button class="copy" type="button" @click="copy">
          <svg aria-hidden="true" viewBox="0 0 20 20" fill="none">
            <path d="M7 6.5V5.3A2.3 2.3 0 0 1 9.3 3h5.4A2.3 2.3 0 0 1 17 5.3v5.4a2.3 2.3 0 0 1-2.3 2.3h-1.2" />
            <rect x="3" y="7" width="10" height="10" rx="2.3" />
          </svg>
          {{ copied ? "已复制" : "复制" }}
        </button>
      </div>
    </div>
    <div class="code-body">
      <div class="line-numbers" aria-hidden="true">
        <div v-for="n in lineCount" :key="n">{{ n }}</div>
      </div>
      <pre class="code-viewer"><code v-html="highlightedCode"></code></pre>
    </div>
  </div>
</template>

<style scoped lang="scss">
.code-card {
  overflow: hidden;
  border: 1px solid #dbe2ea;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 3px 14px rgba(24, 34, 48, 0.035);
}
.code-card--embedded { border: 0; border-radius: 0; box-shadow: none; }

.code-head {
  min-height: 60px;
  padding: 0 22px;
  border-bottom: 1px solid #e7ebf0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  background: #fff;
}

.code-title { overflow: hidden; color: #182230; font-size: 16px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; }
.code-tools { margin-left: auto; display: flex; align-items: center; gap: 9px; }
.lang {
  min-height: 30px;
  padding: 0 10px;
  border: 0;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  color: #596579;
  background: #f3f6f9;
  font: 12px/1 "JetBrains Mono", "Cascadia Code", Consolas, monospace;
  white-space: nowrap;
}
.copy {
  min-height: 32px;
  padding: 0 11px;
  border: 1px solid #cdddf4;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #1769e0;
  background: #f7faff;
  font-size: 12px;
  cursor: pointer;
  user-select: none;
  transition: border-color 160ms ease, background-color 160ms ease;
}
.copy svg { width: 15px; height: 15px; }
.copy path, .copy rect { stroke: currentColor; stroke-width: 1.5; }
.copy:hover { border-color: #a9c6ef; background: #edf5ff; }
.copy:focus-visible { outline: 3px solid rgba(23, 105, 224, 0.18); outline-offset: 2px; }
.code-body {
  display: flex;
  color: #1f2937;
  background: #fff;
  font-family: "JetBrains Mono", "Cascadia Code", Consolas, Monaco, "Courier New", monospace;
  font-size: 13.5px;
  line-height: 1.75;
  tab-size: 2;
}
.line-numbers {
  flex: 0 0 auto;
  min-width: 50px;
  padding: 18px 13px;
  border-right: 1px solid #edf0f4;
  color: #a8b1bf;
  background: #f8fafc;
  text-align: right;
  user-select: none;
  font-variant-numeric: tabular-nums;
}
.code-viewer { flex: 1; min-width: 0; margin: 0; padding: 18px 20px; overflow-x: auto; background: #fff; }
.code-viewer code { padding: 0; color: inherit; background: transparent; }

.code-card--dark {
  .code-body { color: #e6edf3; background: #151a22; }
  .line-numbers { border-right-color: #28303b; color: #657184; background: #11161d; }
  .code-viewer { color: #e6edf3; background: #151a22; }
  .code-viewer code { color: #e6edf3; }
  :deep(.hljs-keyword), :deep(.hljs-selector-tag), :deep(.hljs-literal) { color: #d79bf0; }
  :deep(.hljs-type), :deep(.hljs-title), :deep(.hljs-built_in) { color: #70b7ff; }
  :deep(.hljs-number), :deep(.hljs-symbol) { color: #aacb79; }
  :deep(.hljs-string) { color: #e5bf73; }
  :deep(.hljs-comment) { color: #7d8a9a; }
  :deep(.hljs-meta) { color: #84c7ff; }
}

@media (max-width: 640px) {
  .code-head { min-height: 54px; padding-inline: 16px; }
  .code-tools { gap: 6px; }
  .lang { min-height: 28px; padding-inline: 8px; }
  .copy { min-height: 30px; padding-inline: 9px; }
  .code-body { font-size: 12.5px; }
  .line-numbers { min-width: 42px; padding: 15px 10px; }
  .code-viewer { padding: 15px 14px; }
}

@media (prefers-reduced-motion: reduce) {
  .copy { transition: none; }
}
</style>
