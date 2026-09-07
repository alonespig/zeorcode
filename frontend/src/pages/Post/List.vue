<template>
  <div class="post-page page-container">

    <div class="post-layout flex gap-8 items-start">

      <!-- 左边 -->
      <div class="post-main f-panel flex w-[75%] flex-col">

        <!-- 菜单栏 -->
        <div class="post-list-toolbar flex justify-between items-center border-b border-gray-200 px-6 py-2">

          <!-- 菜单 -->
          <div class="post-category-tabs flex gap-8 font-sans font-medium
          text-gray-800 text-[15px]">
            <template v-for="item in menuList" :key="item.id">
              <span @click="clickMenu(item.id)" class="border-b-2 hover:text-blue-400 cursor-pointer px-1 py-1"
                :class="category === item.id
                ? 'text-blue-400 border-blue-400'
                : 'border-transparent'">
                {{ categoryText[item.id] }}
              </span>
            </template>
          </div>

          <!-- 搜索框 -->
          <div class="post-search flex gap-4">
            <el-input v-model="keyword" class="post-search-input h-8 rounded!"></el-input>
            <el-button @click="handleSearch" class="rounded-sm!" :icon="Search" type="primary"></el-button>
          </div>
        </div>

        <div v-loading="loading" class="post-content" :class="{ 'post-content--empty': !posts.length }">
          <PostCard v-for="post in posts" :key="post.id" :post="post" @click-post="goPost"
            @click-problem="goProblem" />

          <el-empty v-if="!loading && !posts.length" :image-size="80" :description="emptyDescription" />
        </div>


        <div v-if="total > pageSize" class="f-pagination">
          <el-pagination size="large" layout="prev, pager, next" :page-size="pageSize" v-model:current-page="page"
            :total="total" @current-change="handlePageChange" />
        </div>
      </div>

      <!-- 右边 -->
      <div class="post-sidebar flex-1 flex flex-col gap-6">
        <el-button @click="goCreate" class="w-full h-10!" type="primary">
          <span class="iconfont icon-bianji4">
            <span class="ml-2">发布一个讨论</span>
          </span>
        </el-button>


        <div v-if="hotPosts.length" class="hot-posts f-panel flex w-70 flex-col gap-2.5 px-4 py-3">
          <div class="pb-1.5 border-gray-300 border-b">
            <span class="iconfont icon-huohua11 text-red-600 mr-2"></span>
            <span class="text-[16px] text-gray-600 font-medium">热门文章</span>
          </div>

          <ul class="min-w-0 px-2 flex flex-col gap-3">
            <li v-for="post in hotPosts" :key="post.id" class="flex items-center gap-2 min-w-0 flex-1">
              <span class="shrink-0 inline-block w-1.5 h-1.5 rounded-full bg-orange-400"></span>
              <span @click="goPost(post.id)"
                class="block cursor-pointer min-w-0 flex-1 leading-tight truncate text-[15px] text-gray-600 font-normal hover:text-blue-400">{{
                  post.title }}</span>
            </li>
          </ul>
        </div>

      </div>

    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PostCard from '@/components/post/PostCard.vue'
import { getPostList } from '@/api/post'
import { useUserStore } from '@/stores/user'
import { categoryList, categoryTitleMap } from '@/constants/index'
import { Search } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()


const clickMenu = (clickCategory) => {
  router.push({
    query: {
      page: 1,
      pageSize: 10,
      category: clickCategory
    }
  })
}

const goPost = (id) => {
  router.push({
    'name': 'BlogDetail',
    params: {
      id
    }
  })
}

const goProblem = (id) => {
  router.push({
    name: 'ProblemDetail',
    params: { id },
  })
}

const menuList = [
  {
    id: 'all',
    name: '全部'
  },
  ...categoryList
]

const categoryText = {
  all: '全部',
  ...categoryTitleMap
}

const category = ref('all')


const posts = ref([])
const total = ref(0)
const loading = ref(true)

const page = ref(Number(route.query.page) || 1)
const pageSize = 10

const keyword = ref(route.query.keyword || '')


const problemID = computed(() => route.query.problemID || '')

const emptyDescription = computed(() => {
  if (keyword.value.trim()) return '没有找到相关内容'
  if (problemID.value) return '该题目下暂无相关内容'
  if (category.value !== 'all') return '该分类下暂无内容'
  return '暂无社区内容'
})

const hotPosts = ref([])

const fetchHotPosts = async () => {
  const res = await getPostList({
    category: 'all',
    sort: 'hot',
    page: 1,
    pageSize: 10,
  })
  hotPosts.value = res.data.list || []
}


const fetchPosts = async () => {
  loading.value = true
  try {
    const res = await getPostList({
      category: category.value,
      problemID: problemID.value || undefined,
      keyword: keyword.value,
      page: page.value,
      pageSize,
    })
    posts.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  router.push({
    query: {
      ...route.query,
      keyword: keyword.value || undefined,
      page: 1,
    },
  })
}

const handlePageChange = () => {
  router.push({
    query: {
      ...route.query,
      page: page.value,
    },
  })
}

const goCreate = () => {
  if (!userStore.isLogin) {
    userStore.promptLogin()
    return
  }
  router.push({
    name: 'BlogCreate',
  })
}

watch(
  () => route.fullPath,
  () => {
    page.value = Number(route.query.page) || 1
    keyword.value = route.query.keyword || ''
    category.value = route.query.category || 'all'
    fetchPosts()
  },
)

onMounted(async () => {
  await fetchPosts()
  await fetchHotPosts()
})
</script>

<style scoped lang="scss">
.post-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.post-search-input {
  width: 240px;
}

.post-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px;
}

.toolbar-title {
  display: flex;
  align-items: baseline;
  gap: 10px;

  h2 {
    margin: 0;
    color: #1f2937;
    font-size: 22px;
  }

  span {
    color: #7a8494;
    font-size: 14px;
  }
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  position: relative;

  .iconfont {
    position: absolute;
    top: 50%;
    left: 12px;
    transform: translateY(-50%);
    color: #98a2b3;
    font-size: 16px;
  }

  input {
    width: 240px;
    height: 36px;
    border: 1px solid #d9dee8;
    border-radius: 18px;
    padding: 0 14px 0 38px;
    outline: none;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;

    &:focus {
      border-color: #2563eb;
      box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
    }
  }
}

.post-list {
  min-height: 260px;
  background: #fff;
  border: 1px solid #edf0f5;
}

.post-content--empty {
  min-height: 280px;
}

.el-button {
  .iconfont {
    margin-right: 5px;
    font-size: 14px;
  }
}

@media (max-width: 1023px) {
  .post-layout {
    flex-direction: column;
    gap: 18px;
  }

  .post-main,
  .post-sidebar,
  .hot-posts {
    width: 100%;
  }

  .post-sidebar {
    gap: 18px;
  }
}

@media (max-width: 767px) {
  .post-layout,
  .post-sidebar {
    gap: 12px;
  }

  .post-list-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    padding: 10px 12px;
  }

  .post-category-tabs {
    width: 100%;
    gap: 24px;
    overflow-x: auto;
    white-space: nowrap;
    scrollbar-width: none;

    &::-webkit-scrollbar {
      display: none;
    }
  }

  .post-search {
    width: 100%;
    gap: 8px;
  }

  .post-search-input {
    width: auto;
    min-width: 0;
    flex: 1;
  }
}
</style>
