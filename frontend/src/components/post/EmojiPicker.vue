<script setup>
import { shallowRef } from 'vue'

const emit = defineEmits(['select'])

const visible = shallowRef(false)
const emojis = [
  '😀', '😄', '😁', '😂', '🥰', '😍', '😊', '😉',
  '😎', '🤔', '😮', '😢', '😭', '😡', '👍', '👎',
  '👏', '🙌', '🎉', '❤️', '💪', '🔥', '✨', '💡',
  '✅', '❌', '👀', '🙏', '🤝', '🚀', '🐛', '💯',
]

const selectEmoji = (emoji) => {
  emit('select', emoji)
  visible.value = false
}
</script>

<template>
  <el-popover v-model:visible="visible" placement="bottom-start" :width="304" trigger="click" :teleported="false">
    <template #reference>
      <el-button class="emoji-trigger" aria-label="选择表情">
        <span class="emoji-face">😊</span>
        <span>表情</span>
      </el-button>
    </template>

    <div class="grid grid-cols-8 gap-1 p-1" aria-label="表情列表">
      <button v-for="emoji in emojis" :key="emoji" type="button"
        class="flex h-8 w-8 items-center justify-center rounded text-xl transition-colors hover:bg-blue-50 focus-visible:bg-blue-50 focus-visible:outline-none"
        :aria-label="`插入表情 ${emoji}`" @click="selectEmoji(emoji)">
        {{ emoji }}
      </button>
    </div>
  </el-popover>
</template>

<style scoped>
.emoji-trigger {
  height: 36px;
  padding: 0 12px;
}

.emoji-face {
  margin-right: 5px;
  font-size: 17px;
  line-height: 1;
}
</style>
