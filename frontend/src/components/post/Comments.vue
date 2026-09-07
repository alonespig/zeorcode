<template>
  <div class="comments">
    <div class="c-head">评论<span class="c-total">{{ total }}</span></div>

    <!-- 顶层发布框 -->
    <div v-if="userStore.isLogin" class="compose">
      <el-avatar :size="42" :src="userStore.user?.avatar" class="av" />
      <div class="compose-body">
        <el-input v-model="newComment" type="textarea" :rows="3" maxlength="500"
          resize="none" show-word-limit placeholder="写下你的想法…" />
        <div class="compose-foot">
          <EmojiPicker @select="insertCommentEmoji" />
          <el-button class="publish-button" type="primary" :icon="Promotion" :loading="posting"
            :disabled="!newComment.trim()" @click="submit">发布评论</el-button>
        </div>
      </div>
    </div>
    <div v-else class="c-login" @click="userStore.promptLogin()">登录后参与评论</div>

    <!-- 一级评论列表 -->
    <div v-loading="loading" class="c-list">
      <div v-for="c in comments" :key="c.id" class="c-item">
        <el-avatar :size="40" :src="c.user?.avatar" class="av" />

        <div class="c-main">
          <div class="c-name"><UserName :name="c.user?.username || '未知用户'" :rating="c.user?.rating" /></div>
          <div class="c-content">{{ c.content }}</div>
          <div class="c-meta">
            <span>{{ fmt(c.createdAt) }}</span>
            <span class="op like" :class="{ liked: c.isLiked }" @click="toggleLike(c)"><span
                class="iconfont icon-dianzan"></span>{{ c.likeCount || '' }}</span>
            <span v-if="c.id !== replyId" class="op" @click="reply(c.id)">回复</span>
            <span v-else class="op active" @click="cancelReply()">取消回复</span>
            <span v-if="canDelete(c)" class="op del" @click="removeComment(c.id)">删除</span>
          </div>

          <!-- 回复框（回复楼主） -->
          <div v-if="c.id === replyId" class="reply-box">
            <EmojiPicker @select="insertReplyEmoji" />
            <el-input v-model="replyText" :placeholder="`回复 ${c.user?.username}`" maxlength="500"
              @keyup.enter="sendReply(c.id)" />
            <el-button class="reply-button" type="primary" :icon="Promotion" :loading="posting"
              :disabled="!replyText.trim()" @click="sendReply(c.id)">发送</el-button>
          </div>

          <!-- 楼内回复 -->
          <div v-if="c.replies?.length" class="replies">
            <div v-for="r in c.replies" :key="r.id" class="r-item">
              <el-avatar :size="26" :src="r.user?.avatar" class="av" />
              <div class="r-main">
                <div class="r-line">
                  <span class="r-name"><UserName :name="r.user?.username" :rating="r.user?.rating" /></span>
                  <template v-if="r.replyTo">
                    <span class="r-at"> 回复 @{{ r.replyTo.username }}</span>
                  </template>
                  <span class="r-colon">：</span>{{ r.content }}
                </div>
                <div class="c-meta">
                  <span>{{ fmt(r.createdAt) }}</span>
                  <span class="op like" :class="{ liked: r.isLiked }" @click="toggleLike(r)"><span
                      class="iconfont icon-dianzan"></span>{{ r.likeCount || '' }}</span>
                  <span v-if="r.id !== replyId" class="op" @click="reply(r.id)">回复</span>
                  <span v-else class="op active" @click="cancelReply()">取消回复</span>
                  <span v-if="canDelete(r)" class="op del" @click="removeComment(r.id)">删除</span>
                </div>
                <div v-if="r.id === replyId" class="reply-box">
                  <EmojiPicker @select="insertReplyEmoji" />
                  <el-input v-model="replyText" :placeholder="`回复 ${r.user?.username}`" maxlength="500"
                    @keyup.enter="sendReply(r.id)" />
                  <el-button class="reply-button" type="primary" :icon="Promotion" :loading="posting"
                    :disabled="!replyText.trim()" @click="sendReply(r.id)">发送</el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <el-empty v-if="!loading && !comments.length" description="暂无评论" :image-size="60" />
    </div>
  </div>
</template>

<script setup>
import UserName from '@/components/UserName.vue'
import EmojiPicker from '@/components/post/EmojiPicker.vue'
import { ref, onMounted } from 'vue'
import { ElMessageBox } from 'element-plus'
import { Promotion } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getPostComments, createPostComment, deletePostComment, toggleCommentLike } from '@/api/post'

const props = defineProps({
  postId: { type: [String, Number], required: true },
})
const emit = defineEmits(['changed'])

const userStore = useUserStore()

const comments = ref([])
const total = ref(0)
const newComment = ref('')
const replyId = ref(0)
const replyText = ref('')
const posting = ref(false)
const loading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const res = await getPostComments(props.postId)
    comments.value = res.data?.list || []
    total.value = res.data?.total ?? comments.value.length
  } finally {
    loading.value = false
  }
}

