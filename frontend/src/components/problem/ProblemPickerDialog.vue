<script setup>
import { computed, ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { getProblem, getProblemList, getTags } from '@/api/problems'

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true,
  },
  existingProblemIds: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update:modelValue', 'add'])

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
const existingIdSet = computed(() => new Set(props.existingProblemIds.map(String)))

const query = ref('')
const selectedTags = ref([])
const tags = ref([])
const list = ref([])
const page = ref(1)
const pageSize = 8
const total = ref(0)
const picked = ref([])
const loading = ref(false)
const adding = ref(false)

const difficultyText = { 1: '简单', 2: '中等', 3: '困难' }
const difficultyClass = (difficulty) => `difficulty-${difficulty || 'unknown'}`
const isExisting = (id) => existingIdSet.value.has(String(id))
const isPicked = (id) => picked.value.includes(String(id))

const togglePick = (id) => {
  if (isExisting(id)) return
  const normalizedId = String(id)
  const index = picked.value.indexOf(normalizedId)
  if (index === -1) picked.value.push(normalizedId)
  else picked.value.splice(index, 1)
}

const loadList = async (targetPage = 1) => {
  page.value = targetPage
  loading.value = true
  try {
    const res = await getProblemList({
      page: targetPage,
      pageSize,
      q: query.value.trim() || undefined,
      tags: selectedTags.value.length ? selectedTags.value.join(',') : undefined,
    })
    list.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const loadTags = async () => {
  if (tags.value.length) return
  try {
    const res = await getTags()
    tags.value = res.data.tags || []
  } catch (err) {
    console.error(err)
  }
}

const handleOpen = () => {
  query.value = ''
  selectedTags.value = []
  picked.value = props.existingProblemIds.map(String)
  loadList(1)
  loadTags()
}

const confirm = async () => {
  const ids = picked.value.filter((id) => !existingIdSet.value.has(id))
  if (!ids.length) {
    visible.value = false
    return
  }
  adding.value = true
  try {
    const results = await Promise.all(ids.map((id) => getProblem(id)))
    emit('add', results.map((result) => result.data))
    visible.value = false
  } catch (err) {
    console.error(err)
  } finally {
    adding.value = false
  }
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
    <div class="picker-search">
      <el-input
        v-model="query"
        class="keyword-input"
        placeholder="搜索题目名称"
        clearable
        @keyup.enter="loadList(1)"
        @clear="loadList(1)"
      />
      <el-select
        v-model="selectedTags"
        class="tag-filter"
        multiple
        filterable
        collapse-tags
        collapse-tags-tooltip
        clearable
        placeholder="按算法标签过滤"
        @change="loadList(1)"
      >
        <el-option v-for="tag in tags" :key="tag.id" :label="tag.name" :value="tag.id" />
      </el-select>
      <el-button type="primary" :icon="Search" @click="loadList(1)">搜索</el-button>
    </div>

    <div v-loading="loading" class="picker-table-wrap">
      <table class="f-table picker-table">
        <colgroup>
          <col style="width: 52px">
          <col style="width: 110px">
          <col style="width: 240px">
          <col style="width: 90px">
          <col>
        </colgroup>
        <thead>
          <tr>
            <th></th>
            <th class="center">#</th>
            <th class="left">题目名称</th>
            <th class="center">难度</th>
            <th class="left">标签</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id" class="pick-row" @click="togglePick(item.id)">
            <td class="center">
              <el-checkbox
                :model-value="isPicked(item.id)"
                :disabled="isExisting(item.id)"
                @click.stop="togglePick(item.id)"
              />
            </td>
            <td class="center problem-id">{{ item.id }}</td>
            <td class="left problem-name">{{ item.name }}</td>
            <td class="center">
              <span class="difficulty" :class="difficultyClass(item.difficulty)">
                {{ difficultyText[item.difficulty] || '-' }}
              </span>
            </td>
            <td class="left tag-cell">
              <el-tag v-for="tag in item.tags || []" :key="tag.id" size="small" type="primary">
                {{ tag.name }}
              </el-tag>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="picker-pagination">
      <el-pagination
        size="small"
        layout="prev, pager, next"
        :page-size="pageSize"
        :total="total"
        :current-page="page"
        @current-change="loadList"
      />
    </div>

    <template #footer>
      <div class="picker-actions">
        <span class="picked-count">已选 {{ picked.length }} 道</span>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="adding" @click="confirm">确定添加</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped lang="scss">
.picker-search {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;

  .keyword-input {
    width: 280px;
  }

  .tag-filter {
    width: 300px;
  }
}

.picker-table-wrap {
  min-height: 180px;
  overflow-x: auto;
}

.picker-table {
  min-width: 860px;
  font-size: 14px;

  .pick-row {
    cursor: pointer;
  }

  th,
  td {
    vertical-align: middle;
  }

  .problem-id,
  .problem-name {
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .problem-id {
    font-variant-numeric: tabular-nums;
  }

  .tag-cell {
    white-space: normal;

    .el-tag {
      margin: 2px 5px 2px 0;
    }
  }
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

.picker-pagination {
  display: flex;
  justify-content: center;
  margin-top: 14px;
}

.picker-actions {
  display: flex;
  align-items: center;
  width: 100%;
}

.picked-count {
  margin-right: auto;
  color: var(--el-text-color-regular);
  font-size: 13px;
}

@media (max-width: 680px) {
  .picker-search {
    align-items: stretch;
    flex-direction: column;

    .keyword-input,
    .tag-filter {
      width: 100%;
    }
  }
}
</style>
