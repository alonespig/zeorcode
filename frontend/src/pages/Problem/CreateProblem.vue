<template>
  <div class="create-problem-page page-container">
    <div class="layout">
      <section class="main">
        <el-card class="form-card f-panel" shadow="never">
          <header class="page-header">
            <div class="title">
              <el-icon><EditPen /></el-icon>
              <span>新建题目</span>
              <span v-if="form.oj" class="remote-badge">
                <el-icon><Link /></el-icon>远程 · {{ form.oj }}-{{ form.remoteProblemId }}
              </span>
            </div>
            <el-button :icon="Download" @click="openImport">从 OJ 导入</el-button>
          </header>
          <el-form :model="form" :rules="rules" ref="formRef" label-position="top">
            <div class="form-section-title">基础信息</div>
            <el-form-item label="标题" prop="name">
              <el-input v-model="form.name" placeholder="请输入标题" />
            </el-form-item>
            <div class="settings-grid">
              <el-form-item label="题目编号" prop="displayId">
                <el-input v-model.trim="form.displayId" placeholder="唯一题号，如 A、1001、P1234" maxlength="32" />
              </el-form-item>
              <el-form-item label="难度" prop="difficulty">
                <el-select v-model="form.difficulty" placeholder="请选择难度" class="full">
                  <el-option label="简单" :value="1" />
                  <el-option label="中等" :value="2" />
                  <el-option label="困难" :value="3" />
                </el-select>
              </el-form-item>
              <el-form-item label="时间限制 (ms)" prop="timeLimit">
                <el-input-number v-model="form.timeLimit" :min="1" class="full" />
              </el-form-item>
              <el-form-item label="内存限制 (MB)" prop="memoryLimit">
                <el-input-number v-model="form.memoryLimit" :min="1" class="full" />
              </el-form-item>
            </div>
            <div class="secondary-settings">
              <el-form-item label="标签">
                <ProblemTagSelector v-model="form.tagsID" :tags="tags" />
              </el-form-item>
              <el-form-item label="隐藏题目" class="hidden-setting">
                <el-switch v-model="form.hidden" />
                <span class="switch-hint">开启后仅管理员可见/可提交</span>
              </el-form-item>
            </div>

            <div class="block">
              <div class="block-header">
                <span class="block-title">题目描述</span>
                <div class="block-tools">
                  <el-checkbox v-model="preview.description">预览</el-checkbox>
                </div>
              </div>
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="10"
                placeholder="栏目内容"
              />
              <div v-if="preview.description" class="preview-box markdown-body" v-html="renderMarkdown(form.description)"></div>
            </div>

            <div class="block">
              <div class="block-header">
                <span class="block-title">输入格式</span>
                <div class="block-tools">
                  <el-checkbox v-model="preview.inputFormat">预览</el-checkbox>
                </div>
              </div>
              <el-input
                v-model="form.inputFormat"
                type="textarea"
                :rows="8"
                placeholder="栏目内容"
              />
              <div v-if="preview.inputFormat" class="preview-box markdown-body" v-html="renderMarkdown(form.inputFormat)"></div>
            </div>

            <div class="block">
              <div class="block-header">
                <span class="block-title">输出格式</span>
                <div class="block-tools">
                  <el-checkbox v-model="preview.outputFormat">预览</el-checkbox>
                </div>
              </div>
              <el-input
                v-model="form.outputFormat"
                type="textarea"
                :rows="8"
                placeholder="栏目内容"
              />
              <div v-if="preview.outputFormat" class="preview-box markdown-body" v-html="renderMarkdown(form.outputFormat)"></div>
            </div>

            <div
              v-for="(sample, index) in form.samples"
              :key="index"
              class="block"
            >
              <div class="block-header">
                <span class="block-title">样例{{ index + 1 }}</span>
                <div class="block-tools">
                  <el-checkbox v-model="preview.sample">预览</el-checkbox>
                  <el-button :icon="Plus" circle @click="addSample" />
                  <el-button
                    v-if="form.samples.length > 1"
                    type="danger"
                    plain
                    @click="removeSample(index)"
                  >
                    删除
                  </el-button>
                </div>
              </div>

              <div class="sample-grid">
                <div class="sample-col">
                  <div class="sample-label">样例输入</div>
                  <el-input
                    v-model="sample.input"
                    type="textarea"
                    :rows="6"
                    placeholder="样例输入"
                  />
                </div>
                <div class="sample-col">
                  <div class="sample-label">样例输出</div>
                  <el-input
                    v-model="sample.output"
                    type="textarea"
                    :rows="6"
                    placeholder="样例输出"
                  />
                </div>
              </div>
              <div class="sample-explain">
                <div class="sample-label">样例解释</div>
                <el-input
                  v-model="sample.explain"
                  type="textarea"
                  :rows="5"
                  placeholder="样例解释"
                />
                <div
                  v-if="preview.sample"
                  class="preview-box markdown-body"
                  v-html="renderMarkdown(sample.explain)"
                ></div>
              </div>
            </div>

            <div class="block">
              <div class="block-header">
                <span class="block-title">数据范围与提示</span>
                <div class="block-tools">
                  <el-checkbox v-model="preview.hint">预览</el-checkbox>
                </div>
              </div>
              <el-input
                v-model="form.hint"
                type="textarea"
                :rows="8"
                placeholder="栏目内容"
              />
              <div v-if="preview.hint" class="preview-box markdown-body" v-html="renderMarkdown(form.hint)"></div>
            </div>

            <div class="form-actions">
              <el-button @click="goBack">取消</el-button>
              <el-button type="primary" @click="submitForm">创建题目</el-button>
            </div>
          </el-form>
        </el-card>
      </section>

    </div>

    <!-- 导入远程题目 -->
    <el-dialog v-model="importVisible" title="导入远程题目" width="560px">
      <div class="import-bar">
        <el-select v-model="importForm.oj" placeholder="OJ" style="width: 120px">
          <el-option v-for="o in ojList" :key="o" :label="o" :value="o" />
        </el-select>
        <el-input
          v-model="importForm.pid"
          placeholder="题号"
          style="flex: 1"
          @keyup.enter="fetchRemote"
        />
        <el-button type="primary" :loading="fetching" @click="fetchRemote">拉取</el-button>
      </div>

      <div v-if="fetched" class="import-preview">
        <div class="ip-head">
          <span class="ip-title">{{ fetched.title }}</span>
          <span class="ip-meta">
            {{ fetched.oj }} · {{ fetched.remoteProblemId }}&nbsp;|&nbsp;{{ fetched.timeLimit }} ms / {{ fetched.memoryLimit }} MB
          </span>
        </div>
        <div class="ip-body markdown-body" v-html="renderMarkdown(fetched.description)"></div>
      </div>

      <template #footer>
        <el-checkbox v-model="importForm.asRemote" class="import-remote-check">
          远程判题（提交到原 OJ 评测，需在后台配好该 OJ 账号；否则导入为本地题）
        </el-checkbox>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!fetched" @click="applyImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from "vue";
