<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRouter, onBeforeRouteLeave } from "vue-router";
import { ArrowLeft, Rank, View } from "@element-plus/icons-vue";
import { ElMessageBox } from "element-plus";
import { message } from "@/utils/message";
import { getProblem, getTagList, type TagItem } from "@/api/admin/problems";
import {
  createProblemSet,
  getProblemSetDetail,
  updateProblemSet,
  type SaveProblemSetParams
} from "@/api/admin/problemsets";
import ProblemPickerDialog from "./ProblemPickerDialog.vue";
import ProblemPreviewDialog from "./ProblemPreviewDialog.vue";

defineOptions({ name: "ProblemSetForm" });

const props = defineProps<{
  /** 编辑时传入题单对外 id；新建时不传 */
  problemSetId?: number;
}>();

interface ProblemSetFormModel {
  title: string;
  description: string;
  /** 0 草稿 / 1 已发布 */
  published: number;
  /** 0 公开 / 1 需邀请码 */
  visibility: number;
  inviteCode: string;
  tagIds: number[];
}

const router = useRouter();
const isEdit = computed(() => props.problemSetId !== undefined);
// 编辑回填时记录的原始可见性：0 公开 / 1 需邀请码
const originalVisibility = ref(0);
// 邀请码输入框提示：编辑原邀请码题单且留空时保留原码，其余场景需填写
const inviteCodePlaceholder = computed(() =>
  isEdit.value && originalVisibility.value === 1
    ? "留空则保留原邀请码"
    : "请输入邀请码"
);

const tagOptions = ref<TagItem[]>([]);
const problems = ref<{ id: string; name: string }[]>([]);
const form = reactive<ProblemSetFormModel>({
  title: "",
  description: "",
  published: 1,
  visibility: 0,
  inviteCode: "",
  tagIds: []
});

const loadingDetail = ref(false);
const saving = ref(false);
const dirty = ref(false);

// 除初次回填外，任何表单或题目变化都标记为未保存
watch(
  [form, problems],
  () => {
    if (!loadingDetail.value) dirty.value = true;
  },
  { deep: true, flush: "sync" }
);

const pidInput = ref("");
const addingPid = ref(false);
const pickerVisible = ref(false);
const previewVisible = ref(false);
const previewProblemId = ref("");
const existingIds = computed(() => problems.value.map(p => p.id));

const dragIndex = ref(-1);
const dragOverIndex = ref(-1);

async function loadTags() {
  try {
    const res = await getTagList();
    tagOptions.value = res.data?.tags ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  }
}

async function loadDetail() {
  if (props.problemSetId === undefined) return;
  loadingDetail.value = true;
  try {
    const res = await getProblemSetDetail(props.problemSetId);
    const d = res.data;
    if (!d) return;
    form.title = d.title;
    form.description = d.description ?? "";
    form.published = d.published;
    form.visibility = d.visibility;
    form.inviteCode = "";
    originalVisibility.value = d.visibility;
    form.tagIds = (d.tags ?? []).map(t => t.id);
    problems.value = (d.problems ?? []).map(p => ({ id: p.id, name: p.name }));
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loadingDetail.value = false;
    dirty.value = false;
  }
}

async function addProblemById() {
  const id = pidInput.value.trim();
  if (!id) {
    message("请输入题号", { type: "warning" });
    return;
  }
  if (problems.value.some(p => p.id === id)) {
    message("题目已添加", { type: "warning" });
    return;
  }
  addingPid.value = true;
  try {
    const res = await getProblem(id);
    const problem = res.data;
    if (!problem) return;
    problems.value.push({ id: problem.id, name: problem.name });
    pidInput.value = "";
  } catch {
    // 错误信息已由 http 拦截器统一提示（如题目不存在）
  } finally {
    addingPid.value = false;
  }
}

function removeProblem(index: number) {
  problems.value.splice(index, 1);
}

function openProblemPreview(problemId: string) {
  previewProblemId.value = problemId;
  previewVisible.value = true;
}

function addPickedProblems(items: { id: string; name: string }[]) {
  items.forEach(item => {
    if (!problems.value.some(p => p.id === item.id)) {
      problems.value.push({ id: item.id, name: item.name });
    }
  });
}

function onDragStart(index: number) {
  dragIndex.value = index;
}

function onDragOver(index: number) {
  dragOverIndex.value = index;
}

function onDragEnd() {
  dragIndex.value = -1;
  dragOverIndex.value = -1;
}

function onDrop(index: number) {
  const from = dragIndex.value;
  if (from !== -1 && from !== index) {
    const [moved] = problems.value.splice(from, 1);
    problems.value.splice(index, 0, moved);
  }
  onDragEnd();
}

