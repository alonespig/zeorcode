<template>
  <section class="settings-card">
    <header>
      <h3>{{ block.title }}</h3>
      <p>{{ block.description }}</p>
    </header>
    <el-form label-position="top" class="settings-form" @submit.prevent>
      <div class="two-columns">
        <el-form-item label="作业名称">
          <el-input v-model="form.title" maxlength="100" :disabled="disabled" />
        </el-form-item>
        <el-form-item label="题目数量">
          <el-input-number v-model="form.problemCount" :min="1" :max="30" :disabled="disabled" controls-position="right" />
        </el-form-item>
      </div>
      <div class="two-columns">
        <el-form-item label="难度">
          <el-select v-model="form.difficulty" :disabled="disabled">
            <el-option label="不限难度" :value="0" />
            <el-option label="入门" :value="1" />
            <el-option label="普及" :value="2" />
            <el-option label="提高" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="历史作业题目">
          <el-select v-model="form.repeatPolicy" :disabled="disabled">
            <el-option label="排除以前使用过的题目" value="exclude_all" />
            <el-option label="允许重复使用" value="allow_history" />
          </el-select>
        </el-form-item>
      </div>
      <el-form-item label="开始和结束时间">
        <el-date-picker
          v-model="timeRange"
          type="datetimerange"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DD HH:mm:ss"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :disabled="disabled"
          :default-time="defaultTimes"
        />
      </el-form-item>
      <el-form-item label="作业说明（可选）">
        <el-input v-model="form.description" type="textarea" :rows="2" maxlength="1000" :disabled="disabled" />
      </el-form-item>
    </el-form>
    <footer>
      <span v-if="disabled">此设置已提交</span>
      <button v-else type="button" :disabled="!valid || submitting" @click="submit">生成题目方案</button>
    </footer>
  </section>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'

const props = defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
const emit = defineEmits(['submit'])
const defaults = props.block.payload || {}
const form = reactive({
  title: defaults.defaultTitle || '',
  description: '',
  problemCount: defaults.defaultProblemCount || 8,
  difficulty: defaults.defaultDifficulty || 0,
  repeatPolicy: defaults.defaultRepeatPolicy || 'exclude_all',
})
const timeRange = ref([defaults.defaultStartTime, defaults.defaultEndTime].filter(Boolean))
const defaultTimes = [new Date(2000, 1, 1, 8, 0, 0), new Date(2000, 1, 1, 23, 59, 0)]
const valid = computed(() => form.title.trim() && timeRange.value?.length === 2)
const submit = () => emit('submit', props.block.requestId, {
  ...form,
  title: form.title.trim(),
  startTime: timeRange.value[0],
  endTime: timeRange.value[1],
})
</script>

<style scoped>
.settings-card { margin-top: 14px; padding: 19px; border: 1px solid #dfe6f0; border-radius: 7px; background: #fbfcfe; }
header h3 { margin: 0; color: #202c40; font-size: 15px; }
header p { margin: 6px 0 0; color: #7c879a; font-size: 13px; }
.settings-form { margin-top: 17px; }
.two-columns { display: grid; grid-template-columns: minmax(0, 1.55fr) minmax(150px, .75fr); gap: 15px; }
.settings-card :deep(.el-form-item) { margin-bottom: 15px; }
.settings-card :deep(.el-form-item__label) { color: #596579; font-size: 12px; }
.settings-card :deep(.el-select), .settings-card :deep(.el-input-number), .settings-card :deep(.el-date-editor) { width: 100%; }
footer { display: flex; justify-content: flex-end; align-items: center; }
footer span { color: #9da6b5; font-size: 12px; }
footer button { height: 35px; padding: 0 17px; color: #fff; border-radius: 5px; background: #367bf5; font-size: 13px; }
footer button:disabled { background: #bdc9da; cursor: not-allowed; }
@media (max-width: 660px) { .two-columns { grid-template-columns: 1fr; gap: 0; } }
</style>
