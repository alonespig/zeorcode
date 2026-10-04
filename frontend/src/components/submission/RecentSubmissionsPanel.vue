<script setup>
import { computed } from "vue";
import FTag from "@/components/FTag.vue";

const props = defineProps({
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

const rows = computed(() => props.submissions.map((item) => {
  const timestamp = item.createdAt || "";
  const parts = timestamp.match(/^(\d{4}-\d{2}-\d{2})[ T](\d{2}:\d{2})/);
  return { ...item, date: parts ? parts[1] : timestamp || "—", time: parts ? parts[2] : "" };
}));
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
    <table v-else-if="rows.length" class="recent-submissions" aria-label="最近提交记录">
      <colgroup>
        <col class="recent-submissions__id-column" />
        <col class="recent-submissions__time-column" />
        <col class="recent-submissions__result-column" />
      </colgroup>
      <thead>
        <tr><th scope="col">提交</th><th scope="col">时间</th><th scope="col">结果</th></tr>
      </thead>
      <tbody>
        <tr v-for="item in rows" :key="item.id">
          <td>
            <button type="button" class="recent-submissions__id" :aria-label="`查看提交 ${item.id}`"
              @click="emit('click-id', item.id)">{{ item.id }}</button>
          </td>
          <td>
            <span class="recent-submissions__timestamp" :title="item.createdAt">
              <span>{{ item.date }}</span>
              <span v-if="item.time">{{ item.time }}</span>
            </span>
          </td>
          <td><FTag :result="item.result" /></td>
        </tr>
      </tbody>
    </table>
    <div v-else class="px-4 py-6 text-center text-sm text-gray-400">暂无提交</div>
  </section>
</template>

<style scoped>
.recent-submissions {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  color: var(--table-text);
  font-size: 13px;
  text-align: center;
}

.recent-submissions__id-column { width: 33%; }
.recent-submissions__time-column { width: 34%; }
.recent-submissions__result-column { width: 33%; }

.recent-submissions th,
.recent-submissions td {
  padding: 10px 6px;
  border-bottom: 1px solid var(--table-border);
  vertical-align: middle;
}

.recent-submissions th + th,
.recent-submissions td + td { border-left: 1px solid var(--table-border); }
.recent-submissions th { padding-block: 8px; background: var(--table-header-bg); color: var(--table-header-text); font-size: 12px; font-weight: 500; }
.recent-submissions tbody tr:last-child td { border-bottom: 0; }
.recent-submissions tbody tr:hover { background: var(--table-row-hover); }

.recent-submissions__id {
  max-width: 100%;
  padding: 4px 0;
  color: var(--color-primary);
  font: inherit;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.recent-submissions__id:hover { text-decoration: underline; }
.recent-submissions__id:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.recent-submissions__timestamp { display: flex; flex-direction: column; font-size: 12px; font-variant-numeric: tabular-nums; line-height: 1.5; overflow-wrap: anywhere; }
.recent-submissions :deep(.ftag) { white-space: normal; overflow-wrap: anywhere; line-height: 1.5; }
</style>
