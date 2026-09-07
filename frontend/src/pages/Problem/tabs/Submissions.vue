<template>
  <div class="f-panel overflow-hidden">
    <div class="border-b border-gray-100 px-4 py-3 text-base font-medium text-gray-700">提交记录 · 本题</div>

    <SubmissionStatus v-if="tableData.length" :tableData="tableData" @click-id="goDetail" @click-user="goUser" />
    <el-empty v-else description="还没有提交记录" />

    <div v-if="total > pageSize" class="f-pagination">
      <el-pagination layout="prev, pager, next" :current-page="page" :page-size="pageSize" :total="total"
        @current-change="handlePageChange" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SubmissionStatus from '@/components/SubmissionStatus.vue'
import { getSubmit } from '@/api/user'

const route = useRoute()
const router = useRouter()

const problemID = route.params.id
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 15

const fetchData = async () => {
  const res = await getSubmit({ problemID, page: page.value, pageSize })
  tableData.value = res.data.list || []
  total.value = res.data.total || 0
}

const goDetail = (id) => router.push({ name: 'SubmissionDetail', params: { id } })
const goUser = (id) => router.push({ name: 'User', params: { id } })

const handlePageChange = (val) => {
  page.value = val
  fetchData()
}

onMounted(fetchData)
</script>
