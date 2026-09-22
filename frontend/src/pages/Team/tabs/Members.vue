<template>
  <div class="f-panel">
    <div class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-5 py-3">
      <h2 class="text-[15px] font-semibold text-gray-700">成员列表</h2>
      <span class="text-sm text-gray-400">
        {{ keyword.trim() ? `${filteredMembers.length} / ${members.length} 人` : `${members.length} 人` }}
      </span>

      <div class="ml-auto flex items-center gap-2.5">
        <el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索成员…" class="w-56!" />
        <el-button v-if="canManage" type="primary" plain @click="importVisible = true">导入学生</el-button>
      </div>
    </div>

    <div v-if="filteredMembers.length" class="grid grid-cols-1 gap-3 p-5 md:grid-cols-2">
      <div v-for="m in filteredMembers" :key="m.uid"
        class="flex items-center gap-3 rounded border border-gray-200 bg-white p-3">
        <el-avatar :size="34" :src="m.avatar || undefined" />
        <div class="min-w-0">
          <div class="truncate font-medium text-blue-500">{{ m.realName || m.username }}</div>
          <div class="truncate text-xs text-gray-400">{{ m.studentNo || '暂无学号' }}</div>
        </div>
        <div class="ml-auto flex items-center gap-2">
          <el-tag size="small" :type="roleMeta(m.role).type" effect="light">{{ roleMeta(m.role).text }}</el-tag>

          <!-- 管理操作：仅所有者能设/撤管理员；所有者和管理员能移出（规则在后端兜底） -->
          <template v-if="canManage && m.role !== 2">
            <el-button v-if="m.role === 0 && canOwn" link size="small" type="primary"
              @click="setRole(m, 1)">设为管理员</el-button>
            <el-button v-if="m.role === 1 && canOwn" link size="small" @click="setRole(m, 0)">取消管理</el-button>
            <el-button v-if="canRemove(m)" link size="small" type="danger" @click="remove(m)">移出</el-button>
          </template>
        </div>
      </div>
    </div>

    <el-empty v-else :description="members.length ? '未找到匹配的成员' : '暂无成员'" :image-size="80" />

    <StudentImportDialog v-if="canManage" v-model="importVisible" :team-id="route.params.id" @imported="loadMembers" />
  </div>
</template>

<script setup>
import { ref, shallowRef, computed, inject, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Search } from "@element-plus/icons-vue";
import { getTeamMembers, setMemberRole, removeMember } from "@/api/team";
import StudentImportDialog from "@/pages/Team/components/StudentImportDialog.vue";

const team = inject("team");
const route = useRoute();
const members = ref([]);
const keyword = ref("");
const importVisible = shallowRef(false);

const canManage = computed(() => team.value.canManage);
const canOwn = computed(() => team.value.myRole === 2);
const filteredMembers = computed(() => {
  const q = keyword.value.trim().toLocaleLowerCase();
  if (!q) return members.value;
  return members.value.filter((member) =>
    [member.realName, member.studentNo, member.username, String(member.uid)].some((value) =>
      String(value || "").toLocaleLowerCase().includes(q)
    )
  );
});

const roleMeta = (r) => ({
  0: { text: "成员", type: "info" },
  1: { text: "管理员", type: "primary" },
  2: { text: "所有者", type: "warning" },
}[r] || { text: "-", type: "info" });

// 管理员不能移出另一个管理员，只能移出普通成员（规则与后端一致）
const canRemove = (m) => (canOwn.value ? m.role !== 2 : m.role === 0);

const loadMembers = async () => {
  try {
    // 用路由参数而非 inject 的 team.value.id：onMounted 时 team 可能还没拉回来
    const res = await getTeamMembers(route.params.id);
    members.value = res.data?.list || [];
  } catch (err) {
    console.error(err);
  }
};

const setRole = async (m, role) => {
  try {
    await setMemberRole(team.value.id, m.uid, role);
    ElMessage.success(role === 1 ? "已设为管理员" : "已取消管理");
    loadMembers();
  } catch (err) {
    console.error(err);
  }
};

const remove = async (m) => {
  try {
    await ElMessageBox.confirm(`确认将「${m.realName || m.username}」移出团队？`, "移出成员", { type: "warning" });
  } catch {
    return;
  }
  try {
    await removeMember(team.value.id, m.uid);
    ElMessage.success("已移出");
    loadMembers();
  } catch (err) {
    console.error(err);
  }
};

onMounted(loadMembers);
</script>
