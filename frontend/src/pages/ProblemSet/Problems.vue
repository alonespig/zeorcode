<script setup>
import { CloseBold, Select } from "@element-plus/icons-vue";
import { tagStyle } from "@/utils/tag";

defineProps({
  detail: { type: Object, required: true },
});

const DIFF_TEXT = { 1: "简单", 2: "中等", 3: "困难" };
const diffTextClass = (difficulty) =>
  ({ 1: "text-emerald-600", 2: "text-amber-500", 3: "text-red-500" })[difficulty] || "text-gray-400";
const diffColor = (difficulty) =>
  ({ 1: "#2f9e44", 2: "#e8930c", 3: "#e5484d" })[difficulty] || "#c0c4cc";
const rate = (problem) =>
  problem.submitCount ? Math.floor((100 * problem.acceptedCount) / problem.submitCount) : 0;
</script>

<template>
  <section class="f-panel">
    <div class="border-b border-gray-100 px-5 py-3.5">
      <h2 class="text-[15px] font-semibold text-gray-700">题目列表</h2>
    </div>
    <el-table :data="detail.problems" style="width: 100%">
      <el-table-column label="#" width="60" align="center">
        <template #default="{ $index }">
          <span class="text-[13px] text-gray-500">{{ $index + 1 }}</span>
        </template>
      </el-table-column>

      <el-table-column label="状态" width="70" align="center">
        <template #default="{ row }">
          <el-icon v-if="row.status === 1" color="#2f9e44" :size="17">
            <Select />
          </el-icon>
          <el-icon v-else-if="row.status != null" color="#e5484d" :size="17">
            <CloseBold />
          </el-icon>
        </template>
      </el-table-column>

      <el-table-column prop="id" label="题号" width="110" align="center" />
      <el-table-column label="题目" min-width="220">
        <template #default="{ row }">
          <router-link class="font-medium text-blue-500 hover:text-blue-400" :to="`/problem/${row.id}`">
            {{ row.name }}
          </router-link>
        </template>
      </el-table-column>

      <el-table-column label="难度" width="90" align="center">
        <template #default="{ row }">
          <span class="text-[14px] font-medium" :class="diffTextClass(row.difficulty)">
            {{ DIFF_TEXT[row.difficulty] || "-" }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="算法标签" min-width="180">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1.5">
            <el-tag v-for="tag in row.tags" :key="tag.id" effect="light" :style="tagStyle(tag)">
              {{ tag.name }}
            </el-tag>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="通过率" width="140" align="center">
        <template #default="{ row }">
          <div v-if="row.submitCount > 0" class="flex items-center gap-2">
            <div class="h-1.5 w-[74px] overflow-hidden rounded bg-gray-100">
              <div
                class="h-full rounded"
                :style="{ width: `${rate(row)}%`, backgroundColor: diffColor(row.difficulty) }"
              ></div>
            </div>
            <span class="w-9 text-[12.5px] tabular-nums text-gray-500">{{ rate(row) }}%</span>
          </div>
          <span v-else class="text-gray-300">-</span>
        </template>
      </el-table-column>

      <template #empty>
        <el-empty description="该题单还没有添加题目" :image-size="80" />
      </template>
    </el-table>
  </section>
</template>
