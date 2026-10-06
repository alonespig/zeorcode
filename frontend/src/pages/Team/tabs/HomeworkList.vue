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
      <el-table-column v-if="hasActions" label="操作" width="130" align="center" fixed="right">
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

  </div>
</template>

<script setup>
import { ref, computed, inject, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { getHomeworkList, deleteHomework } from "@/api/team";
import CapsuleProgress from "@/components/homework/CapsuleProgress.vue";

const team = inject("team");
const route = useRoute();
const router = useRouter();

const list = ref([]);
const hasActions = computed(() => list.value.some((homework) => homework.canEdit || homework.canDelete));
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

const openCreate = () => {
  router.push({ name: "HomeworkCreate", params: { id: route.params.id } });
};

const openEdit = (row) => {
  router.push({ name: "HomeworkEdit", params: { id: route.params.id, hid: row.id } });
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

onMounted(loadList);
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
