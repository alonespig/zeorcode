<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="t">题单管理</span>
          <span class="sub">共 {{ total }} 个（含草稿）</span>
        </div>
        <el-button type="primary" @click="openCreate">
          <el-icon class="mr-1">
            <Plus />
          </el-icon>新建题单
        </el-button>
      </div>
    </template>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="编号" width="80" align="center" />
      <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.published === 1 ? 'success' : 'info'" size="small" effect="light">
            {{ row.published === 1 ? '已上架' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="可见性" width="110" align="center">
        <template #default="{ row }">
          <el-tag :type="row.visibility === 1 ? 'warning' : 'success'" size="small" effect="light">
            {{ row.visibility === 1 ? '邀请码' : '公开' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="标签" width="180">
        <template #default="{ row }">
          <el-tag v-for="tag in row.tags" :key="tag.id" size="small" effect="plain" class="tag-item">
            {{ tag.name }}
          </el-tag>
          <span v-if="!row.tags?.length" class="text-empty">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="problemCount" label="题目数" width="90" align="center" />
      <el-table-column label="最近更新" width="170">
        <template #default="{ row }">{{ row.updatedAt }}</template>
      </el-table-column>
      <el-table-column label="操作" width="130" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" :loading="row._busy" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="还没有题单" :image-size="80" />
      </template>
    </el-table>

    <div class="pager">
      <el-pagination background layout="total, prev, pager, next" :current-page="page" :page-size="pageSize"
        :total="total" @current-change="handlePageChange" />
    </div>

    <!-- 新建 / 编辑 -->
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑题单' : '新建题单'" width="860px" top="6vh"
      :close-on-click-modal="false">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题" required>
          <el-input v-model="form.title" maxlength="100" show-word-limit placeholder="题单标题" />
        </el-form-item>

        <el-form-item label="标签">
          <el-select v-model="form.tagIds" multiple filterable placeholder="选择标签（可多选）" style="width: 100%">
            <el-option v-for="tag in tagList" :key="tag.id" :label="tag.name" :value="tag.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <!-- 用项目的 MdEditor 包装组件（全局自动注册），已处理图片上传与 mdHeadingId -->
          <MdEditor v-model="form.description" height="260px" editor-id="problemset-admin-editor" />
        </el-form-item>

        <el-form-item label="可见性">
          <el-radio-group v-model="form.visibility">
            <el-radio :value="0">公开</el-radio>
            <el-radio :value="1">需要邀请码</el-radio>
          </el-radio-group>
          <el-input v-if="form.visibility === 1" v-model="form.inviteCode" maxlength="32"
            :placeholder="editingId ? '留空则保留原邀请码' : '请输入邀请码'"
            class="code-input" />
        </el-form-item>

        <el-form-item label="上架">
          <el-switch v-model="publishedOn" />
          <span class="hint">关闭 = 草稿，前台完全看不到</span>
        </el-form-item>

        <el-form-item label="题目">
          <div class="problem-block">
            <div class="add-row">
              <el-input v-model="queryPid" placeholder="输入题号，如 P1001" class="pid-input"
                @keyup.enter="addProblem" />
              <el-button @click="addProblem">添加</el-button>
              <el-button type="primary" @click="openPicker">从题库选择</el-button>
              <span class="hint">已选 {{ problemList.length }} 道 · 拖动行可调整顺序</span>
            </div>

            <table v-if="problemList.length" class="pick-table">
              <thead>
                <tr>
                  <th style="width:44px"></th>
                  <th style="width:54px">#</th>
                  <th style="width:120px">题号</th>
                  <th>题目</th>
                  <th style="width:70px">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(problem, index) in problemList" :key="problem.id" draggable="true"
                  :class="{ dragging: dragIndex === index, 'drag-over': dragOverIndex === index }"
                  @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)" @dragend="onDragEnd"
                  @drop="onDrop(index)">
                  <td class="drag-handle">⋮⋮</td>
                  <td>{{ index + 1 }}</td>
                  <td>{{ problem.id }}</td>
                  <td>{{ problem.name }}</td>
                  <td>
                    <el-button link type="danger" size="small" @click="removeProblem(index)">移除</el-button>
                  </td>
                </tr>
              </tbody>
            </table>
            <el-empty v-else description="还没有添加题目" :image-size="60" />
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button :loading="saving" @click="save(0)">存为草稿</el-button>
        <el-button type="primary" :loading="saving" @click="save(1)">保存并上架</el-button>
      </template>
    </el-dialog>

    <!-- 从题库选择（移植自 Contest/CreateContest.vue，去掉分值与气球颜色） -->
    <el-dialog v-model="pickerVisible" title="从题库选择" width="720px" top="8vh">
      <el-input v-model="pickerQ" placeholder="搜索题号或标题" clearable :prefix-icon="Search" class="mb-3"
        @keyup.enter="loadPicker(1)" @clear="loadPicker(1)" />
      <el-table :data="pickerList" @row-click="(row) => togglePick(row.id)">
        <el-table-column width="50">
          <template #default="{ row }">
            <el-checkbox :model-value="isPicked(row.id)" :disabled="inList(row.id)"
              @click.stop="togglePick(row.id)" />
          </template>
        </el-table-column>
        <el-table-column label="题号" width="120" prop="id" />
        <el-table-column label="标题" min-width="180">
          <template #default="{ row }">
            {{ row.name }}
            <span v-if="inList(row.id)" class="text-empty">（已在题单中）</span>
          </template>
        </el-table-column>
        <el-table-column label="标签" width="180">
          <template #default="{ row }">
            <el-tag v-for="tag in row.tags" :key="tag.id" size="small" effect="plain" class="tag-item">
              {{ tag.name }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div class="picker-foot">
        <span class="hint">已选 <b>{{ picked.length }}</b> 道</span>
        <el-pagination background layout="prev, pager, next" :current-page="pickerPage" :page-size="pickerPageSize"
          :total="pickerTotal" @current-change="loadPicker" />
      </div>
      <template #footer>
        <el-button @click="pickerVisible = false">取消</el-button>
        <el-button type="primary" :loading="picking" @click="confirmPick">确定添加</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Search } from "@element-plus/icons-vue";
import {
  getAdminProblemSetList,
  createProblemSet,
  updateProblemSet,
  deleteProblemSet,
  getProblemSetDetail,
} from "@/api/problemset";
import { getProblem, getProblemList, getTags } from "@/api/problems";

const list = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const tagList = ref([]);

const dialogVisible = ref(false);
const saving = ref(false);
const editingId = ref(null);
const problemList = ref([]);
const queryPid = ref("");

const blankForm = () => ({
  title: "",
  description: "",
  visibility: 0,
  inviteCode: "",
  tagIds: [],
  published: 1,
});
const form = ref(blankForm());

// 开关与 0/1 之间转一层，避免表单里混用布尔和数字
const publishedOn = computed({
  get: () => form.value.published === 1,
  set: (val) => (form.value.published = val ? 1 : 0),
});

const loadList = async () => {
  loading.value = true;
  try {
    const res = await getAdminProblemSetList({ page: page.value, pageSize: pageSize.value });
    list.value = res.data?.list || [];
    total.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const loadTags = async () => {
  try {
    const res = await getTags();
    tagList.value = res.data?.tags || [];
  } catch (err) {
    console.error(err);
  }
};

const handlePageChange = (val) => {
  page.value = val;
  loadList();
};

const openCreate = () => {
  editingId.value = null;
  form.value = blankForm();
  problemList.value = [];
  queryPid.value = "";
  dialogVisible.value = true;
};

const openEdit = async (row) => {
  try {
    // 列表不含描述和题目，编辑前拉一次详情（管理员看得到草稿和未解锁内容）
    const res = await getProblemSetDetail(row.id);
    const d = res.data;
    editingId.value = row.id;
    form.value = {
      title: d.title,
      description: d.description || "",
      visibility: d.visibility,
      inviteCode: "",
      tagIds: (d.tags || []).map((t) => t.id),
      published: d.published,
    };
    problemList.value = (d.problems || []).map((p) => ({ id: p.id, name: p.name }));
    queryPid.value = "";
    dialogVisible.value = true;
  } catch (err) {
    console.error(err);
  }
};

const save = async (published) => {
  if (!form.value.title.trim()) {
    ElMessage.warning("请填写标题");
    return;
  }
  if (!editingId.value && form.value.visibility === 1 && !form.value.inviteCode.trim()) {
    ElMessage.warning("选择邀请码可见时必须填写邀请码");
    return;
  }
  const payload = {
    title: form.value.title.trim(),
    description: form.value.description,
    published,
    visibility: form.value.visibility,
    inviteCode: form.value.inviteCode.trim(),
    tagIds: form.value.tagIds,
    problems: problemList.value.map((p) => String(p.id)),
  };
  saving.value = true;
  try {
    if (editingId.value) {
      await updateProblemSet(editingId.value, payload);
    } else {
      await createProblemSet(payload);
    }
    ElMessage.success(published === 1 ? "已保存并上架" : "已存为草稿");
    dialogVisible.value = false;
    loadList();
  } catch (err) {
    console.error(err);
  } finally {
    saving.value = false;
  }
};

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除题单「${row.title}」？删除后不可恢复。`, "删除题单", { type: "warning" });
  } catch {
    return;
  }
  row._busy = true;
  try {
    await deleteProblemSet(row.id);
    ElMessage.success("已删除");
    loadList();
  } catch (err) {
    console.error(err);
  } finally {
    row._busy = false;
  }
};

// ---------- 题目增删 ----------
const addProblem = async () => {
  const id = queryPid.value.trim();
  if (!id) {
    ElMessage.warning("请输入题号");
    return;
  }
  if (problemList.value.some((p) => String(p.id) === id)) {
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

const removeProblem = (index) => {
  problemList.value.splice(index, 1);
};

// ---------- 拖拽排序（原生 HTML5 DnD，无额外依赖） ----------
const dragIndex = ref(-1);
const dragOverIndex = ref(-1);
const onDragStart = (i) => { dragIndex.value = i; };
const onDragOver = (i) => { dragOverIndex.value = i; };
const onDragEnd = () => { dragIndex.value = -1; dragOverIndex.value = -1; };
const onDrop = (i) => {
  const from = dragIndex.value;
  if (from !== -1 && from !== i) {
    const [moved] = problemList.value.splice(from, 1);
    problemList.value.splice(i, 0, moved);
  }
  onDragEnd();
};

// ---------- 从题库选择 ----------
const pickerVisible = ref(false);
const picking = ref(false);
const pickerQ = ref("");
const pickerList = ref([]);
const pickerPage = ref(1);
const pickerPageSize = 8;
const pickerTotal = ref(0);
const picked = ref([]);

const inList = (id) => problemList.value.some((p) => p.id === id);
const isPicked = (id) => picked.value.includes(id);
const togglePick = (id) => {
  if (inList(id)) return; // 已在题单中的不可取消
  const idx = picked.value.indexOf(id);
  if (idx === -1) picked.value.push(id);
  else picked.value.splice(idx, 1);
};

const loadPicker = async (p = 1) => {
  pickerPage.value = p;
  try {
    const res = await getProblemList({ page: p, pageSize: pickerPageSize, q: pickerQ.value || undefined });
    pickerList.value = res.data?.list || [];
    pickerTotal.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  }
};

const openPicker = () => {
  picked.value = problemList.value.map((p) => p.id); // 已添加的预先勾选
  pickerQ.value = "";
  pickerVisible.value = true;
  loadPicker(1);
};

const confirmPick = async () => {
  const toAdd = picked.value.filter((id) => !inList(id));
  if (toAdd.length === 0) {
    pickerVisible.value = false;
    return;
  }
  picking.value = true;
  try {
    const results = await Promise.all(toAdd.map((id) => getProblem(id)));
    results.forEach((res) => problemList.value.push({ id: res.data.id, name: res.data.name }));
    pickerVisible.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    picking.value = false;
  }
};

onMounted(() => {
  loadList();
  loadTags();
});
</script>

<style scoped>
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.head-title .t {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.head-title .sub {
  margin-left: 10px;
  font-size: 13px;
  color: #909399;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.tag-item {
  margin-right: 4px;
}

.text-empty {
  color: #c0c4cc;
}

.mr-1 {
  margin-right: 4px;
}

.mb-3 {
  margin-bottom: 12px;
}

.hint {
  font-size: 13px;
  color: #909399;
  margin-left: 8px;
}

.code-input {
  max-width: 200px;
  margin-left: 12px;
}

.problem-block {
  width: 100%;
}

.add-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.pid-input {
  max-width: 180px;
}

.pick-table {
  width: 100%;
  border-collapse: collapse;
  background: #fafbfc;
  border: 1px solid var(--table-border);
  border-radius: var(--radius-base);
  font-size: 14px;
}

.pick-table th {
  height: 40px;
  background: var(--table-header-bg);
  color: var(--table-header-text);
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  padding: 0 12px;
  border-bottom: 1px solid var(--table-border);
}

.pick-table td {
  padding: 9px 12px;
  border-bottom: 1px solid var(--table-row-border);
}

.pick-table tr:last-child td {
  border-bottom: none;
}

.pick-table tr.dragging {
  opacity: .5;
}

.pick-table tr.drag-over {
  background: var(--table-row-hover);
  outline: 1px dashed var(--color-primary);
}

.drag-handle {
  cursor: grab;
  color: #c0c4cc;
  user-select: none;
  letter-spacing: 1px;
}

.picker-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 12px;
}
</style>
