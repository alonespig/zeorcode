<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import { useUserStoreHook } from "@/store/modules/user";
import {
  getAdminUserList,
  setUserRole,
  setUserStatus,
  type AdminUserItem
} from "@/api/admin/users";
import UserImportDialog from "./components/UserImportDialog.vue";

defineOptions({ name: "AdminUsers" });

/** 表格行数据，附加按行的提交中状态，用于防重复点击 */
type UserRow = AdminUserItem & {
  _statusBusy?: boolean;
  _roleBusy?: boolean;
};

/** 角色选项集中映射（对应后端 dto.AdminUserItem.Role） */
const ROLE_OPTIONS = [
  { value: 0, label: "普通用户" },
  { value: 1, label: "管理员" }
];

/** 状态集中映射（对应后端 dto.AdminUserItem.Status） */
const STATUS_LABELS: Record<number, string> = { 0: "正常", 1: "封禁" };
const STATUS_TAG_TYPES: Record<number, "success" | "danger"> = {
  0: "success",
  1: "danger"
};

const userStore = useUserStoreHook();
// 当前登录管理员对外用户号，禁止修改自己的角色
const currentUserId = computed(() => userStore.user?.id);

const dataList = ref<UserRow[]>([]);
const loading = ref(false);
const importVisible = ref(false);
const keyword = ref("");
const role = ref<number>();
const status = ref<number>();

const pagination = reactive<PaginationProps>({
  pageSize: 20,
  currentPage: 1,
  total: 0,
  pageSizes: [10, 20, 50, 100],
  layout: "total, sizes, prev, pager, next, jumper"
});

const columns: TableColumnList = [
  { label: "用户号", prop: "id", width: 90, align: "center" },
  {
    label: "用户名",
    prop: "username",
    minWidth: 140,
    showOverflowTooltip: true
  },
  { label: "学号", prop: "studentNo", minWidth: 130, slot: "studentNo" },
  { label: "姓名", prop: "realName", minWidth: 110, showOverflowTooltip: true },
  { label: "邮箱", prop: "email", minWidth: 190, slot: "email" },
  { label: "角色", slot: "role", width: 150, align: "center" },
  { label: "状态", slot: "status", width: 90, align: "center" },
  { label: "注册时间", slot: "createdAt", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 110,
    fixed: "right",
    align: "center"
  }
];

function formatTime(ts: number): string {
  if (!ts) return "—";
  const d = new Date(ts * 1000);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

async function onSearch() {
  loading.value = true;
  try {
    const res = await getAdminUserList({
      page: pagination.currentPage,
      pageSize: pagination.pageSize,
      q: keyword.value.trim() || undefined,
      role: role.value,
      status: status.value
    });
    dataList.value = res.data?.list ?? [];
    pagination.total = res.data?.total ?? 0;
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function applyFilter() {
  pagination.currentPage = 1;
  onSearch();
}

function resetFilter() {
  keyword.value = "";
  role.value = undefined;
  status.value = undefined;
  applyFilter();
}

function onCurrentChange(page: number) {
  pagination.currentPage = page;
  onSearch();
}

function onSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.currentPage = 1;
  onSearch();
}

async function toggleBan(row: UserRow) {
  const ban = row.status !== 1;
  try {
    await ElMessageBox.confirm(
      `确认${ban ? "封禁" : "解封"}用户「${row.username}」？${
        ban ? "封禁后该用户将无法登录并被立即踢下线。" : ""
      }`,
      ban ? "封禁用户" : "解封用户",
      { type: "warning" }
    );
  } catch {
    return;
  }
  if (row._statusBusy) return;
  row._statusBusy = true;
  try {
    await setUserStatus(row.id, { status: ban ? 1 : 0 });
    message(ban ? "已封禁" : "已解封", { type: "success" });
    onSearch();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._statusBusy = false;
  }
}

async function changeRole(row: UserRow, newRole: number) {
  if (row.id === currentUserId.value) return;
  const toAdmin = newRole === 1;
  try {
    await ElMessageBox.confirm(
      `确认将用户「${row.username}」${
        toAdmin ? "设为管理员" : "降为普通用户"
      }？该用户会被立即踢下线，需重新登录后新权限才生效。`,
      toAdmin ? "设为管理员" : "取消管理员",
      { type: "warning" }
    );
  } catch {
    return;
  }
  if (row._roleBusy) return;
  row._roleBusy = true;
  try {
    await setUserRole(row.id, { role: newRole });
    message(toAdmin ? "已设为管理员" : "已降为普通用户", { type: "success" });
    onSearch();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._roleBusy = false;
  }
}

function handleImported() {
  pagination.currentPage = 1;
  onSearch();
}

onMounted(onSearch);
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="onSearch">
      <template #title>
        <div class="flex items-center">
          <span>用户管理</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ pagination.total }} 名用户
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button type="primary" :icon="Plus" @click="importVisible = true">
          导入用户
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <el-input
            v-model="keyword"
            class="w-64"
            clearable
            placeholder="用户号、用户名、学号、姓名或邮箱"
            @keyup.enter="applyFilter"
            @clear="applyFilter"
          />
          <el-select
            v-model="role"
            class="w-36"
            clearable
            placeholder="全部角色"
            @change="applyFilter"
          >
            <el-option
              v-for="option in ROLE_OPTIONS"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
          <el-select
            v-model="status"
            class="w-32"
            clearable
            placeholder="全部状态"
            @change="applyFilter"
          >
            <el-option label="正常" :value="0" />
            <el-option label="封禁" :value="1" />
          </el-select>
          <el-button type="primary" @click="applyFilter">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </div>
        <PureTable
          border
          adaptive
          table-layout="auto"
          align-whole="center"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="pagination"
          @page-size-change="onSizeChange"
          @page-current-change="onCurrentChange"
        >
          <template #studentNo="{ row }">{{ row.studentNo || "—" }}</template>
          <template #email="{ row }">{{ row.email || "—" }}</template>
          <template #role="{ row }">
            <el-tooltip
              content="不能修改自己的角色"
              placement="top"
              :disabled="row.id !== currentUserId"
            >
              <span>
                <el-select
                  :model-value="row.role"
                  size="small"
                  style="width: 110px"
                  :disabled="row._roleBusy || row.id === currentUserId"
                  @change="changeRole(row, $event)"
                >
                  <el-option
                    v-for="opt in ROLE_OPTIONS"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                  />
                </el-select>
              </span>
            </el-tooltip>
          </template>
          <template #status="{ row }">
            <el-tag :type="STATUS_TAG_TYPES[row.status]" size="small">
              {{ STATUS_LABELS[row.status] }}
            </el-tag>
          </template>
          <template #createdAt="{ row }">{{
            formatTime(row.createdAt)
          }}</template>
          <template #operation="{ row }">
            <el-button
              v-if="row.role !== 1"
              link
              size="small"
              :type="row.status === 1 ? 'success' : 'danger'"
              :loading="row._statusBusy"
              @click="toggleBan(row)"
            >
              {{ row.status === 1 ? "解封" : "封禁" }}
            </el-button>
            <span v-else class="text-gray-400">—</span>
          </template>
        </PureTable>
      </template>
    </PureTableBar>

    <UserImportDialog v-model="importVisible" @imported="handleImported" />
  </div>
</template>
