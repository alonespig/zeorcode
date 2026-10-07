<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import {
  Collection,
  DataAnalysis,
  DocumentChecked,
  Loading,
  Postcard,
  User
} from "@element-plus/icons-vue";
import {
  getDashboardOverview,
  type DashboardOverview
} from "@/api/admin/dashboard";

defineOptions({ name: "AdminDashboard" });

const router = useRouter();
const loading = ref(false);
const overview = ref<DashboardOverview>();

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

const maxTrendCount = computed(() =>
  Math.max(
    1,
    ...(overview.value?.submissionTrend.map(item => item.count) ?? [1])
  )
);

const statusMap: Record<
  number,
  { label: string; type: "success" | "warning" | "danger" | "info" }
> = {
  0: { label: "等待评测", type: "info" },
  1: { label: "Accepted", type: "success" },
  2: { label: "Memory Limit", type: "danger" },
  3: { label: "Time Limit", type: "danger" },
  4: { label: "Runtime Error", type: "danger" },
  5: { label: "Wrong Answer", type: "danger" },
  6: { label: "Compile Error", type: "danger" },
  7: { label: "Unknown", type: "info" },
  8: { label: "Presentation Error", type: "warning" }
};

function statusInfo(status: number) {
  return (
    statusMap[status] || { label: `状态 ${status}`, type: "info" as const }
  );
}

function shortDate(value: string) {
  return value.slice(5).replace("-", "/");
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

onMounted(loadOverview);
</script>

<template>
  <div v-loading="loading" class="dashboard-page">
    <header class="dashboard-header">
      <div>
        <h2>工作台</h2>
        <p>查看平台运行概况和需要处理的事项</p>
      </div>
      <el-button @click="loadOverview">刷新数据</el-button>
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

    <div class="dashboard-grid">
      <section class="panel trend-panel">
        <header>
          <div>
            <h3>近 7 天提交趋势</h3>
            <p>按提交创建时间统计</p>
          </div>
        </header>
        <div class="trend-chart">
          <div
            v-for="item in overview?.submissionTrend || []"
            :key="item.date"
            class="trend-column"
          >
            <span class="trend-value">{{ item.count }}</span>
            <div class="bar-track">
              <span
                :style="{
                  height: `${Math.max(3, (item.count / maxTrendCount) * 100)}%`
                }"
              />
            </div>
            <small>{{ shortDate(item.date) }}</small>
          </div>
        </div>
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
            type="button"
            @click="router.push('/content/problems/create')"
          >
            <strong>新建题目</strong><span>录入题面与测试数据</span>
          </button>
          <button
            type="button"
            @click="router.push('/content/contests/create')"
          >
            <strong>创建比赛</strong><span>配置赛制和比赛题目</span>
          </button>
          <button type="button" @click="router.push('/notifications')">
            <strong>发布通知</strong><span>向全部用户发送系统消息</span>
          </button>
          <button type="button" @click="router.push('/agent')">
            <strong>AI 助手</strong><span>通过对话完成教学任务</span>
          </button>
        </div>
      </section>
    </div>

    <section class="panel recent-panel">
      <header>
        <div>
          <h3>最近提交</h3>
          <p>平台最新的 8 条评测记录</p>
        </div>
        <el-button
          text
          type="primary"
          @click="router.push('/judge/submissions')"
          >查看全部</el-button
        >
      </header>
      <el-table
        :data="overview?.recentSubmissions || []"
        empty-text="暂无提交记录"
      >
        <el-table-column prop="id" label="提交号" width="130" align="center" />
        <el-table-column label="用户" min-width="150">
          <template #default="{ row }"
            ><strong>{{ row.username }}</strong
            ><small class="secondary">#{{ row.userID }}</small></template
          >
        </el-table-column>
        <el-table-column label="题目" min-width="220">
          <template #default="{ row }"
            ><span class="problem-id">{{ row.problemID }}</span
            >{{ row.problemName }}</template
          >
        </el-table-column>
        <el-table-column label="结果" width="160" align="center">
          <template #default="{ row }"
            ><el-tag :type="statusInfo(row.status).type" effect="plain">{{
              statusInfo(row.status).label
            }}</el-tag></template
          >
        </el-table-column>
        <el-table-column
          prop="createdAt"
          label="提交时间"
          width="175"
          align="center"
        />
      </el-table>
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
.dashboard-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(300px, 0.8fr);
  gap: 14px;
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
  height: 235px;
  display: flex;
  align-items: stretch;
  gap: 13px;
  padding: 25px 25px 18px;
}
.trend-column {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.trend-value {
  height: 23px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}
.bar-track {
  width: min(34px, 65%);
  min-height: 130px;
  flex: 1;
  display: flex;
  align-items: flex-end;
  background: var(--el-fill-color-light);
}
.bar-track span {
  width: 100%;
  display: block;
  min-height: 3px;
  background: var(--el-color-primary);
  transition: height 0.3s ease;
}
.trend-column small {
  margin-top: 8px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
}
.quick-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1px;
  background: var(--el-border-color-lighter);
}
.quick-list button {
  min-height: 117px;
  padding: 17px;
  background: #fff;
  text-align: left;
}
.quick-list button:hover {
  background: var(--el-fill-color-extra-light);
}
.quick-list strong,
.quick-list span {
  display: block;
}
.quick-list strong {
  color: var(--el-color-primary);
  font-size: 13px;
}
.quick-list span {
  margin-top: 7px;
  color: var(--el-text-color-secondary);
  font-size: 11px;
  line-height: 1.55;
}
.recent-panel {
  margin-top: 14px;
}
.secondary {
  margin-left: 7px;
  color: var(--el-text-color-placeholder);
  font-weight: 400;
}
.problem-id {
  margin-right: 9px;
  color: var(--el-color-primary);
  font-weight: 600;
}
@media (max-width: 1280px) {
  .summary-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
@media (max-width: 900px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
