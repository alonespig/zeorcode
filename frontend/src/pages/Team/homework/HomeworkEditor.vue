<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { ArrowDown, ArrowLeft, ArrowUp, Plus } from "@element-plus/icons-vue";
import { createHomework, getHomeworkDetail, updateHomework } from "@/api/team";
import { getProblem } from "@/api/problems";
import ProblemPickerDialog from "@/components/problem/ProblemPickerDialog.vue";
import ProblemSetPickerDialog from "@/components/problem/ProblemSetPickerDialog.vue";

const route = useRoute();
const router = useRouter();

const isEdit = computed(() => Boolean(route.params.hid));
const pageTitle = computed(() => isEdit.value ? "编辑作业" : "布置作业");
const loading = ref(false);
const saving = ref(false);
const pickerVisible = ref(false);
const problemSetPickerVisible = ref(false);
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

const moveProblem = (index, offset) => {
  const target = index + offset;
  if (target < 0 || target >= problemList.value.length) return;
  const [problem] = problemList.value.splice(index, 1);
  problemList.value.splice(target, 0, problem);
};

const removeProblem = (index) => problemList.value.splice(index, 1);

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
      <header class="editor-heading">
        <el-button text :icon="ArrowLeft" class="back-button" @click="backToList">返回作业列表</el-button>
        <h2>{{ pageTitle }}</h2>
      </header>

      <el-form :model="form" label-position="top" class="editor-form">
        <section class="editor-section">
          <h3 class="section-title">基本信息</h3>
          <div class="basic-fields">
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

        <section class="editor-section problem-section">
          <h3 class="section-title">题目配置</h3>
          <div class="problem-toolbar">
            <el-input v-model="queryPid" placeholder="输入题号，如 L101" class="problem-id-input"
              @keyup.enter="addProblem" />
            <el-button :icon="Plus" @click="addProblem">添加</el-button>
            <el-button type="primary" plain @click="pickerVisible = true">从题库选择</el-button>
            <el-button type="primary" plain @click="problemSetPickerVisible = true">从题单导入</el-button>
          </div>

          <el-table v-if="problemList.length" :data="problemList" class="problem-table">
            <el-table-column label="#" width="64" align="center">
              <template #default="{ $index }">{{ $index + 1 }}</template>
            </el-table-column>
            <el-table-column prop="id" label="题号" width="130" align="center" />
            <el-table-column label="题目" min-width="240">
              <template #default="{ row }">
                <span class="font-medium text-gray-800">{{ row.name }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="170" align="center">
              <template #default="{ $index }">
                <div class="problem-actions">
                  <el-tooltip content="上移" placement="top">
                    <el-button link :icon="ArrowUp" :disabled="$index === 0" aria-label="上移题目"
                      @click="moveProblem($index, -1)" />
                  </el-tooltip>
                  <el-tooltip content="下移" placement="top">
                    <el-button link :icon="ArrowDown" :disabled="$index === problemList.length - 1"
                      aria-label="下移题目" @click="moveProblem($index, 1)" />
                  </el-tooltip>
                  <span class="problem-actions__divider"></span>
                  <el-button link type="danger" @click="removeProblem($index)">移除</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="还没有添加题目" :image-size="64">
            <el-button type="primary" plain @click="pickerVisible = true">从题库选择</el-button>
            <el-button type="primary" plain @click="problemSetPickerVisible = true">从题单导入</el-button>
          </el-empty>
        </section>

        <section class="editor-section">
          <h3 class="section-title">作业说明</h3>
          <el-form-item class="description-field">
            <MdEditor v-model="form.description" height="280px" editor-id="homework-page-editor" />
          </el-form-item>
        </section>

        <div class="editor-actions">
          <el-button @click="backToList">取消</el-button>
          <el-button type="primary" :loading="saving" @click="save">
            {{ isEdit ? "保存修改" : "布置作业" }}
          </el-button>
        </div>
      </el-form>
    </section>

    <ProblemPickerDialog v-model="pickerVisible" :existing-problem-ids="existingProblemIds"
      @add="addPickedProblems" />
    <ProblemSetPickerDialog v-model="problemSetPickerVisible" :existing-problem-ids="existingProblemIds"
      @add="addPickedProblems" />
  </div>
</template>

<style scoped>
.homework-editor-page {
  min-width: 0;
}

.editor-shell {
  overflow: hidden;
}

.editor-heading {
  padding: 16px 22px 20px;
  border-bottom: 1px solid #e8edf3;
}

.back-button {
  margin-left: -10px;
  color: #64748b;
}

.editor-heading h2 {
  margin-top: 8px;
  color: #172033;
  font-size: 22px;
  font-weight: 650;
  line-height: 1.35;
}

.editor-form {
  display: block;
}

.editor-section {
  padding: 22px 24px 24px;
}

.editor-section + .editor-section {
  border-top: 1px solid #eef2f6;
}

.section-title {
  margin-bottom: 18px;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.basic-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 0.9fr);
  gap: 20px;
}

.basic-fields :deep(.el-date-editor) {
  width: 100%;
}

.description-field {
  margin-bottom: 0;
}

.problem-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.problem-id-input {
  width: 220px;
}

.problem-table {
  border: 1px solid #e8edf3;
  border-radius: 4px;
}

.problem-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.problem-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.problem-actions__divider {
  width: 1px;
  height: 14px;
  background: #e2e8f0;
}

.editor-actions {
  position: sticky;
  z-index: 10;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  padding: 14px 18px;
  border-top: 1px solid #dfe6ef;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 -5px 14px rgba(15, 23, 42, 0.04);
  backdrop-filter: blur(8px);
}

@media (max-width: 760px) {
  .editor-actions {
    align-items: center;
  }

  .basic-fields {
    grid-template-columns: 1fr;
  }

  .editor-section {
    padding: 18px 16px 20px;
  }

  .problem-id-input {
    width: 100%;
  }
}
</style>
