<template>
  <div class="contest-time-info f-shadow">
    <div class="contest-name">
      <span>{{ contestData?.name }}</span>
    </div>

    <div class="time-line">
      <span class="time-item">
        <span class="time-title">开始时间：</span>
        {{ fmt(contestData?.startTime) }}
      </span>
      <span class="time-item">
        <span class="time-title">结束时间：</span>
        {{ fmt(contestData?.endTime) }}
      </span>
    </div>

    <div class="slider">
      <el-slider v-model="progressValue" :format-tooltip="formatTooltip"
        :style="{ '--el-slider-main-bg-color': 'var(--judge-accepted)', '--el-slider-height': '10px' }" />
    </div>

    <div class="time-desc">
      <el-tag effect="dark" :type="statusTagType" size="large">
        <span v-if="progress.status === 0">
          距离开始还有 {{ pad(progress.hour) }}:{{ pad(progress.minute) }}
        </span>
        <span v-else-if="progress.status === 1">
          距离结束还有 {{ pad(progress.hour) }}:{{ pad(progress.minute) }}
        </span>
        <span v-else>已结束</span>
      </el-tag>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from "vue";
import dayjs from "dayjs";

const props = defineProps({
  contestData: {
    type: Object,
    default: () => ({ startTime: 0, endTime: 0, name: " " }),
  },
});

const now = ref(Date.now());
let timer = null;

const contestData = ref({ startTime: 0, endTime: 0, name: " " });
watch(() => props.contestData, (newVal) => { contestData.value = newVal; });

const progress = computed(() => {
  const nowTime = now.value;
  const start = contestData.value.startTime;
  const end = contestData.value.endTime;
  const total = end - start;

  if (!start || !end || total <= 0) {
    return { status: -1, nowProgress: 0 };
  }
  // 未开始
  if (nowTime < start) {
    const remain = start - nowTime;
    return {
      status: 0,
      nowProgress: 0,
      hour: remain / 1000 / 3600,
      minute: (remain / 1000 / 60) % 60,
      second: (remain / 1000) % 60,
    };
  }
  // 进行中
  if (nowTime < end) {
    const passed = nowTime - start;
    const remain = end - nowTime;
    return {
      status: 1,
      nowProgress: (passed / total) * 100,
      hour: remain / 1000 / 3600,
      minute: (remain / 1000 / 60) % 60,
      second: (remain / 1000) % 60,
    };
  }
  // 已结束
  return { status: 2, nowProgress: 100, hour: 0, minute: 0, second: 0 };
});

// 只读进度：getter 返回进度百分比，setter 为空 —— 滑块不可拖动改值
const progressValue = computed({
  get: () => progress.value.nowProgress || 0,
  set: () => { },
});

const statusTagType = computed(
  () => ({ 0: "warning", 1: "success", 2: "info" }[progress.value.status] || "info")
);

const pad = (n) => String(Math.floor(n || 0)).padStart(2, "0");

const fmt = (t) => (t ? dayjs(t).format("YYYY-MM-DD HH:mm") : "--");

const secFmt = (s) => {
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  return `${pad(h)}:${pad(m)}`;
};

// tooltip 显示滑块当前位置对应的「已进行时间」
const formatTooltip = (val) => {
  const start = contestData.value.startTime;
  const end = contestData.value.endTime;
  const total = Math.max(0, (end - start) / 1000);
  return secFmt((val / 100) * total);
};

onMounted(() => {
  timer = setInterval(() => { now.value = Date.now(); }, 1000);
});
onUnmounted(() => clearInterval(timer));
</script>

<style lang="scss" scoped>
.contest-time-info {
  width: 100%;
  background-color: #ffffff;
  padding: 16px 24px;
  border-radius: 4px;

  .contest-name {
    width: 100%;
    text-align: center;
    margin-bottom: 14px;
    font-size: 26px;
    font-weight: 600;
    color: #000;
    white-space: normal;
    word-break: break-word;
  }

  .time-line {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 14px;
    margin-bottom: 2px;

    .time-title {
      font-weight: 600;
      color: #000;
    }
  }

  .slider {
    padding: 0 6px;
  }

  .time-desc {
    text-align: center;
    margin-top: 8px;
  }
}
</style>
