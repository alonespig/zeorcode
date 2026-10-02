<script setup>
import { ref, shallowRef, computed, inject, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Search } from "@element-plus/icons-vue";
import { getTeamMembers, removeMember } from "@/api/team";
import StudentImportDialog from "@/pages/Team/components/StudentImportDialog.vue";
import MemberTable from "@/pages/Team/components/MemberTable.vue";

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

const loadMembers = async () => {
  try {
    // 用路由参数而非 inject 的 team.value.id：onMounted 时 team 可能还没拉回来
    const res = await getTeamMembers(route.params.id);
    members.value = res.data?.list || [];
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

<template>
  <div class="f-panel">
    <div class="members-toolbar">
      <h2 class="text-[15px] font-semibold text-gray-700">成员列表</h2>
      <span class="text-sm text-gray-400">
        {{ keyword.trim() ? `${filteredMembers.length} / ${members.length} 人` : `${members.length} 人` }}
      </span>

      <div class="members-toolbar__actions">
        <el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索姓名、学号或账号" aria-label="搜索成员" class="members-search" />
        <el-button v-if="canManage" type="primary" plain @click="importVisible = true">导入学生</el-button>
      </div>
    </div>

    <MemberTable :members="filteredMembers" :can-manage="canManage" :can-own="canOwn"
      :empty-text="members.length ? '未找到匹配的成员' : '暂无成员'"
      @remove="remove" />

    <StudentImportDialog v-if="canManage" v-model="importVisible" :team-id="route.params.id" @imported="loadMembers" />
  </div>
</template>

<style scoped>
.members-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.members-toolbar__actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
}

.members-search {
  width: 260px;
}

@media (max-width: 640px) {
  .members-toolbar {
    padding: 14px 16px;
  }

  .members-toolbar__actions {
    width: 100%;
    margin-left: 0;
  }

  .members-search {
    flex: 1;
    min-width: 0;
    width: auto;
  }
}
</style>
