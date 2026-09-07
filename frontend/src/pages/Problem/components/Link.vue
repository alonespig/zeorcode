<template>
  <a :href="href" target="_blank" class="my-link">{{ title }}</a>
</template>

<script setup>
defineProps({
  href: {
    type: String,
    default: '',
  },
  title: {
    type: String,
    default: '',
  },
})
</script>

<style scoped>
.my-link {
  font-size: 16px;
  font-weight: 400;
  line-height: 1.5;
  border-color: rgb(49, 73, 149);
  color: rgb(52, 152, 219);
}
</style>
