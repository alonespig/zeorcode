<template>
  <div class="markdown-body" v-html="html"></div>
</template>

<script setup>
import { computed } from 'vue'
import { renderMarkdown } from '@/utils/markdown'

const props = defineProps({
  content: {
    type: String,
    default: '',
  }
})

const html = computed(() => renderMarkdown(props.content))
</script>


<style scoped>
.markdown-body :deep(img) {
  max-width: 100%;
}
</style>
