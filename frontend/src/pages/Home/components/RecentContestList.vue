<script setup>
import { computed } from 'vue'
import { useNow } from '@vueuse/core'
import dayjs from 'dayjs'
import { ArrowRight, Calendar, Clock, Trophy, User } from '@element-plus/icons-vue'
import { CONTEST_STATUS } from '@/constants/index'

const props = defineProps({
  contests: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['openContest', 'openAll'])
const now = useNow({ interval: 1000 })

const recentContests = computed(() => {
  const current = now.value.getTime()

  return props.contests
    .map((contest) => createViewModel(contest, current))
    .filter((contest) => contest.status !== CONTEST_STATUS.ENDED)
    .sort((a, b) => {
      if (a.status !== b.status) return b.status - a.status
      return a.startAt - b.startAt
    })
    .slice(0, 3)
})

function createViewModel(contest, current) {
  const startAt = new Date(contest.startTime).getTime()
  const endAt = new Date(contest.endTime).getTime()
  const status = current < startAt
    ? CONTEST_STATUS.NOT_STARTED
    : current < endAt
      ? CONTEST_STATUS.RUNNING
      : CONTEST_STATUS.ENDED
  const remainTime = status === CONTEST_STATUS.NOT_STARTED ? startAt - current : endAt - current

  return {
    ...contest,
    startAt,
    status,
    startText: Number.isFinite(startAt) ? dayjs(startAt).format('MM-DD HH:mm') : '--',
    countdownLabel: status === CONTEST_STATUS.RUNNING ? '距离结束' : '距离开始',
    countdownText: formatCountdown(remainTime),
  }
}

function formatCountdown(ms) {
  if (!Number.isFinite(ms) || ms <= 0) return '00:00:00'

  const totalSeconds = Math.floor(ms / 1000)
  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (days > 0) return `${days} 天 ${hours} 小时`

  const pad = (value) => String(value).padStart(2, '0')
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
}
</script>

<template>
  <section class="recent-contests f-card">
    <div class="f-card-head flex items-center gap-2.5">
      <!-- <span class="flex h-8 w-8 items-center justify-center rounded-md text-blue-500"> -->
      <div class="f-card-head f-blue">
        <span class="f-blue">
          <el-icon :size="18">
            <Trophy />
          </el-icon>
        </span>
        <b class="text-[19px] font-medium text-blue-500">近期比赛</b>
      </div>

      <button type="button"
        class="ml-auto inline-flex items-center gap-0.5 text-[13px] text-blue-500 transition-colors hover:text-blue-400"
        @click="emit('openAll')">
        查看全部 <el-icon :size="13">
          <ArrowRight />
        </el-icon>
      </button>
    </div>

    <div v-if="recentContests.length" class="mt-3 flex flex-col gap-2.5">
      <button v-for="contest in recentContests" :key="contest.id" type="button"
        class="contest-summary group relative w-full overflow-hidden border border-gray-200 bg-white px-3.5 py-3 text-left"
        :aria-label="`查看比赛：${contest.name}`" @click="emit('openContest', contest.id)">
        <span class="absolute inset-y-0 left-0 w-[3px]"
          :class="contest.status === CONTEST_STATUS.RUNNING ? 'bg-emerald-500' : 'bg-blue-400'" />

        <div class="flex min-w-0 items-center gap-2">
          <span class="h-2 w-2 shrink-0 rounded-full"
            :class="contest.status === CONTEST_STATUS.RUNNING ? 'bg-emerald-500' : 'bg-blue-400'" />
          <b class="truncate text-[16px] font-medium text-gray-700 transition-colors group-hover:text-blue-600">
            {{ contest.name }}
          </b>
          <el-icon :size="14" class="ml-auto shrink-0 text-gray-300 transition-colors group-hover:text-blue-400">
            <ArrowRight />
          </el-icon>
        </div>

        <div class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-[13px] text-gray-600">
          <span class="inline-flex items-center gap-1">
            <el-icon :size="13">
              <Calendar />
            </el-icon>
            {{ contest.startText }}
          </span>
          <span class="inline-flex items-center gap-1">
            <el-icon :size="13">
              <Clock />
            </el-icon>
            {{ contest.countdownLabel }}
            <b class="font-medium tabular-nums text-gray-600">{{ contest.countdownText }}</b>
          </span>
          <span class="inline-flex items-center gap-1">
            <el-icon :size="13">
              <User />
            </el-icon>
            {{ contest.participants ?? 0 }} 人
          </span>
        </div>
      </button>
    </div>

    <div v-else class="py-7 text-center text-sm text-gray-400">
      近期暂无比赛
    </div>
  </section>
</template>

<style scoped>
.recent-contests {
  width: 100%;
}

.f-card-head {
  display: flex;
  align-items: center;
  justify-items: center;
  gap: 6px;
  font-size: 19px;
}

.contest-summary {
  border-radius: var(--radius-small);
  transition: border-color 160ms ease, box-shadow 160ms ease, transform 160ms ease;
}

.contest-summary:hover {
  border-color: #bfdbfe;
  box-shadow: var(--shadow-soft);
  transform: translateY(-1px);
}

@media (prefers-reduced-motion: reduce) {
  .contest-summary {
    transition: none;
  }

  .contest-summary:hover {
    transform: none;
  }
}
</style>
