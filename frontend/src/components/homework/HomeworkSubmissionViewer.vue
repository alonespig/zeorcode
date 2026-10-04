<script setup>
import { shallowRef, watch } from "vue";
import { getHomeworkSubmissionDetail } from "@/api/team";
import { JUDGE_STATUS } from "@/constants/index";
import HomeworkSubmissionDialog from "./HomeworkSubmissionDialog.vue";

const visible = defineModel({ type: Boolean, default: false });
const props = defineProps({
  homeworkId: { type: [String, Number], required: true },
  submissionId: { type: [String, Number], default: null },
});
const emit = defineEmits(["settled"]);
const submission = shallowRef({});
const loading = shallowRef(false);

watch(
  () => [visible.value, props.homeworkId, props.submissionId],
  ([isVisible, homeworkId, submissionId], _previous, onCleanup) => {
    let active = true;
    let timer;
    onCleanup(() => {
      active = false;
      window.clearTimeout(timer);
    });
    submission.value = {};
    loading.value = false;
    if (!isVisible || !homeworkId || !submissionId) return;
    loading.value = true;

    const load = async () => {
      try {
        const res = await getHomeworkSubmissionDetail(homeworkId, submissionId);
        if (!active) return;
        submission.value = res.data || {};
        const status = submission.value.status;
        if (status === JUDGE_STATUS.PENDING || status === JUDGE_STATUS.JUDGING) {
          timer = window.setTimeout(load, 1500);
        } else {
          emit("settled");
        }
      } catch {
        // 请求层展示后端错误；关闭弹窗，避免残留空详情或旧提交内容。
        if (active) visible.value = false;
      } finally {
        if (active) loading.value = false;
      }
    };
    load();
  },
  { immediate: true },
);
</script>

<template>
  <HomeworkSubmissionDialog v-model="visible" :submission="submission" :loading="loading" />
</template>
