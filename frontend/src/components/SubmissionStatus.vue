<template>
  <table class="data-table submission-table">
    <colgroup>
      <col :style="{ width: showRejudge ? '7%' : '8%' }">
      <col :style="{ width: showRejudge ? '18%' : '20%' }">
      <col v-if="!hideUser" :style="{ width: showRejudge ? '12%' : '14%' }">
      <col style="width: 12%">
      <col style="width: 8%">
      <col style="width: 8%">
      <col :style="{ width: showRejudge ? '9%' : '10%' }">
      <col :style="{ width: showRejudge ? '15%' : '20%' }">
      <col v-if="showRejudge" style="width: 11%">
    </colgroup>
    <thead>
      <tr>
        <th class="center">#</th>
        <th class="center">题目</th>
        <th v-if="!hideUser" class="left">用户</th>
        <th class="center">结果</th>
        <th class="center">时间</th>
        <th class="center">内存</th>
        <th class="center">语言</th>
        <th class="center">提交时间</th>
        <th v-if="showRejudge" class="center">操作</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="item in tableData" :key="item.id">
        <td>
          <span class="cursor-pointer hover:text-blue-400" @click="handleClickId(item.id)">
            {{ item.id }}
          </span>
        </td>
        <td>
          <span class="cursor-pointer text-blue-400 font-medium" @click="handleClickProblemId(item.problemID)">
            {{ item.problemName }}
          </span>
        </td>
        <td v-if="!hideUser" class="left">
          <UserName :name="item.userName" :rating="item.rating" clickable @click="handleClickUser(item.userID)" />
        </td>
        <td>
          <FTag :result="item.result" />
        </td>
        <td>{{ formatTime(item.timeUsed) }}</td>
        <td>{{ formatMemory(item.memoryUsed) }}</td>
        <td>{{ item.language }}</td>
        <td>{{ item.createdAt }}</td>
        <td v-if="showRejudge">
          <el-button class="rejudge-btn"  type="primary" :icon="RefreshRight" :loading="item.rejudging"
            @click="$emit('rejudge', item)">重判</el-button>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
import FTag from "@/components/FTag.vue";
import UserName from "@/components/UserName.vue";
import { RefreshRight } from '@element-plus/icons-vue'
import { formatTime, formatMemory } from '@/utils/format'

defineProps({
  tableData: {
    type: Array,
    default: () => [],
  },
  // 隐藏"用户"列：个人主页提交记录里每行都是同一人，无需重复展示
  hideUser: {
    type: Boolean,
    default: false,
  },
  // 显示"操作"列（超管重判），默认关闭
  showRejudge: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['click-id', 'click-problem-id', 'click-user', 'rejudge'])

const handleClickId = (id) => {
  emit('click-id', id)
}

const handleClickProblemId = (id) => {
  emit('click-problem-id', id)
}

const handleClickUser = (id) => {
  emit('click-user', id)
}
</script>

<style scoped lang="scss">
.submission-table tbody {
  font-weight: 500;
  letter-spacing: 0.025em;
}

/* 重判按钮图标调大一点，和文字更协调 */
.rejudge-btn :deep(.el-icon) {
  font-size: 16px;
}
</style>
