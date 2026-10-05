<script setup lang="ts">
import { useRoute } from "vue-router";
import ProblemSetForm from "./components/ProblemSetForm.vue";

defineOptions({ name: "AdminProblemSetsEdit" });

const route = useRoute();
const problemSetId = Number(route.params.id);
</script>

<template>
  <ProblemSetForm :problem-set-id="problemSetId" />
</template>
