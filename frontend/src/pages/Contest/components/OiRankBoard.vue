<template>
  <div class="py-2">
    <!-- OI 赛中封榜 -->
    <div v-if="contest?.frozen"
      class="flex items-center gap-2 rounded border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
      <el-icon>
        <Lock />
      </el-icon>
      OI 赛制赛中封榜，成绩将在比赛结束后公布。
    </div>

    <el-scrollbar v-else>
      <table :style="{ width: `${400 + problems.length * 80}px` }" class="
        font-sans
        table-fixed border-collapse border rounded-2xl
        bg-white text-[13px] text-gray-800 tabular-nums
        [&_th]:px-2 [&_th]:py-2 [&_th]:text-center
        [&_td]:px-2 [&_td]:py-2 [&_td]:text-center
        [&_th]:border [&_th]:border-gray-200
        [&_td]:border [&_td]:border-gray-200">
        <colgroup>
          <col style="width: 70px;" />
          <col style="width: 240px;" />
          <col style="width: 90px;" />
          <col v-for="problem in problems" :key="problem.label" style="width: 80px;">
          <col v-if="contest.status === 2" style="width: 90px;">
        </colgroup>

        <thead>
          <tr class="h-11.5 border-b border-gray-200 bg-[#f7f8fa] tracking-wide text-[13px] text-gray-600">
            <th>排名</th>
            <th>参赛者</th>
            <th>总分</th>
            <th v-for="problem in problems" :key="problem.label">
              <span class="font-semibold">{{ problem.label }}</span>
              <!-- CF：题号下显示该题分值（只显示数字） -->
              <div v-if="isCF" class="text-[11px] font-normal text-gray-400">{{ problem.maxScore }}</div>
            </th>
            <th v-if="contest.status === 2">Final</th>
          </tr>
        </thead>

        <tbody class="font-[Arial,'Noto_Sans_SC',sans-serif] font-medium text-gray-800 text-sm">
          <tr v-for="item in rankList" :key="item.user?.id" class="even:bg-[#F8F8F8]">
            <td>{{ item.rank }}</td>
            <td>
              <div class="flex items-center justify-between">
                <img class="w-10 h-10 rounded-full" :src="item.user.avatar || ''" alt="">
                <span class="flex items-center gap-1">
                  <UserName :name="item.user?.nickname" :rating="item.user?.rating" />
                </span>
              </div>
            </td>
            <td class="total">{{ item.totalScore }}</td>

            <td class="td-problem font-[Arial,'Noto_Sans_SC',sans-serif]" :class="cellClass(problem)"
              v-for="problem in item.problems" :key="problem.label">
              <div class="tracking-wider">{{ problem?.score === 0 ? '' : problem?.score }}</div>
              <!-- CF：分数下面显示 AC 时间 -->
              <div v-if="isCF && problem?.acTime" class="text-[10px] font-medium text-gray-500">
                {{ problem.acTime }}
              </div>
            </td>
            <td v-if="contest.status === 2">
              <span v-if="item.ratingDelta != null" class="text-base font-medium tabular-nums"
                :class="item.ratingDelta >= 0 ? 'text-emerald-600' : 'text-red-500'">{{ item.ratingDelta >= 0 ? '+' +
                  item.ratingDelta : item.ratingDelta }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </el-scrollbar>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Lock } from '@element-plus/icons-vue'
import UserName from '@/components/UserName.vue'

const props = defineProps({
  contest: { type: Object, default: () => ({}) },
  problems: { type: Array, default: () => [] },
  rankList: { type: Array, default: () => [] },
})

// CF 赛制：题头显示题目分值、格子分数下显示 AC 时间
const isCF = computed(() => props.contest?.rule === 'CF')

const maxOf = (label) => props.problems.find((p) => p.label === label)?.maxScore || 0

// CF：AC 即绿（分数随时间衰减，不看是否满分；CF 未 AC 时得分恒为 0）
// OI/IOI：满分绿 / 部分黄 / 0 分不染色
const cellClass = (problem) => {
  const score = problem?.score ?? 0
  if (isCF.value) {
    return score > 0 ? 'cf-ac' : '' // CF：白底 + 绿字，不染背景
  }
  const max = maxOf(problem?.label)
  if (max > 0 && score >= max) return 'full'
  if (score > 0) return 'partial'
  return ''
}
</script>

<style scoped lang="scss">
tbody tr.me td:not(.td-problem) {
  background-color: #fff8e6;
}

.total {
  font-weight: 700;
  color: #1f6feb;
}

.full {
  background-color: #a8e6a1;
}

/* CF：白底 + 绿字（AC 格）*/
.cf-ac {
  color: #00AA00;
  font-size: 15px;
  font-weight: 600;
}

.partial {
  background-color: #fdf0c2;
}
</style>
