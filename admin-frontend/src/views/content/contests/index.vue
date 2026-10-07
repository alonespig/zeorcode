<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import dayjs from "dayjs";
import { ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  getAdminContestList,
  recomputeContest,
  rejudgeContest,
  setContestArchived,
  type ContestItem
} from "@/api/admin/contests";

defineOptions({ name: "AdminContests" });

type ContestRow = ContestItem & { _busy?: boolean };

const STATUS_LABELS = ["未开始", "进行中", "已结束"];
const STATUS_TYPES: Array<"info" | "success"> = ["info", "success", "info"];
const RULE_OPTIONS = [
  { value: 1, label: "ACM" },
  { value: 2, label: "OI" },
  { value: 3, label: "IOI" },
  { value: 4, label: "CF" }
];

const router = useRouter();
const dataList = ref<ContestRow[]>([]);
const loading = ref(false);
const keyword = ref("");
const ruleType = ref<number>();
const status = ref<number>();
const archived = ref<boolean>();
const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

const columns: TableColumnList = [
  { label: "编号", prop: "id", width: 90, align: "center" },
  {
    label: "比赛名称",
    prop: "name",
    minWidth: 210,
    showOverflowTooltip: true
  },
  { label: "赛制", prop: "type", width: 90, align: "center" },
  { label: "状态", slot: "status", width: 130, align: "center" },
  { label: "开始时间", slot: "startTime", width: 170, align: "center" },
  { label: "时长", slot: "duration", width: 100, align: "center" },
  { label: "参赛人数", prop: "participants", width: 110, align: "center" },
  { label: "属性", slot: "flags", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 270,
    fixed: "right",
    align: "center"
  }
];

async function onSearch() {
  loading.value = true;
  try {
    const res = await getAdminContestList({
      page: pagination.currentPage,
      pageSize: pagination.pageSize,
      keyword: keyword.value.trim() || undefined,
      type: ruleType.value,
      status: status.value,
      archived: archived.value
    });
    dataList.value = res.data?.list ?? [];
    pagination.total = res.data?.total ?? 0;
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function applyFilter() {
  pagination.currentPage = 1;
  onSearch();
}

function resetFilter() {
  keyword.value = "";
  ruleType.value = undefined;
  status.value = undefined;
  archived.value = undefined;
  applyFilter();
}

async function toggleArchive(row: ContestRow) {
  const next = !row.archived;
  if (next && row.status === 1) {
    message("进行中的比赛不能归档", { type: "warning" });
    return;
  }
  try {
    await ElMessageBox.confirm(
      next
        ? `确认归档比赛「${row.name}」？归档后前台将无法查看和访问该比赛。`
        : `确认恢复比赛「${row.name}」？恢复后比赛将重新在前台显示。`,
      next ? "归档比赛" : "恢复比赛",
      { type: next ? "warning" : "info" }
    );
  } catch {
    return;
  }
  if (row._busy) return;
  row._busy = true;
  try {
    await setContestArchived(row.id, next);
    message(next ? "比赛已归档" : "比赛已恢复", { type: "success" });
    await onSearch();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

function onCurrentChange(page: number) {
  pagination.currentPage = page;
  onSearch();
}

function onSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.currentPage = 1;
  onSearch();
}

async function runRejudge(row: ContestRow) {
  try {
    await ElMessageBox.confirm(
      `确认重判比赛「${row.name}」的全部提交？该操作会产生大量评测任务。`,
      "整场重判",
      { type: "warning" }
    );
  } catch {
    return;
  }
  if (row._busy) return;
  row._busy = true;
  try {
    const res = await rejudgeContest(row.id);
    message(`已将 ${res.data?.rejudged ?? 0} 条提交加入重判队列`, {
      type: "success"
    });
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

async function rebuildRank(row: ContestRow) {
  try {
    await ElMessageBox.confirm(
      `确认根据提交明细重算比赛「${row.name}」的排行榜？`,
      "重算排行榜",
      { type: "warning" }
    );
  } catch {
    return;
  }
  if (row._busy) return;
  row._busy = true;
  try {
    await recomputeContest(row.id);
    message("排行榜已重算", { type: "success" });
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

onMounted(onSearch);
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="onSearch">
      <template #title>
        <div class="flex items-center">
          <span>比赛管理</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ pagination.total }} 场
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button
          type="primary"
          :icon="Plus"
          @click="router.push('/content/contests/create')"
        >
          新建比赛
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <el-input
            v-model="keyword"
            class="w-56"
            clearable
            placeholder="比赛名称"
            @keyup.enter="applyFilter"
            @clear="applyFilter"
          />
          <el-select
            v-model="ruleType"
            class="w-28"
            clearable
            placeholder="赛制"
            @change="applyFilter"
          >
            <el-option
              v-for="item in RULE_OPTIONS"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-select
            v-model="status"
            class="w-32"
            clearable
            placeholder="状态"
            @change="applyFilter"
          >
            <el-option label="未开始" :value="0" />
            <el-option label="进行中" :value="1" />
            <el-option label="已结束" :value="2" />
          </el-select>
          <el-select
            v-model="archived"
            class="w-32"
            clearable
            placeholder="归档状态"
            @change="applyFilter"
          >
            <el-option label="正常" :value="false" />
            <el-option label="已归档" :value="true" />
          </el-select>
          <el-button @click="resetFilter">重置</el-button>
        </div>
        <PureTable
          border
          adaptive
          table-layout="auto"
          align-whole="center"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="pagination"
          @page-size-change="onSizeChange"
          @page-current-change="onCurrentChange"
        >
          <template #status="{ row }">
            <el-tag :type="STATUS_TYPES[row.status]" size="small">
              {{ STATUS_LABELS[row.status] ?? "未知" }}
            </el-tag>
            <el-tag v-if="row.archived" class="ml-1" size="small" type="info">
              已归档
            </el-tag>
          </template>
          <template #startTime="{ row }">
            {{ dayjs(row.startTime).format("YYYY-MM-DD HH:mm") }}
          </template>
          <template #duration="{ row }">{{ row.duration }} 分钟</template>
          <template #flags="{ row }">
            <el-tag v-if="row.rated" size="small" type="warning" class="mr-1">
              Rated
            </el-tag>
            <el-tag v-if="row.needInviteCode" size="small" type="info">
              邀请码
            </el-tag>
            <span
              v-if="!row.rated && !row.needInviteCode"
              class="text-gray-400"
            >
              —
            </span>
          </template>
          <template #operation="{ row }">
            <el-button
              link
              size="small"
              type="primary"
              @click="router.push(`/content/contests/${row.id}/edit`)"
            >
              编辑
            </el-button>
            <el-button
              link
              size="small"
              type="warning"
              :loading="row._busy"
              @click="runRejudge(row)"
            >
              重判
            </el-button>
            <el-button
              link
              size="small"
              :loading="row._busy"
              @click="rebuildRank(row)"
            >
              重算榜单
            </el-button>
            <el-button
              link
              size="small"
              :type="row.archived ? 'success' : 'danger'"
              :disabled="!row.archived && row.status === 1"
              :loading="row._busy"
              @click="toggleArchive(row)"
            >
              {{ row.archived ? "恢复" : "归档" }}
            </el-button>
          </template>
        </PureTable>
      </template>
    </PureTableBar>
  </div>
</template>
