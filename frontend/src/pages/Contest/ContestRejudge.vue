<template>
  <div class="f-panel overflow-hidden">
    <!-- 头 -->
    <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4">
      <div class="text-[15px] font-semibold text-gray-800">
        重判 <span class="ml-2 text-xs font-normal text-gray-400">共 {{ problems.length }} 题</span>
      </div>
      <el-button type="warning" :icon="RefreshRight" :loading="wholeLoading" @click="rejudgeWhole">重判整场</el-button>
    </div>

    <!-- 题目表 -->
    <table class="w-full border-collapse text-[13px]">
      <thead>
        <tr class="[&>th]:bg-[#f7f8fa] [&>th]:border-b [&>th]:border-gray-100 [&>th]:px-5 [&>th]:py-2.5
          [&>th]:text-[12px] [&>th]:font-semibold [&>th]:text-gray-500">
          <th class="text-left" style="width: 90px">题号</th>
          <th class="text-left">题目</th>
          <th class="text-center" style="width: 140px">通过 / 提交</th>
          <th class="text-right" style="width: 150px">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in problems" :key="p.problemID"
          class="[&>td]:border-b [&>td]:border-gray-50 [&>td]:px-5 [&>td]:py-3 hover:bg-[#fafbfc]">
          <td>
            <span class="inline-flex items-center gap-2">
              <span class="inline-block h-2.5 w-2.5 rounded-full" :style="{ background: p.color || '#909399' }"></span>
              <span class="font-bold text-gray-700">{{ p.label }}</span>
            </span>
          </td>
          <td class="font-medium text-[#2f6ad0]">{{ p.name }}</td>
          <td class="text-center tabular-nums text-gray-600">
            <span class="font-semibold text-emerald-600">{{ p.acceptedCount }}</span>
            <span class="mx-1 text-gray-300">/</span>{{ p.totalCount }}
          </td>
          <td class="text-right">
            <el-button type="primary" size="small" :icon="RefreshRight" :loading="p._loading"
              @click="rejudgeProblem(p)">重判本题</el-button>
          </td>
        </tr>
        <tr v-if="!problems.length">
          <td colspan="4" class="px-5 py-8 text-center text-sm text-gray-400">本场暂无题目</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { RefreshRight } from '@element-plus/icons-vue'
import { getContestProblemList, rejudgeContest, rejudgeContestProblem } from '@/api/contest'

const route = useRoute()
const problems = ref([])
const wholeLoading = ref(false)

const fetchProblems = async () => {
  const res = await getContestProblemList(route.params.id)
  problems.value = (res.data.problemList || []).map((p) => ({ ...p, _loading: false }))
}

const rejudgeProblem = async (p) => {
  try {
    await ElMessageBox.confirm(`确认重判「${p.label}. ${p.name}」的全部提交？`, '重判本题', { type: 'warning' })
  } catch {
    return
  }
  p._loading = true
  try {
    const res = await rejudgeContestProblem(route.params.id, p.problemID)
    ElMessage.success(`已加入重判队列 · ${res.data?.rejudged ?? 0} 份`)
  } catch (err) {
    console.error(err)
  } finally {
    p._loading = false
  }
}

const rejudgeWhole = async () => {
  try {
    await ElMessageBox.confirm('将把本场全部提交重新评测，判题期间榜单会随进度自动更新。确认继续？', '重判整场', {
      type: 'warning',
      confirmButtonText: '确认重判',
    })
  } catch {
    return
  }
  wholeLoading.value = true
  try {
    const res = await rejudgeContest(route.params.id)
    ElMessage.success(`已加入重判队列 · ${res.data?.rejudged ?? 0} 份`)
  } catch (err) {
    console.error(err)
  } finally {
    wholeLoading.value = false
  }
}

onMounted(fetchProblems)
</script>
