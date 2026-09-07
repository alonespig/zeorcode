<template>
  <section class="result-card">
    <div class="result-icon"><el-icon><CircleCheckFilled /></el-icon></div>
    <div class="result-copy">
      <h3>{{ block.title }}</h3>
      <p>{{ block.payload?.title }}</p>
    </div>
    <RouterLink v-if="block.payload?.link" :to="block.payload.link">查看作业</RouterLink>
  </section>
</template>

<script setup>
import { CircleCheckFilled } from '@element-plus/icons-vue'
defineProps({ block: { type: Object, required: true } })
</script>

<style scoped>
.result-card { margin-top: 14px; min-height: 72px; display: flex; align-items: center; gap: 12px; padding: 15px 17px; border: 1px solid #cce8d8; border-radius: 6px; background: #f3fbf6; }
.result-icon { color: #29a565; font-size: 27px; }
.result-copy { min-width: 0; flex: 1; }
h3 { margin: 0; color: #20583b; font-size: 14px; }
p { margin: 4px 0 0; color: #69907b; font-size: 12px; }
a { height: 33px; display: inline-flex; align-items: center; padding: 0 13px; color: #238352; border: 1px solid #9fd3b6; border-radius: 5px; background: #fff; font-size: 12px; }
</style>
