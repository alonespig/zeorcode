<script setup>
defineProps({
  tags: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

defineEmits(["edit", "delete"]);
</script>

<template>
  <el-table v-loading="loading" :data="tags" stripe>
    <el-table-column prop="id" label="ID" width="100" />
    <el-table-column prop="name" label="标签名称" min-width="280" show-overflow-tooltip>
      <template #default="{ row }">
        <el-tag effect="plain" size="large">{{ row.name }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="problemCount" label="关联题目" min-width="140" />
    <el-table-column prop="problemSetCount" label="关联题单" min-width="140" />
    <el-table-column label="操作" width="150" fixed="right">
      <template #default="{ row }">
        <el-button type="primary" size="small" link @click="$emit('edit', row)">编辑</el-button>
        <el-button type="danger" size="small" link @click="$emit('delete', row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>
