<script setup>
import { computed } from "vue";

const props = defineProps({
  acceptedCount: {
    type: Number,
    default: 0,
  },
  submitCount: {
    type: Number,
    default: 0,
  },
});

const percentage = computed(() => {
  if (props.submitCount <= 0) return 0;
  const value = Math.floor((100 * props.acceptedCount) / props.submitCount);
  return Math.min(100, Math.max(0, value));
});

const fillColor = computed(() => {
  if (percentage.value < 30) return "#e5484d";
  if (percentage.value < 60) return "#e8930c";
  return "#2f9e44";
});

const fillStyle = computed(() => ({
  width: `${percentage.value}%`,
  backgroundColor: fillColor.value,
}));

const detailText = computed(
  () => `通过 ${props.acceptedCount} 次，共提交 ${props.submitCount} 次`,
);
</script>

<template>
  <div v-if="submitCount > 0" class="pass-rate" :title="detailText">
    <div
      class="pass-rate-track"
      role="progressbar"
      aria-label="题目通过率"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-valuenow="percentage"
    >
      <div class="pass-rate-fill" :style="fillStyle"></div>
    </div>
    <span class="pass-rate-value">{{ percentage }}%</span>
  </div>
  <span v-else class="pass-rate-empty">—</span>
</template>

<style scoped>
.pass-rate {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.pass-rate-track {
  width: 104px;
  height: 10px;
  overflow: hidden;
  border-radius: 999px;
  background: #eef1f5;
  box-shadow: inset 0 0 0 1px rgb(148 163 184 / 12%);
}

.pass-rate-fill {
  height: 100%;
  border-radius: inherit;
  transition: width 180ms ease;
}

.pass-rate-value {
  width: 42px;
  color: #606b7a;
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  text-align: left;
}

.pass-rate-empty {
  color: #b7bec8;
}

@media (prefers-reduced-motion: reduce) {
  .pass-rate-fill {
    transition: none;
  }
}
</style>
