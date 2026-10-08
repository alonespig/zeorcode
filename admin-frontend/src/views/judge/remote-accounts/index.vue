<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import { ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  createRemoteAccount,
  deleteRemoteAccount,
  getRemoteAccounts,
  getRemoteOJList,
  updateRemoteAccount,
  type CreateRemoteAccountParams,
  type RemoteAccountItem,
  type RemoteAuthType,
  type UpdateRemoteAccountParams
} from "@/api/admin/remote-accounts";

defineOptions({ name: "AdminRemoteAccounts" });

type RemoteAccountRow = RemoteAccountItem & { _busy?: boolean };

interface AccountForm {
  oj: string;
  authType: RemoteAuthType;
  username: string;
  secret: string;
  enabled: boolean;
}

const AUTH_LABELS: Record<RemoteAuthType, string> = {
  password: "账号密码",
  cookie: "Cookie / Token"
};

const dataList = ref<RemoteAccountRow[]>([]);
const ojOptions = ref<string[]>([]);
const loading = ref(false);
const dialogVisible = ref(false);
const submitting = ref(false);
const editingId = ref<number>();
const formRef = ref<FormInstance>();
const form = reactive<AccountForm>(blankForm());

const dialogTitle = computed(() =>
  editingId.value === undefined ? "新建远程账号" : "编辑远程账号"
);

const rules: FormRules = {
  oj: [{ required: true, message: "请选择 OJ", trigger: "change" }],
  authType: [{ required: true, message: "请选择认证方式", trigger: "change" }]
};

const columns: TableColumnList = [
  { label: "OJ", prop: "oj", width: 130, align: "center" },
  {
    label: "账号",
    slot: "username",
    minWidth: 170,
    showOverflowTooltip: true
  },
  { label: "认证方式", slot: "authType", width: 130, align: "center" },
  { label: "凭证", slot: "secret", width: 100, align: "center" },
  { label: "有效性", slot: "valid", width: 100, align: "center" },
  { label: "占用状态", slot: "busy", width: 110, align: "center" },
  { label: "启用", slot: "enabled", width: 100, align: "center" },
  { label: "创建时间", slot: "createdAt", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 140,
    fixed: "right",
    align: "center"
  }
];

function blankForm(): AccountForm {
  return {
    oj: "",
    authType: "cookie",
    username: "",
    secret: "",
    enabled: true
  };
}

function fillForm(value: AccountForm) {
  form.oj = value.oj;
  form.authType = value.authType;
  form.username = value.username;
  form.secret = value.secret;
  form.enabled = value.enabled;
}

function formatCreatedAt(seconds: number) {
  if (!seconds) return "—";
  return new Date(seconds * 1000).toLocaleString("zh-CN", { hour12: false });
}

async function loadAccounts() {
  loading.value = true;
  try {
    const res = await getRemoteAccounts();
    dataList.value = res.data?.list ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

async function loadOJOptions() {
  try {
    const res = await getRemoteOJList();
    ojOptions.value = res.data?.list ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  }
}

function openCreate() {
  editingId.value = undefined;
  fillForm(blankForm());
  dialogVisible.value = true;
  nextTick(() => formRef.value?.clearValidate());
}

function openEdit(row: RemoteAccountRow) {
  editingId.value = row.id;
  fillForm({
    oj: row.oj,
    authType: row.authType,
    username: row.username,
    secret: "",
    enabled: row.enabled
  });
  dialogVisible.value = true;
  nextTick(() => formRef.value?.clearValidate());
}

async function save() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid || submitting.value) return;

  const common: UpdateRemoteAccountParams = {
    authType: form.authType,
    username: form.username.trim(),
    secret: form.secret,
    enabled: form.enabled
  };
  submitting.value = true;
  try {
    if (editingId.value === undefined) {
      const data: CreateRemoteAccountParams = { oj: form.oj, ...common };
      await createRemoteAccount(data);
      message("远程账号已创建", { type: "success" });
    } else {
      await updateRemoteAccount(editingId.value, common);
      message("远程账号已更新", { type: "success" });
    }
    dialogVisible.value = false;
    await loadAccounts();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    submitting.value = false;
  }
}

