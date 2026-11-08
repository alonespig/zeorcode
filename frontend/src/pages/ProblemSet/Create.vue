<template>
  <div class="create-problemset-page page-container">
    <section class="editor-panel f-panel">
      <header class="page-header">
        <el-button :icon="ArrowLeft" @click="cancel">返回</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </header>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <section class="form-section">
          <div class="basic-grid">
            <el-form-item label="题单名称" prop="title">
              <el-input v-model="form.title" maxlength="100" show-word-limit placeholder="请输入题单名称" />
            </el-form-item>
            <el-form-item label="算法标签">
              <el-select v-model="form.tagIds" multiple filterable collapse-tags collapse-tags-tooltip
                placeholder="选择标签" class="w-full">
                <el-option v-for="tag in tags" :key="tag.id" :label="tag.name" :value="tag.id" />
              </el-select>
            </el-form-item>
          </div>
        </section>

        <section class="form-section">
          <div class="setting-grid">
            <el-form-item label="访问方式">
              <el-radio-group v-model="form.visibility">
                <el-radio :value="0">公开</el-radio>
                <el-radio :value="1">邀请码可见</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="发布状态">
              <el-radio-group v-model="form.published">
                <el-radio :value="1">已发布</el-radio>
                <el-radio :value="0">草稿</el-radio>
              </el-radio-group>
              <span class="field-hint publish-hint">草稿仅管理员可见</span>
            </el-form-item>
          </div>
          <el-form-item v-if="form.visibility === 1" label="邀请码" class="invite-item">
            <el-input v-model="form.inviteCode" maxlength="32" placeholder="请输入邀请码" />
          </el-form-item>
        </section>

        <section class="form-section">
          <div class="section-head">
            <div class="section-title-row">
              <h2 class="section-title">题目列表</h2>
              <span class="problem-count">{{ problemList.length }} 道</span>
            </div>
            <span class="field-hint">拖动行可调整顺序</span>
          </div>
          <div class="problem-toolbar">
            <el-input v-model="queryPid" placeholder="输入题号，如 L101" clearable @keyup.enter="addProblem"
              class="pid-input" />
            <el-button :icon="Plus" @click="addProblem">添加</el-button>
            <el-button type="primary" plain :icon="Search" @click="pickerVisible = true">从题库选择</el-button>
          </div>

          <div v-if="problemList.length" class="problem-table-wrap">
            <table class="f-table problem-table">
              <colgroup>
                <col style="width: 52px">
                <col style="width: 70px">
                <col style="width: 140px">
                <col>
                <col style="width: 80px">
              </colgroup>
              <thead>
                <tr>
                  <th></th>
                  <th class="center">#</th>
                  <th>题号</th>
                  <th>题目</th>
                  <th class="center">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(problem, index) in problemList" :key="problem.id" draggable="true"
                  :class="{ dragging: dragIndex === index, 'drag-over': dragOverIndex === index }"
                  @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)" @dragend="onDragEnd"
                  @drop="onDrop(index)">
                  <td class="center drag-handle" title="拖动排序">
                    <el-icon><Rank /></el-icon>
                  </td>
                  <td class="center">{{ index + 1 }}</td>
                  <td class="problem-id">{{ problem.id }}</td>
                  <td>{{ problem.name }}</td>
                  <td class="center">
                    <el-button link type="danger" @click="removeProblem(index)">移除</el-button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <el-empty v-else description="还没有添加题目" :image-size="72" />
        </section>

        <section class="form-section">
          <el-form-item label="题单说明">
            <MdEditor v-model="form.description" height="260px" editor-id="problemset-create-editor" />
          </el-form-item>
        </section>
      </el-form>

      <div class="form-actions">
        <el-button @click="cancel">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </div>
    </section>

    <ProblemPickerDialog v-model="pickerVisible" :existing-problem-ids="existingProblemIds"
      @add="addPickedProblems" />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { onBeforeRouteLeave, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { ArrowLeft, Plus, Rank, Search } from "@element-plus/icons-vue";
import { createProblemSet } from "@/api/problemset";
import { getProblem, getTags } from "@/api/problems";
import MdEditor from "@/components/MdEditor.vue";
import ProblemPickerDialog from "@/components/problem/ProblemPickerDialog.vue";

const router = useRouter();
const formRef = ref(null);
const tags = ref([]);
const queryPid = ref("");
const problemList = ref([]);
const pickerVisible = ref(false);
const submitting = ref(false);
const leavingAfterSave = ref(false);

const form = ref({
  title: "",
  description: "",
  visibility: 0,
  inviteCode: "",
  tagIds: [],
  published: 1,
});

const rules = {
  title: [{ required: true, message: "请输入题单名称", trigger: "blur" }],
};

const existingProblemIds = computed(() => problemList.value.map((problem) => problem.id));
const hasChanges = computed(() => Boolean(
  form.value.title.trim()
  || form.value.description.trim()
  || form.value.tagIds.length
  || form.value.visibility === 1
  || form.value.published !== 1
  || problemList.value.length
));

const addProblem = async () => {
  const id = queryPid.value.trim();
  if (!id) {
    ElMessage.warning("请输入题号");
    return;
  }
  if (problemList.value.some((problem) => String(problem.id) === id)) {
    ElMessage.warning("该题目已经在题单中");
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
  problems.forEach((problem) => {
    if (!problemList.value.some((item) => String(item.id) === String(problem.id))) {
      problemList.value.push({ id: problem.id, name: problem.name });
    }
  });
};

const removeProblem = (index) => problemList.value.splice(index, 1);

const dragIndex = ref(-1);
const dragOverIndex = ref(-1);
const onDragStart = (index) => { dragIndex.value = index; };
const onDragOver = (index) => { dragOverIndex.value = index; };
const onDragEnd = () => { dragIndex.value = -1; dragOverIndex.value = -1; };
const onDrop = (index) => {
  const from = dragIndex.value;
  if (from !== -1 && from !== index) {
    const [problem] = problemList.value.splice(from, 1);
    problemList.value.splice(index, 0, problem);
  }
  onDragEnd();
};

const submit = async () => {
  try {
    await formRef.value.validate();
  } catch {
    return;
  }
  if (form.value.visibility === 1 && !form.value.inviteCode.trim()) {
    ElMessage.warning("邀请码可见的题单必须填写邀请码");
    return;
  }

  submitting.value = true;
  try {
    const res = await createProblemSet({
      title: form.value.title.trim(),
      description: form.value.description,
      published: form.value.published,
      visibility: form.value.visibility,
      inviteCode: form.value.inviteCode.trim(),
      tagIds: form.value.tagIds,
      problems: problemList.value.map((problem) => String(problem.id)),
    });
    leavingAfterSave.value = true;
    ElMessage.success(form.value.published === 1 ? "题单已创建并发布" : "题单已保存为草稿");
    if (form.value.published === 1) {
      await router.push({ name: "ProblemSetIntro", params: { id: res.data.id } });
    } else {
      await router.push({ name: "ProblemSetList" });
    }
  } catch (err) {
    console.error(err);
  } finally {
    submitting.value = false;
  }
};

const cancel = () => router.push({ name: "ProblemSetList" });

onBeforeRouteLeave(async () => {
  if (leavingAfterSave.value || !hasChanges.value) return true;
  try {
    await ElMessageBox.confirm("当前题单尚未保存，确定离开吗？", "离开创建页面", {
      type: "warning",
      confirmButtonText: "离开",
      cancelButtonText: "继续编辑",
    });
    return true;
  } catch {
    return false;
  }
});

onMounted(async () => {
  try {
    const res = await getTags();
    tags.value = res.data?.tags || [];
  } catch (err) {
    console.error(err);
  }
});
</script>

<style scoped lang="scss">
.create-problemset-page {
  padding-top: 8px;
  padding-bottom: 28px;
}

.editor-panel {
  overflow: hidden;
  padding: 24px 28px 28px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.form-section + .form-section {
  margin-top: 36px;
}

.section-title {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
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

.setting-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
}

.setting-grid :deep(.el-form-item) {
  margin-bottom: 18px;
}

.invite-item {
  margin-bottom: 0;
}

.field-hint {
  display: block;
  margin-top: 2px;
  color: var(--el-text-color-secondary);
  font-size: 12.5px;
}

.publish-hint {
  width: 100%;
}

.w-full {
  width: 100%;
}

.problem-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.pid-input {
  width: 220px;
}

.problem-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--table-border);
  border-radius: var(--radius-base);
}

.problem-table {
  min-width: 620px;

  th,
  td {
    vertical-align: middle;
  }

  .center {
    text-align: center;
  }

  .problem-id {
    font-variant-numeric: tabular-nums;
  }

  .drag-handle {
    color: var(--el-text-color-placeholder);
    cursor: grab;
  }

  tr.dragging {
    opacity: 0.45;
  }

  tr.drag-over {
    background: var(--table-row-hover);
    box-shadow: inset 0 2px 0 var(--color-primary);
  }
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-top: 24px;

  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}

@media (max-width: 820px) {
  .editor-panel {
    padding: 20px 16px;
  }

  .basic-grid,
  .setting-grid {
    grid-template-columns: minmax(0, 1fr);
    gap: 0;
  }

  .setting-grid :deep(.el-form-item) {
    margin-bottom: 18px;
  }

  .problem-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .pid-input {
    width: 100%;
  }
}
</style>
