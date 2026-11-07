<script setup>
import { onMounted, ref, shallowRef } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { getTeamList, joinTeam } from "@/api/team";
import TeamCard from "@/pages/Team/components/TeamCard.vue";
import TeamFilterPanel from "@/pages/Team/components/TeamFilterPanel.vue";
import { useUserStore } from "@/stores/user";

const router = useRouter();
const userStore = useUserStore();

const list = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;
const q = ref("");
const mine = ref(false);
const visibility = shallowRef(null);

const joinVisible = ref(false);
const joining = ref(false);
const joinTarget = ref(null);
const inviteCode = ref("");

const toggleMine = () => {
  if (!userStore.isLogin) return;
  mine.value = !mine.value;
  page.value = 1;
  loadList();
};

const switchVisibility = (value) => {
  if (visibility.value === value) return;
  visibility.value = value;
  page.value = 1;
  loadList();
};

const loadList = async () => {
  try {
    const res = await getTeamList({
      page: page.value,
      pageSize,
      q: q.value || undefined,
      mine: mine.value ? 1 : undefined,
      visibility: visibility.value ?? undefined,
    });
    list.value = res.data?.list || [];
    total.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  }
};

const applyFilter = () => {
  page.value = 1;
  loadList();
};

const goTeam = (row) => router.push(`/team/${row.id}`);

const join = async (row) => {
  if (!userStore.isLogin) {
    userStore.promptLogin();
    return;
  }
  try {
    await ElMessageBox.confirm(`确认加入团队「${row.name}」？`, "加入团队", { type: "info" });
  } catch {
    return;
  }
  try {
    await joinTeam(row.id);
    ElMessage.success("已加入");
    loadList();
  } catch (err) {
    console.error(err);
  }
};

const openJoin = (row) => {
  if (!userStore.isLogin) {
    userStore.promptLogin();
    return;
  }
  joinTarget.value = row;
  inviteCode.value = "";
  joinVisible.value = true;
};

const submitJoin = async () => {
  if (!inviteCode.value.trim()) {
    ElMessage.warning("请输入邀请码");
    return;
  }
  joining.value = true;
  try {
    await joinTeam(joinTarget.value.id, inviteCode.value.trim());
    ElMessage.success("已加入");
    joinVisible.value = false;
    loadList();
  } catch (err) {
    console.error(err);
  } finally {
    joining.value = false;
  }
};

const openCreate = () => {
  router.push({ name: "TeamCreate" });
};

onMounted(loadList);
</script>

<template>
  <div class="team-list page-container">
    <TeamFilterPanel
      v-model:keyword="q"
      :mine="mine"
      :visibility="visibility"
      :total="total"
      :is-login="userStore.isLogin"
      @search="applyFilter"
      @toggle-mine="toggleMine"
      @change-visibility="switchVisibility"
      @create="openCreate"
    />

    <div v-if="list.length" class="team-grid">
      <TeamCard
        v-for="team in list"
        :key="team.id"
        :team="team"
        :is-login="userStore.isLogin"
        @enter="goTeam"
        @join="join"
        @request-join="openJoin"
      />
    </div>
    <div v-else class="empty-panel f-panel">
      <el-empty :description="mine ? '还没加入任何团队' : '暂无团队'" :image-size="80" />
    </div>

    <div v-if="total > pageSize" class="f-pagination">
      <AppPagination v-model:current-page="page" :page-size="pageSize" :total="total"
        @current-change="loadList" />
    </div>

    <!-- 加入非公开团队 -->
    <el-dialog v-model="joinVisible" title="加入团队" width="420px">
      <p class="mb-4 text-sm text-gray-600">「{{ joinTarget?.name }}」需要邀请码才能加入。</p>
      <el-input v-model="inviteCode" placeholder="请输入邀请码" @keyup.enter="submitJoin" />
      <template #footer>
        <el-button @click="joinVisible = false">取消</el-button>
        <el-button type="primary" :loading="joining" @click="submitJoin">加入</el-button>
      </template>
  </el-dialog>
</div>
</template>

<style scoped>
.team-list.page-container { max-width: 1200px; padding-top: 8px; padding-bottom: 24px; }
.team-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
.empty-panel { padding: 28px 0; }
@media (max-width: 1040px) {
  .team-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 680px) {
  .team-grid { grid-template-columns: 1fr; }
}
</style>
