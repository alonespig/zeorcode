<template>
  <div>
    <table v-if="list.length" class="ch-table">
      <thead>
        <tr>
          <th class="l">比赛</th>
          <th class="c">名次</th>
          <th class="r">Rating 变化</th>
          <th class="c hide-sm">时间</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="c in list" :key="c.contestId">
          <td @click="goContest(c.contestId)" class="l">
            <span class="ch-name cursor-pointer">{{ c.contestName }}</span>
            <span class="ch-badge" :class="'b-' + typeKey(c.type)">{{ ruleLabelOf(c.type) }}</span>
          </td>

          <td class="c ch-rank">
            <template v-if="c.status === 'settled'">
              {{ c.rank }}<span v-if="c.total" class="tot"> / {{ c.total }}</span>
            </template>
            <template v-else>—</template>
          </td>
          <td class="r">
            <span v-if="c.status === 'settled'" class="ch-delta">
              <span class="old">{{ c.oldRating }}</span>
              <span class="arw">→</span>
              <span class="new" :style="{ color: ratingTier(c.newRating).color }">{{ c.newRating }}</span>
              <b class="amt" :class="c.delta >= 0 ? 'up' : 'down'">
                {{ c.delta >= 0 ? '▲ +' + c.delta : '▼ ' + fmtNeg(c.delta) }}
              </b>
            </span>
            <span v-else-if="c.status === 'calculating'" class="ch-pill calc"><i class="d"></i>计算中</span>
            <span v-else-if="c.status === 'running'" class="ch-pill live"><i class="d"></i>进行中</span>
            <span v-else-if="c.status === 'upcoming'" class="ch-muted">未开始</span>
            <span v-else class="ch-muted">不计分</span>
          </td>
           <td class="c hide-sm ch-time">{{ fmtDay(c.time) }}</td>
        </tr>
      </tbody>
    </table>

    <div v-else-if="!loading" class="ch-empty">还没有参加比赛</div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUserContestHistory } from '@/api/user'
import { ratingTier, ruleLabelOf, CONTEST_TYPE } from '@/constants/index'

const route = useRoute()
const router = useRouter()

const list = ref([])
const loading = ref(false)

const typeKey = (t) => ({
  [CONTEST_TYPE.ACM]: 'acm',
  [CONTEST_TYPE.OI]: 'oi',
  [CONTEST_TYPE.IOI]: 'ioi',
  [CONTEST_TYPE.CF]: 'cf',
}[t] || 'acm')

const fmtDay = (sec) => {
  if (!sec) return ''
  const d = new Date(Number(sec) * 1000)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

// delta 为负时本身带 '-'，替换成更好看的减号
const fmtNeg = (n) => String(n).replace('-', '−')

const goContest = (id) => router.push(`/contest/${id}`)

const load = async () => {
  loading.value = true
  try {
    const res = await getUserContestHistory(route.params.id)
    list.value = res.data?.list || []
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => route.params.id, load)
</script>

<style scoped lang="scss">
.ch-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.ch-table th {
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: #8a929c;
  padding: 0 10px 9px;
  border-bottom: 1px solid #eef1f4;
  letter-spacing: 0.03em;
}

.ch-table th.r,
.ch-table td.r {
  text-align: right;
}

.ch-table th.c,
.ch-table td.c {
  text-align: center;
}

.ch-table td {
  padding: 11px 10px;
  border-bottom: 1px solid #f2f4f6;
  vertical-align: middle;
}

// .ch-table tbody tr:first-child {
//   cursor: pointer;
// }

.ch-table tbody tr:last-child td {
  border-bottom: none;
}

.ch-table tbody tr:hover td {
  background: #fafbfc;
}

.ch-name {
  font-weight: 500;
  color: #2f6ad0;
}

.ch-badge {
  display: inline-block;
  font-size: 10.5px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  border: 1px solid;
  margin-left: 8px;
  vertical-align: 1px;
}

.b-cf { color: #c026d3; border-color: #e9b8ef; background: #fbf0fc; }
.b-acm { color: #2563eb; border-color: #bcd2f5; background: #f0f5fe; }
.b-ioi { color: #0d9488; border-color: #b4ddd7; background: #eef8f6; }
.b-oi { color: #d97706; border-color: #eccf9e; background: #fdf5e8; }

.ch-time { color: #9aa1ab; }

.ch-rank {
  font-variant-numeric: tabular-nums;
  color: #4b5563;
}

.ch-rank .tot { color: #b0b7c0; }

.ch-delta {
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.ch-delta .old { color: #aeb4bd; }
.ch-delta .arw { color: #cbd0d7; margin: 0 3px; }
.ch-delta .new { font-weight: 600; }
.ch-delta .amt { font-weight: 700; margin-left: 8px; }
.up { color: #16a34a; }
.down { color: #dc2626; }

.ch-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 999px;
  padding: 2px 10px;
  border: 1px solid;
}

.ch-pill .d {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  animation: chpulse 1.3s ease-in-out infinite;
}

.ch-pill.calc {
  color: #b4790a;
  background: #fdf5e6;
  border-color: #f0d9a8;
}

.ch-pill.calc .d { background: #e0912f; }

.ch-pill.live {
  color: #2563eb;
  background: #eef4ff;
  border-color: #c7dbfa;
}

.ch-pill.live .d { background: #2563eb; }

@keyframes chpulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.25; }
}

@media (prefers-reduced-motion: reduce) {
  .ch-pill .d { animation: none; }
}

.ch-muted {
  font-size: 12px;
  color: #a3a9b2;
}

.ch-empty {
  padding: 28px 0;
  text-align: center;
  font-size: 13px;
  color: #a3a9b2;
}

@media (max-width: 680px) {
  .ch-table .hide-sm { display: none; }
}
</style>
