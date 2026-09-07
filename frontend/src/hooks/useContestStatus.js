import { computed } from "vue";
import { useNow } from "@vueuse/core";
import { CONTEST_STATUS } from "@/constants/index";

export function useContestStatus(contestData) {
  const now = useNow({
    interval: 1000,
  });

  const start = computed(() => new Date(contestData.value?.startTime).getTime());

  const end = computed(() => new Date(contestData.value?.endTime).getTime());

  const status = computed(() => {
    const current = now.value.getTime();

    if (current < start.value) {
      return CONTEST_STATUS.NOT_STARTED;
    }

    if (current < end.value) {
      return CONTEST_STATUS.RUNNING;
    }

    return CONTEST_STATUS.ENDED;
  });

  const remainTime = computed(() => {
    const current = now.value.getTime();

    if (status.value === CONTEST_STATUS.NOT_STARTED) {
      return start.value - current;
    }

    if (status.value === CONTEST_STATUS.RUNNING) {
      return end.value - current;
    }

    return 0;
  });

  const statusText = computed(() => {
    switch (status.value) {
      case CONTEST_STATUS.NOT_STARTED:
        return `未开始`;
      case CONTEST_STATUS.RUNNING:
        return `进行中`;
      case CONTEST_STATUS.ENDED:
        return "已结束";
      default:
        return "已结束";
    }
  });

  const statusTime = computed(() => {
    switch (status.value) {
      case CONTEST_STATUS.NOT_STARTED:
        return formatDuration(remainTime.value);
      case CONTEST_STATUS.RUNNING:
        return formatDuration(remainTime.value);
      default:
        return "比赛已结束";
    }
  });

  // 已进行百分比（0~100），供进度条用；未开始 0、已结束 100
  const progress = computed(() => {
    const s = start.value;
    const e = end.value;
    const cur = now.value.getTime();
    if (!Number.isFinite(s) || !Number.isFinite(e) || e <= s) return 0;
    if (cur <= s) return 0;
    if (cur >= e) return 100;
    return ((cur - s) / (e - s)) * 100;
  });

  const statusTimeText = computed(() => {
    switch (status.value) {
      case CONTEST_STATUS.NOT_STARTED:
        return `距离比赛开始还有 ${formatDurationText(remainTime.value)}`;
      case CONTEST_STATUS.RUNNING:
        return `距离比赛结束还有 ${formatDurationText(remainTime.value)}`;
      default:
        return "比赛已结束";
    }
  });

  return {
    status,
    statusText,
    statusTime,
    statusTimeText,
    remainTime,
    progress,
  };
}

function formatDurationText(ms) {
  const totalSeconds = Math.floor(ms / 1000);

  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const resultText = [];

  if (days) {
    resultText.push(`${days}天`);
  }

  if (hours) {
    resultText.push(`${hours}小时`);
  }

  if (minutes) {
    resultText.push(`${minutes}分钟`);
  }

  if (resultText.length === 0) {
    return "不足1分钟";
  }

  return resultText.join("")
}



function formatDuration(ms) {
  const totalSeconds = Math.floor(ms / 1000);

  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const pad = (value) => String(value).padStart(2, "0");

  return `${pad(hours)}:${pad(minutes)}`
}