async function save() {
  if (!form.title.trim()) {
    message("请填写标题", { type: "warning" });
    return;
  }
  // 邀请码可见时必须填写；仅编辑原邀请码题单时留空表示保留原邀请码
  if (
    form.visibility === 1 &&
    (!isEdit.value || originalVisibility.value === 0) &&
    !form.inviteCode.trim()
  ) {
    message("选择邀请码可见时必须填写邀请码", { type: "warning" });
    return;
  }
  const payload: SaveProblemSetParams = {
    title: form.title.trim(),
    description: form.description,
    published: form.published,
    visibility: form.visibility,
    inviteCode: form.inviteCode.trim(),
    tagIds: form.tagIds,
    problems: problems.value.map(p => p.id)
  };
  saving.value = true;
  try {
    if (isEdit.value && props.problemSetId !== undefined) {
      await updateProblemSet(props.problemSetId, payload);
    } else {
      await createProblemSet(payload);
    }
    message("已保存", { type: "success" });
    dirty.value = false;
    router.push("/content/problemsets");
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    saving.value = false;
  }
}

function goBack() {
  router.push("/content/problemsets");
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true;
  try {
    await ElMessageBox.confirm("当前修改尚未保存，确定要离开吗？", "提示", {
      type: "warning"
    });
    return true;
  } catch {
    return false;
  }
});

onMounted(() => {
  loadTags();
  loadDetail();
});
</script>

<template>
  <div class="problem-set-editor">
    <el-card v-loading="loadingDetail" class="editor-card" shadow="never">
      <header class="editor-header">
        <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          保存题单
        </el-button>
      </header>

      <el-form :model="form" label-position="top">
        <section class="editor-section">
          <div class="basic-grid">
            <el-form-item label="标题" required>
              <el-input
                v-model="form.title"
                maxlength="100"
                show-word-limit
                placeholder="请输入题单标题"
              />
            </el-form-item>

            <el-form-item label="标签">
              <el-select
                v-model="form.tagIds"
                multiple
                filterable
                clearable
                placeholder="选择标签（可多选）"
                class="field-full"
              >
                <el-option
                  v-for="tag in tagOptions"
                  :key="tag.id"
                  :label="tag.name"
                  :value="tag.id"
                />
              </el-select>
            </el-form-item>
          </div>
        </section>

        <section class="editor-section">
          <div class="setting-grid">
            <el-form-item label="访问方式">
              <el-radio-group v-model="form.visibility">
                <el-radio :value="0">公开</el-radio>
                <el-radio :value="1">需要邀请码</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="发布状态">
              <el-radio-group v-model="form.published">
                <el-radio :value="1">已发布</el-radio>
                <el-radio :value="0">草稿</el-radio>
              </el-radio-group>
              <p class="field-hint">草稿仅管理员可见</p>
            </el-form-item>

            <el-form-item
              v-if="form.visibility === 1"
              class="grid-full"
              label="邀请码"
            >
              <el-input
                v-model="form.inviteCode"
                maxlength="32"
                :placeholder="inviteCodePlaceholder"
                class="invite-code-input"
              />
            </el-form-item>
          </div>
        </section>

        <section class="editor-section problem-section">
          <div class="section-heading problem-heading">
            <div>
              <div class="section-title-row">
                <h2 class="section-title">题目列表</h2>
                <span class="problem-count">{{ problems.length }} 道</span>
              </div>
            </div>
          </div>

          <div class="problem-toolbar">
            <el-input
              v-model="pidInput"
              class="problem-id-input"
              placeholder="输入题号，如 P1001"
              :disabled="addingPid"
              @keyup.enter="addProblemById"
            />
            <el-button :loading="addingPid" @click="addProblemById">
              添加题号
            </el-button>
            <el-button type="primary" @click="pickerVisible = true">
              从题库选择
            </el-button>
          </div>

          <div v-if="problems.length" class="problem-table">
            <div class="problem-table-header">
              <span />
              <span>序号</span>
              <span>题号</span>
              <span>题目名称</span>
              <span>操作</span>
            </div>
            <div class="problem-table-body">
              <div
                v-for="(problem, index) in problems"
                :key="problem.id"
                class="problem-row"
                :class="{
                  'is-drag-over':
                    dragOverIndex === index && dragIndex !== index,
                  'is-dragging': dragIndex === index
                }"
                draggable="true"
                @dragstart="onDragStart(index)"
                @dragover.prevent="onDragOver(index)"
                @dragend="onDragEnd"
                @drop="onDrop(index)"
              >
                <el-icon class="drag-handle"><Rank /></el-icon>
                <span class="problem-index">{{ index + 1 }}</span>
                <span class="problem-id">{{ problem.id }}</span>
                <span class="problem-name">{{ problem.name }}</span>
                <div class="problem-actions">
                  <el-tooltip content="查看题目" placement="top">
                    <el-button
                      link
                      type="primary"
                      :icon="View"
                      aria-label="查看题目"
                      @click="openProblemPreview(problem.id)"
                    />
                  </el-tooltip>
                  <el-button link type="danger" @click="removeProblem(index)">
                    移除
                  </el-button>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="problem-empty">
            <el-empty description="还没有添加题目" :image-size="72">
              <el-button type="primary" plain @click="pickerVisible = true">
                从题库选择题目
              </el-button>
            </el-empty>
          </div>
        </section>

        <section class="editor-section">
          <el-form-item label="描述">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="6"
              resize="vertical"
              placeholder="介绍训练目标、适合人群或建议完成顺序（支持 Markdown）"
            />
          </el-form-item>
        </section>
      </el-form>

      <div class="editor-actions">
        <el-button :disabled="saving" @click="goBack">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          保存题单
        </el-button>
      </div>
    </el-card>

    <ProblemPickerDialog
      v-model="pickerVisible"
      :existing-ids="existingIds"
      @add="addPickedProblems"
    />
    <ProblemPreviewDialog
      v-model="previewVisible"
      :problem-id="previewProblemId"
    />
  </div>
