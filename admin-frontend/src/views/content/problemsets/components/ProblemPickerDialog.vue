<script setup lang="ts">
import { computed, ref } from "vue";
import { Search, View } from "@element-plus/icons-vue";
import {
  getProblemList,
  getTagList,
  type ProblemItem,
  type TagItem
} from "@/api/admin/problems";
import ProblemPreviewDialog from "./ProblemPreviewDialog.vue";

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

const difficultyClass = (difficulty: number) =>
  `difficulty-${difficulty || "unknown"}`;

const keyword = ref("");
const tagFilter = ref<number[]>([]);
const tags = ref<TagItem[]>([]);
const list = ref<ProblemItem[]>([]);
const page = ref(1);
const pageSize = 8;
const total = ref(0);
const loading = ref(false);
const picked = ref<{ id: string; name: string }[]>([]);
const previewVisible = ref(false);
const previewProblemId = ref("");

const existingIdSet = computed(() => new Set(props.existingIds));

const isExisting = (id: string) => existingIdSet.value.has(id);
const isPicked = (id: string) => picked.value.some(p => p.id === id);

function togglePick(row: ProblemItem) {
  if (isExisting(row.id)) return;
  const index = picked.value.findIndex(p => p.id === row.id);
  if (index === -1) picked.value.push({ id: row.id, name: row.name });
  else picked.value.splice(index, 1);
}

function openPreview(id: string) {
  previewProblemId.value = id;
  previewVisible.value = true;
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
  previewVisible.value = false;
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
    destroy-on-close
    class="problem-picker-dialog"
    @open="handleOpen"
    @close="previewVisible = false"
  >
    <div class="picker-search">
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

    <div v-loading="loading" class="picker-table-wrap">
      <table class="picker-table">
        <colgroup>
          <col style="width: 52px" />
          <col style="width: 110px" />
          <col style="width: 240px" />
          <col style="width: 90px" />
          <col />
          <col style="width: 64px" />
        </colgroup>
        <thead>
          <tr>
            <th />
            <th class="center">#</th>
            <th class="left">题目名称</th>
            <th class="center">难度</th>
            <th class="left">标签</th>
            <th class="center">预览</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in list"
            :key="row.id"
            class="pick-row"
            @click="togglePick(row)"
          >
            <td class="center">
              <el-checkbox
                :model-value="isExisting(row.id) || isPicked(row.id)"
                :disabled="isExisting(row.id)"
                @click.stop
                @change="togglePick(row)"
              />
            </td>
            <td class="center problem-id">{{ row.id }}</td>
            <td class="left problem-name">{{ row.name }}</td>
            <td class="center">
              <span class="difficulty" :class="difficultyClass(row.difficulty)">
                {{ difficultyText[row.difficulty] || "-" }}
              </span>
            </td>
            <td class="left tag-cell">
              <el-tag
                v-for="tag in row.tags"
                :key="tag.id"
                size="small"
                type="primary"
              >
                {{ tag.name }}
              </el-tag>
              <span v-if="!row.tags?.length" class="muted-text">—</span>
            </td>
            <td class="center">
              <el-tooltip content="查看题目内容" placement="top">
                <el-button
                  text
                  type="primary"
                  :icon="View"
                  :aria-label="`查看题目：${row.name}`"
                  @click.stop="openPreview(row.id)"
                />
              </el-tooltip>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!loading && !list.length" class="picker-empty">
        没有找到符合条件的题目
      </div>
    </div>

    <div class="picker-pagination">
      <el-pagination
        background
        size="small"
        layout="prev, pager, next"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="onPageChange"
      />
    </div>

    <template #footer>
      <div class="picker-footer">
        <span class="picked-summary">已选 {{ picked.length }} 道</span>
        <div class="footer-actions">
          <el-button @click="visible = false">取消</el-button>
          <el-button type="primary" @click="confirm">确定添加</el-button>
        </div>
      </div>
    </template>
  </el-dialog>

  <ProblemPreviewDialog
    v-model="previewVisible"
    :problem-id="previewProblemId"
  />
</template>

<style scoped>
.picker-search {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-bottom: 14px;
}

.keyword-input {
  width: 280px;
}

.tag-filter {
  width: 300px;
}

.picker-table-wrap {
  min-height: 180px;
  overflow-x: auto;
}

.picker-table {
  width: 100%;
  font-size: 14px;
  color: #1f2937;
  table-layout: fixed;
  border-collapse: collapse;
  background: var(--el-bg-color);
}

.picker-table th,
.picker-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f3f4f6;
}

.picker-table th {
  height: 46px;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  letter-spacing: 0.025em;
  background: #f7f8fa;
}

.picker-table .center {
  text-align: center;
}

.picker-table .left {
  text-align: left;
}

.picker-table tbody tr {
  transition: background-color 0.15s ease;
}

.picker-table tbody tr:nth-child(even) {
  background: #fafbfc;
}

.picker-table tbody tr:hover {
  background: #eff6ff;
}

.picker-table .pick-row {
  cursor: pointer;
}

.picker-table .problem-id,
.picker-table .problem-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.picker-table .problem-id {
  font-variant-numeric: tabular-nums;
}

.picker-table .tag-cell {
  white-space: normal;
}

.picker-table .tag-cell :deep(.el-tag) {
  margin: 2px 5px 2px 0;
}

.difficulty {
  font-weight: 500;
}

.difficulty-1 {
  color: #059669;
}

.difficulty-2 {
  color: #d97706;
}

.difficulty-3 {
  color: #dc2626;
}

.difficulty-unknown {
  color: var(--el-text-color-placeholder);
}

.muted-text {
  color: var(--el-text-color-secondary);
}

.picker-empty {
  padding: 48px 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
  text-align: center;
}

.picker-pagination {
  display: flex;
  justify-content: center;
  margin: 22px 0 10px;
}

.picker-pagination :deep(.el-pagination) {
  --el-pagination-button-bg-color: #f1f3f5;
  --el-pagination-button-disabled-bg-color: #f6f7f8;
  --el-pagination-border-radius: 2px;
  --el-pagination-item-gap: 0;
}

.picker-pagination :deep(.btn-prev),
.picker-pagination :deep(.btn-next),
.picker-pagination :deep(.el-pager li) {
  margin: 0 6px;
  font-size: 13px;
}

.picker-footer {
  display: flex;
  align-items: center;
  width: 100%;
}

.picked-summary {
  margin-right: auto;
  font-size: 13px;
  color: #4b5563;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

@media (width <= 680px) {
  .picker-search {
    flex-direction: column;
    align-items: stretch;
  }

  .keyword-input,
  .tag-filter {
    width: 100%;
  }
}
</style>
