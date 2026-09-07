<template>
  <div>
    <template v-if="posts.length">
      <PostCard v-for="post in posts" :key="post.id" :post="post" @click-post="goPost" @click-problem="goProblem" />
    </template>
    <el-empty v-else description="还没有发布任何帖子" />

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination layout="prev, pager, next" :current-page="page" :page-size="pageSize" :total="total"
        @current-change="onPage" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PostCard from '@/components/post/PostCard.vue'
import { getPostList } from '@/api/post'

const route = useRoute()
const router = useRouter()

const posts = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 15

const fetchPosts = async () => {
  const res = await getPostList({
    category: 'all',
    userID: Number(route.params.id),
    page: page.value,
    pageSize,
  })
  posts.value = res.data.list || []
  total.value = res.data.total || 0
}

const goPost = (id) => router.push({ name: 'BlogDetail', params: { id } })
const goProblem = (id) => router.push({ name: 'ProblemDetail', params: { id } })

const onPage = (val) => {
  page.value = val
  fetchPosts()
}

onMounted(fetchPosts)
watch(() => route.params.id, () => {
  page.value = 1
  fetchPosts()
})
</script>
