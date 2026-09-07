<script setup>
import { onMounted, ref, shallowRef } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { createTeam, getTeamList, joinTeam } from "@/api/team";
import TeamCard from "@/pages/Team/components/TeamCard.vue";
import TeamFilterPanel from "@/pages/Team/components/TeamFilterPanel.vue";
import SquareImageCropper from "@/components/upload/SquareImageCropper.vue";
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

const createVisible = ref(false);
const creating = ref(false);
const createForm = ref(blankTeam());
const joinVisible = ref(false);
const joining = ref(false);
const joinTarget = ref(null);
const inviteCode = ref("");

function blankTeam() {
  return { name: "", coverUrl: "", description: "", visibility: 0, inviteCode: "" };
}

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
  createForm.value = blankTeam();
  createVisible.value = true;
};

const submitCreate = async () => {
  if (!createForm.value.name.trim()) {
    ElMessage.warning("请填写团队名称");
    return;
  }
  if (createForm.value.visibility === 1 && !createForm.value.inviteCode.trim()) {
    ElMessage.warning("非公开团队必须设置邀请码");
    return;
  }
  creating.value = true;
  try {
    const res = await createTeam({
      ...createForm.value,
      name: createForm.value.name.trim(),
      inviteCode: createForm.value.inviteCode.trim(),
    });
    ElMessage.success("创建成功");
    createVisible.value = false;
    router.push(`/team/${res.data.id}`);
  } catch (err) {
    console.error(err);
  } finally {
    creating.value = false;
  }
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
      <el-pagination size="large" :page-size="pageSize" :pager-count="16" v-model:current-page="page"
        layout="prev, pager, next" :total="total" @current-change="loadList" />
    </div>

    <!-- 创建团队 -->
    <el-dialog v-model="createVisible" title="创建团队" width="min(920px, calc(100vw - 32px))"
      :close-on-click-modal="false">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="团队名称" required>
          <el-input v-model="createForm.name" maxlength="60" show-word-limit placeholder="团队名称" />
        </el-form-item>
        <el-form-item label="团队封面">
          <SquareImageCropper
            v-model="createForm.coverUrl"
            title="裁剪团队封面"
            remove-text="移除封面"
            success-message="团队封面已上传"
            file-prefix="team-cover"
            preview-alt="团队封面预览"
          />
        </el-form-item>
        <el-form-item label="团队简介">
          <MdEditor v-model="createForm.description" height="160px" editor-id="team-create-editor" />
        </el-form-item>
        <el-form-item label="公开度">
          <el-radio-group v-model="createForm.visibility">
            <el-radio :value="0">公开（任何人可加入）</el-radio>
            <el-radio :value="1">非公开（需邀请码）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="createForm.visibility === 1" label="邀请码">
          <el-input v-model="createForm.inviteCode" maxlength="32" placeholder="非公开团队必填" class="w-52!" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

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
