<template>
  <div class="grid grid-cols-1 gap-4 lg:grid-cols-[1fr_300px]">
    <!-- 左：团队简介 -->
    <div class="space-y-4">
      <div class="f-panel">
        <div class="px-5 py-4">
          <MdEditor v-if="team.description" :model-value="team.description" preview-only editor-id="team-desc" />
          <span v-else class="text-sm text-gray-400">暂无简介</span>
        </div>
      </div>
    </div>

    <!-- 右：信息栏 + 操作 -->
    <div class="space-y-4">
      <div class="f-panel">
        <div class="divide-y divide-gray-100 text-sm">
          <div class="flex justify-between px-5 py-2.5">
            <span class="text-gray-600 text-[13px] font-semibold">所有者</span>
            <span class="text-gray-700 text-[14px]">{{ team.owner }}</span>
          </div>
          <div class="flex justify-between px-5 py-2.5">
            <span class="text-gray-600 text-[13px] font-semibold">公开度</span>
            <span class="text-gray-700 text-[14px]">{{ team.visibility === 1 ? "非公开" : "公开" }}</span>
          </div>
          <div class="flex justify-between px-5 py-2.5">
            <span class="text-gray-600 text-[13px] font-semibold">团队编号</span>
            <span class="font-mono text-gray-700 text-[14px]">{{ team.id }}</span>
          </div>
          <div v-if="isTeamMember" class="flex justify-between px-5 py-2.5">
            <span class="text-gray-600 text-[13px] font-semibold">我的角色</span>
            <el-tag size="small" :type="roleTagType">{{ roleText }}</el-tag>
          </div>
        </div>
      </div>

      <!-- 权限操作必须同时满足已登录和服务端授权 -->
      <div v-if="isTeamMember || canManageTeam" class="f-panel flex flex-col gap-2.5 p-4">
        <el-button v-if="canManageTeam" type="primary" @click="openManage">管理团队</el-button>
        <el-button v-if="canManageTeam" @click="openCreateHomework">布置作业</el-button>
        <el-button v-if="canDisbandTeam" type="danger" plain @click="disband">解散团队</el-button>
        <el-button v-else-if="isTeamMember" type="danger" plain @click="quit">退出团队</el-button>
      </div>
      <!-- 游客先登录，避免显示可直接提交的加入操作 -->
      <div v-else-if="!userStore.isLogin" class="f-panel p-4">
        <el-button type="primary" class="w-full" @click="userStore.promptLogin">登录后加入</el-button>
      </div>
      <!-- 非成员且公开团队 → 加入 -->
      <div v-else-if="team.visibility === 0" class="f-panel p-4">
        <el-button type="primary" class="w-full" @click="join">加入团队</el-button>
      </div>
      <!-- 非成员且非公开 → 输入邀请码 -->
      <div v-else class="f-panel p-4">
        <el-input v-model="inviteCode" placeholder="请输入邀请码" class="mb-2.5" @keyup.enter="joinByCode" />
        <el-button type="primary" class="w-full" @click="joinByCode">加入团队</el-button>
      </div>
    </div>

    <!-- 管理团队弹窗 -->
    <el-dialog v-model="manageVisible" title="管理团队" width="min(920px, calc(100vw - 32px))"
      :close-on-click-modal="false">
      <el-form :model="manageForm" label-width="80px">
        <el-form-item label="团队名称" required>
          <el-input v-model="manageForm.name" maxlength="60" show-word-limit />
        </el-form-item>
        <el-form-item label="团队封面">
          <SquareImageCropper
            v-model="manageForm.coverUrl"
            title="裁剪团队封面"
            remove-text="移除封面"
            success-message="团队封面已上传"
            file-prefix="team-cover"
            preview-alt="团队封面预览"
          />
        </el-form-item>
        <el-form-item label="团队简介">
          <MdEditor v-model="manageForm.description" height="180px" editor-id="team-edit-editor" />
        </el-form-item>
        <el-form-item label="公开度">
          <el-radio-group v-model="manageForm.visibility">
            <el-radio :value="0">公开（任何人可加入）</el-radio>
            <el-radio :value="1">非公开（需邀请码）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="manageForm.visibility === 1" label="邀请码">
          <el-input v-model="manageForm.inviteCode" maxlength="32"
            placeholder="请输入团队邀请码" class="w-52!" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="manageVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveManage">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, inject } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { joinTeam, quitTeam, updateTeam, deleteTeam, getTeamDetail } from "@/api/team";
