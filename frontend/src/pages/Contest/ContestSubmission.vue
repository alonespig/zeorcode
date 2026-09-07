<template>
  <div class="f-panel overflow-hidden">
    <!-- 管理员：全场筛选栏 -->
    <div v-if="isAdmin" class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-4 py-3">
      <el-input v-model="filters.username" placeholder="用户名" clearable style="width: 150px"
        @keyup.enter="applyFilter">
        <template #prefix>
          <el-icon>
            <Search />
          </el-icon>
        </template>
      </el-input>
      <el-select v-model="filters.problemID" placeholder="题目" clearable style="width: 200px">
        <el-option v-for="p in problems" :key="p.problemID" :label="`${p.label}. ${p.name}`" :value="p.problemID" />
      </el-select>
      <el-select v-model="filters.status" placeholder="结果" clearable style="width: 150px">
        <el-option v-for="s in statusList" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-button type="primary" :icon="Search" @click="applyFilter">筛选</el-button>
      <el-button @click="resetFilter">重置</el-button>
      <div class="ml-auto flex items-center gap-2">
        <span class="text-xs text-gray-500">自动刷新</span>
        <el-switch v-model="autoRefresh" size="small" />
        <el-button :icon="Refresh" circle size="small" title="刷新" @click="fetchData" />
        <span class="ml-1 text-xs text-gray-400">共 {{ tableData.length }} 份</span>
      </div>
    </div>

    <SubmissionStatus v-if="tableData.length" :tableData="tableData" :hide-user="!isAdmin" :show-rejudge="isAdmin"
      @click-id="goDetail" @click-problem-id="goProblem" @click-user="goUser" @rejudge="handleRejudge" />
    <el-empty v-else description="还没有提交记录" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import SubmissionStatus from '@/components/SubmissionStatus.vue'
import { getSubmission, getContestProblemList } from '@/api/contest'
import { rejudgeSubmission, getSubmitDetail } from '@/api/user'
import { useUserStore } from '@/stores/user'
import {
  PendingCode, AcceptedCode, MemoryLimitExceededCode, TimeLimitExceededCode,
  RuntimeErrorCode, WrongAnswerCode, CompileErrorCode, UnknownErrorCode, JUDGE_STATUS_TEXT,
} from '@/constants/index'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const tableData = ref([])
const problems = ref([])
const filters = reactive({ username: '', problemID: undefined, status: undefined })

const statusList = [
  AcceptedCode, MemoryLimitExceededCode, TimeLimitExceededCode,
  RuntimeErrorCode, WrongAnswerCode, CompileErrorCode, UnknownErrorCode,
].map((id) => ({ id, name: JUDGE_STATUS_TEXT[id] }))

const problemLabelById = computed(() =>
  new Map(problems.value.map((problem) => [String(problem.problemID), problem.label]))
)

const goDetail = (id) => router.push({ name: 'SubmissionDetail', params: { id } })
const goUser = (id) => router.push({ name: 'User', params: { id } })
const goProblem = (pid) => {
  const label = problemLabelById.value.get(String(pid))
  if (!label) {
    ElMessage.warning('题目信息加载中，请稍后重试')
    return
  }
  router.push({ name: 'ContestProblemDetail', params: { id: route.params.id, problemId: label } })
}

const fetchData = async () => {
  const params = isAdmin.value
    ? {
      username: filters.username || undefined,
      problemID: filters.problemID || undefined,
      status: filters.status ?? undefined,
    }
    : undefined
  const res = await getSubmission(route.params.id, params)
  tableData.value = res.data.list || []
}

const applyFilter = () => fetchData()
const resetFilter = () => {
  filters.username = ''
  filters.problemID = undefined
  filters.status = undefined
  fetchData()
}

// ---------- 自动刷新（整表，管理员盯直播/重判进度用）----------
const autoRefresh = ref(false)
let refreshTimer = null
watch(autoRefresh, (on) => {
  if (on) refreshTimer = setInterval(fetchData, 5000)
  else if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null }
})

// ---------- 单条重判 + 轮询该行 ----------
const pollTimers = new Set()
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

const fetchProblems = async () => {
  const res = await getContestProblemList(route.params.id)
  problems.value = res.data.problemList || []
}

onMounted(() => {
  fetchData()
  fetchProblems()
})
onBeforeUnmount(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  pollTimers.forEach(clearTimeout)
  pollTimers.clear()
})
</script>
