<template>
  <div class="rank-page page-container">
    <div class="f-panel overflow-hidden pb-2">
      <div class="flex justify-between items-center px-10 py-5 border-b border-gray-200">
        <div class="flex items-center gap-6">
          <span class="tw-title">排行榜</span>
          <el-radio-group v-model="mode" @change="onModeChange">
            <el-radio-button value="ac">过题数</el-radio-button>
            <el-radio-button value="rating">Rating</el-radio-button>
          </el-radio-group>
        </div>
        <div v-if="mode === 'ac'" class="flex items-center gap-6">
          <el-input v-model="q" class="rounded h-8.5" placeholder="用户名" :prefix-icon="Search" style="width: 200px;" />
          <el-button @click="handleSearch" :icon="Search" class="rounded-xs!" type="primary">搜索</el-button>
        </div>
      </div>

      <div v-loading="loading" class="rank-content" :class="{ 'rank-content--empty': !list.length }">
        <UserRank v-if="mode === 'ac'" :userRankList="list" @click-user="goUser" />
        <RatingRankTable v-else :list="list" @click-user="goUser" />

        <el-empty v-if="!loading && !list.length" :image-size="80" :description="emptyDescription" />
      </div>

      <div v-if="total > pageSize" class="f-pagination">
        <el-pagination layout="prev, pager, next" :current-page="page" :page-size="pageSize" :total="total"
          @current-change="handlePageChange" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, shallowRef, watch } from "vue";
import { useRoute, useRouter } from 'vue-router'
import { userRank, ratingRank } from "@/api/user";
import UserRank from "@/components/rank/UserRank.vue";
import RatingRankTable from "@/components/rank/RatingRankTable.vue";
import { Search } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const q = ref(route.query.q || '')
const mode = ref(route.query.tab === 'rating' ? 'rating' : 'ac')

const page = computed(() => Number(route.query.page || 1))
const pageSize = computed(() => Number(route.query.pageSize || 15))

const list = ref([]);
const total = ref(0)
const loading = shallowRef(true)

const emptyDescription = computed(() => {
  if (mode.value === 'rating') return '暂无 Rating 排名'
  if (route.query.q) return '没有找到相关用户'
  return '暂无排名数据'
})

const goUser = (id) => router.push({ name: 'User', params: { id } })

// 切换榜单：回到第 1 页，清掉搜索
const onModeChange = (val) => {
  router.push({ query: { tab: val === 'rating' ? 'rating' : undefined, page: 1, pageSize: pageSize.value } })
}

const handleSearch = () => {
  router.push({ query: { ...route.query, q: q.value || undefined, page: 1, pageSize: pageSize.value } })
}

const handlePageChange = (val) => {
  router.push({ query: { ...route.query, page: val, pageSize: pageSize.value } })
}

const fetchList = async () => {
  loading.value = true
  try {
    if (mode.value === 'rating') {
      const res = await ratingRank({ page: page.value, pageSize: pageSize.value })
      list.value = res.data?.list || []
      total.value = res.data?.total || 0
    } else {
      const res = await userRank({ q: route.query.q || undefined, page: page.value, pageSize: pageSize.value })
      list.value = res.data?.list || []
      total.value = res.data?.total || 0
    }
  } finally {
    loading.value = false
  }
}

watch(
  () => [route.query.page, route.query.pageSize, route.query.q, route.query.tab],
  async () => {
    q.value = route.query.q || ''
    mode.value = route.query.tab === 'rating' ? 'rating' : 'ac'
    await fetchList()
  },
  { immediate: true }
)
</script>

<style scoped lang="scss">
.rank-page {
  width: 100%;
}

.rank-content--empty {
  min-height: 280px;
}
</style>
