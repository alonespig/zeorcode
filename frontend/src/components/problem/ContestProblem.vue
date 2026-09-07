<template>
  <div class="f-panel px-8 py-5 max-md:px-4 max-md:py-4">
    <h1 class="text-center text-[22px] font-medium text-gray-800 max-md:text-[18px]">
      <span v-if="problem?.label">{{ problem.label }}. </span>{{ problem?.name }}
    </h1>

    <div class="mt-2 flex justify-center gap-2 text-xs
      [&>span]:rounded [&>span]:bg-amber-100 [&>span]:px-3 [&>span]:py-1 [&>span]:text-gray-500">
      <span>时间：<span class="font-mono">{{ problem?.timeLimitMs }} ms</span></span>
      <span>空间：<span class="font-mono">{{ problem?.memoryLimitMb }} MB</span></span>
    </div>

    <section class="mt-2">
      <h2 class="mb-3 text-xl font-medium text-blue-500">题目描述</h2>
      <div class="markdown-body" v-html="renderMarkdown(problem?.description)"></div>
    </section>

    <section class="mt-6">
      <h2 class="mb-3 text-xl font-medium text-blue-500">输入格式</h2>
      <div class="markdown-body" v-html="renderMarkdown(problem?.inputFormat)"></div>
    </section>

    <section class="mt-6">
      <h2 class="mb-3 text-xl font-medium text-blue-500">输出格式</h2>
      <div class="markdown-body" v-html="renderMarkdown(problem?.outputFormat)"></div>
    </section>

    <section v-if="problem?.samples?.length" class="mt-6">
      <h2 class="mb-2 text-xl font-medium text-blue-500">样例</h2>
      <div v-for="(sample, index) in problem.samples" :key="index" class="mb-4 last:mb-0">
        <div class="flex gap-4 max-md:flex-col">
          <div class="min-w-0 flex-1">
            <div class="mb-1.5 flex items-center justify-between">
              <span class="text-sm font-medium text-gray-600">输入 #{{ index + 1 }}</span>
              <button class="rounded border px-2 py-0.5 text-xs transition-colors" :class="copiedKeys.has(`in-${index}`)
                ? 'border-green-600 bg-green-50 text-green-600'
                : 'border-blue-500 text-blue-500 hover:bg-blue-50'" @click="copyText(`in-${index}`, sample.input)">
                {{ copiedKeys.has(`in-${index}`) ? "已复制" : "复制" }}
              </button>
            </div>
            <pre
              class="overflow-x-auto rounded border border-gray-200 bg-gray-50 p-3 font-mono text-[13px] leading-relaxed text-gray-800">{{ sample.input }}</pre>
          </div>

          <div class="min-w-0 flex-1">
            <div class="mb-1.5 flex items-center justify-between">
              <span class="text-sm font-medium text-gray-600">输出 #{{ index + 1 }}</span>
              <button class="rounded border px-2 py-0.5 text-xs transition-colors" :class="copiedKeys.has(`out-${index}`)
                ? 'border-green-600 bg-green-50 text-green-600'
                : 'border-blue-500 text-blue-500 hover:bg-blue-50'" @click="copyText(`out-${index}`, sample.output)">
                {{ copiedKeys.has(`out-${index}`) ? "已复制" : "复制" }}
              </button>
            </div>
            <pre
              class="overflow-x-auto rounded border border-gray-300 p-3 font-mono text-[13px]">{{ sample.output }}</pre>
          </div>
        </div>

        <template v-if="sample.explain">
          <h3 class="mb-1 mt-3 text-sm font-medium text-gray-600">样例说明</h3>
          <div class="markdown-body" v-html="renderMarkdown(sample.explain)"></div>
        </template>
      </div>
    </section>

    <section v-if="problem?.hint" class="mt-6">
      <h2 class="mb-3 text-xl font-medium text-blue-500">数据范围与提示</h2>
      <div class="markdown-body" v-html="renderMarkdown(problem.hint)"></div>
    </section>
  </div>
</template>

<script setup>
import { reactive } from "vue";
import { ElMessage } from "element-plus";
import { renderMarkdown } from "@/utils/markdown";

defineProps({
  problem: {
    type: Object,
    default: () => ({}),
  },
});

// 处于"已复制"状态的按钮 key 集合，锁定期内再次点击忽略
const copiedKeys = reactive(new Set());
const COPIED_DURATION_MS = 1500;

const copyText = async (key, text) => {
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
    setTimeout(() => copiedKeys.delete(key), COPIED_DURATION_MS);
  } catch {
    ElMessage.error("复制失败");
  }
};
</script>
