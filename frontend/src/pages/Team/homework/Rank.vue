<template>
  <div class="f-panel">
    <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3">
      <h3 class="text-base font-semibold text-gray-800">排行榜</h3>
      <el-button size="small" :disabled="!rank.list?.length" @click="exportCsv">导出 CSV</el-button>
    </div>

    <div v-if="rank.list?.length" class="overflow-x-auto p-5">
      <!-- 用 data-table 类对齐项目原生表格/el-table 的表头风格 -->
      <table class="data-table w-full min-w-[640px]">
        <thead>
          <tr>
            <th class="w-16">名次</th>
            <th class="left">成员</th>
            <th class="w-20">总分</th>
            <th class="w-20">已通过</th>
            <th v-for="pid in rank.problemIds" :key="pid" class="min-w-[72px]">{{ pid }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rank.list" :key="row.uid" :class="{ 'is-me': isMe(row) }">
            <td class="font-semibold" :class="row.rank <= 3 ? 'text-amber-600' : 'text-gray-600'">
              #{{ row.rank }}
            </td>
            <td class="left">
              <div class="flex items-center gap-2">
                <el-avatar :size="24" :src="row.avatar || undefined" />
                <span class="font-medium text-gray-800">{{ row.realName || row.username }}</span>
                <span class="text-xs text-gray-400">{{ row.studentNo || '暂无学号' }}</span>
                <el-tag v-if="isMe(row)" size="small">本人</el-tag>
              </div>
            </td>
            <td class="font-bold tabular-nums text-gray-800">{{ row.totalScore }}</td>
            <td class="tabular-nums text-gray-600">{{ row.solvedCount }}</td>
            <td v-for="cell in row.cells" :key="cell.problemId" class="tabular-nums font-medium"
              :class="cellClass(cell)">
              {{ cell.score > 0 ? cell.score : '—' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <el-empty v-else description="暂无成绩" :image-size="80" />
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { downloadHomeworkRank, getHomeworkRank } from "@/api/team";
import { useUserStore } from "@/stores/user";

const hw = inject("homework");
const route = useRoute();
const userStore = useUserStore();
const rank = ref({ list: [], problemIds: [], startTime: "", endTime: "", totalScore: 0 });

const isMe = (row) => userStore.user?.id === row.uid;
const cellClass = (cell) =>
  cell.solved ? "text-emerald-600" : cell.score > 0 ? "text-amber-600" : "text-gray-300";

const loadRank = async () => {
  try {
    // 直接用路由参数，不依赖 inject 的 hw 是否已加载完
    const res = await getHomeworkRank(route.params.hid);
    rank.value = res.data;
  } catch (err) {
    console.error(err);
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
/* 本人行高亮：覆盖 data-table 的斑马纹，用项目 hover 蓝 */
:deep(.data-table tbody tr.is-me),
:deep(.data-table tbody tr.is-me:nth-child(even)) {
  background-color: #eff6ff;
}
</style>
