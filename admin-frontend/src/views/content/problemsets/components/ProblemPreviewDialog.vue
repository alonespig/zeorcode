<script setup lang="ts">
import { nextTick, shallowRef, useTemplateRef, watch } from "vue";
import ProblemSheet from "@/components/ProblemSheet.vue";
import { getProblemForEdit, type ProblemForEdit } from "@/api/admin/problems";

defineOptions({ name: "ProblemPreviewDialog" });

const props = defineProps<{
  problemId: string;
}>();

const visible = defineModel<boolean>({ default: false });
const problem = shallowRef<ProblemForEdit | null>(null);
const loading = shallowRef(false);
const failed = shallowRef(false);
const retry = shallowRef(0);
const previewRef = useTemplateRef<HTMLElement>("previewRef");

watch(
  [visible, () => props.problemId, retry],
  async ([isOpen, problemId], _, onCleanup) => {
    let active = true;
    onCleanup(() => {
      active = false;
    });

    problem.value = null;
    failed.value = false;
    loading.value = false;
    if (!isOpen || !problemId) return;

    loading.value = true;
    try {
      const res = await getProblemForEdit(problemId);
      if (active) {
        problem.value = res.data ?? null;
        failed.value = !res.data;
        await nextTick();
        previewRef.value?.scrollTo({ top: 0 });
      }
    } catch {
      if (active) failed.value = true;
    } finally {
      if (active) loading.value = false;
    }
  },
  { immediate: true }
);
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
    <div
      ref="previewRef"
      v-loading="loading"
      class="problem-preview"
      :aria-busy="loading"
    >
      <el-empty v-if="failed" description="题目加载失败，请重试">
        <el-button type="primary" @click="retry++">重新加载</el-button>
      </el-empty>

      <ProblemSheet v-else-if="problem" :key="problemId" :problem="problem" />
    </div>

    <template #footer>
      <el-button @click="visible = false">返回题单</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.problem-preview {
  min-height: 220px;
  max-height: 70vh;
  overflow: auto;
  overflow-wrap: anywhere;
}
</style>
