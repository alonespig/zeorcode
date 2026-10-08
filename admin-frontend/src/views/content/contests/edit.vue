<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import ContestForm from "./components/ContestForm.vue";

defineOptions({ name: "AdminContestsEdit" });

const route = useRoute();
const contestId = computed(() => Number(route.params.id));
</script>

<template>
  <ContestForm mode="edit" :contest-id="contestId" />
</template>
