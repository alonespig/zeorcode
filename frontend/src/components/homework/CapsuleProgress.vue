<script setup>
import { computed } from "vue";

const props = defineProps({
  label: { type: String, default: "" },
  value: { type: Number, default: 0 },
  max: { type: Number, default: 100 },
  compact: { type: Boolean, default: false },
  ariaLabel: { type: String, default: "进度" },
});

const normalizedMax = computed(() => Math.max(Number(props.max) || 0, 0));
const normalizedValue = computed(() => {
  if (normalizedMax.value === 0) return 0;
  return Math.min(Math.max(Number(props.value) || 0, 0), normalizedMax.value);
});
const percentage = computed(() => {
  if (normalizedMax.value === 0) return 0;
  return Math.round((normalizedValue.value / normalizedMax.value) * 100);
});
const displayText = computed(() => `${normalizedValue.value} / ${normalizedMax.value}`);
</script>

<template>
  <div
    class="capsule-progress"
    :class="{
      'capsule-progress--compact': compact,
      'capsule-progress--complete': percentage === 100,
    }"
    role="progressbar"
    :aria-label="ariaLabel"
    :aria-valuenow="percentage"
    :aria-valuetext="displayText"
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <span class="capsule-progress__header">
      <span v-if="label" class="capsule-progress__name">{{ label }}</span>
      <span class="capsule-progress__value tabular-nums">{{ displayText }}</span>
    </span>
    <span class="capsule-progress__track" aria-hidden="true">
      <span class="capsule-progress__fill" :style="{ width: `${percentage}%` }"></span>
    </span>
  </div>
</template>

<style scoped>
.capsule-progress {
  width: 100%;
  display: grid;
  gap: 7px;
}

.capsule-progress--compact {
  flex: 0 0 132px;
  width: 132px;
  gap: 5px;
}

.capsule-progress__header {
  display: flex;
  min-height: 20px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.capsule-progress__name {
  color: #1f2937;
  font-size: 15px;
  font-weight: 500;
  line-height: 20px;
}

.capsule-progress__value {
  margin-left: auto;
  color: #111827;
  font-size: 14px;
  line-height: 20px;
  text-align: right;
}

.capsule-progress--compact .capsule-progress__header {
  min-height: 16px;
}

.capsule-progress--compact .capsule-progress__value {
  color: #475569;
  font-size: 13px;
  line-height: 16px;
}

.capsule-progress__track {
  position: relative;
  display: block;
  width: 100%;
  height: 8px;
  overflow: hidden;
  border: 1px solid #dbe4ef;
  border-radius: 999px;
  background: #f1f5f9;
}

.capsule-progress__fill {
  position: absolute;
  inset: 0 auto 0 0;
  border-radius: inherit;
  background: #bfdbfe;
  transition: width 220ms ease;
}

.capsule-progress--complete .capsule-progress__track {
  border-color: #b9e4c8;
  background: #edf9f1;
}

.capsule-progress--complete .capsule-progress__fill {
  background: #bbebcb;
}

.capsule-progress--complete .capsule-progress__value {
  color: #217747;
}

@media (prefers-reduced-motion: reduce) {
  .capsule-progress__fill {
    transition: none;
  }
}
</style>
