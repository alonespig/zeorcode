<template>
  <div class="f-panel">
    <div class="flex items-center justify-between border-b border-gray-100 px-4 py-3">
      <span class="text-base font-medium text-gray-700">讨论</span>
      <el-button type="primary" size="small" @click="goCreate">发起讨论</el-button>
    </div>

    <template v-if="posts.length">
      <PostCard v-for="post in posts" :key="post.id" :post="post" :show-problem="false" @click-post="goPost" />
    </template>
    <el-empty v-else description="还没有讨论，来发起第一个吧" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PostCard from '@/components/post/PostCard.vue'
import { getPostList } from '@/api/post'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const problemID = route.params.id
const posts = ref([])

const fetchPosts = async () => {
  const res = await getPostList({ category: 'help', problemID, page: 1, pageSize: 20 })
  posts.value = res.data.list || []
}

const goPost = (id) => router.push({ name: 'BlogDetail', params: { id } })

const goCreate = () => {
  if (!userStore.isLogin) {
    userStore.promptLogin()
    return
  }
  router.push({ name: 'BlogCreate', query: { category: 'help', pid: problemID } })
}

onMounted(fetchPosts)
</script>
