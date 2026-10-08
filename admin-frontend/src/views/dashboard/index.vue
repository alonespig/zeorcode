<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch
} from "vue";
import { useRouter } from "vue-router";
import { useResizeObserver } from "@vueuse/core";
import echarts from "@/plugins/echarts";
import {
  Bell,
  ChatDotRound,
  Collection,
  DataAnalysis,
  DocumentChecked,
  Loading,
  Postcard,
  Refresh,
  User
} from "@element-plus/icons-vue";
import {
  getDashboardOverview,
  type DashboardOverview
} from "@/api/admin/dashboard";
import { getJudgeStatus, type JudgeNodeStatus } from "@/api/admin/judge";

defineOptions({ name: "AdminDashboard" });

const router = useRouter();
const loading = ref(false);
const overview = ref<DashboardOverview>();
const judgeNodes = ref<JudgeNodeStatus[]>([]);
const judgeLoading = ref(false);
const trendChartRef = ref<HTMLDivElement>();
let trendChart: ReturnType<typeof echarts.init> | undefined;

const summaryCards = computed(() => {
  const summary = overview.value?.summary;
  return [
    {
      label: "用户总数",
      value: summary?.users ?? 0,
      icon: User,
      tone: "blue",
      path: "/users"
    },
    {
      label: "题目总数",
      value: summary?.problems ?? 0,
      icon: Collection,
      tone: "cyan",
      path: "/content/problems"
    },
    {
      label: "今日提交",
      value: summary?.todaySubmissions ?? 0,
      icon: DataAnalysis,
      tone: "green",
      path: "/judge/submissions"
    },
    {
      label: "待审核帖子",
      value: summary?.pendingPosts ?? 0,
      icon: Postcard,
      tone: "orange",
      path: "/content/post-review"
    },
    {
      label: "进行中比赛",
      value: summary?.runningContests ?? 0,
      icon: DocumentChecked,
      tone: "purple",
      path: "/content/contests"
    },
    {
      label: "等待评测",
      value: summary?.pendingJudging ?? 0,
      icon: Loading,
      tone: "red",
      path: "/judge/submissions"
    }
  ];
});

const quickActions = [
  {
    title: "新建题目",
    description: "录入题面与测试数据",
    icon: Collection,
    tone: "blue",
    path: "/content/problems/create"
  },
  {
    title: "创建比赛",
    description: "配置赛制和比赛题目",
    icon: DocumentChecked,
    tone: "purple",
    path: "/content/contests/create"
  },
  {
    title: "发布通知",
    description: "向全部用户发送系统消息",
    icon: Bell,
    tone: "orange",
    path: "/notifications"
  },
  {
    title: "AI 助手",
    description: "通过对话完成教学任务",
    icon: ChatDotRound,
    tone: "green",
    path: "/agent"
  }
];

const onlineJudgeCount = computed(
  () => judgeNodes.value.filter(node => node.online).length
);

function shortDate(value: string) {
  return value.slice(5).replace("-", "/");
}

async function renderTrendChart() {
  await nextTick();
  if (!trendChartRef.value) return;
  trendChart ||= echarts.init(trendChartRef.value);
  const trend = overview.value?.submissionTrend ?? [];
  trendChart.setOption(
    {
      animationDuration: 450,
      grid: { left: 44, right: 24, top: 32, bottom: 36 },
      tooltip: {
        trigger: "axis",
        valueFormatter: (value: number) => `${value} 次提交`
      },
      xAxis: {
        type: "category",
        boundaryGap: false,
        data: trend.map(item => shortDate(item.date)),
        axisTick: { show: false },
        axisLine: { lineStyle: { color: "#dcdfe6" } },
        axisLabel: { color: "#909399", fontSize: 11 }
      },
      yAxis: {
        type: "value",
        minInterval: 1,
        min: 0,
        axisLabel: { color: "#909399", fontSize: 11 },
        splitLine: { lineStyle: { color: "#ebeef5", type: "dashed" } }
      },
      series: [
        {
          name: "提交数",
          type: "line",
          data: trend.map(item => item.count),
          smooth: 0.2,
          symbol: "emptyCircle",
          symbolSize: 8,
          showSymbol: true,
          lineStyle: { width: 2, color: "#409eff" },
          itemStyle: {
            color: "#ffffff",
            borderColor: "#409eff",
            borderWidth: 2
          },
          label: {
            show: true,
            position: "top",
            color: "#606266",
            fontSize: 11
          }
        }
      ]
    },
    true
  );
}

async function loadOverview() {
  loading.value = true;
  try {
    const response = await getDashboardOverview();
    overview.value = response.data;
  } finally {
    loading.value = false;
  }
}

