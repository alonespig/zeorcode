<script setup>
import { onMounted, ref, shallowRef, watch } from "vue";
import { useRoute } from "vue-router";
import { InfoFilled } from "@element-plus/icons-vue";
import { getHomeworkSubmissionDetail, getHomeworkSubmissions, getTeamMembers } from "@/api/team";
import HomeworkSubmissionDialog from "@/components/homework/HomeworkSubmissionDialog.vue";

const route = useRoute();
const list = ref([]);
const total = shallowRef(0);
const page = shallowRef(1);
const pageSize = 20;
const canSeeAll = shallowRef(false);
const filterUid = shallowRef(null);
const members = ref([]);
const detailVisible = shallowRef(false);
const detailLoading = shallowRef(false);
const selectedSubmission = shallowRef({});
let detailFetchVersion = 0;

const statusText = (status) => ({
  0: "等待中",
  1: "通过",
  2: "内存超限",
  3: "时间超限",
  4: "运行错误",
  5: "答案错误",
  6: "编译错误",
  7: "系统错误",
  8: "格式错误",
}[status] || `状态${status}`);

const statusClass = (status) => ({
  1: "text-emerald-600",
  6: "text-amber-600",
  0: "text-blue-500",
}[status] || "text-red-500");

const memberName = (member) => member.realName || member.username || "-";
const memberInitial = (member) => memberName(member).slice(0, 1).toUpperCase();
const memberOptionLabel = (member) => member.studentNo
  ? `${memberName(member)}（${member.studentNo}）`
  : memberName(member);

const loadList = async () => {
  try {
    const res = await getHomeworkSubmissions(route.params.hid, {
      page: page.value,
      pageSize,
      uid: canSeeAll.value ? filterUid.value || undefined : undefined,
    });
    list.value = res.data?.list || [];
    total.value = res.data?.total || 0;
    canSeeAll.value = res.data?.canSeeAll ?? false;
  } catch (err) {
    console.error(err);
  }
};

const handleMemberChange = () => {
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

const openDetail = async (row) => {
  const version = ++detailFetchVersion;
  selectedSubmission.value = {};
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    const res = await getHomeworkSubmissionDetail(route.params.hid, row.id);
    if (version === detailFetchVersion) selectedSubmission.value = res.data || {};
  } catch (err) {
    console.error(err);
    if (version === detailFetchVersion) detailVisible.value = false;
  } finally {
    if (version === detailFetchVersion) detailLoading.value = false;
  }
};

watch(canSeeAll, (value) => {
  if (value && !members.value.length) loadMembers();
});

onMounted(loadList);
</script>

<template>
  <div class="f-panel">
    <div class="flex flex-wrap items-center gap-2.5 border-b border-gray-100 px-5 py-3">
      <span class="text-sm text-gray-400">共 {{ total }} 条提交{{ canSeeAll ? "" : "（仅自己的）" }}</span>
      <div v-if="canSeeAll" class="ml-auto flex items-center gap-2.5">
        <el-select v-model="filterUid" placeholder="全部成员" clearable class="w-36!" @change="handleMemberChange">
          <el-option v-for="member in members" :key="member.uid" :label="memberOptionLabel(member)"
            :value="member.uid" />
        </el-select>
      </div>
    </div>

    <el-table :data="list" style="width: 100%">
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
      <el-table-column label="结果" min-width="120" align="center">
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

    <HomeworkSubmissionDialog v-model="detailVisible" :submission="selectedSubmission" :loading="detailLoading" />
  </div>
</template>
