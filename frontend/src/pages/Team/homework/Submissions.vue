<script setup>
import { computed, inject, onMounted, ref, shallowRef, watch } from "vue";
import { useRoute } from "vue-router";
import { InfoFilled } from "@element-plus/icons-vue";
import { getHomeworkSubmissions, getTeamMembers } from "@/api/team";
import HomeworkSubmissionViewer from "@/components/homework/HomeworkSubmissionViewer.vue";
import { JUDGE_STATUS_TEXT, JUDGE_STATUS_CLASS } from "@/constants/index";

const route = useRoute();
const homework = inject("homework");
const problems = computed(() => homework.value.problems || []);
const list = ref([]);
const total = shallowRef(0);
const page = shallowRef(1);
const pageSize = 20;
const canSeeAll = shallowRef(false);
const filterUid = shallowRef(null);
const filterProblemId = shallowRef("");
const listLoading = shallowRef(false);
const members = ref([]);
const detailVisible = shallowRef(false);
const selectedSubmissionId = shallowRef(null);
let listFetchVersion = 0;

const statusText = (status) => JUDGE_STATUS_TEXT[status] || "Unknown";
const statusClass = (status) => JUDGE_STATUS_CLASS[status] || "unknown";

const memberName = (member) => member.realName || member.username || "-";
const memberInitial = (member) => memberName(member).slice(0, 1).toUpperCase();
const memberOptionLabel = (member) => member.studentNo
  ? `${memberName(member)}（${member.studentNo}）`
  : memberName(member);

const loadList = async () => {
  const version = ++listFetchVersion;
  listLoading.value = true;
  try {
    const res = await getHomeworkSubmissions(route.params.hid, {
      page: page.value,
      pageSize,
      uid: canSeeAll.value ? filterUid.value || undefined : undefined,
      problemId: filterProblemId.value || undefined,
    });
    if (version !== listFetchVersion) return;
    list.value = res.data?.list || [];
    total.value = res.data?.total || 0;
    canSeeAll.value = res.data?.canSeeAll ?? false;
  } catch (err) {
    console.error(err);
    if (version === listFetchVersion) {
      list.value = [];
      total.value = 0;
    }
  } finally {
    if (version === listFetchVersion) listLoading.value = false;
  }
};

const handleFilterChange = () => {
  page.value = 1;
  loadList();
};

const loadMembers = async () => {
  try {
    const res = await getTeamMembers(route.params.id);
    members.value = res.data?.list || [];
  } catch (err) {
    console.error(err);
  }
};

const openDetail = (row) => {
  selectedSubmissionId.value = row.id;
  detailVisible.value = true;
};

watch(canSeeAll, (value) => {
  if (value && !members.value.length) loadMembers();
});

onMounted(loadList);
</script>

<template>
  <div class="f-panel">
    <div class="submissions-toolbar">
      <h2 class="text-sm font-medium text-gray-700">提交列表</h2>
      <div class="submissions-filters">
        <el-select v-model="filterProblemId" placeholder="全部题目" aria-label="筛选题目" clearable filterable
          class="submissions-problem-filter" @change="handleFilterChange">
          <el-option label="全部题目" value="" />
          <el-option v-for="problem in problems" :key="problem.id" :value="problem.id"
            :label="`${problem.id} ${problem.name}`" />
        </el-select>
        <el-select v-if="canSeeAll" v-model="filterUid" placeholder="全部成员" aria-label="筛选成员" clearable
          class="submissions-member-filter" @change="handleFilterChange">
          <el-option v-for="member in members" :key="member.uid" :label="memberOptionLabel(member)"
            :value="member.uid" />
        </el-select>
      </div>
    </div>

    <el-table v-loading="listLoading" :data="list" style="width: 100%">
      <el-table-column prop="id" label="提交号" min-width="90" align="center" />
      <el-table-column v-if="canSeeAll" label="成员" min-width="190">
        <template #default="{ row }">
          <div class="flex min-w-0 items-center gap-2.5">
            <el-avatar :size="32" :src="row.avatar || undefined"
              class="shrink-0 bg-blue-50 text-xs font-semibold text-blue-600">
              {{ memberInitial(row) }}
            </el-avatar>
            <div class="min-w-0 leading-tight">
              <div class="truncate text-gray-800" :title="memberName(row)">{{ memberName(row) }}</div>
              <div class="mt-1 truncate text-xs text-gray-400" :title="row.studentNo || '暂无学号'">
                {{ row.studentNo || '暂无学号' }}
              </div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="题号" min-width="100" align="center">
        <template #default="{ row }">
          <router-link class="text-blue-500 hover:underline"
            :to="{ name: 'HomeworkProblemDetail', params: { id: route.params.id, hid: route.params.hid, problemId: row.problemId } }">
            {{ row.problemId }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="结果" min-width="190" align="center">
        <template #default="{ row }">
          <span class="font-medium" :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="得分" min-width="80" align="center">
        <template #default="{ row }">
          <span class="font-medium tabular-nums"
            :class="row.score >= 100 ? 'text-emerald-600' : row.score > 0 ? 'text-amber-600' : 'text-gray-400'">
            {{ row.score }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="语言" min-width="90" align="center">
        <template #default="{ row }">{{ row.language }}</template>
      </el-table-column>
      <el-table-column label="用时" min-width="100" align="center">
        <template #default="{ row }">{{ row.timeUsed ? row.timeUsed + " ms" : "-" }}</template>
      </el-table-column>
      <el-table-column label="提交时间" min-width="170" align="center">
        <template #default="{ row }">
          <span class="tabular-nums text-gray-500">{{ row.createdAt }}</span>
          <el-tooltip v-if="!row.inWindow" content="时间窗外提交，不计入排行榜" placement="top">
            <el-icon class="ml-1 align-middle text-gray-400" :size="13"><InfoFilled /></el-icon>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="80" align="center" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">查看</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无提交" :image-size="80" />
      </template>
    </el-table>

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination v-model:current-page="page" size="large" :page-size="pageSize" :pager-count="7"
        layout="prev, pager, next" :total="total" @current-change="loadList" />
    </div>

    <HomeworkSubmissionViewer v-model="detailVisible" :homework-id="route.params.hid"
      :submission-id="selectedSubmissionId" @settled="loadList" />
  </div>
</template>

<style scoped>
.submissions-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.submissions-filters { display: flex; flex-wrap: wrap; gap: 10px; margin-left: auto; max-width: 100%; }
.submissions-problem-filter { width: 260px; max-width: 100%; }
.submissions-member-filter { width: 160px; max-width: 100%; }

@media (max-width: 640px) {
  .submissions-filters { width: 100%; margin-left: 0; }
  .submissions-problem-filter, .submissions-member-filter { width: 100%; }
}
</style>
