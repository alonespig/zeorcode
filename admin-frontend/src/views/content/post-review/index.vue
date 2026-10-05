<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessageBox } from "element-plus";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  getPendingPosts,
  getPostDetail,
  reviewPost,
  type PostDetailResp,
  type PostItem
} from "@/api/admin/posts";

defineOptions({ name: "AdminPostReview" });

/** 表格行数据，附加按行的提交中状态，用于防重复点击 */
type PostRow = PostItem & { _busy?: boolean };

/** 分类中文名与标签样式（对齐旧前端 categoryTitleMap / categoryStyleMap） */
const CATEGORY_LABELS: Record<string, string> = {
  blog: "博客",
  announcement: "通知",
  solution: "题解",
  help: "讨论"
};
const CATEGORY_TAG_TYPES: Record<
  string,
  "primary" | "success" | "info" | "warning" | "danger"
> = {
  blog: "primary",
  announcement: "danger",
  solution: "warning",
  help: "info"
};

const dataList = ref<PostRow[]>([]);
const loading = ref(false);

const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

const columns: TableColumnList = [
  { label: "编号", prop: "id", width: 80, align: "center" },
  { label: "分类", slot: "category", width: 90, align: "center" },
  { label: "标题", slot: "title", minWidth: 260 },
  { label: "作者", slot: "author", width: 140, align: "center" },
  { label: "关联题目", slot: "problem", minWidth: 150 },
  { label: "提交时间", prop: "createdAt", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 190,
    fixed: "right",
    align: "center"
  }
];

// 详情弹窗
const dialogVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<PostDetailResp | null>(null);
// 请求序号：快速切换/关闭后使过期请求失效，避免旧结果覆盖当前详情
const detailSeq = ref(0);

async function onSearch() {
  loading.value = true;
  try {
    const res = await getPendingPosts({
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    });
    dataList.value = res.data?.list ?? [];
    pagination.total = res.data?.total ?? 0;
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function onCurrentChange(page: number) {
  pagination.currentPage = page;
  onSearch();
}

function onSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.currentPage = 1;
  onSearch();
}

async function openDetail(row: PostRow) {
  const seq = ++detailSeq.value;
  dialogVisible.value = true;
  detail.value = null;
  detailLoading.value = true;
  try {
    const res = await getPostDetail(row.id);
    if (seq !== detailSeq.value) return;
    detail.value = res.data ?? null;
  } catch {
    if (seq !== detailSeq.value) return;
    // 加载失败：错误已统一提示，关闭弹窗
    dialogVisible.value = false;
  } finally {
    if (seq === detailSeq.value) detailLoading.value = false;
  }
}

function onDialogClosed() {
  // 使进行中的请求失效，并复位 loading
  detailSeq.value += 1;
  detail.value = null;
  detailLoading.value = false;
}

/** 操作成功后从审核队列移除；当前页最后一条回退页码并等待刷新完成 */
async function removeFromQueue() {
  if (dataList.value.length === 1 && pagination.currentPage > 1) {
    pagination.currentPage -= 1;
  }
  await onSearch();
}

async function approve(row: PostRow) {
  if (row._busy) return;
  row._busy = true;
  try {
    await reviewPost(row.id, 1);
    message("已通过", { type: "success" });
    await removeFromQueue();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

async function reject(row: PostRow) {
  if (row._busy) return;
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
        inputValidator: (v: string) =>
          v && v.trim() ? true : "拒绝理由不能为空"
      }
    );
    reason = (value ?? "").trim();
  } catch {
    return; // 用户取消
  }
  row._busy = true;
  try {
    await reviewPost(row.id, 2, reason);
    message("已拒绝", { type: "success" });
    await removeFromQueue();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

onMounted(() => {
  onSearch();
});
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="onSearch">
      <template #title>
        <div class="flex items-center">
          <span>帖子审核</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            待审核 {{ pagination.total }} 条
          </span>
        </div>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <PureTable
          border
          adaptive
          table-layout="auto"
          align-whole="center"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="pagination"
          @page-size-change="onSizeChange"
          @page-current-change="onCurrentChange"
        >
          <template #category="{ row }">
            <el-tag
              :type="CATEGORY_TAG_TYPES[row.category] ?? 'info'"
              size="small"
            >
              {{ CATEGORY_LABELS[row.category] ?? row.category }}
            </el-tag>
          </template>
          <template #title="{ row }">
            <div class="flex flex-col items-start">
              <el-link
                type="primary"
                :underline="false"
                @click="openDetail(row)"
              >
                {{ row.title }}
              </el-link>
              <span
                class="mt-0.5 w-full truncate text-xs text-gray-400"
                :title="row.summary"
              >
                {{ row.summary || "—" }}
              </span>
            </div>
          </template>
          <template #author="{ row }">
            {{ row.user?.username ?? "—" }}
          </template>
          <template #problem="{ row }">
            <template v-if="row.problem">
              <span class="mr-1 font-mono text-xs">{{ row.problem.id }}</span>
              <span class="text-xs">{{ row.problem.name }}</span>
            </template>
            <span v-else class="text-gray-400">—</span>
          </template>
          <template #operation="{ row }">
            <el-button
              link
              size="small"
              type="primary"
              @click="openDetail(row)"
            >
              查看
            </el-button>
            <el-button
              link
              size="small"
              type="success"
              :loading="row._busy"
              @click="approve(row)"
            >
              通过
            </el-button>
            <el-button
              link
              size="small"
              type="danger"
              :loading="row._busy"
              @click="reject(row)"
            >
              拒绝
            </el-button>
          </template>
        </PureTable>
      </template>
    </PureTableBar>

    <el-dialog
      v-model="dialogVisible"
      title="帖子详情"
      width="min(760px, calc(100vw - 32px))"
      top="6vh"
      @closed="onDialogClosed"
    >
      <div v-loading="detailLoading" class="min-h-40">
        <template v-if="detail">
          <div class="mb-3 flex items-center gap-2">
            <span class="text-base font-medium">{{ detail.title }}</span>
            <el-tag
              :type="CATEGORY_TAG_TYPES[detail.category] ?? 'info'"
              size="small"
            >
              {{ CATEGORY_LABELS[detail.category] ?? detail.category }}
            </el-tag>
          </div>
          <el-descriptions :column="2" border size="small" class="mb-4">
            <el-descriptions-item label="作者">
              {{ detail.user?.username ?? "—" }}
            </el-descriptions-item>
            <el-descriptions-item label="提交时间">
              {{ detail.createdAt || "—" }}
            </el-descriptions-item>
            <el-descriptions-item label="关联题目">
              <template v-if="detail.problem">
                {{ detail.problem.id }} {{ detail.problem.name }}
              </template>
              <template v-else>—</template>
            </el-descriptions-item>
            <el-descriptions-item label="数据">
              浏览 {{ detail.viewCount }} · 点赞 {{ detail.likeCount }} · 评论
              {{ detail.commentCount }}
            </el-descriptions-item>
          </el-descriptions>
          <div v-if="detail.summary" class="mb-3 text-sm text-gray-500">
            摘要：{{ detail.summary }}
          </div>
          <div class="mb-1 text-sm font-medium">正文</div>
          <div
            class="max-h-96 overflow-auto whitespace-pre-wrap break-words rounded border border-gray-200 bg-gray-50 p-3 text-sm leading-6"
          >
            {{ detail.content }}
          </div>
        </template>
        <el-empty
          v-else-if="!detailLoading"
          description="无数据"
          :image-size="60"
        />
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>
