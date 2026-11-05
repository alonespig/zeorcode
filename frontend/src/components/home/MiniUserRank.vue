<template>
  <section class="user-ranking f-card">
    <div class="f-card-head">
      <span class="f-blue">
        <el-icon :size="18">
          <Histogram />
        </el-icon>
      </span>
      <b class="text-[19px] font-medium text-blue-500">用户排行</b>
      <button type="button" class="rank-more" @click="goRank">
        查看全部
        <el-icon :size="13">
          <ArrowRight />
        </el-icon>
      </button>
    </div>

    <ul v-if="rankList.length" class="rank-list">
      <li
        v-for="(item, index) in rankList"
        :key="item.id ?? index"
        class="rank-row"
      >
        <span class="rank-badge" :class="rankClass(index)">{{ index + 1 }}</span>

        <el-avatar
          :size="32"
          :src="item.avatar"
          class="rank-avatar"
          :class="rankClass(index)"
        >
          {{ item.username?.charAt(0)?.toUpperCase() }}
        </el-avatar>

        <button type="button" class="rank-name" @click="goUser(item.id)">
          {{ item.username }}
        </button>

        <span class="rank-count">
          <b>{{ item.count }}</b>
          <i>题</i>
        </span>
      </li>
    </ul>

    <div v-else class="rank-empty">暂无排名</div>
  </section>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ArrowRight, Histogram } from '@element-plus/icons-vue'

defineProps({
  rankList: {
    type: Array,
    default: () => [],
  },
})

const router = useRouter()

const goUser = (id) => {
  router.push({
    name: 'User',
    params: {
      id,
    },
  })
}

const goRank = () => router.push({ name: 'Rank' })

const rankClass = (index) => {
  if (index === 0) return 'is-gold'
  if (index === 1) return 'is-silver'
  if (index === 2) return 'is-bronze'
  return ''
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

.rank-more {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-left: auto;
  color: #3b82f6;
  font-size: 13px;
  transition: color 0.15s ease;
}

.rank-more:hover {
  color: #60a5fa;
}

.rank-list {
  display: flex;
  flex-direction: column;
  margin-top: 14px;
}

.rank-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin: 0 -8px;
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  border-radius: 6px;
  transition: background-color 0.15s ease;

  &:last-child {
    border-bottom: 0;
  }

  &:hover {
    background-color: #f8fafc;
  }
}

.rank-badge {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: #f3f4f6;
  color: #9ca3af;
  font-size: 12px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;

  &.is-gold {
    background: linear-gradient(135deg, #f7c948, #e0a10d);
    box-shadow: 0 2px 6px rgb(224 161 13 / 30%);
    color: #fff;
  }

  &.is-silver {
    background: linear-gradient(135deg, #cfd6df, #9aa4b2);
    box-shadow: 0 2px 6px rgb(154 164 178 / 30%);
    color: #fff;
  }

  &.is-bronze {
    background: linear-gradient(135deg, #e2ab79, #c17f45);
    box-shadow: 0 2px 6px rgb(193 127 69 / 30%);
    color: #fff;
  }
}

.rank-avatar {
  flex-shrink: 0;
  background: #eff6ff;
  color: #3b82f6;
  font-size: 13px;

  &.is-gold {
    box-shadow: 0 0 0 2px #fde68a;
  }

  &.is-silver {
    box-shadow: 0 0 0 2px #e5e7eb;
  }

  &.is-bronze {
    box-shadow: 0 0 0 2px #fed7aa;
  }
}

.rank-name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  color: #374151;
  font-size: 15px;
  font-weight: 400;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 0.15s ease;

  &:hover {
    color: #3b82f6;
  }
}

.rank-count {
  flex-shrink: 0;
  color: #4b5563;
  font-size: 15px;
  font-variant-numeric: tabular-nums;

  b {
    margin-right: 2px;
    color: #4b5563;
    font-size: 15px;
    font-weight: 400;
  }

  i {
    color: #6b7280;
    font-size: 12px;
    font-style: normal;
  }
}

.rank-empty {
  padding: 24px 0;
  color: #9ca3af;
  font-size: 14px;
  text-align: center;
}
</style>
