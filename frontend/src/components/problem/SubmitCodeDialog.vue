<template>
  <el-dialog v-model="dialogVisible" :title="`提交代码 · ${problemName || ''}`" width="660px" top="6vh"
    :close-on-click-modal="false" class="submit-code-dialog">
    <div class="sc-toolbar">
      <div class="sc-lang">
        <span class="sc-label">语言</span>
        <el-select v-model="language" style="width: 160px" placeholder="选择语言">
          <el-option v-for="item in languageList" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
      </div>
      <div class="sc-limits" v-if="timeLimitMs || memoryLimitMb">
        <span>时间 {{ timeLimitMs }}ms</span>
        <span>内存 {{ memoryLimitMb }}MB</span>
      </div>
    </div>

    <div class="editor-wrap">
      <div ref="gutterRef" class="gutter">
        <div v-for="n in lineCount" :key="n">{{ n }}</div>
      </div>
      <textarea ref="taRef" v-model="code" class="code-area" spellcheck="false" placeholder="在此粘贴 / 编写代码…"
        @scroll="syncScroll" @keydown.tab.prevent="onTab"></textarea>
    </div>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { submit } from "@/api/user";
import { useLanguages } from "@/hooks/useLanguages";

const props = defineProps({
  visible: { type: Boolean, default: false },
  problemId: { type: [String, Number], required: true },
  problemName: { type: String, default: "" },
  timeLimitMs: { type: Number, default: 0 },
  memoryLimitMb: { type: Number, default: 0 },
});
const emit = defineEmits(["update:visible"]);

const router = useRouter();
const { languageList, loadLanguages } = useLanguages();

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit("update:visible", v),
});

const language = ref(null);
const code = ref("");
const submitting = ref(false);

const taRef = ref(null);
const gutterRef = ref(null);

const lineCount = computed(() => Math.max(1, code.value.split("\n").length));

const syncScroll = (e) => {
  if (gutterRef.value) gutterRef.value.scrollTop = e.target.scrollTop;
};

// Tab 键插入 4 个空格而不是切换焦点
const onTab = (e) => {
  const ta = e.target;
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  const val = code.value;
  code.value = val.slice(0, start) + "    " + val.slice(end);
  nextTick(() => {
    ta.selectionStart = ta.selectionEnd = start + 4;
  });
};

// ---------- 草稿暂存 + 记住语言 ----------
const draftKey = computed(() => `submit-code-draft-${props.problemId}`);
const LAST_LANG_KEY = "submit-code-last-language";

const loadDraft = () => {
  let saved = null;
  try {
    saved = JSON.parse(localStorage.getItem(draftKey.value) || "null");
  } catch {
    saved = null;
  }
  const lastLang = localStorage.getItem(LAST_LANG_KEY);
  code.value = saved?.code || "";
  const candidate = Number(saved?.language ?? lastLang);
  language.value = languageList.value.some((item) => item.id === candidate)
    ? candidate
    : languageList.value[0]?.id ?? null;
};

watch([language, code], () => {
  localStorage.setItem(draftKey.value, JSON.stringify({ language: language.value, code: code.value }));
});

// 打开弹窗时回填草稿
watch(
  () => props.visible,
  async (v) => {
    if (v) {
      await loadLanguages();
      loadDraft();
    }
  }
);

// ---------- 提交 ----------
const handleSubmit = async () => {
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
      problemId: props.problemId,
      language: language.value,
      code: code.value,
    });
    localStorage.setItem(LAST_LANG_KEY, language.value);
    ElMessage.success("提交成功");
    dialogVisible.value = false;
    router.push({ name: "SubmissionDetail", params: { id: res.data.submissionID } });
  } catch (err) {
    // 业务错误 msg 已由 axios 拦截器统一弹出
    console.error(err);
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped lang="scss">
.sc-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;

  .sc-lang {
    display: flex;
    align-items: center;
    gap: 10px;

    .sc-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--el-text-color-regular);
    }
  }

  .sc-limits {
    display: flex;
    gap: 16px;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }
}

.editor-wrap {
  display: flex;
  height: 400px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
  background: #fff;
  font-family: "Fira Code", "Consolas", "Courier New", monospace;
  font-size: 14px;
  line-height: 21px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;

  &:focus-within {
    border-color: #409eff;
    box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.12);
  }
}

.gutter {
  flex: 0 0 auto;
  min-width: 46px;
  padding: 12px 8px;
  text-align: right;
  color: #b0b3bd;
  background: #fafafa;
  border-right: 1px solid #eef0f3;
  user-select: none;
  overflow: hidden;
  white-space: pre;
}

.code-area {
  flex: 1;
  border: none;
  outline: none;
  resize: none;
  padding: 12px 14px;
  margin: 0;
  font: inherit;
  color: #1f2937;
  background: transparent;
  tab-size: 4;
  white-space: pre;
  overflow: auto;
}
</style>
