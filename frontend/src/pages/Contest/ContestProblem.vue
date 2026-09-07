<template>
  <div class="f-panel overflow-hidden">
    <ProblemTable :problemList="problemList" @click-problem="handleClickProblem" />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import { getContestProblemList } from "@/api/contest";
import { useRoute, useRouter } from "vue-router";
import ProblemTable from "@/components/contest/ProblemTable.vue";

const problemList = ref(null);
const route = useRoute();
const router = useRouter();
const params = computed(() => route.params);

const handleClickProblem = (label) => {
  router.push({
    path: `/contest/${params.value.id}/problem/${label}`,
  });
}

const handleClickSubmit = (id) => {
  router.push({
    path: `/contest/${params.value.id}/submit`,
    query: {
      id: id
    }
  })
}

onMounted(async () => {
  const res = await getContestProblemList(params.value.id);
  problemList.value = res.data.problemList || {};
});

</script>
