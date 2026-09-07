<template>
  <div class="contest-container page-container f-panel">
    <!-- 顶栏：标题 + 搜索 + 类型/状态筛选 -->
    <div class="flex flex-wrap items-center gap-3 border-b border-gray-200 px-5 py-4 max-md:items-stretch max-md:px-3 max-md:py-3">
      <span class="mr-auto text-xl text-gray-600 max-md:w-full">所有比赛</span>
      <el-input v-model="filters.keyword" placeholder="搜索比赛…" clearable class="w-52! max-md:w-full!" @keyup.enter="handleSearch"
        @clear="handleSearch" />
      <el-select v-model="filters.type" class="w-32! max-md:min-w-30 max-md:flex-1" @change="handleSearch">
        <el-option label="全部类型" :value="0" />
        <el-option label="ACM" :value="1" />
        <el-option label="OI" :value="2" />
        <el-option label="IOI" :value="3" />
        <el-option label="CF" :value="4" />
      </el-select>
      <el-select v-model="filters.status" class="w-32! max-md:min-w-30 max-md:flex-1" @change="handleSearch">
        <el-option label="全部状态" :value="-1" />
        <el-option label="未开始" :value="0" />
        <el-option label="进行中" :value="1" />
        <el-option label="已结束" :value="2" />
      </el-select>
      <el-button type="primary" class="max-md:ml-0! max-md:w-full!" @click="handleSearch">搜索</el-button>
      <el-button v-if="userStore.isAdmin" class="max-md:ml-0! max-md:w-full!" @click="handleCreateContest">创建比赛</el-button>
    </div>

    <div class="flex flex-col">
      <div class="border-b border-gray-100 last:border-b-0" v-for="contest in contestList" :key="contest.id">
        <ContestCard :contest="contest" @go-contest="handleGoContest" />
      </div>
      <el-empty v-if="!contestList.length" description="暂无比赛" />
    </div>

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination layout="prev, pager, next" :current-page="page" :page-size="pageSize" :total="total"
        @current-change="handlePageChange" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from "vue";
import { getContestList } from "@/api/contest";
import { useRoute, useRouter } from "vue-router";
import ContestCard from "./components/ContestCard.vue";
import { useUserStore } from "@/stores/user";

const userStore = useUserStore();
const route = useRoute();
const router = useRouter();

const page = computed(() => Number(route.query.page || 1));
const pageSize = computed(() => Number(route.query.pageSize || 15));
const total = ref(0);
const contestList = ref([]);

// 筛选状态与 URL 同步（type 0=全部；status -1=全部）
const filters = reactive({ keyword: "", type: 0, status: -1 });

const getList = async () => {
  const params = { page: page.value, pageSize: pageSize.value };
  if (route.query.keyword) params.keyword = route.query.keyword;
  if (route.query.type) params.type = Number(route.query.type);
  if (route.query.status != null && route.query.status !== "") {
    params.status = Number(route.query.status);
  }
  const res = await getContestList(params);
  contestList.value = res.data.list || [];
  total.value = res.data.total;
};

// 搜索：回到第 1 页；默认值不写进 URL
const handleSearch = () => {
  const query = { page: 1, pageSize: pageSize.value };
  if (filters.keyword) query.keyword = filters.keyword;
  if (filters.type) query.type = filters.type;
  if (filters.status >= 0) query.status = filters.status;
  router.push({ query });
};

const handleCreateContest = () => router.push({ name: "CreateContest" });
const handleGoContest = (id) => router.push(`/contest/${id}`);
const handlePageChange = (val) => {
  router.push({ query: { ...route.query, page: val, pageSize: pageSize.value } });
};

// URL 变化 → 回填筛选框 + 拉列表（back/forward、直接带参进入都能对上）
watch(
  () => route.query,
  (q) => {
    filters.keyword = q.keyword || "";
    filters.type = Number(q.type || 0);
    filters.status = q.status != null && q.status !== "" ? Number(q.status) : -1;
    getList();
  },
  { immediate: true, deep: true }
);
</script>

<style scoped lang="scss">
.contest-container {
  max-width: 1000px;
}
</style>
