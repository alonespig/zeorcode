<template>
  <table class="data-table">
    <colgroup>
      <col style="width: 20%" />
      <col style="width: 50%" />
      <col style="width: 30%" />
    </colgroup>
    <thead>
      <tr>
        <th>题目ID</th>
        <th>题目标题</th>
        <th>创建时间</th>
      </tr>
    </thead>
    <tbody>
      <tr class="font-[Arial,'Noto_Sans_SC',sans-serif]" v-for="problem in problems" :key="problem.id">
        <td class="center">{{ problem.id }}</td>
        <td>
          <span class="tw-link" @click="clickProblem(problem.id)">
            {{ problem.name }}
          </span>
        </td>
        <td class="center">
          {{ dayjs(problem.createdAt).fromNow() }}
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'
dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

const props = defineProps({
  problems: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['click-problem'])

const clickProblem = (id) => emit('click-problem', id)

</script>

<style scoped></style>
