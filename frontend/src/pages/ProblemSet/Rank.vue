<script setup>
import { computed, shallowRef, watch } from "vue";
import { useRoute } from "vue-router";
import { CloseBold, Female, Male, Select } from "@element-plus/icons-vue";
import { getProblemSetRank } from "@/api/problemset";

defineProps({
  detail: { type: Object, required: true },
});

const route = useRoute();
const loading = shallowRef(false);
const page = shallowRef(1);
const pageSize = 20;
const rank = shallowRef({ total: 0, problemCount: 0, problemIds: [], list: [], mine: null });
let rankSeq = 0;

const mineText = computed(() => {
  const mine = rank.value.mine;
  if (!mine) return "";
  const prefix = mine.rank > 0 ? `我的排名：第 ${mine.rank} 名` : "暂未进入排名";
  return `${prefix} · 已解决 ${mine.solvedCount} / ${rank.value.problemCount}`;
});

const loadRank = async () => {
  const seq = ++rankSeq;
  loading.value = true;
  try {
    const res = await getProblemSetRank(route.params.id, { page: page.value, pageSize });
    if (seq !== rankSeq) return;
    rank.value = res.data;
  } catch (err) {
    if (seq !== rankSeq) return;
    console.error(err);
  } finally {
    if (seq === rankSeq) loading.value = false;
  }
};

const rowClassName = ({ row }) => (row.isSelf ? "is-self" : "");
const cellStatus = (row, index) => row.cells?.[index]?.status ?? null;

watch(
  () => route.params.id,
  () => {
    page.value = 1;
    loadRank();
  },
  { immediate: true },
);
</script>

<template>
  <section class="f-panel" v-loading="loading">
    <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5">
      <h2 class="text-[15px] font-semibold text-gray-700">排行榜</h2>
      <span v-if="mineText" class="text-[13px] text-gray-500">{{ mineText }}</span>
    </div>

    <el-table :data="rank.list" style="width: 100%" :row-class-name="rowClassName">
      <el-table-column label="排名" width="90" align="center" fixed="left">
        <template #default="{ row }">
          <span class="font-semibold tabular-nums" :class="row.rank <= 3 ? 'text-amber-600' : 'text-gray-600'">
            {{ row.rank }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="用户" width="280" fixed="left">
        <template #default="{ row }">
          <router-link :to="`/user/${row.id}`" class="inline-flex items-center gap-2.5 text-gray-800 hover:text-blue-500">
            <el-avatar :size="34" :src="row.avatar || undefined" />
            <span class="font-medium">{{ row.username }}</span>
            <el-tooltip v-if="row.gender === 1" content="男" placement="top">
              <el-icon class="text-blue-500" aria-label="男"><Male /></el-icon>
            </el-tooltip>
            <el-tooltip v-else-if="row.gender === 2" content="女" placement="top">
              <el-icon class="text-pink-500" aria-label="女"><Female /></el-icon>
            </el-tooltip>
            <el-tag v-if="row.isSelf" size="small">本人</el-tag>
          </router-link>
        </template>
      </el-table-column>

      <el-table-column label="已解决" width="120" align="center">
        <template #default="{ row }">
          <span class="font-semibold text-emerald-600 tabular-nums">{{ row.solvedCount }}</span>
          <span class="text-gray-400"> / {{ rank.problemCount }}</span>
        </template>
      </el-table-column>

      <el-table-column
        v-for="(problemId, problemIndex) in rank.problemIds"
        :key="problemId"
        width="100"
        align="center"
      >
        <template #header>
          <router-link class="font-medium text-blue-500 hover:text-blue-400" :to="`/problem/${problemId}`">
            {{ problemId }}
          </router-link>
        </template>
        <template #default="{ row }">
          <el-tooltip v-if="cellStatus(row, problemIndex) === 1" content="已通过" placement="top">
            <el-icon color="#2f9e44" :size="18" aria-label="已通过">
              <Select />
            </el-icon>
          </el-tooltip>
          <el-tooltip v-else-if="cellStatus(row, problemIndex) != null" content="已尝试" placement="top">
            <el-icon color="#e5484d" :size="17" aria-label="已尝试但未通过">
              <CloseBold />
            </el-icon>
          </el-tooltip>
        </template>
      </el-table-column>

      <template #empty>
        <el-empty description="还没有用户尝试题单中的题目" :image-size="80" />
      </template>
    </el-table>

    <div v-if="rank.total > pageSize" class="flex justify-end border-t border-gray-100 px-5 py-3.5">
      <el-pagination
        v-model:current-page="page"
        background
        layout="prev, pager, next"
        :page-size="pageSize"
        :total="rank.total"
        @current-change="loadRank"
      />
    </div>
  </section>
</template>

<style scoped>
:deep(.el-table .is-self td.el-table__cell) {
  background-color: #eff6ff;
}
</style>
