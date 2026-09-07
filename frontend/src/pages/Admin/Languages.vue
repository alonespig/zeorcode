<script setup>
import { computed, onMounted, reactive, ref, shallowRef } from "vue";
import dayjs from "dayjs";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import {
  createAdminLanguage,
  deleteAdminLanguage,
  getAdminLanguages,
  updateAdminLanguage,
} from "@/api/admin";
import { invalidateLanguageCache } from "@/hooks/useLanguages";

const languages = ref([]);
const loading = shallowRef(false);
const saving = shallowRef(false);
const dialogVisible = shallowRef(false);
const editingId = shallowRef(null);
const form = reactive({ name: "", status: 1, sort: 0 });

const dialogTitle = computed(() => editingId.value ? "编辑编程语言" : "新增编程语言");

const loadLanguages = async () => {
  loading.value = true;
  try {
    const res = await getAdminLanguages();
    languages.value = res.data || [];
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const resetForm = () => {
  form.name = "";
  form.status = 1;
  form.sort = languages.value.reduce((max, item) => Math.max(max, item.sort || 0), 0) + 1;
};

const openCreate = () => {
  editingId.value = null;
  resetForm();
  dialogVisible.value = true;
};

const openEdit = (row) => {
  editingId.value = row.id;
  form.name = row.name;
  form.status = row.status;
  form.sort = row.sort;
  dialogVisible.value = true;
};

const payload = () => ({
  name: form.name.trim(),
  status: form.status,
  sort: Number(form.sort) || 0,
});

const save = async () => {
  if (!form.name.trim()) {
    ElMessage.warning("请输入语言名称");
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await updateAdminLanguage(editingId.value, payload());
      ElMessage.success("编程语言已更新");
    } else {
      await createAdminLanguage(payload());
      ElMessage.success("编程语言已添加");
    }
    invalidateLanguageCache();
    dialogVisible.value = false;
    await loadLanguages();
  } catch (err) {
    console.error(err);
  } finally {
    saving.value = false;
  }
};

const toggleStatus = async (row) => {
  const nextStatus = row.status === 1 ? 0 : 1;
  try {
    await updateAdminLanguage(row.id, { name: row.name, status: nextStatus, sort: row.sort });
    row.status = nextStatus;
    invalidateLanguageCache();
    ElMessage.success(nextStatus === 1 ? "语言已启用" : "语言已停用");
  } catch (err) {
    console.error(err);
  }
};

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认删除编程语言「${row.name}」？已有提交记录不会被删除。`,
      "删除编程语言",
      { type: "warning", confirmButtonText: "确认删除" },
    );
  } catch {
    return;
  }
  try {
    await deleteAdminLanguage(row.id);
    invalidateLanguageCache();
    ElMessage.success("编程语言已删除");
    await loadLanguages();
  } catch (err) {
    console.error(err);
  }
};

const formatTime = (value) => value ? dayjs(value).format("YYYY-MM-DD HH:mm") : "—";

onMounted(loadLanguages);
</script>

<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="title">编程语言</span>
          <span class="summary">共 {{ languages.length }} 种语言</span>
        </div>
        <el-button type="primary" @click="openCreate">
          <el-icon class="button-icon"><Plus /></el-icon>
          新增语言
        </el-button>
      </div>
    </template>

    <el-table v-loading="loading" :data="languages" stripe>
      <el-table-column prop="name" label="名称" min-width="180" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-switch
            :model-value="row.status === 1"
            inline-prompt
            active-text="启用"
            inactive-text="停用"
            @change="toggleStatus(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="100" />
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="更新时间" width="180">
        <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" :close-on-click-modal="false">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" maxlength="32" show-word-limit placeholder="例如 c++、java、python" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        </el-form-item>
        <p class="form-hint">名称必须是评测程序已经支持的语言；未知名称会被后端拒绝。</p>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<style scoped>
.card-head,
.head-title {
  display: flex;
  align-items: center;
}

.card-head {
  justify-content: space-between;
  gap: 16px;
}

.title {
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.summary {
  margin-left: 10px;
  color: #909399;
  font-size: 13px;
}

.button-icon {
  margin-right: 4px;
}

.form-hint {
  margin: 2px 0 0 80px;
  color: #909399;
  font-size: 12px;
  line-height: 1.6;
}
</style>
