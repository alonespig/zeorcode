<template>
  <div class="edit-problem-page page-container">
    <div class="layout">
      <section class="main">
        <el-card class="form-card f-panel" shadow="never">
          <header class="page-header">
            <div class="title">
              <el-icon><EditPen /></el-icon>
              <span>编辑题目</span>
            </div>
            <div class="actions">
              <el-button @click="goBack">返回</el-button>
              <el-button type="primary" @click="submitForm">提交</el-button>
            </div>
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
          </el-form>
        </el-card>
      </section>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { ElMessage } from "element-plus";
import { EditPen, Plus } from "@element-plus/icons-vue";
import { getTags, updateProblem, getProblemInfo } from "@/api/problems";
import { renderMarkdown } from "@/utils/markdown";
import ProblemTagSelector from "@/components/problem/ProblemTagSelector.vue";

const formRef = ref(null);
const router = useRouter();

const route = useRoute();
const params = computed(() => route.params);

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
  hidden: false,
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
    const res = await getProblemInfo(params.value.id);
    form.value = {
      displayId: res.data.id, // ProblemModel.ID 即对外题号
      name: res.data.name,
      difficulty: res.data.difficulty,
      timeLimit: res.data.timeLimit,
      memoryLimit: res.data.memoryLimit,
      description: res.data.description,
      inputFormat: res.data.inputFormat ,
      outputFormat: res.data.outputFormat,
      hint: res.data.hint,
      tagsID: res.data.tagsID.map((tag) => tag.id),
      samples: res.data.samples,
      hidden: res.data.hidden,
     };
     if(form.value.samples.length === 0) {
      form.value.samples.push({ input: "", output: "", explain: "" });
     }
  } catch (err) {
    console.error("获取题目失败", err);
  }

  try {
    const res = await getTags();
    tags.value = res.data.tags;
  } catch (err) {
    console.error("获取标签失败", err);
  }

  const saved = localStorage.getItem("edit-problem-draft");
  if (saved) {
    try {
      const parsed = JSON.parse(saved);
      form.value = { ...form.value, ...parsed };
    } catch {
      localStorage.removeItem("edit-problem-draft");
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
      const res = await updateProblem(params.value.id, form.value);
      ElMessage.success("题目更新成功");
      router.push(`/problem/${res.data.id}/file`);
    } catch (err) {
      // 业务错误 msg 已由 axios 拦截器统一弹出
      console.error(err);
    }
  });
};

const goBack = () => router.back();

watch(
  form,
  (val) => {
    localStorage.setItem("edit-problem-draft", JSON.stringify(val));
  },
  { deep: true }
);

onBeforeRouteLeave(() => {
  localStorage.removeItem("edit-problem-draft");
});
</script>

<style scoped>
.edit-problem-page {
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

.page-header .title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
  font-weight: 600;
}

.actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
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
