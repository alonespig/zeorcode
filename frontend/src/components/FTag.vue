<script setup>
import { JUDGE_STATUS_CLASS, JUDGE_STATUS_TEXT } from '@/constants/index'

defineProps({
  result: {
    type: Number,
    default: 0,
  },
})
</script>

<template>
  <span class="ftag" :class="JUDGE_STATUS_CLASS[result] || 'unknown'">
    {{ JUDGE_STATUS_TEXT[result] }}
  </span>
</template>

<style scoped>
/* 结果只用文字色，不增加填充背景，保持原有 OJ 风格。 */
.ftag {
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}
</style>
