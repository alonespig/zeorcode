<template>
  <table class="data-table">
    <colgroup>
      <col style="width: 10%;" />
      <col style="width: 50%;" />
      <col style="width: 20%;" />
      <col style="width: 20%;" />
    </colgroup>
    <thead>
      <tr>
        <th>排名</th>
        <th class="left">用户名</th>
        <th>Rating</th>
        <th>最高分</th>
      </tr>
    </thead>
    <tbody>
      <tr class="font-[Arial,'Noto_Sans_SC',sans-serif] [&>td]:py-2" v-for="u in list" :key="u.id">
        <td>
          <span class="tabular-nums" :class="medalClass(u.index)">{{ u.index }}</span>
        </td>
        <td class="left">
          <div class="flex items-center gap-4">
            <img class="w-10 h-10 rounded-full" :src="u.avatar" alt="">
            <UserName :name="u.username" :rating="u.rating" clickable @click="$emit('click-user', u.id)" />
          </div>
        </td>
        <td class="font-bold tabular-nums" :style="{ color: tierColor(u.rating) }">{{ u.rating }}</td>
        <td class="tabular-nums text-gray-400">{{ u.maxRating || u.rating }}</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
import UserName from '@/components/UserName.vue'
import { ratingTier } from '@/constants/index'

defineProps({
  list: { type: Array, default: () => [] },
})
defineEmits(['click-user'])

const tierColor = (r) => ratingTier(r).color
const medalClass = (index) =>
  index === 1 ? 'font-bold text-[#e6a817]'
    : index === 2 ? 'font-bold text-[#9aa7b4]'
      : index === 3 ? 'font-bold text-[#cd7f47]' : ''
</script>
