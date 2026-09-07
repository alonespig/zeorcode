<template>
  <div class="flex items-center gap-[18px] px-[22px] py-[18px] max-md:grid max-md:grid-cols-[72px_minmax(0,1fr)] max-md:items-start max-md:gap-x-3 max-md:gap-y-3 max-md:px-3 max-md:py-4">
    <!-- 左：封面图（比赛没上传封面 → 默认 ICPC 图；以后创建比赛可上传，取 contest.coverUrl） -->
    <div class="flex h-[100px] w-[100px] shrink-0 max-md:h-18 max-md:w-18
      cursor-pointer items-center justify-center
      overflow-hidden rounded-xl border border-gray-100 bg-white max-md:rounded-lg"
      @click="emit('goContest', contest.id)">
      <img :src="icpcCover" alt="cover" class="h-full w-full object-contain p-[7px]" />
    </div>

    <!-- 中：标题 + 标签 + 信息行 -->
    <div class="min-w-0 flex flex-1 flex-col gap-4 max-md:gap-2.5">
      <div class="flex flex-wrap items-center gap-2">
        <span class="cursor-pointer text-[21px] font-medium text-gray-700 hover:text-blue-600 max-md:text-[18px]"
          @click="emit('goContest', contest.id)">
          {{ contest?.name }}
        </span>

        <!-- 赛制 -->
        <span v-if="contest?.type" class="rounded px-[7px] py-px text-[11.5px] font-medium text-white"
          :class="typeTagClass">
          {{ contest.type }}
        </span>
        <!-- 非公开 -->
        <span v-if="contest?.needInviteCode"
          class="inline-flex items-center gap-[3px] rounded border border-gray-200 bg-gray-50 px-[7px] py-px text-[11.5px] text-gray-500">
          <el-icon :size="12"><Lock /></el-icon>非公开
        </span>
        <!-- Rated -->
        <span v-if="contest?.rated" class="rounded bg-[#f0932b] px-[7px] py-px text-[11.5px] font-medium text-white">
          ★ Rated
        </span>
        <!-- 状态 -->
        <span class="rounded px-[7px] py-px text-[11.5px] font-medium text-white" :class="statusTag.cls">
          {{ statusTag.text }}
        </span>
      </div>

      <div class="flex gap-[20px] text-[14px] text-gray-600 max-md:flex-col max-md:gap-1.5 max-md:text-[13px]">
        <span class="inline-flex items-center gap-1.5">
          <el-icon :size="14" class="text-gray-600"><Calendar /></el-icon>
          <span class="text-gray-600">开始：{{ startText }}</span>
        </span>
        <span class="inline-flex items-center gap-1.5">
          <el-icon :size="14"><Clock /></el-icon>
          <span>时长：{{ durationText }}</span>
        </span>
        <span class="inline-flex items-center gap-1.5">
          <el-icon :size="14"><User /></el-icon>
          <span>参与人数：{{ contest?.participants ?? 0 }}</span>
        </span>
      </div>
    </div>

    <!-- 右：按钮 + 倒计时 -->
    <div class="flex w-[98px] shrink-0 flex-col items-center gap-2 text-sm max-md:col-span-2 max-md:w-full max-md:flex-row max-md:justify-end max-md:gap-3">
      <el-button class="contest-action" :type="action.type" :plain="action.plain"
        @click="emit('goContest', contest.id)">
        {{ action.text }}
      </el-button>
      <span v-if="countdown" class="text-[12.5px] tabular-nums text-gray-400">
        {{ countdown.label }} <b class="font-semibold" :class="countdown.color">{{ countdown.time }}</b>
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed, toRef } from "vue";
import { Clock, User, Lock, Calendar } from "@element-plus/icons-vue";
import { useContestStatus } from "@/hooks/useContestStatus";
import { CONTEST_STATUS } from "@/constants/index";
import icpcCover from "@/image/icpc.png";
import dayjs from "dayjs";

const props = defineProps({
  contest: {
    type: Object,
    required: true,
  },
});
const emit = defineEmits(["goContest"]);

const contestRef = toRef(props, "contest");
const { status, remainTime } = useContestStatus(contestRef);

// 开始时间：跨年才带年份
const startText = computed(() => {
  const t = props.contest?.startTime;
  if (!t) return "--";
  const d = dayjs(t);
  return d.year() === dayjs().year() ? d.format("MM-DD HH:mm") : d.format("YYYY-MM-DD HH:mm");
});

// 时长：用 结束-开始 算，不依赖后端 duration 字段单位
const durationText = computed(() => {
  const s = props.contest?.startTime, e = props.contest?.endTime;
  if (!s || !e) return "-";
  const mins = Math.round((new Date(e) - new Date(s)) / 60000);
  if (mins <= 0) return "-";
  if (mins < 60) return `${mins} 分钟`;
  if (mins % 60 === 0) return `${mins / 60} 小时`;
  return `${Math.floor(mins / 60)} 小时 ${mins % 60} 分`;
});

// 赛制标签底色（注意 IOI 要在 OI 前判断）
const typeTagClass = computed(() => {
  const t = String(props.contest?.type || "").toUpperCase();
  if (t.includes("CF")) return "bg-[#c026d3]";
  if (t.includes("IOI")) return "bg-[#12b3a6]";
  if (t.includes("OI")) return "bg-[#37b24d]";
  if (t.includes("ACM")) return "bg-[#4a90e2]";
  return "bg-gray-400";
});

// 状态小标
const statusTag = computed(() => {
  switch (status.value) {
    case CONTEST_STATUS.NOT_STARTED:
      return { text: "报名中", cls: "bg-[#4dabf7]" };
    case CONTEST_STATUS.RUNNING:
      return { text: "进行中", cls: "bg-[#37b24d]" };
    default:
      return { text: "已结束", cls: "bg-[#adb5bd]" };
  }
});

// 右侧大按钮（都跳到比赛详情，报名/进入的具体逻辑在详情页处理）
const action = computed(() => {
  if (status.value === CONTEST_STATUS.RUNNING)
    return { text: "进入比赛", type: "success", plain: false };
  if (status.value === CONTEST_STATUS.ENDED)
    return { text: "查看", type: "info", plain: true };
  // 未开始
  if (props.contest?.isRegistered)
    return { text: "✓ 已报名", type: "success", plain: true };
  return { text: "报名", type: "primary", plain: false };
});

// 倒计时（未开始→距比赛；进行中→剩；已结束→无）
const countdown = computed(() => {
  if (status.value === CONTEST_STATUS.NOT_STARTED)
    return { label: "距比赛", time: fmtCountdown(remainTime.value), color: "text-[#f0932b]" };
  if (status.value === CONTEST_STATUS.RUNNING)
    return { label: "剩", time: fmtCountdown(remainTime.value), color: "text-emerald-600" };
  return null;
});

function fmtCountdown(ms) {
  if (!ms || ms <= 0) return "00:00";
  const s = Math.floor(ms / 1000);
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  const pad = (n) => String(n).padStart(2, "0");
  return d > 0 ? `${d}天${pad(h)}:${pad(m)}` : `${pad(h)}:${pad(m)}`;
}
</script>

<style scoped>
.contest-action {
  width: 98px;
  height: 32px;
}
</style>
