<template>
  <div class="page-container py-4">
    <!-- 作业头部 + tab 合并成一张卡片，同团队页 -->
    <div class="f-panel">
      <div class="hw-band">
        <div class="hw-band__row">
          <div class="hw-band__main">
            <div class="flex items-center gap-2 text-sm text-gray-500">
              <router-link class="hover:text-blue-500" :to="`/team/${hw.teamId}`">{{ hw.teamName }}</router-link>
              <span class="text-gray-300">/</span>
            </div>
            <div class="mt-1.5 flex flex-wrap items-center gap-3">
              <h1 class="text-xl font-semibold text-gray-800">{{ hw.title }}</h1>
              <el-tag size="small" :type="statusMeta(hw.status).type" effect="light">{{ statusMeta(hw.status).text }}</el-tag>
              <span class="hw-time">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
                  <circle cx="12" cy="12" r="8.5" />
                  <path d="M12 7.8V12l3 1.8" />
                </svg>
                <span>{{ fmtShort(hw.startTime) }} 至 {{ fmtShort(hw.endTime) }}</span>
              </span>
            </div>
          </div>
          <div class="hw-stats">
            <div class="hw-stat">
              <span>题数</span>
              <b>{{ hw.problemCount }}</b>
            </div>
            <div v-if="hw.myScore != null" class="hw-stat">
              <span>完成度</span>
              <b>{{ hw.solvedCount ?? 0 }}<span class="of"> / {{ hw.problemCount }}</span></b>
            </div>
            <div v-if="hw.myScore != null" class="hw-stat">
              <span>得分</span>
              <b>{{ hw.myScore }}<span class="of"> / {{ hw.totalScore }}</span></b>
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-6 border-t border-gray-100 px-6">
        <router-link v-for="tab in tabs" :key="tab.name" :to="tab.to"
          class="flex items-center gap-1.5 border-b-2 border-transparent py-3 text-[15px] font-medium text-gray-700 hover:text-blue-500"
          exact-active-class="border-blue-500! text-blue-500! font-medium">
          {{ tab.label }}
        </router-link>
      </div>
    </div>

    <!-- 内容区不包 f-panel，各 tab 自己渲染独立卡片，间距 mt-4 -->
    <div class="mt-4">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, provide, watch } from "vue";
import { useRoute } from "vue-router";
import { getHomeworkDetail } from "@/api/team";

const route = useRoute();
const hw = ref({});

// 供四个 tab 注入使用，避免重复请求
provide("homework", hw);

const tabs = computed(() => {
  const base = `/team/${route.params.id}/homework/${route.params.hid}`;
  return [
    { name: "HomeworkIntro", label: "简介", to: base },
    { name: "HomeworkProblems", label: "题目列表", to: `${base}/problems` },
    { name: "HomeworkRank", label: "排行榜", to: `${base}/rank` },
    { name: "HomeworkSubmissions", label: "提交", to: `${base}/submissions` },
  ];
});

const statusMeta = (s) => ({
  0: { text: "未开始", type: "warning" },
  1: { text: "进行中", type: "success" },
  2: { text: "已截止", type: "info" },
}[s] || { text: "-", type: "info" });

// 同年省略年份，时间精确到分钟
const fmtShort = (s) => {
  if (!s) return "-";
  return s.slice(0, 4) === String(new Date().getFullYear()) ? s.slice(5, 16) : s.slice(0, 16);
};

const fetchHomework = async () => {
  try {
    const res = await getHomeworkDetail(route.params.hid);
    hw.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

watch(() => route.params.hid, fetchHomework, { immediate: true });
</script>

<style scoped>
.page-container {
  max-width: 1360px;
}

.hw-band {
  padding: 20px 24px 16px;
  background: linear-gradient(180deg, #f8fbff 0%, #fff 100%);
}

.hw-band__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.hw-band__main {
  min-width: 0;
}

.hw-stats {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
}

.hw-stat {
  min-width: 96px;
  padding: 0 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  border-left: 1px solid #e8eef5;
}

.hw-stat:first-child {
  border-left: 0;
}

.hw-stat span {
  color: #374151;
  font-size: 14px;
  font-weight: 600;
}

.hw-stat b {
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.2;
}

.hw-stat .of {
  margin-left: 2px;
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
}

.hw-time {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 14px;
}

.hw-time svg {
  width: 14px;
  height: 14px;
  flex: 0 0 auto;
  color: #94a3b8;
}

@media (max-width: 900px) {
  .hw-band__row {
    flex-wrap: wrap;
  }

  .hw-stats {
    width: 100%;
  }
}
</style>
