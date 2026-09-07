<script setup>
import { computed, onMounted, ref, shallowRef } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Search } from "@element-plus/icons-vue";
import {
  createAdminTag,
  deleteAdminTag,
  getAdminTags,
  updateAdminTag,
} from "@/api/admin";
import AlgorithmTagDialog from "@/pages/Admin/components/AlgorithmTagDialog.vue";
import AlgorithmTagTable from "@/pages/Admin/components/AlgorithmTagTable.vue";

const tags = ref([]);
const loading = shallowRef(false);
const submitting = shallowRef(false);
const dialogVisible = shallowRef(false);
const keyword = shallowRef("");
const editingTag = shallowRef(null);

const filteredTags = computed(() => {
  const query = keyword.value.trim().toLocaleLowerCase();
  if (!query) return tags.value;
  return tags.value.filter((tag) => tag.name.toLocaleLowerCase().includes(query));
});
const dialogTitle = computed(() => editingTag.value ? "编辑算法标签" : "新增算法标签");
const initialName = computed(() => editingTag.value?.name || "");

const loadTags = async () => {
  loading.value = true;
  try {
    const res = await getAdminTags();
    tags.value = res.data?.tags || [];
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const openCreate = () => {
  editingTag.value = null;
  dialogVisible.value = true;
};

const openEdit = (tag) => {
  editingTag.value = tag;
  dialogVisible.value = true;
};

const saveTag = async (name) => {
  if (!name || submitting.value) return;
  submitting.value = true;
  try {
    if (editingTag.value) {
      await updateAdminTag(editingTag.value.id, { name });
      ElMessage.success("算法标签已更新");
    } else {
      await createAdminTag({ name });
      ElMessage.success("算法标签已新增");
    }
    dialogVisible.value = false;
    await loadTags();
  } catch (err) {
    console.error(err);
  } finally {
    submitting.value = false;
  }
};

const removeTag = async (tag) => {
  const usage = `当前关联 ${tag.problemCount} 道题目、${tag.problemSetCount} 个题单`;
  try {
    await ElMessageBox.confirm(
      `确认删除算法标签「${tag.name}」？${usage}，删除后这些关联也会一并移除。`,
      "删除算法标签",
      { type: "warning", confirmButtonText: "确认删除" },
    );
  } catch {
    return;
  }
  try {
    await deleteAdminTag(tag.id);
    ElMessage.success("算法标签已删除");
    await loadTags();
  } catch (err) {
    console.error(err);
  }
};

onMounted(loadTags);
</script>

<template>
  <el-card shadow="never" class="page-card">
    <template #header>
      <div class="card-head">
        <div class="head-title">
          <span class="title">算法标签</span>
          <span class="summary">共 {{ tags.length }} 个标签</span>
        </div>
        <div class="head-actions">
          <el-input v-model="keyword" :prefix-icon="Search" clearable placeholder="搜索标签…" class="tag-search" />
          <el-button type="primary" @click="openCreate">
            <el-icon class="button-icon"><Plus /></el-icon>
            新增标签
          </el-button>
        </div>
      </div>
    </template>

    <AlgorithmTagTable :tags="filteredTags" :loading="loading" @edit="openEdit" @delete="removeTag" />

    <AlgorithmTagDialog
      v-model="dialogVisible"
      :title="dialogTitle"
      :initial-name="initialName"
      :submitting="submitting"
      @submit="saveTag"
    />
  </el-card>
</template>

<style scoped>
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.head-title,
.head-actions {
  display: flex;
  align-items: center;
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

.head-actions {
  gap: 10px;
}

.tag-search {
  width: 240px;
}

.button-icon {
  margin-right: 4px;
}

@media (max-width: 767px) {
  .card-head,
  .head-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .tag-search {
    width: 100%;
  }
}
</style>
