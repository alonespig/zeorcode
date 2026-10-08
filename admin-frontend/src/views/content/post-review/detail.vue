<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowLeft, ChatDotRound, View } from "@element-plus/icons-vue";
import { ElMessageBox } from "element-plus";
import MarkdownPreview from "@/components/MarkdownPreview.vue";
import { message } from "@/utils/message";
import {
  getPostDetail,
  reviewPost,
  type PostDetailResp
} from "@/api/admin/posts";
import { POST_CATEGORY_LABELS, POST_CATEGORY_TAG_TYPES } from "./constants";

defineOptions({ name: "AdminPostReviewDetail" });

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const reviewing = ref(false);
const loadFailed = ref(false);
const detail = ref<PostDetailResp | null>(null);

const postID = computed(() => {
  const value = Array.isArray(route.params.id)
    ? route.params.id[0]
    : route.params.id;
  return Number(value);
});

function listQuery() {
  const query: Record<string, string> = {};
  if (typeof route.query.page === "string") query.page = route.query.page;
  if (typeof route.query.pageSize === "string") {
    query.pageSize = route.query.pageSize;
  }
  return query;
}

function backToList() {
  return router.push({ name: "AdminPostReview", query: listQuery() });
}

async function loadDetail() {
  if (!Number.isInteger(postID.value) || postID.value <= 0) {
    loadFailed.value = true;
    return;
  }
  loading.value = true;
  loadFailed.value = false;
  try {
    const res = await getPostDetail(postID.value);
    detail.value = res.data ?? null;
    loadFailed.value = !detail.value;
  } catch {
    loadFailed.value = true;
  } finally {
    loading.value = false;
  }
}

async function approve() {
  if (!detail.value || reviewing.value) return;
  reviewing.value = true;
  try {
    await reviewPost(detail.value.id, 1);
    message("已通过", { type: "success" });
    await backToList();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    reviewing.value = false;
  }
}

async function reject() {
  if (!detail.value || reviewing.value) return;
  let reason: string;
  try {
    const { value } = await ElMessageBox.prompt(
      "请输入拒绝理由（作者可见）",
      "拒绝帖子",
      {
        confirmButtonText: "拒绝",
        cancelButtonText: "取消",
        type: "warning",
        inputType: "textarea",
        inputPlaceholder: "拒绝理由",
        inputValidator: (input: string) =>
          input && input.trim() ? true : "拒绝理由不能为空"
      }
    );
    reason = (value ?? "").trim();
  } catch {
    return;
  }

  reviewing.value = true;
  try {
    await reviewPost(detail.value.id, 2, reason);
    message("已拒绝", { type: "success" });
    await backToList();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    reviewing.value = false;
  }
}

onMounted(loadDetail);
</script>

<template>
  <div class="post-detail-page">
    <header class="page-header">
      <el-button :icon="ArrowLeft" @click="backToList">返回审核列表</el-button>
      <div v-if="detail?.reviewStatus === 0" class="review-actions">
        <el-button type="success" :loading="reviewing" @click="approve">
          通过
        </el-button>
        <el-button type="danger" :loading="reviewing" @click="reject">
          拒绝
        </el-button>
      </div>
    </header>

    <article v-loading="loading" class="post-panel">
      <template v-if="detail">
        <div class="title-row">
          <h1>{{ detail.title }}</h1>
          <el-tag
            :type="POST_CATEGORY_TAG_TYPES[detail.category] ?? 'info'"
            effect="plain"
          >
            {{ POST_CATEGORY_LABELS[detail.category] ?? detail.category }}
          </el-tag>
          <el-tag
            v-if="detail.reviewStatus !== 0"
            :type="detail.reviewStatus === 1 ? 'success' : 'danger'"
            effect="plain"
          >
            {{ detail.reviewStatus === 1 ? "已通过" : "已拒绝" }}
          </el-tag>
        </div>

        <div class="post-meta">
          <el-avatar :size="36" :src="detail.user?.avatar">
            {{ detail.user?.username?.slice(0, 1).toUpperCase() }}
          </el-avatar>
          <div class="author">
            <strong>{{ detail.user?.username ?? "—" }}</strong>
            <span>{{ detail.createdAt || "—" }}</span>
          </div>
          <div class="post-stats">
            <span
              ><el-icon><View /></el-icon>{{ detail.viewCount }}</span
            >
            <span
              ><el-icon><ChatDotRound /></el-icon
              >{{ detail.commentCount }}</span
            >
            <span>点赞 {{ detail.likeCount }}</span>
          </div>
        </div>

        <div v-if="detail.problem" class="related-problem">
          <span>关联题目</span>
          <strong>{{ detail.problem.id }}</strong>
          <span>{{ detail.problem.name }}</span>
        </div>

        <div v-if="detail.summary" class="post-summary">
          {{ detail.summary }}
        </div>

        <div class="content-heading">正文</div>
        <MarkdownPreview :content="detail.content" />
      </template>

      <el-empty
        v-else-if="!loading && loadFailed"
        description="帖子详情加载失败"
      >
        <el-button type="primary" @click="loadDetail">重新加载</el-button>
      </el-empty>
    </article>
  </div>
</template>

<style scoped>
.post-detail-page {
  min-height: 520px;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.review-actions {
  display: flex;
  align-items: center;
}
.post-panel {
  min-height: 420px;
  padding: 26px 32px 36px;
  border: 1px solid var(--el-border-color-lighter);
  background: #fff;
}
.title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}
.title-row h1 {
  min-width: 0;
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 24px;
  font-weight: 600;
  line-height: 1.4;
}
.post-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.author strong,
.author span {
  display: block;
}
.author strong {
  color: var(--el-text-color-primary);
  font-size: 13px;
}
.author span {
  margin-top: 3px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.post-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-left: auto;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.post-stats span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.related-problem {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 18px;
  color: var(--el-text-color-regular);
  font-size: 13px;
}
.related-problem > span:first-child {
  color: var(--el-text-color-secondary);
}
.related-problem strong {
  color: var(--el-color-primary);
  font-family: Consolas, Monaco, monospace;
}
.post-summary {
  margin-top: 18px;
  padding: 12px 14px;
  border-left: 3px solid var(--el-color-primary-light-5);
  background: var(--el-fill-color-extra-light);
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.7;
}
.content-heading {
  margin: 26px 0 16px;
  color: var(--el-text-color-primary);
  font-size: 15px;
  font-weight: 600;
}
@media (max-width: 720px) {
  .post-panel {
    padding: 20px 18px 28px;
  }
  .post-meta {
    flex-wrap: wrap;
  }
  .post-stats {
    width: 100%;
    margin-left: 46px;
  }
}
</style>
