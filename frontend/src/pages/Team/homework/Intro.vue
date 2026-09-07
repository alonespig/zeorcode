<template>
  <div class="grid grid-cols-1 gap-4 lg:grid-cols-[1fr_300px]">
    <!-- 左：简介内容（tab 已叫「简介」，不再重复标题） -->
    <div class="f-panel">
      <div class="px-5 py-4">
        <MdEditor v-if="hw.description" :model-value="hw.description" preview-only editor-id="homework-desc" />
        <span v-else class="text-sm text-gray-400">暂无简介</span>
      </div>
    </div>

    <!-- 右：信息栏 + 我的进度 -->
    <div class="space-y-4">
      <div class="f-panel">
        <div class="divide-y divide-gray-100 text-[15px]">
          <div class="flex items-center justify-between px-5 py-2.5">
            <span class="font-medium text-gray-700">布置人</span>
            <span class="flex items-center gap-2 text-gray-900">
              <el-avatar :size="24" :src="hw.creatorAvatar || undefined" />
              {{ hw.creator }}
            </span>
          </div>
          <div class="flex justify-between px-5 py-2.5">
            <span class="font-medium text-gray-700">题数</span>
            <span class="tabular-nums text-gray-900">{{ hw.problemCount }}</span>
          </div>
          <div class="flex justify-between px-5 py-2.5">
            <span class="font-medium text-gray-700">开始时间</span>
            <span class="tabular-nums text-gray-900">{{ fmtTime(hw.startTime) }}</span>
          </div>
          <div class="flex justify-between px-5 py-2.5">
            <span class="font-medium text-gray-700">结束时间</span>
            <span class="tabular-nums text-gray-900">{{ fmtTime(hw.endTime) }}</span>
          </div>
        </div>
      </div>

      <!-- 我的完成度 + 我的得分（进度条，同设计稿） -->
      <div v-if="hw.myScore != null" class="f-panel px-5 py-4">
        <CapsuleProgress
          label="我的完成度"
          :value="hw.solvedCount ?? 0"
          :max="hw.problemCount"
          aria-label="我的作业完成度"
        />
        <CapsuleProgress
          class="mt-4"
          label="我的得分"
          :value="hw.myScore"
          :max="hw.totalScore"
          aria-label="我的作业得分"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { inject } from "vue";
import CapsuleProgress from "@/components/homework/CapsuleProgress.vue";

const hw = inject("homework");

// 同年省略年份，与作业列表一致
const fmtTime = (s) => {
  if (!s) return "-";
  return s.slice(0, 4) === String(new Date().getFullYear()) ? s.slice(5) : s;
};

</script>
