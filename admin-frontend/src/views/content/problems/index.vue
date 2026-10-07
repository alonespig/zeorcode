<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { Plus } from "@element-plus/icons-vue";
import type { PaginationProps } from "@pureadmin/table";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  getProblemList,
  getTagList,
  type ProblemItem,
  type TagItem
} from "@/api/admin/problems";

defineOptions({ name: "AdminProblems" });

const DIFFICULTY_LABELS: Record<number, string> = {
  1: "简单",
  2: "中等",
  3: "困难"
};
const DIFFICULTY_TYPES: Record<number, "success" | "warning" | "danger"> = {
  1: "success",
  2: "warning",
  3: "danger"
};

const router = useRouter();
const dataList = ref<ProblemItem[]>([]);
const tagOptions = ref<TagItem[]>([]);
const loading = ref(false);
const keyword = ref("");
const difficulty = ref<number>();
const selectedTags = ref<number[]>([]);

const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

const columns: TableColumnList = [
  { label: "题号", prop: "id", width: 120, align: "center" },
  {
    label: "题目",
    prop: "name",
    minWidth: 220,
    showOverflowTooltip: true
  },
  { label: "难度", slot: "difficulty", width: 100, align: "center" },
  { label: "标签", slot: "tags", minWidth: 180 },
  { label: "类型", slot: "source", width: 110, align: "center" },
  { label: "可见性", slot: "visibility", width: 100, align: "center" },
  { label: "通过率", slot: "acceptance", width: 110, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 150,
    fixed: "right",
    align: "center"
  }
];

function acceptance(row: ProblemItem) {
  if (!row.submitCount) return "—";
  return `${((row.acceptedCount / row.submitCount) * 100).toFixed(1)}%`;
}

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
    const res = await getProblemList({
      page: pagination.currentPage,
      pageSize: pagination.pageSize,
      q: keyword.value.trim() || undefined,
      difficulty: difficulty.value,
      tags: selectedTags.value.length
        ? selectedTags.value.join(",")
        : undefined,
      order: "latest"
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
  difficulty.value = undefined;
  selectedTags.value = [];
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
          <span>题目管理</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ pagination.total }} 道（含隐藏题）
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button
          type="primary"
          :icon="Plus"
          @click="router.push('/content/problems/create')"
        >
          新建题目
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <el-input
            v-model="keyword"
            class="w-56"
            clearable
            placeholder="题号或题目名称"
            @keyup.enter="applyFilter"
            @clear="applyFilter"
          />
          <el-select
            v-model="difficulty"
            class="w-32"
            clearable
            placeholder="难度"
            @change="applyFilter"
          >
            <el-option label="简单" :value="1" />
            <el-option label="中等" :value="2" />
            <el-option label="困难" :value="3" />
          </el-select>
          <el-select
            v-model="selectedTags"
            class="w-64"
            multiple
            filterable
            clearable
            collapse-tags
            collapse-tags-tooltip
            placeholder="按标签筛选"
            @change="applyFilter"
          >
            <el-option
              v-for="tag in tagOptions"
              :key="tag.id"
              :label="tag.name"
              :value="tag.id"
            />
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
          <template #difficulty="{ row }">
            <el-tag :type="DIFFICULTY_TYPES[row.difficulty]" size="small">
              {{ DIFFICULTY_LABELS[row.difficulty] ?? "未知" }}
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
          <template #source="{ row }">
            <el-tag :type="row.oj ? 'warning' : 'info'" size="small">
              {{ row.oj || "本地题" }}
            </el-tag>
          </template>
          <template #visibility="{ row }">
            <el-tag :type="row.hidden ? 'warning' : 'success'" size="small">
              {{ row.hidden ? "隐藏" : "公开" }}
            </el-tag>
          </template>
          <template #acceptance="{ row }">
            {{ acceptance(row) }}
          </template>
          <template #operation="{ row }">
            <el-button
              link
              size="small"
              type="primary"
              @click="router.push(`/content/problems/${row.id}/edit`)"
            >
              编辑
            </el-button>
            <el-button
              link
              size="small"
              type="primary"
              :disabled="Boolean(row.oj)"
              @click="router.push(`/content/problems/${row.id}/testdata`)"
            >
              测试数据
            </el-button>
          </template>
        </PureTable>
      </template>
    </PureTableBar>
  </div>
</template>
