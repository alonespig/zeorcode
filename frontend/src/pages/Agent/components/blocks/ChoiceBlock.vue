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
    <div v-if="allowCustomInput" class="custom-knowledge">
      <div class="custom-heading">
        <span>自定义知识点</span>
        <small>没有合适选项时可以自己添加</small>
      </div>
      <div class="custom-input-row">
        <el-input
          v-model="customInput"
          :maxlength="customInputMaxLength"
          :placeholder="payload.customInputPlaceholder || '输入其他知识点'"
          :disabled="disabled || submitting || customValues.length >= customInputLimit"
          @keyup.enter.prevent="addCustomValue"
        />
        <button
          class="custom-add-button"
          type="button"
          :disabled="!customInput.trim() || disabled || submitting || customValues.length >= customInputLimit || selectionCount >= maxSelections"
          @click="addCustomValue"
        >添加</button>
      </div>
      <div v-if="customValues.length" class="custom-tags">
        <el-tag
          v-for="item in customValues"
          :key="item"
          :closable="!disabled && !submitting"
          disable-transitions
          @close="removeCustomValue(item)"
        >{{ item }}</el-tag>
      </div>
      <small class="selection-limit">系统选项和自定义知识点合计最多 {{ maxSelections }} 个</small>
    </div>
    <footer>
      <span v-if="disabled" class="expired">此选项已处理</span>
      <template v-else>
        <button class="cancel-button" type="button" :disabled="submitting" @click="emit('cancel')">取消任务</button>
        <button class="primary-button" type="button" :disabled="!hasValue || submitting" @click="submit">确认选择</button>
      </template>
    </footer>
  </section>
</template>

<script setup>
import { computed, ref } from 'vue'
import { Check } from '@element-plus/icons-vue'

const props = defineProps({ block: { type: Object, required: true }, disabled: Boolean, submitting: Boolean })
const emit = defineEmits(['submit', 'cancel'])
const singleValue = ref(null)
const multipleValue = ref([])
const customInput = ref('')
const customValues = ref([])
const payload = computed(() => props.block.payload || {})
const allowCustomInput = computed(() => props.block.type === 'multi_select' && Boolean(payload.value.allowCustomInput))
const customInputMaxLength = computed(() => Number(payload.value.customInputMaxLength) || 30)
const customInputLimit = computed(() => Number(payload.value.customInputLimit) || 8)
const maxSelections = 12
const selectionCount = computed(() => multipleValue.value.length + customValues.value.length)
const hasValue = computed(() => props.block.type === 'multi_select' ? selectionCount.value > 0 : singleValue.value !== null)
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

  if (selectionCount.value >= maxSelections) return

  multipleValue.value = [...multipleValue.value, value]
}

const addCustomValue = () => {
  const value = customInput.value.trim()
  if (!value || customValues.value.length >= customInputLimit.value || selectionCount.value >= maxSelections) return

  const normalized = value.toLocaleLowerCase()
  const matchingOption = (props.block.options || []).find(option => String(option.label).trim().toLocaleLowerCase() === normalized)
  if (matchingOption) {
    if (!multipleValue.value.includes(matchingOption.value)) toggleOption(matchingOption.value)
  } else if (!customValues.value.some(item => item.toLocaleLowerCase() === normalized)) {
    customValues.value = [...customValues.value, value]
  }
  customInput.value = ''
}

const removeCustomValue = (value) => {
  customValues.value = customValues.value.filter(item => item !== value)
}

const submit = () => emit(
  'submit',
  props.block.requestId,
  props.block.type === 'multi_select'
    ? [...multipleValue.value, ...customValues.value.map(item => `custom:${item}`)]
    : singleValue.value
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
.custom-knowledge { margin-top: 15px; padding-top: 14px; border-top: 1px solid #e7ebf2; }
.custom-heading { display: flex; align-items: baseline; gap: 9px; margin-bottom: 8px; color: #39465b; font-size: 13px; font-weight: 600; }
.custom-heading small, .selection-limit { color: #929cad; font-size: 11px; font-weight: 400; }
.custom-input-row { display: flex; gap: 8px; }
.custom-add-button { width: 68px; flex: 0 0 68px; color: #367bf5; border: 1px solid #a8c4f7; border-radius: 5px; background: #fff; font-size: 13px; }
.custom-add-button:hover:not(:disabled) { border-color: #699bf6; background: #f2f6ff; }
.custom-add-button:disabled { color: #aab4c3; border-color: #dfe4ec; background: #f5f6f8; cursor: not-allowed; }
.custom-tags { display: flex; flex-wrap: wrap; gap: 7px; margin-top: 10px; }
.selection-limit { display: block; margin-top: 8px; }
footer { display: flex; justify-content: flex-end; gap: 9px; margin-top: 14px; }
footer button { height: 34px; padding: 0 17px; color: #fff; border-radius: 5px; background: #367bf5; font-size: 13px; }
footer button:disabled { background: #bdc9da; cursor: not-allowed; }
.cancel-button { color: #606d82; border: 1px solid #d8dee8; background: #fff; }
.cancel-button:disabled { color: #9da6b5; background: #f5f6f8; }
.expired { color: #9da6b5; font-size: 12px; }
@media (max-width: 660px) { .choice-list.grid { grid-template-columns: 1fr; } }
</style>
