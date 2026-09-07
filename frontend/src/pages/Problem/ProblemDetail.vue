<template>
  <div>
    <!-- 题面 -->
    <div class="f-panel px-8 py-5 max-md:px-4 max-md:py-4">
      <h1 class="text-[22px] font-medium text-gray-800 text-center max-md:text-[18px]">
       {{ problem?.name }}
      </h1>

      <div class="flex justify-center gap-2 mt-2 text-xs
        [&>span]:bg-amber-100 [&>span]:text-gray-500 [&>span]:px-3
        [&>span]:py-1 [&>span]:rounded
      ">
        <span>时间：
          <span class="font-mono">{{ problem?.timeLimitMs }} ms</span>
        </span>
        <span>空间：
          <span class="font-mono">{{ problem?.memoryLimitMb }} MB</span>
        </span>
      </div>

      <section class="mt-2">
        <h2 class="mb-3 text-xl font-medium text-blue-500">题目描述</h2>
        <div class="markdown-body min-h-4" v-html="renderMarkdown(problem?.description)"></div>
      </section>

      <section class="mt-6">
        <h2 class="mb-3 text-xl font-medium text-blue-500">输入格式</h2>
        <div class="markdown-body min-h-4" v-html="renderMarkdown(problem?.inputFormat)"></div>
      </section>

      <section class="mt-6">
        <h2 class="mb-3 text-xl font-medium text-blue-500">输出格式</h2>
        <div class="markdown-body min-h-4" v-html="renderMarkdown(problem?.outputFormat)"></div>
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
                  : 'border-blue-500 text-blue-500 hover:bg-blue-50'"
                  @click="copyText(`in-${index}`, sample.input)">
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
                  : 'border-blue-500 text-blue-500 hover:bg-blue-50'"
                  @click="copyText(`out-${index}`, sample.output)">
                  {{ copiedKeys.has(`out-${index}`) ? "已复制" : "复制" }}
                </button>
              </div>
              <pre
                class="overflow-x-auto rounded border border-gray-300  p-3 font-mono text-[13px]">{{ sample.output }}</pre>
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

    <!-- 编辑器 -->
    <div class="f-panel mt-5 px-8 py-5 max-md:px-4 max-md:py-4">
      <div class="mb-3 flex items-center">
        <span class="text-sm  text-gray-600">语言：</span>
        <el-select v-model="language" style="width: 120px">
          <el-option v-for="item in languages" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>
      <!-- 高度随视口收缩，避免小屏上编辑器独占一整屏 -->
      <Codemirror v-model="code" :extensions="extensions" :style="{ height: 'clamp(240px, 45vh, 400px)' }" />

      <div class="mt-4 flex">
        <el-button :disabled="!userStore.isLogin" type="primary" class="ml-auto max-md:ml-0! max-md:w-full!" :loading="submitting" @click="handleSubmit">
          <span class="iconfont icon-7 mr-2"></span>
          {{ userStore.isLogin ? '提交代码' : '登录后提交' }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, reactive, inject, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getLanguages } from '@/api/languages'
import { ElMessage } from "element-plus";
import { submit } from "@/api/user";
import { useUserStore } from "@/stores/user";
import { renderMarkdown } from "@/utils/markdown";
import { Codemirror } from 'vue-codemirror'
import { EditorView } from '@codemirror/view'
import { getEditorLanguageExtension } from '@/utils/editorLanguage'

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

// 题目数据由 ProblemLayout 注入，避免重复请求
const problem = inject("problem");

const languages = ref()
const code = ref('')
const language = ref()
const submitting = ref(false)

const fetchLanguage = async () => {
  const res = await getLanguages()
  languages.value = res.data
  // 默认选中第一门语言（数字 id），避免下拉空选导致提交时校验失败
  if (languages.value?.length && !language.value) {
    language.value = languages.value[0].id
  }
}

const customTheme = EditorView.theme({
  '&': {
    fontSize: '14px',
    border: '1px solid #d1d5db',
    borderRadius: '2px',
    overflow: 'hidden',
  },
  '&.cm-focused': {
    borderColor: '#d1d5db',
    outline: 'none',
  },
  '.cm-content': {
    fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace',
    fontVariantLigatures: 'contextual',
  },
  '.cm-gutters': {
    fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace',
  },
})

const selectedLanguageName = computed(() =>
  languages.value?.find((item) => item.id === language.value)?.name || ""
)
const extensions = computed(() => {
  const languageExtension = getEditorLanguageExtension(selectedLanguageName.value)
  return languageExtension ? [languageExtension, customTheme] : [customTheme]
})

// 页内提交：未登录先弹登录框；成功后跳评测详情页看结果
const handleSubmit = async () => {
  if (!userStore.isLogin) {
    userStore.promptLogin();
    return;
  }
  if (!language.value) {
    ElMessage.warning("请选择语言");
    return;
  }
  if (!code.value.trim()) {
    ElMessage.warning("请输入代码");
    return;
  }
  submitting.value = true;
  try {
    const res = await submit({
      problemId: route.params.id, // 对外题号（字符串），后端解析成主键
      language: language.value, // 数字 id，对应后端 SubmitCodeReq.Language int
      code: code.value,
    });
    ElMessage.success("提交成功");
    router.push({ name: "SubmissionDetail", params: { id: res.data.submissionID } });
  } catch (err) {
    // 业务错误 msg 已由 axios 拦截器统一弹出
    console.error(err);
  } finally {
    submitting.value = false;
  }
};

// 处于“已复制”状态的按钮 key 集合，锁定期内再次点击忽略
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

onMounted(fetchLanguage);
</script>
