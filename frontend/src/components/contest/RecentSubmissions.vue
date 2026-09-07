<script setup>
import { ref, shallowRef, watch } from 'vue'
import { getSubmission } from '@/api/contest'
import RecentSubmissionsPanel from '@/components/submission/RecentSubmissionsPanel.vue'

const props = defineProps({
  contestId: {
    type: [String, Number],
    required: true,
  },
  problemId: {
    type: [String, Number],
    default: undefined,
  },
})

const emit = defineEmits(['click-id', 'view-all'])

const submissions = ref([])
const loading = shallowRef(false)
let fetchVersion = 0

const fetchSubmissions = async () => {
  const version = ++fetchVersion
  if (!props.contestId || !props.problemId) {
    submissions.value = []
    return
  }

  loading.value = true
  try {
    const res = await getSubmission(props.contestId, { problemID: props.problemId })
    if (version !== fetchVersion) return

    submissions.value = (res.data.list || [])
      .filter((item) => String(item.problemID) === String(props.problemId))
      .slice(0, 4)
  } catch {
    if (version === fetchVersion) submissions.value = []
  } finally {
    if (version === fetchVersion) loading.value = false
  }
}

watch(
  () => [props.contestId, props.problemId],
  fetchSubmissions,
  { immediate: true }
)
</script>

<template>
  <RecentSubmissionsPanel :submissions="submissions" :loading="loading" @click-id="emit('click-id', $event)"
    @view-all="emit('view-all')" />
</template>
