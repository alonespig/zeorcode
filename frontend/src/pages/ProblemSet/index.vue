<template>
  <div class="problemset-list page-container f-panel">
    <!-- 头部：标题 + 数量 · 搜索 + 标签 + 重置（与题库、比赛同一套排布） -->
    <div class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-5 py-4 max-md:px-3 max-md:py-3">
      <h1 class="text-xl font-semibold text-gray-800">题单</h1>
      <span class="text-xs text-gray-400">共 {{ total }} 个</span>
      <div class="ml-auto flex items-center gap-2.5 max-md:w-full">
        <el-input v-model="q" class="w-64! max-md:w-full!" placeholder="搜索题单标题…" clearable :prefix-icon="Search"
          @keyup.enter="applyFilter" @clear="applyFilter" />
        <el-button type="primary" :icon="Search" @click="applyFilter">搜索</el-button>
        <el-button @click="openTagDialog">
          标签
          <span v-if="selectedTags.length"
            class="ml-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-blue-500 px-1 text-[11px] text-white">{{
              selectedTags.length }}</span>
        </el-button>
        <el-button v-if="q || selectedTags.length" type="info" plain :icon="Refresh" @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 已选标签 -->
    <div v-if="selectedTagObjs.length"
      class="flex flex-wrap items-center gap-2 border-b border-gray-100 px-5 py-3 text-[12.5px]">
      <span class="text-[14px] font-medium text-gray-700">已选标签：</span>
      <el-tag v-for="t in selectedTagObjs" :key="t.id" effect="light" :style="tagStyle(t)" closable
        @close="removeTag(t.id)">
        {{ t.name }}
      </el-tag>
      <span class="cursor-pointer text-gray-400 hover:text-red-500" @click="clearTags">清除</span>
    </div>

    <el-table :data="list" style="width: 100%">
      <el-table-column prop="id" label="#" width="90" align="center" />

      <el-table-column label="题单" min-width="260">
        <template #default="{ row }">
          <router-link class="font-medium text-blue-500 hover:text-blue-400"
            :to="{ name: 'ProblemSetDetail', params: { id: row.id } }">
            {{ row.title }}
          </router-link>
          <!-- 非公开题单在标题右边挂一把锁，替代单独的「可见性」列 -->
          <el-tooltip v-if="row.visibility === 1" content="需要邀请码" placement="top">
            <el-icon class="ml-1.5 align-middle text-gray-400" :size="14">
              <Lock />
            </el-icon>
          </el-tooltip>
        </template>
      </el-table-column>

      <el-table-column label="标签" min-width="180">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1.5">
            <el-tag v-for="t in row.tags" :key="t.id" class="!cursor-pointer" role="button"
              :effect="selectedTags.includes(t.id) ? 'dark' : 'light'"
              :type="selectedTags.includes(t.id) ? 'primary' : undefined"
              :style="selectedTags.includes(t.id) ? { '--el-tag-font-size': TAG_FONT_SIZE } : tagStyle(t)"
              @click="setTagFilter(t.id)">
              {{ t.name }}
            </el-tag>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="题目数" width="90" align="center">
        <template #default="{ row }">
          {{ row.problemCount }}
        </template>
      </el-table-column>

      <el-table-column label="我的进度" width="150" align="center">
        <template #default="{ row }">
          <!-- 未登录时后端不返回 solvedCount，整列留空 -->
          <div v-if="row.solvedCount != null && row.problemCount > 0" class="flex items-center gap-2">
            <div class="h-1.5 w-[74px] overflow-hidden rounded bg-gray-100">
              <div class="h-full rounded bg-emerald-500" :style="{ width: percent(row) + '%' }"></div>
            </div>
            <span class="w-9 text-[12.5px] tabular-nums"
              :class="row.solvedCount === row.problemCount ? 'font-medium text-emerald-600' : 'text-gray-500'">
              {{ row.solvedCount }}/{{ row.problemCount }}
            </span>
          </div>
          <span v-else class="text-gray-300">—</span>
        </template>
      </el-table-column>

      <el-table-column label="作者" width="120" align="center">
        <template #default="{ row }">
          {{ row.author || '-' }}
        </template>
      </el-table-column>

      <el-table-column label="最近更新" width="170" align="center">
        <template #default="{ row }">
          <span class="tabular-nums">{{ row.updatedAt }}</span>
        </template>
      </el-table-column>

      <template #empty>
        <el-empty description="暂无题单" :image-size="80" />
      </template>
    </el-table>

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination size="large" :page-size="pageSize" :pager-count="16" v-model:current-page="page"
        layout="prev, pager, next" :total="total" @current-change="pageChange" />
    </div>

    <!-- 标签选择弹窗（与题库一致） -->
    <el-dialog v-model="tagDialogVisible" title="选择标签" width="640px" @open="onDialogOpen">
      <el-input v-model="tagSearch" placeholder="搜索标签…" clearable :prefix-icon="Search" class="mb-3.5" />
      <div class="mb-2 text-xs text-gray-400">全部标签（点击选择，可多选）</div>
      <div class="flex max-h-[240px] flex-wrap gap-2.5 overflow-y-auto py-1">
        <span v-for="t in filteredTags" :key="t.id" @click="toggleTemp(t.id)"
          class="cursor-pointer select-none rounded-full border px-3 py-1 text-[13px] transition-colors" :class="tempSelected.includes(t.id)
            ? 'border-blue-500 bg-blue-500 text-white'
            : 'border-gray-300 bg-white text-gray-600 hover:border-blue-300 hover:text-blue-500'">
          {{ t.name }}
        </span>
        <span v-if="!filteredTags.length" class="py-3 text-sm text-gray-400">没有匹配的标签</span>
      </div>
      <template #footer>
        <div class="flex items-center">
          <span class="text-sm text-gray-500">已选 {{ tempSelected.length }} 个</span>
          <span class="ml-3 cursor-pointer text-sm text-red-400 hover:text-red-500"
            @click="tempSelected = []">清空</span>
          <span class="ml-auto flex gap-2.5">
            <el-button @click="tagDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="confirmTags">确定</el-button>
          </span>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Search, Refresh, Lock } from "@element-plus/icons-vue";
