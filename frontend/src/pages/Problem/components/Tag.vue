<template>
  <span class="my-tag" :class="tagClass">
    {{ title == 1 ? '简单' : title == 2 ? '中等' : title == 3 ? '困难' : '其他' }}
  </span>
</template>


<script setup>
import { computed } from 'vue'
const props = defineProps({
  title: {
    type: Number,
    default: 0,
  },
})
const tagClass = computed(() => {
  switch (props.title) {
    case 1:
      return 'tag-easy'
    case 2:
      return 'tag-medium'
    case 3:
      return 'tag-hard'
    default:
      return 'tag-default'
  }
})
</script>

<style scoped>
.my-tag {
  font-size: 14px;
  display: inline-block;
  padding: 0 7px;
  box-sizing: border-box;
  font-weight: 400;
  line-height: 1.5;
  border: 1px solid #bfbfbf;
  border-radius: 2px;
  color: rgb(255, 255, 255);
  margin-right: 5px;
}

.my-tag:hover {
  cursor: pointer;
}

/* 简单：绿色 */
.tag-easy {
  background-color: rgb(83, 196, 26);
  border-color: rgb(80, 161, 39);
}

/* 中等：黄色 */
.tag-medium {
  background-color: rgb(243, 156, 17);
  border-color: rgb(202, 137, 32);
}

/* 困难：红色 */
.tag-hard {
  background-color: rgb(156, 61, 207);
  border-color: rgb(138, 62, 179);
}

/* 其他：蓝色 */
.tag-default {
  background-color: rgb(41, 73, 180);
  border-color: rgb(49, 73, 149);
}
</style>
