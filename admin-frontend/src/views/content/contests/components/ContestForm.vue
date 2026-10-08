<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import { onBeforeRouteLeave, useRouter } from "vue-router";
import { ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { ArrowDown, ArrowUp, Plus } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import ProblemPickerDialog from "@/views/content/problemsets/components/ProblemPickerDialog.vue";
import { getProblem } from "@/api/admin/problems";
import {
  createContest,
  getContestInfo,
  updateContest,
  type ContestProblemInput
} from "@/api/admin/contests";

const props = defineProps<{
  mode: "create" | "edit";
  contestId?: number;
}>();

interface ContestProblemRow {
  id: string;
  name: string;
  color: string;
  score: number;
}

interface ContestFormState {
  name: string;
  timeRange: [Date, Date] | [];
  ruleType: number;
  rated: boolean;
  description: string;
  coverUrl: string;
  inviteCode: string;
  requireInviteCode: boolean;
}

const RULE_OPTIONS = [
  { value: 1, label: "ACM" },
  { value: 2, label: "OI" },
  { value: 3, label: "IOI" },
  { value: 4, label: "CF" }
];
const COLORS = [
  "#e74c3c",
  "#3498db",
  "#2ecc71",
  "#f1c40f",
  "#9b59b6",
  "#e67e22",
  "#1abc9c",
  "#ff7faa",
  "#34495e"
];

const router = useRouter();
const formRef = ref<FormInstance>();
const loading = ref(false);
const submitting = ref(false);
const initialized = ref(false);
const saved = ref(false);
const initialSnapshot = ref("");
const pickerVisible = ref(false);
const queryProblemID = ref("");
const problems = ref<ContestProblemRow[]>([]);
const started = ref(false);
const form = reactive<ContestFormState>({
  name: "",
  timeRange: [],
  ruleType: 1,
  rated: false,
  description: "",
  coverUrl: "",
  inviteCode: "",
  requireInviteCode: false
});

const isEdit = computed(() => props.mode === "edit");
const isScored = computed(() => form.ruleType !== 1);
const existingIds = computed(() => problems.value.map(item => item.id));
const snapshot = computed(() =>
  JSON.stringify({ form, problems: problems.value })
);
const dirty = computed(
  () => initialized.value && snapshot.value !== initialSnapshot.value
);

const rules: FormRules = {
  name: [{ required: true, message: "请输入比赛名称", trigger: "blur" }],
  timeRange: [
    { required: true, message: "请选择比赛开始和结束时间", trigger: "change" }
  ],
  ruleType: [{ required: true, message: "请选择赛制", trigger: "change" }]
};

function setSnapshot() {
  nextTick(() => {
    initialSnapshot.value = snapshot.value;
    initialized.value = true;
  });
}

function nextColor() {
  return COLORS[problems.value.length % COLORS.length];
}

function addRows(items: Array<{ id: string; name: string }>) {
  for (const item of items) {
    if (problems.value.some(problem => problem.id === item.id)) continue;
    problems.value.push({
      id: item.id,
      name: item.name,
      color: nextColor(),
      score: 100
    });
  }
}

async function addByID() {
  const id = queryProblemID.value.trim();
  if (!id) {
    message("请输入题号", { type: "warning" });
    return;
  }
  if (problems.value.some(item => item.id === id)) {
    message("该题已添加", { type: "warning" });
    return;
  }
  try {
    const res = await getProblem(id);
    if (res.data) addRows([{ id: res.data.id, name: res.data.name }]);
    queryProblemID.value = "";
  } catch {
    // 错误信息已由 http 拦截器统一提示
  }
}

function moveProblem(index: number, offset: number) {
  const target = index + offset;
  if (target < 0 || target >= problems.value.length) return;
  const [item] = problems.value.splice(index, 1);
  problems.value.splice(target, 0, item);
}

function pad(value: number) {
  return String(value).padStart(2, "0");
}

function createProblemPayload(): ContestProblemInput[] {
  return problems.value.map(item => ({
    problemId: item.id,
    color: item.color,
    score: item.score
  }));
}

async function initialize() {
  if (!isEdit.value || !props.contestId) {
    const start = new Date();
    start.setMinutes(Math.ceil(start.getMinutes() / 15) * 15, 0, 0);
    const end = new Date(start.getTime() + 2 * 60 * 60 * 1000);
    form.timeRange = [start, end];
    setSnapshot();
    return;
  }

  loading.value = true;
  try {
    const res = await getContestInfo(props.contestId);
    const value = res.data;
    if (value) {
      form.name = value.title;
      form.timeRange = [new Date(value.startTime), new Date(value.endTime)];
      form.ruleType = value.ruleType;
      form.rated = value.rated;
      form.description = value.description;
      form.coverUrl = value.coverUrl;
      problems.value = (value.problems ?? []).map(item => ({
        id: item.id,
        name: item.name,
        color: item.color,
        score: item.score || 100
      }));
      started.value = Date.now() >= value.startTime;
    }
    setSnapshot();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid || submitting.value) return;
  if (form.timeRange.length !== 2 || form.timeRange[1] <= form.timeRange[0]) {
    message("比赛结束时间必须晚于开始时间", { type: "warning" });
    return;
  }
  if (!problems.value.length) {
    message("请至少添加一道题目", { type: "warning" });
    return;
  }
  if (!isEdit.value && form.requireInviteCode && !form.inviteCode.trim()) {
    message("启用邀请码后必须填写邀请码", { type: "warning" });
    return;
  }

  const [start, end] = form.timeRange;
  submitting.value = true;
  try {
    if (isEdit.value && props.contestId) {
      await updateContest(props.contestId, {
        title: form.name.trim(),
        description: form.description,
        coverUrl: form.coverUrl.trim(),
        ruleType: form.ruleType,
        startTime: start.getTime(),
        endTime: end.getTime(),
        problems: createProblemPayload(),
        rated: form.rated
      });
      message("比赛已更新", { type: "success" });
    } else {
      await createContest({
        name: form.name.trim(),
        contestDate: `${start.getFullYear()}-${pad(start.getMonth() + 1)}-${pad(start.getDate())}`,
        contestTime: `${pad(start.getHours())}:${pad(start.getMinutes())}`,
        description: form.description,
        coverUrl: form.coverUrl.trim(),
        duration: Math.round((end.getTime() - start.getTime()) / 60_000),
        ruleType: form.ruleType,
        problems: createProblemPayload(),
        inviteCode: form.requireInviteCode ? form.inviteCode.trim() : "",
        rated: form.rated
      });
      message("比赛已创建", { type: "success" });
    }
    saved.value = true;
    await router.push("/content/contests");
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    submitting.value = false;
  }
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
  <div v-loading="loading" class="contest-form-page">
    <div class="page-header">
      <div>
        <h2>{{ isEdit ? "编辑比赛" : "新建比赛" }}</h2>
        <p>设置比赛时间、赛制、题目顺序和计分方式。</p>
      </div>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <section class="form-section">
        <h3>基本信息</h3>
        <el-form-item label="比赛名称" prop="name">
          <el-input v-model="form.name" maxlength="60" show-word-limit />
        </el-form-item>
        <div class="settings-grid">
          <el-form-item label="比赛时间" prop="timeRange">
            <el-date-picker
              v-model="form.timeRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              class="w-full!"
            />
          </el-form-item>
          <el-form-item label="赛制" prop="ruleType">
            <el-select
              v-model="form.ruleType"
              class="w-full"
              :disabled="started"
            >
              <el-option
                v-for="item in RULE_OPTIONS"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
        <el-form-item label="封面 URL">
          <el-input
            v-model="form.coverUrl"
            placeholder="可选，不填写时使用默认封面"
          />
        </el-form-item>
        <div class="switch-row">
          <el-checkbox v-model="form.rated">计入 Rating</el-checkbox>
          <template v-if="!isEdit">
            <el-checkbox v-model="form.requireInviteCode">
              需要邀请码
            </el-checkbox>
            <el-input
              v-if="form.requireInviteCode"
              v-model="form.inviteCode"
              class="w-56!"
              maxlength="32"
              placeholder="请输入邀请码"
            />
          </template>
        </div>
      </section>

      <section class="form-section">
        <div class="section-heading">
          <h3>比赛题目</h3>
          <div class="flex gap-2">
            <el-input
              v-model="queryProblemID"
              class="w-40!"
              placeholder="输入题号"
              @keyup.enter="addByID"
            />
            <el-button :icon="Plus" @click="addByID">添加</el-button>
            <el-button type="primary" @click="pickerVisible = true">
              从题库选择
            </el-button>
          </div>
        </div>
        <el-empty
          v-if="!problems.length"
          description="请添加比赛题目"
          :image-size="72"
        />
        <el-table v-else :data="problems" border>
          <el-table-column type="index" label="#" width="60" align="center" />
          <el-table-column prop="id" label="题号" width="120" />
          <el-table-column prop="name" label="题目" min-width="220" />
          <el-table-column label="气球颜色" width="120" align="center">
            <template #default="{ row }">
              <el-color-picker v-model="row.color" :predefine="COLORS" />
            </template>
          </el-table-column>
          <el-table-column
            v-if="isScored"
            label="满分"
            width="130"
            align="center"
          >
            <template #default="{ row }">
              <el-input-number
                v-model="row.score"
                :min="1"
                :max="1000"
                :controls="false"
                class="w-20!"
              />
            </template>
          </el-table-column>
          <el-table-column label="排序" width="110" align="center">
            <template #default="{ $index }">
              <el-button
                link
                :icon="ArrowUp"
                :disabled="$index === 0"
                @click="moveProblem($index, -1)"
              />
              <el-button
                link
                :icon="ArrowDown"
                :disabled="$index === problems.length - 1"
                @click="moveProblem($index, 1)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="90" align="center">
            <template #default="{ $index }">
              <el-button link type="danger" @click="problems.splice($index, 1)">
                移除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="form-section">
        <h3>比赛说明</h3>
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="10"
          placeholder="支持 Markdown 原文"
        />
      </section>

      <div class="form-actions">
        <el-button @click="router.push('/content/contests')">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
          {{ isEdit ? "保存修改" : "创建比赛" }}
        </el-button>
      </div>
    </el-form>

    <ProblemPickerDialog
      v-model="pickerVisible"
      :existing-ids="existingIds"
      @add="addRows"
    />
  </div>
</template>

<style scoped>
.contest-form-page {
  padding: 24px;
  background: var(--el-bg-color);
  border-radius: 4px;
}

.page-header {
  padding-bottom: 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.page-header h2,
.form-section h3 {
  margin: 0;
}

.page-header p {
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
  grid-template-columns: minmax(0, 2fr) minmax(180px, 1fr);
  gap: 16px;
}

.switch-row,
.section-heading,
.form-actions {
  display: flex;
  align-items: center;
}

.switch-row {
  min-height: 32px;
  gap: 24px;
}

.section-heading {
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-heading h3 {
  margin-bottom: 0;
}

.form-actions {
  justify-content: flex-end;
  padding-top: 20px;
  margin-top: 24px;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
