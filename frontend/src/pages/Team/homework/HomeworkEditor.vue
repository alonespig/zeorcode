<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { ArrowLeft, Plus, Rank, View } from "@element-plus/icons-vue";
import { createHomework, getHomeworkDetail, updateHomework } from "@/api/team";
import { getProblem } from "@/api/problems";
import ProblemPickerDialog from "@/components/problem/ProblemPickerDialog.vue";
import ProblemPreviewDialog from "@/components/problem/ProblemPreviewDialog.vue";
import ProblemSetPickerDialog from "@/components/problem/ProblemSetPickerDialog.vue";

const route = useRoute();
const router = useRouter();

const isEdit = computed(() => Boolean(route.params.hid));
const loading = ref(false);
const saving = ref(false);
const pickerVisible = ref(false);
const problemSetPickerVisible = ref(false);
const previewVisible = ref(false);
const previewProblemId = ref("");
const queryPid = ref("");
const problemList = ref([]);
const form = ref(blankForm());

const existingProblemIds = computed(() => problemList.value.map((problem) => problem.id));

function blankForm() {
  return { title: "", description: "", range: [] };
}

const toMinutePrecision = (value) => `${value.slice(0, 16)}:00`;

const backToList = () => {
  router.push({ name: "TeamHomework", params: { id: route.params.id } });
};

const loadEditor = async () => {
  form.value = blankForm();
  problemList.value = [];
  queryPid.value = "";
  if (!isEdit.value) return;

  loading.value = true;
  try {
    const res = await getHomeworkDetail(route.params.hid);
    const detail = res.data;
    form.value = {
      title: detail.title,
      description: detail.description || "",
      range: [detail.startTime, detail.endTime],
    };
    problemList.value = (detail.problems || []).map((problem) => ({
      id: problem.id,
      name: problem.name,
    }));
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const addProblem = async () => {
  const id = queryPid.value.trim();
  if (!id) {
    ElMessage.warning("请输入题号");
    return;
  }
  if (problemList.value.some((problem) => String(problem.id) === id)) {
    ElMessage.warning("题目已添加");
    return;
  }
  try {
    const res = await getProblem(id);
    problemList.value.push({ id: res.data.id, name: res.data.name });
    queryPid.value = "";
  } catch (err) {
    console.error(err);
  }
};

const addPickedProblems = (problems) => {
  const existing = new Set(problemList.value.map((problem) => String(problem.id)));
  problems.forEach((problem) => {
    if (existing.has(String(problem.id))) return;
    problemList.value.push({ id: problem.id, name: problem.name });
    existing.add(String(problem.id));
  });
};

const removeProblem = (index) => problemList.value.splice(index, 1);

const openPreview = (id) => {
  previewProblemId.value = String(id);
  previewVisible.value = true;
};

const dragIndex = ref(-1);
const dragOverIndex = ref(-1);

const onDragStart = (index) => {
  dragIndex.value = index;
};

const onDragOver = (index) => {
  dragOverIndex.value = index;
};

const onDragEnd = () => {
  dragIndex.value = -1;
  dragOverIndex.value = -1;
};

const onDrop = (index) => {
  const from = dragIndex.value;
  if (from !== -1 && from !== index) {
    const [moved] = problemList.value.splice(from, 1);
    problemList.value.splice(index, 0, moved);
  }
  onDragEnd();
};

const save = async () => {
  if (!form.value.title.trim()) {
    ElMessage.warning("请填写作业名称");
    return;
  }
  if (!form.value.range || form.value.range.length !== 2) {
    ElMessage.warning("请选择起止时间");
    return;
  }

  const payload = {
    title: form.value.title.trim(),
    description: form.value.description,
    startTime: toMinutePrecision(form.value.range[0]),
    endTime: toMinutePrecision(form.value.range[1]),
    problems: problemList.value.map((problem) => String(problem.id)),
  };

  saving.value = true;
  try {
    if (isEdit.value) {
      await updateHomework(route.params.hid, payload);
      ElMessage.success("作业已更新");
    } else {
      await createHomework(route.params.id, payload);
      ElMessage.success("作业已布置");
    }
    backToList();
  } catch (err) {
    console.error(err);
  } finally {
    saving.value = false;
  }
};

watch(() => route.params.hid, loadEditor, { immediate: true });
</script>

<template>
  <div v-loading="loading" class="homework-editor-page">
    <section class="f-panel editor-shell">
      <header class="editor-header">
        <el-button :icon="ArrowLeft" @click="backToList">返回</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{ isEdit ? "保存修改" : "布置作业" }}
        </el-button>
      </header>

      <el-form :model="form" label-position="top" class="editor-form">
        <section class="editor-section">
          <div class="section-heading">
            <h2 class="section-title">基本信息</h2>
          </div>
          <div class="basic-grid">
            <el-form-item label="作业名称" required>
              <el-input v-model="form.title" maxlength="100" show-word-limit placeholder="请输入作业名称" />
            </el-form-item>
            <el-form-item label="起止时间" required>
              <el-date-picker v-model="form.range" type="datetimerange" range-separator="至"
                start-placeholder="开始时间" end-placeholder="结束时间" format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DD HH:mm:ss" />
            </el-form-item>
          </div>
        </section>

        <section class="editor-section">
          <div class="section-heading">
            <div class="section-title-row">
              <h2 class="section-title">题目列表</h2>
              <span class="problem-count">{{ problemList.length }} 道</span>
            </div>
          </div>

          <div class="problem-toolbar">
            <el-input v-model="queryPid" placeholder="输入题号，如 L101" class="problem-id-input"
              @keyup.enter="addProblem" />
            <el-button :icon="Plus" @click="addProblem">添加</el-button>
            <el-button type="primary" @click="pickerVisible = true">从题库选择</el-button>
            <el-button @click="problemSetPickerVisible = true">从题单导入</el-button>
          </div>

          <div v-if="problemList.length" class="problem-table">
            <div class="problem-table-header">
              <span />
              <span>序号</span>
              <span>题号</span>
              <span>题目名称</span>
              <span>操作</span>
            </div>
            <div class="problem-table-body">
              <div v-for="(problem, index) in problemList" :key="problem.id" class="problem-row" :class="{
                'is-drag-over': dragOverIndex === index && dragIndex !== index,
                'is-dragging': dragIndex === index,
              }" draggable="true" @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)"
                @dragend="onDragEnd" @drop="onDrop(index)">
                <el-icon class="drag-handle"><Rank /></el-icon>
                <span class="problem-index">{{ index + 1 }}</span>
                <span class="problem-id">{{ problem.id }}</span>
                <span class="problem-name">{{ problem.name }}</span>
                <div class="problem-actions">
                  <el-tooltip content="查看题目" placement="top">
                    <el-button link type="primary" :icon="View" aria-label="查看题目"
                      @click="openPreview(problem.id)" />
                  </el-tooltip>
                  <el-button link type="danger" @click="removeProblem(index)">移除</el-button>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="problem-empty">
            <el-empty description="还没有添加题目" :image-size="72">
              <el-button type="primary" plain @click="pickerVisible = true">从题库选择</el-button>
              <el-button type="primary" plain @click="problemSetPickerVisible = true">从题单导入</el-button>
            </el-empty>
          </div>
        </section>

        <section class="editor-section">
          <div class="section-heading">
            <h2 class="section-title">作业说明</h2>
          </div>
          <el-form-item class="description-field">
            <MdEditor v-model="form.description" height="280px" editor-id="homework-page-editor" />
          </el-form-item>
        </section>
      </el-form>

      <div class="editor-actions">
        <el-button @click="backToList">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{ isEdit ? "保存修改" : "布置作业" }}
        </el-button>
      </div>
    </section>

    <ProblemPickerDialog v-model="pickerVisible" :existing-problem-ids="existingProblemIds"
      @add="addPickedProblems" />
    <ProblemSetPickerDialog v-model="problemSetPickerVisible" :existing-problem-ids="existingProblemIds"
      @add="addPickedProblems" />
    <ProblemPreviewDialog v-model="previewVisible" :problem-id="previewProblemId" />
  </div>
