<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessageBox } from "element-plus";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  getSubmissionDetail,
  getSubmissionList,
  rejudgeSubmission,
  type SubmissionDetail,
  type SubmissionItem
} from "@/api/admin/submissions";

defineOptions({ name: "AdminSubmissions" });

type SubmissionRow = SubmissionItem & { _busy?: boolean };

const STATUS_OPTIONS = [
  { value: 0, label: "Pending", type: "info" },
  { value: 10, label: "Judging", type: "warning" },
  { value: 1, label: "Accepted", type: "success" },
  { value: 5, label: "Wrong Answer", type: "danger" },
  { value: 3, label: "Time Limit Exceeded", type: "danger" },
  { value: 2, label: "Memory Limit Exceeded", type: "danger" },
  { value: 4, label: "Runtime Error", type: "danger" },
  { value: 6, label: "Compilation Error", type: "danger" },
  { value: 8, label: "Presentation Error", type: "warning" },
  { value: 9, label: "System Error", type: "danger" },
  { value: 7, label: "Unknown", type: "info" }
] as const;

const dataList = ref<SubmissionRow[]>([]);
const loading = ref(false);
const username = ref("");
const problemID = ref("");
const status = ref<number>();
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<SubmissionDetail>();
let detailSeq = 0;

const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

const columns: TableColumnList = [
  { label: "提交号", slot: "id", width: 120, align: "center" },
  { label: "题号", prop: "problemID", width: 110, align: "center" },
  {
    label: "题目",
    prop: "problemName",
    minWidth: 180,
    showOverflowTooltip: true
  },
  { label: "用户", prop: "userName", width: 140, align: "center" },
  { label: "结果", slot: "result", width: 170, align: "center" },
  { label: "语言", prop: "language", width: 110, align: "center" },
  { label: "时间", slot: "time", width: 100, align: "center" },
  { label: "内存", slot: "memory", width: 110, align: "center" },
  { label: "提交时间", prop: "createdAt", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 100,
    fixed: "right",
    align: "center"
  }
];

function statusInfo(value: number) {
  return (
    STATUS_OPTIONS.find(item => item.value === value) ?? {
      label: "Unknown",
      type: "info"
    }
  );
}

function formatTime(value: number) {
  if (!value) return "—";
  return value >= 1000 ? `${(value / 1000).toFixed(2)} s` : `${value} ms`;
}

function formatMemory(value: number) {
  if (!value) return "—";
  return value >= 1024 ? `${(value / 1024).toFixed(1)} MB` : `${value} KB`;
}

async function onSearch() {
  loading.value = true;
  try {
    const res = await getSubmissionList({
      page: pagination.currentPage,
      pageSize: pagination.pageSize,
      username: username.value.trim() || undefined,
      problemID: problemID.value.trim() || undefined,
      status: status.value
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
  username.value = "";
  problemID.value = "";
  status.value = undefined;
  applyFilter();
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

async function openDetail(row: SubmissionRow) {
  const seq = ++detailSeq;
  detailVisible.value = true;
  detailLoading.value = true;
  detail.value = undefined;
  try {
    const res = await getSubmissionDetail(row.id);
    if (seq === detailSeq) detail.value = res.data;
  } catch {
    if (seq === detailSeq) detailVisible.value = false;
  } finally {
    if (seq === detailSeq) detailLoading.value = false;
  }
}

function closeDetail() {
  detailSeq += 1;
  detail.value = undefined;
  detailLoading.value = false;
}

async function rejudge(row: SubmissionRow) {
  try {
    await ElMessageBox.confirm(
      `确认重判提交 #${row.id}？当前结果会被重置并重新进入评测队列。`,
      "重判提交",
      { type: "warning" }
    );
  } catch {
    return;
  }
  if (row._busy) return;
  row._busy = true;
  try {
    const res = await rejudgeSubmission(row.id);
    message(`已将 ${res.data?.rejudged ?? 0} 条提交加入重判队列`, {
      type: "success"
    });
    await onSearch();
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
          <span>提交记录</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ pagination.total }} 条
          </span>
        </div>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <el-input
            v-model="username"
            class="w-44!"
            clearable
            placeholder="用户名"
            @keyup.enter="applyFilter"
            @clear="applyFilter"
          />
          <el-input
            v-model="problemID"
            class="w-40!"
            clearable
            placeholder="题号"
            @keyup.enter="applyFilter"
            @clear="applyFilter"
          />
          <el-select
            v-model="status"
            class="w-52!"
            clearable
            placeholder="评测结果"
            @change="applyFilter"
          >
            <el-option
              v-for="item in STATUS_OPTIONS"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-button class="ml-auto!" @click="resetFilter">重置</el-button>
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
          <template #id="{ row }">
            <el-button link type="primary" @click="openDetail(row)">
              #{{ row.id }}
            </el-button>
          </template>
          <template #result="{ row }">
            <el-tag :type="statusInfo(row.result).type" size="small">
              {{ statusInfo(row.result).label }}
            </el-tag>
          </template>
          <template #time="{ row }">{{ formatTime(row.timeUsed) }}</template>
          <template #memory="{ row }">
            {{ formatMemory(row.memoryUsed) }}
          </template>
          <template #operation="{ row }">
            <el-button
              link
              size="small"
              type="warning"
              :loading="row._busy"
              @click="rejudge(row)"
            >
              重判
            </el-button>
          </template>
        </PureTable>
      </template>
    </PureTableBar>

    <el-dialog
      v-model="detailVisible"
      title="提交详情"
      width="820px"
      destroy-on-close
      @closed="closeDetail"
    >
      <div v-loading="detailLoading" class="min-h-56">
        <template v-if="detail">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="提交号">
              #{{ detail.submission.id }}
            </el-descriptions-item>
            <el-descriptions-item label="用户">
              {{ detail.user.name }}
            </el-descriptions-item>
            <el-descriptions-item label="题目">
              {{ detail.problem.id }} {{ detail.problem.name }}
            </el-descriptions-item>
            <el-descriptions-item label="语言">
              {{ detail.submission.language }}
            </el-descriptions-item>
            <el-descriptions-item label="结果">
              {{ statusInfo(detail.submission.status).label }}
            </el-descriptions-item>
            <el-descriptions-item label="提交时间">
              {{ detail.submission.createdAt }}
            </el-descriptions-item>
          </el-descriptions>
          <div class="mt-4">
            <strong class="text-sm">源代码</strong>
            <pre class="code-block">{{ detail.submission.code }}</pre>
          </div>
          <div v-if="detail.submission.compileOutput" class="mt-4">
            <strong class="text-sm">编译输出</strong>
            <pre class="code-block error-output">{{
              detail.submission.compileOutput
            }}</pre>
          </div>
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.code-block {
  max-height: 360px;
  padding: 14px;
  margin: 8px 0 0;
  overflow: auto;
  color: var(--el-text-color-primary);
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
}

.error-output {
  color: var(--el-color-danger);
}
</style>
