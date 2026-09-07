<template>
  <div class="problem-tag-selector">
    <div class="selector-row">
      <el-button @click="openDialog">
        选择标签
        <span v-if="modelValue.length" class="selected-count">{{ modelValue.length }}</span>
      </el-button>
      <div v-if="selectedTags.length" class="selected-tags">
        <el-tag
          v-for="tag in selectedTags"
          :key="tag.id"
          class="selected-tag"
          type="primary"
          effect="plain"
          closable
          @close="removeTag(tag.id)"
        >
          {{ tag.name }}
        </el-tag>
      </div>
      <span v-if="!selectedTags.length" class="empty-text">暂未选择标签</span>
    </div>

    <el-dialog v-model="dialogVisible" title="选择标签" width="640px" append-to-body>
      <el-input v-model="searchKeyword" placeholder="搜索标签…" clearable :prefix-icon="Search" />
      <div class="dialog-hint">全部标签（点击选择，可多选）</div>
      <div class="tag-options">
        <button
          v-for="tag in filteredTags"
          :key="tag.id"
          type="button"
          class="tag-option"
          :class="{ selected: tempSelected.includes(tag.id) }"
          @click="toggleTag(tag.id)"
        >
          {{ tag.name }}
        </button>
        <span v-if="!filteredTags.length" class="no-result">没有匹配的标签</span>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <span class="selected-summary">已选 {{ tempSelected.length }} 个</span>
          <button v-if="tempSelected.length" type="button" class="clear-button" @click="tempSelected = []">
            清空
          </button>
          <div class="dialog-actions">
            <el-button @click="dialogVisible = false">取消</el-button>
            <el-button type="primary" @click="confirmSelection">确定</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { Search } from "@element-plus/icons-vue";

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => [],
  },
  tags: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["update:modelValue"]);

const dialogVisible = ref(false);
const searchKeyword = ref("");
const tempSelected = ref([]);

const selectedTags = computed(() =>
  props.modelValue
    .map((id) => props.tags.find((tag) => tag.id === id))
    .filter(Boolean)
);

const filteredTags = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase();
  if (!keyword) return props.tags;
  return props.tags.filter((tag) => tag.name.toLowerCase().includes(keyword));
});

const openDialog = () => {
  searchKeyword.value = "";
  tempSelected.value = [...props.modelValue];
  dialogVisible.value = true;
};

const toggleTag = (id) => {
  const index = tempSelected.value.indexOf(id);
  if (index === -1) tempSelected.value.push(id);
  else tempSelected.value.splice(index, 1);
};

const removeTag = (id) => {
  emit("update:modelValue", props.modelValue.filter((tagId) => tagId !== id));
};

const confirmSelection = () => {
  emit("update:modelValue", [...tempSelected.value]);
  dialogVisible.value = false;
};
</script>

<style scoped>
.selector-row,
.dialog-footer {
  display: flex;
  align-items: center;
}

.selector-row {
  flex-wrap: wrap;
  gap: 10px;
}

.selected-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  margin-left: 6px;
  padding: 0 5px;
  border-radius: 9px;
  color: #fff;
  background: var(--color-primary);
  font-size: 11px;
}

.empty-text,
.selected-summary {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.selected-tag {
  --el-tag-bg-color: var(--el-color-primary-light-9);
  --el-tag-border-color: var(--el-color-primary-light-7);
  --el-tag-text-color: var(--color-primary);
}

.clear-button {
  color: var(--el-text-color-secondary);
  line-height: 1;
}

.dialog-hint {
  margin: 14px 0 8px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.tag-options {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  max-height: 240px;
  padding: 4px 0;
  overflow-y: auto;
}

.tag-option {
  padding: 5px 12px;
  border: 1px solid var(--el-border-color);
  border-radius: var(--radius-base);
  color: var(--el-text-color-regular);
  background: var(--color-surface);
  font-size: 13px;
  transition: border-color 0.2s, color 0.2s, background-color 0.2s;
}

.tag-option:hover {
  border-color: var(--el-color-primary-light-5);
  color: var(--color-primary);
}

.tag-option.selected {
  border-color: var(--color-primary);
  color: #fff;
  background: var(--color-primary);
}

.no-result {
  padding: 12px 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.dialog-footer {
  width: 100%;
}

.clear-button {
  margin-left: 12px;
}

.clear-button:hover {
  color: var(--color-danger);
}

.dialog-actions {
  display: flex;
  gap: 10px;
  margin-left: auto;
}
</style>
