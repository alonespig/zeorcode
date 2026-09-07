<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="t">帖子审核</span>
          <span class="sub">待审核 {{ total }} 篇</span>
        </div>
        <el-button :icon="Refresh" @click="loadPosts">刷新</el-button>
      </div>
    </template>

    <el-table v-loading="loading" :data="postList" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="分类" width="90">
        <template #default="{ row }">
          <el-tag size="small" effect="light">{{ categoryTitleMap[row.category] || row.category }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="标题" min-width="260">
        <template #default="{ row }">
          <a class="title-link" @click="openPost(row.id)">{{ row.title }}</a>
          <div class="summary">{{ row.summary }}</div>
        </template>
      </el-table-column>
      <el-table-column label="作者" width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ row.user?.username || '—' }}</template>
      </el-table-column>
      <el-table-column label="提交时间" width="170">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="success" size="small" link :loading="row._busy" @click="approve(row)">通过</el-button>
          <el-button type="danger" size="small" link :loading="row._busy" @click="reject(row)">拒绝</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="没有待审核的帖子" :image-size="80" />
      </template>
    </el-table>

    <div class="pager">
      <el-pagination background layout="total, prev, pager, next" :current-page="page" :page-size="pageSize"
        :total="total" @current-change="handlePageChange" />
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Refresh } from "@element-plus/icons-vue";
import { useRouter } from "vue-router";
import { getPendingPosts, reviewPost } from "@/api/post";
import { categoryTitleMap } from "@/constants/index";

const router = useRouter();
const postList = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);

const formatTime = (ts) => {
  if (!ts) return "-";
  const d = new Date(ts * 1000);
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const loadPosts = async () => {
  loading.value = true;
  try {
    const res = await getPendingPosts({ page: page.value, pageSize: pageSize.value });
    postList.value = res.data?.list || [];
    total.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (val) => {
  page.value = val;
  loadPosts();
};

// 管理员可查看未过审帖子（后端放行），新标签页打开预览
const openPost = (id) => {
  const url = router.resolve({ name: "BlogDetail", params: { id } }).href;
  window.open(url, "_blank");
};

const approve = async (row) => {
  row._busy = true;
  try {
    await reviewPost(row.id, 1);
    ElMessage.success("已通过");
    removeRow(row.id);
  } catch (err) {
    console.error(err);
  } finally {
    row._busy = false;
  }
};

const reject = async (row) => {
  let reason = "";
  try {
    const r = await ElMessageBox.prompt("请输入拒绝理由（作者可见）", "拒绝帖子", {
      confirmButtonText: "确认拒绝",
      cancelButtonText: "取消",
      inputPlaceholder: "如：内容与题目无关 / 含广告 …",
    });
    reason = (r.value || "").trim();
  } catch {
    return;
  }
  row._busy = true;
  try {
    await reviewPost(row.id, 2, reason);
    ElMessage.success("已拒绝");
    removeRow(row.id);
  } catch (err) {
    console.error(err);
  } finally {
    row._busy = false;
  }
};

// 审核完从待审列表移除，并修正计数
const removeRow = (id) => {
  postList.value = postList.value.filter((p) => p.id !== id);
  total.value = Math.max(0, total.value - 1);
};

onMounted(loadPosts);
</script>

<style scoped>
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.head-title .t {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.head-title .sub {
  margin-left: 10px;
  font-size: 13px;
  color: #909399;
}

.title-link {
  color: #409eff;
  cursor: pointer;
  font-weight: 500;
}

.title-link:hover {
  text-decoration: underline;
}

.summary {
  margin-top: 2px;
  font-size: 12px;
  color: #9aa1ab;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 460px;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
