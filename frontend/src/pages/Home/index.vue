<template>
  <div class="page-home page-container">
    <div class="left-container">
      <!-- 通知：取 category=announcement 的最新帖子 -->
      <div class="announcement f-card">
        <div class="flex items-center gap-2 text-[19px] text-blue-500">
          <el-icon><Bell /></el-icon>
          <span>通知</span>
          <span class="ml-auto text-[13px] cursor-pointer" @click="goAnnouncements">更多 ›</span>
        </div>
        <ul class="mt-3 flex flex-col gap-2.5">
          <li
            v-for="a in announcements"
            :key="a.id"
            class="announcement-item flex items-center justify-between gap-3 border border-gray-200 border-l-[3px] border-l-blue-400 px-4 py-3 text-[15px]"
          >
            <button
              type="button"
              class="announcement-title truncate text-left text-gray-600"
              @click="goPost(a.id)"
            >
              {{ a.title }}
            </button>
            <span class="shrink-0 font-mono text-xs text-gray-400">{{ a.createdAt }}</span>
          </li>
          <li v-if="!announcements.length" class="py-4 text-center text-sm text-gray-400">暂无通知</li>
        </ul>
      </div>

      <div v-if="userStore.isLogin" class="problem-static f-card">
        <div class="f-head f-blue">
          <el-icon>
            <DataBoard />
          </el-icon>
          <span>近一周题目通过数</span>
        </div>
        <PassCountChart :dates="datas" :counts="counts" />
      </div>

      <div class="problem-list-new f-card">
        <div class="f-head f-blue">
          <el-icon><Timer /></el-icon>
          <span>最新题目</span>
        </div>
        <LastProblemTable :problems="problemList" @click-problem="goProblem" />
      </div>
    </div>
    <div class="right-container">
      <RecentContestList
        :contests="recentContests"
        @open-contest="goContest"
        @open-all="goContests"
      />
      <MiniUserRank :rankList="rankList" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getPostList } from '@/api/post'
import { getUserRecent7DaysAc, userRank } from '@/api/user'
import { getProblemList } from '@/api/problems'
import { getContestList } from '@/api/contest'
import { CONTEST_STATUS } from '@/constants/index'
import { useUserStore } from '@/stores/user'
import PassCountChart from '@/components/user/PassCountChart.vue'
import MiniUserRank from '@/components/home/MiniUserRank.vue'
import LastProblemTable from './components/LastProblemTable.vue'
import RecentContestList from './components/RecentContestList.vue'
import { DataBoard, Timer, Bell } from '@element-plus/icons-vue'

const datas = ref([])
const counts = ref([])
const rankList = ref([])
const problemList = ref([])
const announcements = ref([])
const recentContests = ref([])

const userStore = useUserStore()
const router = useRouter()

const goProblem = (id) => {
  router.push({ name: 'ProblemDetail', params: { id } })
}

const goContest = (id) => {
  router.push({ name: 'ContestDetailView', params: { id } })
}

const goContests = () => {
  router.push({ name: 'ContestList' })
}

const goPost = (id) => {
  router.push({ name: 'BlogDetail', params: { id } })
}

// “更多” → 博客列表的通知分类
const goAnnouncements = () => {
  router.push({ name: 'BlogList', query: { category: 'announcement' } })
}

const fetchAnnouncements = async () => {
  const res = await getPostList({ category: 'announcement', page: 1, pageSize: 5 })
  announcements.value = res.data.list || []
}

const getRecent7DaysAc = async () => {
  // 未登录：没有"我的"过题数可拉，清空即可（拉 /user/0/ac-stats 会被后端当成用户不存在而报错）
  if (!userStore.isLogin) {
    datas.value = []
    counts.value = []
    return
  }
  const res = await getUserRecent7DaysAc(userStore.user.id)
  datas.value = res.data.dates
  counts.value = res.data.counts
}

const getRankList = async () => {
  const res = await userRank({ page: 1, pageSize: 10 })
  rankList.value = res.data.list || []
}

const getLastProblemList = async () => {
  const res = await getProblemList({ page: 1, pageSize: 10, order: 'desc' })
  problemList.value = res.data.list || []
}

const getRecentContestList = async () => {
  const [runningRes, upcomingRes] = await Promise.all([
    getContestList({ page: 1, pageSize: 3, status: CONTEST_STATUS.RUNNING }),
    getContestList({ page: 1, pageSize: 3, status: CONTEST_STATUS.NOT_STARTED }),
  ])
  const contests = [...(runningRes.data.list || []), ...(upcomingRes.data.list || [])]
  recentContests.value = contests.filter(
    (contest, index, list) => list.findIndex((item) => item.id === contest.id) === index,
  )
}

onMounted(() => {
  fetchAnnouncements()
  getRecent7DaysAc()
  getRankList()
  getLastProblemList()
  getRecentContestList()
})
</script>

<style scoped lang="scss">
.page-home {
  width: 100%;
  max-width: 1320px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(370px, 390px);
  gap: 24px;
  align-items: flex-start;

  .left-container {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 20px;

    .problem-static,
    .problem-list-new {
      width: 100%;
      margin-bottom: 0;
      font-size: 19px;

      .f-head {
        display: flex;
        align-items: center;
        gap: 6px;
      }
    }
  }

  .right-container {
    min-width: 0;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
}

.announcement-item {
  border-radius: var(--radius-small);
}

.announcement-title {
  min-width: 0;
  cursor: pointer;
  transition: color 150ms ease;
}

.announcement-title:hover,
.announcement-title:focus-visible {
  color: var(--color-primary);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.announcement-title:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 35%, transparent);
  outline-offset: 3px;
  border-radius: var(--radius-small);
}

@media (max-width: 900px) {
  .page-home {
    grid-template-columns: minmax(0, 1fr);
    gap: 20px;
  }
}

@media (max-width: 767px) {
  .page-home {
    gap: 12px;

    .left-container {
      gap: 12px;

      .problem-static,
      .problem-list-new {
        margin-bottom: 0;
      }
    }

    .right-container {
      gap: 12px;
    }
  }
}
</style>