</template>

<style scoped>
.problem-set-editor {
  min-width: 0;
}

.editor-header {
  display: flex;
  gap: 16px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.section-title,
.section-description,
.field-hint {
  margin: 0;
}

.section-description,
.field-hint {
  font-size: 13px;
  line-height: 20px;
  color: var(--el-text-color-secondary);
}

.editor-card {
  border-color: var(--el-border-color-light);
}

.editor-card :deep(.el-card__body) {
  padding: 24px 28px 28px;
}

.editor-section {
  width: 100%;
}

.editor-section + .editor-section {
  margin-top: 36px;
}

.section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 18px;
}

.section-title-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  color: var(--el-text-color-primary);
}

.basic-grid,
.setting-grid {
  display: grid;
  gap: 24px;
}

.basic-grid :deep(.el-form-item),
.setting-grid :deep(.el-form-item) {
  margin-bottom: 0;
}

.basic-grid {
  grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.85fr);
}

.setting-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.grid-full {
  grid-column: 1 / -1;
}

.field-full {
  width: 100%;
}

.invite-code-input {
  width: min(360px, 100%);
}

.field-hint {
  margin-left: 12px;
}

.problem-count {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  font-size: 12px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 4px;
}

.problem-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 14px;
}

.problem-id-input {
  width: 240px;
}

.problem-table,
.problem-empty {
  overflow: hidden;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
}

.problem-table-header,
.problem-row {
  display: grid;
  grid-template-columns: 42px 64px 150px minmax(220px, 1fr) 120px;
  align-items: center;
  min-width: 700px;
}

.problem-table-header {
  height: 42px;
  padding: 0 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-light);
}

.problem-table-body {
  max-height: 360px;
  overflow: auto;
}

.problem-row {
  min-height: 58px;
  padding: 0 14px;
  font-size: 15px;
  color: var(--el-text-color-regular);
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition:
    background-color 0.16s ease,
    opacity 0.16s ease;
}

.problem-row:last-child {
  border-bottom: 0;
}

.problem-row:hover,
.problem-row.is-drag-over {
  background: var(--el-fill-color-light);
}

.problem-row.is-dragging {
  opacity: 0.55;
}

.drag-handle {
  color: var(--el-text-color-placeholder);
  cursor: grab;
}

.drag-handle:active {
  cursor: grabbing;
}

.problem-index {
  color: var(--el-text-color-secondary);
}

.problem-id {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--el-color-primary);
}

.problem-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.problem-table-header > :last-child {
  text-align: center;
}

.problem-actions {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: center;
}

.problem-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.problem-empty {
  min-height: 220px;
}

.editor-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-top: 24px;
}

@media (width <= 768px) {
  .editor-header {
    align-items: flex-start;
  }

  .editor-card :deep(.el-card__body) {
    padding: 20px 16px;
  }

  .basic-grid,
  .setting-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .grid-full {
    grid-column: auto;
  }

  .field-hint {
    display: block;
    width: 100%;
    margin: 2px 0 0;
  }

  .problem-id-input {
    width: 100%;
  }

  .problem-table {
    overflow-x: auto;
  }
}
</style>
