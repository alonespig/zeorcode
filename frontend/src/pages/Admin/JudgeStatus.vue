<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="t">评测机状态</span>
          <span class="sub">共 {{ list.length }} 台 · 在线 {{ onlineCount }}</span>
        </div>
        <div class="flex items-center gap-2">
          <el-switch v-model="auto" active-text="自动刷新" inline-prompt />
          <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
        </div>
      </div>
    </template>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column label="地址" min-width="280">
        <template #default="{ row }"><span class="mono">{{ row.url }}</span></template>
      </el-table-column>
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="row.online ? 'success' : 'danger'" size="small" effect="light">
            <span class="dot" :class="row.online ? 'on' : 'off'"></span>{{ row.online ? '在线' : '离线' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="延迟" width="100">
        <template #default="{ row }">
          <span v-if="row.online" :class="latClass(row.latencyMs)">{{ row.latencyMs }} ms</span>
          <span v-else class="text-empty">—</span>
        </template>
      </el-table-column>
      <el-table-column label="版本" width="140" show-overflow-tooltip>
        <template #default="{ row }">
          <span :class="{ 'text-empty': !row.version }">{{ row.version || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="错误" width="300">
        <template #default="{ row }">
          <span v-if="row.error" class="err" :title="row.error">{{ row.error }}</span>
          <span v-else class="text-empty">—</span>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="未配置评测机地址（config.yaml 的 judge.urls）" :image-size="80" />
      </template>
    </el-table>

    <p class="hint">状态为 HTTP 服务进程现场探测的结果；判题进程实际路由时也会自动跳过挂掉的实例。</p>
  </el-card>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { Refresh } from "@element-plus/icons-vue";
import { getJudgeStatus } from "@/api/admin";

const list = ref([]);
const loading = ref(false);
const auto = ref(false);
let timer = null;

const onlineCount = computed(() => list.value.filter((x) => x.online).length);

const latClass = (ms) => (ms < 100 ? "lat-good" : ms < 500 ? "lat-mid" : "lat-bad");

const load = async () => {
  loading.value = true;
  try {
    const res = await getJudgeStatus();
    list.value = res.data?.list || [];
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

// 自动刷新：每 10s 探一次
const startAuto = () => {
  stopAuto();
  timer = setInterval(load, 10000);
};
const stopAuto = () => {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
};

watch(auto, (v) => (v ? startAuto() : stopAuto()));

onMounted(load);
onBeforeUnmount(stopAuto);
</script>

<style scoped>
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.head-title .t {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.head-title .sub {
  margin-left: 10px;
  font-size: 13px;
  color: #909399;
}

.mono {
  font-family: "SFMono-Regular", Consolas, monospace;
  font-size: 13px;
}

.dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 5px;
  vertical-align: middle;
}

.dot.on {
  background: #67c23a;
}

.dot.off {
  background: #f56c6c;
}

.lat-good {
  color: #67c23a;
}

.lat-mid {
  color: #e6a23c;
}

.lat-bad {
  color: #f56c6c;
}

.err {
  color: #f56c6c;
  font-size: 12px;
  display: inline-block;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-empty {
  color: #c0c4cc;
}

.hint {
  margin-top: 14px;
  font-size: 12px;
  color: #9aa1ab;
}
</style>
