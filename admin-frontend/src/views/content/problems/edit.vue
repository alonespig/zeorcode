<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import ProblemForm from "./components/ProblemForm.vue";

defineOptions({ name: "AdminProblemsEdit" });

const route = useRoute();
const problemId = computed(() => String(route.params.id ?? ""));
</script>

<template>
  <ProblemForm mode="edit" :problem-id="problemId" />
</template>
