<template>
  <el-drawer v-model="visible" :title="isEdit ? '编辑题单' : '新建题单'" direction="rtl" size="min(960px, 92vw)"
    :close-on-click-modal="false">
    <el-form v-loading="loading" :model="form" label-position="top" class="edit-form">
      <section class="edit-section">
        <h3 class="section-title">基本信息</h3>
        <el-form-item label="标题" required>
          <el-input v-model="form.title" maxlength="100" show-word-limit placeholder="题单标题" />
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="form.tagIds" multiple filterable collapse-tags collapse-tags-tooltip
            placeholder="选择标签（可多选）" class="w-full">
            <el-option v-for="tag in tagList" :key="tag.id" :label="tag.name" :value="tag.id" />
          </el-select>
        </el-form-item>
        <div class="setting-grid">
          <el-form-item label="可见性">
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
            <span class="field-hint publish-hint">草稿仅管理员可见</span>
          </el-form-item>
        </div>
        <el-form-item v-if="form.visibility === 1" label="邀请码">
          <el-input v-model="form.inviteCode" maxlength="32"
            :placeholder="isEdit ? '留空则保留原邀请码' : '请输入邀请码'" />
        </el-form-item>
      </section>

      <section class="edit-section">
        <div class="section-head">
          <h3 class="section-title">题目列表</h3>
          <span class="field-hint">{{ problemList.length }} 道 · 拖动行可调整顺序</span>
        </div>
        <div class="problem-toolbar">
          <el-input v-model="queryPid" placeholder="输入题号，如 L101" class="pid-input"
            @keyup.enter="addProblem" />
          <el-button :icon="Plus" @click="addProblem">添加</el-button>
          <el-button type="primary" plain :icon="Search" @click="pickerVisible = true">从题库选择</el-button>
        </div>
        <table v-if="problemList.length" class="pick-table">
          <thead>
            <tr>
              <th style="width: 44px"></th>
              <th style="width: 54px">#</th>
              <th style="width: 130px">题号</th>
              <th>题目</th>
              <th style="width: 100px">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(problem, index) in problemList" :key="problem.id" draggable="true"
              :class="{ dragging: dragIndex === index, 'drag-over': dragOverIndex === index }"
              @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)"
              @dragend="onDragEnd" @drop="onDrop(index)">
              <td class="drag-handle">⋮⋮</td>
              <td>{{ index + 1 }}</td>
              <td class="mono">{{ problem.id }}</td>
              <td>{{ problem.name }}</td>
              <td>
                <div class="row-actions">
                  <el-tooltip content="查看题目" placement="top">
                    <el-button link type="primary" :icon="View" aria-label="查看题目"
                      @click="openPreview(problem.id)" />
                  </el-tooltip>
                  <el-button link type="danger" @click="removeProblem(index)">移除</el-button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <el-empty v-else description="还没有添加题目" :image-size="72" />
      </section>

      <section class="edit-section">
        <h3 class="section-title">题单说明</h3>
        <MdEditor v-model="form.description" height="260px" editor-id="problemset-edit-editor" />
      </section>
    </el-form>

    <template #footer>
      <div class="drawer-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{ isEdit ? "保存" : "创建" }}
        </el-button>
      </div>
    </template>

    <ProblemPickerDialog v-model="pickerVisible" :existing-problem-ids="existingProblemIds"
      @add="addPickedProblems" />
    <ProblemPreviewDialog v-model="previewVisible" :problem-id="previewProblemId" />
  </el-drawer>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { Plus, Search, View } from "@element-plus/icons-vue";
import { createProblemSet, getProblemSetDetail, updateProblemSet } from "@/api/problemset";
import { getProblem, getTags } from "@/api/problems";
import MdEditor from "@/components/MdEditor.vue";
import ProblemPickerDialog from "@/components/problem/ProblemPickerDialog.vue";
import ProblemPreviewDialog from "@/components/problem/ProblemPreviewDialog.vue";

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  problemSetId: { type: [String, Number], default: "" },
});
const emit = defineEmits(["update:modelValue", "saved"]);

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const isEdit = computed(() => Boolean(props.problemSetId));

const loading = ref(false);
const saving = ref(false);
const tagList = ref([]);
const problemList = ref([]);
const queryPid = ref("");
const pickerVisible = ref(false);
const previewVisible = ref(false);
const previewProblemId = ref("");
const form = ref(blankForm());