</template>

<style scoped>
.homework-editor-page {
  min-width: 0;
}

.editor-shell {
  padding: 24px 28px 28px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.editor-section + .editor-section {
  margin-top: 36px;
}

.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}

.section-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.section-title {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.problem-count {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 4px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  font-size: 12px;
}

.basic-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.85fr);
  gap: 24px;
}

.basic-grid :deep(.el-form-item) {
  margin-bottom: 0;
}

.basic-grid :deep(.el-date-editor) {
  width: 100%;
}

.description-field {
  margin-bottom: 0;
}

.problem-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
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
  border-bottom: 1px solid var(--el-border-color-light);
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  font-size: 14px;
  font-weight: 600;
}

.problem-table-body {
  max-height: 360px;
  overflow: auto;
}

.problem-row {
  min-height: 58px;
  padding: 0 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  color: var(--el-text-color-regular);
  font-size: 15px;
  transition: background-color 0.16s ease, opacity 0.16s ease;
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
  color: var(--el-color-primary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
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
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.problem-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.problem-empty {
  min-height: 220px;
}

.editor-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 24px;
}

@media (max-width: 768px) {
  .editor-shell {
    padding: 20px 16px;
  }

  .basic-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .problem-id-input {
    width: 100%;
  }

  .problem-table {
    overflow-x: auto;
  }
}
</style>
