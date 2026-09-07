<template>
  <div class="f-panel">
    <!-- 未开始且无权提前看 -->
    <div v-if="hw.locked" class="py-16 text-center">
      <el-icon class="text-gray-300" :size="34">
        <Lock />
      </el-icon>
      <h3 class="mt-3 text-base font-medium text-gray-700">作业尚未开始</h3>
      <p class="mt-1 text-[13px] text-gray-400">{{ hw.startTime }} 开始后可见题目</p>
    </div>

    <el-table v-else :data="hw.problems" style="width: 100%">
      <el-table-column label="状态" width="70" align="center">
        <template #default="{ row }">
          <!-- 1=已通过；其他非空=交过没过；null=没做（与题库/题单一致） -->
          <el-icon v-if="row.status === 1" color="#2f9e44" :size="17">
            <Select />
          </el-icon>
          <el-icon v-else-if="row.status != null" color="#e5484d" :size="17">
            <CloseBold />
          </el-icon>
        </template>
      </el-table-column>
      <el-table-column prop="id" label="题号" width="100" align="center" />
      <el-table-column label="题目" min-width="240">
        <template #default="{ row }">
          <router-link class="font-medium text-blue-500 hover:text-blue-400"
            :to="{ name: 'HomeworkProblemDetail', params: { id: hw.teamId, hid: hw.id, problemId: row.id } }">
            {{ row.name }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="难度" width="90" align="center">
        <template #default="{ row }">
          <span class="text-[14px] font-medium" :class="diffClass(row.difficulty)">{{ diffText(row.difficulty) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="我的得分" width="110" align="center">
        <template #default="{ row }">
          <span v-if="row.myScore != null" class="font-semibold tabular-nums"
            :class="row.myScore >= 100 ? 'text-emerald-600' : row.myScore > 0 ? 'text-amber-600' : 'text-gray-500'">
            {{ row.myScore }}
          </span>
          <span v-else class="text-gray-400">—</span>
        </template>
      </el-table-column>
      <el-table-column label="作业通过率" width="140" align="center">
        <template #default="{ row }">
          <span v-if="row.submitCount" class="tabular-nums text-gray-700">{{ rate(row) }}%</span>
          <span v-else class="text-gray-400">-</span>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="该作业还没有题目" :image-size="80" />
      </template>
    </el-table>
  </div>
</template>

<script setup>
import { inject } from "vue";
import { Lock, Select, CloseBold } from "@element-plus/icons-vue";

const hw = inject("homework");

const diffText = (d) => ({ 1: "简单", 2: "中等", 3: "困难" }[d] || "-");
const diffClass = (d) => ({ 1: "text-emerald-600", 2: "text-amber-500", 3: "text-red-500" }[d] || "text-gray-400");
const rate = (row) => (row.submitCount ? Math.floor((100 * row.acceptedCount) / row.submitCount) : 0);
</script>
