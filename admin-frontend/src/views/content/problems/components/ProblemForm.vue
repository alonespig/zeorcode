<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import { onBeforeRouteLeave, useRouter } from "vue-router";
import { ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { Download, Plus } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import { getRemoteOJList } from "@/api/admin/remote-accounts";
import {
  createProblem,
  getProblemForEdit,
  getRemoteProblem,
  getTagList,
  updateProblem,
  type CreateProblemParams,
  type RemoteProblem,
  type SaveProblemParams,
  type TagItem
} from "@/api/admin/problems";

const props = defineProps<{
  mode: "create" | "edit";
  problemId?: string;
}>();

const router = useRouter();
const formRef = ref<FormInstance>();
const loading = ref(false);
const submitting = ref(false);
const initialized = ref(false);
const saved = ref(false);
const initialSnapshot = ref("");
const tagOptions = ref<TagItem[]>([]);

const form = reactive<CreateProblemParams>(blankForm());

const isEdit = computed(() => props.mode === "edit");
const pageTitle = computed(() => (isEdit.value ? "编辑题目" : "新建题目"));
const isRemote = computed(() => Boolean(form.oj));
const currentSnapshot = computed(() => JSON.stringify(form));
const dirty = computed(
  () => initialized.value && currentSnapshot.value !== initialSnapshot.value
);

const rules: FormRules = {
  displayId: [{ required: true, message: "请输入题目编号", trigger: "blur" }],
  name: [{ required: true, message: "请输入题目名称", trigger: "blur" }],
  difficulty: [{ required: true, message: "请选择难度", trigger: "change" }],
  timeLimit: [{ required: true, message: "请输入时间限制", trigger: "change" }],
  memoryLimit: [
    { required: true, message: "请输入内存限制", trigger: "change" }
  ],
  description: [{ required: true, message: "请输入题目描述", trigger: "blur" }],
  inputFormat: [{ required: true, message: "请输入输入格式", trigger: "blur" }],
  outputFormat: [{ required: true, message: "请输入输出格式", trigger: "blur" }]
};

const importVisible = ref(false);
const importing = ref(false);
const remoteOJOptions = ref<string[]>([]);
const remoteForm = reactive({ oj: "", pid: "", asRemote: false });
const remoteResult = ref<RemoteProblem>();

function blankForm(): CreateProblemParams {
  return {
    displayId: "",
    name: "",
    difficulty: 1,
    timeLimit: 1000,
    memoryLimit: 256,
    description: "",
    inputFormat: "",
    outputFormat: "",
    hint: "",
    samples: [{ input: "", output: "", explain: "" }],
    tagsID: [],
    oj: "",
    remoteProblemId: "",
    hidden: false
  };
}

function setSnapshot() {
  nextTick(() => {
    initialSnapshot.value = currentSnapshot.value;
    initialized.value = true;
  });
}

async function initialize() {
  loading.value = true;
  try {
    const tagsPromise = getTagList();
    if (isEdit.value && props.problemId) {
      const [tagsRes, problemRes] = await Promise.all([
        tagsPromise,
        getProblemForEdit(props.problemId)
      ]);
      tagOptions.value = tagsRes.data?.tags ?? [];
      const value = problemRes.data;
      if (value) {
        Object.assign(form, {
          displayId: value.id,
          name: value.name,
          difficulty: value.difficulty,
          timeLimit: value.timeLimit,
          memoryLimit: value.memoryLimit,
          description: value.description,
          inputFormat: value.inputFormat,
          outputFormat: value.outputFormat,
          hint: value.hint,
          samples: value.samples?.length
            ? value.samples
            : [{ input: "", output: "", explain: "" }],
          tagsID: value.tagsID?.map(tag => tag.id) ?? [],
          hidden: value.hidden,
          oj: "",
          remoteProblemId: ""
        });
      }
    } else {
      const tagsRes = await tagsPromise;
      tagOptions.value = tagsRes.data?.tags ?? [];
    }
    setSnapshot();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function addSample() {
  form.samples.push({ input: "", output: "", explain: "" });
}

function removeSample(index: number) {
  if (form.samples.length === 1) return;
  form.samples.splice(index, 1);
}

function normalizedPayload(): SaveProblemParams {
  return {
    displayId: form.displayId.trim(),
    name: form.name.trim(),
    difficulty: form.difficulty,
    timeLimit: form.timeLimit,
    memoryLimit: form.memoryLimit,
    description: form.description,
    inputFormat: form.inputFormat,
    outputFormat: form.outputFormat,
    hint: form.hint,
    samples: form.samples.map(sample => ({ ...sample })),
    tagsID: [...form.tagsID],
    hidden: form.hidden
  };
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid || submitting.value) return;
  if (!form.samples.length) {
    message("请至少保留一组样例", { type: "warning" });
    return;
  }

  submitting.value = true;
  try {
    if (isEdit.value && props.problemId) {
      await updateProblem(props.problemId, normalizedPayload());
      message("题目已更新", { type: "success" });
      saved.value = true;
      await router.push("/content/problems");
    } else {
      const res = await createProblem({
        ...normalizedPayload(),
        oj: form.oj,
        remoteProblemId: form.remoteProblemId
      });
      message("题目已创建", { type: "success" });
      saved.value = true;
      const id = res.data?.id;
      await router.push(
        id && !form.oj
          ? `/content/problems/${id}/testdata`
          : "/content/problems"
      );
    }
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    submitting.value = false;
  }
}

async function openRemoteImport() {
  importVisible.value = true;
  remoteResult.value = undefined;
  if (!remoteOJOptions.value.length) {
    try {
      const res = await getRemoteOJList();
      remoteOJOptions.value = res.data?.list ?? [];
      if (!remoteForm.oj) remoteForm.oj = remoteOJOptions.value[0] ?? "";
    } catch {
      // 错误信息已由 http 拦截器统一提示
    }
  }
}

async function fetchRemote() {
  if (!remoteForm.oj || !remoteForm.pid.trim()) {
    message("请选择 OJ 并填写题号", { type: "warning" });
    return;
  }
  importing.value = true;
  try {
    const res = await getRemoteProblem({
      oj: remoteForm.oj,
      pid: remoteForm.pid.trim()
    });
    remoteResult.value = res.data;
  } catch {
    remoteResult.value = undefined;
  } finally {
    importing.value = false;
  }
}

function applyRemoteImport() {
  const value = remoteResult.value;
  if (!value) return;
  form.displayId ||= value.remoteProblemId;
  form.name = value.title;
  form.timeLimit = value.timeLimit;
  form.memoryLimit = value.memoryLimit;
  form.description = value.description;
  form.inputFormat = value.inputFormat;
  form.outputFormat = value.outputFormat;
  form.hint = value.hint ?? "";
  form.samples = value.samples?.length
    ? value.samples.map(sample => ({ ...sample }))
    : [{ input: "", output: "", explain: "" }];
  form.oj = remoteForm.asRemote ? value.oj : "";
  form.remoteProblemId = remoteForm.asRemote ? value.remoteProblemId : "";
  importVisible.value = false;
  message("远程题目已填入表单", { type: "success" });
}

onBeforeRouteLeave(async () => {
  if (!dirty.value || saved.value) return true;
  try {
    await ElMessageBox.confirm(
      "当前修改尚未保存，确认离开此页面？",
      "离开页面",
      {
        type: "warning"
      }
    );
    return true;
  } catch {
    return false;
  }
});

onMounted(initialize);
</script>

<template>
  <div v-loading="loading" class="problem-form-page">
    <div class="form-header">
      <div>
        <h2>{{ pageTitle }}</h2>
        <p>编辑题面、限制、标签与样例；Markdown 内容将按原文保存。</p>
      </div>
      <el-button v-if="!isEdit" :icon="Download" @click="openRemoteImport">
        从远程 OJ 导入
      </el-button>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <section class="form-section">
        <h3>基本信息</h3>
        <el-form-item label="题目名称" prop="name">
          <el-input v-model="form.name" maxlength="128" />
        </el-form-item>
        <div class="settings-grid">
          <el-form-item label="题号" prop="displayId">
            <el-input v-model="form.displayId" maxlength="32" />
          </el-form-item>
          <el-form-item label="难度" prop="difficulty">
            <el-select v-model="form.difficulty" class="w-full">
              <el-option label="简单" :value="1" />
              <el-option label="中等" :value="2" />
              <el-option label="困难" :value="3" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间限制（ms）" prop="timeLimit">
            <el-input-number
              v-model="form.timeLimit"
              :min="1"
              controls-position="right"
              class="w-full!"
            />
          </el-form-item>
          <el-form-item label="内存限制（MB）" prop="memoryLimit">
            <el-input-number
              v-model="form.memoryLimit"
              :min="1"
              controls-position="right"
              class="w-full!"
            />
          </el-form-item>
        </div>
        <el-form-item label="标签">
          <el-select
            v-model="form.tagsID"
            multiple
            filterable
            clearable
            class="w-full"
            placeholder="请选择标签"
          >
            <el-option
              v-for="tag in tagOptions"
              :key="tag.id"
              :label="tag.name"
              :value="tag.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="可见性">
          <div class="flex items-center gap-3">
            <el-switch v-model="form.hidden" />
            <span class="text-xs text-gray-400">
              隐藏后仅管理员可见并可提交
            </span>
          </div>
        </el-form-item>
        <el-alert
          v-if="isRemote"
          :title="`远程评测：${form.oj} / ${form.remoteProblemId}`"
          type="warning"
          show-icon
          :closable="false"
        />
      </section>

      <section class="form-section">
        <h3>题面内容</h3>
        <el-form-item label="题目描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="10" />
        </el-form-item>
        <div class="content-grid">
          <el-form-item label="输入格式" prop="inputFormat">
            <el-input v-model="form.inputFormat" type="textarea" :rows="7" />
          </el-form-item>
          <el-form-item label="输出格式" prop="outputFormat">
            <el-input v-model="form.outputFormat" type="textarea" :rows="7" />
          </el-form-item>
        </div>
        <el-form-item label="数据范围与提示">
          <el-input v-model="form.hint" type="textarea" :rows="6" />
        </el-form-item>
      </section>

      <section class="form-section">
        <div class="section-heading">
          <h3>样例</h3>
          <el-button :icon="Plus" @click="addSample">添加样例</el-button>
        </div>
        <div
          v-for="(sample, index) in form.samples"
          :key="index"
          class="sample-item"
        >
          <div class="sample-heading">
            <strong>样例 {{ index + 1 }}</strong>
            <el-button
              v-if="form.samples.length > 1"
              link
              type="danger"
              @click="removeSample(index)"
            >
              删除
            </el-button>
          </div>
          <div class="content-grid">
            <el-form-item label="输入">
              <el-input v-model="sample.input" type="textarea" :rows="6" />
            </el-form-item>
            <el-form-item label="输出">
              <el-input v-model="sample.output" type="textarea" :rows="6" />
            </el-form-item>
          </div>
          <el-form-item label="解释">
            <el-input v-model="sample.explain" type="textarea" :rows="3" />
          </el-form-item>
        </div>
      </section>

      <div class="form-actions">
        <el-button @click="router.push('/content/problems')">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
          {{ isEdit ? "保存修改" : "创建题目" }}
        </el-button>
      </div>
    </el-form>

    <el-dialog
      v-model="importVisible"
      title="从远程 OJ 导入"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="flex gap-2">
        <el-select v-model="remoteForm.oj" class="w-36!" placeholder="OJ">
          <el-option
            v-for="oj in remoteOJOptions"
            :key="oj"
            :label="oj"
            :value="oj"
          />
        </el-select>
        <el-input
          v-model="remoteForm.pid"
          placeholder="远程题号"
          @keyup.enter="fetchRemote"
        />
        <el-button type="primary" :loading="importing" @click="fetchRemote">
          拉取
        </el-button>
      </div>
      <div v-if="remoteResult" class="remote-result">
        <strong>{{ remoteResult.title }}</strong>
        <span>
          {{ remoteResult.oj }} {{ remoteResult.remoteProblemId }} ·
          {{ remoteResult.timeLimit }} ms / {{ remoteResult.memoryLimit }} MB
        </span>
      </div>
      <template #footer>
        <el-checkbox v-model="remoteForm.asRemote">
          使用远程 OJ 判题
        </el-checkbox>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!remoteResult"
          @click="applyRemoteImport"
        >
          导入表单
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.problem-form-page {
  padding: 24px;
  background: var(--el-bg-color);
  border-radius: 4px;
}

.form-header,
.section-heading,
.sample-heading,
.form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.form-header {
  padding-bottom: 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.form-header h2,
.form-section h3 {
  margin: 0;
}

.form-header p {
  margin: 6px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.form-section {
  padding-top: 24px;
  margin-top: 24px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.form-section:first-of-type {
  margin-top: 0;
  border-top: 0;
}

.form-section h3 {
  margin-bottom: 18px;
  font-size: 16px;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.sample-item {
  padding: 16px;
  margin-top: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
}

.sample-heading {
  margin-bottom: 12px;
}

.form-actions {
  justify-content: flex-end;
  padding-top: 20px;
  margin-top: 24px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.remote-result {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 14px;
  margin-top: 16px;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
}

.remote-result span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

@media (max-width: 1100px) {
  .settings-grid,
  .content-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
