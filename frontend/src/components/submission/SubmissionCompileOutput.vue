<script setup>
import { onScopeDispose, shallowRef } from "vue";

const props = defineProps({
  output: { type: String, default: "" },
  restricted: { type: Boolean, default: false },
  embedded: { type: Boolean, default: false },
});

const copied = shallowRef(false);
let copiedTimer = null;

async function copyOutput() {
  if (!props.output) return;
  try {
    await navigator.clipboard.writeText(props.output);
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
  <section class="compile-panel" :class="{ 'compile-panel--embedded': embedded }" aria-labelledby="compile-output-title">
    <header class="panel-head">
      <h2 id="compile-output-title">编译器输出</h2>
      <button v-if="output" class="copy-button" type="button" @click="copyOutput">{{ copied ? "已复制" : "复制信息" }}</button>
    </header>
    <pre v-if="output" class="terminal">{{ output }}</pre>
    <div v-else class="empty-state">{{ restricted ? "仅提交者和管理员可以查看编译器输出。" : "暂无编译器输出。" }}</div>
  </section>
</template>

<style scoped>
.compile-panel { overflow: hidden; border: 1px solid #d7dde6; border-radius: 10px; background: #fff; box-shadow: 0 3px 14px rgba(24, 34, 48, 0.035); }
.compile-panel--embedded { border: 0; border-radius: 0; box-shadow: none; }
.panel-head { min-height: 58px; padding: 0 22px; border-bottom: 1px solid #e5e9f0; display: flex; align-items: center; gap: 12px; }
.panel-head h2 { margin: 0; color: #182230; font-size: 16px; font-weight: 640; }
.copy-button { min-height: 32px; margin-left: auto; padding: 0 8px; border: 0; border-radius: 5px; color: #1769e0; background: transparent; cursor: pointer; font-size: 12px; }
.copy-button:hover { background: #edf5ff; }
.copy-button:focus-visible { outline: 3px solid rgba(23, 105, 224, 0.2); outline-offset: 2px; }
.terminal { margin: 20px 22px 22px; padding: 16px 18px; overflow-x: auto; border: 1px solid #f0d2cf; border-radius: 4px; color: #9f2f2b; background: #fff7f6; font: 12px/1.75 "Cascadia Code", "JetBrains Mono", Consolas, monospace; white-space: pre-wrap; word-break: break-word; }
.empty-state { padding: 48px 22px; color: #7b8697; font-size: 13px; text-align: center; }
</style>
