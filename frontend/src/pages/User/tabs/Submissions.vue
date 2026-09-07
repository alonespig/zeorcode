<template>
  <div>
    <SubmissionStatus v-if="tableData.length" :tableData="tableData" hide-user @click-id="goDetail"
      @click-problem-id="goProblem" />
    <el-empty v-else description="还没有提交记录" />

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination layout="prev, pager, next" :current-page="page" :page-size="pageSize" :total="total"
        @current-change="onPage" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SubmissionStatus from '@/components/SubmissionStatus.vue'
import { getSubmit } from '@/api/user'

const route = useRoute()
const router = useRouter()

const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 15

const fetchData = async () => {
  const res = await getSubmit({ userID: Number(route.params.id), page: page.value, pageSize })
  tableData.value = res.data.list || []
  total.value = res.data.total || 0
}

const goDetail = (id) => router.push({ name: 'SubmissionDetail', params: { id } })
const goProblem = (id) => router.push({ name: 'ProblemDetail', params: { id } })

const onPage = (val) => {
  page.value = val
  fetchData()
}

onMounted(fetchData)
watch(() => route.params.id, () => {
  page.value = 1
  fetchData()
})
</script>
