<template>
  <section class="text-input-card">
    <header>
      <h3>{{ block.title }}</h3>
      <p v-if="block.description">{{ block.description }}</p>
    </header>
    <el-input
      v-model="value"
      type="textarea"
      :rows="3"
      :maxlength="maxLength"
      show-word-limit
      resize="none"
      :placeholder="payload.placeholder || '请输入补充要求'"
      :disabled="disabled || submitting"
    />
    <footer>
      <span v-if="disabled">此设置已提交</span>
      <template v-else>
        <button class="cancel-button" type="button" :disabled="submitting" @click="emit('cancel')">取消任务</button>
        <button class="skip-button" type="button" :disabled="submitting" @click="submit('')">跳过</button>
        <button class="primary-button" type="button" :disabled="!value.trim() || submitting" @click="submit(value)">继续</button>
      </template>
    </footer>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
const emit = defineEmits(['submit', 'cancel'])
const value = ref('')
const payload = computed(() => props.block.payload || {})
const maxLength = computed(() => Number(payload.value.maxLength) || 500)
const submit = (content) => emit('submit', props.block.requestId, content.trim())
</script>

<style scoped>
.text-input-card { margin-top: 14px; padding: 19px; border: 1px solid #dfe6f0; border-radius: 7px; background: #fbfcfe; }
header { margin-bottom: 15px; }
h3 { margin: 0; color: #202c40; font-size: 15px; }
p { margin: 6px 0 0; color: #7c879a; font-size: 13px; line-height: 1.6; }
footer { display: flex; justify-content: flex-end; align-items: center; gap: 9px; margin-top: 14px; }
footer span { color: #9da6b5; font-size: 12px; }
button { height: 34px; padding: 0 16px; border-radius: 5px; font-size: 13px; }
.primary-button { color: #fff; background: #367bf5; }
.skip-button, .cancel-button { color: #606d82; border: 1px solid #d8dee8; background: #fff; }
.cancel-button { margin-right: auto; }
button:disabled { opacity: .55; cursor: not-allowed; }
</style>
