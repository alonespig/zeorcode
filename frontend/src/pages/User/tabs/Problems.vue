<template>
  <div class="px-6 py-5">
    <!-- 已解决 -->
    <section>
      <h3 class="mb-3 text-sm font-semibold text-gray-700">
        <span class="flex items-center text-base gap-1 text-blue-500">
          <el-icon>
            <DocumentChecked />
          </el-icon>
          <span>
            已解决
            <span class="ml-1 text-xs font-normal text-gray-400">
              {{ solveItem.length }} 题
            </span>
          </span>
        </span>
      </h3>
      <div class="flex flex-wrap gap-2">
        <span v-for="p in solveItem" :key="p.id" @click="goProblem(p.id)"
          class="cursor-pointer rounded border border-emerald-200 bg-emerald-50 px-2.5 py-1 text-[13px] text-emerald-700 hover:border-blue-400 hover:text-blue-500">
          {{ p.id }}. {{ p.name }}
        </span>
        <span v-if="!solveItem.length" class="py-1 text-sm text-gray-400">暂无题目</span>
      </div>
    </section>

    <!-- 尝试过 -->
    <section class="mt-7">
      <h3 class="mb-3 text-base font-semibold text-gray-700">
        <span class="flex items-center text-base gap-1 text-blue-500">
          <el-icon>
            <DocumentDelete />
          </el-icon>
          <span>
            尝试过
            <span class="ml-1 text-xs font-normal text-gray-400">
              {{ unsolveItem.length }} 题
            </span>
          </span>
        </span>
      </h3>
      <div class="flex flex-wrap gap-2">
        <span v-for="p in unsolveItem" :key="p.id" @click="goProblem(p.id)"
          class="cursor-pointer rounded border border-amber-200 bg-amber-50 px-2.5 py-1 text-[13px] text-amber-700 hover:border-blue-400 hover:text-blue-500">
          {{ p.id }}. {{ p.name }}
        </span>
        <span v-if="!unsolveItem.length" class="py-1 text-sm text-gray-400">暂无题目</span>
      </div>
    </section>

    <!-- 近七天通过数 -->
    <section class="mt-7">
      <h3 class="mb-3 text-base flex gap-1 items-center font-semibold text-blue-500">
        <el-icon>
          <DataLine />
        </el-icon>
        近七天通过数
      </h3>
      <div class="rounded-lg border border-gray-100 bg-white px-2 pt-2">
        <PassCountChart :dates="dates" :counts="counts" />
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, inject, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PassCountChart from '@/components/user/PassCountChart.vue'
import { getUserRecent7DaysAc } from '@/api/user'
import { DocumentChecked, DocumentDelete, DataLine } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const profile = inject('userProfile')
const solveItem = computed(() => profile.value?.solveItem || [])
const unsolveItem = computed(() => profile.value?.unsolveItem || [])

const dates = ref([])
const counts = ref([])

const goProblem = (id) => router.push({ name: 'ProblemDetail', params: { id } })

const fetchChart = async () => {
  const res = await getUserRecent7DaysAc(route.params.id)
  dates.value = res.data.dates
  counts.value = res.data.counts
}

onMounted(fetchChart)
watch(() => route.params.id, fetchChart)
</script>
