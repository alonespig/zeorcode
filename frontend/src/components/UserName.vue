<template>
  <span class="font-medium font-sans" :style="{ color: tier.color }" :class="{ 'cursor-pointer hover:underline': clickable }"
    @click="onClick">{{ name }}</span>
</template>

<script setup>
import { computed } from 'vue'
import { ratingTier } from '@/constants/index'

const props = defineProps({
  name: { type: String, default: '' },
  rating: { type: Number, default: 0 },
  clickable: { type: Boolean, default: false },
})
const emit = defineEmits(['click'])

const tier = computed(() => ratingTier(props.rating))
const onClick = () => {
  if (props.clickable) emit('click')
}
</script>
