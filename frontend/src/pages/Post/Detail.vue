<template>
  <div class="mx-auto max-w-[1340px] px-5 pb-5">
    <div class="flex items-start gap-[18px]">
      <!-- 主区 -->
      <div class="flex min-w-0 flex-1 flex-col gap-[18px]">
        <article v-loading="loading" class="f-panel px-7 py-6">
          <template v-if="post">
            <!-- 审核状态提示（仅作者/管理员能进到未过审的帖子） -->
            <el-alert v-if="post.reviewStatus === 0" type="warning" :closable="false" show-icon class="mb-3"
              title="该帖子正在等待管理员审核，通过后才会对其他人可见。" />
            <el-alert v-else-if="post.reviewStatus === 2" type="error" :closable="false" show-icon class="mb-3"
              :title="`该帖子未通过审核${post.rejectReason ? '：' + post.rejectReason : ''}。修改后将重新提交审核。`" />

            <!-- 标题 + 分类 -->
            <div class="flex flex-wrap items-center gap-2.5">
              <h1 class="min-w-0 text-[25px] font-medium leading-snug text-gray-800 break-words">{{ post.title }}</h1>
              <span class="shrink-0 rounded border px-2 py-0.5 text-xs font-medium"
                :class="categoryStyleMap[post.category] || categoryStyleMap.default">
                {{ categoryTitleMap[post.category] || '文章' }}
              </span>
            </div>

            <!-- 作者行 -->
            <div class="mt-4 flex items-center gap-3">
              <el-avatar :size="38" :src="post.user?.avatar">
                {{ post.user?.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <div class="min-w-0">
                <div class="text-sm"><UserName :name="post.user?.username || '未知用户'" :rating="post.user?.rating" /></div>
                <div class="text-[12.5px] text-gray-400">发布于 {{ post.createdAt }}</div>
              </div>
              <div class="ml-auto flex items-center gap-4 text-[13px] text-gray-500">
                <span class="inline-flex items-center gap-1"><span class="iconfont icon-yanjing"></span>{{ post.viewCount
                  }}</span>
                <span class="inline-flex items-center gap-1"><span class="iconfont icon-taolun1"></span>{{
                  post.commentCount }}</span>
                <template v-if="canManagePost">
                  <el-button size="small" @click="goEdit">
                    <span class="iconfont icon-bianji1 mr-1"></span>编辑
                  </el-button>
                  <el-button size="small" type="danger" plain @click="handleDeletePost">删除</el-button>
                </template>
              </div>
            </div>

            <!-- 关联题目 + 分隔线 -->
            <div class="mt-3.5 flex flex-wrap items-center gap-2.5 border-b border-gray-100 pb-4">
              <span v-if="post.problem" class="inline-flex items-center gap-1.5 text-[13px] text-gray-500">
                <el-icon>
                  <Link />
                </el-icon>关联题目
                <RouterLink class="text-blue-500 hover:underline" :to="`/problem/${post.problem.id}`">
                  {{ post.problem.id }}. {{ post.problem.name }}
                </RouterLink>
              </span>
            </div>

            <!-- 正文 -->
            <div class="pt-5">
              <MdEditor :model-value="post.content" preview-only :editor-id="mdId" />
            </div>

            <!-- 点赞 -->
            <div class="mt-7 flex justify-center">
              <el-button :type="post.isLiked ? 'primary' : 'default'" @click="handleLike">
                <span class="iconfont icon-dianzan mr-1.5"></span>点赞 {{ post.likeCount }}
              </el-button>
            </div>
          </template>
        </article>

        <!-- 评论 -->
        <section ref="commentsSection" class="f-panel scroll-mt-4 px-6 py-5">
          <Comments :post-id="postID" @changed="fetchPost" />
        </section>
      </div>

      <!-- 目录与文章快捷操作 -->
      <aside class="post-sidebar sticky top-4 hidden shrink-0 flex-col gap-3 lg:flex">
        <div class="f-panel overflow-hidden">
          <div class="px-4 pb-2 pt-3.5 text-[15px] font-medium text-gray-800">目录</div>
          <div class="catalog-scroll overflow-auto px-4 pb-3.5">
            <MdCatalog :editor-id="mdId" :scroll-element="scrollElement" :md-heading-id="mdHeadingId" />
          </div>
        </div>

        <nav class="flex flex-col items-center gap-[18px]" aria-label="文章快捷导航">
          <el-tooltip content="返回顶部" placement="left">
            <button type="button" class="post-side-action" aria-label="返回顶部" @click="scrollToTop">
              <el-icon><Top /></el-icon>
            </button>
          </el-tooltip>
          <el-tooltip content="查看评论" placement="left">
            <button type="button" class="post-side-action" aria-label="查看评论" @click="scrollToComments">
              <el-icon><ChatDotRound /></el-icon>
            </button>
          </el-tooltip>
        </nav>
      </aside>
    </div>
  </div>
</template>

<script setup>
import UserName from '@/components/UserName.vue'
import { computed, onMounted, ref, useTemplateRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Link, Top } from '@element-plus/icons-vue'
import { MdCatalog } from 'md-editor-v3'
import MdEditor from '@/components/MdEditor.vue'
import Comments from '@/components/post/Comments.vue'
import { mdHeadingId } from '@/utils/md'
import { deletePost, getPost, togglePostLike } from '@/api/post'
import { useUserStore } from '@/stores/user'
import { categoryTitleMap, categoryStyleMap } from '@/constants'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const post = ref(null)
const loading = ref(false)
const commentsSection = useTemplateRef('commentsSection')

// 目录：editorId 要和正文 MdEditor 预览一致
const mdId = 'post-preview'
const scrollElement = document.documentElement

const postID = computed(() => Number(route.params.id))
const canManagePost = computed(() => {
  const current = userStore.user
  return Boolean(current && post.value && (current.id === post.value.user?.id || userStore.isAdmin))
})

const fetchPost = async () => {
  loading.value = true
  try {
    const res = await getPost(postID.value)
    post.value = res.data
  } finally {
    loading.value = false
  }
}

const handleLike = async () => {
  if (!userStore.isLogin) {
    userStore.promptLogin()
    return
  }
  const res = await togglePostLike(postID.value)
  post.value.isLiked = res.data.liked
  post.value.likeCount += res.data.liked ? 1 : -1
}

const goEdit = () => {
  router.push({ name: 'BlogEdit', params: { id: post.value.id } })
}

const scrollToComments = () => {
  commentsSection.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const handleDeletePost = async () => {
  await ElMessageBox.confirm('确认删除这篇内容吗？', '删除确认', { type: 'warning' })
  await deletePost(postID.value)
  ElMessage.success('已删除')
  router.push({ name: 'BlogList' })
}

const formatTime = (value) => {
  if (!value) return ''
  return new Date(Number(value) * 1000).toLocaleString()
}

onMounted(fetchPost)
</script>

<style scoped>
.post-sidebar {
  width: 250px;
}

.catalog-scroll {
  max-height: min(500px, calc(100vh - 210px));
}

.post-side-action {
  display: inline-flex;
  width: 48px;
  height: 48px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border-soft);
  border-radius: 50%;
  background: var(--color-surface);
  color: var(--color-text);
  box-shadow: var(--shadow-soft);
  font-size: 20px;
  transition: color 0.2s ease, border-color 0.2s ease, background-color 0.2s ease;
}

.post-side-action:hover {
  border-color: var(--el-color-primary-light-5);
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
</style>
