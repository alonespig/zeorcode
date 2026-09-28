<script setup>
import { computed, shallowRef, watch } from 'vue'
import { getProblem } from '@/api/problems'
import ContestProblem from './ContestProblem.vue'

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  problemId: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
const problem = shallowRef(null)
const loading = shallowRef(false)
const failed = shallowRef(false)
const retry = shallowRef(0)

watch(
  [() => props.modelValue, () => props.problemId, retry],
  async ([isOpen, id], _, onCleanup) => {
    // 关闭或切换题目后，忽略旧请求，防止旧题面覆盖新预览。
    let active = true
    onCleanup(() => { active = false })
    problem.value = null
    failed.value = false
    loading.value = false
    if (!isOpen || !id) return
    loading.value = true
    try {
      const res = await getProblem(id)
      if (active) {
        problem.value = res.data
        failed.value = !res.data
      }
    } catch {
      if (active) failed.value = true
    } finally {
      if (active) loading.value = false
    }
  },
  { immediate: true },
)
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="`题目预览 · ${problemId}`"
    width="min(960px, calc(100vw - 32px))"
    top="5vh"
    append-to-body
    destroy-on-close
  >
    <div v-loading="loading" class="problem-preview" :aria-busy="loading">
      <el-empty v-if="failed" description="题目加载失败，请重试">
        <el-button type="primary" @click="retry++">重新加载</el-button>
      </el-empty>
      <ContestProblem v-else-if="problem" :key="problemId" :problem="problem" />
    </div>
    <template #footer>
      <el-button @click="visible = false">返回选题</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.problem-preview {
  min-height: 200px;
  max-height: 70vh;
  overflow: auto;
  overflow-wrap: anywhere;
}
</style>
