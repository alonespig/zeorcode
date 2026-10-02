<script setup>
defineProps({
  members: { type: Array, required: true },
  canManage: Boolean,
  canOwn: Boolean,
  emptyText: { type: String, default: "暂无成员" },
});

const emit = defineEmits(["remove"]);
</script>

<template>
  <el-table :data="members" row-key="uid" class="member-table" aria-label="团队成员列表">
    <el-table-column label="账号" min-width="260">
      <template #default="{ row }">
        <div class="member-account">
          <el-avatar :size="34" :src="row.avatar || undefined" class="member-avatar">
            {{ (row.realName || row.username || '?').slice(0, 1) }}
          </el-avatar>
          <span class="member-username" :title="row.username">{{ row.username }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="姓名" min-width="140" show-overflow-tooltip>
      <template #default="{ row }">{{ row.realName || '—' }}</template>
    </el-table-column>
    <el-table-column label="性别" min-width="80" align="center">
      <template #default="{ row }">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '未设置' }}</template>
    </el-table-column>
    <el-table-column label="学号" min-width="190" show-overflow-tooltip>
      <template #default="{ row }"><span class="member-number">{{ row.studentNo || '—' }}</span></template>
    </el-table-column>
    <el-table-column label="加入时间" min-width="150" align="center">
      <template #default="{ row }"><span class="member-number">{{ row.joinedAt?.slice(0, 10) || '—' }}</span></template>
    </el-table-column>
    <el-table-column v-if="canManage" label="操作" min-width="80" align="center">
      <template #default="{ row }">
        <el-button v-if="row.role !== 2 && (canOwn || row.role === 0)" link type="danger"
          @click="emit('remove', row)">移出</el-button>
        <span v-else class="member-unavailable">—</span>
      </template>
    </el-table-column>
    <template #empty>
      <el-empty :description="emptyText" :image-size="80" />
    </template>
  </el-table>
</template>

<style scoped>
.member-table {
  width: 100%;
}

.member-table :deep(.el-table__cell) {
  padding-top: 14px;
  padding-bottom: 14px;
}

.member-table :deep(.el-table__cell:first-child .cell) {
  padding-left: 24px;
}

.member-account {
  display: flex;
  align-items: center;
  gap: 12px;
}

.member-avatar {
  flex: 0 0 auto;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.member-username {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--el-text-color-primary);
  font-weight: 500;
}

.member-number {
  font-variant-numeric: tabular-nums;
}

.member-unavailable {
  color: var(--el-text-color-placeholder);
}
</style>
