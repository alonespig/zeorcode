<template>
  <section class="preview-card">
    <div class="preview-title">
      <div class="preview-icon"><el-icon><DocumentChecked /></el-icon></div>
      <div>
        <h3>{{ block.title }}</h3>
        <p>{{ block.description }}</p>
      </div>
    </div>
    <dl>
      <div><dt>团队</dt><dd>{{ payload.teamName }}</dd></div>
      <div><dt>作业名称</dt><dd>{{ payload.title }}</dd></div>
      <div><dt>时间</dt><dd>{{ payload.startTime }} 至 {{ payload.endTime }}</dd></div>
      <div><dt>题目</dt><dd>{{ payload.problemCount }} 题 · {{ (payload.knowledgeTags || []).join('、') }}</dd></div>
    </dl>
    <div class="preview-footer">
      <span v-if="disabled">此操作已处理</span>
      <template v-else>
        <button class="cancel" type="button" :disabled="submitting" @click="$emit('reject', payload.actionId, payload.draftVersion)">取消</button>
        <button class="confirm" type="button" :disabled="submitting" @click="$emit('approve', payload.actionId, payload.draftVersion)">
          确认创建
        </button>
      </template>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { DocumentChecked } from '@element-plus/icons-vue'
const props = defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
defineEmits(['approve', 'reject'])
const payload = computed(() => props.block.payload || {})
</script>

<style scoped>
.preview-card { margin-top: 12px; padding: 18px; border: 1px solid #cadeff; border-left: 3px solid #4d86ed; border-radius: 5px; background: #f7faff; }
.preview-title { display: flex; gap: 11px; }
.preview-icon { width: 34px; height: 34px; flex: 0 0 34px; display: grid; place-items: center; color: #367bf5; border-radius: 6px; background: #e7f0ff; }
h3 { margin: 0; color: #20304b; font-size: 15px; }
p { margin: 4px 0 0; color: #7e899b; font-size: 12px; }
dl { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); margin: 17px 0 0; border: 1px solid #e1e8f2; background: #fff; }
dl > div { min-height: 58px; padding: 10px 13px; border-right: 1px solid #edf0f5; border-bottom: 1px solid #edf0f5; }
dl > div:nth-child(2n) { border-right: 0; }
dl > div:nth-last-child(-n+2) { border-bottom: 0; }
dt { color: #929cac; font-size: 11px; }
dd { margin: 5px 0 0; color: #344158; font-size: 13px; font-weight: 600; }
.preview-footer { min-height: 35px; display: flex; align-items: center; justify-content: flex-end; gap: 9px; margin-top: 14px; }
.preview-footer span { color: #929cac; font-size: 12px; }
button { height: 34px; padding: 0 16px; border-radius: 5px; font-size: 13px; }
.cancel { color: #606d82; border: 1px solid #d8dee8; background: #fff; }
.confirm { color: #fff; background: #367bf5; }
button:disabled { opacity: .55; cursor: not-allowed; }
@media (max-width: 660px) { dl { grid-template-columns: 1fr; } dl > div { border-right: 0; } dl > div:nth-last-child(-n+2) { border-bottom: 1px solid #edf0f5; } dl > div:last-child { border-bottom: 0; } }
</style>
