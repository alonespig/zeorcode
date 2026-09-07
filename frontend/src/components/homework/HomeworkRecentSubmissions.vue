<script setup>
import { ref, shallowRef, watch } from "vue";
import { getHomeworkSubmissions } from "@/api/team";
import RecentSubmissionsPanel from "@/components/submission/RecentSubmissionsPanel.vue";

const props = defineProps({
  homeworkId: {
    type: [String, Number],
    required: true,
  },
  problemId: {
    type: [String, Number],
    required: true,
  },
});

const emit = defineEmits(["click-id", "view-all"]);
const submissions = ref([]);
const loading = shallowRef(false);
let fetchVersion = 0;

const fetchSubmissions = async () => {
  const version = ++fetchVersion;
  if (!props.homeworkId || !props.problemId) {
    submissions.value = [];
    return;
  }

  loading.value = true;
  try {
    const res = await getHomeworkSubmissions(props.homeworkId, {
      page: 1,
      pageSize: 4,
      problemId: props.problemId,
    });
    if (version !== fetchVersion) return;
    submissions.value = (res.data?.list || []).map((item) => ({ ...item, result: item.status }));
  } catch {
    if (version === fetchVersion) submissions.value = [];
  } finally {
    if (version === fetchVersion) loading.value = false;
  }
};

watch(
  () => [props.homeworkId, props.problemId],
  fetchSubmissions,
  { immediate: true }
);
</script>

<template>
  <RecentSubmissionsPanel :submissions="submissions" :loading="loading" @click-id="emit('click-id', $event)"
    @view-all="emit('view-all')" />
</template>
