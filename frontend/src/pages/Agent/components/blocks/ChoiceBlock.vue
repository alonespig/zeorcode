<template>
  <section class="interaction-card">
    <header>
      <h3>{{ block.title }}</h3>
      <p v-if="block.description">{{ block.description }}</p>
    </header>
    <div :class="['choice-list', { grid: block.type === 'multi_select' }]">
      <button
        v-for="option in block.options || []"
        :key="`${typeof option.value}:${String(option.value)}`"
        type="button"
        :class="['choice-item', { selected: isSelected(option.value), disabled: option.disabled }]"
        :disabled="disabled || submitting || option.disabled"
        :role="block.type === 'multi_select' ? 'checkbox' : 'radio'"
        :aria-checked="isSelected(option.value)"
        @click="toggleOption(option.value)"
      >
        <span class="choice-check"><el-icon><Check /></el-icon></span>
        <span class="choice-copy">
          <span class="choice-title">
            {{ option.label }}
            <small v-if="option.recommended">推荐</small>
          </span>
          <span v-if="option.description" class="choice-description">{{ option.description }}</span>
        </span>
      </button>
    </div>
    <footer>
      <span v-if="disabled" class="expired">此选项已处理</span>
      <button v-else type="button" :disabled="!hasValue || submitting" @click="submit">
        确认选择
      </button>
    </footer>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'
import { Check } from '@element-plus/icons-vue'

const props = defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
const emit = defineEmits(['submit'])
const singleValue = ref(null)
const multipleValue = ref([])
const hasValue = computed(() => props.block.type === 'multi_select' ? multipleValue.value.length > 0 : singleValue.value !== null)
const isSelected = (value) => props.block.type === 'multi_select'
  ? multipleValue.value.includes(value)
  : singleValue.value === value

const toggleOption = (value) => {
  if (props.block.type !== 'multi_select') {
    singleValue.value = value
    return
  }

  if (multipleValue.value.includes(value)) {
    multipleValue.value = multipleValue.value.filter(item => item !== value)
    return
  }

  multipleValue.value = [...multipleValue.value, value]
}

const submit = () => emit(
  'submit',
  props.block.requestId,
  props.block.type === 'multi_select' ? [...multipleValue.value] : singleValue.value
)
</script>

<style scoped>
.interaction-card { margin-top: 14px; padding: 18px; border: 1px solid #e1e7f0; border-radius: 7px; background: #fbfcfe; }
header h3 { margin: 0; color: #202c40; font-size: 15px; }
header p { margin: 6px 0 0; color: #768196; font-size: 13px; line-height: 1.6; }
.choice-list { display: flex; flex-direction: column; gap: 8px; margin-top: 15px; }
.choice-list.grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.choice-item { width: 100%; min-height: 48px; display: flex; align-items: center; gap: 11px; padding: 10px 12px; border: 1px solid #e2e7ef; border-radius: 6px; background: #fff; cursor: pointer; font: inherit; text-align: left; transition: .15s; }
.choice-item:hover { border-color: #9abcf8; }
.choice-item.selected { border-color: #699bf6; background: #f2f6ff; }
.choice-item:disabled { cursor: not-allowed; opacity: .55; }
.choice-check { width: 18px; height: 18px; flex: 0 0 18px; display: grid; place-items: center; color: transparent; border: 1px solid #cbd3df; border-radius: 4px; font-size: 12px; }
.choice-item.selected .choice-check { color: #fff; border-color: #367bf5; background: #367bf5; }
.choice-title { display: flex; align-items: center; gap: 7px; color: #39465b; font-size: 13px; font-weight: 600; }
.choice-title small { padding: 2px 5px; color: #22744b; background: #e9f8ef; border-radius: 3px; font-size: 10px; }
.choice-description { display: block; margin-top: 3px; color: #8993a4; font-size: 11px; }
footer { display: flex; justify-content: flex-end; margin-top: 14px; }
footer button { height: 34px; padding: 0 17px; color: #fff; border-radius: 5px; background: #367bf5; font-size: 13px; }
footer button:disabled { background: #bdc9da; cursor: not-allowed; }
.expired { color: #9da6b5; font-size: 12px; }
@media (max-width: 660px) { .choice-list.grid { grid-template-columns: 1fr; } }
</style>
