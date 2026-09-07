<template>
  <div class="user-ranking f-card">
    <div class="f-card-head f-blue">
      <el-icon>
        <Histogram />
      </el-icon>
      <span>用户排行</span>
    </div>

    <ul class="mt-1">
      <li v-for="(item, index) in rankList" :key="index"
        class="flex items-center  gap-2 px-6 py-2.5 border-b border-gray-100 last:border-b-0 hover:bg-slate-50">
        <!-- 名次：前三金银铜，其余灰数字 -->
        <span class="flex h-5.5 w-5.5 shrink-0 items-center justify-center rounded font-sans text-xs font-bold"
          :class="rankClass(index)">
          {{ index + 1 }}
        </span>

        <div class="ml-20 flex items-center">
          <el-avatar :size="28" :src="item.avatar" class="shrink-0 text-xs">
            {{ item.username?.charAt(0)?.toUpperCase() }}
          </el-avatar>
          <span @click="goUser(item.id)"
          class="ml-2 max-w-36 truncate text-[15px] text-gray-700
          cursor-pointer hover:text-blue-500
          ">
            {{ item.username }}
          </span>
        </div>

        <span class="ml-auto shrink-0 text-[15px] text-gray-600 tabular-nums">
          {{ item.count }}
          <i class="not-italic font-sans text-xs text-gray-500">题</i>
        </span>
      </li>
      <li v-if="!rankList.length" class="py-4 text-center text-sm text-gray-400">暂无排名</li>
    </ul>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { Histogram } from '@element-plus/icons-vue'
import { id } from 'element-plus/es/locale/index.mjs';

defineProps({
  rankList: {
    type: Array,
    default: () => [],
  },
})

const router = useRouter();

const goUser = (id) => {
  router.push({
    name: 'User',
    params: {
      id
    }
  })
}

// 前三金 / 银 / 铜，其余灰色数字（无底色）
const rankClass = (index) => {
  if (index === 0) return 'bg-[#f2b01e] text-white'
  if (index === 1) return 'bg-[#b6bcc6] text-white'
  if (index === 2) return 'bg-[#cd8b52] text-white'
  return 'text-gray-400'
}
</script>

<style scoped lang="scss">
.user-ranking {
  width: 100%;
  background-color: #fff;

  .f-card-head {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 19px;
  }
}
</style>