async function loadJudgeStatus() {
  if (judgeLoading.value) return;
  judgeLoading.value = true;
  try {
    const response = await getJudgeStatus();
    judgeNodes.value = response.data?.list ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    judgeLoading.value = false;
  }
}

async function refreshDashboard() {
  await Promise.allSettled([loadOverview(), loadJudgeStatus()]);
}

watch(() => overview.value?.submissionTrend, renderTrendChart, { deep: true });
useResizeObserver(trendChartRef, () => trendChart?.resize());

onMounted(() => {
  renderTrendChart();
  refreshDashboard();
});

onBeforeUnmount(() => {
  trendChart?.dispose();
  trendChart = undefined;
});
</script>

<template>
  <div v-loading="loading" class="dashboard-page">
    <header class="dashboard-header">
      <div>
        <h2>工作台</h2>
        <p>查看平台运行概况和需要处理的事项</p>
      </div>
      <el-button @click="refreshDashboard">刷新数据</el-button>
    </header>

    <section class="summary-grid">
      <button
        v-for="card in summaryCards"
        :key="card.label"
        type="button"
        class="summary-card"
        @click="router.push(card.path)"
      >
        <span :class="['summary-icon', card.tone]"
          ><el-icon><component :is="card.icon" /></el-icon
        ></span>
        <span
          ><small>{{ card.label }}</small
          ><strong>{{ card.value }}</strong></span
        >
      </button>
    </section>

    <section class="panel trend-panel">
      <header>
        <div>
          <h3>近 7 天提交趋势</h3>
          <p>全站所有用户，按提交创建时间统计</p>
        </div>
      </header>
      <div
        ref="trendChartRef"
        class="trend-chart"
        role="img"
        aria-label="近七天提交数量折线图"
      />
    </section>

    <section class="panel quick-panel">
      <header>
        <div>
          <h3>快捷入口</h3>
          <p>常用后台操作</p>
        </div>
      </header>
      <div class="quick-list">
        <button
          v-for="action in quickActions"
          :key="action.path"
          type="button"
          @click="router.push(action.path)"
        >
          <span :class="['quick-icon', action.tone]">
            <el-icon><component :is="action.icon" /></el-icon>
          </span>
          <span class="quick-content">
            <strong>{{ action.title }}</strong>
            <small>{{ action.description }}</small>
          </span>
          <span class="quick-arrow">›</span>
        </button>
      </div>
    </section>

    <section class="panel judge-panel">
      <header>
        <div>
          <h3>评测服务状态</h3>
          <p>实时探测已配置的评测节点</p>
        </div>
        <div class="judge-actions">
          <span
            :class="[
              'judge-summary',
              judgeNodes.length > 0 && onlineJudgeCount === judgeNodes.length
                ? 'healthy'
                : 'warning'
            ]"
          >
            <i />{{ onlineJudgeCount }} / {{ judgeNodes.length }} 在线
          </span>
          <el-button text type="primary" @click="router.push('/judge/nodes')">
            查看详情
          </el-button>
          <el-button
            text
            :icon="Refresh"
            :loading="judgeLoading"
            aria-label="刷新评测服务状态"
            @click="loadJudgeStatus"
          />
        </div>
      </header>
      <div v-if="judgeNodes.length" class="judge-list">
        <article v-for="node in judgeNodes" :key="node.url" class="judge-node">
          <span :class="['judge-dot', node.online ? 'online' : 'offline']" />
          <div class="judge-node-main">
            <strong>{{ node.url }}</strong>
            <small>{{ node.version || "版本未知" }}</small>
          </div>
          <span
            v-if="node.online"
            :class="[
              'judge-latency',
              node.latencyMs < 100
                ? 'fast'
                : node.latencyMs < 500
                  ? 'medium'
                  : 'slow'
            ]"
          >
            {{ node.latencyMs }} ms
          </span>
          <el-tag v-else size="small" type="danger" effect="plain">
            离线
          </el-tag>
          <p v-if="node.error" class="judge-error">{{ node.error }}</p>
        </article>
      </div>
      <div v-else v-loading="judgeLoading" class="judge-empty">
        未配置评测服务节点
      </div>
    </section>
  </div>
</template>

