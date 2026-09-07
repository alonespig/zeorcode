<template>
  <div class="page-container py-5">
    <div class="overflow-hidden rounded border border-gray-200 bg-white">
      <div class="flex items-center justify-between border-b border-gray-100 px-[22px] py-4">
        <h1 class="m-0 text-lg font-semibold text-gray-800">通知</h1>
        <el-button :icon="CircleCheck" :disabled="!unread.total" @click="markAllRead">
          全部标记为已读
        </el-button>
      </div>

      <!-- tab -->
      <div class="flex gap-7 border-b border-gray-100 px-[22px]">
        <button v-for="t in tabs" :key="t.key"
          class="flex items-center gap-1.5 border-b-2 py-3.5 text-[15px] transition-colors" :class="active === t.key
            ? 'border-blue-500 font-medium text-blue-500'
            : 'border-transparent text-gray-600 hover:text-blue-500'" @click="switchTab(t.key)">
          {{ t.label }}
          <span v-if="badge(t.key)"
            class="inline-block min-w-4 rounded-full bg-red-500 px-1 text-center text-[11px] leading-4 text-white">{{
              badge(t.key) }}</span>
        </button>
      </div>

      <!-- 列表 -->
      <div v-loading="loading">
        <div v-for="n in list" :key="n.id"
          class="relative flex cursor-pointer items-start gap-3.5 border-b border-gray-100 px-[22px] py-4 transition-colors hover:bg-[#fafbfc]"
          :class="{ 'bg-[#f5f9ff]': !n.isRead }" @click="open(n)">
          <span v-if="!n.isRead" class="absolute left-2 top-5 h-1.5 w-1.5 rounded-full bg-blue-500"></span>
          <img v-if="n.actor?.avatar" :src="n.actor.avatar" class="h-11 w-11 shrink-0 rounded-full object-cover" alt="">
          <div v-else class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-base font-medium text-white"
            :style="{ background: avatarBg(n) }">{{ avatarText(n) }}</div>
          <div class="min-w-0 flex-1 pt-0.5">
            <div class="text-[15px] leading-6 text-gray-700">
              <template v-if="n.actor">
                <UserName :name="n.actor.username" :rating="n.actor.rating" />{{ actionText(n)
                }}<span v-if="n.title" class="font-medium text-gray-800">《{{ n.title }}》</span>
              </template>
              <span v-else class="font-medium text-gray-800">{{ n.title }}</span>
            </div>
            <div v-if="n.content" class="mt-1 truncate text-sm leading-5 text-gray-500">{{ n.content }}</div>
            <div class="mt-1.5 text-[13px] leading-5 text-gray-400">{{ n.createdAt }}</div>
          </div>
        </div>
        <el-empty v-if="!loading && !list.length" description="暂无通知" :image-size="70" />
      </div>

      <div v-if="total > pageSize" class="flex items-center justify-center border-t border-gray-100 px-[22px] py-3">
        <el-pagination layout="prev, pager, next" :current-page="page"
          :page-size="pageSize" :total="total" @current-change="handlePage" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { CircleCheck } from '@element-plus/icons-vue'
import UserName from '@/components/UserName.vue'
import { getNotifications, getUnreadCount, markNotificationRead } from '@/api/notification'

const router = useRouter()

const tabs = [
  { key: '', label: '全部' },
  { key: 'comment,reply', label: '评论回复' },
  { key: 'like', label: '点赞' },
  { key: 'system,rating', label: '系统' },
]
const active = ref('')
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 15
const loading = ref(false)
const unread = ref({ comment: 0, reply: 0, like: 0, system: 0, rating: 0, total: 0 })

const actionText = (n) => {
  if (n.type === 'like') return n.sourceType === 'comment' ? ' 赞了你的评论' : ' 赞了你的帖子'
  return {
    comment: ' 评论了你的帖子',
    reply: ' 回复了你的评论',
  }[n.type] || ''
}

const avatarText = (n) =>
  n.actor ? (n.actor.username?.[0]?.toUpperCase() || '?') : (n.type === 'rating' ? 'R' : '系')
const avatarBg = (n) =>
  n.type === 'like' ? '#e24b4a' : n.type === 'rating' ? '#e6a817' : n.type === 'system' ? '#8b929a' : '#2f7fe0'

const badge = (key) => {
  const u = unread.value
  if (key === '') return u.total
  if (key === 'comment,reply') return u.comment + u.reply
  if (key === 'like') return u.like
  if (key === 'system,rating') return u.system + u.rating
  return 0
}

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getNotifications({ type: active.value || undefined, page: page.value, pageSize })
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const fetchUnread = async () => {
  const res = await getUnreadCount().catch(() => null)
  if (res?.data) unread.value = res.data
}

const switchTab = (key) => {
  active.value = key
  page.value = 1
  fetchList()
}
const handlePage = (p) => {
  page.value = p
  fetchList()
}

const open = async (n) => {
  if (!n.isRead) {
    await markNotificationRead({ id: n.id }).catch(() => { })
    n.isRead = true
    fetchUnread()
  }
  if (n.link) router.push(n.link)
}

const markAllRead = async () => {
  await markNotificationRead({}).catch(() => { })
  list.value.forEach((n) => (n.isRead = true))
  fetchUnread()
}

onMounted(() => {
  fetchList()
  fetchUnread()
})
</script>
