<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="t">用户管理</span>
          <span class="sub">共 {{ total }} 名用户</span>
        </div>
        <el-button type="primary" @click="importVisible = true">
          <el-icon class="mr-1">
            <Plus />
          </el-icon>导入用户
        </el-button>
      </div>
    </template>

    <el-table v-loading="loading" :data="userList" stripe>
      <el-table-column prop="id" label="ID" width="90" />
      <el-table-column prop="username" label="用户名" min-width="160" show-overflow-tooltip />
      <el-table-column prop="studentNo" label="学号" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.studentNo || '—' }}</template>
      </el-table-column>
      <el-table-column prop="realName" label="姓名" min-width="120" show-overflow-tooltip />
      <el-table-column prop="email" label="邮箱" min-width="210" show-overflow-tooltip />
      <el-table-column label="角色" width="140">
        <template #default="{ row }">
          <!-- 用单向 :model-value 而非 v-model：取消确认时父组件不改 row.role，
               el-select 会自动回弹到原值，无需手动回滚 -->
          <el-tooltip content="不能修改自己的角色" placement="top" :disabled="row.id !== currentUserId">
            <span>
              <el-select :model-value="row.role" size="small" style="width: 110px"
                :disabled="row._roleBusy || row.id === currentUserId" @change="(val) => changeRole(row, val)">
                <el-option :value="0" label="普通用户" />
                <el-option :value="1" label="管理员" />
              </el-select>
            </span>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'danger' : 'success'" size="small" effect="light">
            {{ row.status === 1 ? '封禁' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="signature" label="签名" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">
          <span :class="{ 'text-empty': !row.signature }">{{ row.signature || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.role !== 1" :type="row.status === 1 ? 'success' : 'danger'" size="small" link
            :loading="row._busy" @click="toggleBan(row)">
            {{ row.status === 1 ? '解封' : '封禁' }}
          </el-button>
          <span v-else class="text-empty">—</span>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination background layout="total, prev, pager, next" :current-page="page" :page-size="pageSize"
        :total="total" @current-change="handlePageChange" />
    </div>

    <UserImportDialog v-model="importVisible" @imported="handleImported" />
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { getAdminUserList, setUserStatus, setUserRole } from "@/api/admin";
import { useUserStore } from "@/stores/user";
import UserImportDialog from "@/pages/Admin/components/UserImportDialog.vue";

const userStore = useUserStore();
// row.id 与 user.id 同为对外用户号（UID），可直接比较
const currentUserId = computed(() => userStore.user?.id);

const userList = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);

const importVisible = ref(false);

const formatTime = (ts) => {
  if (!ts) return "-";
  const d = new Date(ts * 1000);
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const loadUsers = async () => {
  loading.value = true;
  try {
    const res = await getAdminUserList({ page: page.value, pageSize: pageSize.value });
    userList.value = res.data?.list || [];
    total.value = res.data?.total || 0;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (val) => {
  page.value = val;
  loadUsers();
};

const toggleBan = async (row) => {
  const ban = row.status !== 1;
  try {
    await ElMessageBox.confirm(
      `确认${ban ? "封禁" : "解封"}用户「${row.username}」？${ban ? "封禁后该用户将无法登录并被立即踢下线。" : ""}`,
      ban ? "封禁用户" : "解封用户",
      { type: "warning" }
    );
  } catch {
    return;
  }
  row._busy = true;
  try {
    await setUserStatus(row.id, ban ? 1 : 0);
    row.status = ban ? 1 : 0;
    ElMessage.success(ban ? "已封禁" : "已解封");
  } catch (err) {
    console.error(err);
  } finally {
    row._busy = false;
  }
};

const changeRole = async (row, newRole) => {
  const toAdmin = newRole === 1;
  try {
    await ElMessageBox.confirm(
      `确认将用户「${row.username}」${toAdmin ? "设为管理员" : "降为普通用户"}？该用户会被立即踢下线，需重新登录后新权限才生效。`,
      toAdmin ? "设为管理员" : "取消管理员",
      { type: "warning" }
    );
  } catch {
    return; // 取消：未改 row.role，下拉自动回弹
  }
  row._roleBusy = true;
  try {
    await setUserRole(row.id, newRole);
    row.role = newRole;
    ElMessage.success(toAdmin ? "已设为管理员" : "已降为普通用户");
  } catch (err) {
    console.error(err);
  } finally {
    row._roleBusy = false;
  }
};

const handleImported = () => {
  page.value = 1;
  loadUsers();
};

onMounted(loadUsers);
</script>

<style scoped>
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.head-title .t {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.head-title .sub {
  margin-left: 10px;
  font-size: 13px;
  color: #909399;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.text-empty {
  color: #c0c4cc;
}

.mr-1 {
  margin-right: 4px;
}

</style>
