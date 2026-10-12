<script setup>
import { computed, reactive } from 'vue'

const props = defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
const emit = defineEmits(['submit', 'cancel'])
const defaults = props.block.payload || {}
const form = reactive({
  title: defaults.defaultTitle || '',
  description: '',
  problemCount: defaults.defaultProblemCount || 8,
  published: defaults.defaultPublished ?? 0,
  visibility: defaults.defaultVisibility ?? 0,
  inviteCode: '',
})
const valid = computed(() => Boolean(
  form.title.trim()
  && form.problemCount >= 1
  && form.problemCount <= 30
  && (form.visibility !== 1 || form.inviteCode.trim()),
))

const submit = () => emit('submit', props.block.requestId, {
  ...form,
  title: form.title.trim(),
  description: form.description.trim(),
  inviteCode: form.inviteCode.trim(),
})
</script>

<template>
  <section class="settings-card">
    <header class="settings-header">
      <h3 class="settings-title">{{ block.title }}</h3>
      <p class="settings-description">{{ block.description }}</p>
    </header>
    <div class="selection-summary">
      <span><small>训练类型</small>{{ defaults.trainingType }}</span>
      <span><small>难度</small>{{ defaults.difficulty }}</span>
      <span v-if="defaults.knowledgePoints?.length"><small>知识点</small>{{ defaults.knowledgePoints.join('、') }}</span>
      <span v-if="defaults.requirements"><small>补充要求</small>{{ defaults.requirements }}</span>
    </div>
    <el-form label-position="top" class="settings-form" @submit.prevent>
      <div class="basic-fields">
        <el-form-item label="题单名称">
          <el-input v-model="form.title" maxlength="100" show-word-limit :disabled="disabled" />
        </el-form-item>
        <el-form-item label="题目数量">
          <el-input-number v-model="form.problemCount" :min="1" :max="30" :disabled="disabled" controls-position="right" />
        </el-form-item>
      </div>
      <el-form-item label="题单说明（可选，支持 Markdown）">
        <el-input v-model="form.description" type="textarea" :rows="3" maxlength="5000" :disabled="disabled" />
      </el-form-item>
      <div class="publish-fields">
        <el-form-item label="发布状态">
          <el-select v-model="form.published" :disabled="disabled">
            <el-option label="保存为草稿" :value="0" />
            <el-option label="立即发布" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="访问方式">
          <el-select v-model="form.visibility" :disabled="disabled">
            <el-option label="公开访问" :value="0" />
            <el-option label="邀请码访问" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="form.visibility === 1" label="邀请码">
          <el-input v-model="form.inviteCode" maxlength="32" placeholder="请输入邀请码" :disabled="disabled" />
        </el-form-item>
      </div>
    </el-form>
    <footer class="settings-footer">
      <span v-if="disabled">此设置已提交</span>
      <template v-else>
        <button class="cancel-button" type="button" :disabled="submitting" @click="emit('cancel')">取消任务</button>
        <button class="primary-button" type="button" :disabled="!valid || submitting" @click="submit">生成题目方案</button>
      </template>
    </footer>
  </section>
</template>

<style scoped>
.settings-card { margin-top: 14px; padding: 19px; border: 1px solid #dfe6f0; border-radius: 7px; background: #fbfcfe; }
.settings-title { margin: 0; color: #202c40; font-size: 15px; }
.settings-description { margin: 6px 0 0; color: #7c879a; font-size: 13px; }
.selection-summary { display: flex; flex-wrap: wrap; gap: 18px; margin-top: 15px; padding: 10px 12px; border: 1px solid #e4e9f1; border-radius: 5px; background: #fff; color: #3d495c; font-size: 12px; }
.selection-summary span { display: flex; align-items: center; gap: 7px; }
.selection-summary small { color: #929cad; font-size: 11px; }
.settings-form { margin-top: 17px; }
.basic-fields { display: grid; grid-template-columns: minmax(0, 1fr) 150px; gap: 15px; }
.publish-fields { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 15px; }
.settings-card :deep(.el-form-item) { margin-bottom: 15px; }
.settings-card :deep(.el-form-item__label) { color: #596579; font-size: 12px; }
.settings-card :deep(.el-select), .settings-card :deep(.el-input-number) { width: 100%; }
.settings-footer { display: flex; justify-content: flex-end; align-items: center; gap: 9px; }
.settings-footer span { color: #9da6b5; font-size: 12px; }
.settings-footer button { height: 35px; padding: 0 17px; border-radius: 5px; font-size: 13px; }
.primary-button { color: #fff; background: #367bf5; }
.primary-button:disabled { background: #bdc9da; cursor: not-allowed; }
.cancel-button { color: #606d82; border: 1px solid #d8dee8; background: #fff; }
.cancel-button:disabled { color: #9da6b5; background: #f5f6f8; }
@media (max-width: 760px) { .basic-fields, .publish-fields { grid-template-columns: 1fr; gap: 0; } }
</style>
