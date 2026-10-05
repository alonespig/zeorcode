<script setup lang="ts">
import { computed, ref } from "vue";
import { Search } from "@element-plus/icons-vue";
import {
  getProblemList,
  getTagList,
  type ProblemItem,
  type TagItem
} from "@/api/admin/problems";

defineOptions({ name: "ProblemPickerDialog" });

const props = defineProps<{
  /** 已在题单中的对外题号，用于禁用并标记为已选 */
  existingIds: string[];
}>();

const emit = defineEmits<{
  add: [problems: { id: string; name: string }[]];
}>();

const visible = defineModel<boolean>({ default: false });

const difficultyText: Record<number, string> = {
  1: "简单",
  2: "中等",
  3: "困难"
};

const keyword = ref("");
const tagFilter = ref<number[]>([]);
const tags = ref<TagItem[]>([]);
const list = ref<ProblemItem[]>([]);
const page = ref(1);
const pageSize = 8;
const total = ref(0);
const loading = ref(false);
const picked = ref<{ id: string; name: string }[]>([]);

const existingIdSet = computed(() => new Set(props.existingIds));

const isExisting = (id: string) => existingIdSet.value.has(id);
const isPicked = (id: string) => picked.value.some(p => p.id === id);

function togglePick(row: ProblemItem) {
  if (isExisting(row.id)) return;
  const index = picked.value.findIndex(p => p.id === row.id);
  if (index === -1) picked.value.push({ id: row.id, name: row.name });
  else picked.value.splice(index, 1);
}

async function loadList(targetPage = 1) {
  page.value = targetPage;
  loading.value = true;
  try {
    const res = await getProblemList({
      page: targetPage,
      pageSize,
      q: keyword.value.trim() || undefined,
      tags: tagFilter.value.length ? tagFilter.value.join(",") : undefined
    });
    list.value = res.data?.list ?? [];
    total.value = res.data?.total ?? 0;
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

async function loadTags() {
  if (tags.value.length) return;
  try {
    const res = await getTagList();
    tags.value = res.data?.tags ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  }
}

function onSearch() {
  loadList(1);
}

function onPageChange(p: number) {
  loadList(p);
}

function handleOpen() {
  keyword.value = "";
  tagFilter.value = [];
  picked.value = [];
  loadList(1);
  loadTags();
}

function confirm() {
  if (picked.value.length) {
    emit(
      "add",
      picked.value.map(p => ({ id: p.id, name: p.name }))
    );
  }
  visible.value = false;
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="从题库选择"
    width="min(960px, calc(100vw - 32px))"
    top="8vh"
    @open="handleOpen"
  >
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <el-input
        v-model="keyword"
        class="w-64"
        placeholder="搜索题目名称或题号"
        clearable
        @keyup.enter="onSearch"
        @clear="onSearch"
      />
      <el-select
        v-model="tagFilter"
        multiple
        filterable
        collapse-tags
        collapse-tags-tooltip
        clearable
        placeholder="按算法标签过滤"
        class="w-72"
        @change="onSearch"
      >
        <el-option
          v-for="tag in tags"
          :key="tag.id"
          :label="tag.name"
          :value="tag.id"
        />
      </el-select>
      <el-button type="primary" :icon="Search" @click="onSearch">
        搜索
      </el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="list"
      row-key="id"
      border
      max-height="420"
    >
      <el-table-column width="50" align="center">
        <template #default="{ row }">
          <el-checkbox
            :model-value="isExisting(row.id) || isPicked(row.id)"
            :disabled="isExisting(row.id)"
            @change="togglePick(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="id" label="题号" width="110" />
      <el-table-column
        prop="name"
        label="题目"
        min-width="220"
        show-overflow-tooltip
      />
      <el-table-column label="难度" width="90" align="center">
        <template #default="{ row }">
          {{ difficultyText[row.difficulty] || "—" }}
        </template>
      </el-table-column>
      <el-table-column label="标签" min-width="180">
        <template #default="{ row }">
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
      </el-table-column>
    </el-table>

    <div class="mt-4 flex justify-center">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="onPageChange"
      />
    </div>

    <template #footer>
      <span class="mr-auto text-sm text-gray-500">
        已选 {{ picked.length }} 道
      </span>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="confirm">确定添加</el-button>
    </template>
  </el-dialog>
</template>