async function toggleEnabled(row: RemoteAccountRow, enabled: boolean) {
  if (row._busy) return;
  const previous = row.enabled;
  row.enabled = enabled;
  row._busy = true;
  try {
    await updateRemoteAccount(row.id, {
      authType: row.authType,
      username: row.username,
      secret: "",
      enabled
    });
    message(enabled ? "账号已启用" : "账号已停用", { type: "success" });
  } catch {
    row.enabled = previous;
  } finally {
    row._busy = false;
  }
}

async function remove(row: RemoteAccountRow) {
  try {
    await ElMessageBox.confirm(
      `确认删除 ${row.oj} 的账号${row.username ? `「${row.username}」` : ""}？删除后不可恢复。`,
      "删除远程账号",
      { type: "warning" }
    );
  } catch {
    return;
  }

  if (row._busy) return;
  row._busy = true;
  try {
    await deleteRemoteAccount(row.id);
    message("远程账号已删除", { type: "success" });
    await loadAccounts();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

onMounted(() => {
  loadAccounts();
  loadOJOptions();
});
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="loadAccounts">
      <template #title>
        <div class="flex items-center">
          <span>远程账号</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ dataList.length }} 个
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button type="primary" :icon="Plus" @click="openCreate">
          新建账号
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <PureTable
          border
          adaptive
          table-layout="auto"
          align-whole="center"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
        >
          <template #username="{ row }">
            {{ row.username || "—" }}
          </template>
          <template #authType="{ row }">
            {{ AUTH_LABELS[row.authType] ?? row.authType }}
          </template>
          <template #secret="{ row }">
            <el-tag :type="row.hasSecret ? 'success' : 'info'" size="small">
              {{ row.hasSecret ? "已设置" : "未设置" }}
            </el-tag>
          </template>
          <template #valid="{ row }">
            <el-tag :type="row.valid ? 'success' : 'info'" size="small">
              {{ row.valid ? "有效" : "未验证" }}
            </el-tag>
          </template>
          <template #busy="{ row }">
            <el-tag :type="row.busy ? 'warning' : 'info'" size="small">
              {{ row.busy ? "占用中" : "空闲" }}
            </el-tag>
          </template>
          <template #enabled="{ row }">
            <el-switch
              :model-value="row.enabled"
              :loading="row._busy"
              @change="value => toggleEnabled(row, Boolean(value))"
            />
          </template>
          <template #createdAt="{ row }">
            {{ formatCreatedAt(row.created_at) }}
          </template>
          <template #operation="{ row }">
            <el-button link size="small" type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button
              link
              size="small"
              type="danger"
              :loading="row._busy"
              @click="remove(row)"
            >
              删除
            </el-button>
          </template>
        </PureTable>
        <p class="mt-3 text-xs text-gray-400">
          凭证由后端加密保存，页面不会返回或展示明文；编辑时留空可保留原凭证。
        </p>
      </template>
    </PureTableBar>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      destroy-on-close
      :close-on-click-modal="false"
      @closed="formRef?.resetFields()"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="88px">
        <el-form-item label="OJ" prop="oj">
          <el-select
            v-model="form.oj"
            filterable
            placeholder="请选择 OJ"
            class="w-full"
            :disabled="editingId !== undefined"
          >
            <el-option
              v-for="oj in ojOptions"
              :key="oj"
              :label="oj"
              :value="oj"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="认证方式" prop="authType">
          <el-radio-group v-model="form.authType">
            <el-radio value="password">账号密码</el-radio>
            <el-radio value="cookie">Cookie / Token</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.authType === 'password'" label="用户名">
          <el-input v-model="form.username" placeholder="登录用户名" />
        </el-form-item>
        <el-form-item :label="form.authType === 'password' ? '密码' : 'Cookie'">
          <el-input
            v-model="form.secret"
            :type="form.authType === 'password' ? 'password' : 'textarea'"
            :rows="form.authType === 'cookie' ? 4 : undefined"
            :show-password="form.authType === 'password'"
            :placeholder="
              editingId === undefined
                ? form.authType === 'password'
                  ? '登录密码'
                  : 'Cookie / Token'
                : '留空表示不修改原凭证'
            "
          />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="save">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
