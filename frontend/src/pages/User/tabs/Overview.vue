<template>
  <div class="px-6 py-5">
    <!-- Rating 走势（始终显示，没记录时是空图） -->
    <section>
      <div class="mb-3 flex items-center justify-between">
        <span class="flex items-center gap-1.5 text-base font-semibold text-blue-500">
          <el-icon><TrendCharts /></el-icon>
          <span>Rating 走势</span>
        </span>
        <span class="text-xs text-gray-400">{{ ratingSummary }}</span>
      </div>
      <div class="rounded-lg border border-gray-100 bg-white px-4 pt-3.5 pb-2">
        <RatingChart :history="ratingHistory" />
      </div>
    </section>

    <!-- 比赛记录 -->
    <section class="mt-7">
      <div class="mb-3 flex items-center justify-between">
        <span class="flex items-center gap-1.5 text-base font-semibold text-blue-500">
          <el-icon><DataAnalysis /></el-icon>
          <span>比赛记录</span>
        </span>
      </div>
      <ContestHistory />
    </section>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import RatingChart from '@/components/user/RatingChart.vue'
import ContestHistory from '@/components/user/ContestHistory.vue'
import { TrendCharts, DataAnalysis } from '@element-plus/icons-vue'

// rating（当前分 + 历史）由 UserLayout 注入
const ratingData = inject('userRating')
const ratingHistory = computed(() => ratingData?.value?.history || [])

const ratingSummary = computed(() => {
  const d = ratingData?.value
  if (!d || !ratingHistory.value.length) return '暂无 rated 记录'
  return `当前 ${d.rating} · 最高 ${d.maxRating} · ${ratingHistory.value.length} 场`
})
</script>