import { useRouter, onBeforeRouteLeave } from "vue-router";
import { ElMessage } from "element-plus";
import { EditPen, Plus, Download, Link } from "@element-plus/icons-vue";
import { getTags, createProblem } from "@/api/problems";
import { getRemoteOJList, getRemoteProblem } from "@/api/remote";
import { renderMarkdown } from "@/utils/markdown";
import ProblemTagSelector from "@/components/problem/ProblemTagSelector.vue";

const formRef = ref(null);
const router = useRouter();

const preview = ref({
  description: false,
  inputFormat: false,
  outputFormat: false,
  sample: false,
  hint: false,
});

const form = ref({
  displayId: "",
  name: "",
  difficulty: null,
  timeLimit: 1000,
  memoryLimit: 256,
  description: "",
  inputFormat: "",
  outputFormat: "",
  hint: "",
  tagsID: [],
  samples: [
    { input: "", output: "", explain: "" },
  ],
  oj: "",              // 非空表示远程题
  remoteProblemId: "",
  hidden: false,       // 隐藏题：仅管理员可见/可提交
});

const rules = {
  displayId: [{ required: true, message: "请输入题目编号", trigger: "blur" }],
  name: [{ required: true, message: "题目名称不能为空", trigger: "blur" }],
  difficulty: [{ required: true, message: "请选择难度", trigger: "change" }],
  timeLimit: [{ required: true, message: "请输入时间限制", trigger: "change" }],
  memoryLimit: [{ required: true, message: "请输入内存限制", trigger: "change" }],
  description: [{ required: true, message: "请输入题目描述", trigger: "blur" }],
};

const tags = ref([]);

onMounted(async () => {
  try {
    const res = await getTags();
    tags.value = res.data.tags;
  } catch (err) {
    console.error("获取标签失败", err);
  }

  const saved = localStorage.getItem("create-problem-draft");
  if (saved) {
    try {
      const parsed = JSON.parse(saved);
      form.value = { ...form.value, ...parsed };
    } catch {
      localStorage.removeItem("create-problem-draft");
    }
  }
});

const addSample = () => {
  form.value.samples.push({ input: "", output: "", explain: "" });
};

const removeSample = (index) => {
  if (form.value.samples.length <= 1) return;
  form.value.samples.splice(index, 1);
};

const submitForm = () => {
  formRef.value.validate(async (valid) => {
    if (!valid) return;
    try {
      const res = await createProblem(form.value);
      ElMessage.success("题目创建成功");
      router.push(`/problem/${res.data.id}/file`);
    } catch (err) {
      // 业务错误 msg 已由 axios 拦截器统一弹出
      console.error(err);
    }
  });
};

