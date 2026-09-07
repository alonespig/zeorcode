<template>
  <div class="py-2">
    <el-scrollbar>
      <table :style="{ width: `${520 + problems.length * 80}px` }" class="font-[Arial,'Noto_Sans_SC',sans-serif] table-fixed border-collapse border
        bg-white text-[13px] text-gray-800 tabular-nums
        [&_th]:px-2 [&_th]:py-2 [&_th]:text-center
        [&_td]:px-2 [&_td]:py-2 [&_td]:text-center
        [&_th]:border [&_th]:border-gray-200
        [&_td]:border [&_td]:border-gray-200">
        <colgroup>
          <col style="width: 70px;" />
          <col style="width: 240px;" />
          <col style="width: 90px;" />
          <col style="width: 120px;" />
          <col v-for="problem in problems" :key="problem.label" style="width: 80px;">
        </colgroup>

        <thead>
          <tr class="h-11.5 border-b border-gray-200 bg-[#f7f8fa] tracking-wide text-gray-700">
            <th>排名</th>
            <th>参赛者</th>
            <th>AC</th>
            <th>罚时</th>
            <th v-for="problem in problems" :key="problem.label">
              <span class="font-semibold">{{ problem.label }}</span>
            </th>
          </tr>
        </thead>

        <tbody class="font-[Arial,'Noto_Sans_SC',sans-serif] font-medium text-gray-800 text-sm">
          <tr v-for="item in rankList" :key="item.user?.id" :class="{ me: item.isSelf }">
            <td>{{ item.rank }}</td>
            <td>
              <div class="flex items-center justify-between">
                <img class="w-10 h-10 rounded-full" :src="item.user.avatar || ''" alt="">
                <span class="flex items-center gap-1">
                  <UserName :name="item.user?.nickname" :rating="item.user?.rating" />
                  <span v-if="item.ratingDelta != null" class="text-xs tabular-nums"
                    :class="item.ratingDelta >= 0 ? 'text-emerald-600' : 'text-red-500'">{{ item.ratingDelta >= 0 ? '+' +
                      item.ratingDelta : item.ratingDelta }}</span>
                </span>
              </div>
            </td>
            <td class="tabular-nums">{{ item.acCount }}</td>
            <td>{{ item.penalty }}</td>
            <td class="td-problem" :class="getStatus(problem?.status, problem?.first)"
              v-for="problem in item.problems" :key="problem.label">
              <div>
                <span v-if="problem?.status !== PendingCode">
                  {{ problem?.status === AcceptedCode ? '+' : '-' }}
                </span>
                <span>{{ problem?.tries === 0 ? '' : problem?.tries }}</span>
              </div>
              <div v-if="problem?.status === AcceptedCode">
                <span>{{ problem?.acTime }}</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </el-scrollbar>
  </div>
</template>

<script setup>
import { PendingCode, AcceptedCode, WrongAnswerCode } from '@/constants/index'
import UserName from '@/components/UserName.vue'

defineProps({
  contest: { type: Object, default: () => ({}) },
  problems: { type: Array, default: () => [] },
  rankList: { type: Array, default: () => [] },
})

const getStatus = (status, first) => {
  if (first === 1) return 'first'
  if (status === AcceptedCode) return 'ac'
  if (status === WrongAnswerCode) return 'wa'
  return ''
}
</script>

<style scoped lang="scss">
tbody tr.me td:not(.td-problem) {
  background-color: #fff8e6;
}

.ac {
  background-color: #69ec5f;
}

.first {
  background-color: #3DB03D;
}

.wa {
  background-color: #e73434;
}
</style>
