<template>
  <div>
    <div class="my-rank">
      <template v-if="contest?.frozen">
        OI 赛制赛中封榜，成绩将在比赛结束后公布。
      </template>
      <template v-else-if="myRank.found">
        我的排名：<b>第 {{ myRank.rank }} 名</b>
        <template v-if="isScored">
          <span class="sep">·</span> 总分 {{ myRank.totalScore }}
        </template>
        <template v-else>
          <span class="sep">·</span> 通过 {{ myRank.acCount }}
          <span class="sep">·</span> 罚时 {{ myRank.penalty }}
        </template>
      </template>
      <template v-else>你还没有该比赛的名次记录</template>
    </div>

    <component :is="board" :contest="contest" :problems="problems" :rank-list="rankList" />
  </div>
</template>

<script setup>
import { getContestRank, getMyContestRank } from '@/api/contest'
import { useRoute } from 'vue-router'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import AcmRankBoard from './components/AcmRankBoard.vue'
import OiRankBoard from './components/OiRankBoard.vue'

const route = useRoute()

const contest = ref({})
const problems = ref([])
const rankList = ref([])
const myRank = ref({ found: false })

// rule 非 ACM（OI/IOI）走得分榜
const isScored = computed(() => !!contest.value?.rule && contest.value.rule !== 'ACM')
const board = computed(() => (isScored.value ? OiRankBoard : AcmRankBoard))

let timer = null

const loadRank = async () => {
  try {
    const res = await getContestRank(route.params.id)
    contest.value = res.data.contest || {}
    problems.value = res.data.problems || []
    rankList.value = res.data.rankList || []
    const mine = await getMyContestRank(route.params.id)
    myRank.value = mine.data || { found: false }
  } catch (err) {
    console.error(err)
  }
}

onMounted(() => {
  loadRank()
  timer = setInterval(loadRank, 5000) // 轮询刷新（与后端缓存 TTL 对齐）
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped lang="scss">
.my-rank {
  padding: 8px 12px;
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
  background: #f4f8ff;
  border: 1px solid #d6e4ff;
  border-radius: 4px;

  b {
    color: #1f6feb;
  }

  .sep {
    margin: 0 8px;
    color: #c0c4cc;
  }
}
</style>
