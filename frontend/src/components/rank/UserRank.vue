<template>
  <table class="data-table">
    <colgroup>
      <col style="width: 10%;" />
      <col style="width: 20%;" />
      <col style="width: 40%;" />
      <col style="width: 15%;" />
      <col style="width: 15%;" />
    </colgroup>
    <thead>
      <tr>
        <th>排名</th>
        <th class="left">用户名</th>
        <th>个性签名</th>
        <th>通过题目数</th>
        <th>提交通过率</th>
      </tr>
    </thead>
    <tbody>
      <tr class="font-[Arial,'Noto_Sans_SC',sans-serif] [&>td]:py-2" v-for="user in userRankList" :key="user.id">
        <td> {{ user.index }} </td>
        <td class="left">
          <div class="flex items-center gap-4">
            <img class="w-10 h-10 rounded-full" :src="user.avatar" alt="">
            <UserName :name="user.username" :rating="user.rating" clickable @click="clickUser(user.id)" />
          </div>
        </td>
        <td>{{ user.signature }}</td>
        <td> {{ user.count ?? 0 }} </td>
        <td> {{ passRate(user) }}% </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
import UserName from '@/components/UserName.vue'

defineProps({
  userRankList: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['click-user'])

// AC 提交次数 / 总提交次数，与题目页的“提交通过率”口径一致
const passRate = (user) => {
  if (!user.submitCount) return 0
  const rate = Math.round(((user.acSubmitCount ?? 0) / user.submitCount) * 100)
  return Math.min(100, Math.max(0, rate))
}

const clickUser = (id) => emit('click-user', id)
</script>