import SquareImageCropper from "@/components/upload/SquareImageCropper.vue";
import { useUserStore } from "@/stores/user";

const team = inject("team");
const router = useRouter();
const userStore = useUserStore();
const inviteCode = ref("");

const isTeamMember = computed(() => userStore.isLogin && team.value.myRole != null);
const canManageTeam = computed(() => userStore.isLogin && Boolean(team.value.canManage));
const canDisbandTeam = computed(() => userStore.isLogin && team.value.myRole === 2);

const roleText = computed(() => ({ 0: "成员", 1: "管理员", 2: "所有者" }[team.value.myRole] || "成员"));
const roleTagType = computed(() => ({ 2: "warning", 1: "primary", 0: "info" }[team.value.myRole] || "info"));

// ===== 加入 / 退出 =====
const join = async () => {
  try {
    await joinTeam(team.value.id);
    ElMessage.success("已加入");
    location.reload();
  } catch (err) {
    console.error(err);
  }
};

const joinByCode = async () => {
  if (!inviteCode.value.trim()) {
    ElMessage.warning("请输入邀请码");
    return;
  }
  try {
    await joinTeam(team.value.id, inviteCode.value.trim());
    ElMessage.success("已加入");
    location.reload();
  } catch (err) {
    console.error(err);
  }
};

const quit = async () => {
  try {
    await ElMessageBox.confirm(`确认退出团队「${team.value.name}」？`, "退出团队", { type: "warning" });
  } catch {
    return;
  }
  try {
    await quitTeam(team.value.id);
    ElMessage.success("已退出");
    router.push("/team");
  } catch (err) {
    console.error(err);
  }
};

const disband = async () => {
  try {
    await ElMessageBox.confirm(
      `确认解散团队「${team.value.name}」？所有成员、作业都会被删除，此操作不可恢复。`,
      "解散团队",
      { type: "error", confirmButtonText: "确认解散" }
    );
  } catch {
    return;
  }
  try {
    await deleteTeam(team.value.id);
    ElMessage.success("已解散");
    router.push("/team");
  } catch (err) {
    console.error(err);
  }
};

// ===== 管理团队 =====
const manageVisible = ref(false);
const saving = ref(false);
const manageForm = ref({ name: "", coverUrl: "", description: "", visibility: 0, inviteCode: "" });

const openManage = () => {
  manageForm.value = {
    name: team.value.name,
    coverUrl: team.value.coverUrl || "",
    description: team.value.description || "",
    visibility: team.value.visibility,
    inviteCode: team.value.inviteCode || "",
  };
  manageVisible.value = true;
};

const saveManage = async () => {
  if (!manageForm.value.name.trim()) {
    ElMessage.warning("请填写团队名称");
    return;
  }
  if (manageForm.value.visibility === 1 && !manageForm.value.inviteCode.trim()) {
    ElMessage.warning("非公开团队必须设置邀请码");
    return;
  }
  saving.value = true;
  try {
    await updateTeam(team.value.id, {
      name: manageForm.value.name.trim(),
      coverUrl: manageForm.value.coverUrl,
      description: manageForm.value.description,
      visibility: manageForm.value.visibility,
      inviteCode: manageForm.value.inviteCode.trim(),
    });
    ElMessage.success("已保存");
    manageVisible.value = false;
    // 重新拉详情，更新横幅里的名称/公开度等
    const res = await getTeamDetail(team.value.id);
    team.value = res.data;
  } catch (err) {
    console.error(err);
  } finally {
    saving.value = false;
  }
};

const openCreateHomework = () => router.push(`/team/${team.value.id}/homework?create=1`);
</script>
