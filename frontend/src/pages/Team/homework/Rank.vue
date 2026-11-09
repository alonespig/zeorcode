<template>
  <div class="f-panel">
    <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5">
      <div class="flex items-center gap-3">
        <h2 class="text-[15px] font-semibold text-gray-700">排行榜</h2>
        <span class="text-[13px] text-gray-400">共 {{ rank.problemIds?.length ?? 0 }} 题</span>
      </div>
      <div class="flex items-center gap-4">
        <span v-if="mineText" class="text-[13px] text-gray-500">{{ mineText }}</span>
        <el-button size="small" :disabled="!rank.list?.length" @click="exportCsv">导出 CSV</el-button>
      </div>
    </div>

    <el-table
      v-loading="loading"
      :data="pagedList"
      style="width: 100%"
      :row-class-name="rowClassName"
    >
      <el-table-column label="排名" width="80" align="center" fixed="left">
        <template #default="{ row }">
          <span class="font-semibold tabular-nums" :class="row.rank <= 3 ? 'text-amber-600' : 'text-gray-600'">
            {{ row.rank }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="成员" width="300" fixed="left">
        <template #default="{ row }">
          <router-link :to="`/user/${row.uid}`" class="flex min-w-0 items-center gap-2.5 text-gray-800 hover:text-blue-500">
            <el-avatar class="shrink-0" :size="34" :src="row.avatar || undefined" />
            <span class="min-w-0 truncate font-medium">{{ row.realName || row.username }}</span>
            <el-tooltip v-if="row.gender === 1" content="男" placement="top">
              <el-icon class="shrink-0 text-blue-500" aria-label="男"><Male /></el-icon>
            </el-tooltip>
            <el-tooltip v-else-if="row.gender === 2" content="女" placement="top">
              <el-icon class="shrink-0 text-pink-500" aria-label="女"><Female /></el-icon>
            </el-tooltip>
            <span v-if="row.studentNo" class="shrink-0 whitespace-nowrap text-[13px] text-gray-500 tabular-nums">{{ row.studentNo }}</span>
            <el-tag v-if="isMe(row)" class="shrink-0" size="small">本人</el-tag>
          </router-link>
        </template>
      </el-table-column>

      <el-table-column label="总分" width="110" align="center">
        <template #default="{ row }">
          <span class="font-semibold tabular-nums" :class="row.totalScore > 0 ? 'text-emerald-600' : 'text-gray-400'">
            {{ row.totalScore }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="已通过" width="110" align="center">
        <template #default="{ row }">
          <span class="font-semibold text-gray-800 tabular-nums">{{ row.solvedCount }}</span>
          <span class="text-gray-400"> / {{ rank.problemIds.length }}</span>
        </template>
      </el-table-column>

      <el-table-column
        v-for="(problemId, problemIndex) in rank.problemIds"
        :key="problemId"
        width="84"
        align="center"
      >
        <template #header>
          <router-link class="font-medium text-blue-500 hover:text-blue-400" :to="problemLink(problemId)">
            {{ problemId }}
          </router-link>
        </template>
        <template #default="{ row }">
          <span class="font-medium tabular-nums" :class="cellClass(row.cells?.[problemIndex])">
            {{ cellText(row.cells?.[problemIndex]) }}
          </span>
        </template>
      </el-table-column>

      <template #empty>
        <el-empty description="暂无成绩" :image-size="80" />
      </template>
    </el-table>

    <div v-if="total > pageSize" class="flex justify-center border-t border-gray-100 px-5 py-3.5">
      <AppPagination v-model:current-page="page" :page-size="pageSize" :total="total" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { Female, Male } from "@element-plus/icons-vue";
import { downloadHomeworkRank, getHomeworkRank } from "@/api/team";
import { useUserStore } from "@/stores/user";

const route = useRoute();
const userStore = useUserStore();
const loading = ref(false);
const rank = ref({ list: [], problemIds: [], startTime: "", endTime: "", totalScore: 0 });

const page = ref(1);
const pageSize = 20;
const total = computed(() => rank.value.list?.length ?? 0);
const pagedList = computed(() => {
  const start = (page.value - 1) * pageSize;
  return (rank.value.list ?? []).slice(start, start + pageSize);
});

const isMe = (row) => userStore.user?.id === row.uid;

const mineText = computed(() => {
  const mine = rank.value.list?.find(isMe);
  if (!mine) return "";
  return `我的排名：第 ${mine.rank} 名 · 总分 ${mine.totalScore} · 已通过 ${mine.solvedCount}/${rank.value.problemIds?.length ?? 0}`;
});

const rowClassName = ({ row }) => (isMe(row) ? "is-me" : "");

const problemLink = (problemId) =>
  `/team/${route.params.id}/homework/${route.params.hid}/problem/${problemId}`;

const cellText = (cell) => (cell?.submitted ? cell.score : "");

const cellClass = (cell) => {
  if (!cell?.submitted) return "";
  if (cell.solved) return "text-emerald-600";
  return cell.score > 0 ? "text-amber-600" : "text-red-500";
};

const loadRank = async () => {
  loading.value = true;
  try {
    const res = await getHomeworkRank(route.params.hid);
    rank.value = res.data;
    const maxPage = Math.max(1, Math.ceil((res.data?.list?.length ?? 0) / pageSize));
    if (page.value > maxPage) page.value = maxPage;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

// 导出 CSV：同源下载，带 cookie，后端返回 attachment
const exportCsv = async () => {
  try {
    const response = await downloadHomeworkRank(route.params.hid);
    const contentType = response.headers?.["content-type"] || "";
    if (contentType.includes("application/json")) {
      const payload = JSON.parse(await response.data.text());
      ElMessage.error(payload.msg || "导出失败");
      return;
    }
    const url = URL.createObjectURL(response.data);
    const a = document.createElement("a");
    a.href = url;
    a.download = "rank.csv";
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  } catch (err) {
    console.error(err);
  }
};

onMounted(loadRank);
</script>

<style scoped>
/* 本人行高亮：覆盖 el-table 默认斑马纹，用项目 hover 蓝 */
:deep(.el-table tr.is-me > td.el-table__cell) {
  background-color: #eff6ff;
}

/* 横向滚动条：默认 6px 太细且只在悬停时出现，加粗并常驻，方便题目列很多时拖动 */
:deep(.el-scrollbar__bar.is-horizontal) {
  left: 0;
  right: 0;
  bottom: 2px;
  height: 9px;
  opacity: 1;
}

:deep(.el-scrollbar__bar.is-horizontal > div) {
  background-color: #cbd5e1;
  border-radius: 999px;
}

:deep(.el-scrollbar__bar.is-horizontal:hover > div) {
  background-color: #94a3b8;
}
</style>
