<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { message } from "@/utils/message";
import {
  getPermissionCatalog,
  getUserPermissions,
  setUserPermissions,
  type PermissionGroup
} from "@/api/admin/permissions";

defineOptions({ name: "PermissionDialog" });

const props = defineProps<{
  userId: number | string;
  username: string;
}>();

const emit = defineEmits<{ saved: [] }>();

const visible = defineModel<boolean>({ default: false });

const groups = ref<PermissionGroup[]>([]);
const selected = reactive(new Set<string>());
const loading = ref(false);
const loaded = ref(false);
const failed = ref(false);
const saving = ref(false);
const search = ref("");
const template = ref("");

const TEMPLATES: Record<string, string[]> = {
  content: [
    "dashboard.read",
    "problem.read",
    "problem.write",
    "problem.archive",
    "problemset.read",
    "problemset.write",
    "problemset.delete",
    "contest.read",
    "contest.write",
    "contest.rejudge",
    "tag.read",
    "tag.write",
    "post.review"
  ],
  judge: [
    "dashboard.read",
    "submission.read",
    "submission.rejudge",
    "judge.node.read",
    "judge.node.manage",
    "judge.language.read",
    "judge.language.manage",
    "judge.remote.read",
    "judge.remote.manage"
  ],
  user: [
    "dashboard.read",
    "user.read",
    "user.manage",
    "user.import",
    "notification.send"
  ],
  audit: ["dashboard.read", "audit.read", "submission.read"]
};

const filteredGroups = computed(() => {
  const keyword = search.value.trim().toLowerCase();
  if (!keyword) return groups.value;
  return groups.value
    .map(group => {
      const permissions = group.permissions.filter(
        item =>
          item.label.toLowerCase().includes(keyword) ||
          item.key.toLowerCase().includes(keyword)
      );
      return { ...group, permissions };
    })
    .filter(group => group.permissions.length > 0);
});

const totalSelected = computed(() => selected.size);

function isSelected(key: string) {
  return selected.has(key);
}

function groupChecked(group: PermissionGroup) {
  return (
    group.permissions.length > 0 &&
    group.permissions.every(item => selected.has(item.key))
  );
}

function groupIndeterminate(group: PermissionGroup) {
  const count = group.permissions.filter(item => selected.has(item.key)).length;
  return count > 0 && count < group.permissions.length;
}

function toggle(key: string, checked: boolean) {
  // 工作台是管理员的默认权限，不允许取消。
  if (key === "dashboard.read") {
    selected.add(key);
    return;
  }
  if (checked) selected.add(key);
  else selected.delete(key);
}

function toggleGroup(group: PermissionGroup, checked: boolean) {
  group.permissions.forEach(item => toggle(item.key, checked));
}

function selectAll() {
  groups.value.forEach(group =>
    group.permissions.forEach(item => selected.add(item.key))
  );
}

function clearAll() {
  selected.clear();
  selected.add("dashboard.read");
}

function applyTemplate(value: string) {
  const list = TEMPLATES[value];
  if (!list) return;
  selected.clear();
  list.forEach(key => selected.add(key));
  template.value = "";
}

async function load() {
  if (!visible.value || !props.userId) return;
  loading.value = true;
  loaded.value = false;
  failed.value = false;
  search.value = "";
  template.value = "";
  selected.clear();
  try {
    const [catalogRes, permissionRes] = await Promise.all([
      getPermissionCatalog(),
      getUserPermissions(props.userId)
    ]);
    groups.value = catalogRes.data?.groups ?? [];
    permissionRes.data?.permissions?.forEach(key => selected.add(key));
    loaded.value = true;
  } catch {
    groups.value = [];
    failed.value = true;
  } finally {
    loading.value = false;
  }
}

async function save() {
  // 未成功加载时禁止保存，避免用空集合覆盖现有权限。
  if (!loaded.value || loading.value || failed.value || saving.value) return;
  saving.value = true;
  try {
    await setUserPermissions(props.userId, [...selected]);
    message("权限已保存，目标用户刷新页面后生效", { type: "success" });
    emit("saved");
    visible.value = false;
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    saving.value = false;
  }
}

watch([visible, () => props.userId], load, { immediate: true });
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="`权限配置 · ${username}`"
    width="min(780px, calc(100vw - 32px))"
    top="6vh"
    append-to-body
    destroy-on-close
    class="permission-dialog"
  >
    <div class="permission-toolbar">
      <el-input
        v-model="search"
        class="w-56!"
        placeholder="搜索权限名称…"
        clearable
      />
      <span class="permission-count">
        已选 <b>{{ totalSelected }}</b> 项
      </span>
      <div class="ml-auto flex items-center gap-3">
        <el-button link type="primary" @click="selectAll">全选</el-button>
        <el-button link @click="clearAll">清空</el-button>
        <el-select
          v-model="template"
          class="w-32!"
          placeholder="常用模板"
          @change="applyTemplate"
        >
          <el-option label="内容管理" value="content" />
          <el-option label="评测管理" value="judge" />
          <el-option label="用户管理" value="user" />
          <el-option label="只读审计" value="audit" />
        </el-select>
      </div>
    </div>

    <div v-loading="loading" class="permission-body">
      <div v-if="filteredGroups.length" class="permission-groups">
        <section
          v-for="group in filteredGroups"
          :key="group.key"
          class="permission-group"
        >
          <div class="group-head">
            <span class="group-name">{{ group.label }}</span>
            <el-checkbox
              :model-value="groupChecked(group)"
              :indeterminate="groupIndeterminate(group)"
              @change="value => toggleGroup(group, Boolean(value))"
            >
              全选
            </el-checkbox>
          </div>
          <div class="group-items">
            <el-checkbox
              v-for="item in group.permissions"
              :key="item.key"
              :model-value="isSelected(item.key)"
              :disabled="item.key === 'dashboard.read'"
              :title="
                item.key === 'dashboard.read'
                  ? '管理员默认拥有工作台权限'
                  : undefined
              "
              @change="value => toggle(item.key, Boolean(value))"
            >
              {{ item.label }}
            </el-checkbox>
          </div>
        </section>
      </div>
      <el-empty
        v-else-if="!loading && failed"
        :image-size="72"
        description="权限加载失败"
      >
        <el-button type="primary" @click="load">重新加载</el-button>
      </el-empty>
      <el-empty
        v-else-if="!loading"
        :image-size="72"
        description="没有匹配的权限"
      />
    </div>

    <template #footer>
      <div class="permission-footer">
        <span class="permission-count">
          已选 <b>{{ totalSelected }}</b> 项
        </span>
        <div class="flex gap-2.5">
          <el-button @click="visible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="saving"
            :disabled="!loaded || loading"
            @click="save"
          >
            保存
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.permission-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.permission-count {
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.permission-count b {
  color: var(--el-color-primary);
  font-weight: 600;
}

.permission-body {
  max-height: 58vh;
  min-height: 200px;
  margin-top: 12px;
  overflow-y: auto;
}

.permission-groups {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.permission-group {
  padding: 11px 13px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}

.group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.group-name {
  color: var(--el-text-color-primary);
  font-size: 13.5px;
  font-weight: 600;
}

.group-items {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px 10px;
  margin-top: 8px;
}

.permission-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

@media (max-width: 720px) {
  .permission-groups {
    grid-template-columns: 1fr;
  }
}
</style>
