<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="t">远程账号</span>
          <span class="sub">配置各 OJ 的提交账号 / Cookie / Token，供远程拉题与判题复用；凭证 AES 加密存储</span>
        </div>
        <el-button type="primary" @click="openCreate">
          <el-icon class="mr-1">
            <Plus />
          </el-icon>新建账号
        </el-button>
      </div>
    </template>

    <el-table v-loading="loading" :data="accounts" stripe>
      <el-table-column prop="oj" label="OJ" width="120" show-overflow-tooltip />
      <el-table-column label="账号" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">{{ row.username || "—" }}</template>
      </el-table-column>
      <el-table-column label="认证方式" width="120">
        <template #default="{ row }">{{ AUTH_LABEL[row.authType] || row.authType }}</template>
      </el-table-column>
      <el-table-column label="凭证" width="100">
        <template #default="{ row }">
          <el-tag :type="row.hasSecret ? 'success' : 'info'" size="small" effect="plain">
            {{ row.hasSecret ? '已设置' : '未设置' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <span class="dot" :class="row.valid ? 'ok' : 'bad'"></span>{{ row.valid ? "有效" : "未验证" }}
        </template>
      </el-table-column>
      <el-table-column label="占用" width="90">
        <template #default="{ row }">
          <el-tag :type="row.busy ? 'warning' : 'info'" size="small" effect="plain">
            {{ row.busy ? '占用中' : '空闲' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="启用" width="80">
        <template #default="{ row }">
          <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="130" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑账号' : '新建账号'" width="480px">
      <el-form :model="form" label-width="84px">
        <el-form-item label="OJ">
          <el-select v-model="form.oj" placeholder="选择 OJ" :disabled="!!editingId" style="width: 100%">
            <el-option v-for="oj in ojOptions" :key="oj" :label="oj" :value="oj" />
          </el-select>
        </el-form-item>
        <el-form-item label="认证方式">
          <el-radio-group v-model="form.authType">
            <el-radio value="password">账号密码</el-radio>
            <el-radio value="cookie">Cookie / Token</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="form.authType === 'password'">
          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="登录用户名" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password
              :placeholder="editingId ? '留空表示不修改' : '登录密码'" />
          </el-form-item>
        </template>

        <el-form-item v-else label="Cookie">
          <el-input v-model="form.cookie" type="textarea" :rows="4"
            :placeholder="editingId ? '留空表示不修改' : '粘贴浏览器登录后的 Cookie / Token'" />
        </el-form-item>

        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import {
  getRemoteAccounts,
  createRemoteAccount,
  updateRemoteAccount,
  deleteRemoteAccount,
} from "@/api/admin";
import { getRemoteOJList } from "@/api/remote";

const AUTH_LABEL = { password: "账号密码", cookie: "Cookie/Token" };

const loading = ref(false);
const saving = ref(false);
const accounts = ref([]);
const ojOptions = ref(["HDU", "POJ", "Codeforces", "LibreOJ", "Nowcoder"]);

const dialogVisible = ref(false);
const editingId = ref(null);
const form = ref(blankForm());

function blankForm() {
  return { oj: "", authType: "cookie", username: "", password: "", cookie: "", enabled: true };
}

async function loadAccounts() {
  loading.value = true;
  try {
    const res = await getRemoteAccounts();
    accounts.value = res.data?.list || [];
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
}

async function loadOJ() {
  try {
    const res = await getRemoteOJList();
    if (res.data?.list?.length) ojOptions.value = res.data.list;
  } catch {
    // 拉不到就用默认列表
  }
}

function openCreate() {
  editingId.value = null;
  form.value = blankForm();
  dialogVisible.value = true;
}

function openEdit(row) {
  editingId.value = row.id;
  form.value = {
    oj: row.oj,
    authType: row.authType,
    username: row.username || "",
    password: "",
    cookie: "",
    enabled: row.enabled,
  };
  dialogVisible.value = true;
}

function buildPayload() {
  const secret = form.value.authType === "password" ? form.value.password : form.value.cookie;
  return {
    oj: form.value.oj,
    authType: form.value.authType,
    username: form.value.username,
    secret,
    enabled: form.value.enabled,
  };
}

async function save() {
  if (!form.value.oj) {
    ElMessage.warning("请选择 OJ");
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await updateRemoteAccount(editingId.value, buildPayload());
    } else {
      await createRemoteAccount(buildPayload());
    }
    ElMessage.success("已保存");
    dialogVisible.value = false;
    loadAccounts();
  } catch (err) {
    console.error(err);
  } finally {
    saving.value = false;
  }
}

async function toggleEnabled(row) {
  try {
    await updateRemoteAccount(row.id, {
      authType: row.authType,
      username: row.username,
      enabled: row.enabled,
      secret: "",
    });
  } catch (err) {
    row.enabled = !row.enabled; // 失败回滚
    console.error(err);
  }
}

function remove(row) {
  ElMessageBox.confirm(`确定删除 ${row.oj} 的账号？`, "提示", { type: "warning" })
    .then(async () => {
      await deleteRemoteAccount(row.id);
      ElMessage.success("已删除");
      loadAccounts();
    })
    .catch(() => { });
}

onMounted(() => {
  loadAccounts();
  loadOJ();
});
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
  font-size: 12px;
  color: #909399;
}

.dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  margin-right: 5px;
  vertical-align: middle;
}

.dot.ok {
  background: #67c23a;
}

.dot.bad {
  background: #c0c4cc;
}

.mr-1 {
  margin-right: 4px;
}
</style>
