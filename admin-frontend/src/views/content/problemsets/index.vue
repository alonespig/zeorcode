<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import { getTagList, type TagItem } from "@/api/admin/problems";
import {
  deleteProblemSet,
  getAdminProblemSetList,
  type ProblemSetItem
} from "@/api/admin/problemsets";

defineOptions({ name: "AdminProblemSets" });

/** 表格行数据，附加按行的提交中状态，用于防重复点击 */
type ProblemSetRow = ProblemSetItem & { _busy?: boolean };

/** 状态与可见性集中映射（对应后端 dto.ProblemSetItemResp） */
const PUBLISHED_LABELS: Record<number, string> = { 0: "草稿", 1: "已发布" };
const PUBLISHED_TAG_TYPES: Record<number, "info" | "success"> = {
  0: "info",
  1: "success"
};
const VISIBILITY_LABELS: Record<number, string> = { 0: "公开", 1: "邀请码" };
const VISIBILITY_TAG_TYPES: Record<number, "success" | "warning"> = {
  0: "success",
  1: "warning"
};

const router = useRouter();

const dataList = ref<ProblemSetRow[]>([]);
const loading = ref(false);
const tagOptions = ref<TagItem[]>([]);

const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

// 筛选条件，仅使用后端真实支持的查询参数
const keyword = ref("");
const tagFilter = ref<number[]>([]);
const visibilityFilter = ref<number | undefined>(undefined);

const columns: TableColumnList = [
  { label: "编号", prop: "id", width: 90, align: "center" },
  {
    label: "标题",
    prop: "title",
    minWidth: 200,
    showOverflowTooltip: true
  },
  { label: "状态", slot: "published", width: 90, align: "center" },
  { label: "可见性", slot: "visibility", width: 110, align: "center" },
  { label: "标签", slot: "tags", minWidth: 180 },
  { label: "题目数", prop: "problemCount", width: 90, align: "center" },
  { label: "最近更新", prop: "updatedAt", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 130,
    fixed: "right",
    align: "center"
  }
];

async function loadTags() {
  try {
    const res = await getTagList();
    tagOptions.value = res.data?.tags ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  }
}

async function onSearch() {
  loading.value = true;
  try {
    const res = await getAdminProblemSetList({
      page: pagination.currentPage,
      pageSize: pagination.pageSize,
      q: keyword.value.trim() || undefined,
      tags: tagFilter.value.length ? tagFilter.value.join(",") : undefined,
      visibility:
        visibilityFilter.value === 0 || visibilityFilter.value === 1
          ? visibilityFilter.value
          : undefined
    });
    dataList.value = res.data?.list ?? [];
    pagination.total = res.data?.total ?? 0;
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function onFilterChange() {
  pagination.currentPage = 1;
  onSearch();
}

function onReset() {
  keyword.value = "";
  tagFilter.value = [];
  visibilityFilter.value = undefined;
  pagination.currentPage = 1;
  onSearch();
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

function goCreate() {
  router.push("/content/problemsets/create");
}

function goEdit(row: ProblemSetRow) {
  router.push(`/content/problemsets/${row.id}/edit`);
}

async function remove(row: ProblemSetRow) {
  try {
    await ElMessageBox.confirm(
      `确认删除题单「${row.title}」？删除后不可恢复。`,
      "删除题单",
      { type: "warning" }
    );
  } catch {
    return;
  }
  if (row._busy) return;
  row._busy = true;
  try {
    await deleteProblemSet(row.id);
    message("已删除", { type: "success" });
    // 删除当前页最后一条数据时回退页码
    if (dataList.value.length === 1 && pagination.currentPage > 1) {
      pagination.currentPage -= 1;
    }
    onSearch();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

onMounted(() => {
  loadTags();
  onSearch();
});
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="onSearch">
      <template #title>
        <div class="flex items-center">
          <span>题单管理</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ pagination.total }} 个（含草稿）
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button type="primary" :icon="Plus" @click="goCreate">
          新建题单
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <el-input
            v-model="keyword"
            class="w-56"
            placeholder="标题关键字"
            clearable
            @keyup.enter="onFilterChange"
            @clear="onFilterChange"
          />
          <el-select
            v-model="tagFilter"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            clearable
            placeholder="按标签筛选"
            class="w-64"
            @change="onFilterChange"
          >
            <el-option
              v-for="tag in tagOptions"
              :key="tag.id"
              :label="tag.name"
              :value="tag.id"
            />
          </el-select>
          <el-select
            v-model="visibilityFilter"
            clearable
            placeholder="可见性"
            class="w-32"
            @change="onFilterChange"
          >
            <el-option label="公开" :value="0" />
            <el-option label="邀请码" :value="1" />
          </el-select>
          <el-button @click="onReset">重置</el-button>
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
          <template #published="{ row }">
            <el-tag :type="PUBLISHED_TAG_TYPES[row.published]" size="small">
              {{ PUBLISHED_LABELS[row.published] }}
            </el-tag>
          </template>
          <template #visibility="{ row }">
            <el-tag :type="VISIBILITY_TAG_TYPES[row.visibility]" size="small">
              {{ VISIBILITY_LABELS[row.visibility] }}
            </el-tag>
          </template>
          <template #tags="{ row }">
            <el-tag
              v-for="tag in row.tags"
              :key="tag.id"
              size="small"
              class="mr-1"
            >
              {{ tag.name }}
            </el-tag>
            <span v-if="!row.tags?.length" class="text-gray-400">—</span>
          </template>
          <template #operation="{ row }">
            <el-button link size="small" type="primary" @click="goEdit(row)">
              编辑
            </el-button>
            <el-button
              link
              size="small"
              type="danger"
              :loading="row._busy"
              @click="remove(row)"
            >
              删除
            </el-button>
          </template>
        </PureTable>
      </template>
    </PureTableBar>
  </div>
</template>