const submit = async () => {
  if (!userStore.isLogin) {
    userStore.promptLogin()
    return
  }
  if (!newComment.value.trim()) return
  posting.value = true
  try {
    await createPostComment(props.postId, { content: newComment.value.trim(), parentID: 0 })
    newComment.value = ''
    await load()
    emit('changed')
  } finally {
    posting.value = false
  }
}

const reply = (id) => {
  if (!userStore.isLogin) {
    userStore.promptLogin()
    return
  }
  replyId.value = id
  replyText.value = ''
}
const cancelReply = () => {
  replyId.value = 0
  replyText.value = ''
}

const appendEmoji = (content, emoji) => `${content}${emoji}`.slice(0, 500)
const insertCommentEmoji = (emoji) => {
  newComment.value = appendEmoji(newComment.value, emoji)
}
const insertReplyEmoji = (emoji) => {
  replyText.value = appendEmoji(replyText.value, emoji)
}

const toggleLike = async (c) => {
  if (!userStore.isLogin) {
    userStore.promptLogin()
    return
  }
  try {
    const res = await toggleCommentLike(c.id)
    c.isLiked = res.data.liked
    c.likeCount = Math.max(0, (c.likeCount || 0) + (res.data.liked ? 1 : -1))
  } catch (err) {
    console.error(err)
  }
}

// parentId = 被回复评论的 id；后端据此挂到对应顶层评论并算出 @谁
const sendReply = async (parentId) => {
  if (!replyText.value.trim()) return
  posting.value = true
  try {
    await createPostComment(props.postId, { content: replyText.value.trim(), parentID: parentId })
    cancelReply()
    await load()
    emit('changed')
  } finally {
    posting.value = false
  }
}

const canDelete = (c) =>
  userStore.isAdmin || (userStore.user && c.user?.id === userStore.user.id)

const removeComment = (id) => {
  ElMessageBox.confirm('确定删除这条评论？', '提示', { type: 'warning' })
    .then(async () => {
      await deletePostComment(id)
      await load()
      emit('changed')
    })
    .catch(() => {})
}

const fmt = (ts) => {
  if (!ts) return ''
  const d = new Date(Number(ts) * 1000)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

onMounted(load)
</script>

<style scoped lang="scss">
.c-head {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  border-bottom: 1px solid #edf0f5;
  padding-bottom: 12px;
  margin-bottom: 16px;

  .c-total {
    font-size: 13px;
    color: #7a8494;
    font-weight: 400;
    margin-left: 8px;
  }
}

.av {
  flex-shrink: 0;
}

/* 顶层发布框 */
.compose {
  display: flex;
  gap: 14px;
  margin-bottom: 6px;

  .compose-body {
    flex: 1;
  }

  .compose-foot {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 10px;
  }
}

.publish-button {
  min-width: 108px;
  height: 36px;
}

.c-login {
  font-size: 14px;
  color: #3867a8;
  cursor: pointer;
  margin-bottom: 8px;
}

.c-list {
  min-height: 80px;
}

/* 一级评论 */
.c-item {
  display: flex;
  gap: 14px;
  padding-top: 20px;
  margin-top: 16px;
  border-top: 1px solid #edf0f5;

  .c-main {
    flex: 1;
    min-width: 0;
  }

  .c-name {
    font-size: 14px;
    color: #3867a8;
    font-weight: 500;
  }

  .c-content {
    font-size: 15px;
    color: #374151;
    line-height: 1.75;
    margin: 6px 0;
    white-space: pre-wrap;
    word-break: break-word;
  }
}

.c-meta {
  font-size: 13px;
  color: #7a8494;
  display: flex;
  gap: 12px;
  align-items: center;

  .op {
    cursor: pointer;

    &:hover,
    &.active {
      color: #3867a8;
    }

    &.del:hover {
      color: #f56c6c;
    }

    &.like {
      display: inline-flex;
      align-items: center;
      gap: 2px;

      .iconfont {
        font-size: 13px;
      }

      &.liked {
        color: #e24b4a;
      }
    }
  }
}

/* 回复输入框 */
.reply-box {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  margin-top: 12px;
}

.reply-button {
  min-width: 84px;
  height: 36px;
}

/* 楼内回复块 */
.replies {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #f7f8fa;
  border-radius: var(--radius-base);
  padding: 14px 16px;
  margin-top: 12px;
}

.r-item {
  display: flex;
  gap: 8px;

  .r-main {
    flex: 1;
    min-width: 0;
  }

  .r-line {
    font-size: 14px;
    line-height: 1.7;
    color: #374151;
    word-break: break-word;

    .r-name {
      color: #6b7280;
    }

    .r-at {
      color: #e6a23c;
    }

    .r-colon {
      color: #6b7280;
    }
  }
}

@media (max-width: 640px) {
  .reply-box {
    flex-wrap: wrap;

    :deep(.el-input) {
      order: -1;
      width: 100%;
    }
  }
}
</style>
