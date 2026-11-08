<template>
  <div class="submission-page page-container">
    <div class="f-panel">
      <!-- header -->
      <div class="flex items-center justify-between px-10 pl-6 py-5 border-b border-gray-100 max-lg:flex-col max-lg:items-stretch max-lg:gap-4 max-md:px-3 max-md:py-4">
        <h1 class="text-xl text-gray-600">评测结果</h1>
        <ul class="flex gap-8 max-lg:flex-wrap max-lg:gap-3 max-md:grid max-md:grid-cols-2 max-md:gap-2">
          <li class="max-md:col-span-2">
            <span class="text-sm text-gray-600">状态：</span>
            <el-select v-model="filters.status" placeholder="全部" clearable class="w-56! max-md:w-full!"
              popper-class="status-filter-popper">
              <template #label="{ label, value }">
                <span class="status-option" :class="statusClass(value)">
                  <span class="status-text">{{ label }}</span>
                </span>
              </template>
              <el-option v-for="item in statusList" :key="item.id" :label="item.name" :value="item.id">
                <span class="status-option" :class="statusClass(item.id)">
                  <span class="status-text">{{ item.name }}</span>
                  <el-icon v-if="filters.status === item.id" class="status-check" :size="14">
                    <Check />
                  </el-icon>
                </span>
              </el-option>
            </el-select>
          </li>
          <li class="max-md:col-span-2">
            <el-input v-model="filters.username" placeholder="提交者" clearable class="w-30! max-md:w-full!"
              @keyup.enter="handleSearch">
              <template #prefix>
                <el-icon>
                  <User />
                </el-icon>
              </template>
            </el-input>
          </li>
          <li class="max-md:col-span-2">
            <button type="button" class="mine-toggle" :class="{ 'mine-toggle--active': onlyMine }"
              :disabled="!userStore.isLogin" :aria-pressed="onlyMine" @click="onlyMine = !onlyMine">
              <el-icon :size="15"><UserFilled /></el-icon>
              我的提交
            </button>
          </li>
          <li>
            <el-button type="info" plain class="max-md:ml-0! max-md:w-full!" @click="handleReset">
              <el-icon>
                <Refresh />
              </el-icon>
              重置
            </el-button>
          </li>
          <li>
            <el-button type="primary" class="max-md:ml-0! max-md:w-full!" @click="handleSearch">
              <el-icon>
                <Search />
              </el-icon>
              筛选
            </el-button>
          </li>
        </ul>
      </div>


      <el-table class="submission-table" :data="tableData" stripe style="width: 100%">
        <el-table-column label="#" width="104" align="center">
          <template #default="{ row }">
            <span @click="handleClickId(row.id)"
              class="cursor-pointer text-[15px] font-normal text-blue-500 hover:text-blue-400!">
              {{ row.id }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="problemName" label="题目" min-width="180">
          <template #default="{ row }">
            <span @click="handleClickProblemId(row.problemID)" class="text-blue-400! cursor-pointer">{{ row.problemName
            }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="userName" label="用户名" min-width="130">
          <template #default="{ row }">
            <UserName :name="row.userName" :rating="row.rating" clickable @click="handleClickUser(row.userID)" />
          </template>
        </el-table-column>
        <el-table-column prop="result" label="结果" min-width="130">
          <template #default="{ row }">
            <FTag :result="row.result" />
          </template>
        </el-table-column>
        <el-table-column label="时间" width="100">
          <template #default="{ row }">
            {{ formatTime(row.timeUsed) }}
          </template>
        </el-table-column>
        <el-table-column label="内存" width="100">
          <template #default="{ row }">
            {{ formatMemory(row.memoryUsed) }}
          </template>
        </el-table-column>
        <el-table-column prop="language" label="语言" width="100" />
        <el-table-column prop="createdAt" label="提交时间" width="180" />
        <el-table-column v-if="userStore.can('submission.rejudge')" label="操作" width="140">
          <template #default="{ row }">
            <el-button type="primary" @click="handleRejudge(row)">
              <el-icon>
                <Refresh />
              </el-icon>
              重新评测
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="f-pagination">
        <AppPagination layout="total, prev, pager, next" :current-page="page" :page-size="pageSize" :total="total"
          @current-change="handlePageChange" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch, computed, onBeforeUnmount } from 'vue'
import { getSubmit, rejudgeSubmission, getSubmitDetail } from '@/api/user'
import {
  PendingCode, AcceptedCode, MemoryLimitExceededCode, TimeLimitExceededCode,
  RuntimeErrorCode, WrongAnswerCode, CompileErrorCode, UnknownErrorCode,
  PresentationErrorCode, JUDGE_STATUS_TEXT,
  JUDGE_STATUS_CLASS
} from '@/constants/index'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { formatTime, formatMemory } from '@/utils/format'
import FTag from "@/components/FTag.vue";
import UserName from "@/components/UserName.vue";
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, User, UserFilled, Check } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()


const tableData = ref([])

const total = ref(0)

// 筛选表单草稿：一组相关字段用 reactive 打包（读写不用 .value，整体好传）。
// 初始值从 URL 回填，保证刷新/分享后筛选条件不丢。
const filters = reactive({
  status: route.query.status != null ? Number(route.query.status) : undefined,
  username: route.query.username || '',
})

const statusList = [
  AcceptedCode,
  PresentationErrorCode,
  WrongAnswerCode,
  TimeLimitExceededCode,
  MemoryLimitExceededCode,
  RuntimeErrorCode,
  CompileErrorCode,
  UnknownErrorCode,
].map(id => ({ id, name: JUDGE_STATUS_TEXT[id] }));