const goBack = () => router.back();

// ===== 导入远程题目 =====
const importVisible = ref(false);
const ojList = ref([]);
const importForm = ref({ oj: "", pid: "", asRemote: false });
const fetching = ref(false);
const fetched = ref(null);

const openImport = async () => {
  importVisible.value = true;
  if (!ojList.value.length) {
    try {
      const res = await getRemoteOJList();
      ojList.value = res.data.list || [];
      if (ojList.value.length && !importForm.value.oj) {
        importForm.value.oj = ojList.value[0];
      }
    } catch (err) {
      console.error(err);
    }
  }
};

const fetchRemote = async () => {
  if (!importForm.value.oj || !importForm.value.pid) {
    ElMessage.warning("请选择 OJ 并填写题号");
    return;
  }
  fetching.value = true;
  try {
    const res = await getRemoteProblem({
      oj: importForm.value.oj,
      pid: importForm.value.pid.trim(),
    });
    fetched.value = res.data;
  } catch (err) {
    console.error(err);
  } finally {
    fetching.value = false;
  }
};

const applyImport = () => {
  const p = fetched.value;
  if (!p) return;
  form.value.name = p.title;
  form.value.description = p.description;
  form.value.inputFormat = p.inputFormat;
  form.value.outputFormat = p.outputFormat;
  form.value.timeLimit = p.timeLimit;
  form.value.memoryLimit = p.memoryLimit;
  form.value.hint = p.hint || "";
  if (p.samples && p.samples.length) {
    form.value.samples = p.samples.map((s) => ({
      input: s.input,
      output: s.output,
      explain: s.explain || "",
    }));
  }
  // 勾了「远程判题」才打远程标记(worker 会提交到原 OJ 判)；否则导入为本地题、本地 go-judge 评测
  if (importForm.value.asRemote) {
    form.value.oj = p.oj;
    form.value.remoteProblemId = p.remoteProblemId;
  } else {
    form.value.oj = "";
    form.value.remoteProblemId = "";
  }
  importVisible.value = false;
  fetched.value = null;
  ElMessage.success("已导入到表单");
};

watch(
  form,
  (val) => {
    localStorage.setItem("create-problem-draft", JSON.stringify(val));
  },
  { deep: true }
);

onBeforeRouteLeave(() => {
  localStorage.removeItem("create-problem-draft");
});
</script>

<style scoped>
.create-problem-page {
  padding-bottom: 28px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 16px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--color-border-soft);
}

.remote-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 10px;
  font-size: 12px;
  color: var(--color-primary);
  background: var(--el-color-primary-light-9);
  padding: 3px 8px;
  border-radius: var(--radius-small);
}

.import-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.import-preview {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.import-preview .ip-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 10px;
}

.import-preview .ip-title {
  font-size: 15px;
  font-weight: 500;
}

.import-preview .ip-meta {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
}

.import-preview .ip-body {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  padding: 12px 14px;
  font-size: 13px;
  max-height: 200px;
  overflow: auto;
}

.page-header .title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
  font-weight: 600;
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding-top: 20px;
  margin-top: 20px;
  border-top: 1px solid var(--color-border-soft);
}

.form-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.layout {
  display: block;
}

.form-card :deep(.el-card__body) {
  padding: var(--space-card);
}

.form-section-title {
  margin-bottom: 14px;
  color: var(--el-text-color-primary);
  font-size: 15px;
  font-weight: 600;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.secondary-settings {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.secondary-settings :deep(.el-form-item) {
  margin-bottom: 0;
}

.hidden-setting :deep(.el-form-item__content) {
  min-height: 32px;
}

.block {
  padding-top: 20px;
  margin-top: 20px;
  border-top: 1px solid var(--color-border-soft);
}

.block-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.block-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.block-tools {
  display: flex;
  align-items: center;
  gap: 10px;
}


.sample-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
}

.sample-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-bottom: 6px;
}

.sample-explain {
  margin-top: 6px;
}

.preview-box {
  margin-top: 12px;
  padding: 14px;
  border: 1px dashed var(--color-border-soft);
  border-radius: var(--radius-base);
  background: var(--table-striped-bg);
  color: var(--el-text-color-regular);
  line-height: 1.8;
}

.preview-box :deep(p) {
  margin: 0 0 12px;
}

.preview-box :deep(p:last-child) {
  margin-bottom: 0;
}

.preview-box :deep(pre) {
  background: var(--color-page-bg);
  padding: 10px;
  border-radius: var(--radius-small);
  overflow-x: auto;
}

.full {
  width: 100%;
}

.switch-hint {
  margin-left: 10px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

@media (max-width: 1100px) {
  .settings-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
