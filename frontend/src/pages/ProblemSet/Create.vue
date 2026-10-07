<template>
  <div class="create-problemset-page page-container">
    <section class="editor-panel f-panel">
      <header class="page-header">
        <div class="page-title">
          <el-button :icon="ArrowLeft" text aria-label="返回题单列表" @click="cancel" />
          <div>
            <h1>创建题单</h1>
          </div>
        </div>
      </header>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <section class="form-section">
          <h2>基本信息</h2>
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

          <el-form-item label="访问方式">
            <div class="access-setting">
              <el-radio-group v-model="form.visibility">
                <el-radio :value="0">公开</el-radio>
                <el-radio :value="1">邀请码可见</el-radio>
              </el-radio-group>
              <el-input v-if="form.visibility === 1" v-model="form.inviteCode" maxlength="32"
                placeholder="请输入邀请码" class="invite-input" />
            </div>
          </el-form-item>
        </section>

        <section class="form-section">
          <div class="section-header">
            <div>
              <h2>题目配置</h2>
              <p>题目将按照列表顺序展示，可拖动调整顺序</p>
            </div>
            <div class="problem-actions">
              <el-input v-model="queryPid" placeholder="输入题号，如 L101" clearable @keyup.enter="addProblem" />
              <el-button :icon="Plus" @click="addProblem">添加</el-button>
              <el-button type="primary" plain :icon="Search" @click="pickerVisible = true">从题库选择</el-button>
            </div>
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
          <h2>题单说明</h2>
          <MdEditor v-model="form.description" height="260px" editor-id="problemset-create-editor" />
        </section>

        <footer class="form-actions">
          <el-button @click="cancel">取消</el-button>
          <el-button :loading="submitting" @click="submit(0)">保存草稿</el-button>
          <el-button type="primary" :loading="submitting" @click="submit(1)">创建并发布</el-button>
        </footer>
      </el-form>
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

const submit = async (published) => {
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
      published,
      visibility: form.value.visibility,
      inviteCode: form.value.inviteCode.trim(),
      tagIds: form.value.tagIds,
      problems: problemList.value.map((problem) => String(problem.id)),
    });
    leavingAfterSave.value = true;
    ElMessage.success(published === 1 ? "题单已创建并发布" : "题单已保存为草稿");
    if (published === 1) {
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
}

.page-header {
  padding: 18px 24px;
  border-bottom: 1px solid var(--color-border-soft);
}

.page-title {
  display: flex;
  align-items: center;
  gap: 8px;

  h1 {
    margin: 0;
    color: var(--el-text-color-primary);
    font-size: 21px;
    font-weight: 600;
  }

}

.form-section {
  padding: 22px 24px;
  border-bottom: 1px solid var(--color-border-soft);

  h2 {
    margin: 0 0 16px;
    color: var(--el-text-color-primary);
    font-size: 16px;
    font-weight: 600;
  }
}

.basic-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(280px, 1fr);
  gap: 20px;
}

.w-full {
  width: 100%;
}

.access-setting {
  display: flex;
  align-items: center;
  gap: 20px;
}

.invite-input {
  width: 260px;
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 14px;

  h2 {
    margin-bottom: 4px;
  }

  p {
    margin: 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
}

.problem-actions {
  display: flex;
  align-items: center;
  gap: 8px;

  .el-input {
    width: 210px;
  }
}

.problem-table-wrap {
  overflow-x: auto;
  border: 1px solid var(--table-border);
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
  justify-content: flex-end;
  gap: 10px;
  padding: 18px 24px;

  :deep(.el-button + .el-button) {
    margin-left: 0;
  }
}

@media (max-width: 820px) {
  .basic-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .section-header,
  .problem-actions,
  .access-setting {
    align-items: stretch;
    flex-direction: column;
  }

  .problem-actions .el-input,
  .invite-input {
    width: 100%;
  }
}
</style>
