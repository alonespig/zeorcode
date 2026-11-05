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
const pageSize = 10;
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

function rowClassName({ row }: { row: ProblemItem }) {
  if (isExisting(row.id)) return "is-existing";
  if (isPicked(row.id)) return "is-picked";
  return "";
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="从题库选择"
    width="min(960px, calc(100vw - 32px))"
    top="6vh"
    destroy-on-close
    class="problem-picker-dialog"
    @open="handleOpen"
  >
    <div class="picker-toolbar">
      <el-input
        v-model="keyword"
        class="keyword-input"
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
        class="tag-filter"
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

    <div class="picker-table-wrap">
      <el-table
        v-loading="loading"
        :data="list"
        :row-class-name="rowClassName"
        row-key="id"
        border
        height="430"
        empty-text="没有找到符合条件的题目"
        @row-click="togglePick"
      >
        <el-table-column width="52" align="center">
          <template #default="{ row }">
            <el-checkbox
              :model-value="isExisting(row.id) || isPicked(row.id)"
              :disabled="isExisting(row.id)"
              @click.stop
              @change="togglePick(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="id" label="题号" width="120" />
        <el-table-column
          prop="name"
          label="题目名称"
          min-width="240"
          show-overflow-tooltip
        />
        <el-table-column label="难度" width="90" align="center">
          <template #default="{ row }">
            {{ difficultyText[row.difficulty] || "—" }}
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="tag in row.tags"
              :key="tag.id"
              size="small"
              class="tag-item"
            >
              {{ tag.name }}
            </el-tag>
            <span v-if="!row.tags?.length" class="muted-text">—</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="picker-pagination">
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
      <div class="picker-footer">
        <span class="picked-summary">本次已选 {{ picked.length }} 道</span>
        <div class="footer-actions">
          <el-button @click="visible = false">取消</el-button>
          <el-button type="primary" @click="confirm">确定添加</el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.picker-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  padding: 12px;
  margin-bottom: 14px;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}

.keyword-input {
  width: 280px;
}

.tag-filter {
  flex: 1;
  min-width: 260px;
}

.picker-table-wrap {
  min-height: 430px;
}

.picker-table-wrap :deep(.el-table__row) {
  cursor: pointer;
}

.picker-table-wrap :deep(.el-table__row.is-picked) {
  background: var(--el-color-primary-light-9);
}

.picker-table-wrap :deep(.el-table__row.is-existing) {
  color: var(--el-text-color-placeholder);
  cursor: not-allowed;
  background: var(--el-fill-color-light);
}

.tag-item {
  margin: 2px 4px 2px 0;
}

.muted-text,
.picked-summary {
  color: var(--el-text-color-secondary);
}

.picker-pagination {
  display: flex;
  justify-content: center;
  min-height: 32px;
  margin-top: 16px;
}

.picker-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.picked-summary {
  font-size: 14px;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

@media (width <= 720px) {
  .keyword-input,
  .tag-filter {
    flex: none;
    width: 100%;
  }
}
</style>
