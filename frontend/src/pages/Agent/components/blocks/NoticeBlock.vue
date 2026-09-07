<template>
  <section class="notice-card">
    <div class="notice-heading">
      <el-icon><Warning /></el-icon>
      <div><h3>{{ block.title }}</h3><p>{{ block.description }}</p></div>
    </div>
    <div class="notice-actions">
      <button
        v-for="option in block.options || []"
        :key="option.value"
        type="button"
        :disabled="disabled || submitting || option.disabled"
        @click="$emit('submit', block.requestId, option.value)"
      >{{ option.label }}</button>
    </div>
    <span v-if="disabled" class="expired">此选项已处理</span>
  </section>
</template>

<script setup>
import { Warning } from '@element-plus/icons-vue'
defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
defineEmits(['submit'])
</script>

<style scoped>
.notice-card { margin-top: 14px; padding: 17px; border: 1px solid #f0ddb5; border-radius: 6px; background: #fffaf0; }
.notice-heading { display: flex; align-items: flex-start; gap: 10px; color: #bd7b12; }
h3 { margin: 0; color: #70501e; font-size: 14px; }
p { margin: 5px 0 0; color: #8d7654; font-size: 12px; }
.notice-actions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 14px; }
button { min-height: 33px; padding: 0 13px; color: #75531f; border: 1px solid #e8d2a6; border-radius: 5px; background: #fff; font-size: 12px; }
button:hover:not(:disabled) { border-color: #d3a653; background: #fff7e8; }
button:disabled { opacity: .45; cursor: not-allowed; }
.expired { display: block; margin-top: 11px; color: #a79270; font-size: 11px; }
</style>