const statusClass = (id) => JUDGE_STATUS_CLASS[id] || 'unknown'

// 从路由中读取参数
const page = computed(() => Number(route.query.page || 1))
const pageSize = computed(() => Number(route.query.pageSize || 15))

// 我的提交：开启时清空提交者输入，避免两个条件互相干扰
const onlyMine = computed({
  get: () => route.query.userID != null && String(route.query.userID) === String(userStore.user?.id),
  set: (val) => {
    const query = { ...route.query, page: 1, pageSize: pageSize.value }
    if (val && userStore.user?.id != null) {
      query.userID = userStore.user.id
      delete query.username
      filters.username = ''
    } else {
      delete query.userID
    }
    router.push({ query })
  },
})

const handleClickId = (id) => {
  router.push({
    path: `/submission/${id}`,
  })
}

const handleClickProblemId = (problemID) => {
  router.push({
    path: `/problem/${problemID}`,
  })
}

const handleClickUser = (id) => {
  router.push({ name: 'User', params: { id } })
}

// 获取数据（生效的筛选条件以 URL 为准；undefined 的参数 axios 会自动忽略）
const getList = async () => {
  const res = await getSubmit({
    page: page.value,
    pageSize: pageSize.value,
    status: route.query.status != null ? Number(route.query.status) : undefined,
    username: route.query.username || undefined,
    userID: route.query.userID != null ? Number(route.query.userID) : undefined,
  })
  tableData.value = res.data.list
  total.value = res.data.total
}

// 点“筛选”：把草稿写进 URL 并回到第 1 页
const handleSearch = () => {
  const query = { page: 1, pageSize: pageSize.value }
  if (filters.status != null) query.status = filters.status
  if (filters.username) query.username = filters.username
  else if (route.query.userID != null) query.userID = route.query.userID
  router.push({ query })
}

// 点“重置”：清空草稿与 URL 筛选
const handleReset = () => {
  filters.status = undefined
  filters.username = ''
  router.push({ query: { page: 1, pageSize: pageSize.value } })
}

// 页码改变
const handlePageChange = (val) => {
  router.push({
    query: {
      ...route.query,
      page: val,
      pageSize: pageSize.value,
    }
  })
}

// ---------- 重判（超管）----------
const pollTimers = new Set()

// 重判后轮询该行状态，直到离开 Pending（最多 ~60s）
const pollRow = (item) => {
  let tries = 0
  const tick = async () => {
    tries++
    let done = true
    try {
      const res = await getSubmitDetail(item.id)
      const sub = res.data?.submission
      if (sub) {
        item.result = sub.status
        item.timeUsed = sub.time
        item.memoryUsed = sub.memory
        done = sub.status !== PendingCode || tries >= 40
      }
    } catch {
      done = true
    }
    if (!done) {
      const t = setTimeout(tick, 1500)
      pollTimers.add(t)
    }
  }
  const t = setTimeout(tick, 1500)
  pollTimers.add(t)
}

const handleRejudge = async (item) => {
  try {
    await ElMessageBox.confirm(`确认重判提交 #${item.id}？`, '重判', { type: 'warning' })
  } catch {
    return
  }
  item.rejudging = true
  try {
    await rejudgeSubmission(item.id)
    item.result = PendingCode
    item.timeUsed = 0
    item.memoryUsed = 0
    ElMessage.success('已加入重判队列')
    pollRow(item)
  } catch (err) {
    console.error(err)
  } finally {
    item.rejudging = false
  }
}

onBeforeUnmount(() => {
  pollTimers.forEach(clearTimeout)
  pollTimers.clear()
})

// 监听路由变化自动重新请求
watch(
  () => [route.query.page, route.query.pageSize, route.query.status, route.query.username],
  getList,
  { immediate: true }
)
</script>

<style scoped>
.submission-page.page-container {
  max-width: 1350px;
}

.submission-table :deep(.el-table__cell) {
  text-align: center;
}

.mine-toggle {
  height: 32px;
  padding: 0 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid #d8e0ea;
  border-radius: 4px;
  color: #5b6675;
  background: #fff;
  font-size: 14px;
  cursor: pointer;
  transition: border-color 150ms ease, color 150ms ease, background-color 150ms ease;
}

.mine-toggle:hover:not(:disabled) {
  border-color: #aebdce;
  color: #344256;
}

.mine-toggle--active {
  border-color: #409eff;
  color: #409eff;
  background: #ecf5ff;
  font-weight: 500;
}

.mine-toggle--active:hover:not(:disabled) {
  border-color: #409eff;
  color: #409eff;
}

.mine-toggle:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.status-option {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.status-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  font-size: 14px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-check {
  flex-shrink: 0;
  margin-left: auto;
  color: inherit;
}

:global(.el-select__popper.status-filter-popper) {
  padding: 4px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  box-shadow: 0 10px 30px rgb(15 23 42 / 12%);
}

:global(.status-filter-popper .el-select-dropdown__list) {
  padding: 0;
}

:global(.status-filter-popper .el-select-dropdown__item) {
  height: 34px;
  margin: 2px 0;
  padding: 0 10px;
  border-radius: 6px;
  line-height: 34px;
}

:global(.status-filter-popper .el-select-dropdown__item.is-hovering) {
  background: #f5f7fa;
}

:global(.status-filter-popper .el-select-dropdown__item.is-selected) {
  background: #eff6ff;
}

:global(.status-filter-popper .el-select-dropdown__item.is-selected .status-text) {
  font-weight: 600;
}
</style>
