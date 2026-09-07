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
        <th>通过率</th>
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
        <td> {{ user.count }} </td>
        <td> {{ user.submitCount === 0 ? '0' : (user.count / user.submitCount * 100).toFixed(0) }}% </td>
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

const clickUser = (id) => emit('click-user', id)
</script>
