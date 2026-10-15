<script setup>
const props = defineProps({
  currentPage: { type: Number, default: 1 },
  pageSize: { type: Number, default: 20 },
  total: { type: Number, default: 0 },
  pagerCount: { type: Number, default: 9 },
  layout: { type: String, default: 'prev, pager, next' },
  compact: Boolean,
  background: { type: Boolean, default: true },
})

const emit = defineEmits(['update:currentPage', 'current-change'])

const handleCurrentChange = (value) => {
  emit('update:currentPage', value)
  emit('current-change', value)
}
</script>

<template>
  <el-pagination
    class="app-pagination"
    :class="{ 'app-pagination--compact': props.compact }"
    :current-page="props.currentPage"
    :page-size="props.pageSize"
    :total="props.total"
    :pager-count="props.pagerCount"
    :layout="props.layout"
    :size="props.compact ? 'small' : 'default'"
    :background="props.background"
    @current-change="handleCurrentChange"
  />
</template>

<style scoped>
.app-pagination {
  --el-pagination-button-bg-color: #f1f3f5;
  --el-pagination-button-disabled-bg-color: #f6f7f8;
  --el-pagination-button-width: 34px;
  --el-pagination-button-height: 34px;
  --el-pagination-border-radius: 2px;
  --el-pagination-item-gap: 0;
}

.app-pagination :deep(.btn-prev),
.app-pagination :deep(.btn-next),
.app-pagination :deep(.el-pager li) {
  margin: 0 4px;
  font-size: 14px;
  font-weight: 400;
}

.app-pagination :deep(.el-pager li.is-active) {
  font-weight: 500;
}

.app-pagination--compact {
  --el-pagination-button-width-small: 28px;
  --el-pagination-button-height-small: 28px;
}

.app-pagination--compact :deep(.btn-prev),
.app-pagination--compact :deep(.btn-next),
.app-pagination--compact :deep(.el-pager li) {
  margin: 0 2px;
  font-size: 13px;
}

@media (max-width: 640px) {
  .app-pagination {
    --el-pagination-button-width: 30px;
    --el-pagination-button-height: 30px;
  }

  .app-pagination :deep(.btn-prev),
  .app-pagination :deep(.btn-next),
  .app-pagination :deep(.el-pager li) {
    margin: 0 2px;
    font-size: 13px;
  }
}
</style>
