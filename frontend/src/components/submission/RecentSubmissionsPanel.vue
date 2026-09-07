<script setup>
import FTag from "@/components/FTag.vue";

defineProps({
  submissions: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["click-id", "view-all"]);
</script>

<template>
  <section class="f-panel overflow-hidden" aria-label="本题最近提交">
    <div class="flex items-center justify-between border-b border-gray-100 px-4 py-2.5">
      <h2 class="text-sm font-medium text-gray-700">最近提交</h2>
      <button type="button" class="text-xs text-blue-500 transition-colors hover:text-blue-600 hover:underline"
        @click="emit('view-all')">
        查看全部
      </button>
    </div>

    <div v-if="loading" class="px-4 py-6 text-center text-sm text-gray-400">加载中...</div>
    <div v-else-if="submissions.length">
      <div
        class="grid grid-cols-[54px_68px_minmax(0,1fr)] gap-2 border-b border-gray-100 bg-gray-50 px-4 py-1.5 text-xs text-gray-400">
        <span>提交</span>
        <span>时间</span>
        <span class="text-right">结果</span>
      </div>
      <button v-for="item in submissions" :key="item.id" type="button"
        class="grid w-full grid-cols-[54px_68px_minmax(0,1fr)] items-center gap-2 border-b border-gray-100 px-4 py-2.5 text-left transition-colors last:border-b-0 hover:bg-gray-50"
        @click="emit('click-id', item.id)">
        <span class="truncate font-mono text-[13px] font-medium text-blue-500">#{{ item.id }}</span>
        <span class="truncate font-mono text-xs text-gray-400" :title="item.createdAt">{{ item.createdAt }}</span>
        <span class="min-w-0 overflow-hidden text-right">
          <FTag :result="item.result" />
        </span>
      </button>
    </div>
    <div v-else class="px-4 py-6 text-center text-sm text-gray-400">暂无提交</div>
  </section>
</template>
