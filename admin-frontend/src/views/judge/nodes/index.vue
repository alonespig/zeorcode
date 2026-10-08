<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { Refresh } from "@element-plus/icons-vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { getJudgeStatus, type JudgeNodeStatus } from "@/api/admin/judge";

defineOptions({ name: "AdminJudgeNodes" });

const dataList = ref<JudgeNodeStatus[]>([]);
const loading = ref(false);
const autoRefresh = ref(false);
let refreshTimer: ReturnType<typeof setInterval> | undefined;

const onlineCount = computed(
  () => dataList.value.filter(item => item.online).length
);

const columns: TableColumnList = [
  { label: "地址", prop: "url", minWidth: 280 },
  { label: "状态", slot: "status", width: 110, align: "center" },
  { label: "延迟", slot: "latency", width: 110, align: "center" },
  {
    label: "版本",
    prop: "version",
    width: 160,
    align: "center",
    showOverflowTooltip: true
  },
  {
    label: "错误信息",
    slot: "error",
    minWidth: 260,
    showOverflowTooltip: true
  }
];

function latencyType(ms: number) {
  if (ms < 100) return "text-green-600";
  if (ms < 500) return "text-orange-500";
  return "text-red-500";
}

async function loadStatus() {
  if (loading.value) return;
  loading.value = true;
  try {
    const res = await getJudgeStatus();
    dataList.value = res.data?.list ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function stopAutoRefresh() {
  if (refreshTimer !== undefined) {
    clearInterval(refreshTimer);
    refreshTimer = undefined;
  }
}

watch(autoRefresh, enabled => {
  stopAutoRefresh();
  if (enabled) refreshTimer = setInterval(loadStatus, 10_000);
});

onMounted(loadStatus);
onBeforeUnmount(stopAutoRefresh);
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="loadStatus">
      <template #title>
        <div class="flex items-center">
          <span>评测机状态</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ dataList.length }} 台，在线 {{ onlineCount }} 台
          </span>
        </div>
      </template>
      <template #buttons>
        <div class="flex items-center gap-3">
          <el-switch
            v-model="autoRefresh"
            inline-prompt
            active-text="自动"
            inactive-text="手动"
          />
          <el-button :icon="Refresh" :loading="loading" @click="loadStatus">
            刷新
          </el-button>
        </div>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <PureTable
          border
          adaptive
          table-layout="auto"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          empty-text="未配置评测机地址"
        >
          <template #status="{ row }">
            <el-tag :type="row.online ? 'success' : 'danger'" size="small">
              {{ row.online ? "在线" : "离线" }}
            </el-tag>
          </template>
          <template #latency="{ row }">
            <span v-if="row.online" :class="latencyType(row.latencyMs)">
              {{ row.latencyMs }} ms
            </span>
            <span v-else class="text-gray-400">—</span>
          </template>
          <template #error="{ row }">
            <span v-if="row.error" class="text-red-500">{{ row.error }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>
        </PureTable>
        <p class="mt-3 text-xs text-gray-400">
          页面展示后端对各评测实例的实时探活结果；开启自动刷新后每 10
          秒更新一次。
        </p>
      </template>
    </PureTableBar>
  </div>
</template>
