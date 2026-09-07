<template>
  <div class="problemset page-container">
    <div class="overflow-hidden rounded-lg border border-gray-200 bg-white">

      <!-- 头部：标题 + 题数 · 搜索 + 标签 + 创建 -->
      <div class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-5 py-4">
        <h1 class="text-xl font-semibold text-gray-800">题库</h1>
        <span class="text-xs text-gray-400">共 {{ total }} 题</span>
        <div class="ml-auto flex items-center gap-2.5">
          <el-input v-model="q" class="w-64!" placeholder="搜索题目标题…" clearable :prefix-icon="Search"
            @keyup.enter="applyFilter" @clear="applyFilter" />
          <el-button type="primary" :icon="Search" @click="applyFilter">搜索</el-button>
          <el-button @click="openTagDialog">
            标签
            <span v-if="selectedTags.length"
              class="ml-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-blue-500 px-1 text-[11px] text-white">{{
                selectedTags.length }}</span>
          </el-button>
          <el-button v-if="q || selectedTags.length" type="info" plain :icon="Refresh" @click="handleReset">重置</el-button>
          <el-button v-permission="['admin']" :icon="Edit" @click="goCreate">创建题目</el-button>
        </div>
      </div>

      <!-- 已选标签 -->
      <div v-if="selectedTagObjs.length"
        class="flex flex-wrap items-center gap-2 border-b border-gray-100 px-5 py-3 text-[12.5px]">
        <span class="text-[14px] font-medium text-gray-700">已选标签：</span>
        <el-tag
          v-for="t in selectedTagObjs"
          :key="t.id"
          effect="light"
          :style="tagStyle(t)"
          closable
          @close="removeTag(t.id)"
        >
          {{ t.name }}
        </el-tag>
        <span class="cursor-pointer text-gray-400 hover:text-red-500" @click="clearTags">清除</span>
      </div>

      <!-- 题目表 -->
      <el-table :data="problemList" style="width: 100%">
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.status === 1" color="#2f9e44" :size="17">
              <Select />
            </el-icon>
            <el-icon v-else-if="row.status === 2" color="#e5484d" :size="17">
              <CloseBold />
            </el-icon>
          </template>
        </el-table-column>

        <el-table-column prop="id" label="#" width="100" align="center" />

        <el-table-column label="题目" width="280">
          <template #default="{ row }">
            <router-link class="font-[Arial,'Noto_Sans_SC',sans-serif] font-medium text-blue-500 hover:text-blue-400"
              :to="`/problem/${row.id}`">
              {{ row.name }}
            </router-link>
            <span v-if="row.hidden"
              class="ml-4 inline-flex items-center
               rounded border border-gray-300 bg-gray-50 px-1.5
               align-middle text-[11px] text-gray-500">
              隐藏
            </span>
          </template>
        </el-table-column>

        <el-table-column label="难度" width="90" align="center">
          <template #default="{ row }">
            <span class="text-[14px] font-medium"
              :class="diffTextClass(row.difficulty)">
              {{ DIFF_TEXT[row.difficulty] || '-' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column min-width="100">
          <template #header>
            <div class="flex items-center gap-2">
              <span>算法标签</span>
              <el-switch class="h-5! leading-5!" width="45px" v-model="tagsShow"
                style="--el-switch-off-color: #d9d9d9" />
            </div>
          </template>
          <template #default="{ row }">
            <div v-if="tagsShow" class="flex flex-wrap gap-1.5">
              <el-tag
                v-for="t in row.tags"
                :key="t.id"
                class="!cursor-pointer"
                role="button"
                :effect="selectedTags.includes(t.id) ? 'dark' : 'light'"
                :type="selectedTags.includes(t.id) ? 'primary' : undefined"
                :style="selectedTags.includes(t.id) ? { '--el-tag-font-size': TAG_FONT_SIZE } : tagStyle(t)"
                @click="setTagFilter(t.id)"
              >
                {{ t.name }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="通过率" width="180" align="center">
          <template #default="{ row }">
            <PassRateBar :accepted-count="row.acceptedCount" :submit-count="row.submitCount" />
          </template>
        </el-table-column>

        <el-table-column label="提交" width="100" align="center">
          <template #default="{ row }">
            <span class="text-[14px] tabular-nums text-gray-700">
              {{ row.submitCount > 0 ? fmtCount(row.submitCount) : "-" }}
            </span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="f-pagination">
        <el-pagination size="large" :page-size="pageSize" :pager-count="16" v-model:current-page="page"
          @current-change="pageChange" layout="prev, pager, next" :total="total" />
      </div>
    </div>

    <!-- 标签选择弹窗 -->
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
          <span class="ml-3 cursor-pointer text-sm text-red-400 hover:text-red-500" @click="tempSelected = []">清空</span>
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
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProblemList, getTags } from '@/api/problems'
import { Search, Select, CloseBold, Edit, Refresh } from '@element-plus/icons-vue'
import PassRateBar from '@/components/problem/PassRateBar.vue'
// 标签配色抽到 utils/tag.js，题库与题单共用，避免两处各写一份导致视觉不一致
import { tagStyle, TAG_FONT_SIZE } from '@/utils/tag'

const route = useRoute()
const router = useRouter()

const goCreate = () => router.push({ name: 'CreateProblem' })

const DIFF_TEXT = { 1: '简单', 2: '中等', 3: '困难' }
const diffTextClass = (d) => ({ 1: 'text-emerald-600', 2: 'text-amber-500', 3: 'text-red-500' }[d] || 'text-gray-400')
// 提交数：>=1000 显示 x.xk
const fmtCount = (n) => {
  n = n || 0
  return n >= 1000 ? (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k' : String(n)
}

const problemList = ref([])
const total = ref(0)
const pageSize = 20
const tagsShow = ref(true) // 「算法标签」列显隐开关

// 筛选草稿（v-model），生效以 URL 为准
const q = ref(route.query.q || '')
const selectedTags = ref(route.query.tags ? String(route.query.tags).split(',').map(Number) : [])
const page = ref(Number(route.query.page) || 1)

// 全部标签
const tagsList = ref([])
const selectedTagObjs = computed(() =>
  selectedTags.value.map((id) => tagsList.value.find((t) => t.id === id)).filter(Boolean)
)

// 从 URL 同步草稿（支持前进/后退/重置）
const syncFromUrl = () => {
  q.value = route.query.q || ''
  selectedTags.value = route.query.tags ? String(route.query.tags).split(',').map(Number) : []
  page.value = Number(route.query.page) || 1
}

const buildQuery = () => {
  const query = { page: 1, pageSize }
  if (q.value) query.q = q.value
  if (selectedTags.value.length) query.tags = selectedTags.value.join(',')
  return query
}

const applyFilter = () => router.push({ query: buildQuery() })
const pageChange = () => router.push({ query: { ...route.query, page: page.value } })

// 点“重置”：清空搜索词与标签筛选，回到第 1 页
const handleReset = () => {
  q.value = ''
  selectedTags.value = []
  router.push({ query: { page: 1, pageSize } })
}

const removeTag = (id) => {
  selectedTags.value = selectedTags.value.filter((t) => t !== id)
  applyFilter()
}
const clearTags = () => {
  selectedTags.value = []
  applyFilter()
}
// 点列表里的标签：清掉之前的标签，只按这一个标签筛选
const setTagFilter = (id) => {
  selectedTags.value = [id]
  applyFilter()
}

// ---------- 标签弹窗 ----------
const tagDialogVisible = ref(false)
const tagSearch = ref('')
const tempSelected = ref([])
const filteredTags = computed(() => {
  const kw = tagSearch.value.trim()
  return kw ? tagsList.value.filter((t) => t.name.includes(kw)) : tagsList.value
})
const openTagDialog = () => { tagDialogVisible.value = true }
const onDialogOpen = () => {
  tagSearch.value = ''
  tempSelected.value = [...selectedTags.value]
}
const toggleTemp = (id) => {
  const i = tempSelected.value.indexOf(id)
  if (i === -1) tempSelected.value.push(id)
  else tempSelected.value.splice(i, 1)
}
const confirmTags = () => {
  selectedTags.value = [...tempSelected.value]
  tagDialogVisible.value = false
  applyFilter()
}

// ---------- 数据 ----------
const getList = async () => {
  const res = await getProblemList({
    page: Number(route.query.page) || 1,
    pageSize,
    q: route.query.q || undefined,
    tags: route.query.tags || undefined,
  })
  problemList.value = res.data.list || []
  total.value = res.data.total || 0
}

const fetchTags = async () => {
  const res = await getTags()
  tagsList.value = res.data.tags || []
}

watch(() => route.query, () => { syncFromUrl(); getList() })

onMounted(() => {
  getList()
  fetchTags()
})
</script>

<style scoped>
.problemset.page-container {
  --table-cell-padding-y: 14px;
  max-width: 1360px;
  padding-top: 8px;
}

</style>
