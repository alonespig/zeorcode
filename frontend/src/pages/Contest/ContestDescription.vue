<template>
  <main class="f-panel px-6 py-4 font-sans">
    <MdEditor :model-value="contest?.description" preview-only code-theme="github" />
  </main>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { getContestDesc } from "@/api/contest";
import { useRoute } from "vue-router";
import MdEditor from "@/components/MdEditor.vue";

const contest = ref(null);
const route = useRoute();
const params = computed(() => route.params);

onMounted(async () => {
  const res = await getContestDesc(params.value.id);
  contest.value = res.data || {};
});
</script>
