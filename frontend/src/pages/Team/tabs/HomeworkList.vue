<template>
  <div class="f-panel">
    <div class="flex items-center justify-between border-b border-gray-100 px-6 py-3.5">
      <span class="text-sm text-gray-500">共 {{ total }} 个作业</span>
      <el-button v-if="team.canManage" type="primary" :icon="Plus" @click="openCreate">布置作业</el-button>
    </div>

    <el-table class="homework-table" :data="list" style="width: 100%">
      <el-table-column label="作业名称" min-width="260" align="center">
        <template #default="{ row }">
          <router-link class="font-medium text-blue-500 hover:text-blue-400"
            :to="`/team/${team.id}/homework/${row.id}`">
            {{ row.title }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="110" align="center">
        <template #default="{ row }">
          <el-tag size="small" :type="statusMeta(row.status).type" effect="light">
            {{ statusMeta(row.status).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="题数" width="90" align="center">
        <template #default="{ row }">{{ row.problemCount }}</template>
      </el-table-column>
      <el-table-column label="我的进度" min-width="210" align="center">
        <template #default="{ row }">
          <div v-if="row.solvedCount != null && row.problemCount > 0" class="homework-progress-cell">
            <CapsuleProgress
              :value="row.solvedCount"
              :max="row.problemCount"
              compact
              :aria-label="`已完成 ${row.solvedCount} 题，共 ${row.problemCount} 题`"
            />
          </div>
          <span v-else class="text-gray-300">—</span>
        </template>
      </el-table-column>
      <el-table-column label="开始时间" width="170" align="center">
        <template #default="{ row }">
          <span class="tabular-nums text-gray-600">{{ fmtTime(row.startTime) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="结束时间" width="170" align="center">
        <template #default="{ row }">
          <span class="tabular-nums text-gray-600">{{ fmtTime(row.endTime) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="130" align="center" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.canEdit" link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="row.canDelete" link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="还没有作业" :image-size="80" />
      </template>
    </el-table>

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination size="large" :page-size="pageSize" :pager-count="16" v-model:current-page="page"
        layout="prev, pager, next" :total="total" @current-change="loadList" />
    </div>

    <!-- 布置 / 编辑作业 -->
    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑作业' : '布置作业'" width="780px" top="6vh"
      :close-on-click-modal="false">
      <el-form :model="form" label-width="80px">
        <el-form-item label="作业名称" required>
          <el-input v-model="form.title" maxlength="100" show-word-limit placeholder="作业名称" />
        </el-form-item>
        <el-form-item label="起止时间" required>
          <el-date-picker v-model="form.range" type="datetimerange" range-separator="→"
            start-placeholder="开始时间" end-placeholder="结束时间" format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DD HH:mm:ss" />
        </el-form-item>
        <el-form-item label="作业简介">
          <MdEditor v-model="form.description" height="200px" editor-id="homework-editor" />
        </el-form-item>
        <el-form-item label="题目">
          <div class="w-full">
            <div class="mb-2.5 flex flex-wrap items-center gap-2">
              <el-input v-model="queryPid" placeholder="输入题号，如 L101" class="w-40!" @keyup.enter="addProblem" />
              <el-button @click="addProblem">添加</el-button>
              <el-button type="primary" plain @click="openPicker">从题库选择</el-button>
              <span class="text-xs text-gray-400">已选 {{ problemList.length }} 道 · 拖动调整顺序 · 每题 100 分</span>
            </div>
            <el-table v-if="problemList.length" :data="problemList" size="small" class="border border-gray-200 rounded">
              <el-table-column width="44">
                <template #default="{ $index }"><span class="cursor-grab text-gray-400">⋮⋮</span></template>
              </el-table-column>
              <el-table-column label="#" width="54" align="center">
                <template #default="{ $index }">{{ $index + 1 }}</template>
              </el-table-column>
              <el-table-column prop="id" label="题号" width="120" />
              <el-table-column prop="name" label="题目" />
              <el-table-column label="操作" width="70" align="center">
                <template #default="{ $index }">
                  <el-button link type="danger" size="small" @click="removeProblem($index)">移除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="还没有添加题目" :image-size="50" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 从题库选择 -->
    <el-dialog v-model="pickerVisible" title="从题库选择" width="680px" top="8vh">
      <el-input v-model="pickerQ" placeholder="搜索题号或标题" clearable :prefix-icon="Search" class="mb-3"
        @keyup.enter="loadPicker(1)" @clear="loadPicker(1)" />
      <el-table :data="pickerList" size="small" @row-click="(row) => togglePick(row.id)">
        <el-table-column width="50">
          <template #default="{ row }">
            <el-checkbox :model-value="isPicked(row.id)" :disabled="inList(row.id)" @click.stop="togglePick(row.id)" />
          </template>
        </el-table-column>
        <el-table-column prop="id" label="题号" width="120" />
        <el-table-column label="标题" min-width="180">
          <template #default="{ row }">
            {{ row.name }}
            <span v-if="inList(row.id)" class="text-gray-400">（已添加）</span>
          </template>
        </el-table-column>
        <el-table-column label="难度" width="90" align="center">
          <template #default="{ row }">{{ diffText(row.difficulty) }}</template>
        </el-table-column>
      </el-table>
      <div class="mt-3 flex items-center justify-between">
        <span class="text-sm text-gray-500">已选 <b class="text-blue-500">{{ picked.length }}</b> 道</span>
        <el-pagination background layout="prev, pager, next" :current-page="pickerPage" :page-size="pickerPageSize"
          :total="pickerTotal" @current-change="loadPicker" />
      </div>
      <template #footer>
        <el-button @click="pickerVisible = false">取消</el-button>
        <el-button type="primary" :loading="picking" @click="confirmPick">确定添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, inject, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Search } from "@element-plus/icons-vue";
import { getHomeworkList, createHomework, updateHomework, deleteHomework, getHomeworkDetail } from "@/api/team";
import { getProblem, getProblemList } from "@/api/problems";
import CapsuleProgress from "@/components/homework/CapsuleProgress.vue";

const team = inject("team");
const route = useRoute();
const router = useRouter();

const list = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = 10;

const statusMeta = (s) => ({
  0: { text: "未开始", type: "warning" },
  1: { text: "进行中", type: "success" },
  2: { text: "已截止", type: "info" },
}[s] || { text: "-", type: "info" });

// 同年省略年份，列表统一显示到分钟。
const fmtTime = (s) => {
  if (!s) return "-";
  return s.slice(0, 4) === String(new Date().getFullYear()) ? s.slice(5, 16) : s.slice(0, 16);
};

const diffText = (d) => ({ 1: "简单", 2: "中等", 3: "困难" }[d] || "-");

const loadList = async () => {
  try {
    // 用路由参数而非 inject 的 team.value.id：onMounted 时 team 可能还没拉回来
    const res = await getHomeworkList(route.params.id, { page: page.value, pageSize });
    list.value = res.data?.list || [];
    total.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  }
};

// ===== 布置/编辑 =====
const dialogVisible = ref(false);
const saving = ref(false);
const editingId = ref(null);
const queryPid = ref("");
const problemList = ref([]);
const form = ref(blankForm());

function blankForm() {
  return { title: "", description: "", range: [] };
}

const toMinutePrecision = (value) => `${value.slice(0, 16)}:00`;

const openCreate = () => {
  editingId.value = null;
  form.value = blankForm();
  problemList.value = [];
  queryPid.value = "";
  dialogVisible.value = true;
};

const openEdit = async (row) => {
  try {
    const res = await getHomeworkDetail(row.id);
    const d = res.data;
    editingId.value = row.id;
    form.value = {
      title: d.title,
      description: d.description || "",
      range: [d.startTime, d.endTime],
    };
    problemList.value = (d.problems || []).map((p) => ({ id: p.id, name: p.name }));
    queryPid.value = "";
    dialogVisible.value = true;
  } catch (err) {
    console.error(err);
  }
};

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

const removeProblem = (i) => problemList.value.splice(i, 1);

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
    problems: problemList.value.map((p) => String(p.id)),
  };
  saving.value = true;
  try {
    if (editingId.value) {
      await updateHomework(editingId.value, payload);
      ElMessage.success("已更新");
    } else {
      await createHomework(team.value.id, payload);
      ElMessage.success("已布置");
    }
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
    await ElMessageBox.confirm(`确认删除作业「${row.title}」？`, "删除作业", { type: "warning" });
  } catch {
    return;
  }
  try {
    await deleteHomework(row.id);
    ElMessage.success("已删除");
    if (list.value.length === 1 && page.value > 1) page.value -= 1;
    await loadList();
  } catch (err) {
    console.error(err);
  }
};

// ===== 从题库选择 =====
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
  if (inList(id)) return;
  const i = picked.value.indexOf(id);
  if (i === -1) picked.value.push(id);
  else picked.value.splice(i, 1);
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
  picked.value = problemList.value.map((p) => p.id);
  pickerQ.value = "";
  pickerVisible.value = true;
  loadPicker(1);
};

const confirmPick = async () => {
  const toAdd = picked.value.filter((id) => !inList(id));
  if (!toAdd.length) {
    pickerVisible.value = false;
    return;
  }
  picking.value = true;
  try {
    const results = await Promise.all(toAdd.map((id) => getProblem(id)));
    results.forEach((r) => problemList.value.push({ id: r.data.id, name: r.data.name }));
    pickerVisible.value = false;
  } catch (err) {
    console.error(err);
  } finally {
    picking.value = false;
  }
};

onMounted(loadList);

// 从概览页「布置作业」跳过来带 ?create=1 时自动打开弹窗
watch(
  () => [route.query.create, team.value.id, team.value.canManage],
  async ([create, teamID, canManage]) => {
    if (create !== "1" || teamID == null) return;
    if (canManage) openCreate();
    const query = { ...route.query };
    delete query.create;
    await router.replace({ query });
  },
  { immediate: true }
);
</script>

<style scoped>
.homework-progress-cell {
  display: flex;
  justify-content: center;
}

:deep(.homework-table th.el-table__cell) {
  padding-block: 13px;
}

:deep(.homework-table td.el-table__cell) {
  padding-block: 16px;
}

</style>