<style scoped>
.dashboard-page {
  min-height: 560px;
}
.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.dashboard-header h2 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 21px;
}
.dashboard-header p,
.panel header p {
  margin: 5px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 12px;
}
.summary-card {
  min-height: 98px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter);
  background: #fff;
  text-align: left;
  transition: 0.16s;
}
.summary-card:hover {
  border-color: var(--el-color-primary-light-6);
  box-shadow: 0 5px 18px rgb(31 45 61 / 6%);
  transform: translateY(-1px);
}
.summary-card span:last-child {
  min-width: 0;
}
.summary-card small,
.summary-card strong {
  display: block;
}
.summary-card small {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  white-space: nowrap;
}
.summary-card strong {
  margin-top: 7px;
  color: var(--el-text-color-primary);
  font-size: 23px;
}
.summary-icon {
  width: 39px;
  height: 39px;
  display: grid;
  flex: 0 0 39px;
  place-items: center;
  font-size: 19px;
}
.summary-icon.blue {
  color: #3b82f6;
  background: #eff6ff;
}
.summary-icon.cyan {
  color: #0891b2;
  background: #ecfeff;
}
.summary-icon.green {
  color: #16a34a;
  background: #f0fdf4;
}
.summary-icon.orange {
  color: #ea580c;
  background: #fff7ed;
}
.summary-icon.purple {
  color: #7c3aed;
  background: #f5f3ff;
}
.summary-icon.red {
  color: #dc2626;
  background: #fef2f2;
}
.trend-panel {
  margin-top: 14px;
}
.panel {
  border: 1px solid var(--el-border-color-lighter);
  background: #fff;
}
.panel > header {
  min-height: 62px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.panel h3 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 15px;
}
.trend-chart {
  height: 270px;
  width: 100%;
}
.quick-panel {
  margin-top: 14px;
}
.quick-list {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  padding: 14px 18px 18px;
}
.quick-list button {
  min-width: 0;
  min-height: 76px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--el-border-color-lighter);
  background: #fff;
  text-align: left;
  transition: 0.16s ease;
}
.quick-list button:hover {
  border-color: var(--el-color-primary-light-6);
  background: var(--el-color-primary-light-9);
  box-shadow: 0 4px 14px rgb(31 45 61 / 6%);
  transform: translateY(-1px);
}
.quick-icon {
  width: 40px;
  height: 40px;
  display: grid;
  flex: 0 0 40px;
  place-items: center;
  font-size: 19px;
}
.quick-icon.blue {
  color: #2563eb;
  background: #eff6ff;
}
.quick-icon.purple {
  color: #7c3aed;
  background: #f5f3ff;
}
.quick-icon.orange {
  color: #ea580c;
  background: #fff7ed;
}
.quick-icon.green {
  color: #16a34a;
  background: #f0fdf4;
}
.quick-content {
  min-width: 0;
  flex: 1;
}
.quick-content strong,
.quick-content small {
  display: block;
}
.quick-content strong {
  color: var(--el-text-color-primary);
  font-size: 13px;
}
.quick-content small {
  margin-top: 5px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
  line-height: 1.55;
}
.quick-arrow {
  color: var(--el-text-color-placeholder);
  font-size: 21px;
}
.judge-panel {
  margin-top: 14px;
}
.judge-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}
.judge-summary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.judge-summary i,
.judge-dot {
  width: 7px;
  height: 7px;
  flex: 0 0 7px;
  border-radius: 50%;
}
.judge-summary.healthy i,
.judge-dot.online {
  background: var(--el-color-success);
  box-shadow: 0 0 0 3px var(--el-color-success-light-9);
}
.judge-summary.warning i,
.judge-dot.offline {
  background: var(--el-color-danger);
  box-shadow: 0 0 0 3px var(--el-color-danger-light-9);
}
.judge-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(310px, 1fr));
  gap: 10px;
  padding: 14px 18px 18px;
}
.judge-node {
  min-width: 0;
  min-height: 66px;
  display: grid;
  grid-template-columns: 10px minmax(0, 1fr) auto;
  align-items: center;
  column-gap: 11px;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-extra-light);
}
.judge-node-main {
  min-width: 0;
}
.judge-node-main strong,
.judge-node-main small {
  display: block;
}
.judge-node-main strong {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.judge-node-main small {
  margin-top: 5px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}
.judge-latency {
  font-size: 12px;
  font-weight: 600;
}
.judge-latency.fast {
  color: var(--el-color-success);
}
.judge-latency.medium {
  color: var(--el-color-warning);
}
.judge-latency.slow {
  color: var(--el-color-danger);
}
.judge-error {
  grid-column: 2 / 4;
  margin: 7px 0 0;
  overflow: hidden;
  color: var(--el-color-danger);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.judge-empty {
  min-height: 88px;
  display: grid;
  place-items: center;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
@media (max-width: 1280px) {
  .summary-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (max-width: 900px) {
  .quick-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