import { getProblemSetList } from "@/api/problemset";
import { getTags } from "@/api/problems";
import { tagStyle, TAG_FONT_SIZE } from "@/utils/tag";

const route = useRoute();
const router = useRouter();

const list = ref([]);
const total = ref(0);
const pageSize = 20;

// 筛选草稿（v-model），生效以 URL 为准 —— 与题库同一套做法，支持前进/后退/分享链接
const q = ref(route.query.q || "");
const selectedTags = ref(route.query.tags ? String(route.query.tags).split(",").map(Number) : []);
const page = ref(Number(route.query.page) || 1);

const tagsList = ref([]);
const selectedTagObjs = computed(() =>
  selectedTags.value.map((id) => tagsList.value.find((t) => t.id === id)).filter(Boolean)
);

const percent = (row) => Math.round((row.solvedCount / row.problemCount) * 100);

const syncFromUrl = () => {
  q.value = route.query.q || "";
  selectedTags.value = route.query.tags ? String(route.query.tags).split(",").map(Number) : [];
  page.value = Number(route.query.page) || 1;
};

const buildQuery = () => {
  const query = { page: 1, pageSize };
  if (q.value) query.q = q.value;
  if (selectedTags.value.length) query.tags = selectedTags.value.join(",");
  return query;
};

const applyFilter = () => router.push({ query: buildQuery() });
const pageChange = () => router.push({ query: { ...route.query, page: page.value } });

// 点「重置」：清空搜索词与标签筛选，回到第 1 页
const handleReset = () => {
  q.value = "";
  selectedTags.value = [];
  router.push({ query: { page: 1, pageSize } });
};

const removeTag = (id) => {
  selectedTags.value = selectedTags.value.filter((t) => t !== id);
  applyFilter();
};
const clearTags = () => {
  selectedTags.value = [];
  applyFilter();
};
// 点列表里的标签：清掉之前的，只按这一个筛选
const setTagFilter = (id) => {
  selectedTags.value = [id];
  applyFilter();
};

// ---------- 标签弹窗 ----------
const tagDialogVisible = ref(false);
const tagSearch = ref("");
const tempSelected = ref([]);
const filteredTags = computed(() => {
  const kw = tagSearch.value.trim();
  return kw ? tagsList.value.filter((t) => t.name.includes(kw)) : tagsList.value;
});
const openTagDialog = () => { tagDialogVisible.value = true; };
const onDialogOpen = () => {
  tagSearch.value = "";
  tempSelected.value = [...selectedTags.value];
};
const toggleTemp = (id) => {
  const i = tempSelected.value.indexOf(id);
  if (i === -1) tempSelected.value.push(id);
  else tempSelected.value.splice(i, 1);
};
const confirmTags = () => {
  selectedTags.value = [...tempSelected.value];
  tagDialogVisible.value = false;
  applyFilter();
};

// ---------- 数据 ----------
const getList = async () => {
  try {
    const res = await getProblemSetList({
      page: Number(route.query.page) || 1,
      pageSize,
      q: route.query.q || undefined,
      tags: route.query.tags || undefined,
    });
    list.value = res.data?.list || [];
    total.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  }
};

const fetchTags = async () => {
  try {
    const res = await getTags();
    tagsList.value = res.data?.tags || [];
  } catch (err) {
    console.error(err);
  }
};

watch(() => route.query, () => { syncFromUrl(); getList(); });

onMounted(() => {
  getList();
  fetchTags();
});
</script>

<style scoped>
.problemset-list {
  padding-top: 8px;
}
</style>
