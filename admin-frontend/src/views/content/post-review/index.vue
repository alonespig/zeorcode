<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessageBox } from "element-plus";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import { getPendingPosts, reviewPost, type PostItem } from "@/api/admin/posts";
import { POST_CATEGORY_LABELS, POST_CATEGORY_TAG_TYPES } from "./constants";

defineOptions({ name: "AdminPostReview" });

/** 表格行数据，附加按行的提交中状态，用于防重复点击 */
type PostRow = PostItem & { _busy?: boolean };

const route = useRoute();
const router = useRouter();
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

function openDetail(row: PostRow) {
  router.push({
    name: "AdminPostReviewDetail",
    params: { id: String(row.id) },
    query: {
      page: String(pagination.currentPage),
      pageSize: String(pagination.pageSize)
    }
  });
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
  const page = Number(route.query.page);
  const pageSize = Number(route.query.pageSize);
  if (Number.isInteger(page) && page > 0) pagination.currentPage = page;
  if (pagination.pageSizes?.includes(pageSize)) pagination.pageSize = pageSize;
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
          table-layout="fixed"
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
              :type="POST_CATEGORY_TAG_TYPES[row.category] ?? 'info'"
              size="small"
            >
              {{ POST_CATEGORY_LABELS[row.category] ?? row.category }}
            </el-tag>
          </template>
          <template #title="{ row }">
            <div class="min-w-0 overflow-hidden text-left">
              <el-link
                type="primary"
                :underline="false"
                class="max-w-full"
                :title="row.title"
                @click="openDetail(row)"
              >
                <span class="block truncate">{{ row.title }}</span>
              </el-link>
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
            <div class="flex items-center justify-center whitespace-nowrap">
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
            </div>
          </template>
        </PureTable>
      </template>
    </PureTableBar>
  </div>
</template>