function blankForm() {
  return { title: "", description: "", visibility: 0, inviteCode: "", tagIds: [], published: 1 };
}

const existingProblemIds = computed(() => problemList.value.map((problem) => problem.id));

const load = async () => {
  form.value = blankForm();
  problemList.value = [];
  queryPid.value = "";
  loading.value = true;
  try {
    const tagsRes = await getTags().catch(() => ({ data: { tags: [] } }));
    tagList.value = tagsRes.data?.tags || [];
    if (!isEdit.value) return;

    const detailRes = await getProblemSetDetail(props.problemSetId);
    const detail = detailRes.data || {};
    form.value = {
      title: detail.title || "",
      description: detail.description || "",
      visibility: detail.visibility ?? 0,
      inviteCode: "",
      tagIds: (detail.tags || []).map((tag) => tag.id),
      published: detail.published ?? 1,
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

const removeProblem = (index) => problemList.value.splice(index, 1);

const openPreview = (id) => {
  previewProblemId.value = String(id);
  previewVisible.value = true;
};

const addPickedProblems = (problems) => {
  const existing = new Set(problemList.value.map((problem) => String(problem.id)));
  problems.forEach((problem) => {
    if (existing.has(String(problem.id))) return;
    problemList.value.push({ id: problem.id, name: problem.name });
    existing.add(String(problem.id));
  });
};

const dragIndex = ref(-1);
const dragOverIndex = ref(-1);
const onDragStart = (index) => { dragIndex.value = index; };
const onDragOver = (index) => { dragOverIndex.value = index; };
const onDragEnd = () => { dragIndex.value = -1; dragOverIndex.value = -1; };
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
    ElMessage.warning("请填写标题");
    return;
  }
  if (!isEdit.value && form.value.visibility === 1 && !form.value.inviteCode.trim()) {
    ElMessage.warning("选择邀请码可见时必须填写邀请码");
    return;
  }
  const payload = {
    title: form.value.title.trim(),
    description: form.value.description,
    published: form.value.published,
    visibility: form.value.visibility,
    inviteCode: form.value.inviteCode.trim(),
    tagIds: form.value.tagIds,
    problems: problemList.value.map((problem) => String(problem.id)),
  };
  saving.value = true;
  try {
    if (isEdit.value) {
      await updateProblemSet(props.problemSetId, payload);
      ElMessage.success("题单已保存");
    } else {
      await createProblemSet(payload);
      ElMessage.success("题单已创建");
    }
    emit("saved");
    visible.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    saving.value = false;
  }
};

watch(() => props.modelValue, (open) => {
  if (open) load();
});
</script>

<style scoped>
.edit-form {
  padding: 0 2px;
}

.drawer-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 0;
}

.edit-section + .edit-section {
  margin-top: 28px;
}

.section-title {
  margin: 0 0 14px;
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-head .section-title {
  margin: 0;
}

.setting-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
}

.field-hint {
  display: block;
  margin-top: 2px;
  color: #909399;
  font-size: 12.5px;
}

.publish-hint {
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
  max-width: 200px;
}

.pick-table {
  width: 100%;
  overflow: hidden;
  border: 1px solid var(--table-border);
  border-radius: var(--radius-base);
  border-collapse: collapse;
  background: #fafbfc;
  font-size: 14px;
}

.pick-table th {
  height: 40px;
  padding: 0 12px;
  border-bottom: 1px solid var(--table-border);
  color: var(--table-header-text);
  background: var(--table-header-bg);
  font-size: 13px;
  font-weight: 600;
  text-align: left;
}

.pick-table td {
  padding: 9px 12px;
  border-bottom: 1px solid var(--table-row-border);
}

.pick-table tr:last-child td {
  border-bottom: none;
}

.pick-table tr.dragging {
  opacity: 0.5;
}

.pick-table tr.drag-over {
  background: #eef5ff;
}

.drag-handle {
  color: #b8c2cf;
  cursor: grab;
  letter-spacing: 1px;
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.row-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.mono {
  font-variant-numeric: tabular-nums;
}

@media (max-width: 760px) {
  .setting-grid {
    grid-template-columns: 1fr;
  }
}
</style>
