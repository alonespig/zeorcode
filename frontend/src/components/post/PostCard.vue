<template>
  <article
    class="post-card border-b border-gray-200 px-[18px] py-3.5 font-[Arial,'Noto_Sans_SC',sans-serif]">
    <!-- 标题行：标题优先（深色加粗），分类标签靠右。仅标题可点进帖子 -->
    <div class="post-card-title flex items-center justify-between gap-3">
      <span class="cursor-pointer truncate text-[16px] font-medium text-gray-700 hover:text-blue-500"
        @click="clickPost">
        {{ post.title }}
      </span>
      <span class="inline-flex shrink-0 items-center rounded border px-2 py-0.5 text-xs font-medium leading-5"
        :class="categoryStyleMap[post.category] || categoryStyleMap.default">
        {{ categoryTitleMap[post.category] }}
      </span>
    </div>

    <!-- 摘要：后端自动从正文截取，有才显示 -->
    <p v-if="post.summary" class="mt-1.5 truncate text-[13px] text-gray-400">{{ post.summary }}</p>

    <!-- 元信息：小头像+作者 · 时间 · 关联题目 …… 右侧浏览/赞/评论 -->
    <div class="post-card-meta mt-2.5 flex items-center gap-2.5 text-[12.5px] text-gray-400">
      <span class="post-card-author flex items-center gap-1.5 text-gray-500">
        <img v-if="post.user?.avatar" :src="post.user.avatar" class="h-5 w-5 rounded-full object-cover" alt="">
        <UserName :name="post.user?.username || '未知用户'" :rating="post.user?.rating" />
      </span>
      <span class="text-gray-300">·</span>
      <span class="post-card-date">{{ post.createdAt }}</span>

      <template v-if="showProblem && post.problem">
        <span class="text-gray-300">·</span>
        <span class="post-card-problem flex items-center cursor-pointer gap-1 text-blue-400 hover:underline" @click.stop="clickProblem">
          <el-icon><Link /></el-icon>{{ post.problem.name }}
        </span>
      </template>

      <span class="post-card-stats ml-auto flex items-center gap-4 tabular-nums [&_.iconfont]:mr-1 [&_.iconfont]:text-gray-400">
        <span><span class="iconfont icon-yanjing"></span>{{ post.viewCount }}</span>
        <span><span class="iconfont icon-dianzan"></span>{{ post.likeCount }}</span>
        <span><span class="iconfont icon-taolun1"></span>{{ post.commentCount }}</span>
      </span>
    </div>
  </article>
</template>

<script setup>
import { categoryTitleMap, categoryStyleMap } from '@/constants/index'
import UserName from '@/components/UserName.vue'
import { Link } from '@element-plus/icons-vue'
const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
  // 是否显示"关联题目"：全站列表 true；题目页题解/讨论 tab 传 false（已在某题下，重复）
  showProblem: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['click-post', 'click-problem'])

const clickPost = () => {
  emit('click-post', props.post.id)
}

const clickProblem = () => {
  emit('click-problem', props.post.problem.id)
}
</script>

<style scoped lang="scss">
@media (max-width: 560px) {
  .post-card {
    padding: 14px 12px;
  }

  .post-card-title {
    align-items: flex-start;
  }

  .post-card-meta {
    flex-wrap: wrap;
    column-gap: 8px;
    row-gap: 6px;
  }

  .post-card-author,
  .post-card-date {
    flex-shrink: 0;
  }

  .post-card-problem {
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .post-card-stats {
    width: 100%;
    margin-left: 0;
    justify-content: flex-end;
  }
}
</style>
