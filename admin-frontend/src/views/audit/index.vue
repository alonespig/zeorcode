<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { PureTableBar } from "@/components/RePureTableBar";
import { getAuditLogs, type AuditLogItem } from "@/api/admin/audit";

defineOptions({ name: "AdminAuditLogs" });

const dataList = ref<AuditLogItem[]>([]);
const loading = ref(false);
const keyword = ref("");
const method = ref("");
const success = ref<boolean>();
const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

const columns: TableColumnList = [
  { label: "编号", prop: "id", width: 90, align: "center" },
  { label: "管理员", slot: "actor", minWidth: 150 },
  { label: "方法", slot: "method", width: 90, align: "center" },
  { label: "接口", prop: "path", minWidth: 230, showOverflowTooltip: true },
  { label: "操作对象", prop: "target", minWidth: 150, slot: "target" },
  { label: "结果", slot: "result", width: 110, align: "center" },
  { label: "业务码", prop: "code", width: 100, align: "center" },
  { label: "来源 IP", prop: "clientIP", width: 140, align: "center" },
  { label: "操作时间", prop: "createdAt", width: 175, align: "center" }
];

async function onSearch() {
  loading.value = true;
  try {
    const response = await getAuditLogs({
      page: pagination.currentPage,
      pageSize: pagination.pageSize,
      q: keyword.value.trim() || undefined,
      method: method.value || undefined,
      success: success.value
    });
    dataList.value = response.data?.list || [];
    pagination.total = response.data?.total || 0;
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
  method.value = "";
  success.value = undefined;
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

function methodType(value: string) {
  if (value === "DELETE") return "danger";
  if (value === "POST") return "success";
  if (value === "PUT" || value === "PATCH") return "warning";
  return "info";
}

onMounted(onSearch);
</script>

<template>
  <PureTableBar :columns="columns" @refresh="onSearch">
    <template #title>
      <div class="flex items-center">
        <span>操作日志</span>
        <span class="ml-2 text-xs font-normal text-gray-400"
          >共 {{ pagination.total }} 条</span
        >
      </div>
    </template>
    <template v-slot="{ size, dynamicColumns }">
      <div class="mb-3 flex flex-wrap items-center gap-2">
        <el-input
          v-model="keyword"
          class="w-64"
          clearable
          placeholder="管理员、接口或操作对象"
          @keyup.enter="applyFilter"
          @clear="applyFilter"
        />
        <el-select
          v-model="method"
          class="w-32"
          clearable
          placeholder="请求方法"
          @change="applyFilter"
        >
          <el-option
            v-for="item in ['POST', 'PUT', 'PATCH', 'DELETE']"
            :key="item"
            :label="item"
            :value="item"
          />
        </el-select>
        <el-select
          v-model="success"
          class="w-32"
          clearable
          placeholder="执行结果"
          @change="applyFilter"
        >
          <el-option label="成功" :value="true" />
          <el-option label="失败" :value="false" />
        </el-select>
        <el-button type="primary" @click="applyFilter">查询</el-button>
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
        <template #actor="{ row }">
          <strong>{{ row.actorName || "未知管理员" }}</strong>
          <small class="ml-2 text-gray-400">#{{ row.actorID || "—" }}</small>
        </template>
        <template #method="{ row }"
          ><el-tag :type="methodType(row.method)" effect="plain">{{
            row.method
          }}</el-tag></template
        >
        <template #target="{ row }">{{ row.target || "—" }}</template>
        <template #result="{ row }"
          ><el-tag :type="row.success ? 'success' : 'danger'">{{
            row.success ? "成功" : "失败"
          }}</el-tag></template
        >
      </PureTable>
    </template>
  </PureTableBar>
</template>
